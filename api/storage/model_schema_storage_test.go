//go:build integration_local || integration
// +build integration_local integration

package storage

import (
	"context"
	"testing"

	"github.com/caraml-dev/merlin/database"
	"github.com/caraml-dev/merlin/mlp"
	"github.com/caraml-dev/merlin/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func Test_modelSchemaStorage_Save(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		env1 := models.Environment{
			Name:                   "env1",
			Cluster:                "k8s",
			IsPredictionJobEnabled: true,
		}
		db.Create(&env1)

		p := mlp.Project{
			ID:                1,
			Name:              "project",
			MLFlowTrackingURL: "http://mlflow:5000",
		}

		m := models.Model{
			ID:           1,
			ProjectID:    models.ID(p.ID),
			ExperimentID: 1,
			Name:         "model",
			Type:         models.ModelTypeSkLearn,
		}
		db.Create(&m)

		modelSchemaStorage := NewModelSchemaStorage(db)
		modelSchema := &models.ModelSchema{
			ModelID: m.ID,
			Spec: &models.SchemaSpec{
				PredictionIDColumn: "prediction_id",
				TagColumns:         []string{"tag"},
				FeatureTypes: map[string]models.ValueType{
					"featureA": models.Float64,
					"featureB": models.Float64,
					"featureC": models.Int64,
					"featureD": models.Boolean,
				},
				ModelPredictionOutput: &models.ModelPredictionOutput{
					BinaryClassificationOutput: &models.BinaryClassificationOutput{
						ActualLabelColumn:     "actual_label",
						NegativeClassLabel:    "negative",
						PositiveClassLabel:    "positive",
						PredictionLabelColumn: "prediction_label",
						PredictionScoreColumn: "prediction_score",
					},
				},
			},
		}
		schema, err := modelSchemaStorage.Save(context.Background(), modelSchema)
		require.NoError(t, err)
		assert.Equal(t, models.ID(1), schema.ID)

		schema.Spec.ModelPredictionOutput = &models.ModelPredictionOutput{
			RankingOutput: &models.RankingOutput{
				PredictionGroudIDColumn: "session_id",
				RankScoreColumn:         "score",
				RelevanceScoreColumn:    "relevance_score",
			},
		}
		_, err = modelSchemaStorage.Save(context.Background(), schema)
		require.NoError(t, err)
		newSchema, err := modelSchemaStorage.FindByID(context.Background(), models.ID(1), m.ID)
		require.NoError(t, err)
		assert.Equal(t, newSchema.Spec.ModelPredictionOutput, &models.ModelPredictionOutput{
			RankingOutput: &models.RankingOutput{
				PredictionGroudIDColumn: "session_id",
				RankScoreColumn:         "score",
				RelevanceScoreColumn:    "relevance_score",
			},
		})
	})
}

