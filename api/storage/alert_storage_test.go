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

// +build integration_local integration

package storage

import (
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"

	"github.com/gojek/merlin/it/database"
	"github.com/gojek/merlin/mlp"
	"github.com/gojek/merlin/models"
)

func populateAlertTable(db *gorm.DB) []*models.ModelEndpointAlert {
	isDefaultTrue := true

	db = db.LogMode(true)

	env1 := models.Environment{
		Name:      "env-1",
		Cluster:   "k8s",
		IsDefault: &isDefaultTrue,
	}
	db.Create(&env1)

	project1 := mlp.Project{
		Name:              "project-1",
		MlflowTrackingUrl: "http://mlflow:5000",
	}

	model1 := &models.Model{
		Id:           1,
		ProjectId:    models.Id(project1.Id),
		ExperimentId: 1,
		Name:         "model-1",
		Type:         models.ModelTypeSkLearn,
	}
	db.Create(model1)

	model2 := &models.Model{
		Id:           2,
		ProjectId:    models.Id(project1.Id),
		ExperimentId: 2,
		Name:         "model-2",
		Type:         models.ModelTypeSkLearn,
	}
	db.Create(model2)

	model1Endpoint1 := &models.ModelEndpoint{
		Id:              1,
		ModelId:         model1.Id,
		Status:          models.EndpointServing,
		EnvironmentName: env1.Name,
	}
	db.Create(model1Endpoint1)

	model1Endpoint2 := &models.ModelEndpoint{
		Id:              2,
		ModelId:         model1.Id,
		Status:          models.EndpointServing,
		EnvironmentName: env1.Name,
	}
	db.Create(model1Endpoint2)

	model2Endpoint3 := &models.ModelEndpoint{
		Id:              3,
		ModelId:         model1.Id,
		Status:          models.EndpointServing,
		EnvironmentName: env1.Name,
	}
	db.Create(model2Endpoint3)

	alerts := []*models.ModelEndpointAlert{}

	alertModel1Endpoint1 := &models.ModelEndpointAlert{
		ModelId:         model1.Id,
		ModelEndpointId: model1Endpoint1.Id,
		EnvironmentName: env1.Name,
		TeamName:        "datascience",
		AlertConditions: models.AlertConditions{
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeThroughput,
				Severity:   models.AlertConditionSeverityWarning,
				Target:     10,
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeThroughput,
				Severity:   models.AlertConditionSeverityCritical,
				Target:     1,
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeLatency,
				Severity:   models.AlertConditionSeverityWarning,
				Target:     10,
				Percentile: 99,
				Unit:       "ms",
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeLatency,
				Severity:   models.AlertConditionSeverityCritical,
				Target:     15,
				Percentile: 99,
				Unit:       "ms",
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeErrorRate,
				Severity:   models.AlertConditionSeverityWarning,
				Target:     1,
				Unit:       "%",
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeErrorRate,
				Severity:   models.AlertConditionSeverityCritical,
				Target:     5,
				Unit:       "%",
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeCPU,
				Severity:   models.AlertConditionSeverityWarning,
				Target:     90,
				Unit:       "%",
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeCPU,
				Severity:   models.AlertConditionSeverityCritical,
				Target:     95,
				Unit:       "%",
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeMemory,
				Severity:   models.AlertConditionSeverityWarning,
				Target:     90,
				Unit:       "%",
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeMemory,
				Severity:   models.AlertConditionSeverityCritical,
				Target:     95,
				Unit:       "%",
			},
		},
	}
	alerts = append(alerts, alertModel1Endpoint1)

	alertModel1Endpoint2 := &models.ModelEndpointAlert{
		ModelId:         model1.Id,
		ModelEndpointId: model1Endpoint2.Id,
		EnvironmentName: env1.Name,
		TeamName:        "datascience",
		AlertConditions: models.AlertConditions{
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeThroughput,
				Severity:   models.AlertConditionSeverityWarning,
				Target:     10,
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeThroughput,
				Severity:   models.AlertConditionSeverityCritical,
				Target:     1,
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeLatency,
				Severity:   models.AlertConditionSeverityWarning,
				Target:     10,
				Percentile: 99,
				Unit:       "ms",
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeLatency,
				Severity:   models.AlertConditionSeverityCritical,
				Target:     15,
				Percentile: 99,
				Unit:       "ms",
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeErrorRate,
				Severity:   models.AlertConditionSeverityWarning,
				Target:     1,
				Unit:       "%",
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeErrorRate,
				Severity:   models.AlertConditionSeverityCritical,
				Target:     5,
				Unit:       "%",
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeCPU,
				Severity:   models.AlertConditionSeverityWarning,
				Target:     90,
				Unit:       "%",
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeCPU,
				Severity:   models.AlertConditionSeverityCritical,
				Target:     95,
				Unit:       "%",
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeMemory,
				Severity:   models.AlertConditionSeverityWarning,
				Target:     90,
				Unit:       "%",
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeMemory,
				Severity:   models.AlertConditionSeverityCritical,
				Target:     95,
				Unit:       "%",
			},
		},
	}
	alerts = append(alerts, alertModel1Endpoint2)

	alertModel2Endpoint3 := &models.ModelEndpointAlert{
		ModelId:         model2.Id,
		ModelEndpointId: model2Endpoint3.Id,
		EnvironmentName: env1.Name,
		TeamName:        "datascience",
		AlertConditions: models.AlertConditions{
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeThroughput,
				Severity:   models.AlertConditionSeverityWarning,
				Target:     10,
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeThroughput,
				Severity:   models.AlertConditionSeverityCritical,
				Target:     1,
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeLatency,
				Severity:   models.AlertConditionSeverityWarning,
				Target:     10,
				Percentile: 99,
				Unit:       "ms",
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeLatency,
				Severity:   models.AlertConditionSeverityCritical,
				Target:     15,
				Percentile: 99,
				Unit:       "ms",
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeErrorRate,
				Severity:   models.AlertConditionSeverityWarning,
				Target:     1,
				Unit:       "%",
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeErrorRate,
				Severity:   models.AlertConditionSeverityCritical,
				Target:     5,
				Unit:       "%",
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeCPU,
				Severity:   models.AlertConditionSeverityWarning,
				Target:     90,
				Unit:       "%",
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeCPU,
				Severity:   models.AlertConditionSeverityCritical,
				Target:     95,
				Unit:       "%",
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeMemory,
				Severity:   models.AlertConditionSeverityWarning,
				Target:     90,
				Unit:       "%",
			},
			&models.AlertCondition{
				Enabled:    true,
				MetricType: models.AlertConditionTypeMemory,
				Severity:   models.AlertConditionSeverityCritical,
				Target:     95,
				Unit:       "%",
			},
		},
	}
	alerts = append(alerts, alertModel2Endpoint3)

	for _, alert := range alerts {
		db.Create(alert)
	}

	return alerts
}

