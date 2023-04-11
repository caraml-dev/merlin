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

package api

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/caraml-dev/merlin/config"
	"github.com/caraml-dev/merlin/mlp"
	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/pkg/deployment"
	"github.com/caraml-dev/merlin/pkg/protocol"
	"github.com/caraml-dev/merlin/pkg/transformer"
	feastmocks "github.com/caraml-dev/merlin/pkg/transformer/feast/mocks"
	"github.com/caraml-dev/merlin/service/mocks"
	"github.com/feast-dev/feast/sdk/go/protos/feast/core"
	"github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/gojek/mlp/api/client"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"k8s.io/apimachinery/pkg/api/resource"
)

func TestListEndpoint(t *testing.T) {
	uuid := uuid.New()
	testCases := []struct {
		desc            string
		vars            map[string]string
		modelService    func() *mocks.ModelsService
		versionService  func() *mocks.VersionsService
		endpointService func() *mocks.EndpointsService
		expected        *Response
	}{
		{
			desc: "Should success list endpoints",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:        models.ID(1),
					Name:      "Model 1",
					ProjectID: models.ID(1),
					Type:      "pyfunc",
					Project: mlp.Project(client.Project{
						ID:   1,
						Name: "sample",
					}),
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:          models.ID(1),
					ModelID:     models.ID(1),
					RunID:       "runID",
					MlflowURL:   "http://mlflow.com",
					ArtifactURI: "http://artifact.com",
				}, nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("ListEndpoints", context.Background(), mock.Anything, mock.Anything).Return([]*models.VersionEndpoint{
					{
						ID:             uuid,
						VersionID:      models.ID(1),
						VersionModelID: models.ID(1),
						Status:         models.EndpointServing,
						URL:            "http://endpoint-1.com",
						Environment: &models.Environment{
							ID:      models.ID(1),
							Name:    "dev",
							Cluster: "dev",
						},
					},
				}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusOK,
				data: []*models.VersionEndpoint{
					{
						ID:             uuid,
						VersionID:      models.ID(1),
						VersionModelID: models.ID(1),
						Status:         models.EndpointServing,
						URL:            "http://endpoint-1.com",
						Environment: &models.Environment{
							ID:      models.ID(1),
							Name:    "dev",
							Cluster: "dev",
						},
						MonitoringURL: "http://grafana?var-cluster=dev&var-model=Model+1&var-model_version=Model+1-1&var-project=sample",
					},
				},
			},
		},
		{
			desc: "Should return 404 if model is not found",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(nil, gorm.ErrRecordNotFound)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				return svc
			},
			expected: &Response{
				code: http.StatusNotFound,
				data: Error{Message: "Model with given `model_id: 1` not found"},
			},
		},
		{
			desc: "Should return 404 if model version is not found",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:        models.ID(1),
					Name:      "Model 1",
					ProjectID: models.ID(1),
					Type:      "pyfunc",
					Project: mlp.Project(client.Project{
						ID:   1,
						Name: "sample",
					}),
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(nil, gorm.ErrRecordNotFound)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				return svc
			},
			expected: &Response{
				code: http.StatusNotFound,
				data: Error{Message: "Version with given `version_id: 1` not found"},
			},
		},
		{
			desc: "Should return 500 if there is error when fetching version endpoint in the db",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:        models.ID(1),
					Name:      "Model 1",
					ProjectID: models.ID(1),
					Type:      "pyfunc",
					Project: mlp.Project(client.Project{
						ID:   1,
						Name: "sample",
					}),
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:          models.ID(1),
					ModelID:     models.ID(1),
					RunID:       "runID",
					MlflowURL:   "http://mlflow.com",
					ArtifactURI: "http://artifact.com",
				}, nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("ListEndpoints", context.Background(), mock.Anything, mock.Anything).Return(nil, fmt.Errorf("DB is down"))
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "DB is down"},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			modelSvc := tC.modelService()
			versionSvc := tC.versionService()
			endpointSvc := tC.endpointService()

			ctl := &EndpointsController{
				AppContext: &AppContext{
					ModelsService:    modelSvc,
					VersionsService:  versionSvc,
					EndpointsService: endpointSvc,
					MonitoringConfig: config.MonitoringConfig{
						MonitoringEnabled: true,
						MonitoringBaseURL: "http://grafana",
					},
					AlertEnabled: true,
				},
			}
			resp := ctl.ListEndpoint(&http.Request{}, tC.vars, nil)
			assert.Equal(t, tC.expected, resp)
		})
	}
}

func TestGetEndpoint(t *testing.T) {
	uuid := uuid.New()
	testCases := []struct {
		desc            string
		vars            map[string]string
		modelService    func() *mocks.ModelsService
		versionService  func() *mocks.VersionsService
		endpointService func() *mocks.EndpointsService
		expected        *Response
	}{
		{
			desc: "Should success get endpoint",
			vars: map[string]string{
				"model_id":    "1",
				"version_id":  "1",
				"endpoint_id": uuid.String(),
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", context.Background(), models.ID(1)).Return(&models.Model{
					ID:        models.ID(1),
					Name:      "Model 1",
					ProjectID: models.ID(1),
					Type:      "pyfunc",
					Project: mlp.Project(client.Project{
						ID:   1,
						Name: "sample",
					}),
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", context.Background(), models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:          models.ID(1),
					ModelID:     models.ID(1),
					RunID:       "runID",
					MlflowURL:   "http://mlflow.com",
					ArtifactURI: "http://artifact.com",
				}, nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("FindByID", context.Background(), uuid).Return(&models.VersionEndpoint{
					ID:             uuid,
					VersionID:      models.ID(1),
					VersionModelID: models.ID(1),
					Status:         models.EndpointServing,
					URL:            "http://endpoint-1.com",
					Environment: &models.Environment{
						ID:      models.ID(1),
						Name:    "dev",
						Cluster: "dev",
					},
				}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusOK,
				data: &models.VersionEndpoint{
					ID:             uuid,
					VersionID:      models.ID(1),
					VersionModelID: models.ID(1),
					Status:         models.EndpointServing,
					URL:            "http://endpoint-1.com",
					Environment: &models.Environment{
						ID:      models.ID(1),
						Name:    "dev",
						Cluster: "dev",
					},
					MonitoringURL: "http://grafana?var-cluster=dev&var-model=Model+1&var-model_version=Model+1-1&var-project=sample",
				},
			},
		},
		{
			desc: "Should return 404 if model is not found",
			vars: map[string]string{
				"model_id":    "1",
				"version_id":  "1",
				"endpoint_id": uuid.String(),
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", context.Background(), models.ID(1)).Return(nil, gorm.ErrRecordNotFound)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				return svc
			},
			expected: &Response{
				code: http.StatusNotFound,
				data: Error{Message: "Model with given `model_id: 1` not found"},
			},
		},
		{
			desc: "Should return 404 if model version is not found",
			vars: map[string]string{
				"model_id":    "1",
				"version_id":  "1",
				"endpoint_id": uuid.String(),
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", context.Background(), models.ID(1)).Return(&models.Model{
					ID:        models.ID(1),
					Name:      "Model 1",
					ProjectID: models.ID(1),
					Type:      "pyfunc",
					Project: mlp.Project(client.Project{
						ID:   1,
						Name: "sample",
					}),
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", context.Background(), models.ID(1), models.ID(1), mock.Anything).Return(nil, gorm.ErrRecordNotFound)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				return svc
			},
			expected: &Response{
				code: http.StatusNotFound,
				data: Error{Message: "Version with given `version_id: 1` not found"},
			},
		},
		{
			desc: "Should return 500 if there is error when fetching version endpoint in the db",
			vars: map[string]string{
				"model_id":    "1",
				"version_id":  "1",
				"endpoint_id": uuid.String(),
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", context.Background(), models.ID(1)).Return(&models.Model{
					ID:        models.ID(1),
					Name:      "Model 1",
					ProjectID: models.ID(1),
					Type:      "pyfunc",
					Project: mlp.Project(client.Project{
						ID:   1,
						Name: "sample",
					}),
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", context.Background(), models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:          models.ID(1),
					ModelID:     models.ID(1),
					RunID:       "runID",
					MlflowURL:   "http://mlflow.com",
					ArtifactURI: "http://artifact.com",
				}, nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("FindByID", context.Background(), uuid).Return(nil, fmt.Errorf("DB is down"))
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: fmt.Sprintf("Error while getting version endpoint with id %v", uuid)},
			},
		},
		{
			desc: "Should return 404 if there is no version endpoint in the db",
			vars: map[string]string{
				"model_id":    "1",
				"version_id":  "1",
				"endpoint_id": uuid.String(),
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", context.Background(), models.ID(1)).Return(&models.Model{
					ID:        models.ID(1),
					Name:      "Model 1",
					ProjectID: models.ID(1),
					Type:      "pyfunc",
					Project: mlp.Project(client.Project{
						ID:   1,
						Name: "sample",
					}),
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", context.Background(), models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:          models.ID(1),
					ModelID:     models.ID(1),
					RunID:       "runID",
					MlflowURL:   "http://mlflow.com",
					ArtifactURI: "http://artifact.com",
				}, nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("FindByID", context.Background(), uuid).Return(nil, gorm.ErrRecordNotFound)
				return svc
			},
			expected: &Response{
				code: http.StatusNotFound,
				data: Error{Message: fmt.Sprintf("Version endpoint with id %s not found", uuid)},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			modelSvc := tC.modelService()
			versionSvc := tC.versionService()
			endpointSvc := tC.endpointService()

			ctl := &EndpointsController{
				AppContext: &AppContext{
					ModelsService:    modelSvc,
					VersionsService:  versionSvc,
					EndpointsService: endpointSvc,
					MonitoringConfig: config.MonitoringConfig{
						MonitoringEnabled: true,
						MonitoringBaseURL: "http://grafana",
					},
					AlertEnabled: true,
				},
			}
			resp := ctl.GetEndpoint(&http.Request{}, tC.vars, nil)
			assert.Equal(t, tC.expected, resp)
		})
	}
}

func TestListContainers(t *testing.T) {
	uuid := uuid.New()
	testCases := []struct {
		desc            string
		vars            map[string]string
		modelService    func() *mocks.ModelsService
		versionService  func() *mocks.VersionsService
		endpointService func() *mocks.EndpointsService
		expected        *Response
	}{
		{
			desc: "Should success list containers",
			vars: map[string]string{
				"model_id":    "1",
				"version_id":  "1",
				"endpoint_id": uuid.String(),
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", context.Background(), models.ID(1)).Return(&models.Model{
					ID:        models.ID(1),
					Name:      "Model 1",
					ProjectID: models.ID(1),
					Type:      "pyfunc",
					Project: mlp.Project(client.Project{
						ID:   1,
						Name: "sample",
					}),
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", context.Background(), models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:          models.ID(1),
					ModelID:     models.ID(1),
					RunID:       "runID",
					MlflowURL:   "http://mlflow.com",
					ArtifactURI: "http://artifact.com",
				}, nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("ListContainers", context.Background(), mock.Anything, mock.Anything, uuid).Return([]*models.Container{
					{
						Name:              "pod-1",
						PodName:           "pod-1-1",
						Namespace:         "default",
						Cluster:           "dev",
						GcpProject:        "dev-proj",
						VersionEndpointID: uuid,
					},
				}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusOK,
				data: []*models.Container{
					{
						Name:              "pod-1",
						PodName:           "pod-1-1",
						Namespace:         "default",
						Cluster:           "dev",
						GcpProject:        "dev-proj",
						VersionEndpointID: uuid,
					},
				},
			},
		},
		{
			desc: "Should return 404 if model is not found",
			vars: map[string]string{
				"model_id":    "1",
				"version_id":  "1",
				"endpoint_id": uuid.String(),
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(nil, gorm.ErrRecordNotFound)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				return svc
			},
			expected: &Response{
				code: http.StatusNotFound,
				data: Error{Message: "Model with given `model_id: 1` not found"},
			},
		},
		{
			desc: "Should return 404 if model version is not found",
			vars: map[string]string{
				"model_id":    "1",
				"version_id":  "1",
				"endpoint_id": uuid.String(),
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:        models.ID(1),
					Name:      "Model 1",
					ProjectID: models.ID(1),
					Type:      "pyfunc",
					Project: mlp.Project(client.Project{
						ID:   1,
						Name: "sample",
					}),
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(nil, gorm.ErrRecordNotFound)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				return svc
			},
			expected: &Response{
				code: http.StatusNotFound,
				data: Error{Message: "Version with given `version_id: 1` not found"},
			},
		},
		{
			desc: "Should return 500 if there is error when fetching list of containers ",
			vars: map[string]string{
				"model_id":    "1",
				"version_id":  "1",
				"endpoint_id": uuid.String(),
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:        models.ID(1),
					Name:      "Model 1",
					ProjectID: models.ID(1),
					Type:      "pyfunc",
					Project: mlp.Project(client.Project{
						ID:   1,
						Name: "sample",
					}),
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:          models.ID(1),
					ModelID:     models.ID(1),
					RunID:       "runID",
					MlflowURL:   "http://mlflow.com",
					ArtifactURI: "http://artifact.com",
				}, nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("ListContainers", context.Background(), mock.Anything, mock.Anything, uuid).Return(nil, fmt.Errorf("DB is down"))
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: fmt.Sprintf("Error while getting container for endpoint with id %v", uuid)},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			modelSvc := tC.modelService()
			versionSvc := tC.versionService()
			endpointSvc := tC.endpointService()

			ctl := &EndpointsController{
				AppContext: &AppContext{
					ModelsService:    modelSvc,
					VersionsService:  versionSvc,
					EndpointsService: endpointSvc,
					MonitoringConfig: config.MonitoringConfig{
						MonitoringEnabled: true,
						MonitoringBaseURL: "http://grafana",
					},
					AlertEnabled: true,
				},
			}
			resp := ctl.ListContainers(&http.Request{}, tC.vars, nil)
			assert.Equal(t, tC.expected, resp)
		})
	}
}

