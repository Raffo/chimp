package backend

type CreateRequest struct {
	BaseRequest
}

type UpdateRequest struct {
	BaseRequest
}

// deploy request
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

// deploy result //TODO do we need to add more info?
type CommonResult string

// Actions on Artifacts
const (
	DEPLOY = iota + 1
	LIST
	INFO
	DELETE
)

// request about exactly one deploy artifact
type ArtifactRequest struct {
	Action int // DELETE, MONITOR, ..
	Name   string
	Labels map[string]string // ["env": "live", "instance": "shop"]
}

type Replica struct {
	Status     string       `json:"status"`
	Endpoints  []string     `json:"endpoints"`
	Ports      []*PortType  `json:"ports"`
	Containers []*Container `json:"containers"`
}

// retrieve information artifact
type Artifact struct {
	Name              string            `json:"name"`
	Message           string            `json:"message"`
	Status            string            `json:"status"`
	Labels            map[string]string `json:"labels"`
	Env               map[string]string `json:"env"`
	RunningReplicas   []*Replica        `json:"runningReplicas"`
	RequestedReplicas int               `json:"requestedReplicas"`
	CPUS              float64           `json:"cpus"`
	Memory            float64           `json:"memory"`
}

type PortType struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
}

//Labels will probably be there, eventually. Makes particularly sense kubernetes, not for marathon
type Container struct {
	ImageURL string            `json:"imageURL"`
	Ports    []*PortType       `json:"ports"`
	Status   string            `json:"status"`
	LogInfo  map[string]string `json:"loginfo"`
	Volumes  []*Volume         `json:"volumes"`
}

type ListDeployments []string

type ScaleRequest struct {
	Name     string
	Replicas int
}

type Volume struct {
	Name          string `json:"name"`
	ContainerPath string `json:"containerPath"`
	HostPath      string `json:"hostPath"`
	Mode          string `json:"mode"`
}
