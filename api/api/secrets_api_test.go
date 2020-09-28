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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/gojek/merlin/mlp"
	"github.com/gojek/merlin/service/mocks"
)

func TestCreateSecret(t *testing.T) {
	testCases := []struct {
		desc               string
		vars               map[string]string
		existingProject    mlp.Project
		errFetchingProject error
		body               interface{}
		savedSecret        mlp.Secret
		errSaveSecret      error
		expectedResponse   *ApiResponse
	}{
		{
			desc: "Should success",
			vars: map[string]string{
				"project_id": "1",
				"user":       "admin@domain.com",
			},
			body: &mlp.Secret{
				Name: "name",
				Data: `{"id": 3}`,
			},
			expectedResponse: &ApiResponse{
				code: 201,
				data: mlp.Secret{
					Id:   int32(1),
					Name: "name",
					Data: "encryptedData",
				},
			},
			existingProject: mlp.Project{
				Id:   1,
				Name: "project",
				Administrators: []string{
					"admin@domain.com",
				},
			},
			savedSecret: mlp.Secret{
				Id:   int32(1),
				Name: "name",
				Data: "encryptedData",
			},
		},
		{
			desc: "Should return not found if project is not exist",
			vars: map[string]string{
				"project_id": "1",
				"user":       "admin@domain.com",
			},
			body: &mlp.Secret{
				Name: "name",
				Data: `{"id": 3}`,
			},
			expectedResponse: &ApiResponse{
				code: 404,
				data: Error{"Project with given `project_id: 1` not found"},
			},
			errFetchingProject: fmt.Errorf("project not found"),
		},
		{
			desc: "Should return internal server error when failed save secret",
			vars: map[string]string{
				"project_id": "1",
				"user":       "admin@domain.com",
			},
			body: &mlp.Secret{
				Name: "name",
				Data: `{"id": 3}`,
			},
			expectedResponse: &ApiResponse{
				code: 500,
				data: Error{"db is down"},
			},
			existingProject: mlp.Project{
				Id:   1,
				Name: "project",
				Administrators: []string{
					"admin@domain.com",
				},
			},
			errSaveSecret: fmt.Errorf("db is down"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			projectService := &mocks.ProjectsService{}
			projectService.On("GetByID", mock.Anything, int32(1)).Return(tC.existingProject, tC.errFetchingProject)

			secretService := &mocks.SecretService{}
			secretService.On("Create", mock.Anything, int32(1), mock.Anything).Return(tC.savedSecret, tC.errSaveSecret)

			controller := &SecretsController{
				AppContext: &AppContext{
					SecretService:   secretService,
					ProjectsService: projectService,
				},
			}

			res := controller.CreateSecret(&http.Request{}, tC.vars, tC.body)
			assert.Equal(t, tC.expectedResponse, res)
		})
	}
}

