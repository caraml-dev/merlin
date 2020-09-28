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

// +build integration integration_local

package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"

	"github.com/gojek/merlin/istio"
	"github.com/gojek/merlin/istio/mocks"
	"github.com/gojek/merlin/it/database"
	"github.com/gojek/merlin/mlp"
	"github.com/gojek/merlin/models"
)

var env1Name = "id-env"
var env2Name = "th-env"

func TestModelEndpointsService_FindById(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		endpoints := populateModelEndpointTable(db)

		env1Istio := &mocks.Client{}
		env2Istio := &mocks.Client{}

		controllers := map[string]istio.Client{env1Name: env1Istio, env2Name: env2Istio}
		environment := "dev"
		endpointSvc := newModelEndpointsService(controllers, db, environment)

		actualEndpoint, err := endpointSvc.FindById(context.Background(), endpoints[0].Id)

		assert.NoError(t, err)
		assert.NotNil(t, actualEndpoint)
		assert.NotNil(t, actualEndpoint.Environment)
		assert.Equal(t, actualEndpoint.Id, endpoints[0].Id)
		assert.NotNil(t, actualEndpoint.Model)
	})
}

func TestModelEndpointsService_ListEndpoints(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		endpoints := populateModelEndpointTable(db)

		env1Istio := &mocks.Client{}
		env2Istio := &mocks.Client{}

		controllers := map[string]istio.Client{env1Name: env1Istio, env2Name: env2Istio}
		environment := "dev"
		endpointSvc := newModelEndpointsService(controllers, db, environment)

		actualEndpoints, err := endpointSvc.ListModelEndpoints(context.Background(), 1)
		assert.NoError(t, err)
		assert.Equal(t, len(endpoints), len(actualEndpoints))
	})
}

func TestModelEndpointsService_ListModelEndpointsInProject(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		expectedEndpoints := populateModelEndpointTable(db)
		expectedIndoEndpoints := []*models.ModelEndpoint{}
		for _, endpoint := range expectedEndpoints {
			if endpoint.EnvironmentName == env1Name {
				expectedIndoEndpoints = append(expectedIndoEndpoints, endpoint)
			}
		}

		env1Istio := &mocks.Client{}
		env2Istio := &mocks.Client{}

		controllers := map[string]istio.Client{env1Name: env1Istio, env2Name: env2Istio}
		environment := "dev"
		endpointSvc := newModelEndpointsService(controllers, db, environment)

		allEndpoints, err := endpointSvc.ListModelEndpointsInProject(context.Background(), 1, "")
		assert.NoError(t, err)
		assert.Equal(t, len(expectedEndpoints), len(allEndpoints))

		indonesiaEndpoints, err := endpointSvc.ListModelEndpointsInProject(context.Background(), 1, "id")
		assert.NoError(t, err)
		assert.Equal(t, len(expectedIndoEndpoints), len(indonesiaEndpoints))
	})
}

func populateModelEndpointTable(db *gorm.DB) []*models.ModelEndpoint {
	db = db.LogMode(true)
	isDefaultTrue := true
	p := mlp.Project{
		Id:                1,
		Name:              "project",
		MlflowTrackingUrl: "http://mlflow:5000",
	}

	m := models.Model{
		ProjectId:    models.Id(p.Id),
		ExperimentId: 1,
		Name:         "model",
		Type:         "sklearn",
	}
	db.Create(&m)

	v := models.Version{
		ModelId:     m.Id,
		RunId:       "1",
		ArtifactUri: "gcs:/mlp/1/1",
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
		Id:              uuid.New(),
		VersionId:       v.Id,
		VersionModelId:  m.Id,
		Status:          "serving",
		EnvironmentName: env1.Name,
	}
	db.Create(&ep1)
	ep2 := models.VersionEndpoint{
		Id:              uuid.New(),
		VersionId:       v.Id,
		VersionModelId:  m.Id,
		Status:          "serving",
		EnvironmentName: env2.Name,
	}
	db.Create(&ep2)

	mep1 := models.ModelEndpoint{
		Id:      1,
		ModelId: m.Id,
		Status:  "serving",
		URL:     "localhost",
		Rule: &models.ModelEndpointRule{
			Destination: []*models.ModelEndpointRuleDestination{
				{
					VersionEndpointID: ep1.Id,
					VersionEndpoint:   &ep1,
					Weight:            100,
				},
			},
		},
		EnvironmentName: env1.Name,
	}
	db.Create(&mep1)

	mep2 := models.ModelEndpoint{
		Id:      2,
		ModelId: m.Id,
		Status:  "serving",
		URL:     "localhost",
		Rule: &models.ModelEndpointRule{
			Destination: []*models.ModelEndpointRuleDestination{
				{
					VersionEndpointID: ep2.Id,
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
