package marathon

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	marathon "github.com/gambol99/go-marathon"
	"github.com/golang/glog"
	"github.com/zalando-techmonkeys/chimp/conf"
	. "github.com/zalando-techmonkeys/chimp/types"
)

//MarathonBackend is the wrapper for the marathon API
type MarathonBackend struct {
	Client marathon.Marathon
}

//New returns an instance of the marathon backend
func New() *MarathonBackend {
	ma := MarathonBackend{}
	ma.Client = initMarathonClient()
	return &ma
}

// Constants
// DOCKERCFG to access private docker repo to pull docker images
// MARATHONURL to access mesos/marathon cluster
const (
	DOCKERCFG     string = "file:///root/.dockercfg"
	DOCKERVERSION string = "1.9"
)

// getMarathonClient connects to mesos cluster
// returns marathon interface like tasks, applications
// groups, deployment, subscriptions, ...
func initMarathonClient() marathon.Marathon {
	config := marathon.NewDefaultConfig()
	chimpConfig := conf.New()
	config.URL = chimpConfig.Endpoint
	if chimpConfig.MarathonAuth.Enabled {
		config.HttpBasicAuthUser = chimpConfig.MarathonAuth.MarathonHttpUser
		config.HttpBasicPassword = chimpConfig.MarathonAuth.MarathonHttpPassword
	}
	client, err := marathon.NewClient(config)
	if err != nil {
		glog.Fatalf("Failed to create a client for marathon, error: %s", err)
	}
	return client
}

// GetAppNames returns all currently listed applications from marathon
// marathon.Applications is a struct with a lot of details
// about all listed applications
func (mb MarathonBackend) GetAppNames(filter map[string]string) ([]string, error) {
	marathonFilter := make(url.Values, 1)
	var arr []string

	if filter["team"] != "" {
		arr = append(arr, fmt.Sprintf("team==%s", filter["team"]))
	}

	marathonFilter["label"] = arr
	applications, err := mb.Client.Applications(marathonFilter)
	if err != nil {
		glog.Errorf("Could not get applications, error %s", err)
		return nil, err
	}
	deployedApps := make([]string, len(applications.Apps))
	for i, application := range applications.Apps {
		deployedApps[i] = application.ID
	}
	return deployedApps, nil
}

// GetApp returns a specific application from marathon
// marathon.Application is a struct with a lot of details
// about the application itself
func (mb MarathonBackend) GetApp(req *ArtifactRequest) (*Artifact, error) {
	application, err := mb.Client.Application(req.Name)
	if err != nil {
		glog.Errorf("Could not get application %s, error: %s", req.Name, err)
		return nil, err
	}
	var status = "RUNNING" //this is just our base case. we then check the status below
	var message string
	if !application.AllTaskRunning() {
		//deploying or waiting or failed, we just don't know!
		//TODO: this is due to marathon API. This must be discussed and improved in marathon!
		status = "DEPLOYING/WAITING"
		//also in case of errors the app is never "FAILED" when the policy is to accept deployments
		//and try to retry till more resources are available
		if application.LastTaskFailure != nil {
			message = fmt.Sprintf("%s, %s AT %s", application.LastTaskFailure.State,
				application.LastTaskFailure.Message, application.LastTaskFailure.Timestamp)
		}
	}
	endpoints := make([]string, 0, len(application.Tasks))

	//transforming the data coming from kubernetes into chimp structure
	replicas := make([]*Replica, 0, len(application.Tasks))
	for _, replica := range application.Tasks {
		//copying container data structure
		containers := make([]*Container, 0, 1)
		status := true
		statString := "OK"
		for _, hc := range replica.HealthCheckResult {
			if hc.Alive == false {
				status = false
			}
		}
		if !status {
			statString = "NOT ALIVE"
		}

		containerName, err := buildContainerName(replica.Host, replica.ID)
		logInfo := map[string]string{}
		if err == nil { //If the URL cannot be build I don't want to encounter any other problem
			remoteURL := "https://www.scalyr.com/events?mode=log&filter=$logfile%3D%27%2Ffluentd%2F%2F" + containerName + "%27%20$serverHost%3D%27" + strings.Split(replica.Host, ".")[0] + "%27"
			logInfo = map[string]string{"containerName": containerName, "remoteURL": remoteURL}
		}

		container := Container{
			ImageURL: application.Container.Docker.Image,
			Status:   statString,
			LogInfo:  logInfo,
		}
		containers = append(containers, &container)

		ports := make([]*PortType, 0, len(replica.Ports))
		for _, port := range replica.Ports {
			ports = append(ports, &PortType{Port: port, Protocol: ""})
		}
		endpoints = append(endpoints, fmt.Sprintf("http://%s:%s/", replica.Host, intslice2str(replica.Ports, "")))
		replica := Replica{Status: statString, Containers: containers, Endpoints: endpoints, Ports: ports} //HACK, this shouldn't be added only one time
		endpoints = nil
		replicas = append(replicas, &replica)
	}

	var ep string
	if strings.HasPrefix(application.ID, "/") {
		ep = application.ID[1:len(application.ID)]
	} else {
		ep = application.ID
	}
	endpoint := fmt.Sprintf(conf.New().EndpointPattern, ep)

	artifact := Artifact{
		Name:              application.ID,
		Message:           message,
		Status:            status,
		Labels:            application.Labels,
		Env:               application.Env,
		RunningReplicas:   replicas,
		RequestedReplicas: application.Instances,
		CPUS:              application.CPUs,
		Memory:            application.Mem,
		Endpoint:          endpoint,
	}

	return &artifact, nil
}

