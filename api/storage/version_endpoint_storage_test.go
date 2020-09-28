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

		actualEndpoint, err := endpointSvc.Get(endpoints[0].Id)

		assert.NoError(t, err)
		assert.NotNil(t, actualEndpoint)
		assert.NotNil(t, actualEndpoint.Environment)
		assert.Equal(t, actualEndpoint.Id, endpoints[0].Id)
	})
}

func TestVersionEndpointsStorage_ListEndpoints(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		endpoints := populateVersionEndpointTable(db)

		endpointSvc := NewVersionEndpointStorage(db)

		actualEndpoints, err := endpointSvc.ListEndpoints(&models.Model{Id: 1}, &models.Version{Id: 1})
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
			Id:   1,
			Name: "model",
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, count)
	})
}

func populateVersionEndpointTable(db *gorm.DB) []*models.VersionEndpoint {
	isDefaultTrue := true
	p := mlp.Project{
		Name:              "project",
		MlflowTrackingUrl: "http://mlflow:5000",
	}
	db.Create(&p)

	m := models.Model{
		Id:           1,
		ProjectId:    models.Id(p.Id),
		ExperimentId: 1,
		Name:         "model",
		Type:         models.ModelTypeSkLearn,
	}
	db.Create(&m)

	v := models.Version{
		ModelId:     m.Id,
		RunId:       "1",
		ArtifactUri: "gcs:/mlp/1/1",
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
		Id:              uuid.New(),
		VersionId:       v.Id,
		VersionModelId:  m.Id,
		Status:          "pending",
		EnvironmentName: env1.Name,
	}
	db.Create(&ep1)
	ep2 := models.VersionEndpoint{
		Id:              uuid.New(),
		VersionId:       v.Id,
		VersionModelId:  m.Id,
		Status:          "terminated",
		EnvironmentName: env1.Name,
	}
	db.Create(&ep2)
	ep3 := models.VersionEndpoint{
		Id:              uuid.New(),
		VersionId:       v.Id,
		VersionModelId:  m.Id,
		Status:          "pending",
		EnvironmentName: env2.Name,
	}
	db.Create(&ep3)
	return []*models.VersionEndpoint{&ep1, &ep2, &ep3}
}
