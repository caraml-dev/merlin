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
	"net/url"
	"testing"

	uuid2 "github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/caraml-dev/merlin/config"
	"github.com/caraml-dev/merlin/mlp"
	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/service"
	"github.com/caraml-dev/merlin/service/mocks"
	"github.com/caraml-dev/mlp/api/client"
)

func TestList(t *testing.T) {
	testCases := []struct {
		desc                 string
		vars                 map[string]string
		modelService         func() *mocks.ModelsService
		versionService       func() *mocks.VersionsService
		predictionJobService func() *mocks.PredictionJobService
		expected             *Response
	}{
		{
			desc: "Should succcess list prediction job",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
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
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				svc.On("ListPredictionJobs", context.Background(), mock.Anything, &service.ListPredictionJobQuery{
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
					},
				}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusOK,
				data: []*models.PredictionJob{
					{
						ID:              models.ID(1),
						Name:            "prediction-job-1",
						ProjectID:       models.ID(1),
						VersionID:       models.ID(1),
						VersionModelID:  models.ID(1),
						EnvironmentName: "dev",
					},
				},
			},
		},
		{
			desc: "Should return 500 if error fetching model",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(nil, fmt.Errorf("Error creating secret: db is down"))
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error getting model / version: error retrieving model with id: 1"},
			},
		},
		{
			desc: "Should return 500 if list prediction job returning error",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
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
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				svc.On("ListPredictionJobs", context.Background(), mock.Anything, &service.ListPredictionJobQuery{
					ModelID:   models.ID(1),
					VersionID: models.ID(1),
				}).Return(nil, fmt.Errorf("Connection refused"))
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error listing prediction jobs: Connection refused"},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			modelSvc := tC.modelService()
			versionSvc := tC.versionService()
			predictionJobSvc := tC.predictionJobService()
			ctl := &PredictionJobController{
				AppContext: &AppContext{
					ModelsService:        modelSvc,
					VersionsService:      versionSvc,
					PredictionJobService: predictionJobSvc,
					MonitoringConfig: config.MonitoringConfig{
						MonitoringEnabled: true,
						MonitoringBaseURL: "http://grafana",
					},
					AlertEnabled: true,
				},
			}
			resp := ctl.List(&http.Request{}, tC.vars, nil)
			assertEqualResponses(t, tC.expected, resp)
		})
	}
}

func TestGet(t *testing.T) {
	trueBoolean := true
	testCases := []struct {
		desc                 string
		vars                 map[string]string
		modelService         func() *mocks.ModelsService
		versionService       func() *mocks.VersionsService
		envService           func() *mocks.EnvironmentService
		predictionJobService func() *mocks.PredictionJobService
		expected             *Response
	}{
		{
			desc: "Should succcess get prediction job",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
				"job_id":     "1",
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
				svc.On("GetDefaultPredictionJobEnvironment").Return(&models.Environment{
					ID:                     models.ID(1),
					Name:                   "dev",
					Cluster:                "dev",
					Region:                 "id",
					GcpProject:             "id-proj",
					IsPredictionJobEnabled: true,
					IsDefaultPredictionJob: &trueBoolean,
				}, nil)
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				svc.On("GetPredictionJob", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&models.PredictionJob{
					ID:              models.ID(1),
					Name:            "prediction-job-1",
					ProjectID:       models.ID(1),
					VersionID:       models.ID(1),
					VersionModelID:  models.ID(1),
					EnvironmentName: "dev",
				}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusOK,
				data: &models.PredictionJob{
					ID:              models.ID(1),
					Name:            "prediction-job-1",
					ProjectID:       models.ID(1),
					VersionID:       models.ID(1),
					VersionModelID:  models.ID(1),
					EnvironmentName: "dev",
				},
			},
		},
		{
			desc: "Should return 500 if error fetching model",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(nil, fmt.Errorf("Error creating secret: db is down"))
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error getting model / version: error retrieving model with id: 1"},
			},
		},
		{
			desc: "Should return 500 if get default env prediction job returning error",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
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
				svc.On("GetDefaultPredictionJobEnvironment").Return(nil, fmt.Errorf("Error creating secret: db is down"))
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Unable to find default environment, specify environment target for deployment: Error creating secret: db is down"},
			},
		},
		{
			desc: "Should return 500 if error when get prediction job",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
				"job_id":     "1",
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
				svc.On("GetDefaultPredictionJobEnvironment").Return(&models.Environment{
					ID:                     models.ID(1),
					Name:                   "dev",
					Cluster:                "dev",
					Region:                 "id",
					GcpProject:             "id-proj",
					IsPredictionJobEnabled: true,
					IsDefaultPredictionJob: &trueBoolean,
				}, nil)
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				svc.On("GetPredictionJob", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("Connection refused"))
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error getting prediction job: Connection refused"},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			modelSvc := tC.modelService()
			versionSvc := tC.versionService()
			predictionJobSvc := tC.predictionJobService()
			envSvc := tC.envService()
			ctl := &PredictionJobController{
				AppContext: &AppContext{
					ModelsService:        modelSvc,
					VersionsService:      versionSvc,
					PredictionJobService: predictionJobSvc,
					EnvironmentService:   envSvc,
					MonitoringConfig: config.MonitoringConfig{
						MonitoringEnabled: true,
						MonitoringBaseURL: "http://grafana",
					},
					AlertEnabled: true,
				},
			}
			resp := ctl.Get(&http.Request{}, tC.vars, nil)
			assertEqualResponses(t, tC.expected, resp)
		})
	}
}

