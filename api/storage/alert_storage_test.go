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

	"github.com/caraml-dev/merlin/it/database"
	"github.com/caraml-dev/merlin/mlp"
	"github.com/caraml-dev/merlin/models"
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
		ID:           1,
		ProjectID:    models.ID(project1.Id),
		ExperimentID: 1,
		Name:         "model-1",
		Type:         models.ModelTypeSkLearn,
	}
	db.Create(model1)

	model2 := &models.Model{
		ID:           2,
		ProjectID:    models.ID(project1.Id),
		ExperimentID: 2,
		Name:         "model-2",
		Type:         models.ModelTypeSkLearn,
	}
	db.Create(model2)

	model1Endpoint1 := &models.ModelEndpoint{
		ID:              1,
		ModelID:         model1.ID,
		Status:          models.EndpointServing,
		EnvironmentName: env1.Name,
	}
	db.Create(model1Endpoint1)

	model1Endpoint2 := &models.ModelEndpoint{
		ID:              2,
		ModelID:         model1.ID,
		Status:          models.EndpointServing,
		EnvironmentName: env1.Name,
	}
	db.Create(model1Endpoint2)

	model2Endpoint3 := &models.ModelEndpoint{
		ID:              3,
		ModelID:         model1.ID,
		Status:          models.EndpointServing,
		EnvironmentName: env1.Name,
	}
	db.Create(model2Endpoint3)

	alerts := []*models.ModelEndpointAlert{}

	alertModel1Endpoint1 := &models.ModelEndpointAlert{
		ModelID:         model1.ID,
		ModelEndpointID: model1Endpoint1.ID,
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
		ModelID:         model1.ID,
		ModelEndpointID: model1Endpoint2.ID,
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
		ModelID:         model2.ID,
		ModelEndpointID: model2Endpoint3.ID,
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

		actualAlerts, err := alertStorage.ListModelEndpointAlerts(models.ID(1))
		assert.Nil(t, err)
		assert.NotNil(t, actualAlerts)

		for _, actualAlert := range actualAlerts {
			assert.Equal(t, "1", actualAlert.ModelID.String())
		}
	})
}

func Test_alertStorage_GetModelEndpointAlert(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		_ = populateAlertTable(db)

		alertStorage := NewAlertStorage(db)

		actualAlert, err := alertStorage.GetModelEndpointAlert(models.ID(1), models.ID(1))
		assert.Nil(t, err)
		assert.NotNil(t, actualAlert)
		assert.Equal(t, "1", actualAlert.ModelID.String())
		assert.Equal(t, "1", actualAlert.ModelEndpointID.String())

		actualAlert, err = alertStorage.GetModelEndpointAlert(models.ID(2), models.ID(3))
		assert.Nil(t, err)
		assert.NotNil(t, actualAlert)
		assert.Equal(t, "2", actualAlert.ModelID.String())
		assert.Equal(t, "3", actualAlert.ModelEndpointID.String())
	})
}

func Test_alertStorage_CreateUpdateModelEndpointAlert(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		_ = populateAlertTable(db)

		alertStorage := NewAlertStorage(db)

		alertModel1Endpoint3 := &models.ModelEndpointAlert{
			ModelID:         1,
			ModelEndpointID: 3,
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

		actualAlert, err := alertStorage.GetModelEndpointAlert(models.ID(1), models.ID(3))
		assert.Nil(t, err)
		assert.NotNil(t, actualAlert)
		assert.Equal(t, "1", actualAlert.ModelID.String())
		assert.Equal(t, "3", actualAlert.ModelEndpointID.String())

		actualAlert.AlertConditions = models.AlertConditions{}

		err = alertStorage.UpdateModelEndpointAlert(actualAlert)

		updatedAlert, err := alertStorage.GetModelEndpointAlert(models.ID(1), models.ID(3))
		assert.Nil(t, err)
		assert.NotNil(t, actualAlert)
		assert.Empty(t, updatedAlert.AlertConditions)
	})
}

func Test_alertStorage_DeleteModelEndpointAlert(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		_ = populateAlertTable(db)

		alertStorage := NewAlertStorage(db)

		err := alertStorage.DeleteModelEndpointAlert(models.ID(1), models.ID(1))
		assert.Nil(t, err)

		actualAlerts, err := alertStorage.ListModelEndpointAlerts(models.ID(1))
		assert.Nil(t, err)
		assert.NotNil(t, actualAlerts)

		for _, actualAlert := range actualAlerts {
			if actualAlert.ModelID.String() == "1" && actualAlert.ModelEndpointID.String() == "1" {
				t.Errorf("Failed to delete model endpoint alert")
			}
		}
	})
}
