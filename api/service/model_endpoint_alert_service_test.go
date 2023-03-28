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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/caraml-dev/merlin/mlp"
	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/pkg/gitlab"
	gitlabmocks "github.com/caraml-dev/merlin/pkg/gitlab/mocks"
	storagemocks "github.com/caraml-dev/merlin/storage/mocks"
	wardenmocks "github.com/caraml-dev/merlin/warden/mocks"
)

func Test_modelEndpointAlertService_ListTeams(t *testing.T) {
	wardenClient := &wardenmocks.Client{}
	wardenClient.On("GetAllTeams").Return([]string{"datascience"}, nil)

	svc := modelEndpointAlertService{
		wardenClient: wardenClient,
	}

	teams, err := svc.ListTeams()
	assert.Nil(t, err)
	assert.NotEmpty(t, teams)
	assert.Equal(t, "datascience", teams[0])
}

func Test_modelEndpointAlertService_CreateModelEndpointAlert(t *testing.T) {
	mockGitlabClient := &gitlabmocks.Client{}
	mockGitlabClient.On("CreateFile", mock.Anything).
		Run(func(args mock.Arguments) {
			assert.Len(t, args, 1)

			opt := args[0].(gitlab.CreateFileOptions)
			yamlBytes, err := os.ReadFile("./testdata/model_endpoint_alert.yaml")
			assert.Nil(t, err)
			assert.Equal(t, string(yamlBytes), opt.Content)
		}).
		Return(nil)

	mockAlertStorage := &storagemocks.AlertStorage{}
	mockAlertStorage.On("CreateModelEndpointAlert", mock.Anything).Return(nil)

	alert := &models.ModelEndpointAlert{
		ModelID: 1,
		Model: &models.Model{
			Name: "model-1",
			Project: mlp.Project{
				Name: "project-1",
			},
		},
		ModelEndpointID: models.ID(1),
		ModelEndpoint: &models.ModelEndpoint{
			ID: models.ID(1),
			Environment: &models.Environment{
				Cluster: "cluster-1",
			},
		},
		EnvironmentName: "env-1",
		TeamName:        "team-1",
		AlertConditions: models.AlertConditions{
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeThroughput,
				Severity:   models.AlertConditionSeverityWarning,
				Target:     1,
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeThroughput,
				Severity:   models.AlertConditionSeverityCritical,
				Target:     2,
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeLatency,
				Severity:   models.AlertConditionSeverityWarning,
				Target:     3,
				Percentile: 99,
				Unit:       "ms",
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeLatency,
				Severity:   models.AlertConditionSeverityCritical,
				Target:     4,
				Percentile: 99,
				Unit:       "ms",
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeLatency,
				Severity:   models.AlertConditionSeverityWarning,
				Target:     5,
				Percentile: 95,
				Unit:       "ms",
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeLatency,
				Severity:   models.AlertConditionSeverityCritical,
				Target:     6,
				Percentile: 95,
				Unit:       "ms",
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeErrorRate,
				Severity:   models.AlertConditionSeverityWarning,
				Target:     7,
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeErrorRate,
				Severity:   models.AlertConditionSeverityCritical,
				Target:     8,
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeCPU,
				Severity:   models.AlertConditionSeverityWarning,
				Target:     9,
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeCPU,
				Severity:   models.AlertConditionSeverityCritical,
				Target:     10,
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeMemory,
				Severity:   models.AlertConditionSeverityWarning,
				Target:     11,
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeMemory,
				Severity:   models.AlertConditionSeverityCritical,
				Target:     12,
			},
		},
	}

	s := &modelEndpointAlertService{
		alertStorage:     mockAlertStorage,
		gitlabClient:     mockGitlabClient,
		alertRepository:  "merlin/alerts",
		alertBranch:      "main",
		dashboardBaseURL: "http://dashboard.dev/",
	}

	savedAlert, err := s.CreateModelEndpointAlert("author-test", alert)
	assert.Nil(t, err)
	assert.Equal(t, alert, savedAlert)
}
