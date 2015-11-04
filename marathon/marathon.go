package marathon

import (
	"fmt"
	"strings"
	"net/url"

	marathon "github.com/gambol99/go-marathon"
	"github.com/golang/glog"
	"github.com/zalando-techmonkeys/chimp/backend"
	"github.com/zalando-techmonkeys/chimp/conf"
)

type MarathonBackend struct {
	Client marathon.Marathon
}

func New() *MarathonBackend {
	ma := MarathonBackend{}
	ma.Client = initMarathonClient()
	return &ma
}

// Constants
// DOCKERCFG to access private docker repo to pull docker images
// MARATHONURL to access mesos/marathon cluster
const (
	DOCKERCFG string = "file:///root/.dockercfg"
)

// getMarathonClient connects to mesos cluster
// returns marathon interface like tasks, applications
// groups, deployment, subscriptions, ...
func initMarathonClient() marathon.Marathon {
	config := marathon.NewDefaultConfig()
	config.URL = conf.New().Endpoint
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
	arr := make([]string, 0)
	arr = append(arr, filter["team"])
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
func (mb MarathonBackend) GetApp(req *backend.ArtifactRequest) (*backend.Artifact, error) {
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

	tasks := make([]*Task, 0, len(application.Tasks))

	for _, task := range application.Tasks {
		status := true
		statString := "OK"
		for _, hc := range task.HealthCheckResult {
			if hc.Alive == false {
				status = false
			}
		}
		if !status {
			statString = "NOT ALIVE"
		}
		tasks = append(tasks, &Task{Name: task.AppID, Host: task.Host, Ports: task.Ports, Status: statString}) //TODO: also include protocol information with port
	}

	var endpoints []string = make([]string, 0, len(tasks))

	//transforming the data coming from kubernetes into chimp structure
	replicas := make([]*backend.Replica, 0, len(tasks))
	for _, replica := range tasks {
		//copying container data structure
		containers := make([]*backend.Container, 0, 1)
		containers = append(containers, &backend.Container{ImageURL: application.Container.Docker.Image, Status: replica.Status})

		//TODO translating between type, good for decoupling, must be removed!
		ports := make([]*backend.PortType, 0, len(replica.Ports))
		for _, port := range replica.Ports {
			ports = append(ports, &backend.PortType{Port: port, Protocol: ""})
		}
		endpoints = append(endpoints, fmt.Sprintf("http://%s:%s/", replica.Host, intslice2str(replica.Ports, "")))
		replica := backend.Replica{Status: replica.Status, Containers: containers, Endpoints: endpoints, Ports: ports} //HACK, this shouldn't be added only one time
		endpoints = nil
		replicas = append(replicas, &replica)
	}

	artifact := backend.Artifact{
		Name:              application.ID,
		Message:           message,
		Status:            status,
		Labels:            application.Labels,
		Env:               application.Env,
		RunningReplicas:   replicas,
		RequestedReplicas: application.Instances,
		CPUS:              application.CPUs,
		Memory:            application.Mem,
	}

	return &artifact, nil
}

// Deploy deploys a new application
// takes CreateRequest from backend as argument
func (mb MarathonBackend) Deploy(cr *backend.CreateRequest) (string, error) {
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
	}
	app.Labels = labels

	app.Container.Docker.Container(imageurl).ForcePullImage = true

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
func (mb MarathonBackend) Scale(scale *backend.ScaleRequest) (string, error) {
	app, err := mb.Client.ScaleApplicationInstances(scale.Name, scale.Replicas, false) //FORCE disabled by default
	if err != nil {
		glog.Errorf("Could not scale application %s, error: %s", scale.Name, err)
		return "", err
	}
	glog.Infof("Successfully scaled application, appID: %s", app.DeploymentID)
	return app.DeploymentID, nil
}

// Delete provides to delete applications by name
func (mb MarathonBackend) Delete(delReq *backend.ArtifactRequest) (string, error) {
	// if application already exists, it will be deleted
	app, err := mb.Client.DeleteApplication(delReq.Name)
	if err != nil {
		glog.Errorf("Could not delete application %s, error: %s", delReq.Name, err)
		return "", err
	}
	glog.Infof("Successfully deleted application %s", app.DeploymentID)
	return app.DeploymentID, nil
}

func (mb MarathonBackend) UpdateDeployment(req *backend.UpdateRequest) (string, error) {
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
