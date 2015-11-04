package api

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/golang/glog"
	backend "github.com/zalando-techmonkeys/chimp/backend"
	"github.com/zalando-techmonkeys/chimp/conf"
	"github.com/zalando-techmonkeys/chimp/marathon"
	mock "github.com/zalando-techmonkeys/chimp/mockbackend"
)

type DeployRequest struct {
	Name        string            // "shop"
	Labels      map[string]string // {"env": "live", "project": "shop"}
	Env         map[string]string // {"FOO": "bar", ..}
	Replicas    int               // 4, creates 4 given container
	Ports       []int
	ImageURL    string
	CPULimit    int
	MemoryLimit string
	Force       bool
}

// If we need data bound to the backend we put it here, p.e. channels to communicate with frontend
type Backend struct {
	BackendType string
	Backend     backend.Backend
}

// Bootstrap backend
var se Backend = Backend{
	BackendType: conf.New().BackendType,
}

func Start() {
	glog.Infof("backend type from config: %s", se.BackendType)
	//Where the backend switch happens
	switch se.BackendType {
	case "marathon":
		se.Backend = marathon.New()
	case "mock":
		se.Backend = mock.New()
	}
}

//BackendError is the erro representation that should be consumed by "frontend" serving layer
//resembles (but no need to be 1:1) what frontend needs to give to the user which is based on our RESTful API Guidelines doc.
type BackendError struct {
	Status int
	Title  string //error message coming from THIS layer
	Detail string //error message coming from backends
}

func rootHandler(ginCtx *gin.Context) {
	config := conf.New()
	ginCtx.JSON(http.StatusOK, gin.H{"chimp-server": fmt.Sprintf("Build Time: %s - Git Commit Hash: %s", config.VersionBuildStamp, config.VersionGitHash)})
}

//deployList is used to get a list of all the running application
func deployList(ginCtx *gin.Context) {
	team, uid := buildTeamLabel(ginCtx)
	all := ginCtx.Query("all")
	var filter map[string]string = nil
	if all == "" {
 		filter = make(map[string]string, 2)
		filter["uid"] = uid
		filter["team"] = team
	}
	result, err := se.Backend.GetAppNames(filter)
	if err != nil {
		glog.Errorf("Could not get artifacts from backend for LIST request, caused by: %s", err.Error())
	}
	ginCtx.JSON(http.StatusOK, gin.H{"deployments": result})
}

func deployInfo(ginCtx *gin.Context) {
	name := ginCtx.Params.ByName("name")
	glog.Infof("retrieve info by name: %s", name)
	var arReq backend.ArtifactRequest = backend.ArtifactRequest{Action: backend.INFO, Name: name}
	result, err := se.Backend.GetApp(&arReq)
	if err != nil {
		glog.Errorf("Could not get artifact from backend for INFO request with name %s, caused by: %s", name, err.Error())
		ginCtx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Could not get artifact from backend for INFO request with name %s, caused by: %s", name, err)})
		return
	}
	ginCtx.JSON(http.StatusOK, result)
}

func deployCreate(ginCtx *gin.Context) {
	givenDeploy, err := commonDeploy(ginCtx)
	team, uid := buildTeamLabel(ginCtx)
	if givenDeploy.Labels == nil {
		givenDeploy.Labels = make(map[string]string, 2)
	}
	givenDeploy.Labels[team] = uid

	ginCtx.Set("data", givenDeploy)
	if err != nil {
		glog.Errorf("Could not update deploy, caused by: %s", err.Error())
		ginCtx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		ginCtx.Error(err)
		return
	}

	memoryLimit, e := mapMemory(givenDeploy.MemoryLimit)
	if e != nil {
		glog.Errorf("Could not create a deploy, caused by: %s", e.Error())
		ginCtx.JSON(http.StatusBadRequest, gin.H{"error": e.Error()})
		ginCtx.Error(err)
		return
	}

	var beReq *backend.CreateRequest = &backend.CreateRequest{BaseRequest: backend.BaseRequest{
		Name: givenDeploy.Name, Ports: givenDeploy.Ports, Labels: givenDeploy.Labels, ImageURL: givenDeploy.ImageURL, Env: givenDeploy.Env, Replicas: givenDeploy.Replicas,
		CPULimit: givenDeploy.CPULimit, MemoryLimit: memoryLimit, Force: givenDeploy.Force}}
	beRes, err := se.Backend.Deploy(beReq)
	if err != nil {
		glog.Errorf("Could not create a deploy, caused by: %s", err.Error())
		ginCtx.JSON(http.StatusNotAcceptable, gin.H{"error": err.Error()})
		ginCtx.Error(err)
		return
	}
	glog.Infof("Deployed: %+v\n", beRes)
	ginCtx.JSON(http.StatusOK, gin.H{"name": beRes})
}

