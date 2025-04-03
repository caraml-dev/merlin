package deployment

import (
	"context"
	"fmt"
	"time"

	"net/http"
	"reflect"
	"strconv"
	"testing"

	"github.com/caraml-dev/merlin/config"
	"github.com/caraml-dev/merlin/models"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"

	fakeappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1/fake"
	fakecorev1 "k8s.io/client-go/kubernetes/typed/core/v1/fake"

	ktesting "k8s.io/client-go/testing"
)

const (
	getMethod    = "get"
	createMethod = "create"
	updateMethod = "update"
	deleteMethod = "delete"

	secretResource     = "secrets"
	deploymentResource = "deployments"
)

type deploymentStatus string

const (
	noStatus     deploymentStatus = "no_status"
	onProgress   deploymentStatus = "on_progress"
	ready        deploymentStatus = "ready"
	timeoutError deploymentStatus = "timeout_error"

	namespace                = "caraml-observability"
	serviceAccountSecretName = "caraml-observability-sa-secret"
)

func createDeploymentSpec(data *models.WorkerData, resourceRequest corev1.ResourceList, resourceLimit corev1.ResourceList, imageName string, serviceAccountSecretName string) *appsv1.Deployment {
	labels := data.Metadata.ToLabel()
	labels[appLabelKey] = data.Metadata.App
	numReplicas := int32(2)
	cfgVolName := "config-volume"
	depl := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s-mlobs", data.Project, data.ModelName),
			Namespace: namespace,
			Labels:    labels,
			Annotations: map[string]string{
				PublisherRevisionAnnotationKey: strconv.Itoa(data.Revision),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": data.Metadata.App,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
					Annotations: map[string]string{
						PublisherRevisionAnnotationKey: strconv.Itoa(data.Revision),
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "worker",
							Image: imageName,
							Command: []string{
								"python",
								"-m",
								"publisher",
								"+environment=config",
							},
							ImagePullPolicy: corev1.PullIfNotPresent,

							Resources: corev1.ResourceRequirements{
								Requests: resourceRequest,
								Limits:   resourceLimit,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      cfgVolName,
									MountPath: "/mlobs/observation-publisher/conf/environment",
									ReadOnly:  true,
								},
								{
									Name:      "iam-secret",
									MountPath: fmt.Sprintf("/iam/%s", serviceAccountSecretName),
									ReadOnly:  true,
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "prom-metric",
									ContainerPort: 8000,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "GOOGLE_APPLICATION_CREDENTIALS",
									Value: fmt.Sprintf("/iam/%s/service-account.json", serviceAccountSecretName),
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: cfgVolName,
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: fmt.Sprintf("%s-%s-config", data.Project, data.ModelName),
								},
							},
						},
						{
							Name: "iam-secret",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: serviceAccountSecretName,
								},
							},
						},
					},
				},
			},
			Replicas: &numReplicas,
		},
	}

	return depl
}

func changeDeploymentStatus(depl *appsv1.Deployment, status deploymentStatus, revision int) *appsv1.Deployment {
	updatedDepl := depl.DeepCopy()
	numReplicas := int32(2)
	var detailStatus appsv1.DeploymentStatus
	if status == onProgress {
		detailStatus = appsv1.DeploymentStatus{
			Replicas:            numReplicas + 1,
			UnavailableReplicas: numReplicas,
			UpdatedReplicas:     1,
		}
	} else if status == timeoutError {
		detailStatus = appsv1.DeploymentStatus{
			Conditions: []appsv1.DeploymentCondition{
				{
					Type:   appsv1.DeploymentProgressing,
					Reason: timeoutReason,
				},
			},
			Replicas:        numReplicas,
			UpdatedReplicas: 1,
		}
	} else {
		detailStatus = appsv1.DeploymentStatus{
			Replicas:          numReplicas,
			UpdatedReplicas:   numReplicas,
			AvailableReplicas: numReplicas,
		}
	}
	updatedDepl.Status = detailStatus
	updatedDepl.Annotations[k8sRevisionAnnotation] = strconv.Itoa(revision)
	return updatedDepl
}

type deploymentWatchReactor struct {
	result chan watch.Event
}

func newDeploymentWatchReactor(depl *appsv1.Deployment) *deploymentWatchReactor {
	w := &deploymentWatchReactor{result: make(chan watch.Event, 1)}
	w.result <- watch.Event{Type: watch.Added, Object: depl}
	return w
}

func (w *deploymentWatchReactor) Handles(action ktesting.Action) bool {
	return action.GetResource().Resource == deploymentResource
}

func (w *deploymentWatchReactor) React(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
	return true, watch.NewProxyWatcher(w.result), nil
}

