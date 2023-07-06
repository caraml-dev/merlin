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
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/caraml-dev/merlin/models"
)

const maxMessageChar = 2048

type VersionEndpointStorage interface {
	ListEndpoints(model *models.Model, version *models.Version) (endpoints []*models.VersionEndpoint, err error)
	Get(uuid.UUID) (*models.VersionEndpoint, error)
	Save(endpoint *models.VersionEndpoint) error
	CountEndpoints(environment *models.Environment, model *models.Model) (int, error)
}

type versionEndpointStorage struct {
	db *gorm.DB
}

func NewVersionEndpointStorage(db *gorm.DB) VersionEndpointStorage {
	return &versionEndpointStorage{db: db}
}

func (v *versionEndpointStorage) ListEndpoints(model *models.Model, version *models.Version) (endpoints []*models.VersionEndpoint, err error) {
	err = v.query().Where("version_endpoints.version_model_id = ? AND version_endpoints.version_id = ?", model.ID, version.ID).Find(&endpoints).Error
	return
}

func (v *versionEndpointStorage) Get(uuid uuid.UUID) (*models.VersionEndpoint, error) {
	ve := &models.VersionEndpoint{}
	if err := v.query().Where("version_endpoints.id = ?", uuid.String()).Find(&ve).Error; err != nil {
		return nil, err
	}
	return ve, nil
}

func (v *versionEndpointStorage) Save(endpoint *models.VersionEndpoint) error {
	sanitizeEndpoint(endpoint)
	return v.db.Save(&endpoint).Error
}

func sanitizeEndpoint(endpoint *models.VersionEndpoint) {
	message := endpoint.Message
	if len(message) > maxMessageChar {
		message = message[:maxMessageChar]
	}
	endpoint.Message = message
}

func (v *versionEndpointStorage) CountEndpoints(environment *models.Environment, model *models.Model) (int, error) {
	var count int64
	err := v.query().
		Model(&models.VersionEndpoint{}).
		Where("version_endpoints.environment_name = ? AND version_endpoints.version_model_id = ? AND version_endpoints.status IN ('pending', 'running', 'serving')", environment.Name, model.ID).
		Count(&count).Error
	return int(count), err
}

func (v *versionEndpointStorage) query() *gorm.DB {
	return v.db.
		Preload("Environment").
		Preload("Transformer").
		Joins("JOIN environments on environments.name = version_endpoints.environment_name")
}