func deployUpsert(ginCtx *gin.Context) {
	deploy, err := commonDeploy(ginCtx)
	ginCtx.Set("data", deploy)
	if err != nil {
		glog.Errorf("Could not update deploy, caused by: %s", err.Error())
		ginCtx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		ginCtx.Error(err)
		return
	}

	memoryLimit, err := mapMemory(deploy.MemoryLimit)
	if err != nil {
		glog.Errorf("Could not create a deploy, caused by: %s", err.Error())
		ginCtx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		ginCtx.Error(err)
		return
	}

	var beReq *backend.UpdateRequest = &backend.UpdateRequest{BaseRequest: backend.BaseRequest{Name: deploy.Name, Ports: deploy.Ports, Labels: deploy.Labels, ImageURL: deploy.ImageURL, Env: deploy.Env, Replicas: deploy.Replicas,
		CPULimit: deploy.CPULimit, MemoryLimit: memoryLimit, Force: deploy.Force}}

	_, err = se.Backend.UpdateDeployment(beReq)
	if err != nil {
		glog.Errorf("Could not update deploy, caused by: %s", err.Error())
		ginCtx.JSON(http.StatusNotAcceptable, gin.H{"error": err.Error()})
		ginCtx.Error(err)
		return
	}
	glog.Infof("Deployment updated")
	ginCtx.JSON(http.StatusOK, gin.H{})
}

func deployDelete(ginCtx *gin.Context) {
	name := ginCtx.Params.ByName("name")
	glog.Info("delete by name: %s", name)
	var ar backend.ArtifactRequest = backend.ArtifactRequest{Action: backend.DELETE, Name: name}
	_, err := se.Backend.Delete(&ar)
	if err != nil {
		glog.Errorf("Could not get artifact from backend for CANCEL request with name %s, caused by: %s", name, err.Error())
		ginCtx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		ginCtx.Error(err)
		return
	}
	ginCtx.JSON(http.StatusOK, gin.H{})
}

func deployReplicasModify(ginCtx *gin.Context) {
	name := ginCtx.Params.ByName("name")
	num := ginCtx.Params.ByName("num")
	glog.Info("scaling %s to %d instances", name, num)
	replicas, err := strconv.Atoi(num)
	if err != nil {
		glog.Errorf("Could not change instances for %s, caused by %s", name, err)
		ginCtx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		ginCtx.Error(err)
		return
	}
	var beReq *backend.ScaleRequest = &backend.ScaleRequest{Name: name, Replicas: replicas}

	_, err = se.Backend.Scale(beReq)
	if err != nil {
		glog.Errorf("Could not change instances for %s, caused by: %s", name, err.Error())
		ginCtx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		ginCtx.Error(err)
		return
	}
	ginCtx.JSON(http.StatusOK, gin.H{})
}

func commonDeploy(ginCtx *gin.Context) (DeployRequest, error) {
	ginCtx.Request.ParseForm()
	var givenDeploy DeployRequest
	ginCtx.BindWith(&givenDeploy, binding.JSON)
	glog.Infof("given %+v", givenDeploy)
	return givenDeploy, nil
}

//TODO this a very hacky implementation of resource handling.
//A much better implementation is found in "github.com/GoogleCloudPlatform/kubernetes/pkg/api/resource"
//We do not support this, but we only express in MB because this is the only amount supported by
//marathon. I (rdifazio) really believe this a limitation in terms  of expressibite, but we keep it
//simple for now.
func mapMemory(memory string) (int, error) {
	re, err := regexp.Compile(`^([0-9]*)(MB){0,1}$`)
	if err != nil {
		return 0, err
	}
	res := re.FindStringSubmatch(memory)
	if len(res) == 0 {
		return 0, errors.New("Memory formatting is wrong")
	}
	return strconv.Atoi(res[1])
}

func buildTeamLabel(ginCtx *gin.Context) (string, string) {
	uid, uidSet := ginCtx.Get("uid")
	team, teamSet := ginCtx.Get("team")
	if uidSet && teamSet {
		return team.(string), uid.(string)
	} else {
		return "", ""
	}
}