func Test_deployer_Deploy(t *testing.T) {
	consumerConfig := config.ObservabilityPublisher{
		ArizeSink: config.ArizeSink{
			APIKey:              "api-key",
			SpaceKey:            "space-key",
			EnabledModelSerials: "project-1_model-1",
		},
		BigQuerySink: config.BigQuerySink{
			Project: "bq-project",
			Dataset: "dataset",
			TTLDays: 10,
			Enabled: true,
		},
		MaxComputeSink: config.MaxComputeSink{
			Project:         "max-project",
			Dataset:         "dataset",
			TTLDays:         0,
			AccessKeyID:     "key",
			AccessKeySecret: "secret",
			AccessUrl:       "url",
			Enabled:         true,
		},
		KafkaConsumer: config.KafkaConsumer{
			Brokers:   "broker-1",
			GroupID:   "group-id",
			BatchSize: 100,
			AdditionalConsumerConfig: map[string]string{
				"auto.offset.reset": "latest",
				"fetch.min.bytes":   "1024000",
			},
		},
		ImageName: "observability-publisher:v0.0",
		DefaultResources: config.ResourceRequestsLimits{
			Requests: config.Resource{
				CPU:    "1",
				Memory: "1Gi",
			},
			Limits: config.Resource{
				Memory: "1Gi",
			},
		},
		EnvironmentName:          "dev",
		Replicas:                 2,
		TargetNamespace:          namespace,
		ServiceAccountSecretName: serviceAccountSecretName,
		DeploymentTimeout:        5 * time.Second,
	}
	requestResource := corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse("1"),
		corev1.ResourceMemory: resource.MustParse("1Gi"),
	}
	limitResource := corev1.ResourceList{
		corev1.ResourceMemory: resource.MustParse("1Gi"),
	}
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
				NegativeClassLabel:    "negative",
				PositiveClassLabel:    "positive",
				PredictionLabelColumn: "prediction_label",
				PredictionScoreColumn: "prediction_score",
				OutputClass:           models.BinaryClassification,
			},
		},
	}
	tests := []struct {
		name            string
		data            *models.WorkerData
		kubeClient      kubernetes.Interface
		consumerConfig  config.ObservabilityPublisher
		resourceRequest corev1.ResourceList
		resourceLimit   corev1.ResourceList
		expectedErr     error
	}{
		{
			name: "fresh deployment",
			data: &models.WorkerData{
				Project:         "project-1",
				ModelSchemaSpec: schemaSpec,
				ModelName:       "model-1",
				ModelVersion:    "1",
				Revision:        1,
				Metadata: models.Metadata{
					App:       "model-1-observability-publisher",
					Component: "worker",
					Stream:    "stream",
					Team:      "team",
				},
				TopicSource: "caraml-project-1-model-1-1-prediction-log",
			},
			consumerConfig: consumerConfig,

			resourceRequest: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1"),
				corev1.ResourceMemory: resource.MustParse("1Gi"),
			},
			resourceLimit: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse("1Gi"),
			},

			kubeClient: func() kubernetes.Interface {
				clientSet := fake.NewSimpleClientset()
				secretAPI := clientSet.CoreV1().Secrets(namespace).(*fakecorev1.FakeSecrets)
				prependGetSecretReactor(t, secretAPI, nil, nil)
				prependUpsertSecretReactor(t, secretAPI, []*corev1.Secret{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "project-1-model-1-config",
							Namespace: namespace,
							Labels: map[string]string{
								"app":          "model-1-observability-publisher",
								"component":    "worker",
								"orchestrator": "merlin",
								"stream":       "stream",
								"team":         "team",
								"environment":  "",
							},
						},
						StringData: map[string]string{
							"config.yaml": "project: project-1\nmodel_id: model-1\nmodel_version: \"1\"\ninference_schema:\n  session_id_column: session_id\n  row_id_column: row_id\n  model_prediction_output:\n    actual_score_column: \"\"\n    negative_class_label: negative\n    prediction_score_column: prediction_score\n    prediction_label_column: prediction_label\n    positive_class_label: positive\n    score_threshold: null\n    output_class: BinaryClassificationOutput\n  tag_columns:\n  - tag\n  feature_types:\n    featureA: float64\n    featureB: float64\n    featureC: int64\n    featureD: boolean\n  feature_orders: []\nobservation_sinks:\n- type: BIGQUERY\n  config:\n    project: bq-project\n    dataset: dataset\n    ttl_days: 10\n- type: MAXCOMPUTE\n  config:\n    project: max-project\n    dataset: dataset\n    ttl_days: 0\n    access_key_id: key\n    access_key_secret: secret\n    access_url: url\n- type: ARIZE\n  config:\n    api_key: api-key\n    space_key: space-key\nobservation_source:\n  type: KAFKA\n  config:\n    topic: caraml-project-1-model-1-1-prediction-log\n    bootstrap_servers: broker-1\n    group_id: group-id\n    batch_size: 100\n    additional_consumer_config:\n      auto.offset.reset: latest\n      fetch.min.bytes: \"1024000\"\n",
						},
					}}, nil, false)
				deploymentAPI := clientSet.AppsV1().Deployments(namespace).(*fakeappsv1.FakeDeployments)
				preprendGetDeploymentReactor(t, deploymentAPI, nil, nil)
				depl := createDeploymentSpec(&models.WorkerData{
					Project:         "project-1",
					ModelSchemaSpec: schemaSpec,
					ModelName:       "model-1",
					ModelVersion:    "1",
					Revision:        1,
					Metadata: models.Metadata{
						App:       "model-1-observability-publisher",
						Component: "worker",
						Stream:    "stream",
						Team:      "team",
					},
				}, requestResource, limitResource, consumerConfig.ImageName, consumerConfig.ServiceAccountSecretName)
				prependUpsertDeploymentReactor(t, deploymentAPI, depl, nil, false)

				updatedDepl := changeDeploymentStatus(depl, ready, 1)
				deplWatchReactor := newDeploymentWatchReactor(updatedDepl)
				clientSet.WatchReactionChain = []ktesting.WatchReactor{deplWatchReactor}

				return clientSet
			}(),
		},
		{
			name: "fresh deployment failed; failed create secret",
			data: &models.WorkerData{
				Project:         "project-1",
				ModelSchemaSpec: schemaSpec,
				ModelName:       "model-1",
				ModelVersion:    "1",
				Revision:        1,
				Metadata: models.Metadata{
					App:       "model-1-observability-publisher",
					Component: "worker",
					Stream:    "stream",
					Team:      "team",
				},
				TopicSource: "caraml-project-1-model-1-1-prediction-log",
			},
			consumerConfig: consumerConfig,

			resourceRequest: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1"),
				corev1.ResourceMemory: resource.MustParse("1Gi"),
			},
			resourceLimit: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse("1Gi"),
			},

			kubeClient: func() kubernetes.Interface {
				clientSet := fake.NewSimpleClientset()
				secretAPI := clientSet.CoreV1().Secrets(namespace).(*fakecorev1.FakeSecrets)
				prependGetSecretReactor(t, secretAPI, nil, nil)
				prependUpsertSecretReactor(t, secretAPI, []*corev1.Secret{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "project-1-model-1-config",
							Namespace: namespace,
							Labels: map[string]string{
								"app":          "model-1-observability-publisher",
								"component":    "worker",
								"orchestrator": "merlin",
								"stream":       "stream",
								"team":         "team",
								"environment":  "",
							},
						},
						StringData: map[string]string{
							"config.yaml": "project: project-1\nmodel_id: model-1\nmodel_version: \"1\"\ninference_schema:\n  session_id_column: session_id\n  row_id_column: row_id\n  model_prediction_output:\n    actual_score_column: \"\"\n    negative_class_label: negative\n    prediction_score_column: prediction_score\n    prediction_label_column: prediction_label\n    positive_class_label: positive\n    score_threshold: null\n    output_class: BinaryClassificationOutput\n  tag_columns:\n  - tag\n  feature_types:\n    featureA: float64\n    featureB: float64\n    featureC: int64\n    featureD: boolean\n  feature_orders: []\nobservation_sinks:\n- type: BIGQUERY\n  config:\n    project: bq-project\n    dataset: dataset\n    ttl_days: 10\n- type: MAXCOMPUTE\n  config:\n    project: max-project\n    dataset: dataset\n    ttl_days: 0\n    access_key_id: key\n    access_key_secret: secret\n    access_url: url\n- type: ARIZE\n  config:\n    api_key: api-key\n    space_key: space-key\nobservation_source:\n  type: KAFKA\n  config:\n    topic: caraml-project-1-model-1-1-prediction-log\n    bootstrap_servers: broker-1\n    group_id: group-id\n    batch_size: 100\n    additional_consumer_config:\n      auto.offset.reset: latest\n      fetch.min.bytes: \"1024000\"\n",
						},
					}}, fmt.Errorf("deployment control plane is down"), false)
				return clientSet
			}(),
			expectedErr: fmt.Errorf("deployment control plane is down"),
		},
		{
			name: "fresh deployment failed; failed during deployment",
			data: &models.WorkerData{
				Project:         "project-1",
				ModelSchemaSpec: schemaSpec,
				ModelName:       "model-1",
				ModelVersion:    "1",
				Revision:        1,
				Metadata: models.Metadata{
					App:       "model-1-observability-publisher",
					Component: "worker",
					Stream:    "stream",
					Team:      "team",
				},
				TopicSource: "caraml-project-1-model-1-1-prediction-log",
			},
			consumerConfig: consumerConfig,

			resourceRequest: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1"),
				corev1.ResourceMemory: resource.MustParse("1Gi"),
			},
			resourceLimit: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse("1Gi"),
			},

			kubeClient: func() kubernetes.Interface {
				clientSet := fake.NewSimpleClientset()
				secretAPI := clientSet.CoreV1().Secrets(namespace).(*fakecorev1.FakeSecrets)
				prependGetSecretReactor(t, secretAPI, nil, nil)
				prependUpsertSecretReactor(t, secretAPI, []*corev1.Secret{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "project-1-model-1-config",
							Namespace: namespace,
							Labels: map[string]string{
								"app":          "model-1-observability-publisher",
								"component":    "worker",
								"orchestrator": "merlin",
								"stream":       "stream",
								"team":         "team",
								"environment":  "",
							},
						},
						StringData: map[string]string{
							"config.yaml": "project: project-1\nmodel_id: model-1\nmodel_version: \"1\"\ninference_schema:\n  session_id_column: session_id\n  row_id_column: row_id\n  model_prediction_output:\n    actual_score_column: \"\"\n    negative_class_label: negative\n    prediction_score_column: prediction_score\n    prediction_label_column: prediction_label\n    positive_class_label: positive\n    score_threshold: null\n    output_class: BinaryClassificationOutput\n  tag_columns:\n  - tag\n  feature_types:\n    featureA: float64\n    featureB: float64\n    featureC: int64\n    featureD: boolean\n  feature_orders: []\nobservation_sinks:\n- type: BIGQUERY\n  config:\n    project: bq-project\n    dataset: dataset\n    ttl_days: 10\n- type: MAXCOMPUTE\n  config:\n    project: max-project\n    dataset: dataset\n    ttl_days: 0\n    access_key_id: key\n    access_key_secret: secret\n    access_url: url\n- type: ARIZE\n  config:\n    api_key: api-key\n    space_key: space-key\nobservation_source:\n  type: KAFKA\n  config:\n    topic: caraml-project-1-model-1-1-prediction-log\n    bootstrap_servers: broker-1\n    group_id: group-id\n    batch_size: 100\n    additional_consumer_config:\n      auto.offset.reset: latest\n      fetch.min.bytes: \"1024000\"\n",
						},
					}}, nil, false)
				deploymentAPI := clientSet.AppsV1().Deployments(namespace).(*fakeappsv1.FakeDeployments)
				preprendGetDeploymentReactor(t, deploymentAPI, nil, nil)
				depl := createDeploymentSpec(&models.WorkerData{
					Project:         "project-1",
					ModelSchemaSpec: schemaSpec,
					ModelName:       "model-1",
					ModelVersion:    "1",
					Revision:        1,
					Metadata: models.Metadata{
						App:       "model-1-observability-publisher",
						Component: "worker",
						Stream:    "stream",
						Team:      "team",
					},
				}, requestResource, limitResource, consumerConfig.ImageName, consumerConfig.ServiceAccountSecretName)
				prependUpsertDeploymentReactor(t, deploymentAPI, depl, fmt.Errorf("control plane is down"), false)
				prependDeleteSecretReactor(t, secretAPI, "project-1-model-1-config", nil)
				prependDeleteDeploymentReactor(t, deploymentAPI, "project-1-model-1-mlobs", nil)
				return clientSet
			}(),
			expectedErr: fmt.Errorf("control plane is down"),
		},
		{
			name: "redeployment",
			data: &models.WorkerData{
				Project:         "project-1",
				ModelSchemaSpec: schemaSpec,
				ModelName:       "model-1",
				ModelVersion:    "2",
				Revision:        2,
				Metadata: models.Metadata{
					App:       "model-1-observability-publisher",
					Component: "worker",
					Stream:    "stream",
					Team:      "team",
				},
				TopicSource: "caraml-project-1-model-1-2-prediction-log",
			},
			consumerConfig: consumerConfig,

			resourceRequest: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1"),
				corev1.ResourceMemory: resource.MustParse("1Gi"),
			},
			resourceLimit: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse("1Gi"),
			},

			kubeClient: func() kubernetes.Interface {
				clientSet := fake.NewSimpleClientset()
				secretAPI := clientSet.CoreV1().Secrets(namespace).(*fakecorev1.FakeSecrets)
				prependGetSecretReactor(t, secretAPI, &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "project-1-model-1-config",
						Namespace: namespace,
						Labels: map[string]string{
							"app":          "model-1-observability-publisher",
							"component":    "worker",
							"orchestrator": "merlin",
							"stream":       "stream",
							"team":         "team",
							"environment":  "",
						},
					},
					StringData: map[string]string{
						"config.yaml": "project: project-1\nmodel_id: model-1\nmodel_version: \"1\"\ninference_schema:\n  session_id_column: session_id\n  row_id_column: row_id\n  model_prediction_output:\n    actual_score_column: \"\"\n    negative_class_label: negative\n    prediction_score_column: prediction_score\n    prediction_label_column: prediction_label\n    positive_class_label: positive\n    score_threshold: null\n    output_class: BinaryClassificationOutput\n  tag_columns:\n  - tag\n  feature_types:\n    featureA: float64\n    featureB: float64\n    featureC: int64\n    featureD: boolean\n  feature_orders: []\nobservation_sinks:\n- type: BIGQUERY\n  config:\n    project: bq-project\n    dataset: dataset\n    ttl_days: 10\n- type: MAXCOMPUTE\n  config:\n    project: max-project\n    dataset: dataset\n    ttl_days: 0\n    access_key_id: key\n    access_key_secret: secret\n    access_url: url\n- type: ARIZE\n  config:\n    api_key: api-key\n    space_key: space-key\nobservation_source:\n  type: KAFKA\n  config:\n    topic: caraml-project-1-model-1-1-prediction-log\n    bootstrap_servers: broker-1\n    group_id: group-id\n    batch_size: 100\n    additional_consumer_config:\n      auto.offset.reset: latest\n      fetch.min.bytes: \"1024000\"\n",
					},
				}, nil)
				prependUpsertSecretReactor(t, secretAPI, []*corev1.Secret{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "project-1-model-1-config",
							Namespace: namespace,
							Labels: map[string]string{
								"app":          "model-1-observability-publisher",
								"component":    "worker",
								"orchestrator": "merlin",
								"stream":       "stream",
								"team":         "team",
								"environment":  "",
							},
						},
						StringData: map[string]string{
							"config.yaml": "project: project-1\nmodel_id: model-1\nmodel_version: \"2\"\ninference_schema:\n  session_id_column: session_id\n  row_id_column: row_id\n  model_prediction_output:\n    actual_score_column: \"\"\n    negative_class_label: negative\n    prediction_score_column: prediction_score\n    prediction_label_column: prediction_label\n    positive_class_label: positive\n    score_threshold: null\n    output_class: BinaryClassificationOutput\n  tag_columns:\n  - tag\n  feature_types:\n    featureA: float64\n    featureB: float64\n    featureC: int64\n    featureD: boolean\n  feature_orders: []\nobservation_sinks:\n- type: BIGQUERY\n  config:\n    project: bq-project\n    dataset: dataset\n    ttl_days: 10\n- type: MAXCOMPUTE\n  config:\n    project: max-project\n    dataset: dataset\n    ttl_days: 0\n    access_key_id: key\n    access_key_secret: secret\n    access_url: url\n- type: ARIZE\n  config:\n    api_key: api-key\n    space_key: space-key\nobservation_source:\n  type: KAFKA\n  config:\n    topic: caraml-project-1-model-1-2-prediction-log\n    bootstrap_servers: broker-1\n    group_id: group-id\n    batch_size: 100\n    additional_consumer_config:\n      auto.offset.reset: latest\n      fetch.min.bytes: \"1024000\"\n",
						},
					}}, nil, true)
				deploymentAPI := clientSet.AppsV1().Deployments(namespace).(*fakeappsv1.FakeDeployments)
				depl := createDeploymentSpec(&models.WorkerData{
					Project:         "project-1",
					ModelSchemaSpec: schemaSpec,
					ModelName:       "model-1",
					ModelVersion:    "2",
					Revision:        2,
					Metadata: models.Metadata{
						App:       "model-1-observability-publisher",
						Component: "worker",
						Stream:    "stream",
						Team:      "team",
					},
				}, requestResource, limitResource, consumerConfig.ImageName, consumerConfig.ServiceAccountSecretName)
				preprendGetDeploymentReactor(t, deploymentAPI, depl, nil)
				prependUpsertDeploymentReactor(t, deploymentAPI, depl, nil, true)

				updatedDepl := changeDeploymentStatus(depl, ready, 2)
				deplWatchReactor := newDeploymentWatchReactor(updatedDepl)
				clientSet.WatchReactionChain = []ktesting.WatchReactor{deplWatchReactor}

				return clientSet
			}(),
		},
		{
			name: "redeployment failed; timeout waiting for deployment",
			data: &models.WorkerData{
				Project:         "project-1",
				ModelSchemaSpec: schemaSpec,
				ModelName:       "model-1",
				ModelVersion:    "2",
				Revision:        2,
				Metadata: models.Metadata{
					App:       "model-1-observability-publisher",
					Component: "worker",
					Stream:    "stream",
					Team:      "team",
				},
				TopicSource: "caraml-project-1-model-1-2-prediction-log",
			},
			consumerConfig: consumerConfig,

			resourceRequest: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1"),
				corev1.ResourceMemory: resource.MustParse("1Gi"),
			},
			resourceLimit: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse("1Gi"),
			},

			kubeClient: func() kubernetes.Interface {
				clientSet := fake.NewSimpleClientset()
				secretAPI := clientSet.CoreV1().Secrets(namespace).(*fakecorev1.FakeSecrets)
				prependGetSecretReactor(t, secretAPI, &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "project-1-model-1-config",
						Namespace: namespace,
						Labels: map[string]string{
							"app":          "model-1-observability-publisher",
							"component":    "worker",
							"orchestrator": "merlin",
							"stream":       "stream",
							"team":         "team",
							"environment":  "",
						},
					},
					StringData: map[string]string{
						"config.yaml": "model_id: model-1\nmodel_version: \"1\"\ninference_schema:\n  session_id_column: session_id\n  row_id_column: row_id\n  model_prediction_output:\n    actual_score_column: \"\"\n    negative_class_label: negative\n    prediction_score_column: prediction_score\n    prediction_label_column: prediction_label\n    positive_class_label: positive\n    score_threshold: null\n    output_class: BinaryClassificationOutput\n  tag_columns:\n  - tag\n  feature_types:\n    featureA: float64\n    featureB: float64\n    featureC: int64\n    featureD: boolean\n  feature_orders: []\nobservation_sinks:\n- type: BIGQUERY\n  config:\n    project: bq-project\n    dataset: dataset\n    ttl_days: 10\n- type: MAXCOMPUTE\n  config:\n    project: max-project\n    dataset: dataset\n    ttl_days: 0\n    access_key_id: key\n    access_key_secret: secret\n    access_url: url\n- type: ARIZE\n  config:\n    api_key: api-key\n    space_key: space-key\nobservation_source:\n  type: KAFKA\n  config:\n    topic: caraml-project-1-model-1-1-prediction-log\n    bootstrap_servers: broker-1\n    group_id: group-id\n    batch_size: 100\n    additional_consumer_config:\n      auto.offset.reset: latest\n      fetch.min.bytes: \"1024000\"\n",
					},
				}, nil)
				prependUpsertSecretReactor(t, secretAPI, []*corev1.Secret{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "project-1-model-1-config",
							Namespace: namespace,
							Labels: map[string]string{
								"app":          "model-1-observability-publisher",
								"component":    "worker",
								"orchestrator": "merlin",
								"stream":       "stream",
								"team":         "team",
								"environment":  "",
							},
						},
						StringData: map[string]string{
							"config.yaml": "project: project-1\nmodel_id: model-1\nmodel_version: \"2\"\ninference_schema:\n  session_id_column: session_id\n  row_id_column: row_id\n  model_prediction_output:\n    actual_score_column: \"\"\n    negative_class_label: negative\n    prediction_score_column: prediction_score\n    prediction_label_column: prediction_label\n    positive_class_label: positive\n    score_threshold: null\n    output_class: BinaryClassificationOutput\n  tag_columns:\n  - tag\n  feature_types:\n    featureA: float64\n    featureB: float64\n    featureC: int64\n    featureD: boolean\n  feature_orders: []\nobservation_sinks:\n- type: BIGQUERY\n  config:\n    project: bq-project\n    dataset: dataset\n    ttl_days: 10\n- type: MAXCOMPUTE\n  config:\n    project: max-project\n    dataset: dataset\n    ttl_days: 0\n    access_key_id: key\n    access_key_secret: secret\n    access_url: url\n- type: ARIZE\n  config:\n    api_key: api-key\n    space_key: space-key\nobservation_source:\n  type: KAFKA\n  config:\n    topic: caraml-project-1-model-1-2-prediction-log\n    bootstrap_servers: broker-1\n    group_id: group-id\n    batch_size: 100\n    additional_consumer_config:\n      auto.offset.reset: latest\n      fetch.min.bytes: \"1024000\"\n",
						},
					}, {
						ObjectMeta: metav1.ObjectMeta{
							Name:      "project-1-model-1-config",
							Namespace: namespace,
							Labels: map[string]string{
								"app":          "model-1-observability-publisher",
								"component":    "worker",
								"orchestrator": "merlin",
								"stream":       "stream",
								"team":         "team",
								"environment":  "",
							},
						},
						StringData: map[string]string{
							"config.yaml": "model_id: model-1\nmodel_version: \"1\"\ninference_schema:\n  session_id_column: session_id\n  row_id_column: row_id\n  model_prediction_output:\n    actual_score_column: \"\"\n    negative_class_label: negative\n    prediction_score_column: prediction_score\n    prediction_label_column: prediction_label\n    positive_class_label: positive\n    score_threshold: null\n    output_class: BinaryClassificationOutput\n  tag_columns:\n  - tag\n  feature_types:\n    featureA: float64\n    featureB: float64\n    featureC: int64\n    featureD: boolean\n  feature_orders: []\nobservation_sinks:\n- type: BIGQUERY\n  config:\n    project: bq-project\n    dataset: dataset\n    ttl_days: 10\n- type: MAXCOMPUTE\n  config:\n    project: max-project\n    dataset: dataset\n    ttl_days: 0\n    access_key_id: key\n    access_key_secret: secret\n    access_url: url\n- type: ARIZE\n  config:\n    api_key: api-key\n    space_key: space-key\nobservation_source:\n  type: KAFKA\n  config:\n    topic: caraml-project-1-model-1-1-prediction-log\n    bootstrap_servers: broker-1\n    group_id: group-id\n    batch_size: 100\n    additional_consumer_config:\n      auto.offset.reset: latest\n      fetch.min.bytes: \"1024000\"\n",
						},
					},
				}, nil, true)
				deploymentAPI := clientSet.AppsV1().Deployments(namespace).(*fakeappsv1.FakeDeployments)
				depl := createDeploymentSpec(&models.WorkerData{
					Project:         "project-1",
					ModelSchemaSpec: schemaSpec,
					ModelName:       "model-1",
					ModelVersion:    "2",
					Revision:        2,
					Metadata: models.Metadata{
						App:       "model-1-observability-publisher",
						Component: "worker",
						Stream:    "stream",
						Team:      "team",
					},
				}, requestResource, limitResource, consumerConfig.ImageName, consumerConfig.ServiceAccountSecretName)
				preprendGetDeploymentReactor(t, deploymentAPI, depl, nil)
				prependUpsertDeploymentReactor(t, deploymentAPI, depl, nil, true)

				updatedDepl := changeDeploymentStatus(depl, timeoutError, 2)
				deplWatchReactor := newDeploymentWatchReactor(updatedDepl)
				clientSet.WatchReactionChain = []ktesting.WatchReactor{deplWatchReactor}

				return clientSet
			}(),
			expectedErr: fmt.Errorf(`deployment "project-1-model-1-mlobs" exceeded its progress deadline`),
		},
	}
	for _, tt := range tests {
		depl := &deployer{
			kubeClient:      tt.kubeClient,
			consumerConfig:  tt.consumerConfig,
			resourceRequest: tt.resourceRequest,
			resourceLimit:   tt.resourceLimit,
		}
		err := depl.Deploy(context.Background(), tt.data)
		assert.Equal(t, tt.expectedErr, err)
	}
}