func TestStop(t *testing.T) {
	trueBoolean := true
	testCases := []struct {
		desc                 string
		vars                 map[string]string
		modelService         func() *mocks.ModelsService
		versionService       func() *mocks.VersionsService
		envService           func() *mocks.EnvironmentService
		predictionJobService func() *mocks.PredictionJobService
		expected             *Response
	}{
		{
			desc: "Should succcess stop prediction job",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
				"job_id":     "1",
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
				svc.On("GetDefaultPredictionJobEnvironment").Return(&models.Environment{
					ID:                     models.ID(1),
					Name:                   "dev",
					Cluster:                "dev",
					Region:                 "id",
					GcpProject:             "id-proj",
					IsPredictionJobEnabled: true,
					IsDefaultPredictionJob: &trueBoolean,
				}, nil)
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				svc.On("StopPredictionJob", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&models.PredictionJob{
					ID:              models.ID(1),
					Name:            "prediction-job-1",
					ProjectID:       models.ID(1),
					VersionID:       models.ID(1),
					VersionModelID:  models.ID(1),
					EnvironmentName: "dev",
				}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusNoContent,
			},
		},
		{
			desc: "Should return 500 if error fetching model",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(nil, fmt.Errorf("Error creating secret: db is down"))
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error getting model / version: error retrieving model with id: 1"},
			},
		},
		{
			desc: "Should return 500 if get default env prediction job returning error",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
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
				svc.On("GetDefaultPredictionJobEnvironment").Return(nil, fmt.Errorf("Error creating secret: db is down"))
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Unable to find default environment, specify environment target for deployment: Error creating secret: db is down"},
			},
		},
		{
			desc: "Should return 500 if error when stop prediction job",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
				"job_id":     "1",
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
				svc.On("GetDefaultPredictionJobEnvironment").Return(&models.Environment{
					ID:                     models.ID(1),
					Name:                   "dev",
					Cluster:                "dev",
					Region:                 "id",
					GcpProject:             "id-proj",
					IsPredictionJobEnabled: true,
					IsDefaultPredictionJob: &trueBoolean,
				}, nil)
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				svc.On("StopPredictionJob", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("Connection refused"))
				return svc
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{Message: "Error stopping prediction job: Connection refused"},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			modelSvc := tC.modelService()
			versionSvc := tC.versionService()
			predictionJobSvc := tC.predictionJobService()
			envSvc := tC.envService()
			ctl := &PredictionJobController{
				AppContext: &AppContext{
					ModelsService:        modelSvc,
					VersionsService:      versionSvc,
					PredictionJobService: predictionJobSvc,
					EnvironmentService:   envSvc,
					MonitoringConfig: config.MonitoringConfig{
						MonitoringEnabled: true,
						MonitoringBaseURL: "http://grafana",
					},
					AlertEnabled: true,
				},
			}
			resp := ctl.Stop(&http.Request{}, tC.vars, nil)
			assertEqualResponses(t, tC.expected, resp)
		})
	}
}

