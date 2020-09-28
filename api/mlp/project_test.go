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

package mlp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProject(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		switch r.URL.Path {
		case "/projects":
			if r.Method == "GET" {
				w.Write([]byte(`[{
					"id": 1,
					"name": "project-1"
				}]`))
			} else if r.Method == "POST" {
				w.Write([]byte(`{
					"id": 1,
					"name": "project-1"
				}`))
			}
		case "/projects/1":
			if r.Method == "GET" {
				w.Write([]byte(`{
					"id": 1,
					"name": "project-1"
				}`))
			} else if r.Method == "PUT" {
				w.Write([]byte(`{
					"id": 1,
					"name": "project-1",
					"readers": ["user@domain.com"]
				}`))
			}
		}
	}))
	defer ts.Close()

	ctx := context.Background()

	c := NewAPIClient(&http.Client{}, ts.URL, "")

	project1, err := c.CreateProject(ctx, Project{Name: "project-1"})
	assert.Nil(t, err)
	assert.NotEmpty(t, project1.Id)
	assert.Equal(t, "project-1", project1.Name)

	project1, err = c.UpdateProject(ctx, Project{
		Id:      1,
		Name:    "project-1",
		Readers: []string{"user@domain.com"},
	})
	assert.Nil(t, err)
	assert.NotEmpty(t, project1.Readers)

	project1, err = c.GetProjectByID(ctx, int32(1))
	assert.Nil(t, err)
	assert.Equal(t, int32(1), project1.Id)

	project1, err = c.GetProjectByName(ctx, "project-1")
	assert.Nil(t, err)
	assert.Equal(t, "project-1", project1.Name)

	projects, err := c.ListProjects(ctx, "project-1")
	assert.Nil(t, err)
	assert.NotNil(t, projects)
}

func TestProject_MlflowExperimentURL(t *testing.T) {
	type args struct {
		experimentID string
	}
	tests := []struct {
		name string
		p    Project
		args args
		want string
	}{
		{
			"1",
			Project{
				MlflowTrackingUrl: "http://mlflow",
			},
			args{"1"},
			"http://mlflow/#/experiments/1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.MlflowExperimentURL(tt.args.experimentID); got != tt.want {
				t.Errorf("Project.MlflowExperimentURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProject_MlflowRunURL(t *testing.T) {
	type args struct {
		experimentID string
		runID        string
	}
	tests := []struct {
		name string
		p    Project
		args args
		want string
	}{
		{
			"1",
			Project{
				MlflowTrackingUrl: "http://mlflow",
			},
			args{"1", "1"},
			"http://mlflow/#/experiments/1/runs/1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.MlflowRunURL(tt.args.experimentID, tt.args.runID); got != tt.want {
				t.Errorf("Project.MlflowRunURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