// Deploy deploys a new application
// takes CreateRequest from backend as argument
func (mb MarathonBackend) Deploy(cr *CreateRequest) (string, error) {
	glog.Infof("Deploying a new application with name %s", cr.Name)
	app := marathon.NewDockerApplication()
	id := cr.Name
	ports := cr.Ports
	cpu := float64(cr.CPULimit)
	storage := 0.0 //TODO: setup limit for storage
	memory := float64(cr.MemoryLimit)
	labels := cr.Labels
	imageurl := cr.ImageURL
	env := cr.Env
	replicas := cr.Replicas

	app.Name(id)
	app.Uris = strings.Fields(DOCKERCFG)
	app.CPU(cpu).Memory(memory).Storage(storage).Count(replicas)
	app.Env = env
	portmappings := make([]*marathon.PortMapping, 0, len(ports))
	for _, port := range ports {
		portmappings = append(portmappings, &marathon.PortMapping{ContainerPort: port, HostPort: 0, Protocol: "tcp"}) //TODO: change to protocol params, we probably want to have UDP too.
	}

	app.Container.Docker.PortMappings = portmappings
	//fluentd implementation
	if conf.New().FluentdEnabled {
		app.Container.Docker.Parameter("log-driver", "fluentd")
		app.Container.Docker.Parameter("log-opt", "\"fluentd-address=localhost:24224\"")
		if DOCKERVERSION == "1.9" {
			//this is unsupported if docker < 1.9
			app.Container.Docker.Parameter("log-opt", "\"tag={{.ImageName}}/{{.Name}}\"")
		} else {
			app.Container.Docker.Parameter("log-opt", "\"fluentd-tag={{.Name}}\"")
		}

	}
	app.Labels = labels

	app.Container.Docker.Container(imageurl).ForcePullImage = true
	volumes := make([]*marathon.Volume, 0, len(cr.Volumes))
	for _, volume := range cr.Volumes {
		volumes = append(volumes, &marathon.Volume{ContainerPath: volume.ContainerPath, HostPath: volume.HostPath, Mode: volume.Mode})
	}

	app.Container.Volumes = volumes

	//forcing basic health checks by default.
	//TODO  must be configurable later.
	checks := make([]*marathon.HealthCheck, 0, 2)
	httpHealth := marathon.HealthCheck{Protocol: "HTTP", Path: "/health", GracePeriodSeconds: 3, IntervalSeconds: 10, MaxConsecutiveFailures: 10}
	cmdHealth := marathon.HealthCheck{Protocol: "COMMAND", Command: &marathon.Command{Value: "curl -f -X GET  http://$HOST:$PORT0/health"}, MaxConsecutiveFailures: 10}
	checks = append(checks, &httpHealth)
	checks = append(checks, &cmdHealth)
	//app.HealthChecks = checks //FIXME health check has been removed because they were constantly killing the instances.
	application, err := mb.Client.CreateApplication(app)
	glog.Info(application) //TODO do we want to get some more information? Container IDs? I guess they can be not stable
	if err != nil {
		glog.Errorf("Could not create application %s, error %s", app.ID, err)
		return "", err
	}
	glog.Infof("Application was created, %s", app.ID)
	return app.ID, nil

}