func Test_deployer_Undeploy(t *testing.T) {

	testCases := []struct {
		name        string
		data        *models.WorkerData
		kubeClient  kubernetes.Interface
		expectedErr error
	}{
		{
			name: "success undeploy",
			data: &models.WorkerData{
				Project:      "project-1",
				ModelName:    "model-1",
				ModelVersion: "1",
				Revision:     1,
				Metadata: models.Metadata{
					App:       "model-1-observability-publisher",
					Component: "worker",
					Stream:    "stream",
					Team:      "team",
				},
			},
			kubeClient: func() kubernetes.Interface {
				clientSet := fake.NewSimpleClientset()
				deploymentAPI := clientSet.AppsV1().Deployments(namespace).(*fakeappsv1.FakeDeployments)
				prependDeleteDeploymentReactor(t, deploymentAPI, "project-1-model-1-mlobs", nil)

				secretAPI := clientSet.CoreV1().Secrets(namespace).(*fakecorev1.FakeSecrets)
				prependDeleteSecretReactor(t, secretAPI, "project-1-model-1-config", nil)
				return clientSet
			}(),
		},
		{
			name: "failed undeploy; error when delete deployment",
			data: &models.WorkerData{
				Project:      "project-1",
				ModelName:    "model-1",
				ModelVersion: "1",
				Revision:     1,
				Metadata: models.Metadata{
					App:       "model-1-observability-publisher",
					Component: "worker",
					Stream:    "stream",
					Team:      "team",
				},
			},
			kubeClient: func() kubernetes.Interface {
				clientSet := fake.NewSimpleClientset()
				deploymentAPI := clientSet.AppsV1().Deployments(namespace).(*fakeappsv1.FakeDeployments)
				prependDeleteDeploymentReactor(t, deploymentAPI, "project-1-model-1-mlobs", fmt.Errorf("control plane is down"))
				return clientSet
			}(),
			expectedErr: fmt.Errorf("control plane is down"),
		},
		{
			name: "faile undeploy; error when delete secret",
			data: &models.WorkerData{
				Project:      "project-1",
				ModelName:    "model-1",
				ModelVersion: "1",
				Revision:     1,
				Metadata: models.Metadata{
					App:       "model-1-observability-publisher",
					Component: "worker",
					Stream:    "stream",
					Team:      "team",
				},
			},
			kubeClient: func() kubernetes.Interface {
				clientSet := fake.NewSimpleClientset()
				deploymentAPI := clientSet.AppsV1().Deployments(namespace).(*fakeappsv1.FakeDeployments)
				prependDeleteDeploymentReactor(t, deploymentAPI, "project-1-model-1-mlobs", nil)

				secretAPI := clientSet.CoreV1().Secrets(namespace).(*fakecorev1.FakeSecrets)
				prependDeleteSecretReactor(t, secretAPI, "project-1-model-1-config", fmt.Errorf("control plane is down"))
				return clientSet
			}(),
			expectedErr: fmt.Errorf("control plane is down"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			consumerConfig := config.ObservabilityPublisher{
				TargetNamespace: namespace,
			}
			depl := &deployer{
				kubeClient:     tC.kubeClient,
				consumerConfig: consumerConfig,
			}
			err := depl.Undeploy(context.Background(), tC.data)
			assert.Equal(t, tC.expectedErr, err)
		})
	}
}

