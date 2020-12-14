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

func TestDeploymentStorage_List(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		deploymentStorage := NewDeploymentStorage(db)
		isDefaultTrue := true

		p := mlp.Project{
			Name:              "project",
			MlflowTrackingUrl: "http://mlflow:5000",
		}
		db.Create(&p)

		m := models.Model{
			ID:           1,
			ProjectID:    models.ID(p.Id),
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

		errRaised := models.VersionEndpoint{
			ID:              uuid.New(),
			VersionID:       v.ID,
			VersionModelID:  m.ID,
			Status:          "pending",
			EnvironmentName: env1.Name,
		}
		db.Create(&errRaised)

		deploy1 := &models.Deployment{
			ProjectID:         models.ID(p.Id),
			VersionID:         v.ID,
			VersionModelID:    m.ID,
			VersionEndpointID: errRaised.ID,
			Status:            models.EndpointServing,
			Error:             "",
			CreatedUpdated:    models.CreatedUpdated{},
		}

		deploy2 := &models.Deployment{
			ProjectID:         models.ID(p.Id),
			VersionID:         v.ID,
			VersionModelID:    m.ID,
			VersionEndpointID: errRaised.ID,
			Status:            models.EndpointServing,
			Error:             "",
			CreatedUpdated:    models.CreatedUpdated{},
		}

		_, err := deploymentStorage.Save(deploy1)
		assert.NoError(t, err)

		_, err = deploymentStorage.Save(deploy2)
		assert.NoError(t, err)

		jobs, err := deploymentStorage.ListInModel(&m)
		assert.Len(t, jobs, 2)
	})
}

func TestDeploymentStorage_GetFirstSuccessModelVersionPerModel(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		deploymentStorage := NewDeploymentStorage(db)
		isDefaultTrue := true

		p := mlp.Project{
			Name:              "project",
			MlflowTrackingUrl: "http://mlflow:5000",
		}
		db.Create(&p)

		m := models.Model{
			ID:           1,
			ProjectID:    models.ID(p.Id),
			ExperimentID: 1,
			Name:         "model",
			Type:         models.ModelTypeSkLearn,
		}
		db.Create(&m)

		v1 := models.Version{
			ModelID:     m.ID,
			RunID:       "1",
			ArtifactURI: "gcs:/mlp/1/1",
		}
		db.Create(&v1)

		v2 := models.Version{
			ModelID:     m.ID,
			RunID:       "1",
			ArtifactURI: "gcs:/mlp/1/1",
		}
		db.Create(&v2)

		env1 := models.Environment{
			Name:      "env1",
			Cluster:   "k8s",
			IsDefault: &isDefaultTrue,
		}
		db.Create(&env1)

		e1 := models.VersionEndpoint{
			ID:              uuid.New(),
			VersionID:       v1.ID,
			VersionModelID:  m.ID,
			Status:          models.EndpointFailed,
			EnvironmentName: env1.Name,
		}
		db.Create(&e1)

		e2 := models.VersionEndpoint{
			ID:              uuid.New(),
			VersionID:       v2.ID,
			VersionModelID:  m.ID,
			Status:          models.EndpointServing,
			EnvironmentName: env1.Name,
		}
		db.Create(&e2)

		deploy1 := &models.Deployment{
			ProjectID:         models.ID(p.Id),
			VersionID:         v1.ID,
			VersionModelID:    m.ID,
			VersionEndpointID: e1.ID,
			Status:            models.EndpointFailed,
			Error:             "failed deployment",
			CreatedUpdated:    models.CreatedUpdated{},
		}

		deploy2 := &models.Deployment{
			ProjectID:         models.ID(p.Id),
			VersionID:         v2.ID,
			VersionModelID:    m.ID,
			VersionEndpointID: e1.ID,
			Status:            models.EndpointServing,
			Error:             "",
			CreatedUpdated:    models.CreatedUpdated{},
		}

		_, err := deploymentStorage.Save(deploy1)
		assert.NoError(t, err)

		_, err = deploymentStorage.Save(deploy2)
		assert.NoError(t, err)

		resultMap, err := deploymentStorage.GetFirstSuccessModelVersionPerModel()
		assert.NoError(t, err)
		assert.Equal(t, v2.ID, resultMap[m.ID])
	})
}
