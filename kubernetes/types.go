package kubernetes

//KubernetesDeployRequest struct to request creation of replication controllers
type KubernetesDeployRequest struct {
	Name        string
	Ports       []int
	Labels      map[string]string
	ImageURL    string
	Env         map[string]string
	Replicas    int
	CPULimit    int
	MemoryLimit int
}

//KubernetesDeployResponse struct used as response to every deploy
type KubernetesDeployResponse struct {
	Status string
	Name   string
}

//KubernetesDeleteRequest struct to request delete of replication controllers
type KubernetesDeleteRequest struct {
	Name string
}

//KubernetesDeployInfoResponse struct to give information to the caller about the status
type KubernetesDeployInfoRequest struct {
	Name   string
	Labels map[string]string
}

type ContainerInfo struct {
	ImageURL string
	Status   string
	Message  string
}

type KubernetesPodInfo struct {
	Status     string
	Endpoints  []string
	Ports      []*KubePort
	Containers []*ContainerInfo
}

type KubePort struct {
	Port     int
	Protocol string
}

//KubernetesDeployInfoResponse struct to give information to the caller about the status
//TODO: status must be somehow an aggregate of what is appening of the containers?
type KubernetesDeployInfoResponse struct {
	Name      string
	Status    string //If not all of the pods/containers are running, we ruturn a "pending" state, "done" otherwise
	Instances int    //TODO remove this field as it can be easily inferred
	Env       map[string]string
	Labels    map[string]string
	Replicas  []*KubernetesPodInfo
}
