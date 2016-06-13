// +build ignore

package kubernetes

import (
	"fmt"
	"strconv"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	k8s "github.com/GoogleCloudPlatform/kubernetes/pkg/client"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/fields"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"
	"github.com/golang/glog"
	"github.com/zalando/chimp/conf"
)

var SERVICE_SUFFIX = "-service"

//global init for client
var client = getKubernetesClient()

func getKubernetesClient() *k8s.Client {
	client := k8s.NewOrDie(&k8s.Config{Host: conf.New().Endpoint, Version: "v1"}) //PANIC if config not correct
	return client
}

//getKubernetesVersion returs the version of the API
func getKubernetesVersion() string {
	version, err := client.ServerVersion()
	if err != nil {
		glog.Error("Error getting kubernetes server version ", err)
		return ""
	}
	return version.String()
}

//GetPods gets all the pods running in the cluster or info on one pod
func GetListOfNames() ([]string, error) {
	podsList, err := client.Pods(api.NamespaceDefault).List(labels.Everything(), fields.Everything())
	if err != nil {
		glog.Error("Unable to get list of pods. REASON: ", err)
		return nil, err
	}
	pods := make([]string, 0, len(podsList.Items))

	for _, pod := range podsList.Items {
		pods = append(pods, pod.Name)
	}
	return pods, nil
}

func DeployInfoKubernetes(req *KubernetesDeployInfoRequest) (*KubernetesDeployInfoResponse, error) {
	rcs, err := client.ReplicationControllers(api.NamespaceDefault).Get(req.Name)
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	pods, err := client.Pods(api.NamespaceDefault).List(labels.Set(rcs.Labels).AsSelector(), fields.Everything())
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	endpoints, err := client.Endpoints(api.NamespaceDefault).Get(req.Name + SERVICE_SUFFIX)
	if err != nil {
		glog.Error(err)
		return nil, err
	}
	subset := api.EndpointSubset{}
	eps := make([]string, 0, len(subset.Addresses))
	ports := make([]*KubePort, 0, len(subset.Ports))
	if len(endpoints.Subsets) > 0 {
		subset = endpoints.Subsets[0]
		for _, address := range subset.Addresses {
			eps = append(eps, address.IP)
		}
		for _, port := range subset.Ports {
			ports = append(ports, &KubePort{Port: port.Port, Protocol: string(port.Protocol)})
		}
	}

	podsInfo := make([]*KubernetesPodInfo, 0, len(pods.Items))
	var deploymentReady bool = true
	for _, item := range pods.Items {
		//ips := []string{}
		//ips = append(ips, item.Status.PodIP) //DOC we return PodIP for the endpoint
		containers := make([]*ContainerInfo, 0, len(item.Status.ContainerStatuses))
		var podReady bool = true
		for _, container := range item.Status.ContainerStatuses {
			if !container.Ready {
				podReady = false
			}
			status, message := describeStatus(req.Name, container.State)
			containerInfo := ContainerInfo{ImageURL: container.Image, Status: fmt.Sprintf("%s", status), Message: message}
			containers = append(containers, &containerInfo) //TODO this must get the port from the POD.
		}

		info := KubernetesPodInfo{
			Status:     fmt.Sprintf("Ready: %s", strconv.FormatBool(podReady)),
			Containers: containers,
			Ports:      ports,
			Endpoints:  eps,
		}
		if !podReady {
			deploymentReady = false
		}

		podsInfo = append(podsInfo, &info)
	}

	//TODO how should kubernetes respond in case of a deployment that is over? Shall we still provide info?

	response := KubernetesDeployInfoResponse{Name: req.Name,
		Labels:    req.Labels,
		Status:    fmt.Sprintf("Ready: %s", strconv.FormatBool(deploymentReady)),
		Replicas:  podsInfo,
		Instances: rcs.Status.Replicas}
	return &response, nil
}

//CreateService creates a kubernetes service.
//This shouldn't be used directly by any other part of the application for now
func createService(req *KubernetesDeployRequest, force bool) error {
	//building ports datastructure cause kubernetes uses a pretty strange one
	ports := make([]api.ServicePort, 0, len(req.Ports))
	for _, port := range req.Ports {
		ports = append(ports, api.ServicePort{Port: port})
	}
	glog.Info(ports)
	serviceName := req.Name + SERVICE_SUFFIX
	service := api.Service{ObjectMeta: api.ObjectMeta{Name: serviceName}, //CHECK: convention to name services
		Spec: api.ServiceSpec{Selector: req.Labels, Ports: ports, Type: "NodePort"}}
	_, err := client.Services(api.NamespaceDefault).Create(&service)
	if err != nil {
		glog.Error("Cannot create service")
		return err
	}
	return nil
}

//Deploy to deploy something on kubernetes. Creates a service, a replication controller and the given pods
func Deploy(req *KubernetesDeployRequest, force bool) (*KubernetesDeployResponse, error) {
	if force {
		Delete(&KubernetesDeleteRequest{Name: req.Name})
	}

	serviceErr := createService(req, force)
	if serviceErr != nil {
		glog.Error("Error creating the service.")
		return nil, serviceErr //just returning the same error as above
	}

	glog.Infof("req: %+v", req)

	controller := api.ReplicationController{
		ObjectMeta: api.ObjectMeta{
			Name:   req.Name,
			Labels: req.Labels,
		},
		Spec: api.ReplicationControllerSpec{
			Replicas: req.Replicas,
			Selector: req.Labels, //FIXME: replicas and selectors shouldn't be the same thing!
			Template: &api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Labels: req.Labels,
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Name:  req.Name, //FIXME: container name should be different from replica-controller name, yet invisible for the user
							Image: req.ImageURL,
						},
					},
				},
			},
		},
	}

	glog.Infof("controller: %+v", controller)
	result, err := client.ReplicationControllers(api.NamespaceDefault).Create(&controller)
	if err != nil {
		glog.Error("Error starting a replication controller.")
		return nil, err
	}
	glog.Info(result)
	return &KubernetesDeployResponse{Status: strconv.FormatInt(result.Status.ObservedGeneration, 10),
			Name: result.ObjectMeta.Name},
		nil

}

//Delete deletes services, replication containers, pods that were created with Deploy above
func Delete(deleteReq *KubernetesDeleteRequest) error {
	err := client.ReplicationControllers(api.NamespaceDefault).Delete(deleteReq.Name)
	if err != nil {
		glog.Error(err)
		return err
	}
	err = client.Services(api.NamespaceDefault).Delete(deleteReq.Name + SERVICE_SUFFIX)
	if err != nil {
		glog.Error(err)
		return err
	}
	return nil
}

//making human readable the magic that Kubernetes creates!
func describeStatus(stateName string, state api.ContainerState) (string, string) {
	switch {
	case state.Running != nil:
		return fmt.Sprintf("%s: Running\n", stateName), ""
	case state.Waiting != nil:
		if state.Waiting.Reason != "" {
			return fmt.Sprintf("%s: Waiting\n", stateName), fmt.Sprintf("Reason: %s\n", state.Waiting.Reason)
		} else {
			return fmt.Sprintf("%s: Waiting\n", stateName), ""
		}
	case state.Terminated != nil:
		if state.Terminated.Reason != "" {
			return fmt.Sprintf("%s: Terminated\n", stateName), fmt.Sprintf("Reason: %s\n", state.Terminated.Reason)
		} else {
			return fmt.Sprintf("%s: Terminated\n", stateName), ""
		}
	default:
		return fmt.Sprintf("%s: Waiting\n", stateName), ""
	}
}
