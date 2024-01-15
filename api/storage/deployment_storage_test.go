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
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/caraml-dev/merlin/database"
	"github.com/caraml-dev/merlin/mlp"
	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/pkg/deployment"
)

func TestDeploymentStorage_List(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		deploymentStorage := NewDeploymentStorage(db)
		isDefaultTrue := true

		p := mlp.Project{
			Name:              "project",
			MLFlowTrackingURL: "http://mlflow:5000",
		}

		m := models.Model{
			ID:           1,
			ProjectID:    models.ID(p.ID),
			ExperimentID: 1,
			Name:         "model",
			Type:         models.ModelTypeSkLearn,
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
			Name:      "env1",
			Cluster:   "k8s",
			IsDefault: &isDefaultTrue,
		}
		db.Create(&env1)

		endpoint := models.VersionEndpoint{
			ID:              uuid.New(),
			VersionID:       v.ID,
			VersionModelID:  m.ID,
			Status:          "pending",
			EnvironmentName: env1.Name,
			DeploymentMode:  deployment.ServerlessDeploymentMode,
		}
		db.Create(&endpoint)

		deploy1 := &models.Deployment{
			ProjectID:         models.ID(p.ID),
			VersionID:         v.ID,
			VersionModelID:    m.ID,
			VersionEndpointID: endpoint.ID,
			Status:            models.EndpointServing,
			Error:             "",
			CreatedUpdated:    models.CreatedUpdated{},
		}

		deploy2 := &models.Deployment{
			ProjectID:         models.ID(p.ID),
			VersionID:         v.ID,
			VersionModelID:    m.ID,
			VersionEndpointID: endpoint.ID,
			Status:            models.EndpointServing,
			Error:             "",
			CreatedUpdated:    models.CreatedUpdated{},
		}

		_, err := deploymentStorage.Save(deploy1)
		assert.NoError(t, err)

		_, err = deploymentStorage.Save(deploy2)
		assert.NoError(t, err)

		deps, err := deploymentStorage.ListInModel(&m)
		assert.Len(t, deps, 2)
	})
}

func TestGetLatestDeployment(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		deploymentStorage := NewDeploymentStorage(db)
		isDefaultTrue := true

		p := mlp.Project{
			Name:              "project",
			MLFlowTrackingURL: "http://mlflow:5000",
		}

		m := models.Model{
			ID:           1,
			ProjectID:    models.ID(p.ID),
			ExperimentID: 1,
			Name:         "model",
			Type:         models.ModelTypeSkLearn,
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
			Name:      "env1",
			Cluster:   "k8s",
			IsDefault: &isDefaultTrue,
		}
		db.Create(&env1)

		endpoint := models.VersionEndpoint{
			ID:              uuid.New(),
			VersionID:       v.ID,
			VersionModelID:  m.ID,
			Status:          "pending",
			EnvironmentName: env1.Name,
			DeploymentMode:  deployment.ServerlessDeploymentMode,
		}
		db.Create(&endpoint)

		deploy1 := &models.Deployment{
			ProjectID:         models.ID(p.ID),
			VersionID:         v.ID,
			VersionModelID:    m.ID,
			VersionEndpointID: endpoint.ID,
			Status:            models.EndpointServing,
			Error:             "",
			CreatedUpdated: models.CreatedUpdated{
				UpdatedAt: time.Date(2022, 4, 27, 16, 33, 41, 0, time.UTC),
			},
		}

		deploy2 := &models.Deployment{
			ProjectID:         models.ID(p.ID),
			VersionID:         v.ID,
			VersionModelID:    m.ID,
			VersionEndpointID: endpoint.ID,
			Status:            models.EndpointServing,
			Error:             "",
			CreatedUpdated: models.CreatedUpdated{
				UpdatedAt: time.Date(2021, 4, 27, 16, 33, 41, 0, time.UTC),
			},
		}

		_, err := deploymentStorage.Save(deploy1)
		assert.NoError(t, err)

		_, err = deploymentStorage.Save(deploy2)
		assert.NoError(t, err)

		dep, err := deploymentStorage.GetLatestDeployment(m.ID, endpoint.VersionID)
		assert.NoError(t, err)
		assert.WithinDuration(t, deploy1.UpdatedAt, dep.UpdatedAt, 0)
	})
}

