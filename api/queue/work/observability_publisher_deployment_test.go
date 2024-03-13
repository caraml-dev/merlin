package work

import (
	"fmt"
	"testing"

	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/pkg/observability/deployment"
	deploymentMock "github.com/caraml-dev/merlin/pkg/observability/deployment/mocks"
	"github.com/caraml-dev/merlin/pkg/observability/event"
	"github.com/caraml-dev/merlin/queue"
	storageMock "github.com/caraml-dev/merlin/storage/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDeploy(t *testing.T) {
	schemaSpec := &models.SchemaSpec{
		SessionIDColumn: "session_id",
		RowIDColumn:     "row_id",
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
	}
	testCases := []struct {
		desc                          string
		deployer                      *deploymentMock.Deployer
		observabilityPublisherStorage *storageMock.ObservabilityPublisherStorage
		jobData                       *queue.Job
		expectedError                 error
	}{
		{
			desc: "deployment completed; fresh deployment",
			deployer: func() *deploymentMock.Deployer {
				mockDeployer := &deploymentMock.Deployer{}
				mockDeployer.On("GetDeployedManifest", mock.Anything, &models.WorkerData{
					Project:         "project-1",
					ModelSchemaSpec: schemaSpec,
					ModelName:       "model-1",
					ModelVersion:    "1",
					Revision:        1,
					TopicSource:     "caraml-project-1-model-1-1-prediction-log",
				}).Return(&deployment.Manifest{}, nil)
				mockDeployer.On("Deploy", mock.Anything, &models.WorkerData{
					Project:         "project-1",
					ModelSchemaSpec: schemaSpec,
					ModelName:       "model-1",
					ModelVersion:    "1",
					Revision:        1,
					TopicSource:     "caraml-project-1-model-1-1-prediction-log",
				}).Return(nil)
				return mockDeployer
			}(),
			observabilityPublisherStorage: func() *storageMock.ObservabilityPublisherStorage {
				mockStorage := &storageMock.ObservabilityPublisherStorage{}
				mockStorage.On("Get", mock.Anything, model.ID).Return(&models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionModelID:  model.ID,
					VersionID:       models.ID(1),
					Status:          models.Pending,
					Revision:        1,
					ModelSchemaSpec: schemaSpec,
				}, nil)
				mockStorage.On("Update", mock.Anything, &models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionModelID:  models.ID(1),
					VersionID:       model.ID,
					Status:          models.Running,
					Revision:        1,
					ModelSchemaSpec: schemaSpec,
				}, false).Return(&models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionModelID:  model.ID,
					VersionID:       models.ID(1),
					Status:          models.Running,
					Revision:        1,
					ModelSchemaSpec: schemaSpec,
				}, nil)
				return mockStorage
			}(),
			jobData: &queue.Job{
				Name: event.ObservabilityPublisherDeployment,
				Arguments: queue.Arguments{
					dataArgKey: models.ObservabilityPublisherJob{
						ActionType: models.DeployPublisher,
						Publisher: &models.ObservabilityPublisher{
							ID:              models.ID(1),
							VersionModelID:  model.ID,
							VersionID:       models.ID(1),
							Status:          models.Pending,
							Revision:        1,
							ModelSchemaSpec: schemaSpec,
						},
						WorkerData: &models.WorkerData{
							Project:         "project-1",
							ModelSchemaSpec: schemaSpec,
							ModelName:       "model-1",
							ModelVersion:    "1",
							Revision:        1,
							TopicSource:     "caraml-project-1-model-1-1-prediction-log",
						},
					},
				},
			},
		},
		{
			desc: "deployment failed; fresh deployment",
			deployer: func() *deploymentMock.Deployer {
				mockDeployer := &deploymentMock.Deployer{}
				mockDeployer.On("GetDeployedManifest", mock.Anything, &models.WorkerData{
					Project:         "project-1",
					ModelSchemaSpec: schemaSpec,
					ModelName:       "model-1",
					ModelVersion:    "1",
					Revision:        1,
					TopicSource:     "caraml-project-1-model-1-1-prediction-log",
				}).Return(&deployment.Manifest{}, nil)
				mockDeployer.On("Deploy", mock.Anything, &models.WorkerData{
					Project:         "project-1",
					ModelSchemaSpec: schemaSpec,
					ModelName:       "model-1",
					ModelVersion:    "1",
					Revision:        1,
					TopicSource:     "caraml-project-1-model-1-1-prediction-log",
				}).Return(fmt.Errorf("control plane is down"))
				return mockDeployer
			}(),
			observabilityPublisherStorage: func() *storageMock.ObservabilityPublisherStorage {
				mockStorage := &storageMock.ObservabilityPublisherStorage{}
				mockStorage.On("Get", mock.Anything, model.ID).Return(&models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionModelID:  model.ID,
					VersionID:       models.ID(1),
					Status:          models.Pending,
					Revision:        1,
					ModelSchemaSpec: schemaSpec,
				}, nil)
				mockStorage.On("Update", mock.Anything, &models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionModelID:  models.ID(1),
					VersionID:       model.ID,
					Status:          models.Failed,
					Revision:        1,
					ModelSchemaSpec: schemaSpec,
				}, false).Return(&models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionModelID:  model.ID,
					VersionID:       models.ID(1),
					Status:          models.Failed,
					Revision:        1,
					ModelSchemaSpec: schemaSpec,
				}, nil)
				return mockStorage
			}(),
			jobData: &queue.Job{
				Name: event.ObservabilityPublisherDeployment,
				Arguments: queue.Arguments{
					dataArgKey: models.ObservabilityPublisherJob{
						ActionType: models.DeployPublisher,
						Publisher: &models.ObservabilityPublisher{
							ID:              models.ID(1),
							VersionModelID:  model.ID,
							VersionID:       models.ID(1),
							Status:          models.Pending,
							Revision:        1,
							ModelSchemaSpec: schemaSpec,
						},
						WorkerData: &models.WorkerData{
							Project:         "project-1",
							ModelSchemaSpec: schemaSpec,
							ModelName:       "model-1",
							ModelVersion:    "1",
							Revision:        1,
							TopicSource:     "caraml-project-1-model-1-1-prediction-log",
						},
					},
				},
			},
			expectedError: fmt.Errorf("control plane is down"),
		},
		{
			desc:     "deployment requeue due to fail fetch from db",
			deployer: &deploymentMock.Deployer{},
			observabilityPublisherStorage: func() *storageMock.ObservabilityPublisherStorage {
				mockStorage := &storageMock.ObservabilityPublisherStorage{}
				mockStorage.On("Get", mock.Anything, model.ID).Return(nil, fmt.Errorf("database is down"))
				return mockStorage
			}(),
			jobData: &queue.Job{
				Name: event.ObservabilityPublisherDeployment,
				Arguments: queue.Arguments{
					dataArgKey: models.ObservabilityPublisherJob{
						ActionType: models.DeployPublisher,
						Publisher: &models.ObservabilityPublisher{
							ID:              models.ID(1),
							VersionModelID:  model.ID,
							VersionID:       models.ID(1),
							Status:          models.Pending,
							Revision:        1,
							ModelSchemaSpec: schemaSpec,
						},
						WorkerData: &models.WorkerData{
							Project:         "project-1",
							ModelSchemaSpec: schemaSpec,
							ModelName:       "model-1",
							ModelVersion:    "1",
							Revision:        1,
							TopicSource:     "caraml-project-1-model-1-1-prediction-log",
						},
					},
				},
			},
			expectedError: queue.RetryableError{Message: "database is down"},
		},
		{
			desc: "undeployment completed",
			deployer: func() *deploymentMock.Deployer {
				mockDeployer := &deploymentMock.Deployer{}
				mockDeployer.On("GetDeployedManifest", mock.Anything, &models.WorkerData{
					Project:         "project-1",
					ModelSchemaSpec: schemaSpec,
					ModelName:       "model-1",
					ModelVersion:    "1",
					Revision:        1,
					TopicSource:     "caraml-project-1-model-1-1-prediction-log",
				}).Return(&deployment.Manifest{}, nil)
				mockDeployer.On("Undeploy", mock.Anything, &models.WorkerData{
					Project:         "project-1",
					ModelSchemaSpec: schemaSpec,
					ModelName:       "model-1",
					ModelVersion:    "1",
					Revision:        1,
					TopicSource:     "caraml-project-1-model-1-1-prediction-log",
				}).Return(nil)
				return mockDeployer
			}(),
			observabilityPublisherStorage: func() *storageMock.ObservabilityPublisherStorage {
				mockStorage := &storageMock.ObservabilityPublisherStorage{}
				mockStorage.On("Get", mock.Anything, model.ID).Return(&models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionModelID:  model.ID,
					VersionID:       models.ID(1),
					Status:          models.Running,
					Revision:        1,
					ModelSchemaSpec: schemaSpec,
				}, nil)
				mockStorage.On("Update", mock.Anything, &models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionModelID:  models.ID(1),
					VersionID:       model.ID,
					Status:          models.Terminated,
					Revision:        1,
					ModelSchemaSpec: schemaSpec,
				}, false).Return(&models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionModelID:  model.ID,
					VersionID:       models.ID(1),
					Status:          models.Terminated,
					Revision:        1,
					ModelSchemaSpec: schemaSpec,
				}, nil)
				return mockStorage
			}(),
			jobData: &queue.Job{
				Name: event.ObservabilityPublisherDeployment,
				Arguments: queue.Arguments{
					dataArgKey: models.ObservabilityPublisherJob{
						ActionType: models.UndeployPublisher,
						Publisher: &models.ObservabilityPublisher{
							ID:              models.ID(1),
							VersionModelID:  model.ID,
							VersionID:       models.ID(1),
							Status:          models.Pending,
							Revision:        1,
							ModelSchemaSpec: schemaSpec,
						},
						WorkerData: &models.WorkerData{
							Project:         "project-1",
							ModelSchemaSpec: schemaSpec,
							ModelName:       "model-1",
							ModelVersion:    "1",
							Revision:        1,
							TopicSource:     "caraml-project-1-model-1-1-prediction-log",
						},
					},
				},
			},
		},
		{
			desc: "undeployment failed; error during deployment to control plane",
			deployer: func() *deploymentMock.Deployer {
				mockDeployer := &deploymentMock.Deployer{}
				mockDeployer.On("GetDeployedManifest", mock.Anything, &models.WorkerData{
					Project:         "project-1",
					ModelSchemaSpec: schemaSpec,
					ModelName:       "model-1",
					ModelVersion:    "1",
					Revision:        1,
					TopicSource:     "caraml-project-1-model-1-1-prediction-log",
				}).Return(&deployment.Manifest{}, nil)
				mockDeployer.On("Undeploy", mock.Anything, &models.WorkerData{
					Project:         "project-1",
					ModelSchemaSpec: schemaSpec,
					ModelName:       "model-1",
					ModelVersion:    "1",
					Revision:        1,
					TopicSource:     "caraml-project-1-model-1-1-prediction-log",
				}).Return(fmt.Errorf("control plane is down"))
				return mockDeployer
			}(),
			observabilityPublisherStorage: func() *storageMock.ObservabilityPublisherStorage {
				mockStorage := &storageMock.ObservabilityPublisherStorage{}
				mockStorage.On("Get", mock.Anything, model.ID).Return(&models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionModelID:  model.ID,
					VersionID:       models.ID(1),
					Status:          models.Pending,
					Revision:        1,
					ModelSchemaSpec: schemaSpec,
				}, nil)
				mockStorage.On("Update", mock.Anything, &models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionModelID:  models.ID(1),
					VersionID:       model.ID,
					Status:          models.Failed,
					Revision:        1,
					ModelSchemaSpec: schemaSpec,
				}, false).Return(&models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionModelID:  model.ID,
					VersionID:       models.ID(1),
					Status:          models.Failed,
					Revision:        1,
					ModelSchemaSpec: schemaSpec,
				}, nil)
				return mockStorage
			}(),
			jobData: &queue.Job{
				Name: event.ObservabilityPublisherDeployment,
				Arguments: queue.Arguments{
					dataArgKey: models.ObservabilityPublisherJob{
						ActionType: models.UndeployPublisher,
						Publisher: &models.ObservabilityPublisher{
							ID:              models.ID(1),
							VersionModelID:  model.ID,
							VersionID:       models.ID(1),
							Status:          models.Pending,
							Revision:        1,
							ModelSchemaSpec: schemaSpec,
						},
						WorkerData: &models.WorkerData{
							Project:         "project-1",
							ModelSchemaSpec: schemaSpec,
							ModelName:       "model-1",
							ModelVersion:    "1",
							Revision:        1,
							TopicSource:     "caraml-project-1-model-1-1-prediction-log",
						},
					},
				},
			},
			expectedError: fmt.Errorf("control plane is down"),
		},
		{
			desc: "redeployment completed",
			deployer: func() *deploymentMock.Deployer {
				mockDeployer := &deploymentMock.Deployer{}
				mockDeployer.On("GetDeployedManifest", mock.Anything, &models.WorkerData{
					Project:         "project-1",
					ModelSchemaSpec: schemaSpec,
					ModelName:       "model-1",
					ModelVersion:    "1",
					Revision:        2,
					TopicSource:     "caraml-project-1-model-1-1-prediction-log",
				}).Return(&deployment.Manifest{
					Deployment: &v1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "model-1-observability-publisher",
							Namespace: "project-1",
							Annotations: map[string]string{
								deployment.PublisherRevisionAnnotationKey: "1",
							},
						},
						Status: v1.DeploymentStatus{
							UnavailableReplicas: 0,
							AvailableReplicas:   1,
						},
					},
					Secret: &corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name: "model-1-config",
						},
					},
					OnProgress: false,
				}, nil)
				mockDeployer.On("Deploy", mock.Anything, &models.WorkerData{
					Project:         "project-1",
					ModelSchemaSpec: schemaSpec,
					ModelName:       "model-1",
					ModelVersion:    "1",
					Revision:        2,
					TopicSource:     "caraml-project-1-model-1-1-prediction-log",
				}).Return(nil)
				return mockDeployer
			}(),
			observabilityPublisherStorage: func() *storageMock.ObservabilityPublisherStorage {
				mockStorage := &storageMock.ObservabilityPublisherStorage{}
				mockStorage.On("Get", mock.Anything, model.ID).Return(&models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionModelID:  model.ID,
					VersionID:       models.ID(1),
					Status:          models.Pending,
					Revision:        2,
					ModelSchemaSpec: schemaSpec,
				}, nil)
				mockStorage.On("Update", mock.Anything, &models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionModelID:  models.ID(1),
					VersionID:       model.ID,
					Status:          models.Running,
					Revision:        2,
					ModelSchemaSpec: schemaSpec,
				}, false).Return(&models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionModelID:  model.ID,
					VersionID:       models.ID(1),
					Status:          models.Running,
					Revision:        2,
					ModelSchemaSpec: schemaSpec,
				}, nil)
				return mockStorage
			}(),
			jobData: &queue.Job{
				Name: event.ObservabilityPublisherDeployment,
				Arguments: queue.Arguments{
					dataArgKey: models.ObservabilityPublisherJob{
						ActionType: models.DeployPublisher,
						Publisher: &models.ObservabilityPublisher{
							ID:              models.ID(1),
							VersionModelID:  model.ID,
							VersionID:       models.ID(1),
							Status:          models.Pending,
							Revision:        2,
							ModelSchemaSpec: schemaSpec,
						},
						WorkerData: &models.WorkerData{
							Project:         "project-1",
							ModelSchemaSpec: schemaSpec,
							ModelName:       "model-1",
							ModelVersion:    "1",
							Revision:        2,
							TopicSource:     "caraml-project-1-model-1-1-prediction-log",
						},
					},
				},
			},
		},
		{
			desc: "redeployment requeued  due to failed save state",
			deployer: func() *deploymentMock.Deployer {
				mockDeployer := &deploymentMock.Deployer{}
				mockDeployer.On("GetDeployedManifest", mock.Anything, &models.WorkerData{
					Project:         "project-1",
					ModelSchemaSpec: schemaSpec,
					ModelName:       "model-1",
					ModelVersion:    "1",
					Revision:        2,
					TopicSource:     "caraml-project-1-model-1-1-prediction-log",
				}).Return(&deployment.Manifest{
					Deployment: &v1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "model-1-observability-publisher",
							Namespace: "project-1",
							Annotations: map[string]string{
								deployment.PublisherRevisionAnnotationKey: "1",
							},
						},
						Status: v1.DeploymentStatus{
							UnavailableReplicas: 1,
							AvailableReplicas:   1,
						},
					},
					Secret: &corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name: "model-1-config",
						},
					},
					OnProgress: false,
				}, nil)
				mockDeployer.On("Deploy", mock.Anything, &models.WorkerData{
					Project:         "project-1",
					ModelSchemaSpec: schemaSpec,
					ModelName:       "model-1",
					ModelVersion:    "1",
					Revision:        2,
					TopicSource:     "caraml-project-1-model-1-1-prediction-log",
				}).Return(nil)
				return mockDeployer
			}(),
			observabilityPublisherStorage: func() *storageMock.ObservabilityPublisherStorage {
				mockStorage := &storageMock.ObservabilityPublisherStorage{}
				mockStorage.On("Get", mock.Anything, model.ID).Return(&models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionModelID:  model.ID,
					VersionID:       models.ID(1),
					Status:          models.Pending,
					Revision:        2,
					ModelSchemaSpec: schemaSpec,
				}, nil)
				mockStorage.On("Update", mock.Anything, &models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionModelID:  models.ID(1),
					VersionID:       model.ID,
					Status:          models.Running,
					Revision:        2,
					ModelSchemaSpec: schemaSpec,
				}, false).Return(nil, fmt.Errorf("connection is lost"))
				return mockStorage
			}(),
			jobData: &queue.Job{
				Name: event.ObservabilityPublisherDeployment,
				Arguments: queue.Arguments{
					dataArgKey: models.ObservabilityPublisherJob{
						ActionType: models.DeployPublisher,
						Publisher: &models.ObservabilityPublisher{
							ID:              models.ID(1),
							VersionModelID:  model.ID,
							VersionID:       models.ID(1),
							Status:          models.Pending,
							Revision:        2,
							ModelSchemaSpec: schemaSpec,
						},
						WorkerData: &models.WorkerData{
							Project:         "project-1",
							ModelSchemaSpec: schemaSpec,
							ModelName:       "model-1",
							ModelVersion:    "1",
							Revision:        2,
							TopicSource:     "caraml-project-1-model-1-1-prediction-log",
						},
					},
				},
			},
			expectedError: queue.RetryableError{Message: "connection is lost"},
		},
		{
			desc: "deployment request revision somehow greater than from db",
			deployer: func() *deploymentMock.Deployer {
				mockDeployer := &deploymentMock.Deployer{}
				return mockDeployer
			}(),
			observabilityPublisherStorage: func() *storageMock.ObservabilityPublisherStorage {
				mockStorage := &storageMock.ObservabilityPublisherStorage{}
				mockStorage.On("Get", mock.Anything, model.ID).Return(&models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionModelID:  model.ID,
					VersionID:       models.ID(1),
					Status:          models.Pending,
					Revision:        1,
					ModelSchemaSpec: schemaSpec,
				}, nil)
				return mockStorage
			}(),
			jobData: &queue.Job{
				Name: event.ObservabilityPublisherDeployment,
				Arguments: queue.Arguments{
					dataArgKey: models.ObservabilityPublisherJob{
						ActionType: models.DeployPublisher,
						Publisher: &models.ObservabilityPublisher{
							ID:              models.ID(1),
							VersionModelID:  model.ID,
							VersionID:       models.ID(1),
							Status:          models.Pending,
							Revision:        2,
							ModelSchemaSpec: schemaSpec,
						},
						WorkerData: &models.WorkerData{
							Project:         "project-1",
							ModelSchemaSpec: schemaSpec,
							ModelName:       "model-1",
							ModelVersion:    "1",
							Revision:        2,
							TopicSource:     "caraml-project-1-model-1-1-prediction-log",
						},
					},
				},
			},
			expectedError: fmt.Errorf("actual publisher revision should not be lower than the one from submitted job"),
		},
		{
			desc: "deployment request already stale compare from db, hence skipped",
			deployer: func() *deploymentMock.Deployer {
				mockDeployer := &deploymentMock.Deployer{}
				return mockDeployer
			}(),
			observabilityPublisherStorage: func() *storageMock.ObservabilityPublisherStorage {
				mockStorage := &storageMock.ObservabilityPublisherStorage{}
				mockStorage.On("Get", mock.Anything, model.ID).Return(&models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionModelID:  model.ID,
					VersionID:       models.ID(1),
					Status:          models.Running,
					Revision:        2,
					ModelSchemaSpec: schemaSpec,
				}, nil)
				return mockStorage
			}(),
			jobData: &queue.Job{
				Name: event.ObservabilityPublisherDeployment,
				Arguments: queue.Arguments{
					dataArgKey: models.ObservabilityPublisherJob{
						ActionType: models.DeployPublisher,
						Publisher: &models.ObservabilityPublisher{
							ID:              models.ID(1),
							VersionModelID:  model.ID,
							VersionID:       models.ID(1),
							Status:          models.Pending,
							Revision:        1,
							ModelSchemaSpec: schemaSpec,
						},
						WorkerData: &models.WorkerData{
							Project:         "project-1",
							ModelSchemaSpec: schemaSpec,
							ModelName:       "model-1",
							ModelVersion:    "1",
							Revision:        1,
							TopicSource:     "caraml-project-1-model-1-1-prediction-log",
						},
					},
				},
			},
		},
		{
			desc: "deployment request is requeue due to another deployment from previous revision is ongoing",
			deployer: func() *deploymentMock.Deployer {
				mockDeployer := &deploymentMock.Deployer{}
				mockDeployer.On("GetDeployedManifest", mock.Anything, &models.WorkerData{
					Project:         "project-1",
					ModelSchemaSpec: schemaSpec,
					ModelName:       "model-1",
					ModelVersion:    "1",
					Revision:        2,
					TopicSource:     "caraml-project-1-model-1-1-prediction-log",
				}).Return(&deployment.Manifest{
					Deployment: &v1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "model-1-observability-publisher",
							Namespace: "project-1",
							Annotations: map[string]string{
								deployment.PublisherRevisionAnnotationKey: "1",
							},
						},
						Status: v1.DeploymentStatus{
							UnavailableReplicas: 1,
							AvailableReplicas:   1,
						},
					},
					Secret: &corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name: "model-1-config",
						},
					},
					OnProgress: true,
				}, nil)
				return mockDeployer
			}(),
			observabilityPublisherStorage: func() *storageMock.ObservabilityPublisherStorage {
				mockStorage := &storageMock.ObservabilityPublisherStorage{}
				mockStorage.On("Get", mock.Anything, model.ID).Return(&models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionModelID:  model.ID,
					VersionID:       models.ID(1),
					Status:          models.Pending,
					Revision:        2,
					ModelSchemaSpec: schemaSpec,
				}, nil)
				return mockStorage
			}(),
			jobData: &queue.Job{
				Name: event.ObservabilityPublisherDeployment,
				Arguments: queue.Arguments{
					dataArgKey: models.ObservabilityPublisherJob{
						ActionType: models.DeployPublisher,
						Publisher: &models.ObservabilityPublisher{
							ID:              models.ID(1),
							VersionModelID:  model.ID,
							VersionID:       models.ID(1),
							Status:          models.Pending,
							Revision:        2,
							ModelSchemaSpec: schemaSpec,
						},
						WorkerData: &models.WorkerData{
							Project:         "project-1",
							ModelSchemaSpec: schemaSpec,
							ModelName:       "model-1",
							ModelVersion:    "1",
							Revision:        2,
							TopicSource:     "caraml-project-1-model-1-1-prediction-log",
						},
					},
				},
			},
			expectedError: queue.RetryableError{Message: "there is on going deployment for previous revision, this deployment request must wait until previous deployment success"},
		},
		{
			desc: "deployment request is failed due to another deployment from greater revision is ongoing",
			deployer: func() *deploymentMock.Deployer {
				mockDeployer := &deploymentMock.Deployer{}
				mockDeployer.On("GetDeployedManifest", mock.Anything, &models.WorkerData{
					Project:         "project-1",
					ModelSchemaSpec: schemaSpec,
					ModelName:       "model-1",
					ModelVersion:    "1",
					Revision:        2,
					TopicSource:     "caraml-project-1-model-1-1-prediction-log",
				}).Return(&deployment.Manifest{
					Deployment: &v1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "model-1-observability-publisher",
							Namespace: "project-1",
							Annotations: map[string]string{
								deployment.PublisherRevisionAnnotationKey: "3",
							},
						},
						Status: v1.DeploymentStatus{
							UnavailableReplicas: 1,
							AvailableReplicas:   1,
						},
					},
					Secret: &corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name: "model-1-config",
						},
					},
					OnProgress: true,
				}, nil)
				return mockDeployer
			}(),
			observabilityPublisherStorage: func() *storageMock.ObservabilityPublisherStorage {
				mockStorage := &storageMock.ObservabilityPublisherStorage{}
				mockStorage.On("Get", mock.Anything, model.ID).Return(&models.ObservabilityPublisher{
					ID:              models.ID(1),
					VersionModelID:  model.ID,
					VersionID:       models.ID(1),
					Status:          models.Pending,
					Revision:        2,
					ModelSchemaSpec: schemaSpec,
				}, nil)
				return mockStorage
			}(),
			jobData: &queue.Job{
				Name: event.ObservabilityPublisherDeployment,
				Arguments: queue.Arguments{
					dataArgKey: models.ObservabilityPublisherJob{
						ActionType: models.DeployPublisher,
						Publisher: &models.ObservabilityPublisher{
							ID:              models.ID(1),
							VersionModelID:  model.ID,
							VersionID:       models.ID(1),
							Status:          models.Pending,
							Revision:        2,
							ModelSchemaSpec: schemaSpec,
						},
						WorkerData: &models.WorkerData{
							Project:         "project-1",
							ModelSchemaSpec: schemaSpec,
							ModelName:       "model-1",
							ModelVersion:    "1",
							Revision:        2,
							TopicSource:     "caraml-project-1-model-1-1-prediction-log",
						},
					},
				},
			},
			expectedError: fmt.Errorf("latest deployment already being deployed"),
		},
		{
			desc:                          "deployment fail due to incorrect job data",
			deployer:                      &deploymentMock.Deployer{},
			observabilityPublisherStorage: &storageMock.ObservabilityPublisherStorage{},
			jobData: &queue.Job{
				ID:   1,
				Name: event.ObservabilityPublisherDeployment,
				Arguments: queue.Arguments{
					dataArgKey: "randomString",
				},
			},
			expectedError: fmt.Errorf("job data for ID: 1 is not in ObservabilityPublisherJob type"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			depl := &ObservabilityPublisherDeployment{
				Deployer:                      tC.deployer,
				ObservabilityPublisherStorage: tC.observabilityPublisherStorage,
			}
			err := depl.Deploy(tC.jobData)
			assert.Equal(t, tC.expectedError, err)
			tC.deployer.AssertExpectations(t)
			tC.observabilityPublisherStorage.AssertExpectations(t)

		})
	}
}
