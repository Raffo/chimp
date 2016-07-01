package types

//CreateRequest is the request of deployment
type CreateRequest struct {
	BaseRequest
}

//UpdateRequest is the request for updating an existing deployment
type UpdateRequest struct {
	BaseRequest
}

//BaseRequest represents common data among create/update request
type BaseRequest struct {
	Name        string            // "shop"
	Ports       []int             // [8080]
	Labels      map[string]string // ["env": "live", "instance": "shop"]
	ImageURL    string
	Env         map[string]string // {"FOO": "bar", ..}
	Replicas    int               // 4, creates 4 given container
	CPULimit    int
	MemoryLimit int
	Force       bool
	Volumes     []*Volume
}

// Actions on Artifacts
const (
	DEPLOY = iota + 1
	LIST
	INFO
	DELETE
)

//ArtifactRequest request about exactly one deploy artifact
type ArtifactRequest struct {
	Action int // DELETE, MONITOR, ..
	Name   string
	Labels map[string]string // ["env": "live", "instance": "shop"]
}

//Replica describes the status of an instance of the app
type Replica struct {
	Status     string       `json:"status"`
	Endpoints  []string     `json:"endpoints"`
	Ports      []*PortType  `json:"ports"`
	Containers []*Container `json:"containers"`
}

//Artifact is used to  retrieve information on an app
type Artifact struct {
	Name              string             `json:"name"`
	Message           string             `json:"message"`
	Status            string             `json:"status"`
	Labels            *map[string]string `json:"labels"`
	Env               *map[string]string `json:"env"`
	RunningReplicas   []*Replica         `json:"runningReplicas"`
	RequestedReplicas int                `json:"requestedReplicas"`
	CPUS              float64            `json:"cpus"`
	Memory            float64            `json:"memory"`
	Endpoint          string             `json:"endpoint"`
}

//PortType represents the port and protocol used
type PortType struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
}

//Container represents the information of the particular container of a replica.
//Currently we suppose that one replica has exactly one container which is not true when mapping kubernetes,
//see pod -> containers mappings.
type Container struct {
	ImageURL string            `json:"imageURL"`
	Ports    []*PortType       `json:"ports"`
	Status   string            `json:"status"`
	LogInfo  map[string]string `json:"loginfo"`
	Volumes  []*Volume         `json:"volumes"`
}

//ListDeployments is a list of names of apps currently deployed
type ListDeployments struct {
	Deployments []string `json:"deployments"`
}

//ScaleRequest is a request for scaling an app based on the name of the app
type ScaleRequest struct {
	Name     string
	Replicas int
	Force    bool
}

//Volume represent a volume that can be mounted in an app
type Volume struct {
	Name          string `json:"name"`
	ContainerPath string `json:"containerPath"`
	HostPath      string `json:"hostPath"`
	Mode          string `json:"mode"`
}

//ChimpDefinition is a general definition for an application.
//A chimp definition can contain several deploy requests.
type ChimpDefinition struct {
	DeployRequest []CmdClientRequest
}

//CmdClientRequest is a request to deploy an application
type CmdClientRequest struct {
	Name        string
	ImageURL    string
	Replicas    int
	Ports       []int
	Labels      map[string]string
	Env         map[string]string
	CPULimit    int
	MemoryLimit string
	Force       bool
	Volumes     []*Volume
}

//Error is a small struct for an error type
type Error struct {
	Err string `json:"error"`
}

//DeployRequest is the struct used to represent a request to deploy
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
	Volumes     []*Volume
}