func Test_alertStorage_ListModelEndpointAlerts(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		_ = populateAlertTable(db)

		alertStorage := NewAlertStorage(db)

		actualAlerts, err := alertStorage.ListModelEndpointAlerts(models.Id(1))
		assert.Nil(t, err)
		assert.NotNil(t, actualAlerts)

		for _, actualAlert := range actualAlerts {
			assert.Equal(t, "1", actualAlert.ModelId.String())
		}
	})
}

func Test_alertStorage_GetModelEndpointAlert(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		_ = populateAlertTable(db)

		alertStorage := NewAlertStorage(db)

		actualAlert, err := alertStorage.GetModelEndpointAlert(models.Id(1), models.Id(1))
		assert.Nil(t, err)
		assert.NotNil(t, actualAlert)
		assert.Equal(t, "1", actualAlert.ModelId.String())
		assert.Equal(t, "1", actualAlert.ModelEndpointId.String())

		actualAlert, err = alertStorage.GetModelEndpointAlert(models.Id(2), models.Id(3))
		assert.Nil(t, err)
		assert.NotNil(t, actualAlert)
		assert.Equal(t, "2", actualAlert.ModelId.String())
		assert.Equal(t, "3", actualAlert.ModelEndpointId.String())
	})
}

func Test_alertStorage_CreateUpdateModelEndpointAlert(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		_ = populateAlertTable(db)

		alertStorage := NewAlertStorage(db)

		alertModel1Endpoint3 := &models.ModelEndpointAlert{
			ModelId:         1,
			ModelEndpointId: 3,
			EnvironmentName: "env-1",
			TeamName:        "datascience",
			AlertConditions: models.AlertConditions{
				&models.AlertCondition{
					Enabled:    true,
					MetricType: models.AlertConditionTypeThroughput,
					Severity:   models.AlertConditionSeverityWarning,
					Target:     10,
				},
				&models.AlertCondition{
					Enabled:    true,
					MetricType: models.AlertConditionTypeThroughput,
					Severity:   models.AlertConditionSeverityCritical,
					Target:     1,
				},
				&models.AlertCondition{
					Enabled:    true,
					MetricType: models.AlertConditionTypeLatency,
					Severity:   models.AlertConditionSeverityWarning,
					Target:     10,
					Percentile: 99,
					Unit:       "ms",
				},
				&models.AlertCondition{
					Enabled:    true,
					MetricType: models.AlertConditionTypeLatency,
					Severity:   models.AlertConditionSeverityCritical,
					Target:     15,
					Percentile: 99,
					Unit:       "ms",
				},
				&models.AlertCondition{
					Enabled:    true,
					MetricType: models.AlertConditionTypeErrorRate,
					Severity:   models.AlertConditionSeverityWarning,
					Target:     1,
					Unit:       "%",
				},
				&models.AlertCondition{
					Enabled:    true,
					MetricType: models.AlertConditionTypeErrorRate,
					Severity:   models.AlertConditionSeverityCritical,
					Target:     5,
					Unit:       "%",
				},
				&models.AlertCondition{
					Enabled:    true,
					MetricType: models.AlertConditionTypeCPU,
					Severity:   models.AlertConditionSeverityWarning,
					Target:     90,
					Unit:       "%",
				},
				&models.AlertCondition{
					Enabled:    true,
					MetricType: models.AlertConditionTypeCPU,
					Severity:   models.AlertConditionSeverityCritical,
					Target:     95,
					Unit:       "%",
				},
				&models.AlertCondition{
					Enabled:    true,
					MetricType: models.AlertConditionTypeMemory,
					Severity:   models.AlertConditionSeverityWarning,
					Target:     90,
					Unit:       "%",
				},
				&models.AlertCondition{
					Enabled:    true,
					MetricType: models.AlertConditionTypeMemory,
					Severity:   models.AlertConditionSeverityCritical,
					Target:     95,
					Unit:       "%",
				},
			},
		}

		err := alertStorage.CreateModelEndpointAlert(alertModel1Endpoint3)
		assert.Nil(t, err)

		actualAlert, err := alertStorage.GetModelEndpointAlert(models.Id(1), models.Id(3))
		assert.Nil(t, err)
		assert.NotNil(t, actualAlert)
		assert.Equal(t, "1", actualAlert.ModelId.String())
		assert.Equal(t, "3", actualAlert.ModelEndpointId.String())

		actualAlert.AlertConditions = models.AlertConditions{}

		err = alertStorage.UpdateModelEndpointAlert(actualAlert)

		updatedAlert, err := alertStorage.GetModelEndpointAlert(models.Id(1), models.Id(3))
		assert.Nil(t, err)
		assert.NotNil(t, actualAlert)
		assert.Empty(t, updatedAlert.AlertConditions)
	})
}

func Test_alertStorage_DeleteModelEndpointAlert(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		_ = populateAlertTable(db)

		alertStorage := NewAlertStorage(db)

		err := alertStorage.DeleteModelEndpointAlert(models.Id(1), models.Id(1))
		assert.Nil(t, err)

		actualAlerts, err := alertStorage.ListModelEndpointAlerts(models.Id(1))
		assert.Nil(t, err)
		assert.NotNil(t, actualAlerts)

		for _, actualAlert := range actualAlerts {
			if actualAlert.ModelId.String() == "1" && actualAlert.ModelEndpointId.String() == "1" {
				t.Errorf("Failed to delete model endpoint alert")
			}
		}
	})
}