func Test_deployer_GetDeployedManifest(t *testing.T) {
	requestResource := corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse("1"),
		corev1.ResourceMemory: resource.MustParse("1Gi"),
	}
	limitResource := corev1.ResourceList{
		corev1.ResourceMemory: resource.MustParse("1Gi"),
	}
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
				NegativeClassLabel:    "negative",
				PositiveClassLabel:    "positive",
				PredictionLabelColumn: "prediction_label",
				PredictionScoreColumn: "prediction_score",
				OutputClass:           models.BinaryClassification,
			},
		},
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "project-1-model-1-config",
			Namespace: namespace,
			Labels: map[string]string{
				"app":          "model-1-observability-publisher",
				"component":    "worker",
				"orchestrator": "merlin",
				"stream":       "stream",
				"team":         "team",
				"environment":  "",
			},
		},
		StringData: map[string]string{
			"config.yaml": "model_id: model-1\nmodel_version: \"1\"\ninference_schema:\n  session_id_column: session_id\n  row_id_column: row_id\n  model_prediction_output:\n    actual_score_column: \"\"\n    negative_class_label: negative\n    prediction_score_column: prediction_score\n    prediction_label_column: prediction_label\n    positive_class_label: positive\n    score_threshold: null\n    output_class: BinaryClassificationOutput\n  tag_columns:\n  - tag\n  feature_types:\n    featureA: float64\n    featureB: float64\n    featureC: int64\n    featureD: boolean\n  feature_orders: []\nobservation_sinks:\n- type: BIGQUERY\n  config:\n    project: bq-project\n    dataset: dataset\n    ttl_days: 10\n- type: ARIZE\n  config:\n    api_key: api-key\n    space_key: space-key\n- type: MAXCOMPUTE\n  config:\n    project: max-project\n    dataset: dataset\n    ttl_days: 0\n    access_key_id: key\n    access_key_secret: secret\n    access_url: url\nobservation_source:\n  type: KAFKA\n  config:\n    topic: caraml-project-1-model-1-1-prediction-log\n    bootstrap_servers: broker-1\n    group_id: group-id\n    batch_size: 100\n    additional_consumer_config:\n      auto.offset.reset: latest\n      fetch.min.bytes: \"1024000\"\n",
		},
	}
	workerData := &models.WorkerData{
		Project:         "project-1",
		ModelSchemaSpec: schemaSpec,
		ModelName:       "model-1",
		ModelVersion:    "1",
		Revision:        1,
		Metadata: models.Metadata{
			App:       "model-1-observability-publisher",
			Component: "worker",
			Stream:    "stream",
			Team:      "team",
		},
		TopicSource: "caraml-project-1-model-1-1-prediction-log",
	}
	depl := createDeploymentSpec(workerData, requestResource, limitResource, "image:v0.1", serviceAccountSecretName)
	testCases := []struct {
		name             string
		data             *models.WorkerData
		kubeClient       kubernetes.Interface
		expectedErr      error
		expectedManifest *Manifest
	}{
		{
			name: "success get manifest",
			data: workerData,
			kubeClient: func() kubernetes.Interface {
				clientSet := fake.NewSimpleClientset()
				secretAPI := clientSet.CoreV1().Secrets(namespace).(*fakecorev1.FakeSecrets)
				prependGetSecretReactor(t, secretAPI, secret, nil)

				deploymentAPI := clientSet.AppsV1().Deployments(namespace).(*fakeappsv1.FakeDeployments)
				preprendGetDeploymentReactor(t, deploymentAPI, depl, nil)
				return clientSet
			}(),
			expectedManifest: &Manifest{
				Secret:     secret,
				Deployment: depl,
				OnProgress: true,
			},
		},
		{
			name: "success get manifest; rolled out deployment",
			data: workerData,
			kubeClient: func() kubernetes.Interface {
				clientSet := fake.NewSimpleClientset()
				secretAPI := clientSet.CoreV1().Secrets(namespace).(*fakecorev1.FakeSecrets)
				prependGetSecretReactor(t, secretAPI, secret, nil)

				updatedDeployment := changeDeploymentStatus(depl, ready, workerData.Revision)
				deploymentAPI := clientSet.AppsV1().Deployments(namespace).(*fakeappsv1.FakeDeployments)
				preprendGetDeploymentReactor(t, deploymentAPI, updatedDeployment, nil)
				return clientSet
			}(),
			expectedManifest: func() *Manifest {
				updatedDeployment := changeDeploymentStatus(depl, ready, workerData.Revision)
				manifest := &Manifest{
					Secret:     secret,
					Deployment: updatedDeployment,
					OnProgress: false,
				}
				return manifest
			}(),
		},
		{
			name: "deployment not exist but secret is exist treated as no deployment",
			data: workerData,
			kubeClient: func() kubernetes.Interface {
				clientSet := fake.NewSimpleClientset()
				secretAPI := clientSet.CoreV1().Secrets(namespace).(*fakecorev1.FakeSecrets)
				prependGetSecretReactor(t, secretAPI, secret, nil)

				deploymentAPI := clientSet.AppsV1().Deployments(namespace).(*fakeappsv1.FakeDeployments)
				preprendGetDeploymentReactor(t, deploymentAPI, nil, nil)
				return clientSet
			}(),
			expectedManifest: nil,
		},
		{
			name: "deployment is exist but secret is not, treated as not deployment manifest",
			data: workerData,
			kubeClient: func() kubernetes.Interface {
				clientSet := fake.NewSimpleClientset()
				secretAPI := clientSet.CoreV1().Secrets(namespace).(*fakecorev1.FakeSecrets)
				prependGetSecretReactor(t, secretAPI, nil, nil)

				updatedDeployment := changeDeploymentStatus(depl, ready, workerData.Revision)
				deploymentAPI := clientSet.AppsV1().Deployments(namespace).(*fakeappsv1.FakeDeployments)
				preprendGetDeploymentReactor(t, deploymentAPI, updatedDeployment, nil)
				return clientSet
			}(),
			expectedManifest: nil,
		},
		{
			name: "failed get manifest; error fetching secret",
			data: workerData,
			kubeClient: func() kubernetes.Interface {
				clientSet := fake.NewSimpleClientset()
				secretAPI := clientSet.CoreV1().Secrets(namespace).(*fakecorev1.FakeSecrets)
				prependGetSecretReactor(t, secretAPI, secret, fmt.Errorf("control plane is down"))

				return clientSet
			}(),
			expectedErr: fmt.Errorf("control plane is down"),
		},
		{
			name: "failed get manifest; error fetching deployment",
			data: workerData,
			kubeClient: func() kubernetes.Interface {
				clientSet := fake.NewSimpleClientset()
				deploymentAPI := clientSet.AppsV1().Deployments(namespace).(*fakeappsv1.FakeDeployments)
				preprendGetDeploymentReactor(t, deploymentAPI, depl, fmt.Errorf("control plane is down"))

				return clientSet
			}(),
			expectedErr: fmt.Errorf("control plane is down"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			consumerConfig := config.ObservabilityPublisher{
				TargetNamespace: namespace,
			}
			depl := &deployer{
				kubeClient:     tC.kubeClient,
				consumerConfig: consumerConfig,
			}
			manifest, err := depl.GetDeployedManifest(context.Background(), tC.data)
			assert.Equal(t, tC.expectedErr, err)
			assert.Equal(t, tC.expectedManifest, manifest)
		})
	}
}

