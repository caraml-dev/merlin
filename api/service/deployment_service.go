package service

import (
	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/storage"
)

type DeploymentService interface {
	ListDeployments(modelID, versionID, endpointUUID string) ([]*models.Deployment, error)
}

func NewDeploymentService(storage storage.DeploymentStorage) DeploymentService {
	return &deploymentService{
		storage: storage,
	}
}

type deploymentService struct {
	storage storage.DeploymentStorage
}

func (service *deploymentService) ListDeployments(modelID, versionID, endpointUUID string) ([]*models.Deployment, error) {
	return service.storage.ListInModelVersion(modelID, versionID, endpointUUID)
}
