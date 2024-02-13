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

func Test_obsPublisherStorage(t *testing.T) {
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

		observabilityPublisherStorage := NewObservabilityPublisherStorage(db)
		observabilityPublisher := &models.ObservabilityPublisher{
			VersionModelID:  m.ID,
			VersionID:       version.ID,
			Status:          models.Pending,
			ModelSchemaSpec: version.ModelSchema.Spec,
		}

		ctx := context.Background()
		publisher, err := observabilityPublisherStorage.Create(ctx, observabilityPublisher)
		require.NoError(t, err)
		assert.Equal(t, models.ID(1), publisher.ID)
		assert.Equal(t, 1, publisher.Revision)

		publisherFromDB, fetchErr := observabilityPublisherStorage.GetByModelID(ctx, m.ID)
		require.NoError(t, fetchErr)
		assert.Equal(t, publisher.ID, publisherFromDB.ID)
		assert.Equal(t, publisher.Revision, publisherFromDB.Revision)
		assert.Equal(t, publisher.ModelSchemaSpec, publisherFromDB.ModelSchemaSpec)

		noPublisher, fetchErr := observabilityPublisherStorage.GetByModelID(ctx, models.ID(2))
		require.NoError(t, fetchErr)
		assert.Nil(t, noPublisher)

		publisherFromDB, fetchErr = observabilityPublisherStorage.Get(ctx, publisherFromDB.ID)
		require.NoError(t, fetchErr)
		assert.Equal(t, publisher.ID, publisherFromDB.ID)
		assert.Equal(t, publisher.Revision, publisherFromDB.Revision)
		assert.Equal(t, publisher.ModelSchemaSpec, publisherFromDB.ModelSchemaSpec)

		signal := make(chan bool, 1)
		go func(p *models.ObservabilityPublisher) {
			_p := *p
			updatedPublisher, err := observabilityPublisherStorage.Update(ctx, &_p, true)
			signal <- true
			require.NoError(t, err)
			assert.Equal(t, p.ID, updatedPublisher.ID)
			assert.Equal(t, p.Revision+1, updatedPublisher.Revision)
		}(publisherFromDB)

		<-signal

		// case when update when revision is not matched anymore
		// update with revision 1 but the record already revision 2
		publisherFromDB.Status = models.Running
		conflictedPublisher, err := observabilityPublisherStorage.Update(ctx, publisherFromDB, false)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
		assert.Nil(t, conflictedPublisher)

		publisherFromDB, fetchErr = observabilityPublisherStorage.Get(ctx, publisherFromDB.ID)
		require.NoError(t, fetchErr)
		assert.Equal(t, 2, publisherFromDB.Revision)

	})
}
