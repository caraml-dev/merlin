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

func Test_projectService(t *testing.T) {
	ctx := context.Background()
	mockMlpAPIClient := &mlpMock.APIClient{}
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
	mockMlpAPIClient.On("ListProjects", mock.Anything, "").
		Return(projects, nil)

	// Create new service
	ps, err := NewProjectsService(mockMlpAPIClient)
	require.NoError(t, err)
	assert.NotNil(t, ps)

	// Test List
	got1, err := ps.List(ctx, "")
	assert.Nil(t, err)
	sort.SliceStable(got1, func(i, j int) bool {
		// Sort result by ID
		return got1[i].ID < got1[j].ID
	})
	assert.Equal(t, projects, got1)

	got2, err := ps.List(ctx, "project-1")
	assert.Nil(t, err)
	assert.Equal(t, mlp.Projects{client.Project(project1)}, got2)

	// Test GetByID
	got3, err := ps.GetByID(ctx, int32(1))
	assert.Nil(t, err)
	assert.Equal(t, project1, got3)
}
