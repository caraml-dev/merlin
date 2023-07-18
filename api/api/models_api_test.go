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
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/caraml-dev/merlin/config"
	"github.com/caraml-dev/merlin/service"
	mlflowDeleteServiceMocks "github.com/caraml-dev/mlp/api/pkg/client/mlflow/mocks"
	"github.com/google/uuid"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"github.com/caraml-dev/mlp/api/client"

	"github.com/caraml-dev/merlin/mlp"
	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/service/mocks"
)

func TestListModel(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		desc         string
		vars         map[string]string
		modelService func() *mocks.ModelsService
		expected     *Response
	}{
		{
			desc: "Should success list model",
			vars: map[string]string{
				"project_id": "1",
				"name":       "tensorflow",
			},
			modelService: func() *mocks.ModelsService {
				mockSvc := &mocks.ModelsService{}
				mockSvc.On("ListModels", mock.Anything, models.ID(1), "tensorflow").Return([]*models.Model{
					{
						ID:        models.ID(1),
						Name:      "tensorflow",
						ProjectID: models.ID(1),
						Project: mlp.Project(client.Project{
							ID:                1,
							Name:              "tensorflow",
							MLFlowTrackingURL: "http://mlflow.com",
							Administrators:    nil,
							Readers:           nil,
							Team:              "dsp",
							Stream:            "dsp",
							Labels:            nil,
							CreatedAt:         now,
							UpdatedAt:         now,
						}),
						ExperimentID: models.ID(1),
						Type:         "tensorflow",
						MlflowURL:    "http://mlflow.com",
						CreatedUpdated: models.CreatedUpdated{
							CreatedAt: now,
							UpdatedAt: now,
						},
					},
				}, nil)
				return mockSvc
			},
			expected: &Response{
				code: http.StatusOK,
				data: []*models.Model{
					{
						ID:        models.ID(1),
						Name:      "tensorflow",
						ProjectID: models.ID(1),
						Project: mlp.Project(client.Project{
							ID:                1,
							Name:              "tensorflow",
							MLFlowTrackingURL: "http://mlflow.com",
							Administrators:    nil,
							Readers:           nil,
							Team:              "dsp",
							Stream:            "dsp",
							Labels:            nil,
							CreatedAt:         now,
							UpdatedAt:         now,
						}),
						ExperimentID: models.ID(1),
						Type:         "tensorflow",
						MlflowURL:    "http://mlflow.com",
						CreatedUpdated: models.CreatedUpdated{
							CreatedAt: now,
							UpdatedAt: now,
						},
					},
				},
			},
		},
		{
			desc: "Should failed list model",
			vars: map[string]string{
				"project_id": "1",
				"name":       "tensorflow",
			},
			modelService: func() *mocks.ModelsService {
				mockSvc := &mocks.ModelsService{}
				mockSvc.On("ListModels", mock.Anything, models.ID(1), "tensorflow").Return(nil, fmt.Errorf("MLP API is down"))
				return mockSvc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error listing models: MLP API is down"},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			mockSvc := tC.modelService()
			ctl := &ModelsController{
				AppContext: &AppContext{
					ModelsService: mockSvc,
				},
			}
			resp := ctl.ListModels(&http.Request{}, tC.vars, nil)
			assertEqualResponses(t, tC.expected, resp)
		})
	}
}

