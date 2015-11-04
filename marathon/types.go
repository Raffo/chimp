package marathon

// MarathonDeployRequest provides all expected fields
// to create and update an application
type MarathonDeployRequest struct {
	Name     string            // "shop"
	CPU      float64           // 2
	Memory   float64           // 2048
	Storage  float64           // 0.0
	Ports    []int             // [8080]
	Labels   map[string]string // ["environment": "live", "team": "shop"]
	ImageURL string            
	Env      map[string]string // {"FOO": "bar", ..}
	Replicas int               // 4, creates 4 given container
}

// MarathonDeployResponse provides fields
// to deploy applications
type MarathonDeployResponse struct {
	Status string
	Name   string
}

// MarathonScaleRequest provides expected fields
// to scale a application by total instances
type MarathonScaleRequest struct {
	Name      string
	Instances int
}

//MarathonDeleteRequest provides fields
// to delete applications
type MarathonDeleteRequest struct {
	Name string
}

type MarathonGetAppRequest struct {
	Name string
}

// MarathonGetAppResponse provides fields
// to information of application
type MarathonGetAppResponse struct {
	ID           string
	ImageURL     string
	CPUs         float64
	Mem          float64
	Instances    int
	Ports        []int
	DeploymentID []map[string]string
	Tasks        []*Task
	Labels       map[string]string
	Env          map[string]string
	Status       string
	Message      string
}

type Task struct {
	Name   string
	Status string
	Host   string
	Ports  []int
}