func TestDeploymentStorage_GetFirstSuccessModelVersionPerModel(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		deploymentStorage := NewDeploymentStorage(db)
		isDefaultTrue := true

		p := mlp.Project{
			Name:              "project",
			MLFlowTrackingURL: "http://mlflow:5000",
		}

		m := models.Model{
			ID:           1,
			ProjectID:    models.ID(p.ID),
			ExperimentID: 1,
			Name:         "model",
			Type:         models.ModelTypeSkLearn,
		}
		db.Create(&m)

		v1 := models.Version{
			ModelID:       m.ID,
			RunID:         "1",
			ArtifactURI:   "gcs:/mlp/1/1",
			PythonVersion: "3.7.*",
		}
		db.Create(&v1)

		v2 := models.Version{
			ModelID:       m.ID,
			RunID:         "1",
			ArtifactURI:   "gcs:/mlp/1/1",
			PythonVersion: "3.7.*",
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
			DeploymentMode:  deployment.ServerlessDeploymentMode,
		}
		db.Create(&e1)

		e2 := models.VersionEndpoint{
			ID:              uuid.New(),
			VersionID:       v2.ID,
			VersionModelID:  m.ID,
			Status:          models.EndpointServing,
			EnvironmentName: env1.Name,
			DeploymentMode:  deployment.ServerlessDeploymentMode,
		}
		db.Create(&e2)

		deploy1 := &models.Deployment{
			ProjectID:         models.ID(p.ID),
			VersionID:         v1.ID,
			VersionModelID:    m.ID,
			VersionEndpointID: e1.ID,
			Status:            models.EndpointFailed,
			Error:             "failed deployment",
			CreatedUpdated:    models.CreatedUpdated{},
		}

		deploy2 := &models.Deployment{
			ProjectID:         models.ID(p.ID),
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

func TestDeploymentStorage_OnDeploymentSuccess(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		deploymentStorage := NewDeploymentStorage(db)
		isDefaultTrue := true

		p := mlp.Project{
			Name:              "project",
			MLFlowTrackingURL: "http://mlflow:5000",
		}

		m := models.Model{
			ID:           1,
			ProjectID:    models.ID(p.ID),
			ExperimentID: 1,
			Name:         "model",
			Type:         models.ModelTypeSkLearn,
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
			Name:      "env1",
			Cluster:   "k8s",
			IsDefault: &isDefaultTrue,
		}
		db.Create(&env1)

		endpoint := models.VersionEndpoint{
			ID:              uuid.New(),
			VersionID:       v.ID,
			VersionModelID:  m.ID,
			Status:          "pending",
			EnvironmentName: env1.Name,
			DeploymentMode:  deployment.ServerlessDeploymentMode,
		}
		db.Create(&endpoint)

		deploy1 := &models.Deployment{
			ProjectID:         models.ID(p.ID),
			VersionID:         v.ID,
			VersionModelID:    m.ID,
			VersionEndpointID: endpoint.ID,
			Status:            models.EndpointTerminated,
			Error:             "",
			CreatedUpdated:    models.CreatedUpdated{},
		}
		db.Create(&deploy1)

		deploy2 := &models.Deployment{
			ProjectID:         models.ID(p.ID),
			VersionID:         v.ID,
			VersionModelID:    m.ID,
			VersionEndpointID: endpoint.ID,
			Status:            models.EndpointRunning,
			Error:             "",
			CreatedUpdated:    models.CreatedUpdated{},
		}
		db.Create(&deploy2)

		// To make sure that other deployment is not affected
		otherDeployment := &models.Deployment{
			ProjectID:         models.ID(p.ID),
			VersionID:         v.ID + 1,
			VersionModelID:    m.ID + 1,
			VersionEndpointID: uuid.New(),
			Status:            models.EndpointServing,
			Error:             "",
			CreatedUpdated:    models.CreatedUpdated{},
		}
		db.Create(&otherDeployment)

		newDeployment := &models.Deployment{
			ProjectID:         models.ID(p.ID),
			VersionID:         v.ID,
			VersionModelID:    m.ID,
			VersionEndpointID: endpoint.ID,
			Status:            models.EndpointPending,
			Error:             "",
			CreatedUpdated:    models.CreatedUpdated{},
		}
		db.Create(&newDeployment)

		newDeployment.Status = models.EndpointRunning

		err := deploymentStorage.OnDeploymentSuccess(newDeployment)
		assert.Nil(t, err)

		deps, err := deploymentStorage.ListInModel(&m)
		assert.Len(t, deps, 3)

		for i := range deps {
			if deps[i].ID == newDeployment.ID {
				assert.Equal(t, models.EndpointRunning, deps[i].Status)
			} else {
				assert.Equal(t, models.EndpointTerminated, deps[i].Status)
			}
		}

		// To make sure that other deployment is not affected
		assert.Equal(t, models.EndpointServing, otherDeployment.Status)
	})
}