func TestUpdateSecret(t *testing.T) {
	testCases := []struct {
		desc               string
		vars               map[string]string
		existingProject    mlp.Project
		errFetchingProject error
		body               interface{}
		updatedSecret      mlp.Secret
		errUpdatingSecret  error
		expectedResponse   *ApiResponse
	}{
		{
			desc: "Should responded 204",
			vars: map[string]string{
				"project_id": "1",
				"user":       "admin@domain.com",
				"secret_id":  "1",
			},
			existingProject: mlp.Project{
				Id:   1,
				Name: "project",
				Administrators: []string{
					"admin@domain.com",
				},
			},
			body: &mlp.Secret{
				Name: "name",
				Data: `{"id": 3}`,
			},
			updatedSecret: mlp.Secret{
				Id:   int32(1),
				Name: "name",
				Data: `{"id": 3}`,
			},
			expectedResponse: &ApiResponse{
				code: 200,
				data: mlp.Secret{
					Id:   int32(1),
					Name: "name",
					Data: `{"id": 3}`,
				},
			},
		},
		{
			desc: "Should responded 204 even body is partially there",
			vars: map[string]string{
				"project_id": "1",
				"user":       "admin@domain.com",
				"secret_id":  "1",
			},
			existingProject: mlp.Project{
				Id:   1,
				Name: "project",
				Administrators: []string{
					"admin@domain.com",
				},
			},
			body: &mlp.Secret{
				Name: "name",
			},
			updatedSecret: mlp.Secret{
				Id:   int32(1),
				Name: "name",
				Data: `{"id": 3}`,
			},
			expectedResponse: &ApiResponse{
				code: 200,
				data: mlp.Secret{
					Id:   int32(1),
					Name: "name",
					Data: `{"id": 3}`,
				},
			},
		},
		{
			desc: "Should responded 400 when project_id and secret_id is not integer",
			vars: map[string]string{
				"project_id": "abc",
				"secret_id":  "def",
			},
			existingProject: mlp.Project{
				Id:   1,
				Name: "project",
				Administrators: []string{
					"admin@domain.com",
				},
			},
			body: &mlp.Secret{
				Name: "name",
			},
			updatedSecret: mlp.Secret{
				Id:   int32(1),
				Name: "name",
				Data: `{"id": 3}`,
			},
			expectedResponse: &ApiResponse{
				code: 400,
				data: Error{"project_id and secret_id is not valid"},
			},
		},
		{
			desc: "Should responded 400 when body is invalid",
			vars: map[string]string{
				"project_id": "1",
				"user":       "admin@domain.com",
				"secret_id":  "1",
			},
			existingProject: mlp.Project{
				Id:   1,
				Name: "project",
				Administrators: []string{
					"admin@domain.com",
				},
			},
			body: "body",
			updatedSecret: mlp.Secret{
				Id:   int32(1),
				Name: "name",
				Data: `{"id": 3}`,
			},
			expectedResponse: &ApiResponse{
				code: 400,
				data: Error{"Invalid request body"},
			},
		},
		{
			desc: "Should responded 500 when secret not found",
			vars: map[string]string{
				"project_id": "1",
				"user":       "admin@domain.com",
				"secret_id":  "1",
			},
			existingProject: mlp.Project{
				Id:   1,
				Name: "project",
				Administrators: []string{
					"admin@domain.com",
				},
			},
			body: &mlp.Secret{
				Name: "name",
			},
			errUpdatingSecret: fmt.Errorf("Secret with given `secret_id: 1` and `project_id: 1` not found"),
			expectedResponse: &ApiResponse{
				code: 500,
				data: Error{"Secret with given `secret_id: 1` and `project_id: 1` not found"},
			},
		},
		{
			desc: "Should responded 500 when error updating secret",
			vars: map[string]string{
				"project_id": "1",
				"user":       "admin@domain.com",
				"secret_id":  "1",
			},
			existingProject: mlp.Project{
				Id:   1,
				Name: "project",
				Administrators: []string{
					"admin@domain.com",
				},
			},
			body: &mlp.Secret{
				Name: "name",
			},
			errUpdatingSecret: fmt.Errorf("db is down"),
			expectedResponse: &ApiResponse{
				code: 500,
				data: Error{"db is down"},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			projectService := &mocks.ProjectsService{}
			projectService.On("GetByID", mock.Anything, int32(1)).Return(tC.existingProject, tC.errFetchingProject)

			secretService := &mocks.SecretService{}
			secretService.On("Update", mock.Anything, int32(1), mock.Anything).Return(tC.updatedSecret, tC.errUpdatingSecret)

			controller := &SecretsController{
				AppContext: &AppContext{
					ProjectsService: projectService,
					SecretService:   secretService,
				},
			}

			res := controller.UpdateSecret(&http.Request{}, tC.vars, tC.body)
			assert.Equal(t, tC.expectedResponse, res)
		})
	}
}

