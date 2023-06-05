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
	"github.com/stretchr/testify/mock"

	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/service/mocks"
	"github.com/jinzhu/gorm"
)

func TestListTeams(t *testing.T) {
	testCases := []struct {
		desc     string
		vars     map[string]string
		service  func() *mocks.ModelEndpointAlertService
		expected *Response
	}{
		{
			desc: "Should success list teams",
			service: func() *mocks.ModelEndpointAlertService {
				svc := &mocks.ModelEndpointAlertService{}
				svc.On("ListTeams").Return([]string{
					"dsp",
				}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusOK,
				data: []string{"dsp"},
			},
		},
		{
			desc: "Should return 500 if error fetching teams",
			service: func() *mocks.ModelEndpointAlertService {
				svc := &mocks.ModelEndpointAlertService{}
				svc.On("ListTeams").Return(nil, fmt.Errorf("API is down"))
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error listing teams: API is down"},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			svc := tC.service()
			ctl := &AlertsController{
				AppContext: &AppContext{
					ModelEndpointAlertService: svc,
				},
			}
			resp := ctl.ListTeams(&http.Request{}, tC.vars, nil)
			assertEqualResponses(t, tC.expected, resp)
		})
	}
}

func TestListModelEndpointAlerts(t *testing.T) {
	testCases := []struct {
		desc     string
		vars     map[string]string
		service  func() *mocks.ModelEndpointAlertService
		expected *Response
	}{
		{
			desc: "Should success list model endpoint alerts",
			vars: map[string]string{
				"model_id": "1",
			},
			service: func() *mocks.ModelEndpointAlertService {
				svc := &mocks.ModelEndpointAlertService{}
				svc.On("ListModelAlerts", models.ID(1)).Return([]*models.ModelEndpointAlert{
					{
						ID:              models.ID(1),
						ModelID:         models.ID(1),
						ModelEndpointID: models.ID(1),
						EnvironmentName: "dev",
						AlertConditions: models.AlertConditions{
							{
								Enabled:    true,
								MetricType: models.AlertConditionTypeCPU,
								Severity:   models.AlertConditionSeverityCritical,
							},
						},
					},
				}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusOK,
				data: []*models.ModelEndpointAlert{
					{
						ID:              models.ID(1),
						ModelID:         models.ID(1),
						ModelEndpointID: models.ID(1),
						EnvironmentName: "dev",
						AlertConditions: models.AlertConditions{
							{
								Enabled:    true,
								MetricType: models.AlertConditionTypeCPU,
								Severity:   models.AlertConditionSeverityCritical,
							},
						},
					},
				},
			},
		},
		{
			desc: "Should return 500 if error fetching model alerts",
			vars: map[string]string{
				"model_id": "1",
			},
			service: func() *mocks.ModelEndpointAlertService {
				svc := &mocks.ModelEndpointAlertService{}
				svc.On("ListModelAlerts", models.ID(1)).Return(nil, fmt.Errorf("API is down"))
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error listing alerts for model: API is down"},
			},
		},
		{
			desc: "Should return 400 if there is no model alerts",
			vars: map[string]string{
				"model_id": "1",
			},
			service: func() *mocks.ModelEndpointAlertService {
				svc := &mocks.ModelEndpointAlertService{}
				svc.On("ListModelAlerts", models.ID(1)).Return(nil, gorm.ErrRecordNotFound)
				return svc
			},
			expected: &Response{
				code: http.StatusNotFound,
				data: Error{Message: "Model alert not found: record not found"},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			svc := tC.service()
			ctl := &AlertsController{
				AppContext: &AppContext{
					ModelEndpointAlertService: svc,
				},
			}
			resp := ctl.ListModelEndpointAlerts(&http.Request{}, tC.vars, nil)
			assertEqualResponses(t, tC.expected, resp)
		})
	}
}

func TestGetModelEndpointAlert(t *testing.T) {
	testCases := []struct {
		desc     string
		vars     map[string]string
		service  func() *mocks.ModelEndpointAlertService
		expected *Response
	}{
		{
			desc: "Should success list model endpoint alerts",
			vars: map[string]string{
				"model_id":          "1",
				"model_endpoint_id": "1",
			},
			service: func() *mocks.ModelEndpointAlertService {
				svc := &mocks.ModelEndpointAlertService{}
				svc.On("GetModelEndpointAlert", models.ID(1), models.ID(1)).Return(&models.ModelEndpointAlert{
					ID:              models.ID(1),
					ModelID:         models.ID(1),
					ModelEndpointID: models.ID(1),
					EnvironmentName: "dev",
					AlertConditions: models.AlertConditions{
						{
							Enabled:    true,
							MetricType: models.AlertConditionTypeCPU,
							Severity:   models.AlertConditionSeverityCritical,
						},
					},
				}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusOK,
				data: &models.ModelEndpointAlert{
					ID:              models.ID(1),
					ModelID:         models.ID(1),
					ModelEndpointID: models.ID(1),
					EnvironmentName: "dev",
					AlertConditions: models.AlertConditions{
						{
							Enabled:    true,
							MetricType: models.AlertConditionTypeCPU,
							Severity:   models.AlertConditionSeverityCritical,
						},
					},
				},
			},
		},
		{
			desc: "Should return 500 if error fetching model alerts",
			vars: map[string]string{
				"model_id":          "1",
				"model_endpoint_id": "1",
			},
			service: func() *mocks.ModelEndpointAlertService {
				svc := &mocks.ModelEndpointAlertService{}
				svc.On("GetModelEndpointAlert", models.ID(1), models.ID(1)).Return(nil, fmt.Errorf("API is down"))
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error getting alert for model: API is down"},
			},
		},
		{
			desc: "Should return 404 if there is no model alerts",
			vars: map[string]string{
				"model_id":          "1",
				"model_endpoint_id": "1",
			},
			service: func() *mocks.ModelEndpointAlertService {
				svc := &mocks.ModelEndpointAlertService{}
				svc.On("GetModelEndpointAlert", models.ID(1), models.ID(1)).Return(nil, gorm.ErrRecordNotFound)
				return svc
			},
			expected: &Response{
				code: http.StatusNotFound,
				data: Error{Message: "Model alert not found: record not found"},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			svc := tC.service()
			ctl := &AlertsController{
				AppContext: &AppContext{
					ModelEndpointAlertService: svc,
				},
			}
			resp := ctl.GetModelEndpointAlert(&http.Request{}, tC.vars, nil)
			assertEqualResponses(t, tC.expected, resp)
		})
	}
}

func TestCreateModelEndpointAlert(t *testing.T) {
	testCases := []struct {
		desc                      string
		vars                      map[string]string
		request                   interface{}
		modelEndpointService      func() *mocks.ModelEndpointsService
		modelService              func() *mocks.ModelsService
		modelEndpointAlertService func() *mocks.ModelEndpointAlertService
		expected                  *Response
	}{
		{
			desc: "Should success create model endpoint alert",
			vars: map[string]string{
				"user":              "admin",
				"model_id":          "1",
				"model_endpoint_id": "1",
			},
			request: &models.ModelEndpointAlert{
				ModelID:         models.ID(1),
				ModelEndpointID: models.ID(1),
				EnvironmentName: "dev",
				AlertConditions: models.AlertConditions{
					{
						Enabled:    true,
						MetricType: models.AlertConditionTypeCPU,
						Severity:   models.AlertConditionSeverityCritical,
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
					ExperimentID: 0,
				}, nil)
				return svc
			},
			modelEndpointService: func() *mocks.ModelEndpointsService {
				svc := &mocks.ModelEndpointsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.ModelEndpoint{
					ID:              models.ID(1),
					ModelID:         models.ID(1),
					Status:          models.EndpointRunning,
					URL:             "http://serving.com",
					EnvironmentName: "dev",
				}, nil)
				return svc
			},
			modelEndpointAlertService: func() *mocks.ModelEndpointAlertService {
				svc := &mocks.ModelEndpointAlertService{}
				svc.On("CreateModelEndpointAlert", "admin", mock.Anything).Return(&models.ModelEndpointAlert{
					ID:              models.ID(1),
					ModelID:         models.ID(1),
					ModelEndpointID: models.ID(1),
					EnvironmentName: "dev",
					TeamName:        "dsp",
					AlertConditions: models.AlertConditions{
						{
							Enabled:    true,
							MetricType: models.AlertConditionTypeCPU,
							Severity:   models.AlertConditionSeverityCritical,
						},
					},
				}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusCreated,
				data: &models.ModelEndpointAlert{
					ID:              models.ID(1),
					ModelID:         models.ID(1),
					ModelEndpointID: models.ID(1),
					EnvironmentName: "dev",
					TeamName:        "dsp",
					AlertConditions: models.AlertConditions{
						{
							Enabled:    true,
							MetricType: models.AlertConditionTypeCPU,
							Severity:   models.AlertConditionSeverityCritical,
						},
					},
				},
			},
		},
		{
			desc: "Should return 400 if request is invalid",
			vars: map[string]string{
				"user":              "admin",
				"model_id":          "1",
				"model_endpoint_id": "1",
			},
			request: &models.ModelEndpoint{
				ModelID:         models.ID(1),
				EnvironmentName: "dev",
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				return svc
			},
			modelEndpointService: func() *mocks.ModelEndpointsService {
				svc := &mocks.ModelEndpointsService{}
				return svc
			},
			modelEndpointAlertService: func() *mocks.ModelEndpointAlertService {
				svc := &mocks.ModelEndpointAlertService{}
				return svc
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{Message: "Unable to parse body as model endpoint alert"},
			},
		},
		{
			desc: "Should return 404 if model is not exist",
			vars: map[string]string{
				"user":              "admin",
				"model_id":          "1",
				"model_endpoint_id": "1",
			},
			request: &models.ModelEndpointAlert{
				ModelID:         models.ID(1),
				ModelEndpointID: models.ID(1),
				EnvironmentName: "dev",
				AlertConditions: models.AlertConditions{
					{
						Enabled:    true,
						MetricType: models.AlertConditionTypeCPU,
						Severity:   models.AlertConditionSeverityCritical,
					},
				},
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(nil, gorm.ErrRecordNotFound)
				return svc
			},
			modelEndpointService: func() *mocks.ModelEndpointsService {
				svc := &mocks.ModelEndpointsService{}
				return svc
			},
			modelEndpointAlertService: func() *mocks.ModelEndpointAlertService {
				svc := &mocks.ModelEndpointAlertService{}
				return svc
			},
			expected: &Response{
				code: http.StatusNotFound,
				data: Error{Message: "Model not found: record not found"},
			},
		},
		{
			desc: "Should return 500 if fetching model returning error",
			vars: map[string]string{
				"user":              "admin",
				"model_id":          "1",
				"model_endpoint_id": "1",
			},
			request: &models.ModelEndpointAlert{
				ModelID:         models.ID(1),
				ModelEndpointID: models.ID(1),
				EnvironmentName: "dev",
				AlertConditions: models.AlertConditions{
					{
						Enabled:    true,
						MetricType: models.AlertConditionTypeCPU,
						Severity:   models.AlertConditionSeverityCritical,
					},
				},
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(nil, fmt.Errorf("Error creating secret: db is down"))
				return svc
			},
			modelEndpointService: func() *mocks.ModelEndpointsService {
				svc := &mocks.ModelEndpointsService{}
				return svc
			},
			modelEndpointAlertService: func() *mocks.ModelEndpointAlertService {
				svc := &mocks.ModelEndpointAlertService{}
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error finding model: Error creating secret: db is down"},
			},
		},
		{
			desc: "Should return 404 when model endpoint is not exist",
			vars: map[string]string{
				"user":              "admin",
				"model_id":          "1",
				"model_endpoint_id": "1",
			},
			request: &models.ModelEndpointAlert{
				ModelID:         models.ID(1),
				ModelEndpointID: models.ID(1),
				EnvironmentName: "dev",
				AlertConditions: models.AlertConditions{
					{
						Enabled:    true,
						MetricType: models.AlertConditionTypeCPU,
						Severity:   models.AlertConditionSeverityCritical,
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
					ExperimentID: 0,
				}, nil)
				return svc
			},
			modelEndpointService: func() *mocks.ModelEndpointsService {
				svc := &mocks.ModelEndpointsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(nil, gorm.ErrRecordNotFound)
				return svc
			},
			modelEndpointAlertService: func() *mocks.ModelEndpointAlertService {
				svc := &mocks.ModelEndpointAlertService{}
				return svc
			},
			expected: &Response{
				code: http.StatusNotFound,
				data: Error{Message: "Model endpoint not found: record not found"},
			},
		},
		{
			desc: "Should return 500 when model endpoint fetching return error",
			vars: map[string]string{
				"user":              "admin",
				"model_id":          "1",
				"model_endpoint_id": "1",
			},
			request: &models.ModelEndpointAlert{
				ModelID:         models.ID(1),
				ModelEndpointID: models.ID(1),
				EnvironmentName: "dev",
				AlertConditions: models.AlertConditions{
					{
						Enabled:    true,
						MetricType: models.AlertConditionTypeCPU,
						Severity:   models.AlertConditionSeverityCritical,
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
					ExperimentID: 0,
				}, nil)
				return svc
			},
			modelEndpointService: func() *mocks.ModelEndpointsService {
				svc := &mocks.ModelEndpointsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(nil, fmt.Errorf("Error creating secret: db is down"))
				return svc
			},
			modelEndpointAlertService: func() *mocks.ModelEndpointAlertService {
				svc := &mocks.ModelEndpointAlertService{}
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error getting model endpoint: Error creating secret: db is down"},
			},
		},
		{
			desc: "Should return 500 if error when creating model alert",
			vars: map[string]string{
				"user":              "admin",
				"model_id":          "1",
				"model_endpoint_id": "1",
			},
			request: &models.ModelEndpointAlert{
				ModelID:         models.ID(1),
				ModelEndpointID: models.ID(1),
				EnvironmentName: "dev",
				AlertConditions: models.AlertConditions{
					{
						Enabled:    true,
						MetricType: models.AlertConditionTypeCPU,
						Severity:   models.AlertConditionSeverityCritical,
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
					ExperimentID: 0,
				}, nil)
				return svc
			},
			modelEndpointService: func() *mocks.ModelEndpointsService {
				svc := &mocks.ModelEndpointsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.ModelEndpoint{
					ID:              models.ID(1),
					ModelID:         models.ID(1),
					Status:          models.EndpointRunning,
					URL:             "http://serving.com",
					EnvironmentName: "dev",
				}, nil)
				return svc
			},
			modelEndpointAlertService: func() *mocks.ModelEndpointAlertService {
				svc := &mocks.ModelEndpointAlertService{}
				svc.On("CreateModelEndpointAlert", "admin", mock.Anything).Return(nil, fmt.Errorf("Connection refused"))
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error creating alert: Connection refused"},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			modelSvc := tC.modelService()
			modelEndpointSvc := tC.modelEndpointService()
			modelEndpointAlertSvc := tC.modelEndpointAlertService()
			ctl := &AlertsController{
				AppContext: &AppContext{
					ModelsService:             modelSvc,
					ModelEndpointsService:     modelEndpointSvc,
					ModelEndpointAlertService: modelEndpointAlertSvc,
				},
			}
			resp := ctl.CreateModelEndpointAlert(&http.Request{}, tC.vars, tC.request)
			assertEqualResponses(t, tC.expected, resp)
		})
	}
}

func TestUpdateModelEndpointAlert(t *testing.T) {
	testCases := []struct {
		desc                      string
		vars                      map[string]string
		request                   interface{}
		modelService              func() *mocks.ModelsService
		modelEndpointAlertService func() *mocks.ModelEndpointAlertService
		expected                  *Response
	}{
		{
			desc: "Should success update model endpoint alert",
			vars: map[string]string{
				"user":              "admin",
				"model_id":          "1",
				"model_endpoint_id": "1",
			},
			request: &models.ModelEndpointAlert{
				ModelID:         models.ID(1),
				ModelEndpointID: models.ID(1),
				EnvironmentName: "dev",
				AlertConditions: models.AlertConditions{
					{
						Enabled:    true,
						MetricType: models.AlertConditionTypeCPU,
						Severity:   models.AlertConditionSeverityCritical,
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
					ExperimentID: 0,
				}, nil)
				return svc
			},
			modelEndpointAlertService: func() *mocks.ModelEndpointAlertService {
				svc := &mocks.ModelEndpointAlertService{}
				svc.On("GetModelEndpointAlert", models.ID(1), models.ID(1)).Return(&models.ModelEndpointAlert{
					ID:              models.ID(1),
					ModelID:         models.ID(1),
					ModelEndpointID: models.ID(1),
					EnvironmentName: "dev",
					TeamName:        "dsp",
					AlertConditions: models.AlertConditions{
						{
							Enabled:    true,
							MetricType: models.AlertConditionTypeCPU,
							Severity:   models.AlertConditionSeverityCritical,
						},
					},
				}, nil)
				svc.On("UpdateModelEndpointAlert", "admin", mock.Anything).Return(&models.ModelEndpointAlert{
					ID:              models.ID(1),
					ModelID:         models.ID(1),
					ModelEndpointID: models.ID(1),
					EnvironmentName: "dev",
					TeamName:        "dsp",
					AlertConditions: models.AlertConditions{
						{
							Enabled:    true,
							MetricType: models.AlertConditionTypeCPU,
							Severity:   models.AlertConditionSeverityCritical,
						},
					},
				}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusCreated,
				data: &models.ModelEndpointAlert{
					ID:              models.ID(1),
					ModelID:         models.ID(1),
					ModelEndpointID: models.ID(1),
					EnvironmentName: "dev",
					TeamName:        "dsp",
					AlertConditions: models.AlertConditions{
						{
							Enabled:    true,
							MetricType: models.AlertConditionTypeCPU,
							Severity:   models.AlertConditionSeverityCritical,
						},
					},
				},
			},
		},
		{
			desc: "Should return 400 if request is invalid",
			vars: map[string]string{
				"user":              "admin",
				"model_id":          "1",
				"model_endpoint_id": "1",
			},
			request: &models.ModelEndpoint{
				ModelID:         models.ID(1),
				EnvironmentName: "dev",
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				return svc
			},
			modelEndpointAlertService: func() *mocks.ModelEndpointAlertService {
				svc := &mocks.ModelEndpointAlertService{}
				return svc
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{Message: "Unable to parse body as model endpoint alert"},
			},
		},
		{
			desc: "Should return 404 if model is not exist",
			vars: map[string]string{
				"user":              "admin",
				"model_id":          "1",
				"model_endpoint_id": "1",
			},
			request: &models.ModelEndpointAlert{
				ModelID:         models.ID(1),
				ModelEndpointID: models.ID(1),
				EnvironmentName: "dev",
				AlertConditions: models.AlertConditions{
					{
						Enabled:    true,
						MetricType: models.AlertConditionTypeCPU,
						Severity:   models.AlertConditionSeverityCritical,
					},
				},
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(nil, gorm.ErrRecordNotFound)
				return svc
			},
			modelEndpointAlertService: func() *mocks.ModelEndpointAlertService {
				svc := &mocks.ModelEndpointAlertService{}
				return svc
			},
			expected: &Response{
				code: http.StatusNotFound,
				data: Error{Message: "Model not found: record not found"},
			},
		},
		{
			desc: "Should return 500 if fetching model returning error",
			vars: map[string]string{
				"user":              "admin",
				"model_id":          "1",
				"model_endpoint_id": "1",
			},
			request: &models.ModelEndpointAlert{
				ModelID:         models.ID(1),
				ModelEndpointID: models.ID(1),
				EnvironmentName: "dev",
				AlertConditions: models.AlertConditions{
					{
						Enabled:    true,
						MetricType: models.AlertConditionTypeCPU,
						Severity:   models.AlertConditionSeverityCritical,
					},
				},
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(nil, fmt.Errorf("Error creating secret: db is down"))
				return svc
			},
			modelEndpointAlertService: func() *mocks.ModelEndpointAlertService {
				svc := &mocks.ModelEndpointAlertService{}
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error finding model: Error creating secret: db is down"},
			},
		},
		{
			desc: "Should return 404 when there is no model endpoint alert",
			vars: map[string]string{
				"user":              "admin",
				"model_id":          "1",
				"model_endpoint_id": "1",
			},
			request: &models.ModelEndpointAlert{
				ModelID:         models.ID(1),
				ModelEndpointID: models.ID(1),
				EnvironmentName: "dev",
				AlertConditions: models.AlertConditions{
					{
						Enabled:    true,
						MetricType: models.AlertConditionTypeCPU,
						Severity:   models.AlertConditionSeverityCritical,
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
					ExperimentID: 0,
				}, nil)
				return svc
			},
			modelEndpointAlertService: func() *mocks.ModelEndpointAlertService {
				svc := &mocks.ModelEndpointAlertService{}
				svc.On("GetModelEndpointAlert", models.ID(1), models.ID(1)).Return(nil, gorm.ErrRecordNotFound)
				return svc
			},
			expected: &Response{
				code: http.StatusNotFound,
				data: Error{Message: "Model endpoint alert not found: record not found"},
			},
		},
		{
			desc: "Should return 500 when model endpoint alert fetching return error",
			vars: map[string]string{
				"user":              "admin",
				"model_id":          "1",
				"model_endpoint_id": "1",
			},
			request: &models.ModelEndpointAlert{
				ModelID:         models.ID(1),
				ModelEndpointID: models.ID(1),
				EnvironmentName: "dev",
				AlertConditions: models.AlertConditions{
					{
						Enabled:    true,
						MetricType: models.AlertConditionTypeCPU,
						Severity:   models.AlertConditionSeverityCritical,
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
					ExperimentID: 0,
				}, nil)
				return svc
			},
			modelEndpointAlertService: func() *mocks.ModelEndpointAlertService {
				svc := &mocks.ModelEndpointAlertService{}
				svc.On("GetModelEndpointAlert", models.ID(1), models.ID(1)).Return(nil, fmt.Errorf("Error creating secret: db is down"))
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error finding alert for model endpoint: Error creating secret: db is down"},
			},
		},
		{
			desc: "Should return 500 when failed update model endpoint alert",
			vars: map[string]string{
				"user":              "admin",
				"model_id":          "1",
				"model_endpoint_id": "1",
			},
			request: &models.ModelEndpointAlert{
				ModelID:         models.ID(1),
				ModelEndpointID: models.ID(1),
				EnvironmentName: "dev",
				AlertConditions: models.AlertConditions{
					{
						Enabled:    true,
						MetricType: models.AlertConditionTypeCPU,
						Severity:   models.AlertConditionSeverityCritical,
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
					ExperimentID: 0,
				}, nil)
				return svc
			},

			modelEndpointAlertService: func() *mocks.ModelEndpointAlertService {
				svc := &mocks.ModelEndpointAlertService{}
				svc.On("GetModelEndpointAlert", models.ID(1), models.ID(1)).Return(&models.ModelEndpointAlert{
					ID:              models.ID(1),
					ModelID:         models.ID(1),
					ModelEndpointID: models.ID(1),
					EnvironmentName: "dev",
					TeamName:        "dsp",
					AlertConditions: models.AlertConditions{
						{
							Enabled:    true,
							MetricType: models.AlertConditionTypeCPU,
							Severity:   models.AlertConditionSeverityCritical,
						},
					},
				}, nil)
				svc.On("UpdateModelEndpointAlert", "admin", mock.Anything).Return(nil, fmt.Errorf("Something went wrong"))
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error updating model endpoint alert: Something went wrong"},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			modelSvc := tC.modelService()
			modelEndpointAlertSvc := tC.modelEndpointAlertService()
			ctl := &AlertsController{
				AppContext: &AppContext{
					ModelsService:             modelSvc,
					ModelEndpointAlertService: modelEndpointAlertSvc,
				},
			}
			resp := ctl.UpdateModelEndpointAlert(&http.Request{}, tC.vars, tC.request)
			assertEqualResponses(t, tC.expected, resp)
		})
	}
}
