// +build mock !marathon

package backend

import (
	. "github.com/zalando/chimp/types"
)

//MockBackend is used to have a fake backend that mocks the current behaviour of a backend.
//This is mostly model on marathon, but should be somehow consistent with kubernetes or any other backend.
type MockBackend struct{}

func NewMockBackend() Backend {
	ma := &MockBackend{}
	return ma
}

func init() {
	New = NewMockBackend
}

//GetAppNames is used to get a list of names for app deployed
func (mb *MockBackend) GetAppNames(filter map[string]string) ([]string, error) {
	return []string{"fake-cat"}, nil
}

//GetApp is used to get information related to one specif app
func (mb *MockBackend) GetApp(req *ArtifactRequest) (*Artifact, error) {
	replicas := make([]*Replica, 0, 1)
	containers := make([]*Container, 0, 1)
	containers = append(containers, &Container{ImageURL: "pierone.test.techmonkeys", Status: "OK"})
	replicas = append(replicas, &Replica{Status: "RUNNING", Containers: containers, Endpoints: []string{"localhost:8888"}, Ports: nil})
	artifact := Artifact{
		Name:              "fake-cat",
		Message:           "there should be no message",
		Status:            "status",
		Labels:            &map[string]string{},
		Env:               &map[string]string{},
		RunningReplicas:   replicas,
		RequestedReplicas: 1,
		CPUS:              1,
		Memory:            2048.0,
	}
	return &artifact, nil
}

//Deploy deploys an application. In this case, it only returns the name of the app deployed
func (mb *MockBackend) Deploy(req *CreateRequest) (string, error) {
	return "fake-cat", nil
}

//Scale scales an already deployed application. In this case it only returns the name of the app deleted
func (mb *MockBackend) Scale(scale *ScaleRequest) (string, error) {
	return "fake-cat", nil
}

//Delete deletes and application currently deployed. In this case it only returns the name of the app deleted
func (mb *MockBackend) Delete(deleteReq *ArtifactRequest) (string, error) {
	return "fake-cat", nil
}

//UpdateDeployment updates one application currenly deployed. In this case it only returns the name of the app updated.
func (mb *MockBackend) UpdateDeployment(req *UpdateRequest) (string, error) {
	return "fake-cat", nil
}
