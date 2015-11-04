package mockbackend

import . "github.com/zalando-techmonkeys/chimp/backend"

type MockBackend struct {
	ErrorProbability float32
}

func New() *MockBackend {
	ma := MockBackend{}
	ma.ErrorProbability = 0.0
	return &ma
}

func (mb MockBackend) GetAppNames(filter map[string]string) ([]string, error) {
	return []string{"fake-cat"}, nil
}

func (mb MockBackend) GetApp(req *ArtifactRequest) (*Artifact, error) {
	replicas := make([]*Replica, 0, 1)
	containers := make([]*Container, 0, 1)
	containers = append(containers, &Container{ImageURL: "pierone.test.techmonkeys", Status: "OK"})
	replicas = append(replicas, &Replica{Status: "RUNNING", Containers: containers, Endpoints: []string{"localhost:8888"}, Ports: nil})
	artifact := Artifact{
		Name:              "fake-cat",
		Message:           "there should be no message",
		Status:            "status",
		Labels:            map[string]string{},
		Env:               map[string]string{},
		RunningReplicas:   replicas,
		RequestedReplicas: 1,
		CPUS:              1,
		Memory:            2048.0,
	}
	return &artifact, nil
}

func (mb MockBackend) Deploy(req *CreateRequest) (string, error) {
	return "fake-cat", nil
}

func (mb MockBackend) Scale(scale *ScaleRequest) (string, error) {
	return "fake-cat", nil
}

func (mb MockBackend) Delete(deleteReq *ArtifactRequest) (string, error) {
	return "fake-cat", nil
}

func (mb MockBackend) UpdateDeployment(req *UpdateRequest) (string, error) {
	return "fake-cat", nil
}
