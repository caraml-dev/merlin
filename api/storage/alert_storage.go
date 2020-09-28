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
	"github.com/jinzhu/gorm"

	"github.com/gojek/merlin/models"
)

// AlertStorage interface.
type AlertStorage interface {
	ListModelEndpointAlerts(modelId models.Id) ([]*models.ModelEndpointAlert, error)
	GetModelEndpointAlert(modelId models.Id, modelEndpointId models.Id) (*models.ModelEndpointAlert, error)
	CreateModelEndpointAlert(alert *models.ModelEndpointAlert) error
	UpdateModelEndpointAlert(alert *models.ModelEndpointAlert) error
	DeleteModelEndpointAlert(modelId models.Id, modelEndpointId models.Id) error
}

type alertStorage struct {
	db *gorm.DB
}

// NewAlertStorage initialize new alert storage backed by GitLab repository.
func NewAlertStorage(db *gorm.DB) AlertStorage {
	return &alertStorage{db}
}

func (s *alertStorage) query() *gorm.DB {
	return s.db.
		Preload("Model").
		Preload("ModelEndpoint").
		Preload("ModelEndpoint.Environment").
		Joins("JOIN models on models.id = model_endpoint_alerts.model_id").
		Joins("JOIN model_endpoints on model_endpoints.id = model_endpoint_alerts.model_endpoint_id")
}

func (s *alertStorage) ListModelEndpointAlerts(modelId models.Id) (alerts []*models.ModelEndpointAlert, err error) {
	err = s.query().
		Where("model_endpoint_alerts.model_id = ?", modelId.String()).
		Find(&alerts).
		Error
	return
}

// GetModelEndpointAlert gets an model endpoint alert.
func (s *alertStorage) GetModelEndpointAlert(modelId models.Id, modelEndpointId models.Id) (*models.ModelEndpointAlert, error) {
	var alert models.ModelEndpointAlert
	err := s.query().
		Where("model_endpoint_alerts.model_id = ? AND model_endpoint_alerts.model_endpoint_id = ?",
			modelId.String(), modelEndpointId.String()).
		First(&alert).
		Error
	if err != nil {
		return nil, err
	}
	return &alert, nil
}

// CreateModelEndpointAlert inserts model endpoint alert
func (s *alertStorage) CreateModelEndpointAlert(alert *models.ModelEndpointAlert) error {
	return s.db.Create(&alert).Error
}

// UpdateModelEndpointAlert updates model endpoint alert
func (s *alertStorage) UpdateModelEndpointAlert(alert *models.ModelEndpointAlert) error {
	return s.db.Save(&alert).Error
}

// DeleteModelEndpointAlert deletes a model endpoint alert.
func (s *alertStorage) DeleteModelEndpointAlert(modelId models.Id, modelEndpointId models.Id) error {
	return s.db.
		Where("model_endpoint_alerts.model_id = ? AND model_endpoint_alerts.model_endpoint_id = ?",
			modelId.String(), modelEndpointId.String()).
		Delete(models.ModelEndpointAlert{}).Error
}