func TestCreateEndpoint(t *testing.T) {
	uuid := uuid.New()
	trueBoolean := true
	testCases := []struct {
		desc        string
		vars        map[string]string
		requestBody *models.VersionEndpoint

		modelService    func() *mocks.ModelsService
		versionService  func() *mocks.VersionsService
		endpointService func() *mocks.EndpointsService
		envService      func() *mocks.EnvironmentService

		monitoringConfig          config.MonitoringConfig
		standardTransformerConfig config.StandardTransformerConfig

		feastCoreMock func() *feastmocks.CoreServiceClient

		expected *Response
	}{
		{
			desc: "Should success create endpoint",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.VersionEndpoint{
				ID:              uuid,
				VersionID:       models.ID(1),
				VersionModelID:  models.ID(1),
				ServiceName:     "sample",
				Namespace:       "sample",
				EnvironmentName: "dev",
				Message:         "",
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				EnvVars: models.EnvVars([]models.EnvVar{
					{
						Name:  "WORKER",
						Value: "1",
					},
				}),
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:           models.ID(1),
					Name:         "model-1",
					ProjectID:    models.ID(1),
					Project:      mlp.Project{},
					ExperimentID: 1,
					Type:         "pyfunc",
					MlflowURL:    "",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "",
						Endpoints:    nil,
					},
				}, nil)
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				svc.On("GetDefaultEnvironment").Return(&models.Environment{
					ID:         models.ID(1),
					Name:       "dev",
					Cluster:    "dev",
					IsDefault:  &trueBoolean,
					Region:     "id",
					GcpProject: "dev-proj",
					MaxCPU:     "1",
					MaxMemory:  "1Gi",
				}, nil)
				svc.On("GetEnvironment", "dev").Return(&models.Environment{
					ID:         models.ID(1),
					Name:       "dev",
					Cluster:    "dev",
					IsDefault:  &trueBoolean,
					Region:     "id",
					GcpProject: "dev-proj",
					MaxCPU:     "1",
					MaxMemory:  "1Gi",
				}, nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("CountEndpoints", context.Background(), mock.Anything, mock.Anything).Return(0, nil)
				svc.On("DeployEndpoint", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&models.VersionEndpoint{
					ID:                   uuid,
					VersionID:            models.ID(1),
					VersionModelID:       models.ID(1),
					Status:               models.EndpointRunning,
					URL:                  "http://endpoint.svc",
					ServiceName:          "sample",
					InferenceServiceName: "sample",
					Namespace:            "sample",
					MonitoringURL:        "http://monitoring.com",
					Environment: &models.Environment{
						ID:         models.ID(1),
						Name:       "dev",
						Cluster:    "dev",
						IsDefault:  &trueBoolean,
						Region:     "id",
						GcpProject: "dev-proj",
						MaxCPU:     "1",
						MaxMemory:  "1Gi",
					},
					EnvironmentName: "dev",
					Message:         "",
					ResourceRequest: nil,
					EnvVars: models.EnvVars([]models.EnvVar{
						{
							Name:  "WORKER",
							Value: "1",
						},
					}),
					CreatedUpdated: models.CreatedUpdated{},
				}, nil)
				return svc
			},
			monitoringConfig: config.MonitoringConfig{
				MonitoringEnabled: true,
				MonitoringBaseURL: "http://grafana",
			},
			feastCoreMock: func() *feastmocks.CoreServiceClient {
				return &feastmocks.CoreServiceClient{}
			},
			expected: &Response{
				code: http.StatusCreated,
				data: &models.VersionEndpoint{
					ID:                   uuid,
					VersionID:            models.ID(1),
					VersionModelID:       models.ID(1),
					Status:               models.EndpointRunning,
					URL:                  "http://endpoint.svc",
					ServiceName:          "sample",
					InferenceServiceName: "sample",
					Namespace:            "sample",
					MonitoringURL:        "http://monitoring.com",
					Environment: &models.Environment{
						ID:         models.ID(1),
						Name:       "dev",
						Cluster:    "dev",
						IsDefault:  &trueBoolean,
						Region:     "id",
						GcpProject: "dev-proj",
						MaxCPU:     "1",
						MaxMemory:  "1Gi",
					},
					EnvironmentName: "dev",
					Message:         "",
					ResourceRequest: nil,
					EnvVars: models.EnvVars([]models.EnvVar{
						{
							Name:  "WORKER",
							Value: "1",
						},
					}),
					CreatedUpdated: models.CreatedUpdated{},
				},
			},
		},
		{
			desc: "Should success create endpoint without monitoring url",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.VersionEndpoint{
				ID:              uuid,
				VersionID:       models.ID(1),
				VersionModelID:  models.ID(1),
				ServiceName:     "sample",
				Namespace:       "sample",
				EnvironmentName: "dev",
				Message:         "",
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				EnvVars: models.EnvVars([]models.EnvVar{
					{
						Name:  "WORKER",
						Value: "1",
					},
				}),
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:           models.ID(1),
					Name:         "model-1",
					ProjectID:    models.ID(1),
					Project:      mlp.Project{},
					ExperimentID: 1,
					Type:         "pyfunc",
					MlflowURL:    "",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "",
						Endpoints:    nil,
					},
				}, nil)
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				svc.On("GetDefaultEnvironment").Return(&models.Environment{
					ID:         models.ID(1),
					Name:       "dev",
					Cluster:    "dev",
					IsDefault:  &trueBoolean,
					Region:     "id",
					GcpProject: "dev-proj",
					MaxCPU:     "1",
					MaxMemory:  "1Gi",
				}, nil)
				svc.On("GetEnvironment", "dev").Return(&models.Environment{
					ID:         models.ID(1),
					Name:       "dev",
					Cluster:    "dev",
					IsDefault:  &trueBoolean,
					Region:     "id",
					GcpProject: "dev-proj",
					MaxCPU:     "1",
					MaxMemory:  "1Gi",
				}, nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("CountEndpoints", context.Background(), mock.Anything, mock.Anything).Return(0, nil)
				svc.On("DeployEndpoint", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&models.VersionEndpoint{
					ID:                   uuid,
					VersionID:            models.ID(1),
					VersionModelID:       models.ID(1),
					Status:               models.EndpointRunning,
					URL:                  "http://endpoint.svc",
					ServiceName:          "sample",
					InferenceServiceName: "sample",
					Namespace:            "sample",
					Environment: &models.Environment{
						ID:         models.ID(1),
						Name:       "dev",
						Cluster:    "dev",
						IsDefault:  &trueBoolean,
						Region:     "id",
						GcpProject: "dev-proj",
						MaxCPU:     "1",
						MaxMemory:  "1Gi",
					},
					EnvironmentName: "dev",
					Message:         "",
					ResourceRequest: nil,
					EnvVars: models.EnvVars([]models.EnvVar{
						{
							Name:  "WORKER",
							Value: "1",
						},
					}),
					CreatedUpdated: models.CreatedUpdated{},
				}, nil)
				return svc
			},
			monitoringConfig: config.MonitoringConfig{},
			feastCoreMock: func() *feastmocks.CoreServiceClient {
				return &feastmocks.CoreServiceClient{}
			},
			expected: &Response{
				code: http.StatusCreated,
				data: &models.VersionEndpoint{
					ID:                   uuid,
					VersionID:            models.ID(1),
					VersionModelID:       models.ID(1),
					Status:               models.EndpointRunning,
					URL:                  "http://endpoint.svc",
					ServiceName:          "sample",
					InferenceServiceName: "sample",
					Namespace:            "sample",
					Environment: &models.Environment{
						ID:         models.ID(1),
						Name:       "dev",
						Cluster:    "dev",
						IsDefault:  &trueBoolean,
						Region:     "id",
						GcpProject: "dev-proj",
						MaxCPU:     "1",
						MaxMemory:  "1Gi",
					},
					EnvironmentName: "dev",
					Message:         "",
					ResourceRequest: nil,
					EnvVars: models.EnvVars([]models.EnvVar{
						{
							Name:  "WORKER",
							Value: "1",
						},
					}),
					CreatedUpdated: models.CreatedUpdated{},
				},
			},
		},
		{
			desc: "Should return 400 if UPI is not supported",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.VersionEndpoint{
				ID:              uuid,
				VersionID:       models.ID(1),
				VersionModelID:  models.ID(1),
				ServiceName:     "sample",
				Namespace:       "sample",
				EnvironmentName: "dev",
				Message:         "",
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				Protocol: protocol.UpiV1,
				EnvVars: models.EnvVars([]models.EnvVar{
					{
						Name:  "WORKER",
						Value: "1",
					},
				}),
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:           models.ID(1),
					Name:         "model-1",
					ProjectID:    models.ID(1),
					Project:      mlp.Project{},
					ExperimentID: 1,
					Type:         models.ModelTypeTensorflow,
					MlflowURL:    "",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "",
						Endpoints:    nil,
					},
				}, nil)
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				svc.On("GetDefaultEnvironment").Return(&models.Environment{
					ID:         models.ID(1),
					Name:       "dev",
					Cluster:    "dev",
					IsDefault:  &trueBoolean,
					Region:     "id",
					GcpProject: "dev-proj",
					MaxCPU:     "1",
					MaxMemory:  "1Gi",
				}, nil)
				svc.On("GetEnvironment", "dev").Return(&models.Environment{
					ID:         models.ID(1),
					Name:       "dev",
					Cluster:    "dev",
					IsDefault:  &trueBoolean,
					Region:     "id",
					GcpProject: "dev-proj",
					MaxCPU:     "1",
					MaxMemory:  "1Gi",
				}, nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				return svc
			},
			monitoringConfig: config.MonitoringConfig{
				MonitoringEnabled: true,
				MonitoringBaseURL: "http://grafana",
			},
			feastCoreMock: func() *feastmocks.CoreServiceClient {
				return &feastmocks.CoreServiceClient{}
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{Message: "tensorflow model is not supported by UPI"},
			},
		},
		{
			desc: "Should return 500 if model is not found",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.VersionEndpoint{
				ID:              uuid,
				VersionID:       models.ID(1),
				VersionModelID:  models.ID(1),
				ServiceName:     "sample",
				Namespace:       "sample",
				EnvironmentName: "dev",
				Message:         "",
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				EnvVars: models.EnvVars([]models.EnvVar{
					{
						Name:  "WORKER",
						Value: "1",
					},
				}),
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(nil, gorm.ErrRecordNotFound)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				return svc
			},
			monitoringConfig: config.MonitoringConfig{
				MonitoringEnabled: true,
				MonitoringBaseURL: "http://grafana",
			},
			feastCoreMock: func() *feastmocks.CoreServiceClient {
				return &feastmocks.CoreServiceClient{}
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "model with given id: 1 not found"},
			},
		},
		{
			desc: "Should return 500 if fetching model returning error",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.VersionEndpoint{
				ID:              uuid,
				VersionID:       models.ID(1),
				VersionModelID:  models.ID(1),
				ServiceName:     "sample",
				Namespace:       "sample",
				EnvironmentName: "dev",
				Message:         "",
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				EnvVars: models.EnvVars([]models.EnvVar{
					{
						Name:  "WORKER",
						Value: "1",
					},
				}),
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(nil, fmt.Errorf("DB is down"))
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				return svc
			},
			monitoringConfig: config.MonitoringConfig{
				MonitoringEnabled: true,
				MonitoringBaseURL: "http://grafana",
			},
			feastCoreMock: func() *feastmocks.CoreServiceClient {
				return &feastmocks.CoreServiceClient{}
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "error retrieving model with id: 1"},
			},
		},
		{
			desc: "Should return 500 if default environment not found",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.VersionEndpoint{
				ID:              uuid,
				VersionID:       models.ID(1),
				VersionModelID:  models.ID(1),
				ServiceName:     "sample",
				Namespace:       "sample",
				EnvironmentName: "dev",
				Message:         "",
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				EnvVars: models.EnvVars([]models.EnvVar{
					{
						Name:  "WORKER",
						Value: "1",
					},
				}),
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:           models.ID(1),
					Name:         "model-1",
					ProjectID:    models.ID(1),
					Project:      mlp.Project{},
					ExperimentID: 1,
					Type:         "pyfunc",
					MlflowURL:    "",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "",
						Endpoints:    nil,
					},
				}, nil)
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				svc.On("GetDefaultEnvironment").Return(nil, gorm.ErrRecordNotFound)

				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				return svc
			},
			monitoringConfig: config.MonitoringConfig{
				MonitoringEnabled: true,
				MonitoringBaseURL: "http://grafana",
			},
			feastCoreMock: func() *feastmocks.CoreServiceClient {
				return &feastmocks.CoreServiceClient{}
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Unable to find default environment, specify environment target for deployment"},
			},
		},
		{
			desc: "Should return 404 if environment is not exist",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.VersionEndpoint{
				ID:              uuid,
				VersionID:       models.ID(1),
				VersionModelID:  models.ID(1),
				ServiceName:     "sample",
				Namespace:       "sample",
				EnvironmentName: "dev",
				Message:         "",
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				EnvVars: models.EnvVars([]models.EnvVar{
					{
						Name:  "WORKER",
						Value: "1",
					},
				}),
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:           models.ID(1),
					Name:         "model-1",
					ProjectID:    models.ID(1),
					Project:      mlp.Project{},
					ExperimentID: 1,
					Type:         "pyfunc",
					MlflowURL:    "",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "",
						Endpoints:    nil,
					},
				}, nil)
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				svc.On("GetDefaultEnvironment").Return(&models.Environment{
					ID:         models.ID(1),
					Name:       "dev",
					Cluster:    "dev",
					IsDefault:  &trueBoolean,
					Region:     "id",
					GcpProject: "dev-proj",
					MaxCPU:     "1",
					MaxMemory:  "1Gi",
				}, nil)
				svc.On("GetEnvironment", "dev").Return(nil, gorm.ErrRecordNotFound)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				return svc
			},
			monitoringConfig: config.MonitoringConfig{
				MonitoringEnabled: true,
				MonitoringBaseURL: "http://grafana",
			},
			feastCoreMock: func() *feastmocks.CoreServiceClient {
				return &feastmocks.CoreServiceClient{}
			},
			expected: &Response{
				code: http.StatusNotFound,
				data: Error{Message: "Environment not found: dev"},
			},
		},
		{
			desc: "Should return 400 if deployed endpoint is more than limit",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.VersionEndpoint{
				ID:              uuid,
				VersionID:       models.ID(1),
				VersionModelID:  models.ID(1),
				ServiceName:     "sample",
				Namespace:       "sample",
				EnvironmentName: "dev",
				Message:         "",
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				EnvVars: models.EnvVars([]models.EnvVar{
					{
						Name:  "WORKER",
						Value: "1",
					},
				}),
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:           models.ID(1),
					Name:         "model-1",
					ProjectID:    models.ID(1),
					Project:      mlp.Project{},
					ExperimentID: 1,
					Type:         "pyfunc",
					MlflowURL:    "",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "",
						Endpoints:    nil,
					},
				}, nil)
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				svc.On("GetDefaultEnvironment").Return(&models.Environment{
					ID:         models.ID(1),
					Name:       "dev",
					Cluster:    "dev",
					IsDefault:  &trueBoolean,
					Region:     "id",
					GcpProject: "dev-proj",
					MaxCPU:     "1",
					MaxMemory:  "1Gi",
				}, nil)
				svc.On("GetEnvironment", "dev").Return(&models.Environment{
					ID:         models.ID(1),
					Name:       "dev",
					Cluster:    "dev",
					IsDefault:  &trueBoolean,
					Region:     "id",
					GcpProject: "dev-proj",
					MaxCPU:     "1",
					MaxMemory:  "1Gi",
				}, nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("CountEndpoints", context.Background(), mock.Anything, mock.Anything).Return(5, nil)
				return svc
			},
			monitoringConfig: config.MonitoringConfig{
				MonitoringEnabled: true,
				MonitoringBaseURL: "http://grafana",
			},
			feastCoreMock: func() *feastmocks.CoreServiceClient {
				return &feastmocks.CoreServiceClient{}
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{Message: "Max deployed endpoint reached. Max: 2 Current: 5, undeploy existing endpoint before continuing"},
			},
		},
		{
			desc: "Should return 500 if failed deployed endpoint",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.VersionEndpoint{
				ID:              uuid,
				VersionID:       models.ID(1),
				VersionModelID:  models.ID(1),
				ServiceName:     "sample",
				Namespace:       "sample",
				EnvironmentName: "dev",
				Message:         "",
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				EnvVars: models.EnvVars([]models.EnvVar{
					{
						Name:  "WORKER",
						Value: "1",
					},
				}),
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:           models.ID(1),
					Name:         "model-1",
					ProjectID:    models.ID(1),
					Project:      mlp.Project{},
					ExperimentID: 1,
					Type:         "pyfunc",
					MlflowURL:    "",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "",
						Endpoints:    nil,
					},
				}, nil)
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				svc.On("GetDefaultEnvironment").Return(&models.Environment{
					ID:         models.ID(1),
					Name:       "dev",
					Cluster:    "dev",
					IsDefault:  &trueBoolean,
					Region:     "id",
					GcpProject: "dev-proj",
					MaxCPU:     "1",
					MaxMemory:  "1Gi",
				}, nil)
				svc.On("GetEnvironment", "dev").Return(&models.Environment{
					ID:         models.ID(1),
					Name:       "dev",
					Cluster:    "dev",
					IsDefault:  &trueBoolean,
					Region:     "id",
					GcpProject: "dev-proj",
					MaxCPU:     "1",
					MaxMemory:  "1Gi",
				}, nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("CountEndpoints", context.Background(), mock.Anything, mock.Anything).Return(0, nil)
				svc.On("DeployEndpoint", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("Something went wrong"))
				return svc
			},
			monitoringConfig: config.MonitoringConfig{
				MonitoringEnabled: true,
				MonitoringBaseURL: "http://grafana",
			},
			feastCoreMock: func() *feastmocks.CoreServiceClient {
				return &feastmocks.CoreServiceClient{}
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Unable to deploy model version: Something went wrong"},
			},
		},
		{
			desc: "Should success create endpoint - custom model type",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.VersionEndpoint{
				ID:              uuid,
				VersionID:       models.ID(1),
				VersionModelID:  models.ID(1),
				ServiceName:     "sample",
				Namespace:       "sample",
				EnvironmentName: "dev",
				Message:         "",
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				EnvVars: models.EnvVars([]models.EnvVar{
					{
						Name:  "WORKER",
						Value: "1",
					},
				}),
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:           models.ID(1),
					Name:         "model-1",
					ProjectID:    models.ID(1),
					Project:      mlp.Project{},
					ExperimentID: 1,
					Type:         "custom",
					MlflowURL:    "",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "custom",
						MlflowURL:    "",
						Endpoints:    nil,
					},
					CustomPredictor: &models.CustomPredictor{
						Image:   "gcr.io/custom-predictor:v0.1",
						Command: "./run.sh",
						Args:    "firstArg secondArg",
					},
				}, nil)
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				svc.On("GetDefaultEnvironment").Return(&models.Environment{
					ID:         models.ID(1),
					Name:       "dev",
					Cluster:    "dev",
					IsDefault:  &trueBoolean,
					Region:     "id",
					GcpProject: "dev-proj",
					MaxCPU:     "1",
					MaxMemory:  "1Gi",
				}, nil)
				svc.On("GetEnvironment", "dev").Return(&models.Environment{
					ID:         models.ID(1),
					Name:       "dev",
					Cluster:    "dev",
					IsDefault:  &trueBoolean,
					Region:     "id",
					GcpProject: "dev-proj",
					MaxCPU:     "1",
					MaxMemory:  "1Gi",
				}, nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("CountEndpoints", context.Background(), mock.Anything, mock.Anything).Return(0, nil)
				svc.On("DeployEndpoint", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&models.VersionEndpoint{
					ID:                   uuid,
					VersionID:            models.ID(1),
					VersionModelID:       models.ID(1),
					Status:               models.EndpointRunning,
					URL:                  "http://endpoint.svc",
					ServiceName:          "sample",
					InferenceServiceName: "sample",
					Namespace:            "sample",
					MonitoringURL:        "http://monitoring.com",
					Environment: &models.Environment{
						ID:         models.ID(1),
						Name:       "dev",
						Cluster:    "dev",
						IsDefault:  &trueBoolean,
						Region:     "id",
						GcpProject: "dev-proj",
						MaxCPU:     "1",
						MaxMemory:  "1Gi",
					},
					EnvironmentName: "dev",
					Message:         "",
					ResourceRequest: nil,
					EnvVars: models.EnvVars([]models.EnvVar{
						{
							Name:  "WORKER",
							Value: "1",
						},
					}),
					CreatedUpdated: models.CreatedUpdated{},
				}, nil)
				return svc
			},
			monitoringConfig: config.MonitoringConfig{
				MonitoringEnabled: true,
				MonitoringBaseURL: "http://grafana",
			},
			feastCoreMock: func() *feastmocks.CoreServiceClient {
				return &feastmocks.CoreServiceClient{}
			},
			expected: &Response{
				code: http.StatusCreated,
				data: &models.VersionEndpoint{
					ID:                   uuid,
					VersionID:            models.ID(1),
					VersionModelID:       models.ID(1),
					Status:               models.EndpointRunning,
					URL:                  "http://endpoint.svc",
					ServiceName:          "sample",
					InferenceServiceName: "sample",
					Namespace:            "sample",
					MonitoringURL:        "http://monitoring.com",
					Environment: &models.Environment{
						ID:         models.ID(1),
						Name:       "dev",
						Cluster:    "dev",
						IsDefault:  &trueBoolean,
						Region:     "id",
						GcpProject: "dev-proj",
						MaxCPU:     "1",
						MaxMemory:  "1Gi",
					},
					EnvironmentName: "dev",
					Message:         "",
					ResourceRequest: nil,
					EnvVars: models.EnvVars([]models.EnvVar{
						{
							Name:  "WORKER",
							Value: "1",
						},
					}),
					CreatedUpdated: models.CreatedUpdated{},
				},
			},
		},
		{
			desc: "Should failed create endpoint - custom model type, image not set",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.VersionEndpoint{
				ID:              uuid,
				VersionID:       models.ID(1),
				VersionModelID:  models.ID(1),
				ServiceName:     "sample",
				Namespace:       "sample",
				EnvironmentName: "dev",
				Message:         "",
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				EnvVars: models.EnvVars([]models.EnvVar{
					{
						Name:  "WORKER",
						Value: "1",
					},
				}),
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:           models.ID(1),
					Name:         "model-1",
					ProjectID:    models.ID(1),
					Project:      mlp.Project{},
					ExperimentID: 1,
					Type:         "custom",
					MlflowURL:    "",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "custom",
						MlflowURL:    "",
						Endpoints:    nil,
					},
					CustomPredictor: &models.CustomPredictor{
						Command: "./run.sh",
						Args:    "firstArg secondArg",
					},
				}, nil)
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				return svc
			},
			monitoringConfig: config.MonitoringConfig{
				MonitoringEnabled: true,
				MonitoringBaseURL: "http://grafana",
			},
			feastCoreMock: func() *feastmocks.CoreServiceClient {
				return &feastmocks.CoreServiceClient{}
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{Message: "custom predictor image must be set"},
			},
		},
		{
			desc: "Should success create endpoint with transformer",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.VersionEndpoint{
				ID:              uuid,
				VersionID:       models.ID(1),
				VersionModelID:  models.ID(1),
				ServiceName:     "sample",
				Namespace:       "sample",
				EnvironmentName: "dev",
				Message:         "",
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				EnvVars: models.EnvVars([]models.EnvVar{
					{
						Name:  "WORKER",
						Value: "1",
					},
				}),
				Transformer: &models.Transformer{
					Enabled:         true,
					TransformerType: models.StandardTransformerType,
					EnvVars: models.EnvVars{
						{
							Name: transformer.StandardTransformerConfigEnvName,
							Value: `{
								"transformerConfig": {
								  "preprocess": {
									"inputs": [
									  {
										"feast": [
										  {
											"tableName": "driver_feature_table",
											"project": "merlin",
											"entities": [
											  {
												"name": "merlin_test_driver_id",
												"valueType": "STRING",
												"jsonPath": "$.drivers[*].id"
											  }
											],
											"features": [
											  {
												"name": "merlin_test_driver_features:test_int32",
												"valueType": "INT32",
												"defaultValue": "-1"
											  }
											]
										  },
										  {
											"tableName": "driver_feature_table",
											"project": "merlin",
											"servingUrl": "localhost:6566",
											"entities": [
											  {
												"name": "merlin_test_driver_id",
												"valueType": "STRING",
												"jsonPath": "$.drivers[*].id"
											  }
											],
											"features": [
											  {
												"name": "merlin_test_driver_features:test_int32",
												"valueType": "INT32",
												"defaultValue": "-1"
											  }
											]
										  },
										  {
											"tableName": "driver_feature_table",
											"project": "merlin",
											"servingUrl": "localhost:6567",
											"entities": [
											  {
												"name": "merlin_test_driver_id",
												"valueType": "STRING",
												"jsonPath": "$.drivers[*].id"
											  }
											],
											"features": [
											  {
												"name": "merlin_test_driver_features:test_int32",
												"valueType": "INT32",
												"defaultValue": "-1"
											  }
											]
										  }
										]
									  }
									]
								  }
								}
							  }`,
						},
					},
				},
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:           models.ID(1),
					Name:         "model-1",
					ProjectID:    models.ID(1),
					Project:      mlp.Project{},
					ExperimentID: 1,
					Type:         "pyfunc",
					MlflowURL:    "",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "",
						Endpoints:    nil,
					},
				}, nil)
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				svc.On("GetDefaultEnvironment").Return(&models.Environment{
					ID:         models.ID(1),
					Name:       "dev",
					Cluster:    "dev",
					IsDefault:  &trueBoolean,
					Region:     "id",
					GcpProject: "dev-proj",
					MaxCPU:     "1",
					MaxMemory:  "1Gi",
				}, nil)
				svc.On("GetEnvironment", "dev").Return(&models.Environment{
					ID:         models.ID(1),
					Name:       "dev",
					Cluster:    "dev",
					IsDefault:  &trueBoolean,
					Region:     "id",
					GcpProject: "dev-proj",
					MaxCPU:     "1",
					MaxMemory:  "1Gi",
				}, nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("CountEndpoints", context.Background(), mock.Anything, mock.Anything).Return(0, nil)
				svc.On("DeployEndpoint", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&models.VersionEndpoint{
					ID:                   uuid,
					VersionID:            models.ID(1),
					VersionModelID:       models.ID(1),
					Status:               models.EndpointRunning,
					URL:                  "http://endpoint.svc",
					ServiceName:          "sample",
					InferenceServiceName: "sample",
					Namespace:            "sample",
					MonitoringURL:        "http://monitoring.com",
					Environment: &models.Environment{
						ID:         models.ID(1),
						Name:       "dev",
						Cluster:    "dev",
						IsDefault:  &trueBoolean,
						Region:     "id",
						GcpProject: "dev-proj",
						MaxCPU:     "1",
						MaxMemory:  "1Gi",
					},
					EnvironmentName: "dev",
					Message:         "",
					ResourceRequest: nil,
					EnvVars: models.EnvVars([]models.EnvVar{
						{
							Name:  "WORKER",
							Value: "1",
						},
					}),
					CreatedUpdated: models.CreatedUpdated{},
				}, nil)
				return svc
			},
			monitoringConfig: config.MonitoringConfig{
				MonitoringEnabled: true,
				MonitoringBaseURL: "http://grafana",
			},
			standardTransformerConfig: config.StandardTransformerConfig{
				FeastBigtableConfig: &config.FeastBigtableConfig{
					ServingURL: "localhost:6567",
					Project:    "gcp-project",
					Instance:   "instance",
					AppProfile: "default",
					PoolSize:   3,
				},
				FeastRedisConfig: &config.FeastRedisConfig{
					ServingURL: "localhost:6566",
					RedisAddresses: []string{
						"10.1.1.2", "10.1.1.3",
					},
					PoolSize: 5,
				},
			},
			feastCoreMock: func() *feastmocks.CoreServiceClient {
				client := &feastmocks.CoreServiceClient{}
				client.On("ListEntities", mock.Anything, mock.Anything).
					Return(&core.ListEntitiesResponse{
						Entities: []*core.Entity{
							{
								Spec: &core.EntitySpecV2{
									Name:      "merlin_test_driver_id",
									ValueType: types.ValueType_STRING,
								},
							},
						},
					}, nil)
				client.On("ListFeatures", mock.Anything, mock.Anything).
					Return(&core.ListFeaturesResponse{
						Features: map[string]*core.FeatureSpecV2{
							"merlin_test_driver_features:test_int32": {
								Name:      "test_int32",
								ValueType: types.ValueType_INT32,
							},
						},
					}, nil)
				return client
			},
			expected: &Response{
				code: http.StatusCreated,
				data: &models.VersionEndpoint{
					ID:                   uuid,
					VersionID:            models.ID(1),
					VersionModelID:       models.ID(1),
					Status:               models.EndpointRunning,
					URL:                  "http://endpoint.svc",
					ServiceName:          "sample",
					InferenceServiceName: "sample",
					Namespace:            "sample",
					MonitoringURL:        "http://monitoring.com",
					Environment: &models.Environment{
						ID:         models.ID(1),
						Name:       "dev",
						Cluster:    "dev",
						IsDefault:  &trueBoolean,
						Region:     "id",
						GcpProject: "dev-proj",
						MaxCPU:     "1",
						MaxMemory:  "1Gi",
					},
					EnvironmentName: "dev",
					Message:         "",
					ResourceRequest: nil,
					EnvVars: models.EnvVars([]models.EnvVar{
						{
							Name:  "WORKER",
							Value: "1",
						},
					}),
					CreatedUpdated: models.CreatedUpdated{},
				},
			},
		},
		{
			desc: "Should success create endpoint with transformer; upiv1 and enable prediction logging",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.VersionEndpoint{
				ID:              uuid,
				VersionID:       models.ID(1),
				VersionModelID:  models.ID(1),
				ServiceName:     "sample",
				Namespace:       "sample",
				EnvironmentName: "dev",
				Message:         "",
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				EnvVars: models.EnvVars([]models.EnvVar{
					{
						Name:  "WORKER",
						Value: "1",
					},
				}),
				Protocol: protocol.UpiV1,
				Transformer: &models.Transformer{
					Enabled:         true,
					TransformerType: models.StandardTransformerType,
					EnvVars: models.EnvVars{
						{
							Name: transformer.StandardTransformerConfigEnvName,
							Value: `{
								"transformerConfig": {
								  "preprocess": {
									"inputs": [
									  {
										"autoload": {
										  "tableNames": [
											"rawFeatures",
											"entities"
										  ]
										}
									  },
									  {
										"variables": [
										  {
											"name": "country",
											"jsonPath": "$.prediction_context[0].string_value"
										  }
										]
									  }
									]
								  },
								  "postprocess": {
									"inputs": [
									  {
										"autoload": {
										  "tableNames": [
											"prediction_result"
										  ]
										}
									  }
									],
									"transformations": [
									  {
										"tableTransformation": {
										  "inputTable": "prediction_result",
										  "outputTable": "output_table",
										  "steps": [
											{
											  "updateColumns": [
												{
												  "column": "country",
												  "expression": "country"
												}
											  ]
											}
										  ]
										}
									  }
									],
									"outputs": [
									  {
										"upiPostprocessOutput": {
										  "predictionResultTableName": "output_table"
										}
									  }
									]
								  }
								},
								"predictionLogConfig": {
								  "enable": true,
								  "rawFeaturesTable": "rawFeatures",
								  "entitiesTable": "entities"
								}
							  }`,
						},
					},
				},
				Logger: &models.Logger{
					Prediction: &models.PredictionLoggerConfig{
						Enabled:          true,
						RawFeaturesTable: "rawFeatures",
						EntitiesTable:    "entities",
					},
					Model: &models.LoggerConfig{
						Enabled: true,
						Mode:    models.LogAll,
					},
					Transformer: &models.LoggerConfig{
						Enabled: true,
						Mode:    models.LogAll,
					},
				},
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:           models.ID(1),
					Name:         "model-1",
					ProjectID:    models.ID(1),
					Project:      mlp.Project{},
					ExperimentID: 1,
					Type:         "pyfunc",
					MlflowURL:    "",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "",
						Endpoints:    nil,
					},
				}, nil)
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				svc.On("GetDefaultEnvironment").Return(&models.Environment{
					ID:         models.ID(1),
					Name:       "dev",
					Cluster:    "dev",
					IsDefault:  &trueBoolean,
					Region:     "id",
					GcpProject: "dev-proj",
					MaxCPU:     "1",
					MaxMemory:  "1Gi",
				}, nil)
				svc.On("GetEnvironment", "dev").Return(&models.Environment{
					ID:         models.ID(1),
					Name:       "dev",
					Cluster:    "dev",
					IsDefault:  &trueBoolean,
					Region:     "id",
					GcpProject: "dev-proj",
					MaxCPU:     "1",
					MaxMemory:  "1Gi",
				}, nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("CountEndpoints", context.Background(), mock.Anything, mock.Anything).Return(0, nil)
				svc.On("DeployEndpoint", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&models.VersionEndpoint{
					ID:                   uuid,
					VersionID:            models.ID(1),
					VersionModelID:       models.ID(1),
					Status:               models.EndpointRunning,
					URL:                  "http://endpoint.svc",
					ServiceName:          "sample",
					InferenceServiceName: "sample",
					Namespace:            "sample",
					MonitoringURL:        "http://monitoring.com",
					Environment: &models.Environment{
						ID:         models.ID(1),
						Name:       "dev",
						Cluster:    "dev",
						IsDefault:  &trueBoolean,
						Region:     "id",
						GcpProject: "dev-proj",
						MaxCPU:     "1",
						MaxMemory:  "1Gi",
					},
					Logger: &models.Logger{
						Prediction: &models.PredictionLoggerConfig{
							Enabled:          true,
							RawFeaturesTable: "rawFeatures",
							EntitiesTable:    "entities",
						},
					},
					EnvironmentName: "dev",
					Message:         "",
					ResourceRequest: nil,
					EnvVars: models.EnvVars([]models.EnvVar{
						{
							Name:  "WORKER",
							Value: "1",
						},
					}),
					CreatedUpdated: models.CreatedUpdated{},
				}, nil)
				return svc
			},
			monitoringConfig: config.MonitoringConfig{
				MonitoringEnabled: true,
				MonitoringBaseURL: "http://grafana",
			},
			standardTransformerConfig: config.StandardTransformerConfig{
				FeastBigtableConfig: &config.FeastBigtableConfig{
					ServingURL: "localhost:6567",
					Project:    "gcp-project",
					Instance:   "instance",
					AppProfile: "default",
					PoolSize:   3,
				},
				FeastRedisConfig: &config.FeastRedisConfig{
					ServingURL: "localhost:6566",
					RedisAddresses: []string{
						"10.1.1.2", "10.1.1.3",
					},
					PoolSize: 5,
				},
				Kafka: config.KafkaConfig{
					Topic:               "",
					Brokers:             "brokers",
					CompressionType:     "none",
					MaxMessageSizeBytes: 1048588,
					SerializationFmt:    "protobuf",
				},
			},
			feastCoreMock: func() *feastmocks.CoreServiceClient {
				client := &feastmocks.CoreServiceClient{}
				return client
			},
			expected: &Response{
				code: http.StatusCreated,
				data: &models.VersionEndpoint{
					ID:                   uuid,
					VersionID:            models.ID(1),
					VersionModelID:       models.ID(1),
					Status:               models.EndpointRunning,
					URL:                  "http://endpoint.svc",
					ServiceName:          "sample",
					InferenceServiceName: "sample",
					Namespace:            "sample",
					MonitoringURL:        "http://monitoring.com",
					Environment: &models.Environment{
						ID:         models.ID(1),
						Name:       "dev",
						Cluster:    "dev",
						IsDefault:  &trueBoolean,
						Region:     "id",
						GcpProject: "dev-proj",
						MaxCPU:     "1",
						MaxMemory:  "1Gi",
					},
					EnvironmentName: "dev",
					Message:         "",
					ResourceRequest: nil,
					Logger: &models.Logger{
						Prediction: &models.PredictionLoggerConfig{
							Enabled:          true,
							RawFeaturesTable: "rawFeatures",
							EntitiesTable:    "entities",
						},
					},
					EnvVars: models.EnvVars([]models.EnvVar{
						{
							Name:  "WORKER",
							Value: "1",
						},
					}),
					CreatedUpdated: models.CreatedUpdated{},
				},
			},
		},
		{
			desc: "Failed deploy endpoint due to raw features table is not registered",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.VersionEndpoint{
				ID:              uuid,
				VersionID:       models.ID(1),
				VersionModelID:  models.ID(1),
				ServiceName:     "sample",
				Namespace:       "sample",
				EnvironmentName: "dev",
				Message:         "",
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				EnvVars: models.EnvVars([]models.EnvVar{
					{
						Name:  "WORKER",
						Value: "1",
					},
				}),
				Protocol: protocol.UpiV1,
				Transformer: &models.Transformer{
					Enabled:         true,
					TransformerType: models.StandardTransformerType,
					EnvVars: models.EnvVars{
						{
							Name: transformer.StandardTransformerConfigEnvName,
							Value: `{
								"transformerConfig": {
								  "preprocess": {
									"inputs": [
									  {
										"autoload": {
										  "tableNames": [
											"entities"
										  ]
										}
									  },
									  {
										"variables": [
										  {
											"name": "country",
											"jsonPath": "$.prediction_context[0].string_value"
										  }
										]
									  }
									]
								  },
								  "postprocess": {
									"inputs": [
									  {
										"autoload": {
										  "tableNames": [
											"prediction_result"
										  ]
										}
									  }
									],
									"transformations": [
									  {
										"tableTransformation": {
										  "inputTable": "prediction_result",
										  "outputTable": "output_table",
										  "steps": [
											{
											  "updateColumns": [
												{
												  "column": "country",
												  "expression": "country"
												}
											  ]
											}
										  ]
										}
									  }
									],
									"outputs": [
									  {
										"upiPostprocessOutput": {
										  "predictionResultTableName": "output_table"
										}
									  }
									]
								  }
								},
								"predictionLogConfig": {
								  "enable": true,
								  "rawFeaturesTable": "rawFeatures",
								  "entitiesTable": "entities"
								}
							  }`,
						},
					},
				},
				Logger: &models.Logger{
					Prediction: &models.PredictionLoggerConfig{
						Enabled:          true,
						RawFeaturesTable: "rawFeatures",
						EntitiesTable:    "entities",
					},
					Model: &models.LoggerConfig{
						Enabled: true,
						Mode:    models.LogAll,
					},
					Transformer: &models.LoggerConfig{
						Enabled: true,
						Mode:    models.LogAll,
					},
				},
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:           models.ID(1),
					Name:         "model-1",
					ProjectID:    models.ID(1),
					Project:      mlp.Project{},
					ExperimentID: 1,
					Type:         "pyfunc",
					MlflowURL:    "",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "",
						Endpoints:    nil,
					},
				}, nil)
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				svc.On("GetDefaultEnvironment").Return(&models.Environment{
					ID:         models.ID(1),
					Name:       "dev",
					Cluster:    "dev",
					IsDefault:  &trueBoolean,
					Region:     "id",
					GcpProject: "dev-proj",
					MaxCPU:     "1",
					MaxMemory:  "1Gi",
				}, nil)
				svc.On("GetEnvironment", "dev").Return(&models.Environment{
					ID:         models.ID(1),
					Name:       "dev",
					Cluster:    "dev",
					IsDefault:  &trueBoolean,
					Region:     "id",
					GcpProject: "dev-proj",
					MaxCPU:     "1",
					MaxMemory:  "1Gi",
				}, nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("CountEndpoints", context.Background(), mock.Anything, mock.Anything).Return(0, nil)
				return svc
			},
			monitoringConfig: config.MonitoringConfig{
				MonitoringEnabled: true,
				MonitoringBaseURL: "http://grafana",
			},
			standardTransformerConfig: config.StandardTransformerConfig{
				FeastBigtableConfig: &config.FeastBigtableConfig{
					ServingURL: "localhost:6567",
					Project:    "gcp-project",
					Instance:   "instance",
					AppProfile: "default",
					PoolSize:   3,
				},
				FeastRedisConfig: &config.FeastRedisConfig{
					ServingURL: "localhost:6566",
					RedisAddresses: []string{
						"10.1.1.2", "10.1.1.3",
					},
					PoolSize: 5,
				},
				Kafka: config.KafkaConfig{
					Topic:               "",
					Brokers:             "brokers",
					CompressionType:     "none",
					MaxMessageSizeBytes: 1048588,
					SerializationFmt:    "protobuf",
				},
			},
			feastCoreMock: func() *feastmocks.CoreServiceClient {
				client := &feastmocks.CoreServiceClient{}
				return client
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{
					Message: "variable rawFeatures is not registered",
				},
			},
		},
		{
			desc: "Should success create endpoint with transformer - source set but not serving url",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.VersionEndpoint{
				ID:              uuid,
				VersionID:       models.ID(1),
				VersionModelID:  models.ID(1),
				ServiceName:     "sample",
				Namespace:       "sample",
				EnvironmentName: "dev",
				Message:         "",
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				EnvVars: models.EnvVars([]models.EnvVar{
					{
						Name:  "WORKER",
						Value: "1",
					},
				}),
				Transformer: &models.Transformer{
					Enabled:         true,
					TransformerType: models.StandardTransformerType,
					EnvVars: models.EnvVars{
						{
							Name: transformer.StandardTransformerConfigEnvName,
							Value: `{
								"transformerConfig": {
								  "preprocess": {
									"inputs": [
									  {
										"feast": [
										  {
											"tableName": "driver_feature_table",
											"project": "merlin",
											"entities": [
											  {
												"name": "merlin_test_driver_id",
												"valueType": "STRING",
												"jsonPath": "$.drivers[*].id"
											  }
											],
											"features": [
											  {
												"name": "merlin_test_driver_features:test_int32",
												"valueType": "INT32",
												"defaultValue": "-1"
											  }
											]
										  },
										  {
											"tableName": "driver_feature_table",
											"project": "merlin",
											"source": "REDIS",
											"entities": [
											  {
												"name": "merlin_test_driver_id",
												"valueType": "STRING",
												"jsonPath": "$.drivers[*].id"
											  }
											],
											"features": [
											  {
												"name": "merlin_test_driver_features:test_int32",
												"valueType": "INT32",
												"defaultValue": "-1"
											  }
											]
										  },
										  {
											"tableName": "driver_feature_table",
											"project": "merlin",
											"source": "BIGTABLE",
											"entities": [
											  {
												"name": "merlin_test_driver_id",
												"valueType": "STRING",
												"jsonPath": "$.drivers[*].id"
											  }
											],
											"features": [
											  {
												"name": "merlin_test_driver_features:test_int32",
												"valueType": "INT32",
												"defaultValue": "-1"
											  }
											]
										  }
										]
									  }
									]
								  }
								}
							  }`,
						},
					},
				},
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:           models.ID(1),
					Name:         "model-1",
					ProjectID:    models.ID(1),
					Project:      mlp.Project{},
					ExperimentID: 1,
					Type:         "pyfunc",
					MlflowURL:    "",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "",
						Endpoints:    nil,
					},
				}, nil)
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				svc.On("GetDefaultEnvironment").Return(&models.Environment{
					ID:         models.ID(1),
					Name:       "dev",
					Cluster:    "dev",
					IsDefault:  &trueBoolean,
					Region:     "id",
					GcpProject: "dev-proj",
					MaxCPU:     "1",
					MaxMemory:  "1Gi",
				}, nil)
				svc.On("GetEnvironment", "dev").Return(&models.Environment{
					ID:         models.ID(1),
					Name:       "dev",
					Cluster:    "dev",
					IsDefault:  &trueBoolean,
					Region:     "id",
					GcpProject: "dev-proj",
					MaxCPU:     "1",
					MaxMemory:  "1Gi",
				}, nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("CountEndpoints", context.Background(), mock.Anything, mock.Anything).Return(0, nil)
				svc.On("DeployEndpoint", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&models.VersionEndpoint{
					ID:                   uuid,
					VersionID:            models.ID(1),
					VersionModelID:       models.ID(1),
					Status:               models.EndpointRunning,
					URL:                  "http://endpoint.svc",
					ServiceName:          "sample",
					InferenceServiceName: "sample",
					Namespace:            "sample",
					MonitoringURL:        "http://monitoring.com",
					Environment: &models.Environment{
						ID:         models.ID(1),
						Name:       "dev",
						Cluster:    "dev",
						IsDefault:  &trueBoolean,
						Region:     "id",
						GcpProject: "dev-proj",
						MaxCPU:     "1",
						MaxMemory:  "1Gi",
					},
					EnvironmentName: "dev",
					Message:         "",
					ResourceRequest: nil,
					EnvVars: models.EnvVars([]models.EnvVar{
						{
							Name:  "WORKER",
							Value: "1",
						},
					}),
					CreatedUpdated: models.CreatedUpdated{},
				}, nil)
				return svc
			},
			monitoringConfig: config.MonitoringConfig{
				MonitoringEnabled: true,
				MonitoringBaseURL: "http://grafana",
			},
			standardTransformerConfig: config.StandardTransformerConfig{
				FeastBigtableConfig: &config.FeastBigtableConfig{
					ServingURL: "localhost:6567",
					Project:    "gcp-project",
					Instance:   "instance",
					AppProfile: "default",
					PoolSize:   4,
				},
				FeastRedisConfig: &config.FeastRedisConfig{
					ServingURL: "localhost:6566",
					RedisAddresses: []string{
						"10.1.1.2", "10.1.1.3",
					},
					PoolSize: 5,
				},
			},
			feastCoreMock: func() *feastmocks.CoreServiceClient {
				client := &feastmocks.CoreServiceClient{}
				client.On("ListEntities", mock.Anything, mock.Anything).
					Return(&core.ListEntitiesResponse{
						Entities: []*core.Entity{
							{
								Spec: &core.EntitySpecV2{
									Name:      "merlin_test_driver_id",
									ValueType: types.ValueType_STRING,
								},
							},
						},
					}, nil)
				client.On("ListFeatures", mock.Anything, mock.Anything).
					Return(&core.ListFeaturesResponse{
						Features: map[string]*core.FeatureSpecV2{
							"merlin_test_driver_features:test_int32": {
								Name:      "test_int32",
								ValueType: types.ValueType_INT32,
							},
						},
					}, nil)
				return client
			},
			expected: &Response{
				code: http.StatusCreated,
				data: &models.VersionEndpoint{
					ID:                   uuid,
					VersionID:            models.ID(1),
					VersionModelID:       models.ID(1),
					Status:               models.EndpointRunning,
					URL:                  "http://endpoint.svc",
					ServiceName:          "sample",
					InferenceServiceName: "sample",
					Namespace:            "sample",
					MonitoringURL:        "http://monitoring.com",
					Environment: &models.Environment{
						ID:         models.ID(1),
						Name:       "dev",
						Cluster:    "dev",
						IsDefault:  &trueBoolean,
						Region:     "id",
						GcpProject: "dev-proj",
						MaxCPU:     "1",
						MaxMemory:  "1Gi",
					},
					EnvironmentName: "dev",
					Message:         "",
					ResourceRequest: nil,
					EnvVars: models.EnvVars([]models.EnvVar{
						{
							Name:  "WORKER",
							Value: "1",
						},
					}),
					CreatedUpdated: models.CreatedUpdated{},
				},
			},
		},
		{
			desc: "Should failed create endpoint - transformer config invalid serving url",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.VersionEndpoint{
				ID:              uuid,
				VersionID:       models.ID(1),
				VersionModelID:  models.ID(1),
				ServiceName:     "sample",
				Namespace:       "sample",
				EnvironmentName: "dev",
				Message:         "",
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				EnvVars: models.EnvVars([]models.EnvVar{
					{
						Name:  "WORKER",
						Value: "1",
					},
				}),
				Transformer: &models.Transformer{
					Enabled:         true,
					TransformerType: models.StandardTransformerType,
					EnvVars: models.EnvVars{
						{
							Name: transformer.StandardTransformerConfigEnvName,
							Value: `{
								"transformerConfig": {
								  "preprocess": {
									"inputs": [
									  {
										"feast": [
										  {
											"tableName": "driver_feature_table",
											"project": "merlin",
											"servingUrl": "localhost:6565",
											"entities": [
											  {
												"name": "merlin_test_driver_id",
												"valueType": "STRING",
												"jsonPath": "$.drivers[*].id"
											  }
											],
											"features": [
											  {
												"name": "merlin_test_driver_features:test_int32",
												"valueType": "INT32",
												"defaultValue": "-1"
											  }
											]
										  }
										]
									  }
									]
								  }
								}
							  }`,
						},
					},
				},
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:           models.ID(1),
					Name:         "model-1",
					ProjectID:    models.ID(1),
					Project:      mlp.Project{},
					ExperimentID: 1,
					Type:         "pyfunc",
					MlflowURL:    "",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "",
						Endpoints:    nil,
					},
				}, nil)
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				svc.On("GetDefaultEnvironment").Return(&models.Environment{
					ID:         models.ID(1),
					Name:       "dev",
					Cluster:    "dev",
					IsDefault:  &trueBoolean,
					Region:     "id",
					GcpProject: "dev-proj",
					MaxCPU:     "1",
					MaxMemory:  "1Gi",
				}, nil)
				svc.On("GetEnvironment", "dev").Return(&models.Environment{
					ID:         models.ID(1),
					Name:       "dev",
					Cluster:    "dev",
					IsDefault:  &trueBoolean,
					Region:     "id",
					GcpProject: "dev-proj",
					MaxCPU:     "1",
					MaxMemory:  "1Gi",
				}, nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("CountEndpoints", context.Background(), mock.Anything, mock.Anything).Return(0, nil)
				svc.On("DeployEndpoint", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&models.VersionEndpoint{
					ID:                   uuid,
					VersionID:            models.ID(1),
					VersionModelID:       models.ID(1),
					Status:               models.EndpointRunning,
					URL:                  "http://endpoint.svc",
					ServiceName:          "sample",
					InferenceServiceName: "sample",
					Namespace:            "sample",
					MonitoringURL:        "http://monitoring.com",
					Environment: &models.Environment{
						ID:         models.ID(1),
						Name:       "dev",
						Cluster:    "dev",
						IsDefault:  &trueBoolean,
						Region:     "id",
						GcpProject: "dev-proj",
						MaxCPU:     "1",
						MaxMemory:  "1Gi",
					},
					EnvironmentName: "dev",
					Message:         "",
					ResourceRequest: nil,
					EnvVars: models.EnvVars([]models.EnvVar{
						{
							Name:  "WORKER",
							Value: "1",
						},
					}),
					CreatedUpdated: models.CreatedUpdated{},
				}, nil)
				return svc
			},
			monitoringConfig: config.MonitoringConfig{
				MonitoringEnabled: true,
				MonitoringBaseURL: "http://grafana",
			},
			standardTransformerConfig: config.StandardTransformerConfig{
				FeastBigtableConfig: &config.FeastBigtableConfig{
					ServingURL: "localhost:6567",
				},
				FeastRedisConfig: &config.FeastRedisConfig{
					ServingURL: "localhost:6566",
					RedisAddresses: []string{
						"10.1.1.2", "10.1.1.3",
					},
					PoolSize: 5,
				},
			},
			feastCoreMock: func() *feastmocks.CoreServiceClient {
				return &feastmocks.CoreServiceClient{}
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{Message: "feast source configuration is not valid, servingURL: localhost:6565 source: UNKNOWN"},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			modelSvc := tC.modelService()
			versionSvc := tC.versionService()
			envSvc := tC.envService()
			endpointSvc := tC.endpointService()
			feastCoreMock := tC.feastCoreMock()

			ctl := &EndpointsController{
				AppContext: &AppContext{
					ModelsService:             modelSvc,
					VersionsService:           versionSvc,
					EnvironmentService:        envSvc,
					EndpointsService:          endpointSvc,
					MonitoringConfig:          tC.monitoringConfig,
					AlertEnabled:              true,
					StandardTransformerConfig: tC.standardTransformerConfig,
					FeastCoreClient:           feastCoreMock,
				},
			}
			resp := ctl.CreateEndpoint(&http.Request{}, tC.vars, tC.requestBody)
			assert.Equal(t, tC.expected, resp)
		})
	}
}

