package backend

import . "github.com/zalando-techmonkeys/chimp/types"

//Backend is the interface with all the methods that any backend should implement to be run in chimp
type Backend interface {
	GetAppNames(filter map[string]string) ([]string, error)
	GetApp(req *ArtifactRequest) (*Artifact, error)
	Deploy(req *CreateRequest) (string, error)
	Scale(scale *ScaleRequest) (string, error)
	Delete(deleteReq *ArtifactRequest) (string, error)
	UpdateDeployment(req *UpdateRequest) (string, error)
}

type backendFactory func() Backend

var New backendFactory
