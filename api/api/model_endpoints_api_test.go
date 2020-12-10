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

	"github.com/gojek/merlin/mlp"
	"github.com/gojek/merlin/models"
	"github.com/gojek/merlin/service/mocks"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListModelEndpointInProject(t *testing.T) {
	testCases := []struct {
		desc                 string
		vars                 map[string]string
		modelEndpointService func() *mocks.ModelEndpointsService
		expected             *APIResponse
	}{
		{
			desc: "Should success list model endpoint",
			vars: map[string]string{
				"project_id": "1",
				"region":     "id",
			},
			modelEndpointService: func() *mocks.ModelEndpointsService {
				mockSvc := &mocks.ModelEndpointsService{}
				mockSvc.On("ListModelEndpointsInProject", mock.Anything, models.ID(1), "id").Return([]*models.ModelEndpoint{
					{
						ID:      models.ID(1),
						ModelID: models.ID(1),
						Model: &models.Model{
							ID:           models.ID(1),
							Name:         "Model-1",
							ProjectID:    models.ID(1),
							Project:      mlp.Project{},
							ExperimentID: 0,
							Type:         "pyfunc",
							MlflowURL:    "http://mlflow.com",
							Endpoints:    nil,
						},
					},
				}, nil)
				return mockSvc
			},
			expected: &APIResponse{
				code: http.StatusOK,
				data: []*models.ModelEndpoint{
					{
						ID:      models.ID(1),
						ModelID: models.ID(1),
						Model: &models.Model{
							ID:           models.ID(1),
							Name:         "Model-1",
							ProjectID:    models.ID(1),
							Project:      mlp.Project{},
							ExperimentID: 0,
							Type:         "pyfunc",
							MlflowURL:    "http://mlflow.com",
							Endpoints:    nil,
						},
					},
				},
			},
		},
		{
			desc: "Should return 404 if there is no model endpoint",
			vars: map[string]string{
				"project_id": "1",
				"region":     "id",
			},
			modelEndpointService: func() *mocks.ModelEndpointsService {
				mockSvc := &mocks.ModelEndpointsService{}
				mockSvc.On("ListModelEndpointsInProject", mock.Anything, models.ID(1), "id").Return(nil, gorm.ErrRecordNotFound)
				return mockSvc
			},
			expected: &APIResponse{
				code: http.StatusNotFound,
				data: Error{Message: "Model Endpoints for Project ID 1 not found"},
			},
		},
		{
			desc: "Should return 500 if error when fetching list of model endpoint",
			vars: map[string]string{
				"project_id": "1",
				"region":     "id",
			},
			modelEndpointService: func() *mocks.ModelEndpointsService {
				mockSvc := &mocks.ModelEndpointsService{}
				mockSvc.On("ListModelEndpointsInProject", mock.Anything, models.ID(1), "id").Return(nil, fmt.Errorf("DB is down"))
				return mockSvc
			},
			expected: &APIResponse{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error while getting Model Endpoints for Project ID 1"},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			mockSvc := tC.modelEndpointService()
			ctl := &ModelEndpointsController{
				AppContext: &AppContext{
					ModelEndpointsService: mockSvc,
				},
			}
			resp := ctl.ListModelEndpointInProject(&http.Request{}, tC.vars, nil)
			assert.Equal(t, tC.expected, resp)
		})
	}
}