func TestUpdateEndpoint(t *testing.T) {
	uuid := uuid.New()
	trueBoolean := true
	testCases := []struct {
		desc            string
		vars            map[string]string
		requestBody     *models.VersionEndpoint
		modelService    func() *mocks.ModelsService
		versionService  func() *mocks.VersionsService
		endpointService func() *mocks.EndpointsService
		envService      func() *mocks.EnvironmentService
		expected        *Response
	}{
		{
			desc: "Should success update endpoint",
			vars: map[string]string{
				"model_id":    "1",
				"version_id":  "1",
				"endpoint_id": uuid.String(),
			},
			requestBody: &models.VersionEndpoint{
				ID:              uuid,
				VersionID:       models.ID(1),
				VersionModelID:  models.ID(1),
				Status:          models.EndpointRunning,
				ServiceName:     "sample",
				Namespace:       "sample",
				EnvironmentName: "dev",
				Message:         "",
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				EnvVars: models.EnvVars([]models.EnvVar{
					{
						Name:  "WORKER",
						Value: "1",
					},
				}),
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", context.Background(), models.ID(1)).Return(&models.Model{
					ID:           models.ID(1),
					Name:         "model-1",
					ProjectID:    models.ID(1),
					Project:      mlp.Project{},
					ExperimentID: 1,
					Type:         "pyfunc",
					MlflowURL:    "",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", context.Background(), models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "",
						Endpoints:    nil,
					},
				}, nil)
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				svc.On("GetEnvironment", "dev").Return(&models.Environment{
					ID:         models.ID(1),
					Name:       "dev",
					Cluster:    "dev",
					IsDefault:  &trueBoolean,
					Region:     "id",
					GcpProject: "dev-proj",
					MaxCPU:     "1",
					MaxMemory:  "1Gi",
				}, nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("FindByID", context.Background(), uuid).Return(&models.VersionEndpoint{
					ID:                   uuid,
					VersionID:            models.ID(1),
					VersionModelID:       models.ID(1),
					Status:               models.EndpointPending,
					ServiceName:          "sample",
					InferenceServiceName: "sample",
					Namespace:            "sample",
					URL:                  "http://endpoint.svc",
					MonitoringURL:        "http://monitoring.com",
					Environment: &models.Environment{
						ID:         models.ID(1),
						Name:       "dev",
						Cluster:    "dev",
						IsDefault:  &trueBoolean,
						Region:     "id",
						GcpProject: "dev-proj",
						MaxCPU:     "1",
						MaxMemory:  "1Gi",
					}, EnvironmentName: "dev",
					Message:         "",
					ResourceRequest: nil,
					EnvVars: models.EnvVars([]models.EnvVar{
						{
							Name:  "WORKER",
							Value: "1",
						},
					}),
				}, nil)
				svc.On("DeployEndpoint", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&models.VersionEndpoint{
					ID:                   uuid,
					VersionID:            models.ID(1),
					VersionModelID:       models.ID(1),
					Status:               models.EndpointRunning,
					URL:                  "http://endpoint.svc",
					ServiceName:          "sample",
					InferenceServiceName: "sample",
					Namespace:            "sample",
					MonitoringURL:        "http://monitoring.com",
					Environment: &models.Environment{
						ID:         models.ID(1),
						Name:       "dev",
						Cluster:    "dev",
						IsDefault:  &trueBoolean,
						Region:     "id",
						GcpProject: "dev-proj",
						MaxCPU:     "1",
						MaxMemory:  "1Gi",
					},
					EnvironmentName: "dev",
					Message:         "",
					ResourceRequest: nil,
					EnvVars: models.EnvVars([]models.EnvVar{
						{
							Name:  "WORKER",
							Value: "1",
						},
					}),
					CreatedUpdated: models.CreatedUpdated{},
				}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusOK,
				data: &models.VersionEndpoint{
					ID:                   uuid,
					VersionID:            models.ID(1),
					VersionModelID:       models.ID(1),
					Status:               models.EndpointRunning,
					URL:                  "http://endpoint.svc",
					ServiceName:          "sample",
					InferenceServiceName: "sample",
					Namespace:            "sample",
					MonitoringURL:        "http://monitoring.com",
					Environment: &models.Environment{
						ID:         models.ID(1),
						Name:       "dev",
						Cluster:    "dev",
						IsDefault:  &trueBoolean,
						Region:     "id",
						GcpProject: "dev-proj",
						MaxCPU:     "1",
						MaxMemory:  "1Gi",
					},
					EnvironmentName: "dev",
					Message:         "",
					ResourceRequest: nil,
					EnvVars: models.EnvVars([]models.EnvVar{
						{
							Name:  "WORKER",
							Value: "1",
						},
					}),
					CreatedUpdated: models.CreatedUpdated{},
				},
			},
		},
		{
			desc: "Should return 500 if model is not found",
			vars: map[string]string{
				"model_id":    "1",
				"version_id":  "1",
				"endpoint_id": uuid.String(),
			},
			requestBody: &models.VersionEndpoint{
				ID:              uuid,
				VersionID:       models.ID(1),
				VersionModelID:  models.ID(1),
				Status:          models.EndpointRunning,
				ServiceName:     "sample",
				Namespace:       "sample",
				EnvironmentName: "dev",
				Message:         "",
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				EnvVars: models.EnvVars([]models.EnvVar{
					{
						Name:  "WORKER",
						Value: "1",
					},
				}),
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", context.Background(), models.ID(1)).Return(nil, gorm.ErrRecordNotFound)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "model with given id: 1 not found"},
			},
		},
		{
			desc: "Should 400 if endpoint already serving",
			vars: map[string]string{
				"model_id":    "1",
				"version_id":  "1",
				"endpoint_id": uuid.String(),
			},
			requestBody: &models.VersionEndpoint{
				ID:              uuid,
				VersionID:       models.ID(1),
				VersionModelID:  models.ID(1),
				Status:          models.EndpointRunning,
				ServiceName:     "sample",
				Namespace:       "sample",
				EnvironmentName: "dev",
				Message:         "",
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				EnvVars: models.EnvVars([]models.EnvVar{
					{
						Name:  "WORKER",
						Value: "1",
					},
				}),
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", context.Background(), models.ID(1)).Return(&models.Model{
					ID:           models.ID(1),
					Name:         "model-1",
					ProjectID:    models.ID(1),
					Project:      mlp.Project{},
					ExperimentID: 1,
					Type:         "pyfunc",
					MlflowURL:    "",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", context.Background(), models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "",
						Endpoints:    nil,
					},
				}, nil)
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				svc.On("GetEnvironment", "dev").Return(&models.Environment{
					ID:         models.ID(1),
					Name:       "dev",
					Cluster:    "dev",
					IsDefault:  &trueBoolean,
					Region:     "id",
					GcpProject: "dev-proj",
					MaxCPU:     "1",
					MaxMemory:  "1Gi",
				}, nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("FindByID", context.Background(), uuid).Return(&models.VersionEndpoint{
					ID:                   uuid,
					VersionID:            models.ID(1),
					VersionModelID:       models.ID(1),
					Status:               models.EndpointServing,
					ServiceName:          "sample",
					InferenceServiceName: "sample",
					Namespace:            "sample",
					URL:                  "http://endpoint.svc",
					MonitoringURL:        "http://monitoring.com",
					Environment: &models.Environment{
						ID:         models.ID(1),
						Name:       "dev",
						Cluster:    "dev",
						IsDefault:  &trueBoolean,
						Region:     "id",
						GcpProject: "dev-proj",
						MaxCPU:     "1",
						MaxMemory:  "1Gi",
					}, EnvironmentName: "dev",
					Message:         "",
					ResourceRequest: nil,
					EnvVars: models.EnvVars([]models.EnvVar{
						{
							Name:  "WORKER",
							Value: "1",
						},
					}),
				}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{Message: "Updating endpoint status to running is not allowed when the endpoint is in serving state"},
			},
		},
		{
			desc: "Should 400 if new endpoint environment is different with existing one",
			vars: map[string]string{
				"model_id":    "1",
				"version_id":  "1",
				"endpoint_id": uuid.String(),
			},
			requestBody: &models.VersionEndpoint{
				ID:              uuid,
				VersionID:       models.ID(1),
				VersionModelID:  models.ID(1),
				Status:          models.EndpointRunning,
				ServiceName:     "sample",
				Namespace:       "sample",
				EnvironmentName: "staging",
				Message:         "",
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				EnvVars: models.EnvVars([]models.EnvVar{
					{
						Name:  "WORKER",
						Value: "1",
					},
				}),
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", context.Background(), models.ID(1)).Return(&models.Model{
					ID:           models.ID(1),
					Name:         "model-1",
					ProjectID:    models.ID(1),
					Project:      mlp.Project{},
					ExperimentID: 1,
					Type:         "pyfunc",
					MlflowURL:    "",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", context.Background(), models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "",
						Endpoints:    nil,
					},
				}, nil)
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				svc.On("GetEnvironment", "dev").Return(&models.Environment{
					ID:         models.ID(1),
					Name:       "dev",
					Cluster:    "dev",
					IsDefault:  &trueBoolean,
					Region:     "id",
					GcpProject: "dev-proj",
					MaxCPU:     "1",
					MaxMemory:  "1Gi",
				}, nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("FindByID", context.Background(), uuid).Return(&models.VersionEndpoint{
					ID:                   uuid,
					VersionID:            models.ID(1),
					VersionModelID:       models.ID(1),
					Status:               models.EndpointServing,
					ServiceName:          "sample",
					InferenceServiceName: "sample",
					Namespace:            "sample",
					URL:                  "http://endpoint.svc",
					MonitoringURL:        "http://monitoring.com",
					Environment: &models.Environment{
						ID:         models.ID(1),
						Name:       "dev",
						Cluster:    "dev",
						IsDefault:  &trueBoolean,
						Region:     "id",
						GcpProject: "dev-proj",
						MaxCPU:     "1",
						MaxMemory:  "1Gi",
					}, EnvironmentName: "dev",
					Message:         "",
					ResourceRequest: nil,
					EnvVars: models.EnvVars([]models.EnvVar{
						{
							Name:  "WORKER",
							Value: "1",
						},
					}),
				}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{Message: "Updating environment is not allowed, previous: dev, new: staging"},
			},
		},
		{
			desc: "Should 400 if new endpoint status is pending",
			vars: map[string]string{
				"model_id":    "1",
				"version_id":  "1",
				"endpoint_id": uuid.String(),
			},
			requestBody: &models.VersionEndpoint{
				ID:              uuid,
				VersionID:       models.ID(1),
				VersionModelID:  models.ID(1),
				Status:          models.EndpointPending,
				ServiceName:     "sample",
				Namespace:       "sample",
				EnvironmentName: "dev",
				Message:         "",
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				EnvVars: models.EnvVars([]models.EnvVar{
					{
						Name:  "WORKER",
						Value: "1",
					},
				}),
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", context.Background(), models.ID(1)).Return(&models.Model{
					ID:           models.ID(1),
					Name:         "model-1",
					ProjectID:    models.ID(1),
					Project:      mlp.Project{},
					ExperimentID: 1,
					Type:         "pyfunc",
					MlflowURL:    "",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", context.Background(), models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "",
						Endpoints:    nil,
					},
				}, nil)
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				svc.On("GetEnvironment", "dev").Return(&models.Environment{
					ID:         models.ID(1),
					Name:       "dev",
					Cluster:    "dev",
					IsDefault:  &trueBoolean,
					Region:     "id",
					GcpProject: "dev-proj",
					MaxCPU:     "1",
					MaxMemory:  "1Gi",
				}, nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("FindByID", context.Background(), uuid).Return(&models.VersionEndpoint{
					ID:                   uuid,
					VersionID:            models.ID(1),
					VersionModelID:       models.ID(1),
					Status:               models.EndpointRunning,
					ServiceName:          "sample",
					InferenceServiceName: "sample",
					Namespace:            "sample",
					URL:                  "http://endpoint.svc",
					MonitoringURL:        "http://monitoring.com",
					Environment: &models.Environment{
						ID:         models.ID(1),
						Name:       "dev",
						Cluster:    "dev",
						IsDefault:  &trueBoolean,
						Region:     "id",
						GcpProject: "dev-proj",
						MaxCPU:     "1",
						MaxMemory:  "1Gi",
					}, EnvironmentName: "dev",
					Message:         "",
					ResourceRequest: nil,
					EnvVars: models.EnvVars([]models.EnvVar{
						{
							Name:  "WORKER",
							Value: "1",
						},
					}),
				}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{Message: "Updating endpoint status to pending is not allowed"},
			},
		},
		{
			desc: "Should return 500 if endpoint not found",
			vars: map[string]string{
				"model_id":    "1",
				"version_id":  "1",
				"endpoint_id": uuid.String(),
			},
			requestBody: &models.VersionEndpoint{
				ID:              uuid,
				VersionID:       models.ID(1),
				VersionModelID:  models.ID(1),
				Status:          models.EndpointRunning,
				ServiceName:     "sample",
				Namespace:       "sample",
				EnvironmentName: "dev",
				Message:         "",
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				EnvVars: models.EnvVars([]models.EnvVar{
					{
						Name:  "WORKER",
						Value: "1",
					},
				}),
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", context.Background(), models.ID(1)).Return(&models.Model{
					ID:           models.ID(1),
					Name:         "model-1",
					ProjectID:    models.ID(1),
					Project:      mlp.Project{},
					ExperimentID: 1,
					Type:         "pyfunc",
					MlflowURL:    "",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", context.Background(), models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "",
						Endpoints:    nil,
					},
				}, nil)
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				svc.On("GetEnvironment", "dev").Return(nil, gorm.ErrRecordNotFound)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("FindByID", context.Background(), uuid).Return(&models.VersionEndpoint{
					ID:                   uuid,
					VersionID:            models.ID(1),
					VersionModelID:       models.ID(1),
					Status:               models.EndpointPending,
					ServiceName:          "sample",
					InferenceServiceName: "sample",
					Namespace:            "sample",
					URL:                  "http://endpoint.svc",
					MonitoringURL:        "http://monitoring.com",
					Environment: &models.Environment{
						ID:         models.ID(1),
						Name:       "dev",
						Cluster:    "dev",
						IsDefault:  &trueBoolean,
						Region:     "id",
						GcpProject: "dev-proj",
						MaxCPU:     "1",
						MaxMemory:  "1Gi",
					}, EnvironmentName: "dev",
					Message:         "",
					ResourceRequest: nil,
					EnvVars: models.EnvVars([]models.EnvVar{
						{
							Name:  "WORKER",
							Value: "1",
						},
					}),
				}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusNotFound,
				data: Error{Message: "Environment not found: dev"},
			},
		},
		{
			desc: "Should success update endpoint to terminated",
			vars: map[string]string{
				"model_id":    "1",
				"version_id":  "1",
				"endpoint_id": uuid.String(),
			},
			requestBody: &models.VersionEndpoint{
				ID:              uuid,
				VersionID:       models.ID(1),
				VersionModelID:  models.ID(1),
				Status:          models.EndpointTerminated,
				ServiceName:     "sample",
				Namespace:       "sample",
				EnvironmentName: "dev",
				Message:         "",
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				EnvVars: models.EnvVars([]models.EnvVar{
					{
						Name:  "WORKER",
						Value: "1",
					},
				}),
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", context.Background(), models.ID(1)).Return(&models.Model{
					ID:           models.ID(1),
					Name:         "model-1",
					ProjectID:    models.ID(1),
					Project:      mlp.Project{},
					ExperimentID: 1,
					Type:         "pyfunc",
					MlflowURL:    "",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", context.Background(), models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "",
						Endpoints:    nil,
					},
				}, nil)
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				svc.On("GetEnvironment", "dev").Return(&models.Environment{
					ID:         models.ID(1),
					Name:       "dev",
					Cluster:    "dev",
					IsDefault:  &trueBoolean,
					Region:     "id",
					GcpProject: "dev-proj",
					MaxCPU:     "1",
					MaxMemory:  "1Gi",
				}, nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("FindByID", context.Background(), uuid).Return(&models.VersionEndpoint{
					ID:                   uuid,
					VersionID:            models.ID(1),
					VersionModelID:       models.ID(1),
					Status:               models.EndpointPending,
					ServiceName:          "sample",
					InferenceServiceName: "sample",
					Namespace:            "sample",
					URL:                  "http://endpoint.svc",
					MonitoringURL:        "http://monitoring.com",
					Environment: &models.Environment{
						ID:         models.ID(1),
						Name:       "dev",
						Cluster:    "dev",
						IsDefault:  &trueBoolean,
						Region:     "id",
						GcpProject: "dev-proj",
						MaxCPU:     "1",
						MaxMemory:  "1Gi",
					}, EnvironmentName: "dev",
					Message:         "",
					ResourceRequest: nil,
					EnvVars: models.EnvVars([]models.EnvVar{
						{
							Name:  "WORKER",
							Value: "1",
						},
					}),
				}, nil)
				svc.On("UndeployEndpoint", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&models.VersionEndpoint{
					ID:                   uuid,
					VersionID:            models.ID(1),
					VersionModelID:       models.ID(1),
					Status:               models.EndpointTerminated,
					URL:                  "http://endpoint.svc",
					ServiceName:          "sample",
					InferenceServiceName: "sample",
					Namespace:            "sample",
					MonitoringURL:        "http://monitoring.com",
					Environment: &models.Environment{
						ID:         models.ID(1),
						Name:       "dev",
						Cluster:    "dev",
						IsDefault:  &trueBoolean,
						Region:     "id",
						GcpProject: "dev-proj",
						MaxCPU:     "1",
						MaxMemory:  "1Gi",
					},
					EnvironmentName: "dev",
					Message:         "",
					ResourceRequest: nil,
					EnvVars: models.EnvVars([]models.EnvVar{
						{
							Name:  "WORKER",
							Value: "1",
						},
					}),
					CreatedUpdated: models.CreatedUpdated{},
				}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusOK,
				data: &models.VersionEndpoint{
					ID:                   uuid,
					VersionID:            models.ID(1),
					VersionModelID:       models.ID(1),
					Status:               models.EndpointTerminated,
					URL:                  "http://endpoint.svc",
					ServiceName:          "sample",
					InferenceServiceName: "sample",
					Namespace:            "sample",
					MonitoringURL:        "http://monitoring.com",
					Environment: &models.Environment{
						ID:         models.ID(1),
						Name:       "dev",
						Cluster:    "dev",
						IsDefault:  &trueBoolean,
						Region:     "id",
						GcpProject: "dev-proj",
						MaxCPU:     "1",
						MaxMemory:  "1Gi",
					},
					EnvironmentName: "dev",
					Message:         "",
					ResourceRequest: nil,
					EnvVars: models.EnvVars([]models.EnvVar{
						{
							Name:  "WORKER",
							Value: "1",
						},
					}),
					CreatedUpdated: models.CreatedUpdated{},
				},
			},
		},
		{
			desc: "Should success update endpoint - custom model type",
			vars: map[string]string{
				"model_id":    "1",
				"version_id":  "1",
				"endpoint_id": uuid.String(),
			},
			requestBody: &models.VersionEndpoint{
				ID:              uuid,
				VersionID:       models.ID(1),
				VersionModelID:  models.ID(1),
				Status:          models.EndpointRunning,
				ServiceName:     "sample",
				Namespace:       "sample",
				EnvironmentName: "dev",
				Message:         "",
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				EnvVars: models.EnvVars([]models.EnvVar{
					{
						Name:  "WORKER",
						Value: "1",
					},
				}),
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", context.Background(), models.ID(1)).Return(&models.Model{
					ID:           models.ID(1),
					Name:         "model-1",
					ProjectID:    models.ID(1),
					Project:      mlp.Project{},
					ExperimentID: 1,
					Type:         "custom",
					MlflowURL:    "",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", context.Background(), models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "custom",
						MlflowURL:    "",
						Endpoints:    nil,
					},
					CustomPredictor: &models.CustomPredictor{
						Image: "gcr.io/custom-predictor:v0.1",
					},
				}, nil)
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				svc.On("GetEnvironment", "dev").Return(&models.Environment{
					ID:         models.ID(1),
					Name:       "dev",
					Cluster:    "dev",
					IsDefault:  &trueBoolean,
					Region:     "id",
					GcpProject: "dev-proj",
					MaxCPU:     "1",
					MaxMemory:  "1Gi",
				}, nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("FindByID", context.Background(), uuid).Return(&models.VersionEndpoint{
					ID:                   uuid,
					VersionID:            models.ID(1),
					VersionModelID:       models.ID(1),
					Status:               models.EndpointPending,
					ServiceName:          "sample",
					InferenceServiceName: "sample",
					Namespace:            "sample",
					URL:                  "http://endpoint.svc",
					MonitoringURL:        "http://monitoring.com",
					Environment: &models.Environment{
						ID:         models.ID(1),
						Name:       "dev",
						Cluster:    "dev",
						IsDefault:  &trueBoolean,
						Region:     "id",
						GcpProject: "dev-proj",
						MaxCPU:     "1",
						MaxMemory:  "1Gi",
					}, EnvironmentName: "dev",
					Message:         "",
					ResourceRequest: nil,
					EnvVars: models.EnvVars([]models.EnvVar{
						{
							Name:  "WORKER",
							Value: "1",
						},
					}),
				}, nil)
				svc.On("DeployEndpoint", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&models.VersionEndpoint{
					ID:                   uuid,
					VersionID:            models.ID(1),
					VersionModelID:       models.ID(1),
					Status:               models.EndpointRunning,
					URL:                  "http://endpoint.svc",
					ServiceName:          "sample",
					InferenceServiceName: "sample",
					Namespace:            "sample",
					MonitoringURL:        "http://monitoring.com",
					Environment: &models.Environment{
						ID:         models.ID(1),
						Name:       "dev",
						Cluster:    "dev",
						IsDefault:  &trueBoolean,
						Region:     "id",
						GcpProject: "dev-proj",
						MaxCPU:     "1",
						MaxMemory:  "1Gi",
					},
					EnvironmentName: "dev",
					Message:         "",
					ResourceRequest: nil,
					EnvVars: models.EnvVars([]models.EnvVar{
						{
							Name:  "WORKER",
							Value: "1",
						},
					}),
					CreatedUpdated: models.CreatedUpdated{},
				}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusOK,
				data: &models.VersionEndpoint{
					ID:                   uuid,
					VersionID:            models.ID(1),
					VersionModelID:       models.ID(1),
					Status:               models.EndpointRunning,
					URL:                  "http://endpoint.svc",
					ServiceName:          "sample",
					InferenceServiceName: "sample",
					Namespace:            "sample",
					MonitoringURL:        "http://monitoring.com",
					Environment: &models.Environment{
						ID:         models.ID(1),
						Name:       "dev",
						Cluster:    "dev",
						IsDefault:  &trueBoolean,
						Region:     "id",
						GcpProject: "dev-proj",
						MaxCPU:     "1",
						MaxMemory:  "1Gi",
					},
					EnvironmentName: "dev",
					Message:         "",
					ResourceRequest: nil,
					EnvVars: models.EnvVars([]models.EnvVar{
						{
							Name:  "WORKER",
							Value: "1",
						},
					}),
					CreatedUpdated: models.CreatedUpdated{},
				},
			},
		},
		{
			desc: "Should failed update endpoint - custom model type image is not set",
			vars: map[string]string{
				"model_id":    "1",
				"version_id":  "1",
				"endpoint_id": uuid.String(),
			},
			requestBody: &models.VersionEndpoint{
				ID:              uuid,
				VersionID:       models.ID(1),
				VersionModelID:  models.ID(1),
				Status:          models.EndpointRunning,
				ServiceName:     "sample",
				Namespace:       "sample",
				EnvironmentName: "dev",
				Message:         "",
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				EnvVars: models.EnvVars([]models.EnvVar{
					{
						Name:  "WORKER",
						Value: "1",
					},
				}),
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", context.Background(), models.ID(1)).Return(&models.Model{
					ID:           models.ID(1),
					Name:         "model-1",
					ProjectID:    models.ID(1),
					Project:      mlp.Project{},
					ExperimentID: 1,
					Type:         "custom",
					MlflowURL:    "",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", context.Background(), models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "custom",
						MlflowURL:    "",
						Endpoints:    nil,
					},
					CustomPredictor: &models.CustomPredictor{},
				}, nil)
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				return svc
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{Message: "custom predictor image must be set"},
			},
		},
		{
			desc: "Should success update deployment mode",
			vars: map[string]string{
				"model_id":    "1",
				"version_id":  "1",
				"endpoint_id": uuid.String(),
			},
			requestBody: &models.VersionEndpoint{
				ID:              uuid,
				VersionID:       models.ID(1),
				VersionModelID:  models.ID(1),
				Status:          models.EndpointRunning,
				ServiceName:     "sample",
				Namespace:       "sample",
				EnvironmentName: "dev",
				Message:         "",
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				EnvVars: models.EnvVars([]models.EnvVar{
					{
						Name:  "WORKER",
						Value: "1",
					},
				}),
				DeploymentMode: deployment.RawDeploymentMode,
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", context.Background(), models.ID(1)).Return(&models.Model{
					ID:           models.ID(1),
					Name:         "model-1",
					ProjectID:    models.ID(1),
					Project:      mlp.Project{},
					ExperimentID: 1,
					Type:         "pyfunc",
					MlflowURL:    "",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", context.Background(), models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "",
						Endpoints:    nil,
					},
				}, nil)
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				svc.On("GetEnvironment", "dev").Return(&models.Environment{
					ID:         models.ID(1),
					Name:       "dev",
					Cluster:    "dev",
					IsDefault:  &trueBoolean,
					Region:     "id",
					GcpProject: "dev-proj",
					MaxCPU:     "1",
					MaxMemory:  "1Gi",
				}, nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("FindByID", context.Background(), uuid).Return(&models.VersionEndpoint{
					ID:                   uuid,
					VersionID:            models.ID(1),
					VersionModelID:       models.ID(1),
					Status:               models.EndpointFailed,
					ServiceName:          "sample",
					InferenceServiceName: "sample",
					Namespace:            "sample",
					URL:                  "http://endpoint.svc",
					MonitoringURL:        "http://monitoring.com",
					Environment: &models.Environment{
						ID:         models.ID(1),
						Name:       "dev",
						Cluster:    "dev",
						IsDefault:  &trueBoolean,
						Region:     "id",
						GcpProject: "dev-proj",
						MaxCPU:     "1",
						MaxMemory:  "1Gi",
					}, EnvironmentName: "dev",
					Message:         "",
					ResourceRequest: nil,
					EnvVars: models.EnvVars([]models.EnvVar{
						{
							Name:  "WORKER",
							Value: "1",
						},
					}),
					DeploymentMode: deployment.ServerlessDeploymentMode,
				}, nil)
				svc.On("DeployEndpoint", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&models.VersionEndpoint{
					ID:                   uuid,
					VersionID:            models.ID(1),
					VersionModelID:       models.ID(1),
					Status:               models.EndpointRunning,
					URL:                  "http://endpoint.svc",
					ServiceName:          "sample",
					InferenceServiceName: "sample",
					Namespace:            "sample",
					MonitoringURL:        "http://monitoring.com",
					Environment: &models.Environment{
						ID:         models.ID(1),
						Name:       "dev",
						Cluster:    "dev",
						IsDefault:  &trueBoolean,
						Region:     "id",
						GcpProject: "dev-proj",
						MaxCPU:     "1",
						MaxMemory:  "1Gi",
					},
					EnvironmentName: "dev",
					Message:         "",
					ResourceRequest: nil,
					EnvVars: models.EnvVars([]models.EnvVar{
						{
							Name:  "WORKER",
							Value: "1",
						},
					}),
					DeploymentMode: deployment.RawDeploymentMode,
					CreatedUpdated: models.CreatedUpdated{},
				}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusOK,
				data: &models.VersionEndpoint{
					ID:                   uuid,
					VersionID:            models.ID(1),
					VersionModelID:       models.ID(1),
					Status:               models.EndpointRunning,
					URL:                  "http://endpoint.svc",
					ServiceName:          "sample",
					InferenceServiceName: "sample",
					Namespace:            "sample",
					MonitoringURL:        "http://monitoring.com",
					Environment: &models.Environment{
						ID:         models.ID(1),
						Name:       "dev",
						Cluster:    "dev",
						IsDefault:  &trueBoolean,
						Region:     "id",
						GcpProject: "dev-proj",
						MaxCPU:     "1",
						MaxMemory:  "1Gi",
					},
					EnvironmentName: "dev",
					Message:         "",
					ResourceRequest: nil,
					EnvVars: models.EnvVars([]models.EnvVar{
						{
							Name:  "WORKER",
							Value: "1",
						},
					}),
					DeploymentMode: deployment.RawDeploymentMode,
					CreatedUpdated: models.CreatedUpdated{},
				},
			},
		},
		{
			desc: "Should fail to change deployment type for a serving endpoint endpoint",
			vars: map[string]string{
				"model_id":    "1",
				"version_id":  "1",
				"endpoint_id": uuid.String(),
			},
			requestBody: &models.VersionEndpoint{
				ID:              uuid,
				VersionID:       models.ID(1),
				VersionModelID:  models.ID(1),
				Status:          models.EndpointServing,
				ServiceName:     "sample",
				Namespace:       "sample",
				EnvironmentName: "dev",
				Message:         "",
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				EnvVars: models.EnvVars([]models.EnvVar{
					{
						Name:  "WORKER",
						Value: "1",
					},
				}),
				DeploymentMode: deployment.RawDeploymentMode,
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", context.Background(), models.ID(1)).Return(&models.Model{
					ID:           models.ID(1),
					Name:         "model-1",
					ProjectID:    models.ID(1),
					Project:      mlp.Project{},
					ExperimentID: 1,
					Type:         "pyfunc",
					MlflowURL:    "",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", context.Background(), models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "",
						Endpoints:    nil,
					},
				}, nil)
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				svc.On("GetEnvironment", "dev").Return(&models.Environment{
					ID:         models.ID(1),
					Name:       "dev",
					Cluster:    "dev",
					IsDefault:  &trueBoolean,
					Region:     "id",
					GcpProject: "dev-proj",
					MaxCPU:     "1",
					MaxMemory:  "1Gi",
				}, nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("FindByID", context.Background(), uuid).Return(&models.VersionEndpoint{
					ID:                   uuid,
					VersionID:            models.ID(1),
					VersionModelID:       models.ID(1),
					Status:               models.EndpointServing,
					ServiceName:          "sample",
					InferenceServiceName: "sample",
					Namespace:            "sample",
					URL:                  "http://endpoint.svc",
					MonitoringURL:        "http://monitoring.com",
					Environment: &models.Environment{
						ID:         models.ID(1),
						Name:       "dev",
						Cluster:    "dev",
						IsDefault:  &trueBoolean,
						Region:     "id",
						GcpProject: "dev-proj",
						MaxCPU:     "1",
						MaxMemory:  "1Gi",
					}, EnvironmentName: "dev",
					Message:         "",
					ResourceRequest: nil,
					EnvVars: models.EnvVars([]models.EnvVar{
						{
							Name:  "WORKER",
							Value: "1",
						},
					}),
					DeploymentMode: deployment.ServerlessDeploymentMode,
				}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{Message: "Changing deployment type of a serving model is not allowed, please terminate it first."},
			},
		},
		{
			desc: "Should fail to change deployment type for a running endpoint endpoint",
			vars: map[string]string{
				"model_id":    "1",
				"version_id":  "1",
				"endpoint_id": uuid.String(),
			},
			requestBody: &models.VersionEndpoint{
				ID:              uuid,
				VersionID:       models.ID(1),
				VersionModelID:  models.ID(1),
				Status:          models.EndpointRunning,
				ServiceName:     "sample",
				Namespace:       "sample",
				EnvironmentName: "dev",
				Message:         "",
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    1,
					MaxReplica:    4,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
				EnvVars: models.EnvVars([]models.EnvVar{
					{
						Name:  "WORKER",
						Value: "1",
					},
				}),
				DeploymentMode: deployment.RawDeploymentMode,
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", context.Background(), models.ID(1)).Return(&models.Model{
					ID:           models.ID(1),
					Name:         "model-1",
					ProjectID:    models.ID(1),
					Project:      mlp.Project{},
					ExperimentID: 1,
					Type:         "pyfunc",
					MlflowURL:    "",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", context.Background(), models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "",
						Endpoints:    nil,
					},
				}, nil)
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				svc.On("GetEnvironment", "dev").Return(&models.Environment{
					ID:         models.ID(1),
					Name:       "dev",
					Cluster:    "dev",
					IsDefault:  &trueBoolean,
					Region:     "id",
					GcpProject: "dev-proj",
					MaxCPU:     "1",
					MaxMemory:  "1Gi",
				}, nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("FindByID", context.Background(), uuid).Return(&models.VersionEndpoint{
					ID:                   uuid,
					VersionID:            models.ID(1),
					VersionModelID:       models.ID(1),
					Status:               models.EndpointRunning,
					ServiceName:          "sample",
					InferenceServiceName: "sample",
					Namespace:            "sample",
					URL:                  "http://endpoint.svc",
					MonitoringURL:        "http://monitoring.com",
					Environment: &models.Environment{
						ID:         models.ID(1),
						Name:       "dev",
						Cluster:    "dev",
						IsDefault:  &trueBoolean,
						Region:     "id",
						GcpProject: "dev-proj",
						MaxCPU:     "1",
						MaxMemory:  "1Gi",
					}, EnvironmentName: "dev",
					Message:         "",
					ResourceRequest: nil,
					EnvVars: models.EnvVars([]models.EnvVar{
						{
							Name:  "WORKER",
							Value: "1",
						},
					}),
					DeploymentMode: deployment.ServerlessDeploymentMode,
				}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{Message: "Changing deployment type of a running model is not allowed, please terminate it first."},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			modelSvc := tC.modelService()
			versionSvc := tC.versionService()
			envSvc := tC.envService()
			endpointSvc := tC.endpointService()

			ctl := &EndpointsController{
				AppContext: &AppContext{
					ModelsService:      modelSvc,
					VersionsService:    versionSvc,
					EnvironmentService: envSvc,
					EndpointsService:   endpointSvc,
					MonitoringConfig: config.MonitoringConfig{
						MonitoringEnabled: true,
						MonitoringBaseURL: "http://grafana",
					},
					AlertEnabled: true,
				},
			}
			resp := ctl.UpdateEndpoint(&http.Request{}, tC.vars, tC.requestBody)
			assert.Equal(t, tC.expected, resp)
		})
	}
}

