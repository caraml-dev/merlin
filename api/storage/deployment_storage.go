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

package storage

import (
	"gorm.io/gorm"

	"github.com/caraml-dev/merlin/models"
)

type DeploymentStorage interface {
	// ListInModel return all deployment within a model
	ListInModel(model *models.Model) ([]*models.Deployment, error)
	// Save save the deployment to underlying storage
	Save(deployment *models.Deployment) (*models.Deployment, error)
	// GetFirstSuccessModelVersionPerModel Return mapping of model id and the first model version with a successful model version
	GetFirstSuccessModelVersionPerModel() (map[models.ID]models.ID, error)
	Delete(modelID models.ID, versionID models.ID) error
}

type deploymentStorage struct {
	db *gorm.DB
}

func NewDeploymentStorage(db *gorm.DB) DeploymentStorage {
	return &deploymentStorage{db: db}
}

func (d *deploymentStorage) ListInModel(model *models.Model) ([]*models.Deployment, error) {
	var deployments []*models.Deployment
	err := d.db.Where("version_model_id = ?", model.ID).Find(&deployments).Error
	return deployments, err
}

func (d *deploymentStorage) Save(deployment *models.Deployment) (*models.Deployment, error) {
	err := d.db.Save(deployment).Error
	return deployment, err
}

func (d *deploymentStorage) GetFirstSuccessModelVersionPerModel() (map[models.ID]models.ID, error) {
	rows, err := d.db.Table("deployments").
		Select("version_model_id , min(version_id)").
		Where("status = 'running' or status = 'serving'").
		Group("version_model_id").
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint: errcheck

	resultMap := make(map[models.ID]models.ID)
	for rows.Next() {
		var modelID models.ID
		var versionID models.ID
		if err := rows.Scan(&modelID, &versionID); err != nil {
			return nil, err
		}
		resultMap[modelID] = versionID
	}
	return resultMap, nil
}

func (d *deploymentStorage) Delete(modelID models.ID, versionID models.ID) error {
	return d.db.Where("version_id = ? AND version_model_id = ?", versionID, modelID).Delete(models.Deployment{}).Error
}
