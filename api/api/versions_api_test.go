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
	"net/url"
	"testing"

	"github.com/caraml-dev/merlin/service"
	"github.com/caraml-dev/merlin/webhook"
	webhookMock "github.com/caraml-dev/merlin/webhook/mocks"
	"github.com/google/uuid"

	"github.com/caraml-dev/merlin/config"
	"github.com/caraml-dev/merlin/mlflow"
	mlfmocks "github.com/caraml-dev/merlin/mlflow/mocks"
	"github.com/caraml-dev/merlin/mlp"
	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/service/mocks"
	mlflowDeleteServiceMocks "github.com/caraml-dev/mlp/api/pkg/client/mlflow/mocks"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestGetVersion(t *testing.T) {
	testCases := []struct {
		desc           string
		vars           map[string]string
		versionService func() *mocks.VersionsService
		expected       *Response
	}{
		{
			desc: "Should success get version",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
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
						MlflowURL:    "http://mlflow.com",
					},
					MlflowURL: "http://mlflow.com",
				}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusOK,
				data: &models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "http://mlflow.com",
					},
					MlflowURL: "http://mlflow.com",
				},
			},
		},
		{
			desc: "Should return 404 if version is not found",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(nil, gorm.ErrRecordNotFound)
				return svc
			},
			expected: &Response{
				code: http.StatusNotFound,
				data: Error{Message: "Model version not found: record not found"},
			},
		},
		{
			desc: "Should return 500 if error when fetching version",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(nil, fmt.Errorf("Error creating secret: db is down"))
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error getting model version: Error creating secret: db is downv"},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			versionSvc := tC.versionService()

			ctl := &VersionsController{
				AppContext: &AppContext{
					VersionsService: versionSvc,
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
			}
			resp := ctl.GetVersion(&http.Request{}, tC.vars, nil)
			assertEqualResponses(t, tC.expected, resp)
		})
	}
}

func TestListVersion(t *testing.T) {
	testCases := []struct {
		desc           string
		vars           map[string]string
		versionService func() *mocks.VersionsService
		queryParameter string
		expected       *Response
	}{
		{
			desc: "Should success get version",
			vars: map[string]string{
				"model_id": "1",
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
							Type:         "pyfunc",
							MlflowURL:    "http://mlflow.com",
						},
						MlflowURL: "http://mlflow.com",
					},
				}, "", nil)
				return svc
			},
			expected: &Response{
				code: http.StatusOK,
				data: []*models.Version{
					{
						ID:      models.ID(1),
						ModelID: models.ID(1),
						Model: &models.Model{
							ID:           models.ID(1),
							Name:         "model-1",
							ProjectID:    models.ID(1),
							Project:      mlp.Project{},
							ExperimentID: 1,
							Type:         "pyfunc",
							MlflowURL:    "http://mlflow.com",
						},
						MlflowURL: "http://mlflow.com",
					},
				},
				headers: map[string]string{},
			},
		},
		{
			desc: "Should success get version with pagination",
			vars: map[string]string{
				"model_id": "1",
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
							Type:         "pyfunc",
							MlflowURL:    "http://mlflow.com",
						},
						MlflowURL: "http://mlflow.com",
					},
				}, "NDdfMzQ=", nil)
				return svc
			},
			queryParameter: "limit=30",
			expected: &Response{
				code: http.StatusOK,
				data: []*models.Version{
					{
						ID:      models.ID(1),
						ModelID: models.ID(1),
						Model: &models.Model{
							ID:           models.ID(1),
							Name:         "model-1",
							ProjectID:    models.ID(1),
							Project:      mlp.Project{},
							ExperimentID: 1,
							Type:         "pyfunc",
							MlflowURL:    "http://mlflow.com",
						},
						MlflowURL: "http://mlflow.com",
					},
				},
				headers: map[string]string{
					"Next-Cursor": "NDdfMzQ=",
				},
			},
		},
		{
			desc: "Should return 500 if get version returning error",
			vars: map[string]string{
				"model_id": "1",
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("ListVersions", mock.Anything, models.ID(1), mock.Anything, mock.Anything).Return(nil, "", fmt.Errorf("Error creating secret: db is down"))
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error listing versions for model: Error creating secret: db is down"},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			versionSvc := tC.versionService()

			ctl := &VersionsController{
				AppContext: &AppContext{
					VersionsService: versionSvc,
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
			}
			resp := ctl.ListVersions(&http.Request{URL: &url.URL{RawQuery: tC.queryParameter}}, tC.vars, nil)
			assertEqualResponses(t, tC.expected, resp)
		})
	}
}