func TestDeploymentStorage_Undeploy(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		deploymentStorage := NewDeploymentStorage(db)
		isDefaultTrue := true

		p := mlp.Project{
			Name:              "project",
			MLFlowTrackingURL: "http://mlflow:5000",
		}

		m := models.Model{
			ID:           1,
			ProjectID:    models.ID(p.ID),
			ExperimentID: 1,
			Name:         "model",
			Type:         models.ModelTypeSkLearn,
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
			Name:      "env1",
			Cluster:   "k8s",
			IsDefault: &isDefaultTrue,
		}
		db.Create(&env1)

		endpoint := models.VersionEndpoint{
			ID:              uuid.New(),
			VersionID:       v.ID,
			VersionModelID:  m.ID,
			Status:          "pending",
			EnvironmentName: env1.Name,
			DeploymentMode:  deployment.ServerlessDeploymentMode,
		}
		db.Create(&endpoint)

		deploy1 := &models.Deployment{
			ProjectID:         models.ID(p.ID),
			VersionID:         v.ID,
			VersionModelID:    m.ID,
			VersionEndpointID: endpoint.ID,
			Status:            models.EndpointTerminated,
			Error:             "",
			CreatedUpdated:    models.CreatedUpdated{},
		}
		db.Create(&deploy1)

		deploy2 := &models.Deployment{
			ProjectID:         models.ID(p.ID),
			VersionID:         v.ID,
			VersionModelID:    m.ID,
			VersionEndpointID: endpoint.ID,
			Status:            models.EndpointRunning,
			Error:             "",
			CreatedUpdated:    models.CreatedUpdated{},
		}
		db.Create(&deploy2)

		// To make sure that other deployment is not affected
		otherDeployment := &models.Deployment{
			ProjectID:         models.ID(p.ID),
			VersionID:         v.ID + 1,
			VersionModelID:    m.ID + 1,
			VersionEndpointID: uuid.New(),
			Status:            models.EndpointServing,
			Error:             "",
			CreatedUpdated:    models.CreatedUpdated{},
		}
		db.Create(&otherDeployment)

		err := deploymentStorage.Undeploy(m.ID.String(), v.ID.String(), endpoint.ID.String())
		assert.Nil(t, err)

		deps, err := deploymentStorage.ListInModel(&m)
		assert.Len(t, deps, 2)

		for i := range deps {
			assert.Equal(t, models.EndpointTerminated, deps[i].Status)
		}

		// To make sure that other deployment is not affected
		assert.Equal(t, models.EndpointServing, otherDeployment.Status)
	})
}
