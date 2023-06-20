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

package service

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/caraml-dev/merlin/pkg/deployment"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"github.com/caraml-dev/merlin/config"
	"github.com/caraml-dev/merlin/database"
	"github.com/caraml-dev/merlin/mlp"
	mlpMock "github.com/caraml-dev/merlin/mlp/mocks"
	"github.com/caraml-dev/merlin/models"
)

func TestVersionsService_FindById(t *testing.T) {
	isDefaultTrue := true
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		env := models.Environment{
			Name:      "env1",
			Cluster:   "k8s",
			IsDefault: &isDefaultTrue,
		}
		db.Create(&env)

		p := mlp.Project{
			ID:                1,
			Name:              "project_1",
			MLFlowTrackingURL: "http://mlflow:5000",
		}

		m := models.Model{
			ProjectID:    models.ID(p.ID),
			ExperimentID: 1,
			Name:         "model_1",
			Type:         "other",
		}
		db.Create(&m)

		v := models.Version{
			ModelID:       m.ID,
			RunID:         "1",
			ArtifactURI:   "gcs:/mlp/1/1",
			PythonVersion: "3.7.*",
		}
		db.Create(&v)

		endpoint1 := &models.VersionEndpoint{
			ID:              uuid.New(),
			VersionID:       v.ID,
			VersionModelID:  m.ID,
			Status:          models.EndpointTerminated,
			EnvironmentName: env.Name,
			DeploymentMode:  deployment.ServerlessDeploymentMode,
		}
		db.Create(&endpoint1)

		endpoint2 := &models.VersionEndpoint{
			ID:              uuid.New(),
			VersionID:       v.ID,
			VersionModelID:  m.ID,
			Status:          models.EndpointTerminated,
			EnvironmentName: env.Name,
			DeploymentMode:  deployment.ServerlessDeploymentMode,
		}
		db.Create(&endpoint2)

		endpoint3 := &models.VersionEndpoint{
			ID:              uuid.New(),
			VersionID:       v.ID,
			VersionModelID:  m.ID,
			Status:          models.EndpointTerminated,
			EnvironmentName: env.Name,
			DeploymentMode:  deployment.ServerlessDeploymentMode,
			Transformer: &models.Transformer{
				Enabled: true,
				Image:   "ghcr.io/caraml-dev/merlin-transformer-test",
			},
		}
		db.Create(&endpoint3)

		mockMlpAPIClient := &mlpMock.APIClient{}
		mockMlpAPIClient.On("GetProjectByID", mock.Anything, int32(m.ProjectID)).Return(p, nil)

		versionsService := NewVersionsService(db, mockMlpAPIClient)

		found, err := versionsService.FindByID(context.Background(), m.ID, v.ID, config.MonitoringConfig{MonitoringEnabled: true})
		assert.NoError(t, err)

		assert.NotNil(t, found)
		assert.Equal(t, "http://mlflow:5000/#/experiments/1/runs/1", found.MlflowURL)
		assert.Equal(t, 3, len(found.Endpoints))

		var transformer *models.Transformer
		for _, endpoint := range found.Endpoints {
			if endpoint.Transformer != nil {
				transformer = endpoint.Transformer
			}
		}
		assert.Equal(t, endpoint3.Transformer.Image, transformer.Image)
	})
}

