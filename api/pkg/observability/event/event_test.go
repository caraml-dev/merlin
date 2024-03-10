package event

import (
	"fmt"
	"testing"

	"github.com/caraml-dev/merlin/mlp"
	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/queue"
	queueMock "github.com/caraml-dev/merlin/queue/mocks"
	storageMock "github.com/caraml-dev/merlin/storage/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_eventProducer_ModelEndpointChangeEvent(t *testing.T) {
	model := &models.Model{
		ID:   models.ID(1),
		Name: "model-1",
		Project: mlp.Project{
			Name:   "project-1",
			Stream: "stream",
			Team:   "team",
		},
		ObservabilitySupported: true,
	}
	schemaSpec := &models.SchemaSpec{
		TagColumns: []string{"tag"},
		FeatureTypes: map[string]models.ValueType{
			"featureA": models.Float64,
			"featureB": models.Float64,
			"featureC": models.Int64,
			"featureD": models.Boolean,
		},
		ModelPredictionOutput: &models.ModelPredictionOutput{
			BinaryClassificationOutput: &models.BinaryClassificationOutput{
				NegativeClassLabel:    "negative",
				PositiveClassLabel:    "positive",
				PredictionLabelColumn: "prediction_label",
				PredictionScoreColumn: "prediction_score",
				OutputClass:           models.BinaryClassification,
			},
		},
	}

	regresionSchemaSpec := &models.SchemaSpec{
		TagColumns: []string{"tag"},
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
				OutputClass:           models.Regression,
			},
		},
	}

	modelSchema := &models.ModelSchema{
		ModelID: model.ID,
		ID:      models.ID(1),
		Spec:    schemaSpec,
	}
	tests := []struct {
		name                          string
		jobProducer                   *queueMock.Producer
		observabilityPublisherStorage *storageMock.ObservabilityPublisherStorage
		versionStorage                *storageMock.VersionStorage
		modelEndpoint                 *models.ModelEndpoint
		model                         *models.Model
		expectedError                 error
	}{
		{
			name: "do nothing if model doesn't support model observability",
			jobProducer: func() *queueMock.Producer {
				producer := &queueMock.Producer{}
				return producer
			}(),
			observabilityPublisherStorage: func() *storageMock.ObservabilityPublisherStorage {
				mockStorage := &storageMock.ObservabilityPublisherStorage{}
				return mockStorage
			}(),
			versionStorage: func() *storageMock.VersionStorage {
				mockStorage := &storageMock.VersionStorage{}
				return mockStorage
			}(),
			model: &models.Model{
				ID:                     models.ID(2),
				ObservabilitySupported: false,
			},
			modelEndpoint: &models.ModelEndpoint{
				ID:      models.ID(1),
				ModelID: model.ID,
				Model:   model,
				Status:  models.EndpointServing,
				Rule: &models.ModelEndpointRule{
					Destination: []*models.ModelEndpointRuleDestination{
						{
							VersionEndpoint: &models.VersionEndpoint{
								ID:                       uuid.UUID{},
								VersionID:                models.ID(2),
								Status:                   models.EndpointServing,
								EnableModelObservability: true,
							},
						},
					},
				},
			},
		},
		{
			name: "no deployment; version endpoint model observability is disabled and never been deployed before",
			jobProducer: func() *queueMock.Producer {
				producer := &queueMock.Producer{}
				return producer
			}(),
			observabilityPublisherStorage: func() *storageMock.ObservabilityPublisherStorage {
				mockStorage := &storageMock.ObservabilityPublisherStorage{}
				mockStorage.On("GetByModelID", mock.Anything, models.ID(1)).Return(nil, nil)
				return mockStorage
			}(),
			versionStorage: func() *storageMock.VersionStorage {
				mockStorage := &storageMock.VersionStorage{}
				return mockStorage
			}(),
			model: model,
			modelEndpoint: &models.ModelEndpoint{
				ID:      models.ID(1),
				ModelID: model.ID,
				Model:   model,
				Status:  models.EndpointServing,
				Rule: &models.ModelEndpointRule{
					Destination: []*models.ModelEndpointRuleDestination{
						{
							VersionEndpoint: &models.VersionEndpoint{
								ID:                       uuid.UUID{},
								VersionID:                models.ID(2),
								Status:                   models.EndpointServing,
								EnableModelObservability: false,
							},
						},
					},
				},
			},
		},
		{
			name: "fresh deployment",
			jobProducer: func() *queueMock.Producer {
				producer := &queueMock.Producer{}
				producer.On("EnqueueJob", &queue.Job{
					Name: ObservabilityPublisherDeployment,
					Arguments: queue.Arguments{
						dataArgKey: models.ObservabilityPublisherJob{
							ActionType: models.DeployPublisher,
							Publisher: &models.ObservabilityPublisher{
								ID:              models.ID(1),
								VersionID:       models.ID(2),
								VersionModelID:  models.ID(1),
								ModelSchemaSpec: schemaSpec,
								Revision:        1,
								Status:          models.Pending,
							},
							WorkerData: &models.WorkerData{
								Namespace:       "project-1",
								ModelSchemaSpec: schemaSpec,
								Metadata: models.Metadata{
									App:       "model-1-observability-publisher",
									Component: "worker",
									Stream:    "stream",
									Team:      "team",
								},
								ModelName:    "model-1",
								ModelVersion: "2",
								Revision:     1,
								TopicSource:  "caraml-project-1-model-1-2-prediction-log",
							},
						},
					},
				}).Return(nil)
				return producer
			}(),
			observabilityPublisherStorage: func() *storageMock.ObservabilityPublisherStorage {
				mockStorage := &storageMock.ObservabilityPublisherStorage{}
				mockStorage.On("GetByModelID", mock.Anything, models.ID(1)).Return(nil, nil)
				mockStorage.On("Create", mock.Anything, &models.ObservabilityPublisher{
					VersionID:       models.ID(2),
					VersionModelID:  models.ID(1),
					Status:          models.Pending,
					Revision:        1,
					ModelSchemaSpec: schemaSpec,
				}).Return(&models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionID:       models.ID(2),
					VersionModelID:  models.ID(1),
					Revision:        1,
					Status:          models.Pending,
					ModelSchemaSpec: schemaSpec,
				}, nil)
				return mockStorage
			}(),
			versionStorage: func() *storageMock.VersionStorage {
				mockStorage := &storageMock.VersionStorage{}
				mockStorage.On("FindByID", mock.Anything, models.ID(2), model.ID).Return(&models.Version{
					ID:          models.ID(2),
					ModelID:     model.ID,
					ModelSchema: modelSchema,
					Model:       model,
				}, nil)
				return mockStorage
			}(),
			model: model,
			modelEndpoint: &models.ModelEndpoint{
				ID:      models.ID(1),
				ModelID: model.ID,
				Model:   model,
				Status:  models.EndpointServing,
				Rule: &models.ModelEndpointRule{
					Destination: []*models.ModelEndpointRuleDestination{
						{
							VersionEndpoint: &models.VersionEndpoint{
								ID:                       uuid.UUID{},
								VersionID:                models.ID(2),
								Status:                   models.EndpointServing,
								EnableModelObservability: true,
							},
						},
					},
				},
			},
		},
		{
			name: "fresh deployment request failed - version doesn't have schema ",
			jobProducer: func() *queueMock.Producer {
				producer := &queueMock.Producer{}
				return producer
			}(),
			observabilityPublisherStorage: func() *storageMock.ObservabilityPublisherStorage {
				mockStorage := &storageMock.ObservabilityPublisherStorage{}
				mockStorage.On("GetByModelID", mock.Anything, models.ID(1)).Return(nil, nil)
				return mockStorage
			}(),
			versionStorage: func() *storageMock.VersionStorage {
				mockStorage := &storageMock.VersionStorage{}
				mockStorage.On("FindByID", mock.Anything, models.ID(3), model.ID).Return(&models.Version{
					ID:          models.ID(3),
					ModelID:     model.ID,
					ModelSchema: nil,
					Model:       model,
				}, nil)
				return mockStorage
			}(),
			model: model,
			modelEndpoint: &models.ModelEndpoint{
				ID:      models.ID(1),
				ModelID: model.ID,
				Model:   model,
				Status:  models.EndpointServing,
				Rule: &models.ModelEndpointRule{
					Destination: []*models.ModelEndpointRuleDestination{
						{
							VersionEndpoint: &models.VersionEndpoint{
								ID:                       uuid.UUID{},
								VersionID:                models.ID(3),
								Status:                   models.EndpointServing,
								EnableModelObservability: true,
							},
						},
					},
				},
			},
			expectedError: fmt.Errorf("versionID: 3 in modelID: 1 doesn't have model schema"),
		},
		{
			name: "redeployment - model endpoint change version",
			jobProducer: func() *queueMock.Producer {
				producer := &queueMock.Producer{}
				producer.On("EnqueueJob", &queue.Job{
					Name: ObservabilityPublisherDeployment,
					Arguments: queue.Arguments{
						dataArgKey: models.ObservabilityPublisherJob{
							ActionType: models.DeployPublisher,
							Publisher: &models.ObservabilityPublisher{
								ID:              models.ID(1),
								VersionID:       models.ID(3),
								VersionModelID:  models.ID(1),
								ModelSchemaSpec: regresionSchemaSpec,
								Revision:        2,
								Status:          models.Pending,
							},
							WorkerData: &models.WorkerData{
								Namespace:       "project-1",
								ModelSchemaSpec: regresionSchemaSpec,
								Metadata: models.Metadata{
									App:       "model-1-observability-publisher",
									Component: "worker",
									Stream:    "stream",
									Team:      "team",
								},
								ModelName:    "model-1",
								ModelVersion: "3",
								Revision:     2,
								TopicSource:  "caraml-project-1-model-1-3-prediction-log",
							},
						},
					},
				}).Return(nil)
				return producer
			}(),
			observabilityPublisherStorage: func() *storageMock.ObservabilityPublisherStorage {
				mockStorage := &storageMock.ObservabilityPublisherStorage{}
				mockStorage.On("GetByModelID", mock.Anything, models.ID(1)).Return(&models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionID:       models.ID(2),
					VersionModelID:  models.ID(1),
					Status:          models.Pending,
					ModelSchemaSpec: schemaSpec,
				}, nil)
				mockStorage.On("Update", mock.Anything, &models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionID:       models.ID(3),
					VersionModelID:  models.ID(1),
					Status:          models.Pending,
					ModelSchemaSpec: regresionSchemaSpec,
				}, true).Return(&models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionID:       models.ID(3),
					VersionModelID:  models.ID(1),
					Revision:        2,
					Status:          models.Pending,
					ModelSchemaSpec: regresionSchemaSpec,
				}, nil)
				return mockStorage
			}(),
			versionStorage: func() *storageMock.VersionStorage {
				mockStorage := &storageMock.VersionStorage{}
				mockStorage.On("FindByID", mock.Anything, models.ID(3), model.ID).Return(&models.Version{
					ID:      models.ID(3),
					ModelID: model.ID,
					ModelSchema: &models.ModelSchema{
						ID:      models.ID(1),
						ModelID: model.ID,
						Spec:    regresionSchemaSpec,
					},
					Model: model,
				}, nil)
				return mockStorage
			}(),
			model: model,
			modelEndpoint: &models.ModelEndpoint{
				ID:      models.ID(1),
				ModelID: model.ID,
				Model:   model,
				Status:  models.EndpointServing,
				Rule: &models.ModelEndpointRule{
					Destination: []*models.ModelEndpointRuleDestination{
						{
							VersionEndpoint: &models.VersionEndpoint{
								ID:                       uuid.UUID{},
								VersionID:                models.ID(3),
								Status:                   models.EndpointServing,
								EnableModelObservability: true,
							},
						},
					},
				},
			},
		},
		{
			name: "redeployment request failed - failed get version ",
			jobProducer: func() *queueMock.Producer {
				producer := &queueMock.Producer{}
				return producer
			}(),
			observabilityPublisherStorage: func() *storageMock.ObservabilityPublisherStorage {
				mockStorage := &storageMock.ObservabilityPublisherStorage{}
				mockStorage.On("GetByModelID", mock.Anything, models.ID(1)).Return(&models.ObservabilityPublisher{
					ID:             models.ID(1),
					VersionID:      models.ID(2),
					VersionModelID: model.ID,
					Revision:       1,
					Status:         models.Running,
				}, nil)
				return mockStorage
			}(),
			versionStorage: func() *storageMock.VersionStorage {
				mockStorage := &storageMock.VersionStorage{}
				mockStorage.On("FindByID", mock.Anything, models.ID(4), model.ID).Return(nil, fmt.Errorf("connection is broken"))
				return mockStorage
			}(),
			model: model,
			modelEndpoint: &models.ModelEndpoint{
				ID:      models.ID(1),
				ModelID: model.ID,
				Model:   model,
				Status:  models.EndpointServing,
				Rule: &models.ModelEndpointRule{
					Destination: []*models.ModelEndpointRuleDestination{
						{
							VersionEndpoint: &models.VersionEndpoint{
								ID:                       uuid.UUID{},
								VersionID:                models.ID(4),
								Status:                   models.EndpointServing,
								EnableModelObservability: true,
							},
						},
					},
				},
			},
			expectedError: fmt.Errorf("connection is broken"),
		},
		{
			name: "undeployment - model endpoint is nil",
			jobProducer: func() *queueMock.Producer {
				producer := &queueMock.Producer{}
				producer.On("EnqueueJob", &queue.Job{
					Name: ObservabilityPublisherDeployment,
					Arguments: queue.Arguments{
						dataArgKey: models.ObservabilityPublisherJob{
							ActionType: models.UndeployPublisher,
							Publisher: &models.ObservabilityPublisher{
								ID:              models.ID(1),
								VersionID:       models.ID(2),
								VersionModelID:  models.ID(1),
								ModelSchemaSpec: schemaSpec,
								Revision:        1,
								Status:          models.Pending,
							},
							WorkerData: &models.WorkerData{
								Namespace:       "project-1",
								ModelSchemaSpec: schemaSpec,
								Metadata: models.Metadata{
									App:       "model-1-observability-publisher",
									Component: "worker",
									Stream:    "stream",
									Team:      "team",
								},
								ModelName:    "model-1",
								ModelVersion: "2",
								Revision:     1,
								TopicSource:  "caraml-project-1-model-1-2-prediction-log",
							},
						},
					},
				}).Return(nil)
				return producer
			}(),
			observabilityPublisherStorage: func() *storageMock.ObservabilityPublisherStorage {
				mockStorage := &storageMock.ObservabilityPublisherStorage{}
				mockStorage.On("GetByModelID", mock.Anything, models.ID(1)).Return(&models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionID:       models.ID(2),
					VersionModelID:  models.ID(1),
					Status:          models.Running,
					Revision:        1,
					ModelSchemaSpec: schemaSpec,
				}, nil)
				mockStorage.On("Update", mock.Anything, &models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionID:       models.ID(2),
					VersionModelID:  models.ID(1),
					Status:          models.Pending,
					Revision:        1,
					ModelSchemaSpec: schemaSpec,
				}, false).Return(&models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionID:       models.ID(2),
					VersionModelID:  models.ID(1),
					Revision:        1,
					Status:          models.Pending,
					ModelSchemaSpec: schemaSpec,
				}, nil)
				return mockStorage
			}(),
			versionStorage: func() *storageMock.VersionStorage {
				mockStorage := &storageMock.VersionStorage{}
				mockStorage.On("FindByID", mock.Anything, models.ID(2), model.ID).Return(&models.Version{
					ID:          models.ID(2),
					ModelID:     model.ID,
					ModelSchema: modelSchema,
					Model:       model,
				}, nil)
				return mockStorage
			}(),
			model:         model,
			modelEndpoint: nil,
		},
		{
			name: "undeployment - model observability is disabled",
			jobProducer: func() *queueMock.Producer {
				producer := &queueMock.Producer{}
				producer.On("EnqueueJob", &queue.Job{
					Name: ObservabilityPublisherDeployment,
					Arguments: queue.Arguments{
						dataArgKey: models.ObservabilityPublisherJob{
							ActionType: models.UndeployPublisher,
							Publisher: &models.ObservabilityPublisher{
								ID:              models.ID(1),
								VersionID:       models.ID(2),
								VersionModelID:  models.ID(1),
								ModelSchemaSpec: schemaSpec,
								Revision:        1,
								Status:          models.Pending,
							},
							WorkerData: &models.WorkerData{
								Namespace:       "project-1",
								ModelSchemaSpec: schemaSpec,
								Metadata: models.Metadata{
									App:       "model-1-observability-publisher",
									Component: "worker",
									Stream:    "stream",
									Team:      "team",
								},
								ModelName:    "model-1",
								ModelVersion: "2",
								Revision:     1,
								TopicSource:  "caraml-project-1-model-1-2-prediction-log",
							},
						},
					},
				}).Return(nil)
				return producer
			}(),
			observabilityPublisherStorage: func() *storageMock.ObservabilityPublisherStorage {
				mockStorage := &storageMock.ObservabilityPublisherStorage{}
				mockStorage.On("GetByModelID", mock.Anything, models.ID(1)).Return(&models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionID:       models.ID(2),
					VersionModelID:  models.ID(1),
					Status:          models.Running,
					Revision:        1,
					ModelSchemaSpec: schemaSpec,
				}, nil)
				mockStorage.On("Update", mock.Anything, &models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionID:       models.ID(2),
					VersionModelID:  models.ID(1),
					Status:          models.Pending,
					Revision:        1,
					ModelSchemaSpec: schemaSpec,
				}, false).Return(&models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionID:       models.ID(2),
					VersionModelID:  models.ID(1),
					Revision:        1,
					Status:          models.Pending,
					ModelSchemaSpec: schemaSpec,
				}, nil)
				return mockStorage
			}(),
			versionStorage: func() *storageMock.VersionStorage {
				mockStorage := &storageMock.VersionStorage{}
				mockStorage.On("FindByID", mock.Anything, models.ID(2), model.ID).Return(&models.Version{
					ID:          models.ID(2),
					ModelID:     model.ID,
					ModelSchema: modelSchema,
					Model:       model,
				}, nil)
				return mockStorage
			}(),
			model: model,
			modelEndpoint: &models.ModelEndpoint{
				ID:      models.ID(1),
				ModelID: model.ID,
				Model:   model,
				Status:  models.EndpointServing,
				Rule: &models.ModelEndpointRule{
					Destination: []*models.ModelEndpointRuleDestination{
						{
							VersionEndpoint: &models.VersionEndpoint{
								ID:                       uuid.UUID{},
								VersionID:                models.ID(2),
								Status:                   models.EndpointServing,
								EnableModelObservability: false,
							},
						},
					},
				},
			},
		},
		{
			name: "do nothing; version endpoint model observability is disabled and last state of publisher is terminated",
			jobProducer: func() *queueMock.Producer {
				producer := &queueMock.Producer{}
				return producer
			}(),
			observabilityPublisherStorage: func() *storageMock.ObservabilityPublisherStorage {
				mockStorage := &storageMock.ObservabilityPublisherStorage{}
				mockStorage.On("GetByModelID", mock.Anything, models.ID(1)).Return(&models.ObservabilityPublisher{
					ID:             models.ID(1),
					VersionID:      models.ID(2),
					VersionModelID: model.ID,
					Status:         models.Terminated,
					Revision:       1,
				}, nil)
				return mockStorage
			}(),
			versionStorage: func() *storageMock.VersionStorage {
				mockStorage := &storageMock.VersionStorage{}
				return mockStorage
			}(),
			model: model,
			modelEndpoint: &models.ModelEndpoint{
				ID:      models.ID(1),
				ModelID: model.ID,
				Model:   model,
				Status:  models.EndpointServing,
				Rule: &models.ModelEndpointRule{
					Destination: []*models.ModelEndpointRuleDestination{
						{
							VersionEndpoint: &models.VersionEndpoint{
								ID:                       uuid.UUID{},
								VersionID:                models.ID(3),
								Status:                   models.EndpointServing,
								EnableModelObservability: false,
							},
						},
					},
				},
			},
		},
		{
			name: "undeployment request failed; fail fetch version",
			jobProducer: func() *queueMock.Producer {
				producer := &queueMock.Producer{}
				return producer
			}(),
			observabilityPublisherStorage: func() *storageMock.ObservabilityPublisherStorage {
				mockStorage := &storageMock.ObservabilityPublisherStorage{}
				mockStorage.On("GetByModelID", mock.Anything, models.ID(1)).Return(&models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionID:       models.ID(2),
					VersionModelID:  models.ID(1),
					Status:          models.Running,
					Revision:        1,
					ModelSchemaSpec: schemaSpec,
				}, nil)
				return mockStorage
			}(),
			versionStorage: func() *storageMock.VersionStorage {
				mockStorage := &storageMock.VersionStorage{}
				mockStorage.On("FindByID", mock.Anything, models.ID(2), model.ID).Return(nil, fmt.Errorf("connection error"))
				return mockStorage
			}(),
			model: model,
			modelEndpoint: &models.ModelEndpoint{
				ID:      models.ID(1),
				ModelID: model.ID,
				Model:   model,
				Status:  models.EndpointServing,
				Rule: &models.ModelEndpointRule{
					Destination: []*models.ModelEndpointRuleDestination{
						{
							VersionEndpoint: &models.VersionEndpoint{
								ID:                       uuid.UUID{},
								VersionID:                models.ID(2),
								Status:                   models.EndpointServing,
								EnableModelObservability: false,
							},
						},
					},
				},
			},
			expectedError: fmt.Errorf("connection error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventProducer := NewEventProducer(tt.jobProducer, tt.observabilityPublisherStorage, tt.versionStorage)
			err := eventProducer.ModelEndpointChangeEvent(tt.modelEndpoint, tt.model)
			assert.Equal(t, tt.expectedError, err)
			tt.observabilityPublisherStorage.AssertExpectations(t)
			tt.versionStorage.AssertExpectations(t)
			tt.jobProducer.AssertExpectations(t)
		})
	}
}

