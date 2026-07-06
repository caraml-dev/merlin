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
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/caraml-dev/merlin/log"
	"github.com/caraml-dev/merlin/models"
)

const maxMessageChar = 2048

type VersionEndpointStorage interface {
	ListEndpoints(model *models.Model, version *models.Version) (endpoints []*models.VersionEndpoint, err error)
	Get(uuid.UUID) (*models.VersionEndpoint, error)
	Save(endpoint *models.VersionEndpoint) error
	CountEndpoints(environment *models.Environment, model *models.Model) (int, error)
	Delete(endpoint *models.VersionEndpoint) error
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

	if err := v.db.Save(&endpoint).Error; err != nil {
		if invalid := invalidUTF8Fields(map[string]string{
			"status":                 string(endpoint.Status),
			"url":                    endpoint.URL,
			"service_name":           endpoint.ServiceName,
			"inference_service_name": endpoint.InferenceServiceName,
			"namespace":              endpoint.Namespace,
			"environment_name":       endpoint.EnvironmentName,
			"message":                endpoint.Message,
		}); len(invalid) > 0 {
			log.Errorf("failed to save version_endpoint (id: %s): invalid UTF-8 in column(s) %s: %v", endpoint.ID, strings.Join(invalid, ", "), err)
		}
		return err
	}

	if endpoint.Transformer != nil {
		return v.db.Save(endpoint.Transformer).Error
	}

	return nil
}

// invalidUTF8Fields returns the names of the given columns whose values are not
// valid UTF-8. Postgres reports "invalid byte sequence for encoding UTF8"
// (SQLSTATE 22021) without naming the offending column, so this pinpoints it.
func invalidUTF8Fields(fields map[string]string) []string {
	var invalid []string
	for name, value := range fields {
		if !utf8.ValidString(value) {
			invalid = append(invalid, fmt.Sprintf("%s (len=%d)", name, len(value)))
		}
	}
	return invalid
}

func sanitizeEndpoint(endpoint *models.VersionEndpoint) {
	message := strings.ToValidUTF8(endpoint.Message, "")
	if len(message) > maxMessageChar {
		message = message[:maxMessageChar]
		message = strings.ToValidUTF8(message, "")
	}
	endpoint.Message = message

	// Status only ever holds controlled constants, but sanitize defensively to
	// guarantee a valid UTF-8 byte sequence is persisted (avoids SQLSTATE 22021).
	endpoint.Status = models.EndpointStatus(strings.ToValidUTF8(string(endpoint.Status), ""))
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

func (v *versionEndpointStorage) Delete(endpoint *models.VersionEndpoint) error {
	return v.db.Delete(endpoint).Error
}
