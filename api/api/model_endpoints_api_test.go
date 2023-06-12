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

	"github.com/caraml-dev/merlin/mlp"
	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/service/mocks"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/mock"
)

func TestListModelEndpointInProject(t *testing.T) {
	testCases := []struct {
		desc                 string
		vars                 map[string]string
		modelEndpointService func() *mocks.ModelEndpointsService
		expected             *Response
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
			expected: &Response{
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
			expected: &Response{
				code: http.StatusNotFound,
				data: Error{Message: "Model endpoints not found: record not found"},
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
				mockSvc.On("ListModelEndpointsInProject", mock.Anything, models.ID(1), "id").Return(nil, fmt.Errorf("Error creating secret: db is down"))
				return mockSvc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error listing model endpoints: Error creating secret: db is down"},
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
			assertEqualResponses(t, tC.expected, resp)
		})
	}
}

func TestListModelEndpoints(t *testing.T) {
	testCases := []struct {
		desc                 string
		vars                 map[string]string
		modelEndpointService func() *mocks.ModelEndpointsService
		expected             *Response
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
			expected: &Response{
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
			expected: &Response{
				code: http.StatusNotFound,
				data: Error{Message: "Model endpoints not found: record not found"},
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
				mockSvc.On("ListModelEndpoints", mock.Anything, models.ID(1)).Return(nil, fmt.Errorf("Error creating secret: db is down"))
				return mockSvc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error listing model endpoints: Error creating secret: db is down"},
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
			assertEqualResponses(t, tC.expected, resp)
		})
	}
}

func TestGetModelEndpoint(t *testing.T) {
	testCases := []struct {
		desc                 string
		vars                 map[string]string
		modelEndpointService func() *mocks.ModelEndpointsService
		expected             *Response
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
			expected: &Response{
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
			expected: &Response{
				code: http.StatusNotFound,
				data: Error{Message: "Model endpoint not found: record not found"},
			},
		},
		{
			desc: "Should return 500 if error when fetching list of model endpoint",
			vars: map[string]string{
				"model_endpoint_id": "1",
			},
			modelEndpointService: func() *mocks.ModelEndpointsService {
				mockSvc := &mocks.ModelEndpointsService{}
				mockSvc.On("FindByID", mock.Anything, models.ID(1)).Return(nil, fmt.Errorf("Error creating secret: db is down"))
				return mockSvc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error getting model endpoint: Error creating secret: db is down"},
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
			assertEqualResponses(t, tC.expected, resp)
		})
	}
}
