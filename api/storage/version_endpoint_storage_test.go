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

// +build integration_local integration

package storage

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"

	"github.com/gojek/merlin/it/database"
	"github.com/gojek/merlin/mlp"
	"github.com/gojek/merlin/models"
)

func TestVersionEndpointsStorage_Get(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		endpoints := populateVersionEndpointTable(db)
		endpointSvc := NewVersionEndpointStorage(db)

		actualEndpoint, err := endpointSvc.Get(endpoints[0].ID)

		assert.NoError(t, err)
		assert.NotNil(t, actualEndpoint)
		assert.NotNil(t, actualEndpoint.Environment)
		assert.Equal(t, actualEndpoint.ID, endpoints[0].ID)
	})
}

func TestVersionEndpointsStorage_ListEndpoints(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		endpoints := populateVersionEndpointTable(db)

		endpointSvc := NewVersionEndpointStorage(db)

		actualEndpoints, err := endpointSvc.ListEndpoints(&models.Model{ID: 1}, &models.Version{ID: 1})
		assert.NoError(t, err)
		assert.Equal(t, len(endpoints), len(actualEndpoints))
	})
}

func TestVersionsService_CountEndpoints(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		populateVersionEndpointTable(db)

		endpointSvc := NewVersionEndpointStorage(db)

		count, err := endpointSvc.CountEndpoints(&models.Environment{
			Name: "env1",
		}, &models.Model{
			ID:   1,
			Name: "model",
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, count)
	})
}

func TestVersionEndpointsStorage_GetTransformer(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		endpoints := populateVersionEndpointTable(db)
		endpointSvc := NewVersionEndpointStorage(db)

		actualEndpoint, err := endpointSvc.Get(endpoints[2].ID)

		assert.NoError(t, err)
		assert.NotNil(t, actualEndpoint)
		assert.NotNil(t, actualEndpoint.Transformer)
		assert.Equal(t, actualEndpoint.Transformer.Image, endpoints[2].Transformer.Image)
	})
}

func populateVersionEndpointTable(db *gorm.DB) []*models.VersionEndpoint {
	isDefaultTrue := true
	p := mlp.Project{
		Name:              "project",
		MlflowTrackingURL: "http://mlflow:5000",
	}
	db.Create(&p)

	m := models.Model{
		ID:           1,
		ProjectID:    models.ID(p.ID),
		ExperimentID: 1,
		Name:         "model",
		Type:         models.ModelTypeSkLearn,
	}
	db.Create(&m)

	v := models.Version{
		ModelID:     m.ID,
		RunID:       "1",
		ArtifactURI: "gcs:/mlp/1/1",
	}
	db.Create(&v)

	env1 := models.Environment{
		Name:      "env1",
		Cluster:   "k8s",
		IsDefault: &isDefaultTrue,
	}
	db.Create(&env1)

	env2 := models.Environment{
		Name:    "env2",
		Cluster: "k8s",
	}
	db.Create(&env2)

	ep1 := models.VersionEndpoint{
		ID:              uuid.New(),
		VersionID:       v.ID,
		VersionModelID:  m.ID,
		Status:          "pending",
		EnvironmentName: env1.Name,
	}
	db.Create(&ep1)
	ep2 := models.VersionEndpoint{
		ID:              uuid.New(),
		VersionID:       v.ID,
		VersionModelID:  m.ID,
		Status:          "terminated",
		EnvironmentName: env1.Name,
	}
	db.Create(&ep2)
	ep3 := models.VersionEndpoint{
		ID:              uuid.New(),
		VersionID:       v.ID,
		VersionModelID:  m.ID,
		Status:          "pending",
		EnvironmentName: env2.Name,
		Transformer: &models.Transformer{
			Enabled: true,
			Image:   "ghcr.io/gojek/merlin-transformer-test",
		},
	}
	db.Create(&ep3)
	return []*models.VersionEndpoint{&ep1, &ep2, &ep3}
}