func TestListModelEndpoints(t *testing.T) {
	testCases := []struct {
		desc                 string
		vars                 map[string]string
		modelEndpointService func() *mocks.ModelEndpointsService
		expected             *APIResponse
	}{
		{
			desc: "Should success list model endpoint",
			vars: map[string]string{
				"model_id": "1",
				"region":   "id",
			},
			modelEndpointService: func() *mocks.ModelEndpointsService {
				mockSvc := &mocks.ModelEndpointsService{}
				mockSvc.On("ListModelEndpoints", mock.Anything, models.ID(1)).Return([]*models.ModelEndpoint{
					{
						ID:      models.ID(1),
						ModelID: models.ID(1),
						Model: &models.Model{
							ID:           models.ID(1),
							Name:         "Model-1",
							ProjectID:    models.ID(1),
							Project:      mlp.Project{},
							ExperimentID: 0,
							Type:         "pyfunc",
							MlflowURL:    "http://mlflow.com",
							Endpoints:    nil,
						},
					},
				}, nil)
				return mockSvc
			},
			expected: &APIResponse{
				code: http.StatusOK,
				data: []*models.ModelEndpoint{
					{
						ID:      models.ID(1),
						ModelID: models.ID(1),
						Model: &models.Model{
							ID:           models.ID(1),
							Name:         "Model-1",
							ProjectID:    models.ID(1),
							Project:      mlp.Project{},
							ExperimentID: 0,
							Type:         "pyfunc",
							MlflowURL:    "http://mlflow.com",
							Endpoints:    nil,
						},
					},
				},
			},
		},
		{
			desc: "Should return 404 if there is no model endpoint",
			vars: map[string]string{
				"model_id": "1",
				"region":   "id",
			},
			modelEndpointService: func() *mocks.ModelEndpointsService {
				mockSvc := &mocks.ModelEndpointsService{}
				mockSvc.On("ListModelEndpoints", mock.Anything, models.ID(1)).Return(nil, gorm.ErrRecordNotFound)
				return mockSvc
			},
			expected: &APIResponse{
				code: http.StatusNotFound,
				data: Error{Message: "Model Endpoints for Model ID 1 not found"},
			},
		},
		{
			desc: "Should return 500 if error when fetching list of model endpoint",
			vars: map[string]string{
				"model_id": "1",
				"region":   "id",
			},
			modelEndpointService: func() *mocks.ModelEndpointsService {
				mockSvc := &mocks.ModelEndpointsService{}
				mockSvc.On("ListModelEndpoints", mock.Anything, models.ID(1)).Return(nil, fmt.Errorf("DB is down"))
				return mockSvc
			},
			expected: &APIResponse{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error while getting Model Endpoints for Model ID 1"},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			mockSvc := tC.modelEndpointService()
			ctl := &ModelEndpointsController{
				AppContext: &AppContext{
					ModelEndpointsService: mockSvc,
				},
			}
			resp := ctl.ListModelEndpoints(&http.Request{}, tC.vars, nil)
			assert.Equal(t, tC.expected, resp)
		})
	}
}

func TestGetModelEndpoint(t *testing.T) {
	testCases := []struct {
		desc                 string
		vars                 map[string]string
		modelEndpointService func() *mocks.ModelEndpointsService
		expected             *APIResponse
	}{
		{
			desc: "Should success get model endpoint",
			vars: map[string]string{
				"model_endpoint_id": "1",
			},
			modelEndpointService: func() *mocks.ModelEndpointsService {
				mockSvc := &mocks.ModelEndpointsService{}
				mockSvc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.ModelEndpoint{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "Model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 0,
						Type:         "pyfunc",
						MlflowURL:    "http://mlflow.com",
						Endpoints:    nil,
					},
				}, nil)
				return mockSvc
			},
			expected: &APIResponse{
				code: http.StatusOK,
				data: &models.ModelEndpoint{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "Model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 0,
						Type:         "pyfunc",
						MlflowURL:    "http://mlflow.com",
						Endpoints:    nil,
					},
				},
			},
		},
		{
			desc: "Should return 404 if there is no model endpoint",
			vars: map[string]string{
				"model_endpoint_id": "1",
			},
			modelEndpointService: func() *mocks.ModelEndpointsService {
				mockSvc := &mocks.ModelEndpointsService{}
				mockSvc.On("FindByID", mock.Anything, models.ID(1)).Return(nil, gorm.ErrRecordNotFound)
				return mockSvc
			},
			expected: &APIResponse{
				code: http.StatusNotFound,
				data: Error{Message: "Model endpoint with id 1 not found"},
			},
		},
		{
			desc: "Should return 500 if error when fetching list of model endpoint",
			vars: map[string]string{
				"model_endpoint_id": "1",
			},
			modelEndpointService: func() *mocks.ModelEndpointsService {
				mockSvc := &mocks.ModelEndpointsService{}
				mockSvc.On("FindByID", mock.Anything, models.ID(1)).Return(nil, fmt.Errorf("DB is down"))
				return mockSvc
			},
			expected: &APIResponse{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error while getting model endpoint with id 1"},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			mockSvc := tC.modelEndpointService()
			ctl := &ModelEndpointsController{
				AppContext: &AppContext{
					ModelEndpointsService: mockSvc,
				},
			}
			resp := ctl.GetModelEndpoint(&http.Request{}, tC.vars, nil)
			assert.Equal(t, tC.expected, resp)
		})
	}
}
