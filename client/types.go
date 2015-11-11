package client

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
}

type Error struct {
	Err string `json:"error"`
}

type ListDeployments struct {
	Deployments []string `json:"deployments"`
}
