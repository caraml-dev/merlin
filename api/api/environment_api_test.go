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

	"github.com/gojek/merlin/models"
	"github.com/gojek/merlin/service/mocks"
	"github.com/stretchr/testify/assert"
)

func TestListEnvironments(t *testing.T) {
	testCases := []struct {
		desc       string
		vars       map[string]string
		envService func() *mocks.EnvironmentService
		expected   *APIResponse
	}{
		{
			desc: "Should success list environment",
			vars: map[string]string{
				"name": "dev",
			},
			envService: func() *mocks.EnvironmentService {
				mockSvc := &mocks.EnvironmentService{}
				mockSvc.On("ListEnvironments", "dev").Return([]*models.Environment{
					{
						Id:         models.Id(1),
						Name:       "dev",
						Cluster:    "dev",
						Region:     "id",
						GcpProject: "dev",
						MaxCpu:     "1",
						MaxMemory:  "1Gi",
					},
				}, nil)
				return mockSvc
			},
			expected: &APIResponse{
				code: http.StatusOK,
				data: []*models.Environment{
					{
						Id:         models.Id(1),
						Name:       "dev",
						Cluster:    "dev",
						Region:     "id",
						GcpProject: "dev",
						MaxCpu:     "1",
						MaxMemory:  "1Gi",
					},
				},
			},
		},
		{
			desc: "Should return 500 when failed fetching list of environment",
			vars: map[string]string{
				"name": "dev",
			},
			envService: func() *mocks.EnvironmentService {
				mockSvc := &mocks.EnvironmentService{}
				mockSvc.On("ListEnvironments", "dev").Return(nil, fmt.Errorf("Database is down"))
				return mockSvc
			},
			expected: &APIResponse{
				code: http.StatusInternalServerError,
				data: Error{
					Message: "Database is down",
				},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			mockSvc := tC.envService()
			ctl := &EnvironmentController{
				AppContext: &AppContext{
					EnvironmentService: mockSvc,
				},
			}
			resp := ctl.ListEnvironments(&http.Request{}, tC.vars, nil)
			assert.Equal(t, tC.expected, resp)
		})
	}
}