func TestListContainers_PredictionJob(t *testing.T) {
	trueBoolean := true
	uuid := uuid2.New()
	testCases := []struct {
		desc                 string
		vars                 map[string]string
		modelService         func() *mocks.ModelsService
		versionService       func() *mocks.VersionsService
		envService           func() *mocks.EnvironmentService
		predictionJobService func() *mocks.PredictionJobService
		expected             *Response
	}{
		{
			desc: "Should succcess get list of containers",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
				"job_id":     "1",
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
				svc.On("GetDefaultPredictionJobEnvironment").Return(&models.Environment{
					ID:                     models.ID(1),
					Name:                   "dev",
					Cluster:                "dev",
					Region:                 "id",
					GcpProject:             "id-proj",
					IsPredictionJobEnabled: true,
					IsDefaultPredictionJob: &trueBoolean,
				}, nil)
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				svc.On("GetPredictionJob", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&models.PredictionJob{
					ID:              models.ID(1),
					Name:            "prediction-job-1",
					ProjectID:       models.ID(1),
					VersionID:       models.ID(1),
					VersionModelID:  models.ID(1),
					EnvironmentName: "dev",
				}, nil)
				svc.On("ListContainers", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*models.Container{
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
			desc: "Should return 500 if error fetching model",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(nil, fmt.Errorf("Error creating secret: db is down"))
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error getting model / version: error retrieving model with id: 1"},
			},
		},
		{
			desc: "Should return 500 if get default env prediction job returning error",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
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
				svc.On("GetDefaultPredictionJobEnvironment").Return(nil, fmt.Errorf("Error creating secret: db is down"))
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Unable to find default environment, specify environment target for deployment: Error creating secret: db is down"},
			},
		},
		{
			desc: "Should return 500 if error when get prediction job",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
				"job_id":     "1",
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
				svc.On("GetDefaultPredictionJobEnvironment").Return(&models.Environment{
					ID:                     models.ID(1),
					Name:                   "dev",
					Cluster:                "dev",
					Region:                 "id",
					GcpProject:             "id-proj",
					IsPredictionJobEnabled: true,
					IsDefaultPredictionJob: &trueBoolean,
				}, nil)
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				svc.On("GetPredictionJob", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("Connection refused"))
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error getting prediction job: Connection refused"},
			},
		},
		{
			desc: "Should return 500 if fetching list of containers failed",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
				"job_id":     "1",
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
				svc.On("GetDefaultPredictionJobEnvironment").Return(&models.Environment{
					ID:                     models.ID(1),
					Name:                   "dev",
					Cluster:                "dev",
					Region:                 "id",
					GcpProject:             "id-proj",
					IsPredictionJobEnabled: true,
					IsDefaultPredictionJob: &trueBoolean,
				}, nil)
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				svc.On("GetPredictionJob", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&models.PredictionJob{
					ID:              models.ID(1),
					Name:            "prediction-job-1",
					ProjectID:       models.ID(1),
					VersionID:       models.ID(1),
					VersionModelID:  models.ID(1),
					EnvironmentName: "dev",
				}, nil)
				svc.On("ListContainers", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("Connection refused"))
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error while getting containers for endpoint: Connection refused"},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			modelSvc := tC.modelService()
			versionSvc := tC.versionService()
			predictionJobSvc := tC.predictionJobService()
			envSvc := tC.envService()
			ctl := &PredictionJobController{
				AppContext: &AppContext{
					ModelsService:        modelSvc,
					VersionsService:      versionSvc,
					PredictionJobService: predictionJobSvc,
					EnvironmentService:   envSvc,
					MonitoringConfig: config.MonitoringConfig{
						MonitoringEnabled: true,
						MonitoringBaseURL: "http://grafana",
					},
					AlertEnabled: true,
				},
			}
			resp := ctl.ListContainers(&http.Request{}, tC.vars, nil)
			assertEqualResponses(t, tC.expected, resp)
		})
	}
}

