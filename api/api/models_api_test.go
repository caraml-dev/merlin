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
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

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
				data: Error{Message: "MLP API is down"},
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
			assert.Equal(t, tC.expected, resp)
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
				mockSvc.On("GetByID", mock.Anything, int32(1)).Return(mlp.Project(client.Project{}), fmt.Errorf("Project API is down"))
				return mockSvc
			},
			modelService: func() *mocks.ModelsService {
				mockSvc := &mocks.ModelsService{}
				return mockSvc
			},
			expected: &Response{
				code: http.StatusNotFound,
				data: Error{Message: "Project API is down"},
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
				data: Error{Message: "Model id 1 not found"},
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
				data: Error{Message: "DB is unreachable"},
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
			assert.Equal(t, tC.expected, resp)
		})
	}
}