func toQuantityPointer(quantity resource.Quantity) *resource.Quantity {
	return &quantity
}

func Test_deployer_getResources(t *testing.T) {
	testCases := []struct {
		name                     string
		deployer                 *deployer
		data                     *models.WorkerData
		expectedRequestResources corev1.ResourceList
		expectedLimitResources   corev1.ResourceList
	}{
		{
			name: "worker data doesn't have resource request and limit hence using default",
			deployer: &deployer{
				resourceRequest: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1"),
					corev1.ResourceMemory: resource.MustParse("1Gi"),
				},
				resourceLimit: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1"),
					corev1.ResourceMemory: resource.MustParse("1Gi"),
				},
			},
			data: &models.WorkerData{},
			expectedRequestResources: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1"),
				corev1.ResourceMemory: resource.MustParse("1Gi"),
			},
			expectedLimitResources: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1"),
				corev1.ResourceMemory: resource.MustParse("1Gi"),
			},
		},
		{
			name: "worker data have cpu and memory resource request",
			deployer: &deployer{
				resourceRequest: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1"),
					corev1.ResourceMemory: resource.MustParse("1Gi"),
				},
				resourceLimit: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1"),
					corev1.ResourceMemory: resource.MustParse("1Gi"),
				},
			},
			data: &models.WorkerData{
				ResourceRequest: &models.WorkerResourceRequest{
					CPURequest:    toQuantityPointer(resource.MustParse("2")),
					MemoryRequest: toQuantityPointer(resource.MustParse("2Gi")),
				},
			},
			expectedRequestResources: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("2"),
				corev1.ResourceMemory: resource.MustParse("2Gi"),
			},
			expectedLimitResources: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse("2Gi"),
			},
		},
		{
			name: "worker data have cpu and memory resource request, default doesn't have cpu limit",
			deployer: &deployer{
				resourceRequest: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1"),
					corev1.ResourceMemory: resource.MustParse("1Gi"),
				},
				resourceLimit: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("1Gi"),
				},
			},
			data: &models.WorkerData{
				ResourceRequest: &models.WorkerResourceRequest{
					CPURequest:    toQuantityPointer(resource.MustParse("2")),
					MemoryRequest: toQuantityPointer(resource.MustParse("2Gi")),
				},
			},
			expectedRequestResources: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("2"),
				corev1.ResourceMemory: resource.MustParse("2Gi"),
			},
			expectedLimitResources: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse("2Gi"),
			},
		},
		{
			name: "worker data have cpu resource request but not memory",
			deployer: &deployer{
				resourceRequest: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1"),
					corev1.ResourceMemory: resource.MustParse("1Gi"),
				},
				resourceLimit: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1"),
					corev1.ResourceMemory: resource.MustParse("2Gi"),
				},
			},
			data: &models.WorkerData{
				ResourceRequest: &models.WorkerResourceRequest{
					CPURequest: toQuantityPointer(resource.MustParse("2")),
				},
			},
			expectedRequestResources: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("2"),
				corev1.ResourceMemory: resource.MustParse("1Gi"),
			},
			expectedLimitResources: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse("2Gi"),
			},
		},
		{
			name: "worker data have memory resource request but not cpu",
			deployer: &deployer{
				resourceRequest: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1"),
					corev1.ResourceMemory: resource.MustParse("1Gi"),
				},
				resourceLimit: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1"),
					corev1.ResourceMemory: resource.MustParse("1Gi"),
				},
			},
			data: &models.WorkerData{
				ResourceRequest: &models.WorkerResourceRequest{
					MemoryRequest: toQuantityPointer(resource.MustParse("2Gi")),
				},
			},
			expectedRequestResources: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1"),
				corev1.ResourceMemory: resource.MustParse("2Gi"),
			},
			expectedLimitResources: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1"),
				corev1.ResourceMemory: resource.MustParse("2Gi"),
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			requestResource, limitResource := tC.deployer.getResources(tC.data)
			assert.Equal(t, tC.expectedRequestResources, requestResource)
			assert.Equal(t, tC.expectedLimitResources, limitResource)
		})
	}
}