func TestGetModel(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		desc           string
		vars           map[string]string
		modelService   func() *mocks.ModelsService
		projectService func() *mocks.ProjectsService
		expected       *Response
	}{
		{
			desc: "Should success get model",
			vars: map[string]string{
				"project_id": "1",
				"model_id":   "1",
			},
			projectService: func() *mocks.ProjectsService {
				mockSvc := &mocks.ProjectsService{}
				mockSvc.On("GetByID", mock.Anything, int32(1)).Return(mlp.Project(client.Project{
					ID:                1,
					Name:              "tensorflow",
					MLFlowTrackingURL: "http://mlflow.com",
					Administrators:    nil,
					Readers:           nil,
					Team:              "dsp",
					Stream:            "dsp",
					Labels:            nil,
					CreatedAt:         now,
					UpdatedAt:         now,
				}), nil)
				return mockSvc
			},
			modelService: func() *mocks.ModelsService {
				mockSvc := &mocks.ModelsService{}
				mockSvc.On("FindByID", mock.Anything, models.ID(1)).Return(
					&models.Model{
						ID:        models.ID(1),
						Name:      "tensorflow",
						ProjectID: models.ID(1),
						Project: mlp.Project(client.Project{
							ID:                1,
							Name:              "tensorflow",
							MLFlowTrackingURL: "http://mlflow.com",
							Administrators:    nil,
							Readers:           nil,
							Team:              "dsp",
							Stream:            "dsp",
							Labels:            nil,
							CreatedAt:         now,
							UpdatedAt:         now,
						}),
						ExperimentID: models.ID(1),
						Type:         "tensorflow",
						MlflowURL:    "http://mlflow.com",
						CreatedUpdated: models.CreatedUpdated{
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil)
				return mockSvc
			},
			expected: &Response{
				code: http.StatusOK,
				data: &models.Model{
					ID:        models.ID(1),
					Name:      "tensorflow",
					ProjectID: models.ID(1),
					Project: mlp.Project(client.Project{
						ID:                1,
						Name:              "tensorflow",
						MLFlowTrackingURL: "http://mlflow.com",
						Administrators:    nil,
						Readers:           nil,
						Team:              "dsp",
						Stream:            "dsp",
						Labels:            nil,
						CreatedAt:         now,
						UpdatedAt:         now,
					}),
					ExperimentID: models.ID(1),
					Type:         "tensorflow",
					MlflowURL:    "http://mlflow.com",
					CreatedUpdated: models.CreatedUpdated{
						CreatedAt: now,
						UpdatedAt: now,
					},
				},
			},
		},
		{
			desc: "Should failed if project api called was failing",
			vars: map[string]string{
				"project_id": "1",
				"model_id":   "1",
			},
			projectService: func() *mocks.ProjectsService {
				mockSvc := &mocks.ProjectsService{}
				mockSvc.On("GetByID", mock.Anything, int32(1)).Return(mlp.Project(client.Project{}), fmt.Errorf("Model not found: Model not found: Project API is down"))
				return mockSvc
			},
			modelService: func() *mocks.ModelsService {
				mockSvc := &mocks.ModelsService{}
				return mockSvc
			},
			expected: &Response{
				code: http.StatusNotFound,
				data: Error{Message: "Model not found: Model not found: Model not found: Project API is down"},
			},
		},
		{
			desc: "Should return not found if model is not found",
			vars: map[string]string{
				"project_id": "1",
				"model_id":   "1",
			},
			projectService: func() *mocks.ProjectsService {
				mockSvc := &mocks.ProjectsService{}
				mockSvc.On("GetByID", mock.Anything, int32(1)).Return(mlp.Project(client.Project{
					ID:                1,
					Name:              "tensorflow",
					MLFlowTrackingURL: "http://mlflow.com",
					Administrators:    nil,
					Readers:           nil,
					Team:              "dsp",
					Stream:            "dsp",
					Labels:            nil,
					CreatedAt:         now,
					UpdatedAt:         now,
				}), nil)
				return mockSvc
			},
			modelService: func() *mocks.ModelsService {
				mockSvc := &mocks.ModelsService{}
				mockSvc.On("FindByID", mock.Anything, models.ID(1)).Return(nil, gorm.ErrRecordNotFound)
				return mockSvc
			},
			expected: &Response{
				code: http.StatusNotFound,
				data: Error{Message: "Model not found: record not found"},
			},
		},
		{
			desc: "Should return internal server error if fetching model returning error",
			vars: map[string]string{
				"project_id": "1",
				"model_id":   "1",
			},
			projectService: func() *mocks.ProjectsService {
				mockSvc := &mocks.ProjectsService{}
				mockSvc.On("GetByID", mock.Anything, int32(1)).Return(mlp.Project(client.Project{
					ID:                1,
					Name:              "tensorflow",
					MLFlowTrackingURL: "http://mlflow.com",
					Administrators:    nil,
					Readers:           nil,
					Team:              "dsp",
					Stream:            "dsp",
					Labels:            nil,
					CreatedAt:         now,
					UpdatedAt:         now,
				}), nil)
				return mockSvc
			},
			modelService: func() *mocks.ModelsService {
				mockSvc := &mocks.ModelsService{}
				mockSvc.On("FindByID", mock.Anything, models.ID(1)).Return(nil, fmt.Errorf("DB is unreachable"))
				return mockSvc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error getting model: DB is unreachable"},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			mockSvc := tC.modelService()
			projectSvc := tC.projectService()
			ctl := &ModelsController{
				AppContext: &AppContext{
					ModelsService:   mockSvc,
					ProjectsService: projectSvc,
				},
			}
			resp := ctl.GetModel(&http.Request{}, tC.vars, nil)
			assertEqualResponses(t, tC.expected, resp)
		})
	}
}

func TestDeleteModel(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		desc                 string
		vars                 map[string]string
		projectService       func() *mocks.ProjectsService
		versionService       func() *mocks.VersionsService
		modelsService        func() *mocks.ModelsService
		mlflowDeleteService  func() *mlflowDeleteServiceMocks.Service
		predictionJobService func() *mocks.PredictionJobService
		endpointService      func() *mocks.EndpointsService
		modelEndpointService func() *mocks.ModelEndpointsService
		expected             *Response
	}{
		{
			desc: "Should successfully delete model with pyfunc_v2 type",
			vars: map[string]string{
				"model_id":   "1",
				"project_id": "1",
			},
			projectService: func() *mocks.ProjectsService {
				svc := &mocks.ProjectsService{}
				svc.On("GetByID", mock.Anything, int32(1)).Return(mlp.Project(client.Project{
					ID:                1,
					Name:              "iris",
					MLFlowTrackingURL: "http://mlflow.com",
					Administrators:    nil,
					Readers:           nil,
					Team:              "dsp",
					Stream:            "dsp",
					Labels:            nil,
					CreatedAt:         now,
					UpdatedAt:         now,
				}), nil)
				return svc
			},
			modelsService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:        models.ID(1),
					Name:      "model-1",
					ProjectID: models.ID(1),
					Project: mlp.Project{
						MLFlowTrackingURL: "http://www.notinuse.com",
					},
					ExperimentID: 1,
					Type:         "pyfunc_v2",
					MlflowURL:    "http://mlflow.com",
					Endpoints:    nil,
				}, nil)
				svc.On("Delete", mock.Anything).Return(nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("ListVersions", mock.Anything, models.ID(1), mock.Anything, mock.Anything).Return([]*models.Version{
					{
						ID:      models.ID(1),
						ModelID: models.ID(1),
						Model: &models.Model{
							ID:           models.ID(1),
							Name:         "model-1",
							ProjectID:    models.ID(1),
							Project:      mlp.Project{},
							ExperimentID: 1,
							Type:         "pyfunc_v2",
							MlflowURL:    "http://mlflow.com",
						},
						MlflowURL: "http://mlflow.com",
						RunID:     "runID1",
					},
				}, "", nil)
				svc.On("Delete", mock.Anything).Return(nil)
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				svc.On("ListPredictionJobs", mock.Anything, mock.Anything, &service.ListPredictionJobQuery{
					ModelID:   models.ID(1),
					VersionID: models.ID(1),
				}).Return([]*models.PredictionJob{}, nil)
				return svc
			},
			mlflowDeleteService: func() *mlflowDeleteServiceMocks.Service {
				svc := &mlflowDeleteServiceMocks.Service{}
				svc.On("DeleteExperiment", mock.Anything, "1", mock.Anything).Return(nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("ListEndpoints", mock.Anything, mock.Anything, mock.Anything).Return([]*models.VersionEndpoint{}, nil)
				return svc
			},
			modelEndpointService: func() *mocks.ModelEndpointsService {
				svc := &mocks.ModelEndpointsService{}
				svc.On("ListModelEndpoints", mock.Anything, models.ID(1)).Return([]*models.ModelEndpoint{}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusOK,
				data: models.ID(1),
			},
		},
		{
			desc: "Should successfully delete model with type other than pyfunc_v2",
			vars: map[string]string{
				"model_id":   "1",
				"project_id": "1",
			},
			projectService: func() *mocks.ProjectsService {
				svc := &mocks.ProjectsService{}
				svc.On("GetByID", mock.Anything, int32(1)).Return(mlp.Project(client.Project{
					ID:                1,
					Name:              "iris",
					MLFlowTrackingURL: "http://mlflow.com",
					Administrators:    nil,
					Readers:           nil,
					Team:              "dsp",
					Stream:            "dsp",
					Labels:            nil,
					CreatedAt:         now,
					UpdatedAt:         now,
				}), nil)
				return svc
			},
			modelsService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:        models.ID(1),
					Name:      "model-1",
					ProjectID: models.ID(1),
					Project: mlp.Project{
						MLFlowTrackingURL: "http://www.notinuse.com",
					},
					ExperimentID: 1,
					Type:         "xgboost",
					MlflowURL:    "http://mlflow.com",
					Endpoints:    nil,
				}, nil)
				svc.On("Delete", mock.Anything).Return(nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("ListVersions", mock.Anything, models.ID(1), mock.Anything, mock.Anything).Return([]*models.Version{
					{
						ID:      models.ID(1),
						ModelID: models.ID(1),
						Model: &models.Model{
							ID:           models.ID(1),
							Name:         "model-1",
							ProjectID:    models.ID(1),
							Project:      mlp.Project{},
							ExperimentID: 1,
							Type:         "xgboost",
							MlflowURL:    "http://mlflow.com",
						},
						MlflowURL: "http://mlflow.com",
						RunID:     "runID1",
					},
				}, "", nil)
				svc.On("Delete", mock.Anything).Return(nil)
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				return svc
			},
			mlflowDeleteService: func() *mlflowDeleteServiceMocks.Service {
				svc := &mlflowDeleteServiceMocks.Service{}
				svc.On("DeleteExperiment", mock.Anything, "1", mock.Anything).Return(nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("ListEndpoints", mock.Anything, mock.Anything, mock.Anything).Return([]*models.VersionEndpoint{}, nil)
				return svc
			},
			modelEndpointService: func() *mocks.ModelEndpointsService {
				svc := &mocks.ModelEndpointsService{}
				svc.On("ListModelEndpoints", mock.Anything, models.ID(1)).Return([]*models.ModelEndpoint{}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusOK,
				data: models.ID(1),
			},
		},
		{
			desc: "Should return 400 if there are active Endpoints",
			vars: map[string]string{
				"model_id":   "1",
				"project_id": "1",
			},
			projectService: func() *mocks.ProjectsService {
				svc := &mocks.ProjectsService{}
				svc.On("GetByID", mock.Anything, int32(1)).Return(mlp.Project(client.Project{
					ID:                1,
					Name:              "iris",
					MLFlowTrackingURL: "http://mlflow.com",
					Administrators:    nil,
					Readers:           nil,
					Team:              "dsp",
					Stream:            "dsp",
					Labels:            nil,
					CreatedAt:         now,
					UpdatedAt:         now,
				}), nil)
				return svc
			},
			modelsService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:        models.ID(1),
					Name:      "model-1",
					ProjectID: models.ID(1),
					Project: mlp.Project{
						MLFlowTrackingURL: "http://www.notinuse.com",
					},
					ExperimentID: 1,
					Type:         "xgboost",
					MlflowURL:    "http://mlflow.com",
					Endpoints:    nil,
				}, nil)
				svc.On("Delete", mock.Anything).Return(nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("ListVersions", mock.Anything, models.ID(1), mock.Anything, mock.Anything).Return([]*models.Version{
					{
						ID:      models.ID(1),
						ModelID: models.ID(1),
						Model: &models.Model{
							ID:           models.ID(1),
							Name:         "model-1",
							ProjectID:    models.ID(1),
							Project:      mlp.Project{},
							ExperimentID: 1,
							Type:         "xgboost",
							MlflowURL:    "http://mlflow.com",
						},
						MlflowURL: "http://mlflow.com",
						RunID:     "runID1",
					},
				}, "", nil)
				svc.On("Delete", mock.Anything).Return(nil)
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				return svc
			},
			mlflowDeleteService: func() *mlflowDeleteServiceMocks.Service {
				svc := &mlflowDeleteServiceMocks.Service{}
				svc.On("DeleteExperiment", mock.Anything, "1", mock.Anything).Return(nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("ListEndpoints", mock.Anything, mock.Anything, mock.Anything).Return([]*models.VersionEndpoint{
					{
						ID:             uuid.New(),
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
			modelEndpointService: func() *mocks.ModelEndpointsService {
				svc := &mocks.ModelEndpointsService{}
				return svc
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{Message: "There are active endpoint that still using this model version"},
			},
		},
		{
			desc: "Should return 500 if failed to delete inactive endpoint",
			vars: map[string]string{
				"model_id":   "1",
				"project_id": "1",
			},
			projectService: func() *mocks.ProjectsService {
				svc := &mocks.ProjectsService{}
				svc.On("GetByID", mock.Anything, int32(1)).Return(mlp.Project(client.Project{
					ID:                1,
					Name:              "iris",
					MLFlowTrackingURL: "http://mlflow.com",
					Administrators:    nil,
					Readers:           nil,
					Team:              "dsp",
					Stream:            "dsp",
					Labels:            nil,
					CreatedAt:         now,
					UpdatedAt:         now,
				}), nil)
				return svc
			},
			modelsService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:        models.ID(1),
					Name:      "model-1",
					ProjectID: models.ID(1),
					Project: mlp.Project{
						MLFlowTrackingURL: "http://www.notinuse.com",
					},
					ExperimentID: 1,
					Type:         "xgboost",
					MlflowURL:    "http://mlflow.com",
					Endpoints:    nil,
				}, nil)
				svc.On("Delete", mock.Anything).Return(nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("ListVersions", mock.Anything, models.ID(1), mock.Anything, mock.Anything).Return([]*models.Version{
					{
						ID:      models.ID(1),
						ModelID: models.ID(1),
						Model: &models.Model{
							ID:           models.ID(1),
							Name:         "model-1",
							ProjectID:    models.ID(1),
							Project:      mlp.Project{},
							ExperimentID: 1,
							Type:         "xgboost",
							MlflowURL:    "http://mlflow.com",
						},
						MlflowURL: "http://mlflow.com",
						RunID:     "runID1",
					},
				}, "", nil)
				svc.On("Delete", mock.Anything).Return(nil)
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				return svc
			},
			mlflowDeleteService: func() *mlflowDeleteServiceMocks.Service {
				svc := &mlflowDeleteServiceMocks.Service{}
				svc.On("DeleteExperiment", mock.Anything, "1", mock.Anything).Return(nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("ListEndpoints", mock.Anything, mock.Anything, mock.Anything).Return([]*models.VersionEndpoint{
					{
						ID:             uuid.New(),
						VersionID:      models.ID(1),
						VersionModelID: models.ID(1),
						Status:         models.EndpointFailed,
						URL:            "http://endpoint-1.com",
						Environment: &models.Environment{
							ID:      models.ID(1),
							Name:    "dev",
							Cluster: "dev",
						},
					},
				}, nil)
				svc.On("DeleteEndpoint", mock.Anything, mock.Anything).Return(
					errors.New("failed to delete endpoint"))
				return svc
			},
			modelEndpointService: func() *mocks.ModelEndpointsService {
				svc := &mocks.ModelEndpointsService{}
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Failed to delete endpoint: failed to delete endpoint"},
			},
		},
		{
			desc: "Should return 400 if there are active prediction jobs",
			vars: map[string]string{
				"model_id":   "1",
				"project_id": "1",
			},
			projectService: func() *mocks.ProjectsService {
				svc := &mocks.ProjectsService{}
				svc.On("GetByID", mock.Anything, int32(1)).Return(mlp.Project(client.Project{
					ID:                1,
					Name:              "iris",
					MLFlowTrackingURL: "http://mlflow.com",
					Administrators:    nil,
					Readers:           nil,
					Team:              "dsp",
					Stream:            "dsp",
					Labels:            nil,
					CreatedAt:         now,
					UpdatedAt:         now,
				}), nil)
				return svc
			},
			modelsService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:        models.ID(1),
					Name:      "model-1",
					ProjectID: models.ID(1),
					Project: mlp.Project{
						MLFlowTrackingURL: "http://www.notinuse.com",
					},
					ExperimentID: 1,
					Type:         "pyfunc_v2",
					MlflowURL:    "http://mlflow.com",
					Endpoints:    nil,
				}, nil)
				svc.On("Delete", mock.Anything).Return(nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("ListVersions", mock.Anything, models.ID(1), mock.Anything, mock.Anything).Return([]*models.Version{
					{
						ID:      models.ID(1),
						ModelID: models.ID(1),
						Model: &models.Model{
							ID:           models.ID(1),
							Name:         "model-1",
							ProjectID:    models.ID(1),
							Project:      mlp.Project{},
							ExperimentID: 1,
							Type:         "pyfunc_v2",
							MlflowURL:    "http://mlflow.com",
						},
						MlflowURL: "http://mlflow.com",
						RunID:     "runID1",
					},
				}, "", nil)
				svc.On("Delete", mock.Anything).Return(nil)
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				svc.On("ListPredictionJobs", mock.Anything, mock.Anything, &service.ListPredictionJobQuery{
					ModelID:   models.ID(1),
					VersionID: models.ID(1),
				}).Return([]*models.PredictionJob{
					{
						ID:              models.ID(1),
						Name:            "prediction-job-1",
						ProjectID:       models.ID(1),
						VersionID:       models.ID(1),
						VersionModelID:  models.ID(1),
						EnvironmentName: "dev",
						Status:          models.JobRunning,
					},
				}, nil)
				return svc
			},
			mlflowDeleteService: func() *mlflowDeleteServiceMocks.Service {
				svc := &mlflowDeleteServiceMocks.Service{}
				svc.On("DeleteExperiment", mock.Anything, "1", mock.Anything).Return(nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				return svc
			},
			modelEndpointService: func() *mocks.ModelEndpointsService {
				svc := &mocks.ModelEndpointsService{}
				return svc
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{Message: "There are active prediction job that still using this model version"},
			},
		},
		{
			desc: "Should return 500 if failed to delete inactive jobs",
			vars: map[string]string{
				"model_id":   "1",
				"project_id": "1",
			},
			projectService: func() *mocks.ProjectsService {
				svc := &mocks.ProjectsService{}
				svc.On("GetByID", mock.Anything, int32(1)).Return(mlp.Project(client.Project{
					ID:                1,
					Name:              "iris",
					MLFlowTrackingURL: "http://mlflow.com",
					Administrators:    nil,
					Readers:           nil,
					Team:              "dsp",
					Stream:            "dsp",
					Labels:            nil,
					CreatedAt:         now,
					UpdatedAt:         now,
				}), nil)
				return svc
			},
			modelsService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:        models.ID(1),
					Name:      "model-1",
					ProjectID: models.ID(1),
					Project: mlp.Project{
						MLFlowTrackingURL: "http://www.notinuse.com",
					},
					ExperimentID: 1,
					Type:         "pyfunc_v2",
					MlflowURL:    "http://mlflow.com",
					Endpoints:    nil,
				}, nil)
				svc.On("Delete", mock.Anything).Return(nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("ListVersions", mock.Anything, models.ID(1), mock.Anything, mock.Anything).Return([]*models.Version{
					{
						ID:      models.ID(1),
						ModelID: models.ID(1),
						Model: &models.Model{
							ID:           models.ID(1),
							Name:         "model-1",
							ProjectID:    models.ID(1),
							Project:      mlp.Project{},
							ExperimentID: 1,
							Type:         "pyfunc_v2",
							MlflowURL:    "http://mlflow.com",
						},
						MlflowURL: "http://mlflow.com",
						RunID:     "runID1",
					},
				}, "", nil)
				svc.On("Delete", mock.Anything).Return(nil)
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				svc.On("ListPredictionJobs", mock.Anything, mock.Anything, &service.ListPredictionJobQuery{
					ModelID:   models.ID(1),
					VersionID: models.ID(1),
				}).Return([]*models.PredictionJob{
					{
						ID:              models.ID(1),
						Name:            "prediction-job-1",
						ProjectID:       models.ID(1),
						VersionID:       models.ID(1),
						VersionModelID:  models.ID(1),
						EnvironmentName: "dev",
						Status:          models.JobFailed,
					},
				}, nil)
				svc.On("StopPredictionJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					nil, errors.New("failed to stop prediction job"))
				return svc
			},
			mlflowDeleteService: func() *mlflowDeleteServiceMocks.Service {
				svc := &mlflowDeleteServiceMocks.Service{}
				svc.On("DeleteExperiment", mock.Anything, "1", mock.Anything).Return(nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("ListEndpoints", mock.Anything, mock.Anything, mock.Anything).Return([]*models.VersionEndpoint{}, nil)
				return svc
			},
			modelEndpointService: func() *mocks.ModelEndpointsService {
				svc := &mocks.ModelEndpointsService{}
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Failed stopping prediction job: failed to stop prediction job"},
			},
		},
		{
			desc: "Should return 500 if delete model endpoint failed",
			vars: map[string]string{
				"model_id":   "1",
				"project_id": "1",
			},
			projectService: func() *mocks.ProjectsService {
				svc := &mocks.ProjectsService{}
				svc.On("GetByID", mock.Anything, int32(1)).Return(mlp.Project(client.Project{
					ID:                1,
					Name:              "iris",
					MLFlowTrackingURL: "http://mlflow.com",
					Administrators:    nil,
					Readers:           nil,
					Team:              "dsp",
					Stream:            "dsp",
					Labels:            nil,
					CreatedAt:         now,
					UpdatedAt:         now,
				}), nil)
				return svc
			},
			modelsService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:        models.ID(1),
					Name:      "model-1",
					ProjectID: models.ID(1),
					Project: mlp.Project{
						MLFlowTrackingURL: "http://www.notinuse.com",
					},
					ExperimentID: 1,
					Type:         "xgboost",
					MlflowURL:    "http://mlflow.com",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("ListVersions", mock.Anything, models.ID(1), mock.Anything, mock.Anything).Return([]*models.Version{
					{
						ID:      models.ID(1),
						ModelID: models.ID(1),
						Model: &models.Model{
							ID:           models.ID(1),
							Name:         "model-1",
							ProjectID:    models.ID(1),
							Project:      mlp.Project{},
							ExperimentID: 1,
							Type:         "xgboost",
							MlflowURL:    "http://mlflow.com",
						},
						MlflowURL: "http://mlflow.com",
						RunID:     "runID1",
					},
				}, "", nil)
				svc.On("Delete", mock.Anything).Return(nil)
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				return svc
			},
			mlflowDeleteService: func() *mlflowDeleteServiceMocks.Service {
				svc := &mlflowDeleteServiceMocks.Service{}
				svc.On("DeleteExperiment", mock.Anything, "1", mock.Anything).Return(nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("ListEndpoints", mock.Anything, mock.Anything, mock.Anything).Return([]*models.VersionEndpoint{}, nil)
				return svc
			},
			modelEndpointService: func() *mocks.ModelEndpointsService {
				svc := &mocks.ModelEndpointsService{}
				svc.On("ListModelEndpoints", mock.Anything, models.ID(1)).Return([]*models.ModelEndpoint{
					{
						ID:      models.ID(1),
						ModelID: models.ID(1),
						Model: &models.Model{
							ID:           models.ID(1),
							Name:         "Model-1",
							ProjectID:    models.ID(1),
							Project:      mlp.Project{},
							ExperimentID: 0,
							Type:         "xgboost",
							MlflowURL:    "http://mlflow.com",
							Endpoints:    nil,
						},
					},
				}, nil)
				svc.On("DeleteModelEndpoint", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("failed to delete model endpoint"))
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Unable to delete model endpoint: failed to delete model endpoint"},
			},
		},
		{
			desc: "Should return 500 if delete mlflow run failed",
			vars: map[string]string{
				"model_id":   "1",
				"project_id": "1",
			},
			projectService: func() *mocks.ProjectsService {
				svc := &mocks.ProjectsService{}
				svc.On("GetByID", mock.Anything, int32(1)).Return(mlp.Project(client.Project{
					ID:                1,
					Name:              "iris",
					MLFlowTrackingURL: "http://mlflow.com",
					Administrators:    nil,
					Readers:           nil,
					Team:              "dsp",
					Stream:            "dsp",
					Labels:            nil,
					CreatedAt:         now,
					UpdatedAt:         now,
				}), nil)
				return svc
			},
			modelsService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:        models.ID(1),
					Name:      "model-1",
					ProjectID: models.ID(1),
					Project: mlp.Project{
						MLFlowTrackingURL: "http://www.notinuse.com",
					},
					ExperimentID: 1,
					Type:         "pyfunc_v2",
					MlflowURL:    "http://mlflow.com",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("ListVersions", mock.Anything, models.ID(1), mock.Anything, mock.Anything).Return([]*models.Version{
					{
						ID:      models.ID(1),
						ModelID: models.ID(1),
						Model: &models.Model{
							ID:           models.ID(1),
							Name:         "model-1",
							ProjectID:    models.ID(1),
							Project:      mlp.Project{},
							ExperimentID: 1,
							Type:         "pyfunc_v2",
							MlflowURL:    "http://mlflow.com",
						},
						MlflowURL: "http://mlflow.com",
						RunID:     "runID1",
					},
				}, "", nil)
				svc.On("Delete", mock.Anything).Return(nil)
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				svc.On("ListPredictionJobs", mock.Anything, mock.Anything, &service.ListPredictionJobQuery{
					ModelID:   models.ID(1),
					VersionID: models.ID(1),
				}).Return([]*models.PredictionJob{}, nil)
				return svc
			},
			mlflowDeleteService: func() *mlflowDeleteServiceMocks.Service {
				svc := &mlflowDeleteServiceMocks.Service{}
				svc.On("DeleteExperiment", mock.Anything, "1", mock.Anything).Return(errors.New("failed to delete mlflow experiment"))
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("ListEndpoints", mock.Anything, mock.Anything, mock.Anything).Return([]*models.VersionEndpoint{}, nil)
				return svc
			},
			modelEndpointService: func() *mocks.ModelEndpointsService {
				svc := &mocks.ModelEndpointsService{}
				svc.On("ListModelEndpoints", mock.Anything, models.ID(1)).Return([]*models.ModelEndpoint{}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Delete mlflow experiment failed: failed to delete mlflow experiment"},
			},
		},
		{
			desc: "Should return 500 if delete model version from database failed",
			vars: map[string]string{
				"model_id":   "1",
				"project_id": "1",
			},
			projectService: func() *mocks.ProjectsService {
				svc := &mocks.ProjectsService{}
				svc.On("GetByID", mock.Anything, int32(1)).Return(mlp.Project(client.Project{
					ID:                1,
					Name:              "iris",
					MLFlowTrackingURL: "http://mlflow.com",
					Administrators:    nil,
					Readers:           nil,
					Team:              "dsp",
					Stream:            "dsp",
					Labels:            nil,
					CreatedAt:         now,
					UpdatedAt:         now,
				}), nil)
				return svc
			},
			modelsService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:        models.ID(1),
					Name:      "model-1",
					ProjectID: models.ID(1),
					Project: mlp.Project{
						MLFlowTrackingURL: "http://www.notinuse.com",
					},
					ExperimentID: 1,
					Type:         "pyfunc_v2",
					MlflowURL:    "http://mlflow.com",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("ListVersions", mock.Anything, models.ID(1), mock.Anything, mock.Anything).Return([]*models.Version{
					{
						ID:      models.ID(1),
						ModelID: models.ID(1),
						Model: &models.Model{
							ID:           models.ID(1),
							Name:         "model-1",
							ProjectID:    models.ID(1),
							Project:      mlp.Project{},
							ExperimentID: 1,
							Type:         "pyfunc_v2",
							MlflowURL:    "http://mlflow.com",
						},
						MlflowURL: "http://mlflow.com",
						RunID:     "runID1",
					},
				}, "", nil)
				svc.On("Delete", mock.Anything).Return(errors.New("failed to delete model version"))
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				svc.On("ListPredictionJobs", mock.Anything, mock.Anything, &service.ListPredictionJobQuery{
					ModelID:   models.ID(1),
					VersionID: models.ID(1),
				}).Return([]*models.PredictionJob{}, nil)
				return svc
			},
			mlflowDeleteService: func() *mlflowDeleteServiceMocks.Service {
				svc := &mlflowDeleteServiceMocks.Service{}
				svc.On("DeleteExperiment", mock.Anything, "1", mock.Anything).Return(nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("ListEndpoints", mock.Anything, mock.Anything, mock.Anything).Return([]*models.VersionEndpoint{}, nil)
				return svc
			},
			modelEndpointService: func() *mocks.ModelEndpointsService {
				svc := &mocks.ModelEndpointsService{}
				svc.On("ListModelEndpoints", mock.Anything, models.ID(1)).Return([]*models.ModelEndpoint{}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Delete version id 1 failed: failed to delete model version"},
			},
		},
		{
			desc: "Should return 500 if delete model from database failed",
			vars: map[string]string{
				"model_id":   "1",
				"project_id": "1",
			},
			projectService: func() *mocks.ProjectsService {
				svc := &mocks.ProjectsService{}
				svc.On("GetByID", mock.Anything, int32(1)).Return(mlp.Project(client.Project{
					ID:                1,
					Name:              "iris",
					MLFlowTrackingURL: "http://mlflow.com",
					Administrators:    nil,
					Readers:           nil,
					Team:              "dsp",
					Stream:            "dsp",
					Labels:            nil,
					CreatedAt:         now,
					UpdatedAt:         now,
				}), nil)
				return svc
			},
			modelsService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("Delete", mock.Anything).Return(errors.New("failed to delete model"))
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:        models.ID(1),
					Name:      "model-1",
					ProjectID: models.ID(1),
					Project: mlp.Project{
						MLFlowTrackingURL: "http://www.notinuse.com",
					},
					ExperimentID: 1,
					Type:         "pyfunc_v2",
					MlflowURL:    "http://mlflow.com",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("ListVersions", mock.Anything, models.ID(1), mock.Anything, mock.Anything).Return([]*models.Version{
					{
						ID:      models.ID(1),
						ModelID: models.ID(1),
						Model: &models.Model{
							ID:           models.ID(1),
							Name:         "model-1",
							ProjectID:    models.ID(1),
							Project:      mlp.Project{},
							ExperimentID: 1,
							Type:         "pyfunc_v2",
							MlflowURL:    "http://mlflow.com",
						},
						MlflowURL: "http://mlflow.com",
						RunID:     "runID1",
					},
				}, "", nil)
				svc.On("Delete", mock.Anything).Return(nil)
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				svc.On("ListPredictionJobs", mock.Anything, mock.Anything, &service.ListPredictionJobQuery{
					ModelID:   models.ID(1),
					VersionID: models.ID(1),
				}).Return([]*models.PredictionJob{}, nil)
				return svc
			},
			mlflowDeleteService: func() *mlflowDeleteServiceMocks.Service {
				svc := &mlflowDeleteServiceMocks.Service{}
				svc.On("DeleteExperiment", mock.Anything, "1", mock.Anything).Return(nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("ListEndpoints", mock.Anything, mock.Anything, mock.Anything).Return([]*models.VersionEndpoint{}, nil)
				return svc
			},
			modelEndpointService: func() *mocks.ModelEndpointsService {
				svc := &mocks.ModelEndpointsService{}
				svc.On("ListModelEndpoints", mock.Anything, models.ID(1)).Return([]*models.ModelEndpoint{}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Delete model failed: failed to delete model"},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			versionSvc := tC.versionService
			modelsSvc := tC.modelsService
			mlflowDeleteSvc := tC.mlflowDeleteService
			predictionJobSvc := tC.predictionJobService
			endpointSvc := tC.endpointService
			projectSvc := tC.projectService
			modelEndpointSvc := tC.modelEndpointService

			ctl := &ModelsController{
				AppContext: &AppContext{
					VersionsService:       versionSvc(),
					ModelsService:         modelsSvc(),
					MlflowDeleteService:   mlflowDeleteSvc(),
					PredictionJobService:  predictionJobSvc(),
					EndpointsService:      endpointSvc(),
					ProjectsService:       projectSvc(),
					ModelEndpointsService: modelEndpointSvc(),
					FeatureToggleConfig: config.FeatureToggleConfig{
						AlertConfig: config.AlertConfig{
							AlertEnabled: true,
						},
						MonitoringConfig: config.MonitoringConfig{
							MonitoringEnabled: true,
							MonitoringBaseURL: "http://grafana",
						},
					},
				},
				VersionsController: &VersionsController{
					AppContext: &AppContext{
						VersionsService:      versionSvc(),
						MlflowDeleteService:  mlflowDeleteSvc(),
						PredictionJobService: predictionJobSvc(),
						EndpointsService:     endpointSvc(),
						FeatureToggleConfig: config.FeatureToggleConfig{
							AlertConfig: config.AlertConfig{
								AlertEnabled: true,
							},
							MonitoringConfig: config.MonitoringConfig{
								MonitoringEnabled: true,
								MonitoringBaseURL: "http://grafana",
							},
						},
					},
				},
			}
			resp := ctl.DeleteModel(&http.Request{}, tC.vars, nil)
			assertEqualResponses(t, tC.expected, resp)
		})
	}
}
