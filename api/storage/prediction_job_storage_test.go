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

//go:build integration_local || integration
// +build integration_local integration

package storage

import (
	"fmt"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"

	"github.com/gojek/merlin-pyspark-app/pkg/spec"
	"github.com/gojek/merlin/it/database"
	"github.com/gojek/merlin/mlp"
	"github.com/gojek/merlin/models"
)

func TestPredictionJobStorage_SaveAndGet(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		predJobStore := NewPredictionJobStorage(db)
		isDefaultPredictionJob := true
		env1 := models.Environment{
			Name:                   "env1",
			Cluster:                "k8s",
			IsPredictionJobEnabled: true,
			IsDefaultPredictionJob: &isDefaultPredictionJob,
		}
		db.Create(&env1)

		p := mlp.Project{
			ID:                1,
			Name:              "project",
			MLFlowTrackingURL: "http://mlflow:5000",
		}
		db.Create(&p)

		m := models.Model{
			ID:           1,
			ProjectID:    models.ID(p.ID),
			ExperimentID: 1,
			Name:         "model",
			Type:         models.ModelTypeSkLearn,
		}
		db.Create(&m)

		v := models.Version{
			ModelID:     m.ID,
			RunID:       "1",
			ArtifactURI: "gcs:/mlp/1/1",
		}
		db.Create(&v)

		job := &models.PredictionJob{
			ID:   1,
			Name: fmt.Sprintf("%s-%s-%s", m.Name, v.ID, time.Now()),
			Metadata: models.Metadata{
				Team:   "dsp",
				Stream: "dsp",
				App:    "my-model",
				Labels: mlp.Labels{
					{
						Key:   "my-key",
						Value: "my-value",
					},
				},
			},
			VersionID:       v.ID,
			VersionModelID:  m.ID,
			ProjectID:       models.ID(p.ID),
			EnvironmentName: env1.Name,
			Environment:     &env1,
			Config: &models.Config{
				JobConfig: &spec.PredictionJob{
					Version: "v1",
					Kind:    "prediction_job",
					Name:    "my_prediction_job",
					Source: &spec.PredictionJob_BigquerySource{
						BigquerySource: &spec.BigQuerySource{
							Table:    "project.dataset.source_table",
							Features: []string{"feature1", "feature2", "feature3"},
							Options: map[string]string{
								"src_option1": "value1",
							},
						},
					},
					Model: &spec.Model{
						Type: spec.ModelType_PYFUNC_V2,
						Uri:  "gs://test-bucket",
						Result: &spec.Model_ModelResult{
							Type: spec.ResultType_INTEGER,
						},
						Options: map[string]string{
							"model_option1": "value1",
						},
					},
					Sink: &spec.PredictionJob_BigquerySink{BigquerySink: &spec.BigQuerySink{
						Table:         "project.dataset.sink_table",
						StagingBucket: "gs://staging_bucket",
						ResultColumn:  "prediction",
						SaveMode:      spec.SaveMode_OVERWRITE,
						Options: map[string]string{
							"sink_option1": "value1",
						},
					}},
				},
			},
			Status: models.JobPending,
			Error:  "no-error",
			CreatedUpdated: models.CreatedUpdated{
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}

		err := predJobStore.Save(job)
		assert.NoError(t, err)

		loaded, err := predJobStore.Get(job.ID)
		assert.NoError(t, err)

		assert.Equal(t, job.ID, loaded.ID)
		assert.Equal(t, job.VersionID, loaded.VersionID)
		assert.Equal(t, job.VersionModelID, loaded.VersionModelID)
		assert.Equal(t, job.Config, loaded.Config)
		assert.Equal(t, job.Status, loaded.Status)
		assert.Equal(t, job.Error, loaded.Error)
		assert.Equal(t, job.Environment.Name, loaded.Environment.Name)
		assert.Equal(t, job.EnvironmentName, loaded.EnvironmentName)
		assert.Equal(t, job.Metadata, loaded.Metadata)
	})
}

