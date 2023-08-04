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
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/caraml-dev/merlin/mlp"
	"github.com/caraml-dev/merlin/service/mocks"
)

var (
	projects = mlp.Projects{
		{
			ID:             1,
			Name:           "project-1",
			Administrators: []string{"admin@domain.com"},
			Readers:        []string{"reader@domain.com"},
		},
	}

	project1 = mlp.Project{
		ID:             1,
		Name:           "project-1",
		Administrators: []string{"admin@domain.com"},
		Readers:        []string{"reader@domain.com"},
	}
)

func TestProjectsController_ListProjects(t *testing.T) {
	type args struct {
		r    *http.Request
		vars map[string]string
		in2  interface{}
	}
	tests := []struct {
		name        string
		args        args
		authEnabled bool
		mockFunc    func(*mocks.ProjectsService)
		want        *Response
	}{
		{
			name: "success - user is admin",
			args: args{
				r: &http.Request{},
				vars: map[string]string{
					"name": "project-1",
					"user": "admin@domain.com",
				},
			},
			authEnabled: true,
			mockFunc: func(m *mocks.ProjectsService) {
				m.On("List", mock.Anything, "project-1").Return(projects, nil)
			},
			want: &Response{
				code: http.StatusOK,
				data: projects,
			},
		},
		{
			name: "success - user is reader",
			args: args{
				r: &http.Request{},
				vars: map[string]string{
					"name": "project-1",
					"user": "reader@domain.com",
				},
			},
			authEnabled: true,
			mockFunc: func(m *mocks.ProjectsService) {
				m.On("List", mock.Anything, "project-1").Return(projects, nil)
			},
			want: &Response{
				code: http.StatusOK,
				data: projects,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProjectService := &mocks.ProjectsService{}
			tt.mockFunc(mockProjectService)

			c := &ProjectsController{
				AppContext: &AppContext{
					ProjectsService: mockProjectService,
				},
			}

			got := c.ListProjects(tt.args.r, tt.args.vars, tt.args.in2)
			assertEqualResponses(t, tt.want, got)
		})
	}
}

func TestProjectsController_GetProject(t *testing.T) {
	type args struct {
		r    *http.Request
		vars map[string]string
		body interface{}
	}
	tests := []struct {
		name        string
		args        args
		authEnabled bool
		mockFunc    func(*mocks.ProjectsService)
		want        *Response
	}{
		{
			name: "success - user is admin",
			args: args{
				r: &http.Request{},
				vars: map[string]string{
					"project_id": "1",
					"user":       "admin@domain.com",
				},
			},
			authEnabled: true,
			mockFunc: func(m *mocks.ProjectsService) {
				m.On("GetByID", mock.Anything, int32(1)).Return(project1, nil)
			},
			want: &Response{
				code: http.StatusOK,
				data: project1,
			},
		},
		{
			name: "success - user is reader",
			args: args{
				r: &http.Request{},
				vars: map[string]string{
					"project_id": "1",
					"user":       "reader@domain.com",
				},
			},
			authEnabled: true,
			mockFunc: func(m *mocks.ProjectsService) {
				m.On("GetByID", mock.Anything, int32(1)).Return(project1, nil)
			},
			want: &Response{
				code: http.StatusOK,
				data: project1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProjectService := &mocks.ProjectsService{}
			tt.mockFunc(mockProjectService)

			c := &ProjectsController{
				AppContext: &AppContext{
					ProjectsService:      mockProjectService,
					AuthorizationEnabled: tt.authEnabled,
				},
			}

			got := c.GetProject(tt.args.r, tt.args.vars, tt.args.body)
			assertEqualResponses(t, tt.want, got)
		})
	}
}