func Test_eventProducer_VersionEndpointChangeEvent(t *testing.T) {
	model := &models.Model{
		ID:   models.ID(1),
		Name: "model-1",
		Project: mlp.Project{
			Name:   "project-1",
			Stream: "stream",
			Team:   "team",
		},
		ObservabilitySupported: true,
	}
	schemaSpec := &models.SchemaSpec{
		TagColumns: []string{"tag"},
		FeatureTypes: map[string]models.ValueType{
			"featureA": models.Float64,
			"featureB": models.Float64,
			"featureC": models.Int64,
			"featureD": models.Boolean,
		},
		ModelPredictionOutput: &models.ModelPredictionOutput{
			BinaryClassificationOutput: &models.BinaryClassificationOutput{
				NegativeClassLabel:    "negative",
				PositiveClassLabel:    "positive",
				PredictionLabelColumn: "prediction_label",
				PredictionScoreColumn: "prediction_score",
				OutputClass:           models.BinaryClassification,
			},
		},
	}

	modelSchema := &models.ModelSchema{
		ID:      models.ID(1),
		ModelID: model.ID,
		Spec:    schemaSpec,
	}

	regresionSchemaSpec := &models.SchemaSpec{
		TagColumns: []string{"tag"},
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
				OutputClass:           models.Regression,
			},
		},
	}
	tests := []struct {
		name                          string
		jobProducer                   *queueMock.Producer
		observabilityPublisherStorage *storageMock.ObservabilityPublisherStorage
		versionStorage                *storageMock.VersionStorage
		model                         *models.Model
		versionEndpoint               *models.VersionEndpoint
		expectedError                 error
	}{
		{
			name:                          "do nothing if model not supported observability",
			jobProducer:                   &queueMock.Producer{},
			observabilityPublisherStorage: &storageMock.ObservabilityPublisherStorage{},
			versionStorage:                &storageMock.VersionStorage{},
			model: &models.Model{
				ID:                     models.ID(1),
				ObservabilitySupported: false,
			},
		},
		{
			name: "no deployment; version endpoint model observability is disabled and never been deployed before",
			jobProducer: func() *queueMock.Producer {
				producer := &queueMock.Producer{}
				return producer
			}(),
			observabilityPublisherStorage: func() *storageMock.ObservabilityPublisherStorage {
				mockStorage := &storageMock.ObservabilityPublisherStorage{}
				mockStorage.On("GetByModelID", mock.Anything, models.ID(1)).Return(nil, nil)
				return mockStorage
			}(),
			versionStorage: &storageMock.VersionStorage{},
			model:          model,
			versionEndpoint: &models.VersionEndpoint{
				ID:                       uuid.UUID{},
				VersionID:                models.ID(2),
				Status:                   models.EndpointServing,
				EnableModelObservability: false,
			},
		},
		{
			name: "fresh deployment",
			jobProducer: func() *queueMock.Producer {
				producer := &queueMock.Producer{}
				producer.On("EnqueueJob", &queue.Job{
					Name: ObservabilityPublisherDeployment,
					Arguments: queue.Arguments{
						dataArgKey: models.ObservabilityPublisherJob{
							ActionType: models.DeployPublisher,
							Publisher: &models.ObservabilityPublisher{
								ID:              models.ID(1),
								VersionID:       models.ID(2),
								VersionModelID:  models.ID(1),
								ModelSchemaSpec: schemaSpec,
								Revision:        1,
								Status:          models.Pending,
							},
							WorkerData: &models.WorkerData{
								Namespace:       "project-1",
								ModelSchemaSpec: schemaSpec,
								Metadata: models.Metadata{
									App:       "model-1-observability-publisher",
									Component: "worker",
									Stream:    "stream",
									Team:      "team",
								},
								ModelName:    "model-1",
								ModelVersion: "2",
								Revision:     1,
								TopicSource:  "caraml-project-1-model-1-2-prediction-log",
							},
						},
					},
				}).Return(nil)
				return producer
			}(),
			observabilityPublisherStorage: func() *storageMock.ObservabilityPublisherStorage {
				mockStorage := &storageMock.ObservabilityPublisherStorage{}
				mockStorage.On("GetByModelID", mock.Anything, models.ID(1)).Return(nil, nil)
				mockStorage.On("Create", mock.Anything, &models.ObservabilityPublisher{
					VersionID:       models.ID(2),
					VersionModelID:  models.ID(1),
					Status:          models.Pending,
					Revision:        1,
					ModelSchemaSpec: schemaSpec,
				}).Return(&models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionID:       models.ID(2),
					VersionModelID:  models.ID(1),
					Revision:        1,
					Status:          models.Pending,
					ModelSchemaSpec: schemaSpec,
				}, nil)
				return mockStorage
			}(),
			versionStorage: func() *storageMock.VersionStorage {
				mockStorage := &storageMock.VersionStorage{}
				mockStorage.On("FindByID", mock.Anything, models.ID(2), model.ID).Return(&models.Version{
					ID:          models.ID(2),
					ModelID:     model.ID,
					ModelSchema: modelSchema,
					Model:       model,
				}, nil)
				return mockStorage
			}(),
			model: model,
			versionEndpoint: &models.VersionEndpoint{
				ID:                       uuid.UUID{},
				VersionID:                models.ID(2),
				VersionModelID:           model.ID,
				Status:                   models.EndpointServing,
				EnableModelObservability: true,
			},
		},
		{
			name: "fresh deployment request failed - version doesn't have schema ",
			jobProducer: func() *queueMock.Producer {
				producer := &queueMock.Producer{}
				return producer
			}(),
			observabilityPublisherStorage: func() *storageMock.ObservabilityPublisherStorage {
				mockStorage := &storageMock.ObservabilityPublisherStorage{}
				mockStorage.On("GetByModelID", mock.Anything, models.ID(1)).Return(nil, nil)
				return mockStorage
			}(),
			versionStorage: func() *storageMock.VersionStorage {
				mockStorage := &storageMock.VersionStorage{}
				mockStorage.On("FindByID", mock.Anything, models.ID(3), model.ID).Return(&models.Version{
					ID:          models.ID(3),
					ModelID:     model.ID,
					ModelSchema: nil,
					Model:       model,
				}, nil)
				return mockStorage
			}(),
			model: model,
			versionEndpoint: &models.VersionEndpoint{
				ID:                       uuid.UUID{},
				VersionID:                models.ID(3),
				VersionModelID:           model.ID,
				Status:                   models.EndpointServing,
				EnableModelObservability: true,
			},
			expectedError: fmt.Errorf("versionID: 3 in modelID: 1 doesn't have model schema"),
		},
		{
			name: "redeployment - model endpoint change version",
			jobProducer: func() *queueMock.Producer {
				producer := &queueMock.Producer{}
				producer.On("EnqueueJob", &queue.Job{
					Name: ObservabilityPublisherDeployment,
					Arguments: queue.Arguments{
						dataArgKey: models.ObservabilityPublisherJob{
							ActionType: models.DeployPublisher,
							Publisher: &models.ObservabilityPublisher{
								ID:              models.ID(1),
								VersionID:       models.ID(3),
								VersionModelID:  models.ID(1),
								ModelSchemaSpec: regresionSchemaSpec,
								Revision:        2,
								Status:          models.Pending,
							},
							WorkerData: &models.WorkerData{
								Namespace:       "project-1",
								ModelSchemaSpec: regresionSchemaSpec,
								Metadata: models.Metadata{
									App:       "model-1-observability-publisher",
									Component: "worker",
									Stream:    "stream",
									Team:      "team",
								},
								ModelName:    "model-1",
								ModelVersion: "3",
								Revision:     2,
								TopicSource:  "caraml-project-1-model-1-3-prediction-log",
							},
						},
					},
				}).Return(nil)
				return producer
			}(),
			observabilityPublisherStorage: func() *storageMock.ObservabilityPublisherStorage {
				mockStorage := &storageMock.ObservabilityPublisherStorage{}
				mockStorage.On("GetByModelID", mock.Anything, models.ID(1)).Return(&models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionID:       models.ID(2),
					VersionModelID:  models.ID(1),
					Status:          models.Pending,
					ModelSchemaSpec: schemaSpec,
				}, nil)
				mockStorage.On("Update", mock.Anything, &models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionID:       models.ID(3),
					VersionModelID:  models.ID(1),
					Status:          models.Pending,
					ModelSchemaSpec: regresionSchemaSpec,
				}, true).Return(&models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionID:       models.ID(3),
					VersionModelID:  models.ID(1),
					Revision:        2,
					Status:          models.Pending,
					ModelSchemaSpec: regresionSchemaSpec,
				}, nil)
				return mockStorage
			}(),
			versionStorage: func() *storageMock.VersionStorage {
				mockStorage := &storageMock.VersionStorage{}
				mockStorage.On("FindByID", mock.Anything, models.ID(3), model.ID).Return(&models.Version{
					ID:      models.ID(3),
					ModelID: model.ID,
					ModelSchema: &models.ModelSchema{
						ID:      models.ID(1),
						ModelID: model.ID,
						Spec:    regresionSchemaSpec,
					},
					Model: model,
				}, nil)
				return mockStorage
			}(),
			model: model,
			versionEndpoint: &models.VersionEndpoint{
				ID:                       uuid.UUID{},
				VersionID:                models.ID(3),
				VersionModelID:           model.ID,
				Status:                   models.EndpointServing,
				EnableModelObservability: true,
			},
		},
		{
			name: "undeployment - model observability is disabled",
			jobProducer: func() *queueMock.Producer {
				producer := &queueMock.Producer{}
				producer.On("EnqueueJob", &queue.Job{
					Name: ObservabilityPublisherDeployment,
					Arguments: queue.Arguments{
						dataArgKey: models.ObservabilityPublisherJob{
							ActionType: models.UndeployPublisher,
							Publisher: &models.ObservabilityPublisher{
								ID:              models.ID(1),
								VersionID:       models.ID(2),
								VersionModelID:  models.ID(1),
								ModelSchemaSpec: schemaSpec,
								Revision:        1,
								Status:          models.Pending,
							},
							WorkerData: &models.WorkerData{
								Namespace:       "project-1",
								ModelSchemaSpec: schemaSpec,
								Metadata: models.Metadata{
									App:       "model-1-observability-publisher",
									Component: "worker",
									Stream:    "stream",
									Team:      "team",
								},
								ModelName:    "model-1",
								ModelVersion: "2",
								Revision:     1,
								TopicSource:  "caraml-project-1-model-1-2-prediction-log",
							},
						},
					},
				}).Return(nil)
				return producer
			}(),
			observabilityPublisherStorage: func() *storageMock.ObservabilityPublisherStorage {
				mockStorage := &storageMock.ObservabilityPublisherStorage{}
				mockStorage.On("GetByModelID", mock.Anything, models.ID(1)).Return(&models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionID:       models.ID(2),
					VersionModelID:  models.ID(1),
					Status:          models.Running,
					Revision:        1,
					ModelSchemaSpec: schemaSpec,
				}, nil)
				mockStorage.On("Update", mock.Anything, &models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionID:       models.ID(2),
					VersionModelID:  models.ID(1),
					Status:          models.Pending,
					Revision:        1,
					ModelSchemaSpec: schemaSpec,
				}, false).Return(&models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionID:       models.ID(2),
					VersionModelID:  models.ID(1),
					Revision:        1,
					Status:          models.Pending,
					ModelSchemaSpec: schemaSpec,
				}, nil)
				return mockStorage
			}(),
			versionStorage: func() *storageMock.VersionStorage {
				mockStorage := &storageMock.VersionStorage{}
				mockStorage.On("FindByID", mock.Anything, models.ID(2), model.ID).Return(&models.Version{
					ID:          models.ID(2),
					ModelID:     model.ID,
					ModelSchema: modelSchema,
					Model:       model,
				}, nil)
				return mockStorage
			}(),
			model: model,
			versionEndpoint: &models.VersionEndpoint{
				ID:                       uuid.UUID{},
				VersionID:                models.ID(2),
				VersionModelID:           model.ID,
				Status:                   models.EndpointServing,
				EnableModelObservability: false,
			},
		},
		{
			name: "do nothing; version endpoint model observability is disabled and last state of publisher is terminated",
			jobProducer: func() *queueMock.Producer {
				producer := &queueMock.Producer{}
				return producer
			}(),
			observabilityPublisherStorage: func() *storageMock.ObservabilityPublisherStorage {
				mockStorage := &storageMock.ObservabilityPublisherStorage{}
				mockStorage.On("GetByModelID", mock.Anything, models.ID(1)).Return(&models.ObservabilityPublisher{
					ID:             models.ID(1),
					VersionID:      models.ID(2),
					VersionModelID: model.ID,
					Status:         models.Terminated,
					Revision:       1,
				}, nil)
				return mockStorage
			}(),
			versionStorage: func() *storageMock.VersionStorage {
				mockStorage := &storageMock.VersionStorage{}
				return mockStorage
			}(),
			model: model,
			versionEndpoint: &models.VersionEndpoint{
				ID:                       uuid.UUID{},
				VersionID:                models.ID(3),
				VersionModelID:           model.ID,
				Status:                   models.EndpointServing,
				EnableModelObservability: false,
			},
		},
		{
			name: "undeployment request failed; fail fetch version",
			jobProducer: func() *queueMock.Producer {
				producer := &queueMock.Producer{}
				return producer
			}(),
			observabilityPublisherStorage: func() *storageMock.ObservabilityPublisherStorage {
				mockStorage := &storageMock.ObservabilityPublisherStorage{}
				mockStorage.On("GetByModelID", mock.Anything, models.ID(1)).Return(&models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionID:       models.ID(2),
					VersionModelID:  models.ID(1),
					Status:          models.Running,
					Revision:        1,
					ModelSchemaSpec: schemaSpec,
				}, nil)
				return mockStorage
			}(),
			versionStorage: func() *storageMock.VersionStorage {
				mockStorage := &storageMock.VersionStorage{}
				mockStorage.On("FindByID", mock.Anything, models.ID(2), model.ID).Return(nil, fmt.Errorf("connection error"))
				return mockStorage
			}(),
			model: model,
			versionEndpoint: &models.VersionEndpoint{
				ID:                       uuid.UUID{},
				VersionID:                models.ID(2),
				VersionModelID:           model.ID,
				Status:                   models.EndpointServing,
				EnableModelObservability: false,
			},
			expectedError: fmt.Errorf("connection error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventProducer := NewEventProducer(tt.jobProducer, tt.observabilityPublisherStorage, tt.versionStorage)
			err := eventProducer.VersionEndpointChangeEvent(tt.versionEndpoint, tt.model)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