func TestVersionsService_ListVersions(t *testing.T) {
	isDefaultTrue := true
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		env := models.Environment{
			Name:      "env1",
			Cluster:   "k8s",
			IsDefault: &isDefaultTrue,
		}
		db.Create(&env)

		p := mlp.Project{
			ID:                1,
			Name:              "project_1",
			MLFlowTrackingURL: "http://mlflow:5000",
		}

		m := models.Model{
			ProjectID:    models.ID(p.ID),
			ExperimentID: 1,
			Name:         "model_1",
			Type:         "other",
		}
		db.Create(&m)

		numOfVersions := 10
		indexEndpointActive := map[int]interface{}{
			0: nil, 1: nil, 2: nil, 4: nil, 8: nil, 9: nil,
		}
		for i := 0; i < numOfVersions; i++ {
			v := models.Version{
				ModelID:     m.ID,
				RunID:       fmt.Sprintf("%d", i+1),
				ArtifactURI: "gcs:/mlp/1/1",
				// Version objects will have these labels test cases:
				// +----+------------+------------+------------+------------+
				// | ID | label_key1 | label_val1 | label_key2 | label_val2 |
				// +----+------------+------------+------------+------------+
				// |  1 | foo        | foo0       | bar        | bar0       |
				// |  2 | foo        | foo1       | bar        | bar1       |
				// |  3 | foo        | foo2       | bar        | bar2       |
				// |  4 | foo        | foo0       | bar        | bar3       |
				// |  5 | foo        | foo1       | bar        | bar4       |
				// |  6 | foo        | foo2       | bar        | bar0       |
				// |  7 | foo        | foo0       | bar        | bar1       |
				// |  8 | foo        | foo1       | bar        | bar2       |
				// |  9 | foo        | foo2       | bar        | bar3       |
				// | 10 | foo        | foo0       | bar        | bar4       |
				// +----+------------+------------+------------+------------+
				Labels: map[string]interface{}{
					"foo": fmt.Sprintf("foo%d", i%3),
					"bar": fmt.Sprintf("bar%d", i%5),
				},
			}
			db.Create(&v)
			endpoint1 := &models.VersionEndpoint{
				ID:              uuid.New(),
				VersionID:       v.ID,
				VersionModelID:  m.ID,
				Status:          models.EndpointTerminated,
				EnvironmentName: env.Name,
				DeploymentMode:  deployment.ServerlessDeploymentMode,
			}
			db.Create(&endpoint1)

			endpoint2 := &models.VersionEndpoint{
				ID:              uuid.New(),
				VersionID:       v.ID,
				VersionModelID:  m.ID,
				Status:          models.EndpointTerminated,
				EnvironmentName: env.Name,
				DeploymentMode:  deployment.ServerlessDeploymentMode,
			}
			db.Create(&endpoint2)

			endpoint3Status := models.EndpointTerminated

			if _, ok := indexEndpointActive[i]; ok {
				endpoint3Status = models.EndpointRunning
			}
			endpoint3 := &models.VersionEndpoint{
				ID:              uuid.New(),
				VersionID:       v.ID,
				VersionModelID:  m.ID,
				Status:          endpoint3Status,
				EnvironmentName: env.Name,
				Transformer: &models.Transformer{
					Enabled: true,
					Image:   "ghcr.io/caraml-dev/merlin-transformer-test",
				},
				DeploymentMode: deployment.ServerlessDeploymentMode,
			}
			db.Create(&endpoint3)
		}

		mockMlpAPIClient := &mlpMock.APIClient{}
		mockMlpAPIClient.On("GetProjectByID", mock.Anything, int32(m.ProjectID)).Return(p, nil)

		versionsService := NewVersionsService(db, mockMlpAPIClient)

		numOfFetchedVersions := 0
		cursor := ""
		limit := 3
		for numOfFetchedVersions < numOfVersions {
			versions, nextCursor, err := versionsService.ListVersions(context.Background(), m.ID, config.MonitoringConfig{MonitoringEnabled: true}, VersionQuery{
				PaginationQuery: PaginationQuery{
					Limit:  limit,
					Cursor: cursor,
				},
			})
			assert.NoError(t, err)

			expectedNumOfReturns := limit
			if numOfFetchedVersions+limit > numOfVersions {
				expectedNumOfReturns = numOfVersions - numOfFetchedVersions
			}
			assert.Equal(t, expectedNumOfReturns, len(versions))
			cursor = nextCursor
			numOfFetchedVersions = numOfFetchedVersions + len(versions)
		}

		//search mlflow_run_id 1
		versions, nextCursor, err := versionsService.ListVersions(context.Background(), m.ID, config.MonitoringConfig{MonitoringEnabled: true}, VersionQuery{
			PaginationQuery: PaginationQuery{
				Limit: limit,
			},
			Search: "1",
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(versions))
		assert.Equal(t, "", nextCursor)

		//search environment_name:env1
		versionsBasedOnEnv, cursor, err := versionsService.ListVersions(context.Background(), m.ID, config.MonitoringConfig{MonitoringEnabled: true}, VersionQuery{
			PaginationQuery: PaginationQuery{
				Limit: limit,
			},
			Search: "environment_name:env1",
		})
		assert.NoError(t, err)
		assert.Equal(t, 3, len(versionsBasedOnEnv))
		assert.NotEmpty(t, cursor)

		//search environment_name:env2 not exist env
		versionsBasedOnEnv1, cursor1, err := versionsService.ListVersions(context.Background(), m.ID, config.MonitoringConfig{MonitoringEnabled: true}, VersionQuery{
			PaginationQuery: PaginationQuery{
				Limit: limit,
			},
			Search: "environment_name:env2",
		})
		assert.NoError(t, err)
		assert.Equal(t, 0, len(versionsBasedOnEnv1))
		assert.Equal(t, "", cursor1)

		//search environment_name:env1 and non-matching label
		got, _, err := versionsService.ListVersions(context.Background(), m.ID, config.MonitoringConfig{MonitoringEnabled: true}, VersionQuery{
			Search: "environment_name:env1 labels:baz IN (qux)",
		})
		assert.NoError(t, err)
		assert.Equal(t, 0, len(got))

		//search environment_name:env1 and matching 1 label key with single value
		got, _, err = versionsService.ListVersions(context.Background(), m.ID, config.MonitoringConfig{MonitoringEnabled: true}, VersionQuery{
			Search: "environment_name:env1 labels:foo IN (foo0)",
		})
		wantIDs := []int{1, 10}
		gotIDs := make([]int, 0)
		for _, version := range got {
			gotIDs = append(gotIDs, int(version.ID))
		}
		sort.Ints(gotIDs)
		assert.NoError(t, err)
		assert.Equal(t, len(wantIDs), len(got))
		assert.Equal(t, wantIDs, gotIDs)

		//search environment_name:env1 and matching 1 label key with multiple values
		got, _, err = versionsService.ListVersions(context.Background(), m.ID, config.MonitoringConfig{MonitoringEnabled: true}, VersionQuery{
			Search: "environment_name:env1 labels:foo IN (foo0, foo1)",
		})
		wantIDs = []int{1, 2, 5, 10}
		gotIDs = make([]int, 0)
		for _, version := range got {
			gotIDs = append(gotIDs, int(version.ID))
		}
		sort.Ints(gotIDs)
		assert.NoError(t, err)
		assert.Equal(t, len(wantIDs), len(got))
		assert.Equal(t, wantIDs, gotIDs)

		//search environment_name:env1 and matching multiple label key with some multiple values
		got, _, err = versionsService.ListVersions(context.Background(), m.ID, config.MonitoringConfig{MonitoringEnabled: true}, VersionQuery{
			Search: "environment_name:env1 labels:foo IN (foo0), bar IN (bar1, bar4)",
		})
		wantIDs = []int{10}
		gotIDs = make([]int, 0)
		for _, version := range got {
			gotIDs = append(gotIDs, int(version.ID))
		}
		sort.Ints(gotIDs)
		assert.NoError(t, err)
		assert.Equal(t, len(wantIDs), len(got))
		assert.Equal(t, wantIDs, gotIDs)

		//search environment_name:env1 and matching multiple label key with all multiple values
		got, _, err = versionsService.ListVersions(context.Background(), m.ID, config.MonitoringConfig{MonitoringEnabled: true}, VersionQuery{
			Search: "environment_name:env1 labels:foo IN (foo0, foo1), bar IN (bar1, bar4)",
		})
		wantIDs = []int{2, 5, 10}
		gotIDs = make([]int, 0)
		for _, version := range got {
			gotIDs = append(gotIDs, int(version.ID))
		}
		sort.Ints(gotIDs)
		assert.NoError(t, err)
		assert.Equal(t, len(wantIDs), len(got))
		assert.Equal(t, wantIDs, gotIDs)

		//search non-matching label
		got, _, err = versionsService.ListVersions(context.Background(), m.ID, config.MonitoringConfig{MonitoringEnabled: true}, VersionQuery{
			Search: "labels:foo IN (foo0), megaman in (8), rockman in IN (x4,x5)",
		})
		wantIDs = make([]int, 0)
		gotIDs = make([]int, 0)
		for _, version := range got {
			gotIDs = append(gotIDs, int(version.ID))
		}
		sort.Ints(gotIDs)
		assert.NoError(t, err)
		assert.Equal(t, len(wantIDs), len(got))

		//search matching 1 label key with single value
		got, _, err = versionsService.ListVersions(context.Background(), m.ID, config.MonitoringConfig{MonitoringEnabled: true}, VersionQuery{
			Search: "labels:foo IN (foo0)",
		})
		wantIDs = []int{1, 4, 7, 10}
		gotIDs = make([]int, 0)
		for _, version := range got {
			gotIDs = append(gotIDs, int(version.ID))
		}
		sort.Ints(gotIDs)
		assert.NoError(t, err)
		assert.Equal(t, len(wantIDs), len(got))
		assert.Equal(t, wantIDs, gotIDs)

		//search matching 1 label key with multiple values
		got, _, err = versionsService.ListVersions(context.Background(), m.ID, config.MonitoringConfig{MonitoringEnabled: true}, VersionQuery{
			Search: "labels:foo IN (foo0, foo2)",
		})
		wantIDs = []int{1, 3, 4, 6, 7, 9, 10}
		gotIDs = make([]int, 0)
		for _, version := range got {
			gotIDs = append(gotIDs, int(version.ID))
		}
		sort.Ints(gotIDs)
		assert.NoError(t, err)
		assert.Equal(t, len(wantIDs), len(got))
		assert.Equal(t, wantIDs, gotIDs)

		//search matching multiple label key with some multiple values
		got, _, err = versionsService.ListVersions(context.Background(), m.ID, config.MonitoringConfig{MonitoringEnabled: true}, VersionQuery{
			Search: "labels:foo IN (foo0), bar IN (bar1, bar4)",
		})
		wantIDs = []int{7, 10}
		gotIDs = make([]int, 0)
		for _, version := range got {
			gotIDs = append(gotIDs, int(version.ID))
		}
		sort.Ints(gotIDs)
		assert.NoError(t, err)
		assert.Equal(t, len(wantIDs), len(got))
		assert.Equal(t, wantIDs, gotIDs)

		//search environment_name:env1 and matching multiple label key with all multiple values
		got, _, err = versionsService.ListVersions(context.Background(), m.ID, config.MonitoringConfig{MonitoringEnabled: true}, VersionQuery{
			Search: "labels:foo IN (foo0, foo1), bar IN (bar1, bar4)",
		})
		wantIDs = []int{2, 5, 7, 10}
		gotIDs = make([]int, 0)
		for _, version := range got {
			gotIDs = append(gotIDs, int(version.ID))
		}
		sort.Ints(gotIDs)
		assert.NoError(t, err)
		assert.Equal(t, len(wantIDs), len(got))
		assert.Equal(t, wantIDs, gotIDs)
	})
}

