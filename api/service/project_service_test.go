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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/gojek/mlp/api/client"

	"github.com/gojek/merlin/mlp"
	mlpMock "github.com/gojek/merlin/mlp/mocks"
)

func Test_projectService(t *testing.T) {
	ctx := context.Background()

	mockMlpAPIClient := &mlpMock.APIClient{}
	ps := NewProjectsService(mockMlpAPIClient)
	assert.NotNil(t, ps)

	project1 := mlp.Project{
		ID:   1,
		Name: "project-1",
	}
	projects := mlp.Projects{
		client.Project(project1),
	}

	mockMlpAPIClient.On("ListProjects", mock.Anything, "").
		Return(projects, nil)

	got1, err := ps.List(ctx, "")
	assert.Nil(t, err)
	assert.Equal(t, projects, got1)

	mockMlpAPIClient.On("GetProjectByID", mock.Anything, int32(1)).
		Return(project1, nil)

	got2, err := ps.GetByID(ctx, int32(1))
	assert.Nil(t, err)
	assert.Equal(t, project1, got2)

	mockMlpAPIClient.On("GetProjectByName", mock.Anything, "project-1").
		Return(project1, nil)

	got3, err := ps.GetByName(ctx, "project-1")
	assert.Nil(t, err)
	assert.Equal(t, project1, got3)
}