func TestDeleteSecret(t *testing.T) {
	testCases := []struct {
		desc               string
		vars               map[string]string
		existingProject    mlp.Project
		errFetchingProject error
		errDeletingSecret  error
		expectedResponse   *ApiResponse
	}{
		{
			desc: "Should responsed 204",
			vars: map[string]string{
				"project_id": "1",
				"user":       "admin@domain.com",
				"secret_id":  "1",
			},
			existingProject: mlp.Project{
				Id:   1,
				Name: "project",
				Administrators: []string{
					"admin@domain.com",
				},
			},
			expectedResponse: &ApiResponse{
				code: 204,
				data: nil,
			},
		},
		{
			desc: "Should responsed 400 if project_id or secret_id is invalid",
			vars: map[string]string{
				"project_id": "def",
				"secret_id":  "ghi",
			},
			existingProject: mlp.Project{
				Id:   1,
				Name: "project",
				Administrators: []string{
					"admin@domain.com",
				},
			},
			expectedResponse: &ApiResponse{
				code: 400,
				data: Error{"project_id and secret_id is not valid"},
			},
		},
		{
			desc: "Should responsed 500",
			vars: map[string]string{
				"project_id": "1",
				"user":       "admin@domain.com",
				"secret_id":  "1",
			},
			existingProject: mlp.Project{
				Id:   1,
				Name: "project",
				Administrators: []string{
					"admin@domain.com",
				},
			},
			errDeletingSecret: fmt.Errorf("db is down"),
			expectedResponse: &ApiResponse{
				code: 500,
				data: Error{"db is down"},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			projectService := &mocks.ProjectsService{}
			projectService.On("GetByID", mock.Anything, int32(1)).Return(tC.existingProject, tC.errFetchingProject)

			secretService := &mocks.SecretService{}
			secretService.On("Delete", mock.Anything, int32(1), int32(1)).Return(tC.errDeletingSecret)

			controller := &SecretsController{
				AppContext: &AppContext{
					ProjectsService: projectService,
					SecretService:   secretService,
				},
			}
			res := controller.DeleteSecret(&http.Request{}, tC.vars, nil)
			assert.Equal(t, tC.expectedResponse, res)
		})
	}
}

func TestListSecret(t *testing.T) {
	testCases := []struct {
		desc               string
		vars               map[string]string
		existingProject    mlp.Project
		errFetchingProject error
		secrets            mlp.Secrets
		errListSecrets     error
		expectedResponse   *ApiResponse
	}{
		{
			desc: "Should success",
			vars: map[string]string{
				"project_id": "1",
				"user":       "reader@domain.com",
			},
			existingProject: mlp.Project{
				Id:   1,
				Name: "project",
				Administrators: []string{
					"admin@domain.com",
				},
				Readers: []string{
					"reader@domain.com",
				},
			},
			expectedResponse: &ApiResponse{
				code: 200,
				data: mlp.Secrets{
					{
						Id:   int32(1),
						Name: "name-1",
						Data: "encryptedData",
					},
					{
						Id:   int32(2),
						Name: "name-2",
						Data: "encryptedData",
					},
				},
			},
			secrets: mlp.Secrets{
				{
					Id:   int32(1),
					Name: "name-1",
					Data: "encryptedData",
				},
				{
					Id:   int32(2),
					Name: "name-2",
					Data: "encryptedData",
				},
			},
		},
		{
			desc: "Should return internal server error when listing secrets",
			vars: map[string]string{
				"project_id": "1",
				"user":       "reader@domain.com",
			},
			existingProject: mlp.Project{
				Id:   1,
				Name: "project",
				Administrators: []string{
					"admin@domain.com",
				},
				Readers: []string{
					"reader@domain.com",
				},
			},
			expectedResponse: &ApiResponse{
				code: 500,
				data: Error{"db is down"},
			},
			errListSecrets: fmt.Errorf("db is down"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			projectService := &mocks.ProjectsService{}
			projectService.On("GetByID", mock.Anything, int32(1)).Return(tC.existingProject, tC.errFetchingProject)

			secretService := &mocks.SecretService{}
			secretService.On("List", mock.Anything, int32(1)).Return(tC.secrets, tC.errListSecrets)

			controller := &SecretsController{
				AppContext: &AppContext{
					ProjectsService: projectService,
					SecretService:   secretService,
				},
			}

			res := controller.ListSecret(&http.Request{}, tC.vars, nil)
			assert.Equal(t, tC.expectedResponse, res)
		})
	}
}