func TestCreate(t *testing.T) {
	trueBoolean := true
	testCases := []struct {
		desc                 string
		vars                 map[string]string
		requestBody          interface{}
		modelService         func() *mocks.ModelsService
		versionService       func() *mocks.VersionsService
		envService           func() *mocks.EnvironmentService
		predictionJobService func() *mocks.PredictionJobService
		expected             *Response
	}{
		{
			desc: "Should succcess create prediction job",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
				"job_id":     "1",
			},
			requestBody: &models.PredictionJob{
				Name:            "prediction-job-1",
				ProjectID:       models.ID(1),
				VersionID:       models.ID(1),
				VersionModelID:  models.ID(1),
				EnvironmentName: "dev",
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
				svc.On("GetDefaultPredictionJobEnvironment").Return(&models.Environment{
					ID:                     models.ID(1),
					Name:                   "dev",
					Cluster:                "dev",
					Region:                 "id",
					GcpProject:             "id-proj",
					IsPredictionJobEnabled: true,
					IsDefaultPredictionJob: &trueBoolean,
				}, nil)
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				svc.On("CreatePredictionJob", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&models.PredictionJob{
					ID:              models.ID(1),
					Name:            "prediction-job-1",
					ProjectID:       models.ID(1),
					VersionID:       models.ID(1),
					VersionModelID:  models.ID(1),
					EnvironmentName: "dev",
				}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusOK,
				data: &models.PredictionJob{
					ID:              models.ID(1),
					Name:            "prediction-job-1",
					ProjectID:       models.ID(1),
					VersionID:       models.ID(1),
					VersionModelID:  models.ID(1),
					EnvironmentName: "dev",
				},
			},
		},
		{
			desc: "Should return 500 if error fetching model",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(nil, fmt.Errorf("Error creating secret: db is down"))
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				return svc
			},
			envService: func() *mocks.EnvironmentService {
				svc := &mocks.EnvironmentService{}
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error getting model / version: error retrieving model with id: 1"},
			},
		},
		{
			desc: "Should return 500 if get default env prediction job returning error",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.PredictionJob{
				Name:            "prediction-job-1",
				ProjectID:       models.ID(1),
				VersionID:       models.ID(1),
				VersionModelID:  models.ID(1),
				EnvironmentName: "dev",
			}, modelService: func() *mocks.ModelsService {
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
				svc.On("GetDefaultPredictionJobEnvironment").Return(nil, fmt.Errorf("Error creating secret: db is down"))
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Unable to find default environment, specify environment target for deployment: Error creating secret: db is down"},
			},
		},
		{
			desc: "Should return 400 if error when create prediction job",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.PredictionJob{
				Name:            "prediction-job-1",
				ProjectID:       models.ID(1),
				VersionID:       models.ID(1),
				VersionModelID:  models.ID(1),
				EnvironmentName: "dev",
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
				svc.On("GetDefaultPredictionJobEnvironment").Return(&models.Environment{
					ID:                     models.ID(1),
					Name:                   "dev",
					Cluster:                "dev",
					Region:                 "id",
					GcpProject:             "id-proj",
					IsPredictionJobEnabled: true,
					IsDefaultPredictionJob: &trueBoolean,
				}, nil)
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				svc.On("CreatePredictionJob", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("Connection refused"))
				return svc
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{Message: "Error creating prediction job: Connection refused"},
			},
		},
		{
			desc: "Should return 400 if error request body is not valid",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
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
			requestBody: &models.Environment{
				Name:    "env",
				Cluster: "env",
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
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				return svc
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{Message: "Unable to parse body as prediction job"},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			modelSvc := tC.modelService()
			versionSvc := tC.versionService()
			predictionJobSvc := tC.predictionJobService()
			envSvc := tC.envService()
			ctl := &PredictionJobController{
				AppContext: &AppContext{
					ModelsService:        modelSvc,
					VersionsService:      versionSvc,
					PredictionJobService: predictionJobSvc,
					EnvironmentService:   envSvc,
					MonitoringConfig: config.MonitoringConfig{
						MonitoringEnabled: true,
						MonitoringBaseURL: "http://grafana",
					},
					AlertEnabled: true,
				},
			}
			resp := ctl.Create(&http.Request{}, tC.vars, tC.requestBody)
			assertEqualResponses(t, tC.expected, resp)
		})
	}
}

