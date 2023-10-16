// Copyright 2020 The Merlin Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
