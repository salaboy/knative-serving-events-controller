package models

type CDEventType string

func (e CDEventType) String() string {
	return string(e)
}

const (
	CreateKService CDEventType = "cd.service.created.v1"
)

type KService struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Image     string `json:"image"`
}
