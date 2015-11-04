package backend

type Backend interface {
	GetAppNames(filter map[string]string) ([]string, error)
	GetApp(req *ArtifactRequest) (*Artifact, error)
	Deploy(req *CreateRequest) (string, error)
	Scale(scale *ScaleRequest) (string, error)
	Delete(deleteReq *ArtifactRequest) (string, error)
	UpdateDeployment(req *UpdateRequest) (string, error)
}
