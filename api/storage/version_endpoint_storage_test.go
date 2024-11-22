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

//go:build integration_local || integration
// +build integration_local integration

package storage

import (
	"testing"

	"github.com/caraml-dev/merlin/pkg/protocol"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/caraml-dev/merlin/database"
	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/pkg/deployment"
)

func TestVersionEndpointsStorage_Get(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		endpoints := populateVersionEndpointTable(db)
		endpointSvc := NewVersionEndpointStorage(db)

		for _, ve := range endpoints {
			actualEndpoint, err := endpointSvc.Get(ve.ID)

			assert.NoError(t, err)
			assert.NotNil(t, actualEndpoint)
			assert.NotNil(t, actualEndpoint.Environment)
			assert.Equal(t, actualEndpoint.ID, ve.ID)
			assert.Equal(t, actualEndpoint.Protocol, ve.Protocol)
		}

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

func TestVersionEndpointsStorage_GetModelObservability(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		endpoints := populateVersionEndpointTable(db)
		endpointSvc := NewVersionEndpointStorage(db)

		actualEndpoint, err := endpointSvc.Get(endpoints[2].ID)

		assert.NoError(t, err)
		assert.NotNil(t, actualEndpoint)
		modelObservability := actualEndpoint.ModelObservability
		assert.NotNil(t, modelObservability)
		assert.Equal(t, true, modelObservability.Enabled)
		expectedGroundTruthSource := &models.GroundTruthSource{
			TableURN:             "table_urn",
			EventTimestampColumn: "event_timestamp",
			DWHProject:           "dwh_project",
		}
		assert.Equal(t, expectedGroundTruthSource, modelObservability.GroundTruthSource)
		expectedGroundTruthob := &models.GroundTruthJob{
			CronSchedule:             "0 0 * * *",
			CPURequest:               "1",
			CPULimit:                 nil,
			MemoryRequest:            "1Gi",
			MemoryLimit:              nil,
			StartDayOffsetFromNow:    2,
			EndDayOffsetFromNow:      1,
			GracePeriodDay:           3,
			ServiceAccountSecretName: "service_account_secret_name",
		}
		assert.Equal(t, expectedGroundTruthob, modelObservability.GroundTruthJob)

		cpuQuantity := resource.MustParse("1")
		memoryQuantity := resource.MustParse("1Gi")
		expectedPredictionLogIngestionResourceRequest := &models.WorkerResourceRequest{
			CPURequest:    &cpuQuantity,
			MemoryRequest: &memoryQuantity,
			Replica:       1,
		}
		assert.Equal(t, expectedPredictionLogIngestionResourceRequest, modelObservability.PredictionLogIngestionResourceRequest)
	})
}

func TestVersionEndpointsStorage_Save(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		endpoints := populateVersionEndpointTable(db)
		endpointSvc := NewVersionEndpointStorage(db)

		targetEndpoint, err := endpointSvc.Get(endpoints[1].ID)
		assert.NoError(t, err)
		assert.NotNil(t, targetEndpoint)
		// Set message length more than 2048 char like the allocated in db field
		targetEndpoint.Message = "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"
		expectedMessage := "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"

		err = endpointSvc.Save(targetEndpoint)
		assert.NoError(t, err)

		targetEndpoint, err = endpointSvc.Get(targetEndpoint.ID)
		assert.Equal(t, expectedMessage, targetEndpoint.Message)

		// using short message
		targetEndpoint.Message = "abcdef"
		err = endpointSvc.Save(targetEndpoint)
		assert.NoError(t, err)

		targetEndpoint, err = endpointSvc.Get(targetEndpoint.ID)
		assert.NoError(t, err)
		assert.Equal(t, "abcdef", targetEndpoint.Message)

	})
}

func populateVersionEndpointTable(db *gorm.DB) []*models.VersionEndpoint {
	isDefaultTrue := true

	m := models.Model{
		ID:           1,
		ProjectID:    models.ID(1),
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
		DeploymentMode:  deployment.ServerlessDeploymentMode,
		Protocol:        protocol.UpiV1,
	}
	db.Create(&ep1)
	ep2 := models.VersionEndpoint{
		ID:              uuid.New(),
		VersionID:       v.ID,
		VersionModelID:  m.ID,
		Status:          "terminated",
		EnvironmentName: env1.Name,
		DeploymentMode:  deployment.ServerlessDeploymentMode,
		Protocol:        protocol.HttpJson,
	}
	db.Create(&ep2)
	cpuQuantity := resource.MustParse("1")
	memoryQuantity := resource.MustParse("1Gi")
	ep3 := models.VersionEndpoint{
		ID:              uuid.New(),
		VersionID:       v.ID,
		VersionModelID:  m.ID,
		Status:          "pending",
		EnvironmentName: env2.Name,
		Transformer: &models.Transformer{
			Enabled: true,
			Image:   "ghcr.io/caraml-dev/merlin-transformer-test",
		},
		DeploymentMode: deployment.ServerlessDeploymentMode,
		ModelObservability: &models.ModelObservability{
			Enabled: true,
			GroundTruthSource: &models.GroundTruthSource{
				TableURN:             "table_urn",
				EventTimestampColumn: "event_timestamp",
				DWHProject:           "dwh_project",
			},
			GroundTruthJob: &models.GroundTruthJob{
				CronSchedule:             "0 0 * * *",
				CPURequest:               "1",
				CPULimit:                 nil,
				MemoryRequest:            "1Gi",
				MemoryLimit:              nil,
				StartDayOffsetFromNow:    2,
				EndDayOffsetFromNow:      1,
				GracePeriodDay:           3,
				ServiceAccountSecretName: "service_account_secret_name",
			},
			PredictionLogIngestionResourceRequest: &models.WorkerResourceRequest{
				CPURequest:    &cpuQuantity,
				MemoryRequest: &memoryQuantity,
				Replica:       1,
			},
		},
	}
	db.Create(&ep3)
	return []*models.VersionEndpoint{&ep1, &ep2, &ep3}
}
