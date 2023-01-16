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

//go:build integration || integration_local
// +build integration integration_local

package storage

import (
	"context"
	"testing"

	"github.com/gojek/merlin/pkg/deployment"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"

	"github.com/gojek/merlin/it/database"
	"github.com/gojek/merlin/mlp"
	"github.com/gojek/merlin/models"
)

var env1Name = "id-env"
var env2Name = "th-env"

func TestModelEndpointsStorage_FindById(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		endpoints := populateModelEndpointTable(db)
		storage := NewModelEndpointStorage(db)

		actualEndpoint, err := storage.FindByID(context.Background(), endpoints[0].ID)

		assert.NoError(t, err)
		assert.NotNil(t, actualEndpoint)
		assert.NotNil(t, actualEndpoint.Environment)
		assert.Equal(t, actualEndpoint.ID, endpoints[0].ID)
		assert.NotNil(t, actualEndpoint.Model)
	})
}

func TestModelEndpointsStorage_ListEndpoints(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		endpoints := populateModelEndpointTable(db)
		storage := NewModelEndpointStorage(db)

		actualEndpoints, err := storage.ListModelEndpoints(context.Background(), 1)
		assert.NoError(t, err)
		assert.Equal(t, len(endpoints), len(actualEndpoints))
	})
}

func TestModelEndpointsStorage_ListModelEndpointsInProject(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		expectedEndpoints := populateModelEndpointTable(db)
		var expectedIndoEndpoints []*models.ModelEndpoint
		for _, endpoint := range expectedEndpoints {
			if endpoint.EnvironmentName == env1Name {
				expectedIndoEndpoints = append(expectedIndoEndpoints, endpoint)
			}
		}

		storage := NewModelEndpointStorage(db)
		allEndpoints, err := storage.ListModelEndpointsInProject(context.Background(), 1, "")
		assert.NoError(t, err)
		assert.Equal(t, len(expectedEndpoints), len(allEndpoints))

		indonesiaEndpoints, err := storage.ListModelEndpointsInProject(context.Background(), 1, "id")
		assert.NoError(t, err)
		assert.Equal(t, len(expectedIndoEndpoints), len(indonesiaEndpoints))
	})
}

func TestModelEndpointStorage_Save(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		storage := NewModelEndpointStorage(db)
		endpoints := populateModelEndpointTable(db)

		newVersionEndpoint := &models.VersionEndpoint{
			ID:              uuid.New(),
			VersionID:       1,
			VersionModelID:  1,
			Status:          models.EndpointRunning,
			EnvironmentName: env1Name,
			DeploymentMode:  deployment.ServerlessDeploymentMode,
		}
		db.Create(newVersionEndpoint)

		newEndpoint := endpoints[0]
		oldVeId := newEndpoint.Rule.Destination[0].VersionEndpointID
		newEndpoint.Rule.Destination[0].VersionEndpointID = newVersionEndpoint.ID
		newEndpoint.Rule.Destination[0].VersionEndpoint = newVersionEndpoint

		currentEndpoint, err := storage.FindByID(context.Background(), newEndpoint.ID)
		assert.NoError(t, err)

		err = storage.Save(context.Background(), currentEndpoint, newEndpoint)
		assert.NoError(t, err)

		actual, err := storage.FindByID(context.Background(), newEndpoint.ID)
		assert.Equal(t, newVersionEndpoint.ID, actual.Rule.Destination[0].VersionEndpointID)

		ve, err := storage.(*modelEndpointStorage).versionEndpointStorage.Get(newVersionEndpoint.ID)
		assert.NoError(t, err)
		assert.Equal(t, models.EndpointServing, ve.Status)

		oldVe, err := storage.(*modelEndpointStorage).versionEndpointStorage.Get(oldVeId)
		assert.NoError(t, err)
		assert.Equal(t, models.EndpointRunning, oldVe.Status)
	})
}

func TestModelEndpointStorage_SaveTerminated(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		storage := NewModelEndpointStorage(db)
		endpoints := populateModelEndpointTable(db)

		endpoint := endpoints[0]
		endpoint.Status = models.EndpointTerminated

		ve, err := storage.(*modelEndpointStorage).versionEndpointStorage.Get(endpoint.Rule.Destination[0].VersionEndpointID)
		assert.NoError(t, err)
		assert.Equal(t, models.EndpointServing, ve.Status)

		err = storage.Save(context.Background(), nil, endpoint)
		assert.NoError(t, err)

		ve, err = storage.(*modelEndpointStorage).versionEndpointStorage.Get(endpoint.Rule.Destination[0].VersionEndpointID)
		assert.NoError(t, err)
		assert.Equal(t, models.EndpointRunning, ve.Status)
	})
}

func populateModelEndpointTable(db *gorm.DB) []*models.ModelEndpoint {
	db = db.LogMode(true)
	isDefaultTrue := true
	p := mlp.Project{
		ID:                1,
		Name:              "project",
		MLFlowTrackingURL: "http://mlflow:5000",
	}

	m := models.Model{
		ProjectID:    models.ID(p.ID),
		ExperimentID: 1,
		Name:         "model",
		Type:         "sklearn",
	}
	db.Create(&m)

	v := models.Version{
		ModelID:       m.ID,
		RunID:         "1",
		ArtifactURI:   "gcs:/mlp/1/1",
		PythonVersion: "3.7.*",
	}
	db.Create(&v)

	env1 := models.Environment{
		Name:       env1Name,
		Cluster:    "k8s",
		IsDefault:  &isDefaultTrue,
		Region:     "id",
		GcpProject: "project",
	}
	db.Create(&env1)

	env2 := models.Environment{
		Name:       env2Name,
		Cluster:    "k8s",
		Region:     "th",
		GcpProject: "project",
	}
	db.Create(&env2)

	ep1 := models.VersionEndpoint{
		ID:              uuid.New(),
		VersionID:       v.ID,
		VersionModelID:  m.ID,
		Status:          models.EndpointServing,
		EnvironmentName: env1.Name,
		DeploymentMode:  deployment.ServerlessDeploymentMode,
	}
	db.Create(&ep1)
	ep2 := models.VersionEndpoint{
		ID:              uuid.New(),
		VersionID:       v.ID,
		VersionModelID:  m.ID,
		Status:          models.EndpointServing,
		EnvironmentName: env2.Name,
		DeploymentMode:  deployment.ServerlessDeploymentMode,
	}
	db.Create(&ep2)

	mep1 := models.ModelEndpoint{
		ID:      1,
		ModelID: m.ID,
		Status:  "serving",
		URL:     "localhost",
		Rule: &models.ModelEndpointRule{
			Destination: []*models.ModelEndpointRuleDestination{
				{
					VersionEndpointID: ep1.ID,
					VersionEndpoint:   &ep1,
					Weight:            100,
				},
			},
		},
		EnvironmentName: env1.Name,
	}
	db.Create(&mep1)

	mep2 := models.ModelEndpoint{
		ID:      2,
		ModelID: m.ID,
		Status:  "serving",
		URL:     "localhost",
		Rule: &models.ModelEndpointRule{
			Destination: []*models.ModelEndpointRuleDestination{
				{
					VersionEndpointID: ep2.ID,
					VersionEndpoint:   &ep2,
					Weight:            100,
				},
			},
		},
		EnvironmentName: env2.Name,
	}
	db.Create(&mep2)

	return []*models.ModelEndpoint{&mep1, &mep2}
}