func TestVersionsService_Save(t *testing.T) {
	isDefaultTrue := true
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		env := models.Environment{
			Name:      "env1",
			Cluster:   "k8s",
			IsDefault: &isDefaultTrue,
		}
		db.Create(&env)

		p := mlp.Project{
			ID:                1,
			Name:              "project_1",
			MLFlowTrackingURL: "http://mlflow:5000",
		}

		m := models.Model{
			ProjectID:    models.ID(p.ID),
			ExperimentID: 1,
			Name:         "model_1",
			Type:         "other",
		}
		db.Create(&m)

		mockMlpAPIClient := &mlpMock.APIClient{}
		mockMlpAPIClient.On("GetProjectByID", mock.Anything, int32(m.ProjectID)).Return(p, nil)
		versionsService := NewVersionsService(db, mockMlpAPIClient)

		testCases := []struct {
			desc    string
			version models.Version
		}{
			{
				desc: "Should successfully persist version with labels",
				version: models.Version{
					ModelID:     m.ID,
					RunID:       "1",
					ArtifactURI: "gcs:/mlp/1/1",
					Labels: models.KV{
						"service_type":   "GO-FOOD",
						"targeting_date": "2021-02-01",
					},
				},
			},
			{
				desc: "Should successfully persist version without labels",
				version: models.Version{
					ModelID:     m.ID,
					RunID:       "1",
					ArtifactURI: "gcs:/mlp/1/1",
				},
			},
		}
		for _, tC := range testCases {
			t.Run(tC.desc, func(t *testing.T) {
				v, err := versionsService.Save(context.Background(), &tC.version, config.MonitoringConfig{MonitoringEnabled: true})
				assert.Nil(t, err)
				assert.Equal(t, tC.version.ModelID, v.ModelID)
				assert.Equal(t, tC.version.RunID, v.RunID)
				assert.Equal(t, tC.version.ArtifactURI, v.ArtifactURI)
				assert.Equal(t, tC.version.Labels, v.Labels)
			})
		}
	})
}