func prependGetSecretReactor(t *testing.T, secretAPI *fakecorev1.FakeSecrets, secretRet *corev1.Secret, expectedErr error) {
	secretAPI.Fake.PrependReactor(getMethod, secretResource, func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
		actualAction, ok := action.(ktesting.GetAction)
		if !ok {
			t.Fatalf("unexpected actual action")
		}
		if secretRet == nil {
			return true, nil, &errors.StatusError{
				ErrStatus: metav1.Status{
					Code:   http.StatusNotFound,
					Reason: metav1.StatusReasonNotFound,
				}}
		}

		if actualAction.GetNamespace() != secretRet.GetNamespace() {
			t.Fatalf("different namespace")
		}
		if actualAction.GetName() != secretRet.GetName() {
			t.Fatalf("requested different secret name")
		}

		return true, secretRet, expectedErr
	})

}

func prependUpsertSecretReactor(t *testing.T, secretAPI *fakecorev1.FakeSecrets, requestedSecrets []*corev1.Secret, expectedErr error, updateOperation bool) {
	method := createMethod
	if updateOperation {
		method = updateMethod
	}
	secretAPI.Fake.PrependReactor(method, secretResource, func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
		var actualReqSecret *corev1.Secret
		if updateOperation {
			actualAction := action.(ktesting.UpdateAction)
			actualReqSecret = actualAction.GetObject().(*corev1.Secret)
		} else {
			actualAction := action.(ktesting.CreateAction)
			actualReqSecret = actualAction.GetObject().(*corev1.Secret)
		}

		foundAction := false
		var secret *corev1.Secret
		for _, requestedSecret := range requestedSecrets {
			foundAction = actualReqSecret.Namespace == requestedSecret.Namespace && reflect.DeepEqual(requestedSecret.ObjectMeta, actualReqSecret.ObjectMeta) && reflect.DeepEqual(requestedSecret.StringData, actualReqSecret.StringData)
			if foundAction {
				secret = requestedSecret
				break
			}
		}

		if !foundAction {
			t.Fatalf("actual and expected secret is different")
		}

		return true, secret, expectedErr
	})
}

