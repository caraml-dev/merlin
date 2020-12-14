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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	mux2 "github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/gojek/mlp/api/pkg/authz/enforcer"
	enforcerMock "github.com/gojek/mlp/api/pkg/authz/enforcer/mocks"

	"github.com/gojek/merlin/config"
	"github.com/gojek/merlin/mlp"
	"github.com/gojek/merlin/models"
	"github.com/gojek/merlin/service/mocks"
)

var (
	user   = "user@email.com"
	reject = false

	basePath = "/v1"
)

func TestRejectAuthorization(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		path     string
		resource string
		action   string
		model    *models.Model
	}{
		{
			"reject: list environments",
			http.MethodGet,
			"/v1/environments",
			"environments",
			enforcer.ActionRead,
			nil,
		},
		{
			"reject: list projects",
			http.MethodGet,
			"/v1/projects",
			"projects",
			enforcer.ActionRead,
			nil,
		},
		{
			"reject: list models",
			http.MethodGet,
			"/v1/projects/1/models",
			"projects:1:models",
			enforcer.ActionRead,
			nil,
		},
		{
			"reject: create models",
			http.MethodPost,
			"/v1/projects/1/models",
			"projects:1:models",
			enforcer.ActionCreate,
			nil,
		},
		{
			"reject: list model versions",
			http.MethodGet,
			"/v1/models/2/versions",
			"projects:1:models:2:versions",
			enforcer.ActionRead,
			&models.Model{
				ID:        2,
				ProjectID: 1,
				Project: mlp.Project{
					Id:   1,
					Name: "my-project",
				},
				Name: "my-model",
			},
		},
		{
			"reject: create model versions",
			http.MethodPost,
			"/v1/models/2/versions",
			"projects:1:models:2:versions",
			enforcer.ActionCreate,
			&models.Model{
				ID:        2,
				ProjectID: 1,
				Project: mlp.Project{
					Id:   1,
					Name: "my-project",
				},
				Name: "my-model",
			},
		},
		{
			"reject: get model versions",
			http.MethodGet,
			"/v1/models/2/versions/3",
			"projects:1:models:2:versions:3",
			enforcer.ActionRead,
			&models.Model{
				ID:        2,
				ProjectID: 1,
				Project: mlp.Project{
					Id:   1,
					Name: "my-project",
				},
				Name: "my-model",
			},
		},
		{
			"reject: patch model versions",
			http.MethodPatch,
			"/v1/models/2/versions/3",
			"projects:1:models:2:versions:3",
			enforcer.ActionUpdate,
			&models.Model{
				ID:        2,
				ProjectID: 1,
				Project: mlp.Project{
					Id:   1,
					Name: "my-project",
				},
				Name: "my-model",
			},
		},
		{
			"reject: list model endpoints in project",
			http.MethodGet,
			"/v1/projects/1/model_endpoints",
			"projects:1:model_endpoints",
			enforcer.ActionRead,
			nil,
		},
		{
			"reject: list model endpoints",
			http.MethodGet,
			"/v1/models/2/endpoints",
			"projects:1:models:2:endpoints",
			enforcer.ActionRead,
			&models.Model{
				ID:        2,
				ProjectID: 1,
				Project: mlp.Project{
					Id:   1,
					Name: "my-project",
				},
				Name: "my-model",
			},
		},
		{
			"reject: create model endpoints",
			http.MethodPost,
			"/v1/models/2/endpoints",
			"projects:1:models:2:endpoints",
			enforcer.ActionCreate,
			&models.Model{
				ID:        2,
				ProjectID: 1,
				Project: mlp.Project{
					Id:   1,
					Name: "my-project",
				},
				Name: "my-model",
			},
		},
		{
			"reject: get model endpoints",
			http.MethodGet,
			"/v1/models/2/endpoints/4",
			"projects:1:models:2:endpoints:4",
			enforcer.ActionRead,
			&models.Model{
				ID:        2,
				ProjectID: 1,
				Project: mlp.Project{
					Id:   1,
					Name: "my-project",
				},
				Name: "my-model",
			},
		},
		{
			"reject: update model endpoints",
			http.MethodPut,
			"/v1/models/2/endpoints/4",
			"projects:1:models:2:endpoints:4",
			enforcer.ActionUpdate,
			&models.Model{
				ID:        2,
				ProjectID: 1,
				Project: mlp.Project{
					Id:   1,
					Name: "my-project",
				},
				Name: "my-model",
			},
		},
		{
			"reject: list version endpoints",
			http.MethodGet,
			"/v1/models/2/versions/3/endpoint",
			"projects:1:models:2:versions:3:endpoint",
			enforcer.ActionRead,
			&models.Model{
				ID:        2,
				ProjectID: 1,
				Project: mlp.Project{
					Id:   1,
					Name: "my-project",
				},
				Name: "my-model",
			},
		},
		{
			"reject: create version endpoints",
			http.MethodPost,
			"/v1/models/2/versions/3/endpoint",
			"projects:1:models:2:versions:3:endpoint",
			enforcer.ActionCreate,
			&models.Model{
				ID:        2,
				ProjectID: 1,
				Project: mlp.Project{
					Id:   1,
					Name: "my-project",
				},
				Name: "my-model",
			},
		},
		{
			"reject: delete version endpoints (old)",
			http.MethodDelete,
			"/v1/models/2/versions/3/endpoint",
			"projects:1:models:2:versions:3:endpoint",
			enforcer.ActionDelete,
			&models.Model{
				ID:        2,
				ProjectID: 1,
				Project: mlp.Project{
					Id:   1,
					Name: "my-project",
				},
				Name: "my-model",
			},
		},
		{
			"reject: get version endpoints",
			http.MethodGet,
			"/v1/models/2/versions/3/endpoint/8e9624e0-efd3-44c9-941d-e645d5f680e8",
			"projects:1:models:2:versions:3:endpoint:8e9624e0-efd3-44c9-941d-e645d5f680e8",
			enforcer.ActionRead,
			&models.Model{
				ID:        2,
				ProjectID: 1,
				Project: mlp.Project{
					Id:   1,
					Name: "my-project",
				},
				Name: "my-model",
			},
		},
		{
			"reject: update version endpoints",
			http.MethodPut,
			"/v1/models/2/versions/3/endpoint/8e9624e0-efd3-44c9-941d-e645d5f680e8",
			"projects:1:models:2:versions:3:endpoint:8e9624e0-efd3-44c9-941d-e645d5f680e8",
			enforcer.ActionUpdate,
			&models.Model{
				ID:        2,
				ProjectID: 1,
				Project: mlp.Project{
					Id:   1,
					Name: "my-project",
				},
				Name: "my-model",
			},
		},
		{
			"reject: get version endpoints",
			http.MethodDelete,
			"/v1/models/2/versions/3/endpoint/8e9624e0-efd3-44c9-941d-e645d5f680e8",
			"projects:1:models:2:versions:3:endpoint:8e9624e0-efd3-44c9-941d-e645d5f680e8",
			enforcer.ActionDelete,
			&models.Model{
				ID:        2,
				ProjectID: 1,
				Project: mlp.Project{
					Id:   1,
					Name: "my-project",
				},
				Name: "my-model",
			},
		},
		{
			"reject: list containers",
			http.MethodGet,
			"/v1/models/2/versions/3/endpoint/8e9624e0-efd3-44c9-941d-e645d5f680e8/containers",
			"projects:1:models:2:versions:3:endpoint:8e9624e0-efd3-44c9-941d-e645d5f680e8:containers",
			enforcer.ActionRead,
			&models.Model{
				ID:        2,
				ProjectID: 1,
				Project: mlp.Project{
					Id:   1,
					Name: "my-project",
				},
				Name: "my-model",
			},
		},
		{
			"reject: read log",
			http.MethodGet,
			"/v1/logs?name=pyfunc-image-builder&pod_name=maf-dnf-3&namespace=my-project&cluster=products&gcp_project=&version_endpoint_id=0e1b0dc6-94ee-4417-ad0c-8078f694ac3c&follow=true&timestamps=true&model_id=2",
			"projects:1:logs",
			enforcer.ActionRead,
			&models.Model{
				ID:        2,
				ProjectID: 1,
				Project: mlp.Project{
					Id:   1,
					Name: "my-project",
				},
				Name: "my-model",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEnforcer := &enforcerMock.Enforcer{}
			mockModelService := &mocks.ModelsService{}
			mockEndpointService := &mocks.EndpointsService{}
			r := NewRouter(AppContext{
				EnvironmentService:    &mocks.EnvironmentService{},
				ProjectsService:       &mocks.ProjectsService{},
				ModelsService:         mockModelService,
				ModelEndpointsService: &mocks.ModelEndpointsService{},
				VersionsService:       &mocks.VersionsService{},
				EndpointsService:      mockEndpointService,
				LogService:            &mocks.LogService{},
				Enforcer:              mockEnforcer,
				DB:                    nil,
				AuthorizationEnabled:  true,
				MonitoringConfig: config.MonitoringConfig{
					MonitoringEnabled: true,
					MonitoringBaseURL: "http://grafana",
				},
				AlertEnabled: true,
			})
			if tt.model != nil {
				mockModelService.On("FindByID", mock.Anything, tt.model.ID).Return(tt.model, nil)
			}

			mockEnforcer.On("Enforce", user, tt.resource, tt.action).Return(&reject, nil)

			req, err := http.NewRequest(tt.method, tt.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			req.Header["User-Email"] = []string{user}
			rr := httptest.NewRecorder()

			route := mux2.NewRouter()
			route.PathPrefix(basePath).Handler(
				http.StripPrefix(
					strings.TrimSuffix(basePath, "/"),
					r,
				),
			)
			route.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusUnauthorized, rr.Code)
		})
	}
}
