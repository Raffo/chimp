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
	backend "github.com/zalando/chimp/backend"
	"github.com/zalando/chimp/conf"
	. "github.com/zalando/chimp/types"
	"github.com/zalando/chimp/validators"
)

//Backend contains the current backend
type Backend struct {
	BackendType string
	Backend     backend.Backend
}

// Bootstrap backend
var se = Backend{
	BackendType: conf.New().BackendType,
}

//Start initializes the current backend
func Start() {
	se.Backend = backend.New()
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

func healthHandler(ginCtx *gin.Context) {
	ginCtx.String(http.StatusOK, "OK")
}

//deployList is used to get a list of all the running application
func deployList(ginCtx *gin.Context) {
	team, uid := buildTeamLabel(ginCtx)
	all := ginCtx.Query("all")
	var filter map[string]string
	if all == "" {
		filter = make(map[string]string, 2)
		filter["uid"] = uid
		filter["team"] = team
	}
	result, err := se.Backend.GetAppNames(filter)
	if err != nil {
		glog.Errorf("Could not get artifacts from backend for LIST request, caused by: %s", err.Error())
		ginCtx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Could not get artifact from backend for INFO, caused by: %s", err)})
		return
	}
	ginCtx.JSON(http.StatusOK, gin.H{"deployments": result})
}

func deployInfo(ginCtx *gin.Context) {
	name := ginCtx.Params.ByName("name")
	glog.Infof("retrieve info by name: %s", name)
	var arReq = ArtifactRequest{Action: INFO, Name: name}
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
	validator := validators.New()
	valid, err := validator.Validate(givenDeploy)
	if !valid {
		glog.Errorf("Invalid request, validation not passed.")
		ginCtx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request."})
		ginCtx.Error(errors.New("Invalid request"))
		return
	}
	team, uid := buildTeamLabel(ginCtx)
	if givenDeploy.Labels == nil {
		givenDeploy.Labels = make(map[string]string, 2)
	}
	givenDeploy.Labels["team"] = team
	givenDeploy.Labels["user"] = uid

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

	volumes := make([]*Volume, len(givenDeploy.Volumes))
	for i, vol := range givenDeploy.Volumes {
		volumes[i] = &Volume{HostPath: vol.HostPath, ContainerPath: vol.ContainerPath, Mode: vol.Mode}
	}

	var beReq = &CreateRequest{BaseRequest: BaseRequest{
		Name: givenDeploy.Name, Ports: givenDeploy.Ports, Labels: givenDeploy.Labels, ImageURL: givenDeploy.ImageURL, Env: givenDeploy.Env, Replicas: givenDeploy.Replicas,
		CPULimit: givenDeploy.CPULimit, MemoryLimit: memoryLimit, Force: givenDeploy.Force, Volumes: volumes}}
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
	validator := validators.New()
	valid, err := validator.Validate(deploy)
	if !valid {
		glog.Errorf("Invalid request, validation not passed.")
		ginCtx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request."})
		ginCtx.Error(errors.New("Invalid request"))
		return
	}
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

	var beReq = UpdateRequest{BaseRequest: BaseRequest{Name: deploy.Name, Ports: deploy.Ports, Labels: deploy.Labels, ImageURL: deploy.ImageURL, Env: deploy.Env, Replicas: deploy.Replicas,
		CPULimit: deploy.CPULimit, MemoryLimit: memoryLimit, Force: deploy.Force}}

	_, err = se.Backend.UpdateDeployment(&beReq)
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
	var ar = ArtifactRequest{Action: DELETE, Name: name}
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
	force := ginCtx.Query("force")
	fs := false
	if force == "true" {
		fs = true
	}
	glog.Info("scaling %s to %d instances", name, num)
	replicas, err := strconv.Atoi(num)
	if err != nil {
		glog.Errorf("Could not change instances for %s, caused by %s", name, err)
		ginCtx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		ginCtx.Error(err)
		return
	}
	var beReq = &ScaleRequest{Name: name, Replicas: replicas, Force: fs}

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
	re, err := regexp.Compile(`^([0-9]*)(MB|GB){0,1}$`)
	if err != nil {
		return 0, err
	}
	res := re.FindStringSubmatch(memory)
	if len(res) == 0 {
		return 0, errors.New("Memory formatting is wrong")
	} else if res[2] == "" { //user didn't specify the size, assuming
		return strconv.Atoi(res[1])
	} else {
		val, err := strconv.Atoi(res[1])
		if res[2] == "GB" {
			return val * 1000, err
		} else if res[2] == "MB" {
			return val, err
		} else {
			return 0, errors.New("Memory formatting is wrong")
		}
	}
}

func buildTeamLabel(ginCtx *gin.Context) (string, string) {
	uid, uidSet := ginCtx.Get("uid")
	team, teamSet := ginCtx.Get("team")
	if uidSet && teamSet {
		return team.(string), uid.(string)
	}
	return "", ""
}