func TestPredictionJobStorage_List(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		predJobStore := NewPredictionJobStorage(db)
		isDefaultPredictionJob := true

		env1 := models.Environment{
			Name:                   "env1",
			Cluster:                "k8s",
			IsPredictionJobEnabled: true,
			IsDefaultPredictionJob: &isDefaultPredictionJob,
		}
		db.Create(&env1)

		p := mlp.Project{
			Name:              "project",
			MLFlowTrackingURL: "http://mlflow:5000",
		}
		db.Create(&p)

		m := models.Model{
			ID:           1,
			ProjectID:    models.ID(p.ID),
			ExperimentID: 1,
			Name:         "model",
			Type:         models.ModelTypeSkLearn,
		}
		db.Create(&m)

		v := models.Version{
			ModelID:     m.ID,
			RunID:       "1",
			ArtifactURI: "gcs:/mlp/1/1",
		}
		db.Create(&v)

		job1 := &models.PredictionJob{
			ID:   1,
			Name: fmt.Sprintf("%s-%s-%s", m.Name, v.ID, time.Now()),
			Metadata: models.Metadata{
				Team:   "dsp",
				Stream: "dsp",
				App:    "my-model",
				Labels: mlp.Labels{
					{
						Key:   "my-key",
						Value: "my-value",
					},
				},
			},
			VersionID:       v.ID,
			VersionModelID:  m.ID,
			ProjectID:       models.ID(p.ID),
			Environment:     &env1,
			EnvironmentName: env1.Name,
			Config: &models.Config{
				JobConfig: &spec.PredictionJob{
					Version: "v1",
					Kind:    "prediction_job",
					Name:    "my_prediction_job",
					Source: &spec.PredictionJob_BigquerySource{
						BigquerySource: &spec.BigQuerySource{
							Table:    "project.dataset.source_table",
							Features: []string{"feature1", "feature2", "feature3"},
							Options: map[string]string{
								"src_option1": "value1",
							},
						},
					},
					Model: &spec.Model{
						Type: spec.ModelType_PYFUNC_V2,
						Uri:  "gs://test-bucket",
						Result: &spec.Model_ModelResult{
							Type: spec.ResultType_INTEGER,
						},
						Options: map[string]string{
							"model_option1": "value1",
						},
					},
					Sink: &spec.PredictionJob_BigquerySink{BigquerySink: &spec.BigQuerySink{
						Table:         "project.dataset.sink_table",
						StagingBucket: "gs://staging_bucket",
						ResultColumn:  "prediction",
						SaveMode:      spec.SaveMode_OVERWRITE,
						Options: map[string]string{
							"sink_option1": "value1",
						},
					}},
				},
			},
			Status: models.JobPending,
			Error:  "no-error",
			CreatedUpdated: models.CreatedUpdated{
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}

		job2 := &models.PredictionJob{
			ID:   2,
			Name: fmt.Sprintf("%s-%s-%s", m.Name, v.ID, time.Now()),
			Metadata: models.Metadata{
				Team:   "dsp",
				Stream: "dsp",
				App:    "my-model",
				Labels: mlp.Labels{
					{
						Key:   "my-key",
						Value: "my-value",
					},
				},
			},
			VersionID:       v.ID,
			VersionModelID:  m.ID,
			ProjectID:       models.ID(p.ID),
			Environment:     &env1,
			EnvironmentName: env1.Name,
			Config: &models.Config{
				JobConfig: &spec.PredictionJob{
					Version: "v1",
					Kind:    "prediction_job",
					Name:    "my_prediction_job",
					Source: &spec.PredictionJob_BigquerySource{
						BigquerySource: &spec.BigQuerySource{
							Table:    "project.dataset.source_table",
							Features: []string{"feature1", "feature2", "feature3"},
							Options: map[string]string{
								"src_option1": "value1",
							},
						},
					},
					Model: &spec.Model{
						Type: spec.ModelType_PYFUNC_V2,
						Uri:  "gs://test-bucket",
						Result: &spec.Model_ModelResult{
							Type: spec.ResultType_INTEGER,
						},
						Options: map[string]string{
							"model_option1": "value1",
						},
					},
					Sink: &spec.PredictionJob_BigquerySink{BigquerySink: &spec.BigQuerySink{
						Table:         "project.dataset.sink_table",
						StagingBucket: "gs://staging_bucket",
						ResultColumn:  "prediction",
						SaveMode:      spec.SaveMode_OVERWRITE,
						Options: map[string]string{
							"sink_option1": "value1",
						},
					}},
				},
			},
			Status: models.JobCompleted,
			Error:  "no-error",
			CreatedUpdated: models.CreatedUpdated{
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}

		err := predJobStore.Save(job1)
		assert.NoError(t, err)

		err = predJobStore.Save(job2)
		assert.NoError(t, err)

		jobs, err := predJobStore.List(&models.PredictionJob{
			ProjectID: models.ID(p.ID),
		})
		assert.NoError(t, err)
		assert.Len(t, jobs, 2)

		jobs, err = predJobStore.List(&models.PredictionJob{
			VersionModelID: m.ID,
		})
		assert.NoError(t, err)
		assert.Len(t, jobs, 2)

		jobs, err = predJobStore.List(&models.PredictionJob{
			VersionID: v.ID,
		})
		assert.NoError(t, err)
		assert.Len(t, jobs, 2)
	})
}