func Test_modelSchemaStorage_SaveThroughVersion(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		env1 := models.Environment{
			Name:                   "env1",
			Cluster:                "k8s",
			IsPredictionJobEnabled: true,
		}
		db.Create(&env1)

		p := mlp.Project{
			ID:                1,
			Name:              "project",
			MLFlowTrackingURL: "http://mlflow:5000",
		}

		m := models.Model{
			ID:           1,
			ProjectID:    models.ID(p.ID),
			ExperimentID: 1,
			Name:         "model",
			Type:         models.ModelTypeSkLearn,
		}
		db.Create(&m)

		version := &models.Version{
			ID:        models.ID(1),
			ModelID:   m.ID,
			Model:     &m,
			MlflowURL: "http://mlflow.com",
			ModelSchema: &models.ModelSchema{
				ModelID: m.ID,
				Spec: &models.SchemaSpec{
					PredictionIDColumn: "prediction_id",
					TagColumns:         []string{"tag"},
					FeatureTypes: map[string]models.ValueType{
						"featureA": models.Float64,
						"featureB": models.Float64,
						"featureC": models.Int64,
						"featureD": models.Boolean,
					},
					ModelPredictionOutput: &models.ModelPredictionOutput{
						BinaryClassificationOutput: &models.BinaryClassificationOutput{
							ActualLabelColumn:     "actual_label",
							NegativeClassLabel:    "negative",
							PositiveClassLabel:    "positive",
							PredictionLabelColumn: "prediction_label",
							PredictionScoreColumn: "prediction_score",
						},
					},
				},
			},
		}
		err := db.Save(&version).Error
		require.NoError(t, err)

		modelSchemaStorage := NewModelSchemaStorage(db)
		newSchema, err := modelSchemaStorage.FindByID(context.Background(), models.ID(1), m.ID)
		require.NoError(t, err)

		assert.Equal(t, &models.ModelPredictionOutput{
			BinaryClassificationOutput: &models.BinaryClassificationOutput{
				ActualLabelColumn:     "actual_label",
				NegativeClassLabel:    "negative",
				PositiveClassLabel:    "positive",
				PredictionLabelColumn: "prediction_label",
				PredictionScoreColumn: "prediction_score",
			},
		}, newSchema.Spec.ModelPredictionOutput)
		var versions []*models.Version
		db.Preload("Model").Preload("ModelSchema").Where("model_id = ?", m.ID).Find(&versions)
		assert.Equal(t, 1, len(versions))
		assert.Equal(t, models.ID(1), newSchema.ID)
		assert.Equal(t, &models.ModelPredictionOutput{
			BinaryClassificationOutput: &models.BinaryClassificationOutput{
				ActualLabelColumn:     "actual_label",
				NegativeClassLabel:    "negative",
				PositiveClassLabel:    "positive",
				PredictionLabelColumn: "prediction_label",
				PredictionScoreColumn: "prediction_score",
			},
		}, versions[0].ModelSchema.Spec.ModelPredictionOutput)
	})
}

func Test_modelSchemaStorage_FindAll_Delete(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		env1 := models.Environment{
			Name:                   "env1",
			Cluster:                "k8s",
			IsPredictionJobEnabled: true,
		}
		db.Create(&env1)

		p := mlp.Project{
			ID:                1,
			Name:              "project",
			MLFlowTrackingURL: "http://mlflow:5000",
		}

		m := models.Model{
			ID:           1,
			ProjectID:    models.ID(p.ID),
			ExperimentID: 1,
			Name:         "model",
			Type:         models.ModelTypeSkLearn,
		}
		db.Create(&m)

		modelSchemaStorage := NewModelSchemaStorage(db)
		modelSchemas := []*models.ModelSchema{
			{
				ModelID: m.ID,
				Spec: &models.SchemaSpec{
					PredictionIDColumn: "prediction_id",
					TagColumns:         []string{"tag"},
					FeatureTypes: map[string]models.ValueType{
						"featureA": models.Float64,
						"featureB": models.Float64,
						"featureC": models.Int64,
						"featureD": models.Boolean,
					},
					ModelPredictionOutput: &models.ModelPredictionOutput{
						BinaryClassificationOutput: &models.BinaryClassificationOutput{
							ActualLabelColumn:     "actual_label",
							NegativeClassLabel:    "negative",
							PositiveClassLabel:    "positive",
							PredictionLabelColumn: "prediction_label",
							PredictionScoreColumn: "prediction_score",
						},
					},
				},
			},
			{
				ModelID: m.ID,
				Spec: &models.SchemaSpec{
					PredictionIDColumn: "prediction_id",
					TagColumns:         []string{"tag"},
					FeatureTypes: map[string]models.ValueType{
						"featureA": models.Float64,
						"featureB": models.Float64,
						"featureC": models.Int64,
						"featureD": models.Boolean,
					},
					ModelPredictionOutput: &models.ModelPredictionOutput{
						RankingOutput: &models.RankingOutput{
							PredictionGroudIDColumn: "session_id",
							RankScoreColumn:         "score",
							RelevanceScoreColumn:    "relevance_score",
						},
					},
				},
			},
			{
				ModelID: m.ID,
				Spec: &models.SchemaSpec{
					PredictionIDColumn: "prediction_id",
					TagColumns:         []string{"tag"},
					FeatureTypes: map[string]models.ValueType{
						"featureA": models.Float64,
						"featureB": models.Float64,
						"featureC": models.Int64,
						"featureD": models.Boolean,
					},
					ModelPredictionOutput: &models.ModelPredictionOutput{
						RegressionOutput: &models.RegressionOutput{
							PredictionScoreColumn: "prediction_score",
							ActualScoreColumn:     "actual_score",
						},
					},
				},
			},
		}

		for _, schema := range modelSchemas {
			_, err := modelSchemaStorage.Save(context.Background(), schema)
			require.NoError(t, err)
		}

		allSchemas, err := modelSchemaStorage.FindAll(context.Background(), m.ID)
		require.NoError(t, err)
		assert.Equal(t, len(modelSchemas), len(allSchemas))
		for idx, schema := range allSchemas {
			assert.Equal(t, modelSchemas[idx], schema)
		}

		err = modelSchemaStorage.Delete(context.Background(), &models.ModelSchema{ID: 1, ModelID: m.ID})
		require.NoError(t, err)

		allSchemas, err = modelSchemaStorage.FindAll(context.Background(), m.ID)
		require.NoError(t, err)
		assert.Equal(t, 2, len(allSchemas))
	})
}