func prependDeleteSecretReactor(t *testing.T, secretAPI *fakecorev1.FakeSecrets, secretName string, expectedErr error) {
	secretAPI.Fake.PrependReactor(deleteMethod, secretResource, func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
		actualAction, ok := action.(ktesting.DeleteAction)
		if !ok {
			t.Fatalf("unexpected actual action")
		}
		if actualAction.GetName() != secretName {
			t.Fatalf("requested and actual secret name is not the same")
		}
		return true, nil, expectedErr
	})
}

func preprendGetDeploymentReactor(t *testing.T, deploymentAPI *fakeappsv1.FakeDeployments, deploymentRet *appsv1.Deployment, expectedErr error) {
	deploymentAPI.Fake.PrependReactor(getMethod, deploymentResource, func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
		actualAction, ok := action.(ktesting.GetAction)
		if !ok {
			t.Fatalf("unexpected actual action")
		}
		if deploymentRet == nil {
			return true, nil, &errors.StatusError{
				ErrStatus: metav1.Status{
					Code:   http.StatusNotFound,
					Reason: metav1.StatusReasonNotFound,
				}}
		}
		if actualAction.GetName() != deploymentRet.GetName() {
			t.Fatalf("requested and actual deployment name is different")
		}
		return true, deploymentRet, expectedErr
	})
}