func TestListAllInProject(t *testing.T) {
	testCases := []struct {
		desc                 string
		request              *http.Request
		vars                 map[string]string
		projectService       func() *mocks.ProjectsService
		predictionJobService func() *mocks.PredictionJobService
		expected             *Response
	}{
		{
			desc: "Should success list of prediction job",
			request: &http.Request{URL: &url.URL{
				RawQuery: "name=prediction-job&model_id=1",
			}},
			vars: map[string]string{
				"project_id": "1",
			},
			projectService: func() *mocks.ProjectsService {
				svc := &mocks.ProjectsService{}
				svc.On("GetByID", mock.Anything, int32(1)).Return(mlp.Project(client.Project{
					ID:                1,
					Name:              "project-1",
					MLFlowTrackingURL: "http://mlflow.com",
					Team:              "dsp",
					Stream:            "dsp",
				}), nil)
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				svc.On("ListPredictionJobs", context.Background(), mock.Anything, &service.ListPredictionJobQuery{
					Name:    "prediction-job",
					ModelID: models.ID(1),
				}).Return([]*models.PredictionJob{
					{
						ID:              models.ID(1),
						Name:            "prediction-job-1",
						ProjectID:       models.ID(1),
						VersionID:       models.ID(1),
						VersionModelID:  models.ID(1),
						EnvironmentName: "dev",
					},
				}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusOK,
				data: []*models.PredictionJob{
					{
						ID:              models.ID(1),
						Name:            "prediction-job-1",
						ProjectID:       models.ID(1),
						VersionID:       models.ID(1),
						VersionModelID:  models.ID(1),
						EnvironmentName: "dev",
					},
				},
			},
		},
		{
			desc: "Should return 404 when error fetching project",
			request: &http.Request{URL: &url.URL{
				RawQuery: "name=prediction-job&model_id=1",
			}},
			vars: map[string]string{
				"project_id": "1",
			},
			projectService: func() *mocks.ProjectsService {
				svc := &mocks.ProjectsService{}
				svc.On("GetByID", mock.Anything, int32(1)).Return(mlp.Project{}, fmt.Errorf("API is down"))
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				return svc
			},
			expected: &Response{
				code: http.StatusNotFound,
				data: Error{Message: "Project not found: API is down"},
			},
		},
		{
			desc: "Should return 500 when error get list of prediction jobs",
			request: &http.Request{URL: &url.URL{
				RawQuery: "name=prediction-job&model_id=1",
			}},
			vars: map[string]string{
				"project_id": "1",
			},
			projectService: func() *mocks.ProjectsService {
				svc := &mocks.ProjectsService{}
				svc.On("GetByID", mock.Anything, int32(1)).Return(mlp.Project(client.Project{
					ID:                1,
					Name:              "project-1",
					MLFlowTrackingURL: "http://mlflow.com",
					Team:              "dsp",
					Stream:            "dsp",
				}), nil)
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				svc.On("ListPredictionJobs", context.Background(), mock.Anything, &service.ListPredictionJobQuery{
					Name:    "prediction-job",
					ModelID: models.ID(1),
				}).Return(nil, fmt.Errorf("Error creating secret: db is down"))
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error listing prediction jobs: Error creating secret: db is down"},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			projectSvc := tC.projectService()
			predictionJobSvc := tC.predictionJobService()
			ctl := &PredictionJobController{
				AppContext: &AppContext{
					ProjectsService:      projectSvc,
					PredictionJobService: predictionJobSvc,
					MonitoringConfig: config.MonitoringConfig{
						MonitoringEnabled: true,
						MonitoringBaseURL: "http://grafana",
					},
					AlertEnabled: true,
				},
			}
			resp := ctl.ListAllInProject(tC.request, tC.vars, nil)
			assertEqualResponses(t, tC.expected, resp)
		})
	}
}