func TestPatchVersion(t *testing.T) {
	testCases := []struct {
		desc           string
		requestBody    interface{}
		vars           map[string]string
		versionService func() *mocks.VersionsService
		webhook        func() *webhookMock.Client
		expected       *Response
	}{
		{
			desc: "Should success patch version",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.VersionPatch{Properties: &models.KV{
				"name":       "model-1",
				"created_by": "anonymous",
			}},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(
					&models.Version{
						ID:      models.ID(1),
						ModelID: models.ID(1),
						Model: &models.Model{
							ID:           models.ID(1),
							Name:         "model-1",
							ProjectID:    models.ID(1),
							Project:      mlp.Project{},
							ExperimentID: 1,
							Type:         "pyfunc",
							MlflowURL:    "http://mlflow.com",
						},
						MlflowURL: "http://mlflow.com",
					}, nil)
				svc.On("Save", mock.Anything, &models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "http://mlflow.com",
					},
					MlflowURL: "http://mlflow.com",
					Properties: models.KV{
						"name":       "model-1",
						"created_by": "anonymous",
					},
				}, mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "http://mlflow.com",
					},
					MlflowURL: "http://mlflow.com",
					Properties: models.KV{
						"name":       "model-1",
						"created_by": "anonymous",
					},
				}, nil)
				return svc
			},
			webhook: func() *webhookMock.Client {
				w := webhookMock.NewClient(t)
				w.On("TriggerModelVersionEvent", mock.Anything, webhook.OnModelVersionUpdated, mock.Anything).Return(nil)
				return w
			},
			expected: &Response{
				code: http.StatusOK,
				data: &models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "http://mlflow.com",
					},
					MlflowURL: "http://mlflow.com",
					Properties: models.KV{
						"name":       "model-1",
						"created_by": "anonymous",
					},
				},
			},
		},
		{
			desc: "Should success patch version - patch customer container if model type is CUSTOM",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.VersionPatch{
				CustomPredictor: &models.CustomPredictor{
					Image:   "gcr.io/custom-predictor:v0.1",
					Command: "./run.sh",
					Args:    "firstArg secondArg",
				},
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(
					&models.Version{
						ID:      models.ID(1),
						ModelID: models.ID(1),
						Model: &models.Model{
							ID:           models.ID(1),
							Name:         "model-1",
							ProjectID:    models.ID(1),
							Project:      mlp.Project{},
							ExperimentID: 1,
							Type:         "pyfunc",
							MlflowURL:    "http://mlflow.com",
						},
						MlflowURL:       "http://mlflow.com",
						CustomPredictor: nil,
					}, nil)
				svc.On("Save", mock.Anything, &models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "http://mlflow.com",
					},
					MlflowURL: "http://mlflow.com",
				}, mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "http://mlflow.com",
					},
					MlflowURL: "http://mlflow.com",
					CustomPredictor: &models.CustomPredictor{
						Image:   "gcr.io/custom-predictor:v0.1",
						Command: "./run.sh",
						Args:    "firstArg secondArg",
					},
				}, nil)
				return svc
			},
			webhook: func() *webhookMock.Client {
				w := webhookMock.NewClient(t)
				w.On("TriggerModelVersionEvent", mock.Anything, webhook.OnModelVersionUpdated, mock.Anything).Return(nil)
				return w
			},
			expected: &Response{
				code: http.StatusOK,
				data: &models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "http://mlflow.com",
					},
					MlflowURL: "http://mlflow.com",
					CustomPredictor: &models.CustomPredictor{
						Image:   "gcr.io/custom-predictor:v0.1",
						Command: "./run.sh",
						Args:    "firstArg secondArg",
					},
				},
			},
		},
		{
			desc: "Should success patch version - do nothing when trying patch custom_container where its type is not CUSTOM",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.VersionPatch{
				CustomPredictor: &models.CustomPredictor{
					Image:   "gcr.io/custom-predictor:v0.1",
					Command: "./run.sh",
					Args:    "firstArg secondArg",
				},
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(
					&models.Version{
						ID:      models.ID(1),
						ModelID: models.ID(1),
						Model: &models.Model{
							ID:           models.ID(1),
							Name:         "model-1",
							ProjectID:    models.ID(1),
							Project:      mlp.Project{},
							ExperimentID: 1,
							Type:         "pyfunc",
							MlflowURL:    "http://mlflow.com",
						},
						MlflowURL:       "http://mlflow.com",
						CustomPredictor: nil,
					}, nil)
				svc.On("Save", mock.Anything, &models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "http://mlflow.com",
					},
					MlflowURL: "http://mlflow.com",
				}, mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "http://mlflow.com",
					},
					MlflowURL: "http://mlflow.com",
				}, nil)
				return svc
			},
			webhook: func() *webhookMock.Client {
				w := webhookMock.NewClient(t)
				w.On("TriggerModelVersionEvent", mock.Anything, webhook.OnModelVersionUpdated, mock.Anything).Return(nil)
				return w
			},
			expected: &Response{
				code: http.StatusOK,
				data: &models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "http://mlflow.com",
					},
					MlflowURL: "http://mlflow.com",
				},
			},
		},
		{
			desc: "Return 400 patch version - model type custom but custom predictor object doesn't have image",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.VersionPatch{
				CustomPredictor: &models.CustomPredictor{
					Command: "./run.sh",
					Args:    "firstArg secondArg",
				},
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(
					&models.Version{
						ID:      models.ID(1),
						ModelID: models.ID(1),
						Model: &models.Model{
							ID:           models.ID(1),
							Name:         "model-1",
							ProjectID:    models.ID(1),
							Project:      mlp.Project{},
							ExperimentID: 1,
							Type:         "custom",
							MlflowURL:    "http://mlflow.com",
						},
						MlflowURL:       "http://mlflow.com",
						CustomPredictor: nil,
					}, nil)
				return svc
			},
			webhook: func() *webhookMock.Client {
				return webhookMock.NewClient(t)
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{Message: "Error validating version: custom predictor image must be set"},
			},
		},
		{
			desc: "Should return 404 if version is not found",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.VersionPatch{Properties: &models.KV{
				"name":       "model-1",
				"created_by": "anonymous",
			}},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(
					nil, gorm.ErrRecordNotFound)
				return svc
			},
			webhook: func() *webhookMock.Client {
				return webhookMock.NewClient(t)
			},
			expected: &Response{
				code: http.StatusNotFound,
				data: Error{Message: "Model version not found: record not found"},
			},
		},
		{
			desc: "Should return 500 if version fetching returning error",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.VersionPatch{Properties: &models.KV{
				"name":       "model-1",
				"created_by": "anonymous",
			}},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(
					nil, fmt.Errorf("Error creating secret: db is down"))
				return svc
			},
			webhook: func() *webhookMock.Client {
				return webhookMock.NewClient(t)
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error getting model version: Error creating secret: db is down"},
			},
		},
		{
			desc: "Should return 500 if request body is not valid",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.Model{
				ID: models.ID(1),
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(
					&models.Version{
						ID:      models.ID(1),
						ModelID: models.ID(1),
						Model: &models.Model{
							ID:           models.ID(1),
							Name:         "model-1",
							ProjectID:    models.ID(1),
							Project:      mlp.Project{},
							ExperimentID: 1,
							Type:         "pyfunc",
							MlflowURL:    "http://mlflow.com",
						},
						MlflowURL: "http://mlflow.com",
					}, nil)
				return svc
			},
			webhook: func() *webhookMock.Client {
				return webhookMock.NewClient(t)
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Unable to parse request body"},
			},
		},
		{
			desc: "Should return 500 if save is failing",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.VersionPatch{Properties: &models.KV{
				"name":       "model-1",
				"created_by": "anonymous",
			}},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(
					&models.Version{
						ID:      models.ID(1),
						ModelID: models.ID(1),
						Model: &models.Model{
							ID:           models.ID(1),
							Name:         "model-1",
							ProjectID:    models.ID(1),
							Project:      mlp.Project{},
							ExperimentID: 1,
							Type:         "pyfunc",
							MlflowURL:    "http://mlflow.com",
						},
						MlflowURL: "http://mlflow.com",
					}, nil)
				svc.On("Save", mock.Anything, &models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "http://mlflow.com",
					},
					MlflowURL: "http://mlflow.com",
					Properties: models.KV{
						"name":       "model-1",
						"created_by": "anonymous",
					},
				}, mock.Anything).Return(nil, fmt.Errorf("Error creating secret: db is down"))
				return svc
			},
			webhook: func() *webhookMock.Client {
				return webhookMock.NewClient(t)
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error patching model version: Error creating secret: db is down"},
			},
		},
		{
			desc: "Should success update model schema",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.VersionPatch{
				Properties: &models.KV{
					"name":       "model-1",
					"created_by": "anonymous",
				},
				ModelSchema: &models.ModelSchema{
					Spec: &models.SchemaSpec{
						SessionIDColumn: "session_id",
						RowIDColumn:     "row_id",
						ModelPredictionOutput: &models.ModelPredictionOutput{
							RankingOutput: &models.RankingOutput{
								RankScoreColumn:      "score",
								RelevanceScoreColumn: "relevance_score",
								OutputClass:          models.Ranking,
							},
						},
						FeatureTypes: map[string]models.ValueType{
							"featureA": models.Float64,
							"featureB": models.Int64,
							"featureC": models.Boolean,
						},
					},
					ModelID: models.ID(1),
				},
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(
					&models.Version{
						ID:      models.ID(1),
						ModelID: models.ID(1),
						Model: &models.Model{
							ID:           models.ID(1),
							Name:         "model-1",
							ProjectID:    models.ID(1),
							Project:      mlp.Project{},
							ExperimentID: 1,
							Type:         "pyfunc",
							MlflowURL:    "http://mlflow.com",
						},
						MlflowURL: "http://mlflow.com",
					}, nil)
				svc.On("Save", mock.Anything, &models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "http://mlflow.com",
					},
					MlflowURL: "http://mlflow.com",
					Properties: models.KV{
						"name":       "model-1",
						"created_by": "anonymous",
					},
					ModelSchema: &models.ModelSchema{
						Spec: &models.SchemaSpec{
							SessionIDColumn: "session_id",
							RowIDColumn:     "row_id",
							ModelPredictionOutput: &models.ModelPredictionOutput{
								RankingOutput: &models.RankingOutput{
									RankScoreColumn:      "score",
									RelevanceScoreColumn: "relevance_score",
									OutputClass:          models.Ranking,
								},
							},
							FeatureTypes: map[string]models.ValueType{
								"featureA": models.Float64,
								"featureB": models.Int64,
								"featureC": models.Boolean,
							},
						},
						ModelID: models.ID(1),
					},
				}, mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "http://mlflow.com",
					},
					MlflowURL: "http://mlflow.com",
					Properties: models.KV{
						"name":       "model-1",
						"created_by": "anonymous",
					},
					ModelSchema: &models.ModelSchema{
						Spec: &models.SchemaSpec{
							SessionIDColumn: "session_id",
							RowIDColumn:     "row_id",
							ModelPredictionOutput: &models.ModelPredictionOutput{
								RankingOutput: &models.RankingOutput{
									RankScoreColumn:      "score",
									RelevanceScoreColumn: "relevance_score",
									OutputClass:          models.Ranking,
								},
							},
							FeatureTypes: map[string]models.ValueType{
								"featureA": models.Float64,
								"featureB": models.Int64,
								"featureC": models.Boolean,
							},
						},
						ModelID: models.ID(1),
					},
				}, nil)
				return svc
			},
			webhook: func() *webhookMock.Client {
				w := webhookMock.NewClient(t)
				w.On("TriggerModelVersionEvent", mock.Anything, webhook.OnModelVersionUpdated, mock.Anything).Return(nil)
				return w
			},
			expected: &Response{
				code: http.StatusOK,
				data: &models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "http://mlflow.com",
					},
					MlflowURL: "http://mlflow.com",
					Properties: models.KV{
						"name":       "model-1",
						"created_by": "anonymous",
					},
					ModelSchema: &models.ModelSchema{
						Spec: &models.SchemaSpec{
							SessionIDColumn: "session_id",
							RowIDColumn:     "row_id",
							ModelPredictionOutput: &models.ModelPredictionOutput{
								RankingOutput: &models.RankingOutput{
									RankScoreColumn:      "score",
									RelevanceScoreColumn: "relevance_score",
									OutputClass:          models.Ranking,
								},
							},
							FeatureTypes: map[string]models.ValueType{
								"featureA": models.Float64,
								"featureB": models.Int64,
								"featureC": models.Boolean,
							},
						},
						ModelID: models.ID(1),
					},
				},
			},
		},
		{
			desc: "Should fail update model schema when there is mismatch of model id",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.VersionPatch{
				Properties: &models.KV{
					"name":       "model-1",
					"created_by": "anonymous",
				},
				ModelSchema: &models.ModelSchema{
					Spec: &models.SchemaSpec{
						SessionIDColumn: "session_id",
						RowIDColumn:     "row_id",
						ModelPredictionOutput: &models.ModelPredictionOutput{
							RankingOutput: &models.RankingOutput{
								RankScoreColumn:      "score",
								RelevanceScoreColumn: "relevance_score",
								OutputClass:          models.Ranking,
							},
						},
						FeatureTypes: map[string]models.ValueType{
							"featureA": models.Float64,
							"featureB": models.Int64,
							"featureC": models.Boolean,
						},
					},
					ModelID: models.ID(5),
				},
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(
					&models.Version{
						ID:      models.ID(1),
						ModelID: models.ID(1),
						Model: &models.Model{
							ID:           models.ID(1),
							Name:         "model-1",
							ProjectID:    models.ID(1),
							Project:      mlp.Project{},
							ExperimentID: 1,
							Type:         "pyfunc",
							MlflowURL:    "http://mlflow.com",
						},
						MlflowURL: "http://mlflow.com",
					}, nil)
				return svc
			},
			webhook: func() *webhookMock.Client {
				return webhookMock.NewClient(t)
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{
					Message: "Error validating version: mismatch model id between version and model schema",
				},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			versionSvc := tC.versionService()
			webhook := tC.webhook()

			ctl := &VersionsController{
				AppContext: &AppContext{
					VersionsService: versionSvc,
					FeatureToggleConfig: config.FeatureToggleConfig{
						AlertConfig: config.AlertConfig{
							AlertEnabled: true,
						},
						MonitoringConfig: config.MonitoringConfig{
							MonitoringEnabled: true,
							MonitoringBaseURL: "http://grafana",
						},
					},
					Webhook: webhook,
				},
			}
			resp := ctl.PatchVersion(&http.Request{}, tC.vars, tC.requestBody)
			assertEqualResponses(t, tC.expected, resp)
		})
	}
}