func Test_modelSchemaStorage_FindByID(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		env1 := models.Environment{
			Name:                   "env1",
			Cluster:                "k8s",
			IsPredictionJobEnabled: true,
		}
		db.Create(&env1)

		p := mlp.Project{
			ID:                1,
			Name:              "project",
			MLFlowTrackingURL: "http://mlflow:5000",
		}

		m := models.Model{
			ID:           1,
			ProjectID:    models.ID(p.ID),
			ExperimentID: 1,
			Name:         "model",
			Type:         models.ModelTypeSkLearn,
		}
		db.Create(&m)

		modelSchemaStorage := NewModelSchemaStorage(db)
		modelSchema := &models.ModelSchema{
			ModelID: m.ID,
			Spec: &models.SchemaSpec{
				PredictionIDColumn: "prediction_id",
				TagColumns:         []string{"tag"},
				FeatureTypes: map[string]models.ValueType{
					"featureA": models.Float64,
					"featureB": models.Float64,
					"featureC": models.Int64,
					"featureD": models.Boolean,
				},
				ModelPredictionOutput: &models.ModelPredictionOutput{
					BinaryClassificationOutput: &models.BinaryClassificationOutput{
						ActualLabelColumn:     "actual_label",
						NegativeClassLabel:    "negative",
						PositiveClassLabel:    "positive",
						PredictionLabelColumn: "prediction_label",
						PredictionScoreColumn: "prediction_score",
					},
				},
			},
		}
		schema, err := modelSchemaStorage.Save(context.Background(), modelSchema)
		require.NoError(t, err)
		assert.Equal(t, models.ID(1), schema.ID)

		schema.Spec.ModelPredictionOutput = &models.ModelPredictionOutput{
			RankingOutput: &models.RankingOutput{
				PredictionGroudIDColumn: "session_id",
				RankScoreColumn:         "score",
				RelevanceScoreColumn:    "relevance_score",
			},
		}
		_, err = modelSchemaStorage.Save(context.Background(), schema)
		require.NoError(t, err)
		newSchema, err := modelSchemaStorage.FindByID(context.Background(), models.ID(1), m.ID)
		require.NoError(t, err)
		assert.Equal(t, newSchema.Spec.ModelPredictionOutput, &models.ModelPredictionOutput{
			RankingOutput: &models.RankingOutput{
				PredictionGroudIDColumn: "session_id",
				RankScoreColumn:         "score",
				RelevanceScoreColumn:    "relevance_score",
			},
		})

		schema, err = modelSchemaStorage.FindByID(context.Background(), models.ID(2), m.ID)
		require.Equal(t, gorm.ErrRecordNotFound, err)
	})
}
