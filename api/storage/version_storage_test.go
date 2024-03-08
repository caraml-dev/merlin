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

func Test_versionStorage_FindByID(t *testing.T) {
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

		modelSchema := &models.ModelSchema{
			ID:      models.ID(1),
			ModelID: m.ID,
			Spec: &models.SchemaSpec{
				SessionIDColumn: "prediction_id",
				RowIDColumn:     "row",
				TagColumns:      []string{"tag"},
				FeatureTypes: map[string]models.ValueType{
					"featureA": models.Float64,
					"featureB": models.Float64,
					"featureC": models.Int64,
					"featureD": models.Boolean,
				},
				ModelPredictionOutput: &models.ModelPredictionOutput{
					BinaryClassificationOutput: &models.BinaryClassificationOutput{
						ActualScoreColumn:     "actual_score",
						NegativeClassLabel:    "negative",
						PositiveClassLabel:    "positive",
						PredictionLabelColumn: "prediction_label",
						PredictionScoreColumn: "prediction_score",
						OutputClass:           models.BinaryClassification,
					},
				},
			},
		}
		version := models.Version{
			ID:          models.ID(1),
			ModelID:     m.ID,
			ModelSchema: modelSchema,
		}
		db.Create(&version)

		storage := NewVersionStorage(db)
		v, err := storage.FindByID(context.Background(), version.ID, m.ID)
		require.NoError(t, err)
		assert.Equal(t, v.ID, version.ID)
		assert.Equal(t, &modelSchema.ID, version.ModelSchemaID)

		v, err = storage.FindByID(context.Background(), 2, m.ID)

		require.NoError(t, err)
		assert.Nil(t, v)

	})
}
