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

package service

import (
	"context"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/caraml-dev/mlp/api/client"

	"github.com/caraml-dev/merlin/mlp"
	mlpMock "github.com/caraml-dev/merlin/mlp/mocks"
)

func TestListProjects(t *testing.T) {
	project1 := mlp.Project{
		ID:   1,
		Name: "project-1",
	}
	project2 := mlp.Project{
		ID:   2,
		Name: "project-2",
	}
	projects := mlp.Projects{
		client.Project(project1),
		client.Project(project2),
	}

	// Create new service
	ps, err := NewProjectsService(createMockMLPClient(projects))
	require.NoError(t, err)
	assert.NotNil(t, ps)

	tests := map[string]struct {
		projectName string
		expected    mlp.Projects
	}{
		"list all projects": {"", mlp.Projects(projects)},
		"filter by name":    {"project-1", mlp.Projects{client.Project(project1)}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ps.List(context.Background(), tt.projectName)
			assert.Nil(t, err)
			sort.SliceStable(got, func(i, j int) bool {
				// Sort result by ID
				return got[i].ID < got[j].ID
			})
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestGetProjectByID(t *testing.T) {
	project1 := mlp.Project{
		ID:   1,
		Name: "project-1",
	}
	project2 := mlp.Project{
		ID:   2,
		Name: "project-2",
	}
	projects := mlp.Projects{
		client.Project(project1),
		client.Project(project2),
	}

	// Create new service
	ps, err := NewProjectsService(createMockMLPClient(projects))
	require.NoError(t, err)
	assert.NotNil(t, ps)

	tests := map[string]struct {
		projectID   int32
		expected    mlp.Project
		expectedErr string
	}{
		"project exists":         {1, project1, ""},
		"project does not exist": {100, mlp.Project{}, "Project info for id 100 not found in the cache"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ps.GetByID(context.Background(), tt.projectID)
			if tt.expectedErr != "" {
				assert.EqualError(t, err, tt.expectedErr)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func createMockMLPClient(projects mlp.Projects) mlp.APIClient {
	mockMlpAPIClient := &mlpMock.APIClient{}
	mockMlpAPIClient.On("ListProjects", mock.Anything, "").
		Return(projects, nil)
	return mockMlpAPIClient
}