//Scale is used to scale an application.
//NOTE we should not use this function as our intention is to basically map
//Marathon's REST API. This means we should not use specific ACTIONS, but operations on resources (deployments here)
// Scale provides to scale up or down applications
// Takes two arguments, first one is app name and
// second one is the instances number in total
func (mb MarathonBackend) Scale(scale *ScaleRequest) (string, error) {
	app, err := mb.Client.ScaleApplicationInstances(scale.Name, scale.Replicas, false) //FORCE disabled by default
	if err != nil {
		glog.Errorf("Could not scale application %s, error: %s", scale.Name, err)
		return "", err
	}
	glog.Infof("Successfully scaled application, appID: %s", app.DeploymentID)
	return app.DeploymentID, nil
}

// Delete provides to delete applications by name
func (mb MarathonBackend) Delete(delReq *ArtifactRequest) (string, error) {
	// if application already exists, it will be deleted
	app, err := mb.Client.DeleteApplication(delReq.Name)
	if err != nil {
		glog.Errorf("Could not delete application %s, error: %s", delReq.Name, err)
		return "", err
	}
	glog.Infof("Successfully deleted application %s", app.DeploymentID)
	return app.DeploymentID, nil
}

// UpdateDeployment updates the current deployment. This means the deployment will be restarted
func (mb MarathonBackend) UpdateDeployment(req *UpdateRequest) (string, error) {
	glog.Infof("Updating a previously deployed application")
	app := marathon.NewDockerApplication()
	id := req.Name
	ports := req.Ports
	cpu := float64(req.CPULimit)
	storage := 0.0
	memory := float64(req.MemoryLimit)
	labels := req.Labels
	imageurl := req.ImageURL
	env := req.Env
	replicas := req.Replicas

	app.Name(id)
	app.Uris = strings.Fields(DOCKERCFG)
	app.CPU(cpu)
	app.Memory(memory)
	app.Storage(storage)
	app.Count(replicas)
	app.Env = env
	portmappings := make([]*marathon.PortMapping, 0, len(ports))
	for _, port := range ports {
		portmappings = append(portmappings, &marathon.PortMapping{ContainerPort: port, HostPort: 0, Protocol: "tcp"}) //TODO: change to protocol params, we probably want to have UDP too.
	}

	app.Container.Docker.PortMappings = portmappings
	app.Labels = labels

	app.Container.Docker.Container(imageurl).ForcePullImage = true

	appID, err := mb.Client.UpdateApplication(app)
	return appID.DeploymentID, err
}

func intslice2str(ary []int, sep string) string {
	var str string
	for _, value := range ary {
		str += fmt.Sprintf("%d%s", value, sep)
	}
	return str
}

//this is a definitely not an elegant way of composing the ContainerName, but this information is currently not
//available anywhere else. This MUST be removed or refactored as soon as https://issues.apache.org/jira/browse/MESOS-3688
//is completed and the same information exposed via marathon.
//NOTE: this function ca be further improved as there are many hardcoded infos.
func buildContainerName(host string, ID string) (string, error) {
	containerName := ""
	slaveID := ""
	//call the mesos host (slave) to get the state.json
	req, err := http.NewRequest("GET", fmt.Sprintf("%s://%s:5051/state.json", "http", host), nil)
	if err != nil {
		return "", err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	//parse the state.json to get from "frameworks->execturos->" the entry with id == ID
	var mj interface{}
	err = json.Unmarshal(body, &mj)
	if err != nil {
		return "", err
	}
	data := mj.(map[string]interface{})
	frameworks := data["frameworks"]
	for _, framework := range frameworks.([]interface{}) {
		//check if we are really looking at marathon data as there could be many mesos frameworks
		f := framework.(map[string]interface{})
		if f["name"] == "marathon" {
			executors := f["executors"].([]interface{})
			//iterate on tasks
			for _, executor := range executors {
				exec := executor.(map[string]interface{})
				if exec["id"].(string) == ID { //this is the map we need
					//read field "container" from the entry
					containerName = exec["container"].(string)
					//getting slave ID
					tasks := exec["tasks"].([]interface{})
					if len(tasks) > 0 {
						task := tasks[0].(map[string]interface{})
						slaveID = task["slave_id"].(string)
					}
				}

			}
		}
	}
	//build and return the string
	return fmt.Sprintf("mesos-%s.%s", slaveID, containerName), nil
}