func TestDeleteEndpoint(t *testing.T) {
	uuid := uuid.New()
	trueBoolean := true
	testCases := []struct {
		desc            string
		vars            map[string]string
		requestBody     *models.VersionEndpoint
		modelService    func() *mocks.ModelsService
		versionService  func() *mocks.VersionsService
		endpointService func() *mocks.EndpointsService
		envService      func() *mocks.EnvironmentService
		expected        *Response
	}{
		{
			desc: "Should success delete endpoint",
			vars: map[string]string{
				"model_id":    "1",
				"version_id":  "1",
				"endpoint_id": uuid.String(),
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", context.Background(), models.ID(1)).Return(&models.Model{
					ID:           models.ID(1),
					Name:         "model-1",
					ProjectID:    models.ID(1),
					Project:      mlp.Project{},
					ExperimentID: 1,
					Type:         "pyfunc",
					MlflowURL:    "",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", context.Background(), models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "",
						Endpoints:    nil,
					},
				}, nil)
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				svc.On("GetEnvironment", "dev").Return(&models.Environment{
					ID:         models.ID(1),
					Name:       "dev",
					Cluster:    "dev",
					IsDefault:  &trueBoolean,
					Region:     "id",
					GcpProject: "dev-proj",
					MaxCPU:     "1",
					MaxMemory:  "1Gi",
				}, nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("FindByID", context.Background(), uuid).Return(&models.VersionEndpoint{
					ID:                   uuid,
					VersionID:            models.ID(1),
					VersionModelID:       models.ID(1),
					Status:               models.EndpointRunning,
					ServiceName:          "sample",
					InferenceServiceName: "sample",
					Namespace:            "sample",
					URL:                  "http://endpoint.svc",
					MonitoringURL:        "http://monitoring.com",
					Environment: &models.Environment{
						ID:         models.ID(1),
						Name:       "dev",
						Cluster:    "dev",
						IsDefault:  &trueBoolean,
						Region:     "id",
						GcpProject: "dev-proj",
						MaxCPU:     "1",
						MaxMemory:  "1Gi",
					}, EnvironmentName: "dev",
					Message:         "",
					ResourceRequest: nil,
					EnvVars: models.EnvVars([]models.EnvVar{
						{
							Name:  "WORKER",
							Value: "1",
						},
					}),
				}, nil)
				svc.On("UndeployEndpoint", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&models.VersionEndpoint{
					ID:                   uuid,
					VersionID:            models.ID(1),
					VersionModelID:       models.ID(1),
					Status:               models.EndpointTerminated,
					URL:                  "http://endpoint.svc",
					ServiceName:          "sample",
					InferenceServiceName: "sample",
					Namespace:            "sample",
					MonitoringURL:        "http://monitoring.com",
					Environment: &models.Environment{
						ID:         models.ID(1),
						Name:       "dev",
						Cluster:    "dev",
						IsDefault:  &trueBoolean,
						Region:     "id",
						GcpProject: "dev-proj",
						MaxCPU:     "1",
						MaxMemory:  "1Gi",
					},
					EnvironmentName: "dev",
					Message:         "",
					ResourceRequest: nil,
					EnvVars: models.EnvVars([]models.EnvVar{
						{
							Name:  "WORKER",
							Value: "1",
						},
					}),
					CreatedUpdated: models.CreatedUpdated{},
				}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusOK,
				data: nil,
			},
		},
		{
			desc: "Should return 404 if model is not found",
			vars: map[string]string{
				"model_id":    "1",
				"version_id":  "1",
				"endpoint_id": uuid.String(),
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(nil, gorm.ErrRecordNotFound)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				return svc
			},
			expected: &Response{
				code: http.StatusNotFound,
				data: Error{Message: "model with given id: 1 not found"},
			},
		},
		{
			desc: "Should return 404 if version is not found",
			vars: map[string]string{
				"model_id":    "1",
				"version_id":  "1",
				"endpoint_id": uuid.String(),
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", context.Background(), models.ID(1)).Return(&models.Model{
					ID:           models.ID(1),
					Name:         "model-1",
					ProjectID:    models.ID(1),
					Project:      mlp.Project{},
					ExperimentID: 1,
					Type:         "pyfunc",
					MlflowURL:    "",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", context.Background(), models.ID(1), models.ID(1), mock.Anything).Return(nil, gorm.ErrRecordNotFound)
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				return svc
			},
			expected: &Response{
				code: http.StatusNotFound,
				data: Error{Message: "model version with given id: 1 not found"},
			},
		},
		{
			desc: "Should return 200 if endpoint is not found",
			vars: map[string]string{
				"model_id":    "1",
				"version_id":  "1",
				"endpoint_id": uuid.String(),
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", context.Background(), models.ID(1)).Return(&models.Model{
					ID:           models.ID(1),
					Name:         "model-1",
					ProjectID:    models.ID(1),
					Project:      mlp.Project{},
					ExperimentID: 1,
					Type:         "pyfunc",
					MlflowURL:    "",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", context.Background(), models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "",
						Endpoints:    nil,
					},
				}, nil)
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("FindByID", context.Background(), uuid).Return(nil, gorm.ErrRecordNotFound)
				return svc
			},
			expected: &Response{
				code: http.StatusOK,
				data: fmt.Sprintf("Version endpoint %s is not available", uuid),
			},
		},
		{
			desc: "Should return 500 if error fetching endpoint",
			vars: map[string]string{
				"model_id":    "1",
				"version_id":  "1",
				"endpoint_id": uuid.String(),
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", context.Background(), models.ID(1)).Return(&models.Model{
					ID:           models.ID(1),
					Name:         "model-1",
					ProjectID:    models.ID(1),
					Project:      mlp.Project{},
					ExperimentID: 1,
					Type:         "pyfunc",
					MlflowURL:    "",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", context.Background(), models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "",
						Endpoints:    nil,
					},
				}, nil)
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("FindByID", context.Background(), uuid).Return(nil, fmt.Errorf("DB is down"))
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error while finding endpoint"},
			},
		},
		{
			desc: "Should return 500 if could not find environment",
			vars: map[string]string{
				"model_id":    "1",
				"version_id":  "1",
				"endpoint_id": uuid.String(),
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", context.Background(), models.ID(1)).Return(&models.Model{
					ID:           models.ID(1),
					Name:         "model-1",
					ProjectID:    models.ID(1),
					Project:      mlp.Project{},
					ExperimentID: 1,
					Type:         "pyfunc",
					MlflowURL:    "",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", context.Background(), models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "",
						Endpoints:    nil,
					},
				}, nil)
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				svc.On("GetEnvironment", "dev").Return(nil, gorm.ErrRecordNotFound)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("FindByID", context.Background(), uuid).Return(&models.VersionEndpoint{
					ID:                   uuid,
					VersionID:            models.ID(1),
					VersionModelID:       models.ID(1),
					Status:               models.EndpointRunning,
					ServiceName:          "sample",
					InferenceServiceName: "sample",
					Namespace:            "sample",
					URL:                  "http://endpoint.svc",
					MonitoringURL:        "http://monitoring.com",
					Environment: &models.Environment{
						ID:         models.ID(1),
						Name:       "dev",
						Cluster:    "dev",
						IsDefault:  &trueBoolean,
						Region:     "id",
						GcpProject: "dev-proj",
						MaxCPU:     "1",
						MaxMemory:  "1Gi",
					}, EnvironmentName: "dev",
					Message:         "",
					ResourceRequest: nil,
					EnvVars: models.EnvVars([]models.EnvVar{
						{
							Name:  "WORKER",
							Value: "1",
						},
					}),
				}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Unable to find environment dev"},
			},
		},
		{
			desc: "Should return 400 if endpoint is currently serving",
			vars: map[string]string{
				"model_id":    "1",
				"version_id":  "1",
				"endpoint_id": uuid.String(),
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", context.Background(), models.ID(1)).Return(&models.Model{
					ID:           models.ID(1),
					Name:         "model-1",
					ProjectID:    models.ID(1),
					Project:      mlp.Project{},
					ExperimentID: 1,
					Type:         "pyfunc",
					MlflowURL:    "",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", context.Background(), models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "",
						Endpoints:    nil,
					},
				}, nil)
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				svc.On("GetEnvironment", "dev").Return(&models.Environment{
					ID:         models.ID(1),
					Name:       "dev",
					Cluster:    "dev",
					IsDefault:  &trueBoolean,
					Region:     "id",
					GcpProject: "dev-proj",
					MaxCPU:     "1",
					MaxMemory:  "1Gi",
				}, nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("FindByID", context.Background(), uuid).Return(&models.VersionEndpoint{
					ID:                   uuid,
					VersionID:            models.ID(1),
					VersionModelID:       models.ID(1),
					Status:               models.EndpointServing,
					ServiceName:          "sample",
					InferenceServiceName: "sample",
					Namespace:            "sample",
					URL:                  "http://endpoint.svc",
					MonitoringURL:        "http://monitoring.com",
					Environment: &models.Environment{
						ID:         models.ID(1),
						Name:       "dev",
						Cluster:    "dev",
						IsDefault:  &trueBoolean,
						Region:     "id",
						GcpProject: "dev-proj",
						MaxCPU:     "1",
						MaxMemory:  "1Gi",
					}, EnvironmentName: "dev",
					Message:         "",
					ResourceRequest: nil,
					EnvVars: models.EnvVars([]models.EnvVar{
						{
							Name:  "WORKER",
							Value: "1",
						},
					}),
				}, nil)
				svc.On("UndeployEndpoint", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&models.VersionEndpoint{
					ID:                   uuid,
					VersionID:            models.ID(1),
					VersionModelID:       models.ID(1),
					Status:               models.EndpointTerminated,
					URL:                  "http://endpoint.svc",
					ServiceName:          "sample",
					InferenceServiceName: "sample",
					Namespace:            "sample",
					MonitoringURL:        "http://monitoring.com",
					Environment: &models.Environment{
						ID:         models.ID(1),
						Name:       "dev",
						Cluster:    "dev",
						IsDefault:  &trueBoolean,
						Region:     "id",
						GcpProject: "dev-proj",
						MaxCPU:     "1",
						MaxMemory:  "1Gi",
					},
					EnvironmentName: "dev",
					Message:         "",
					ResourceRequest: nil,
					EnvVars: models.EnvVars([]models.EnvVar{
						{
							Name:  "WORKER",
							Value: "1",
						},
					}),
					CreatedUpdated: models.CreatedUpdated{},
				}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{Message: fmt.Sprintf("Version Endpoints %s is still serving traffic. Please route the traffic to another model version first", uuid)},
			},
		},
		{
			desc: "Should return 500 if failed undeploy endpoint",
			vars: map[string]string{
				"model_id":    "1",
				"version_id":  "1",
				"endpoint_id": uuid.String(),
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", context.Background(), models.ID(1)).Return(&models.Model{
					ID:           models.ID(1),
					Name:         "model-1",
					ProjectID:    models.ID(1),
					Project:      mlp.Project{},
					ExperimentID: 1,
					Type:         "pyfunc",
					MlflowURL:    "",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", context.Background(), models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "",
						Endpoints:    nil,
					},
				}, nil)
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				svc.On("GetEnvironment", "dev").Return(&models.Environment{
					ID:         models.ID(1),
					Name:       "dev",
					Cluster:    "dev",
					IsDefault:  &trueBoolean,
					Region:     "id",
					GcpProject: "dev-proj",
					MaxCPU:     "1",
					MaxMemory:  "1Gi",
				}, nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("FindByID", context.Background(), uuid).Return(&models.VersionEndpoint{
					ID:                   uuid,
					VersionID:            models.ID(1),
					VersionModelID:       models.ID(1),
					Status:               models.EndpointRunning,
					ServiceName:          "sample",
					InferenceServiceName: "sample",
					Namespace:            "sample",
					URL:                  "http://endpoint.svc",
					MonitoringURL:        "http://monitoring.com",
					Environment: &models.Environment{
						ID:         models.ID(1),
						Name:       "dev",
						Cluster:    "dev",
						IsDefault:  &trueBoolean,
						Region:     "id",
						GcpProject: "dev-proj",
						MaxCPU:     "1",
						MaxMemory:  "1Gi",
					}, EnvironmentName: "dev",
					Message:         "",
					ResourceRequest: nil,
					EnvVars: models.EnvVars([]models.EnvVar{
						{
							Name:  "WORKER",
							Value: "1",
						},
					}),
				}, nil)
				svc.On("UndeployEndpoint", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("Connection refused"))
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: fmt.Sprintf("Unable to undeploy version endpoint %s", uuid)},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			modelSvc := tC.modelService()
			versionSvc := tC.versionService()
			envSvc := tC.envService()
			endpointSvc := tC.endpointService()

			ctl := &EndpointsController{
				AppContext: &AppContext{
					ModelsService:      modelSvc,
					VersionsService:    versionSvc,
					EnvironmentService: envSvc,
					EndpointsService:   endpointSvc,
					MonitoringConfig: config.MonitoringConfig{
						MonitoringEnabled: true,
						MonitoringBaseURL: "http://grafana",
					},
					AlertEnabled: true,
				},
			}
			resp := ctl.DeleteEndpoint(&http.Request{}, tC.vars, nil)
			assert.Equal(t, tC.expected, resp)
		})
	}
}