func prependUpsertDeploymentReactor(t *testing.T, deploymentAPI *fakeappsv1.FakeDeployments, requestedDepl *appsv1.Deployment, expectedErr error, updateOperation bool) {
	method := createMethod
	if updateOperation {
		method = updateMethod
	}

	deploymentAPI.Fake.PrependReactor(method, deploymentResource, func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
		var actualReqDepl *appsv1.Deployment
		if updateOperation {
			actualAction := action.(ktesting.CreateAction)
			actualReqDepl = actualAction.GetObject().(*appsv1.Deployment)
		} else {
			actualAction := action.(ktesting.UpdateAction)
			actualReqDepl = actualAction.GetObject().(*appsv1.Deployment)
		}

		if actualReqDepl.Namespace != requestedDepl.GetNamespace() {
			t.Fatalf("different namespace")
		}

		assert.Equal(t, requestedDepl.ObjectMeta, actualReqDepl.ObjectMeta)
		assert.Equal(t, requestedDepl.Spec, actualReqDepl.Spec)
		if !reflect.DeepEqual(requestedDepl.ObjectMeta, actualReqDepl.ObjectMeta) || !reflect.DeepEqual(requestedDepl.Spec, actualReqDepl.Spec) {
			t.Fatalf("actual and expected requested deployment is different")
		}

		return true, requestedDepl, expectedErr
	})
}

func prependDeleteDeploymentReactor(t *testing.T, deploymentAPI *fakeappsv1.FakeDeployments, deploymentName string, expectedErr error) {
	deploymentAPI.Fake.PrependReactor(deleteMethod, deploymentResource, func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
		actualAction, ok := action.(ktesting.DeleteAction)
		if !ok {
			t.Fatalf("unexpected actual action")
		}

		if actualAction.GetName() != deploymentName {
			t.Fatalf("requested and actual deployment name is not the same")
		}
		return true, nil, expectedErr
	})
}