func TestCreateVersion(t *testing.T) {
	testCases := []struct {
		desc           string
		vars           map[string]string
		body           models.VersionPost
		versionService func() *mocks.VersionsService
		mlflowClient   func() *mlfmocks.Client
		modelsService  func() *mocks.ModelsService
		webhook        func() *webhookMock.Client
		expected       *Response
	}{
		{
			desc: "Should successfully create version",
			vars: map[string]string{
				"model_id": "1",
			},
			body: models.VersionPost{
				Labels: models.KV{
					"service.type":     "GO-FOOD",
					"1-targeting_date": "2021-02-01",
					"TheQuickBrownFoxJumpsOverTheLazyDogTheQuickBrownFoxJumpsOverThe": "TheQuickBrownFoxJumpsOverTheLazyDogTheQuickBrownFoxJumpsOverThe",
				},
				PythonVersion: "3.10.*",
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
					Type:         "pyfunc",
					MlflowURL:    "http://mlflow.com",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			mlflowClient: func() *mlfmocks.Client {
				svc := &mlfmocks.Client{}
				svc.On("CreateRun", "1").Return(&mlflow.Run{
					Info: mlflow.Info{
						RunID:       "1",
						ArtifactURI: "artifact/url/run",
					},
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("Save", mock.Anything, &models.Version{
					ModelID:     models.ID(1),
					RunID:       "1",
					ArtifactURI: "artifact/url/run",
					Labels: models.KV{
						"service.type":     "GO-FOOD",
						"1-targeting_date": "2021-02-01",
						"TheQuickBrownFoxJumpsOverTheLazyDogTheQuickBrownFoxJumpsOverThe": "TheQuickBrownFoxJumpsOverTheLazyDogTheQuickBrownFoxJumpsOverThe",
					},
					PythonVersion: "3.10.*",
				}, mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "sklearn",
						MlflowURL:    "http://mlflow.com",
					},
					MlflowURL: "http://mlflow.com",
					Labels: models.KV{
						"service.type":     "GO-FOOD",
						"1-targeting_date": "2021-02-01",
						"TheQuickBrownFoxJumpsOverTheLazyDogTheQuickBrownFoxJumpsOverThe": "TheQuickBrownFoxJumpsOverTheLazyDogTheQuickBrownFoxJumpsOverThe",
					},
					PythonVersion: "3.10.*",
				}, nil)
				return svc
			},
			webhook: func() *webhookMock.Client {
				w := webhookMock.NewClient(t)
				w.On("TriggerModelVersionEvent", mock.Anything, webhook.OnModelVersionCreated, mock.Anything).Return(nil)
				return w
			},
			expected: &Response{
				code: http.StatusCreated,
				data: &models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "sklearn",
						MlflowURL:    "http://mlflow.com",
					},
					MlflowURL: "http://mlflow.com",
					Labels: models.KV{
						"service.type":     "GO-FOOD",
						"1-targeting_date": "2021-02-01",
						"TheQuickBrownFoxJumpsOverTheLazyDogTheQuickBrownFoxJumpsOverThe": "TheQuickBrownFoxJumpsOverTheLazyDogTheQuickBrownFoxJumpsOverThe",
					},
					PythonVersion: "3.10.*",
				},
			},
		},
		{
			desc: "Should fail create version when save model returning error",
			vars: map[string]string{
				"model_id": "1",
			},
			body: models.VersionPost{
				Labels: models.KV{
					"service.type":     "GO-FOOD",
					"1-targeting_date": "2021-02-01",
					"TheQuickBrownFoxJumpsOverTheLazyDogTheQuickBrownFoxJumpsOverThe": "TheQuickBrownFoxJumpsOverTheLazyDogTheQuickBrownFoxJumpsOverThe",
				},
				PythonVersion: "3.10.*",
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
					Type:         "pyfunc",
					MlflowURL:    "http://mlflow.com",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			mlflowClient: func() *mlfmocks.Client {
				svc := &mlfmocks.Client{}
				svc.On("CreateRun", "1").Return(&mlflow.Run{
					Info: mlflow.Info{
						RunID:       "1",
						ArtifactURI: "artifact/url/run",
					},
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("Save", mock.Anything, &models.Version{
					ModelID:     models.ID(1),
					RunID:       "1",
					ArtifactURI: "artifact/url/run",
					Labels: models.KV{
						"service.type":     "GO-FOOD",
						"1-targeting_date": "2021-02-01",
						"TheQuickBrownFoxJumpsOverTheLazyDogTheQuickBrownFoxJumpsOverThe": "TheQuickBrownFoxJumpsOverTheLazyDogTheQuickBrownFoxJumpsOverThe",
					},
					PythonVersion: "3.10.*",
				}, mock.Anything).Return(nil, fmt.Errorf("pq constraint violation"))
				return svc
			},
			webhook: func() *webhookMock.Client {
				return webhookMock.NewClient(t)
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{
					Message: "Failed to save version: pq constraint violation",
				},
			},
		},
		{
			desc: "Should fail label key validation: has emoji inside",
			vars: map[string]string{
				"model_id": "1",
			},
			body: models.VersionPost{
				Labels: models.KV{
					"😊😊😊😊😊": "GO-FOOD",
				},
			},
			modelsService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{}, nil)
				return svc
			},
			mlflowClient: func() *mlfmocks.Client {
				svc := &mlfmocks.Client{}
				svc.On("CreateRun", "1").Return(&mlflow.Run{}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("Save", mock.Anything, &models.Version{}, mock.Anything).Return(&models.Version{}, nil)
				return svc
			},
			webhook: func() *webhookMock.Client {
				return webhookMock.NewClient(t)
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{Message: "Valid label key/values must be 63 characters or less and must be empty or begin and end with an alphanumeric character ([a-z0-9A-Z]) with dashes (-), underscores (_), dots (.), and alphanumerics between."},
			},
		},
		{
			desc: "Should fail label key validation: start with non alphanumeric",
			vars: map[string]string{
				"model_id": "1",
			},
			body: models.VersionPost{
				Labels: models.KV{
					"-service_type": "GO-FOOD",
				},
			},
			modelsService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{}, nil)
				return svc
			},
			mlflowClient: func() *mlfmocks.Client {
				svc := &mlfmocks.Client{}
				svc.On("CreateRun", "1").Return(&mlflow.Run{}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("Save", mock.Anything, &models.Version{}, mock.Anything).Return(&models.Version{}, nil)
				return svc
			},
			webhook: func() *webhookMock.Client {
				return webhookMock.NewClient(t)
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{Message: "Valid label key/values must be 63 characters or less and must be empty or begin and end with an alphanumeric character ([a-z0-9A-Z]) with dashes (-), underscores (_), dots (.), and alphanumerics between."},
			},
		},
		{
			desc: "Should fail label key validation: end with non alphanumeric",
			vars: map[string]string{
				"model_id": "1",
			},
			body: models.VersionPost{
				Labels: models.KV{
					"service_type-": "GO-FOOD",
				},
			},
			modelsService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{}, nil)
				return svc
			},
			mlflowClient: func() *mlfmocks.Client {
				svc := &mlfmocks.Client{}
				svc.On("CreateRun", "1").Return(&mlflow.Run{}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("Save", mock.Anything, &models.Version{}, mock.Anything).Return(&models.Version{}, nil)
				return svc
			},
			webhook: func() *webhookMock.Client {
				return webhookMock.NewClient(t)
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{Message: "Valid label key/values must be 63 characters or less and must be empty or begin and end with an alphanumeric character ([a-z0-9A-Z]) with dashes (-), underscores (_), dots (.), and alphanumerics between."},
			},
		},
		{
			desc: "Should fail label key validation: 64 characters",
			vars: map[string]string{
				"model_id": "1",
			},
			body: models.VersionPost{
				Labels: models.KV{
					"TheQuickBrownFoxJumpsOverTheLazyDogTheQuickBrownFoxJumpsOverTheL": "GO-FOOD",
				},
			},
			modelsService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{}, nil)
				return svc
			},
			mlflowClient: func() *mlfmocks.Client {
				svc := &mlfmocks.Client{}
				svc.On("CreateRun", "1").Return(&mlflow.Run{}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("Save", mock.Anything, &models.Version{}, mock.Anything).Return(&models.Version{}, nil)
				return svc
			},
			webhook: func() *webhookMock.Client {
				return webhookMock.NewClient(t)
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{Message: "Valid label key/values must be 63 characters or less and must be empty or begin and end with an alphanumeric character ([a-z0-9A-Z]) with dashes (-), underscores (_), dots (.), and alphanumerics between."},
			},
		},
		{
			desc: "Should fail label value validation: has emoji inside",
			vars: map[string]string{
				"model_id": "1",
			},
			body: models.VersionPost{
				Labels: models.KV{
					"emoji-label": "😊😊😊😊😊",
				},
			},
			modelsService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{}, nil)
				return svc
			},
			mlflowClient: func() *mlfmocks.Client {
				svc := &mlfmocks.Client{}
				svc.On("CreateRun", "1").Return(&mlflow.Run{}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("Save", mock.Anything, &models.Version{}, mock.Anything).Return(&models.Version{}, nil)
				return svc
			},
			webhook: func() *webhookMock.Client {
				return webhookMock.NewClient(t)
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{Message: "Valid label key/values must be 63 characters or less and must be empty or begin and end with an alphanumeric character ([a-z0-9A-Z]) with dashes (-), underscores (_), dots (.), and alphanumerics between."},
			},
		},
		{
			desc: "Should fail label value validation: start with non alphanumeric",
			vars: map[string]string{
				"model_id": "1",
			},
			body: models.VersionPost{
				Labels: models.KV{
					"service_type": "-GO-FOOD",
				},
			},
			modelsService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{}, nil)
				return svc
			},
			mlflowClient: func() *mlfmocks.Client {
				svc := &mlfmocks.Client{}
				svc.On("CreateRun", "1").Return(&mlflow.Run{}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("Save", mock.Anything, &models.Version{}, mock.Anything).Return(&models.Version{}, nil)
				return svc
			},
			webhook: func() *webhookMock.Client {
				return webhookMock.NewClient(t)
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{Message: "Valid label key/values must be 63 characters or less and must be empty or begin and end with an alphanumeric character ([a-z0-9A-Z]) with dashes (-), underscores (_), dots (.), and alphanumerics between."},
			},
		},
		{
			desc: "Should fail label value validation: end with non alphanumeric",
			vars: map[string]string{
				"model_id": "1",
			},
			body: models.VersionPost{
				Labels: models.KV{
					"service_type": "GO-FOOD-",
				},
			},
			modelsService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{}, nil)
				return svc
			},
			mlflowClient: func() *mlfmocks.Client {
				svc := &mlfmocks.Client{}
				svc.On("CreateRun", "1").Return(&mlflow.Run{}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("Save", mock.Anything, &models.Version{}, mock.Anything).Return(&models.Version{}, nil)
				return svc
			},
			webhook: func() *webhookMock.Client {
				return webhookMock.NewClient(t)
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{Message: "Valid label key/values must be 63 characters or less and must be empty or begin and end with an alphanumeric character ([a-z0-9A-Z]) with dashes (-), underscores (_), dots (.), and alphanumerics between."},
			},
		},
		{
			desc: "Should fail label value validation: 64 characters",
			vars: map[string]string{
				"model_id": "1",
			},
			body: models.VersionPost{
				Labels: models.KV{
					"some_valid_key": "TheQuickBrownFoxJumpsOverTheLazyDogTheQuickBrownFoxJumpsOverTheL",
				},
			},
			modelsService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{}, nil)
				return svc
			},
			mlflowClient: func() *mlfmocks.Client {
				svc := &mlfmocks.Client{}
				svc.On("CreateRun", "1").Return(&mlflow.Run{}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("Save", mock.Anything, &models.Version{}, mock.Anything).Return(&models.Version{}, nil)
				return svc
			},
			webhook: func() *webhookMock.Client {
				return webhookMock.NewClient(t)
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{Message: "Valid label key/values must be 63 characters or less and must be empty or begin and end with an alphanumeric character ([a-z0-9A-Z]) with dashes (-), underscores (_), dots (.), and alphanumerics between."},
			},
		},
		{
			desc: "Should successfully create version without labels, default python version",
			vars: map[string]string{
				"model_id": "1",
			},
			body: models.VersionPost{},
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
					Type:         "pyfunc",
					MlflowURL:    "http://mlflow.com",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			mlflowClient: func() *mlfmocks.Client {
				svc := &mlfmocks.Client{}
				svc.On("CreateRun", "1").Return(&mlflow.Run{
					Info: mlflow.Info{
						RunID:       "1",
						ArtifactURI: "artifact/url/run",
					},
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("Save", mock.Anything, &models.Version{
					ModelID:       models.ID(1),
					RunID:         "1",
					ArtifactURI:   "artifact/url/run",
					PythonVersion: DEFAULT_PYTHON_VERSION,
				}, mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "sklearn",
						MlflowURL:    "http://mlflow.com",
					},
					MlflowURL:     "http://mlflow.com",
					PythonVersion: DEFAULT_PYTHON_VERSION,
				}, nil)
				return svc
			},
			webhook: func() *webhookMock.Client {
				w := webhookMock.NewClient(t)
				w.On("TriggerModelVersionEvent", mock.Anything, webhook.OnModelVersionCreated, mock.Anything).Return(nil)
				return w
			},
			expected: &Response{
				code: http.StatusCreated,
				data: &models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "sklearn",
						MlflowURL:    "http://mlflow.com",
					},
					MlflowURL:     "http://mlflow.com",
					PythonVersion: DEFAULT_PYTHON_VERSION,
				},
			},
		},
		{
			desc: "Should successfully create version with model schema",
			vars: map[string]string{
				"model_id": "1",
			},
			body: models.VersionPost{
				ModelSchema: &models.ModelSchema{
					Spec: &models.SchemaSpec{
						SessionIDColumn: "session_id",
						RowIDColumn:     "row_id",
						ModelPredictionOutput: &models.ModelPredictionOutput{
							RankingOutput: &models.RankingOutput{
								RankScoreColumn:      "score",
								RelevanceScoreColumn: "relevance_score",
								OutputClass:          models.Ranking,
							},
						},
						FeatureTypes: map[string]models.ValueType{
							"featureA": models.Float64,
							"featureB": models.Int64,
							"featureC": models.Boolean,
						},
					},
					ModelID: models.ID(1),
				},
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
					Type:         "pyfunc",
					MlflowURL:    "http://mlflow.com",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			mlflowClient: func() *mlfmocks.Client {
				svc := &mlfmocks.Client{}
				svc.On("CreateRun", "1").Return(&mlflow.Run{
					Info: mlflow.Info{
						RunID:       "1",
						ArtifactURI: "artifact/url/run",
					},
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("Save", mock.Anything, &models.Version{
					ModelID:       models.ID(1),
					RunID:         "1",
					ArtifactURI:   "artifact/url/run",
					PythonVersion: DEFAULT_PYTHON_VERSION,
					ModelSchema: &models.ModelSchema{
						Spec: &models.SchemaSpec{
							SessionIDColumn: "session_id",
							RowIDColumn:     "row_id",
							ModelPredictionOutput: &models.ModelPredictionOutput{
								RankingOutput: &models.RankingOutput{
									RankScoreColumn:      "score",
									RelevanceScoreColumn: "relevance_score",
									OutputClass:          models.Ranking,
								},
							},
							FeatureTypes: map[string]models.ValueType{
								"featureA": models.Float64,
								"featureB": models.Int64,
								"featureC": models.Boolean,
							},
						},
						ModelID: models.ID(1),
					},
				}, mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "sklearn",
						MlflowURL:    "http://mlflow.com",
					},
					MlflowURL:     "http://mlflow.com",
					PythonVersion: DEFAULT_PYTHON_VERSION,
					ModelSchema: &models.ModelSchema{
						Spec: &models.SchemaSpec{
							SessionIDColumn: "session_id",
							RowIDColumn:     "row_id",
							ModelPredictionOutput: &models.ModelPredictionOutput{
								RankingOutput: &models.RankingOutput{
									RankScoreColumn:      "score",
									RelevanceScoreColumn: "relevance_score",
									OutputClass:          models.Ranking,
								},
							},
							FeatureTypes: map[string]models.ValueType{
								"featureA": models.Float64,
								"featureB": models.Int64,
								"featureC": models.Boolean,
							},
						},
						ModelID: models.ID(1),
					},
				}, nil)
				return svc
			},
			webhook: func() *webhookMock.Client {
				w := webhookMock.NewClient(t)
				w.On("TriggerModelVersionEvent", mock.Anything, webhook.OnModelVersionCreated, mock.Anything).Return(nil)
				return w
			},
			expected: &Response{
				code: http.StatusCreated,
				data: &models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "sklearn",
						MlflowURL:    "http://mlflow.com",
					},
					MlflowURL:     "http://mlflow.com",
					PythonVersion: DEFAULT_PYTHON_VERSION,
					ModelSchema: &models.ModelSchema{
						Spec: &models.SchemaSpec{
							SessionIDColumn: "session_id",
							RowIDColumn:     "row_id",
							ModelPredictionOutput: &models.ModelPredictionOutput{
								RankingOutput: &models.RankingOutput{
									RankScoreColumn:      "score",
									RelevanceScoreColumn: "relevance_score",
									OutputClass:          models.Ranking,
								},
							},
							FeatureTypes: map[string]models.ValueType{
								"featureA": models.Float64,
								"featureB": models.Int64,
								"featureC": models.Boolean,
							},
						},
						ModelID: models.ID(1),
					},
				},
			},
		},
		{
			desc: "Should fail create version with model schema, when there is mismatch model if between version and model schema",
			vars: map[string]string{
				"model_id": "1",
			},
			body: models.VersionPost{
				ModelSchema: &models.ModelSchema{
					Spec: &models.SchemaSpec{
						SessionIDColumn: "session_id",
						RowIDColumn:     "row_id",
						ModelPredictionOutput: &models.ModelPredictionOutput{
							RankingOutput: &models.RankingOutput{
								RankScoreColumn:      "score",
								RelevanceScoreColumn: "relevance_score",
								OutputClass:          models.Ranking,
							},
						},
						FeatureTypes: map[string]models.ValueType{
							"featureA": models.Float64,
							"featureB": models.Int64,
							"featureC": models.Boolean,
						},
					},
					ModelID: models.ID(5),
				},
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
					Type:         "pyfunc",
					MlflowURL:    "http://mlflow.com",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			mlflowClient: func() *mlfmocks.Client {
				svc := &mlfmocks.Client{}
				svc.On("CreateRun", "1").Return(&mlflow.Run{
					Info: mlflow.Info{
						RunID:       "1",
						ArtifactURI: "artifact/url/run",
					},
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				return svc
			},
			webhook: func() *webhookMock.Client {
				return webhookMock.NewClient(t)
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{
					Message: "Error validating version: mismatch model id between version and model schema",
				},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			versionSvc := tC.versionService()
			modelsSvc := tC.modelsService()
			mlflowClient := tC.mlflowClient()
			webhook := tC.webhook()

			ctl := &VersionsController{
				AppContext: &AppContext{
					VersionsService: versionSvc,
					FeatureToggleConfig: config.FeatureToggleConfig{
						AlertConfig: config.AlertConfig{
							AlertEnabled: true,
						},
						MonitoringConfig: config.MonitoringConfig{
							MonitoringEnabled: true,
							MonitoringBaseURL: "http://grafana",
						},
					},
					MlflowClient:  mlflowClient,
					ModelsService: modelsSvc,
					Webhook:       webhook,
				},
			}
			resp := ctl.CreateVersion(&http.Request{}, tC.vars, &tC.body)
			assertEqualResponses(t, tC.expected, resp)
		})
	}
}

func TestDeleteVersion(t *testing.T) {
	testCases := []struct {
		desc                 string
		vars                 map[string]string
		versionService       func() *mocks.VersionsService
		modelsService        func() *mocks.ModelsService
		mlflowDeleteService  func() *mlflowDeleteServiceMocks.Service
		predictionJobService func() *mocks.PredictionJobService
		endpointService      func() *mocks.EndpointsService
		webhook              func() *webhookMock.Client
		expected             *Response
	}{
		{
			desc: "Should successfully delete version type pyfunc_v2",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
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
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
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
				}, nil)
				svc.On("Delete", mock.Anything).Return(nil)
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				svc.On("ListPredictionJobs", mock.Anything, mock.Anything, &service.ListPredictionJobQuery{
					ModelID:   models.ID(1),
					VersionID: models.ID(1),
				}, false).Return([]*models.PredictionJob{}, nil, nil)
				return svc
			},
			mlflowDeleteService: func() *mlflowDeleteServiceMocks.Service {
				svc := &mlflowDeleteServiceMocks.Service{}
				svc.On("DeleteRun", mock.Anything, "runID1", mock.Anything, true).Return(nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("ListEndpoints", mock.Anything, mock.Anything, mock.Anything).Return([]*models.VersionEndpoint{}, nil)
				return svc
			},
			webhook: func() *webhookMock.Client {
				w := webhookMock.NewClient(t)
				w.On("TriggerModelVersionEvent", mock.Anything, webhook.OnModelVersionDeleted, mock.Anything).Return(nil)
				return w
			},
			expected: &Response{
				code: http.StatusOK,
				data: models.ID(1),
			},
		},
		{
			desc: "Should successfully delete version with type other than pyfunc_v2",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
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
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
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
				}, nil)
				svc.On("Delete", mock.Anything).Return(nil)
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("ListEndpoints", mock.Anything, mock.Anything, mock.Anything).Return([]*models.VersionEndpoint{}, nil)
				return svc
			},
			mlflowDeleteService: func() *mlflowDeleteServiceMocks.Service {
				svc := &mlflowDeleteServiceMocks.Service{}
				svc.On("DeleteRun", mock.Anything, "runID1", mock.Anything, true).Return(nil)
				return svc
			},
			webhook: func() *webhookMock.Client {
				w := webhookMock.NewClient(t)
				w.On("TriggerModelVersionEvent", mock.Anything, webhook.OnModelVersionDeleted, mock.Anything).Return(nil)
				return w
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
				"version_id": "1",
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
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
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
				}, nil)
				svc.On("Delete", mock.Anything).Return(nil)
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
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
			mlflowDeleteService: func() *mlflowDeleteServiceMocks.Service {
				svc := &mlflowDeleteServiceMocks.Service{}
				svc.On("DeleteRun", mock.Anything, "runID1", mock.Anything, true).Return(nil)
				return svc
			},
			webhook: func() *webhookMock.Client {
				return webhookMock.NewClient(t)
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
				"version_id": "1",
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
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
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
				}, nil)
				svc.On("Delete", mock.Anything).Return(nil)
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
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
			mlflowDeleteService: func() *mlflowDeleteServiceMocks.Service {
				svc := &mlflowDeleteServiceMocks.Service{}
				svc.On("DeleteRun", mock.Anything, "runID1", mock.Anything, true).Return(nil)
				return svc
			},
			webhook: func() *webhookMock.Client {
				return webhookMock.NewClient(t)
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
				"version_id": "1",
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
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
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
				}, nil)
				svc.On("Delete", mock.Anything).Return(nil)
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				svc.On("ListPredictionJobs", mock.Anything, mock.Anything, &service.ListPredictionJobQuery{
					ModelID:   models.ID(1),
					VersionID: models.ID(1),
				}, false).Return([]*models.PredictionJob{
					{
						ID:              models.ID(1),
						Name:            "prediction-job-1",
						ProjectID:       models.ID(1),
						VersionID:       models.ID(1),
						VersionModelID:  models.ID(1),
						EnvironmentName: "dev",
						Status:          models.JobRunning,
					},
				}, nil, nil)
				return svc
			},
			mlflowDeleteService: func() *mlflowDeleteServiceMocks.Service {
				svc := &mlflowDeleteServiceMocks.Service{}
				svc.On("DeleteRun", mock.Anything, "runID1", mock.Anything, true).Return(nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				return svc
			},
			webhook: func() *webhookMock.Client {
				return webhookMock.NewClient(t)
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{Message: "There are active prediction job that still using this model version"},
			},
		},
		{
			desc: "Should return 500 if failed to delete inactive Prediction jobs",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
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
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
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
				}, nil)
				svc.On("Delete", mock.Anything).Return(nil)
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				svc.On("ListPredictionJobs", mock.Anything, mock.Anything, &service.ListPredictionJobQuery{
					ModelID:   models.ID(1),
					VersionID: models.ID(1),
				}, false).Return([]*models.PredictionJob{
					{
						ID:              models.ID(1),
						Name:            "prediction-job-1",
						ProjectID:       models.ID(1),
						VersionID:       models.ID(1),
						VersionModelID:  models.ID(1),
						EnvironmentName: "dev",
						Status:          models.JobFailed,
					},
				}, nil, nil)
				svc.On("StopPredictionJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					nil, errors.New("failed to stop prediction job"))
				return svc
			},
			mlflowDeleteService: func() *mlflowDeleteServiceMocks.Service {
				svc := &mlflowDeleteServiceMocks.Service{}
				svc.On("DeleteRun", mock.Anything, "runID1", mock.Anything, true).Return(nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				return svc
			},
			webhook: func() *webhookMock.Client {
				return webhookMock.NewClient(t)
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Failed stopping prediction job: failed to stop prediction job"},
			},
		},
		{
			desc: "Should return 500 if delete mlflow run failed",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
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
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
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
				}, nil)
				svc.On("Delete", mock.Anything).Return(nil)
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				svc.On("ListPredictionJobs", mock.Anything, mock.Anything, &service.ListPredictionJobQuery{
					ModelID:   models.ID(1),
					VersionID: models.ID(1),
				}, false).Return([]*models.PredictionJob{}, nil, nil)
				return svc
			},
			mlflowDeleteService: func() *mlflowDeleteServiceMocks.Service {
				svc := &mlflowDeleteServiceMocks.Service{}
				svc.On("DeleteRun", mock.Anything, "runID1", mock.Anything, true).Return(errors.New("failed to delete mlflow run"))
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("ListEndpoints", mock.Anything, mock.Anything, mock.Anything).Return([]*models.VersionEndpoint{}, nil)
				return svc
			},
			webhook: func() *webhookMock.Client {
				return webhookMock.NewClient(t)
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Delete mlflow run failed: failed to delete mlflow run"},
			},
		},
		{
			desc: "Should return 500 if delete model version from database failed",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
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
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
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
				}, nil)
				svc.On("Delete", mock.Anything).Return(errors.New("failed to delete model version"))
				return svc
			},
			predictionJobService: func() *mocks.PredictionJobService {
				svc := &mocks.PredictionJobService{}
				svc.On("ListPredictionJobs", mock.Anything, mock.Anything, &service.ListPredictionJobQuery{
					ModelID:   models.ID(1),
					VersionID: models.ID(1),
				}, false).Return([]*models.PredictionJob{}, nil, nil)
				return svc
			},
			mlflowDeleteService: func() *mlflowDeleteServiceMocks.Service {
				svc := &mlflowDeleteServiceMocks.Service{}
				svc.On("DeleteRun", mock.Anything, "runID1", mock.Anything, true).Return(nil)
				return svc
			},
			endpointService: func() *mocks.EndpointsService {
				svc := &mocks.EndpointsService{}
				svc.On("ListEndpoints", mock.Anything, mock.Anything, mock.Anything).Return([]*models.VersionEndpoint{}, nil)
				return svc
			},
			webhook: func() *webhookMock.Client {
				return webhookMock.NewClient(t)
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Delete model version failed: failed to delete model version"},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			versionSvc := tC.versionService()
			modelsSvc := tC.modelsService()
			mlflowDeleteSvc := tC.mlflowDeleteService
			predictionJobSvc := tC.predictionJobService
			endpointService := tC.endpointService
			webhook := tC.webhook()

			ctl := &VersionsController{
				AppContext: &AppContext{
					VersionsService: versionSvc,
					FeatureToggleConfig: config.FeatureToggleConfig{
						AlertConfig: config.AlertConfig{
							AlertEnabled: true,
						},
						MonitoringConfig: config.MonitoringConfig{
							MonitoringEnabled: true,
							MonitoringBaseURL: "http://grafana",
						},
					},
					ModelsService:        modelsSvc,
					MlflowDeleteService:  mlflowDeleteSvc(),
					PredictionJobService: predictionJobSvc(),
					EndpointsService:     endpointService(),
					Webhook:              webhook,
				},
			}
			resp := ctl.DeleteVersion(&http.Request{}, tC.vars, nil)
			assertEqualResponses(t, tC.expected, resp)
		})
	}
}
