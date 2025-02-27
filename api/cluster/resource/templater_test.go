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

package resource

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/caraml-dev/merlin/cluster/labeller"
	kservev1beta1 "github.com/kserve/kserve/pkg/apis/serving/v1beta1"
	kserveconstant "github.com/kserve/kserve/pkg/constants"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	knautoscaling "knative.dev/serving/pkg/apis/autoscaling"
	knserving "knative.dev/serving/pkg/apis/serving"

	"github.com/caraml-dev/merlin/config"
	"github.com/caraml-dev/merlin/mlp"
	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/pkg/autoscaling"
	"github.com/caraml-dev/merlin/pkg/deployment"
	"github.com/caraml-dev/merlin/pkg/protocol"
	transformerpkg "github.com/caraml-dev/merlin/pkg/transformer"
)

const (
	testEnvironmentName  = "dev"
	testOrchestratorName = "merlin"
)

var (
	defaultModelResourceRequests = &config.ResourceRequests{
		MinReplica:    1,
		MaxReplica:    2,
		CPURequest:    resource.MustParse("500m"),
		MemoryRequest: resource.MustParse("500Mi"),
	}

	expDefaultModelResourceRequests = corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    defaultModelResourceRequests.CPURequest,
			corev1.ResourceMemory: defaultModelResourceRequests.MemoryRequest,
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU: resource.MustParse("10"),
			corev1.ResourceMemory: ScaleQuantity(
				defaultModelResourceRequests.MemoryRequest,
				userContainerMemoryLimitRequestFactor,
			),
		},
	}

	defaultTransformerResourceRequests = &config.ResourceRequests{
		MinReplica:    1,
		MaxReplica:    2,
		CPURequest:    resource.MustParse("100m"),
		MemoryRequest: resource.MustParse("500Mi"),
	}

	expDefaultTransformerResourceRequests = corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    defaultTransformerResourceRequests.CPURequest,
			corev1.ResourceMemory: defaultTransformerResourceRequests.MemoryRequest,
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("10"),
			corev1.ResourceMemory: ScaleQuantity(defaultTransformerResourceRequests.MemoryRequest, 2),
		},
	}

	userResourceRequests = &models.ResourceRequest{
		MinReplica:    1,
		MaxReplica:    10,
		CPURequest:    resource.MustParse("1"),
		MemoryRequest: resource.MustParse("1Gi"),
	}

	cpuLimit                          = resource.MustParse("8")
	userResourceRequestsWithCPULimits = &models.ResourceRequest{
		MinReplica:    1,
		MaxReplica:    10,
		CPURequest:    resource.MustParse("1"),
		MemoryRequest: resource.MustParse("1Gi"),
		CPULimit:      &cpuLimit,
	}

	expUserResourceRequests = corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    userResourceRequests.CPURequest,
			corev1.ResourceMemory: userResourceRequests.MemoryRequest,
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("10"),
			corev1.ResourceMemory: ScaleQuantity(userResourceRequests.MemoryRequest, 2),
		},
	}

	expUserResourceRequestsWithCPULimits = corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    userResourceRequestsWithCPULimits.CPURequest,
			corev1.ResourceMemory: userResourceRequestsWithCPULimits.MemoryRequest,
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    *userResourceRequestsWithCPULimits.CPULimit,
			corev1.ResourceMemory: ScaleQuantity(userResourceRequestsWithCPULimits.MemoryRequest, 2),
		},
	}

	testPredictorScale, testTransformerScale = 3, 5

	defaultDeploymentScale = DeploymentScale{
		Predictor:   &testPredictorScale,
		Transformer: &testTransformerScale,
	}

	oneMinuteDuration         = time.Minute * 1
	twoMinuteDuration         = time.Minute * 2
	standardTransformerConfig = config.StandardTransformerConfig{
		ImageName:    "merlin-standard-transformer",
		FeastCoreURL: "core.feast.dev:8081",
		Jaeger: config.JaegerConfig{
			CollectorURL: "http://jaeger-tracing-collector.infrastructure:14268/api/traces",
			SamplerParam: "1",
			Disabled:     "false",
		},
		FeastGPRCConnCount:   5,
		ModelServerConnCount: 3,
		FeastRedisConfig: &config.FeastRedisConfig{
			IsRedisCluster: true,
			ServingURL:     "localhost:6866",
			RedisAddresses: []string{"10.1.1.2", "10.1.1.3"},
			PoolSize:       5,
			MinIdleConn:    2,
		},
		FeastServingKeepAlive: &config.FeastServingKeepAliveConfig{
			Enabled: true,
			Time:    30 * time.Second,
			Timeout: 1 * time.Second,
		},
		FeastBigtableConfig: &config.FeastBigtableConfig{
			ServingURL:        "localhost:6867",
			Project:           "gcp-project",
			Instance:          "instance",
			AppProfile:        "default",
			PoolSize:          4,
			KeepAliveInterval: &twoMinuteDuration,
			KeepAliveTimeout:  &oneMinuteDuration,
		},
		BigtableCredential: "eyJrZXkiOiJ2YWx1ZSJ9",
		ModelClientKeepAlive: &config.ModelClientKeepAliveConfig{
			Enabled: true,
			Time:    45 * time.Second,
			Timeout: 5 * time.Second,
		},
		Kafka: config.KafkaConfig{
			Topic:               "",
			Brokers:             "kafka-brokers",
			CompressionType:     "none",
			MaxMessageSizeBytes: 1048588,
			ConnectTimeoutMS:    1000,
			SerializationFmt:    "protobuf",
		},
	}

	pyfuncPublisherConfig = config.PyFuncPublisherConfig{
		Kafka: config.KafkaConfig{
			Brokers:           "kafka-broker:1111",
			LingerMS:          1000,
			Acks:              0,
			NumPartitions:     24,
			ReplicationFactor: 3,
			AdditionalConfig:  "{}",
		},
		SamplingRatioRate: 0.1,
	}

	userContainerCPUDefaultLimit          = "10"
	userContainerCPULimitRequestFactor    = float64(0)
	userContainerMemoryLimitRequestFactor = float64(2)

	defaultWorkersEnvVarName  = "WORKERS"
	defaultWorkersEnvVarValue = "2"

	defaultEnvVarWithoutCPULimits = corev1.EnvVar{
		Name:  defaultWorkersEnvVarName,
		Value: defaultWorkersEnvVarValue,
	}
)

func TestCreateInferenceServiceSpec(t *testing.T) {
	err := labeller.InitKubernetesLabeller("gojek.com/", "caraml.dev/", testEnvironmentName)
	assert.NoError(t, err)

	defer func() {
		_ = labeller.InitKubernetesLabeller("", "", "")
	}()

	project := mlp.Project{
		Name: "project",
	}
	modelSvc := &models.Service{
		Name:         "my-model-1",
		ModelName:    "my-model",
		Namespace:    project.Name,
		ModelVersion: "1",
		ArtifactURI:  "gs://my-artifacet",
		Metadata: models.Metadata{
			App:       "model",
			Component: models.ComponentModelVersion,
			Stream:    "dsp",
			Team:      "dsp",
			Labels: mlp.Labels{
				{
					Key:   "sample",
					Value: "true",
				},
			},
		},
		Protocol:                  protocol.HttpJson,
		EnabledModelObservability: true,
	}

	modelSvcWithSchema := &models.Service{
		Name:         "my-model-1",
		ModelName:    "my-model",
		Namespace:    project.Name,
		ModelVersion: "1",
		ArtifactURI:  "gs://my-artifacet",
		Metadata: models.Metadata{
			App:       "model",
			Component: models.ComponentModelVersion,
			Stream:    "dsp",
			Team:      "dsp",
			Labels: mlp.Labels{
				{
					Key:   "sample",
					Value: "true",
				},
			},
		},
		Protocol:                  protocol.HttpJson,
		EnabledModelObservability: true,
		ModelSchema: &models.ModelSchema{
			ID:      models.ID(1),
			ModelID: models.ID(1),
			Spec: &models.SchemaSpec{
				SessionIDColumn: "session_id",
				RowIDColumn:     "row_id",
				TagColumns:      []string{"tags"},
				FeatureTypes: map[string]models.ValueType{
					"featureA": models.Float64,
					"featureB": models.Int64,
					"featureC": models.String,
					"featureD": models.Boolean,
				},
				FeatureOrders: []string{"featureA", "featureB", "featureC", "featureD"},
				ModelPredictionOutput: &models.ModelPredictionOutput{
					RankingOutput: &models.RankingOutput{
						RankScoreColumn:      "score",
						RelevanceScoreColumn: "relevance_score",
					},
				},
			},
		},
	}

	queueResourcePercentage := "2"
	storageUri := fmt.Sprintf("%s/model", modelSvc.ArtifactURI)

	// Liveness probe config for the model containers
	probeConfig := createLivenessProbeSpec(protocol.HttpJson, fmt.Sprintf("/v1/models/%s", modelSvc.Name))
	probeConfigUPI := createLivenessProbeSpec(protocol.UpiV1, fmt.Sprintf("/v1/models/%s", modelSvc.Name))

	tests := []struct {
		name               string
		modelSvc           *models.Service
		resourcePercentage string
		deploymentScale    DeploymentScale
		exp                *kservev1beta1.InferenceService
		wantErr            bool
	}{
		{
			name: "tensorflow spec",
			modelSvc: &models.Service{
				Name:         modelSvc.Name,
				ModelName:    modelSvc.ModelName,
				ModelVersion: modelSvc.ModelVersion,
				Namespace:    project.Name,
				ArtifactURI:  modelSvc.ArtifactURI,
				Type:         models.ModelTypeTensorflow,
				Options:      &models.ModelOption{},
				Metadata:     modelSvc.Metadata,
				Protocol:     protocol.HttpJson,
			},
			resourcePercentage: queueResourcePercentage,
			deploymentScale:    defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testPredictorScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						Tensorflow: &kservev1beta1.TFServingSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     expDefaultModelResourceRequests,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{defaultEnvVarWithoutCPULimits},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
		{
			name: "tensorflow wth user-provided env var",
			modelSvc: &models.Service{
				Name:         modelSvc.Name,
				ModelName:    modelSvc.ModelName,
				ModelVersion: modelSvc.ModelVersion,
				Namespace:    project.Name,
				ArtifactURI:  modelSvc.ArtifactURI,
				Type:         models.ModelTypeTensorflow,
				Options:      &models.ModelOption{},
				Metadata:     modelSvc.Metadata,
				Protocol:     protocol.HttpJson,
				EnvVars: models.EnvVars{
					{
						Name: "env1", Value: "env1Value",
					},
					{
						Name: "env2", Value: "env2Value",
					},
				},
			},
			resourcePercentage: queueResourcePercentage,
			deploymentScale:    defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testPredictorScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						Tensorflow: &kservev1beta1.TFServingSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     expDefaultModelResourceRequests,
									LivenessProbe: probeConfig,
									Env: []corev1.EnvVar{
										defaultEnvVarWithoutCPULimits,
										{
											Name: "env1", Value: "env1Value",
										},
										{
											Name: "env2", Value: "env2Value",
										},
									},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
		{
			name: "tensorflow spec as raw deployment",
			modelSvc: &models.Service{
				Name:           modelSvc.Name,
				ModelName:      modelSvc.ModelName,
				ModelVersion:   modelSvc.ModelVersion,
				Namespace:      project.Name,
				ArtifactURI:    modelSvc.ArtifactURI,
				Type:           models.ModelTypeTensorflow,
				Options:        &models.ModelOption{},
				Metadata:       modelSvc.Metadata,
				DeploymentMode: deployment.RawDeploymentMode,
				Protocol:       protocol.HttpJson,
			},
			resourcePercentage: queueResourcePercentage,
			deploymentScale:    defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.RawDeployment),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						Tensorflow: &kservev1beta1.TFServingSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     expDefaultModelResourceRequests,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{defaultEnvVarWithoutCPULimits},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
		{
			name: "tensorflow spec as serverless",
			modelSvc: &models.Service{
				Name:           modelSvc.Name,
				ModelName:      modelSvc.ModelName,
				ModelVersion:   modelSvc.ModelVersion,
				Namespace:      project.Name,
				ArtifactURI:    modelSvc.ArtifactURI,
				Type:           models.ModelTypeTensorflow,
				Options:        &models.ModelOption{},
				Metadata:       modelSvc.Metadata,
				DeploymentMode: deployment.ServerlessDeploymentMode,
				Protocol:       protocol.HttpJson,
			},
			resourcePercentage: queueResourcePercentage,
			deploymentScale:    defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testPredictorScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						Tensorflow: &kservev1beta1.TFServingSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     expDefaultModelResourceRequests,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{defaultEnvVarWithoutCPULimits},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
		{
			name: "tensorflow spec without queue resource percentage",
			modelSvc: &models.Service{
				Name:         modelSvc.Name,
				ModelName:    modelSvc.ModelName,
				ModelVersion: modelSvc.ModelVersion,
				Namespace:    project.Name,
				ArtifactURI:  modelSvc.ArtifactURI,
				Type:         models.ModelTypeTensorflow,
				Options:      &models.ModelOption{},
				Metadata:     modelSvc.Metadata,
				Protocol:     protocol.HttpJson,
			},
			deploymentScale: defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						kserveconstant.DeploymentMode:           string(kserveconstant.Serverless),
						knautoscaling.InitialScaleAnnotationKey: fmt.Sprint(testPredictorScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						Tensorflow: &kservev1beta1.TFServingSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     expDefaultModelResourceRequests,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{defaultEnvVarWithoutCPULimits},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
		{
			name: "xgboost spec",
			modelSvc: &models.Service{
				Name:         modelSvc.Name,
				ModelName:    modelSvc.ModelName,
				ModelVersion: modelSvc.ModelVersion,
				Namespace:    project.Name,
				ArtifactURI:  modelSvc.ArtifactURI,
				Type:         models.ModelTypeXgboost,
				Options:      &models.ModelOption{},
				Metadata:     modelSvc.Metadata,
				Protocol:     protocol.HttpJson,
			},
			resourcePercentage: queueResourcePercentage,
			deploymentScale:    defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testPredictorScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						XGBoost: &kservev1beta1.XGBoostSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     expDefaultModelResourceRequests,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{defaultEnvVarWithoutCPULimits},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
		{
			name: "sklearn spec",
			modelSvc: &models.Service{
				Name:         modelSvc.Name,
				ModelName:    modelSvc.ModelName,
				ModelVersion: modelSvc.ModelVersion,
				Namespace:    project.Name,
				ArtifactURI:  modelSvc.ArtifactURI,
				Type:         models.ModelTypeSkLearn,
				Options:      &models.ModelOption{},
				Metadata:     modelSvc.Metadata,
				Protocol:     protocol.HttpJson,
			},
			resourcePercentage: queueResourcePercentage,
			deploymentScale:    defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testPredictorScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						SKLearn: &kservev1beta1.SKLearnSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     expDefaultModelResourceRequests,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{defaultEnvVarWithoutCPULimits},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
		{
			name: "pytorch spec",
			modelSvc: &models.Service{
				Name:         modelSvc.Name,
				ModelName:    modelSvc.ModelName,
				ModelVersion: modelSvc.ModelVersion,
				Namespace:    project.Name,
				ArtifactURI:  modelSvc.ArtifactURI,
				Type:         models.ModelTypePyTorch,
				Metadata:     modelSvc.Metadata,
				Protocol:     protocol.HttpJson,
			},
			resourcePercentage: queueResourcePercentage,
			deploymentScale:    defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testPredictorScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						PyTorch: &kservev1beta1.TorchServeSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     expDefaultModelResourceRequests,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{defaultEnvVarWithoutCPULimits},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
		{
			name: "pyfunc spec with cpu limits specified",
			modelSvc: &models.Service{
				Name:         modelSvc.Name,
				ModelName:    modelSvc.ModelName,
				ModelVersion: modelSvc.ModelVersion,
				Namespace:    project.Name,
				ArtifactURI:  modelSvc.ArtifactURI,
				Type:         models.ModelTypePyFunc,
				Options: &models.ModelOption{
					PyFuncImageName: "gojek/project-model:1",
				},
				ResourceRequest: userResourceRequestsWithCPULimits,
				Metadata:        modelSvc.Metadata,
				Protocol:        protocol.HttpJson,
			},
			resourcePercentage: queueResourcePercentage,
			deploymentScale:    defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						"prometheus.io/scrape":                                "true",
						"prometheus.io/port":                                  "8080",
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testPredictorScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						PodSpec: kservev1beta1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:          kserveconstant.InferenceServiceContainerName,
									Image:         "gojek/project-model:1",
									Env:           createPyFuncDefaultEnvVarsWithProtocol(modelSvc, protocol.HttpJson).ToKubernetesEnvVars(),
									Resources:     expUserResourceRequestsWithCPULimits,
									LivenessProbe: probeConfig,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &userResourceRequestsWithCPULimits.MinReplica,
							MaxReplicas: userResourceRequestsWithCPULimits.MaxReplica,
						},
					},
				},
			},
		},
		{
			name: "pyfunc spec with user-configured secrets specified",
			modelSvc: &models.Service{
				Name:         modelSvc.Name,
				ModelName:    modelSvc.ModelName,
				ModelVersion: modelSvc.ModelVersion,
				Namespace:    project.Name,
				ArtifactURI:  modelSvc.ArtifactURI,
				Type:         models.ModelTypePyFunc,
				Options: &models.ModelOption{
					PyFuncImageName: "gojek/project-model:1",
				},
				ResourceRequest: userResourceRequestsWithCPULimits,
				Metadata:        modelSvc.Metadata,
				Protocol:        protocol.HttpJson,
				Secrets: models.Secrets{
					{
						MLPSecretName: "SECRET_NAME_1",
						EnvVarName:    "ENV_SECRET_NAME_1",
					},
				},
			},
			resourcePercentage: queueResourcePercentage,
			deploymentScale:    defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						"prometheus.io/scrape":                                "true",
						"prometheus.io/port":                                  "8080",
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testPredictorScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						PodSpec: kservev1beta1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  kserveconstant.InferenceServiceContainerName,
									Image: "gojek/project-model:1",
									Env: MergeEnvVars(
										createPyFuncDefaultEnvVarsWithProtocol(modelSvc, protocol.HttpJson).ToKubernetesEnvVars(),
										[]corev1.EnvVar{
											{
												Name: "ENV_SECRET_NAME_1",
												ValueFrom: &corev1.EnvVarSource{
													SecretKeyRef: &corev1.SecretKeySelector{
														LocalObjectReference: corev1.LocalObjectReference{
															Name: modelSvc.Name,
														},
														Key: "SECRET_NAME_1",
													},
												},
											},
										},
									),
									Resources:     expUserResourceRequestsWithCPULimits,
									LivenessProbe: probeConfig,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &userResourceRequestsWithCPULimits.MinReplica,
							MaxReplicas: userResourceRequestsWithCPULimits.MaxReplica,
						},
					},
				},
			},
		},
		{
			name: "pyfunc spec with liveness probe disabled (old)",
			modelSvc: &models.Service{
				Name:         modelSvc.Name,
				ModelName:    modelSvc.ModelName,
				ModelVersion: modelSvc.ModelVersion,
				Namespace:    project.Name,
				ArtifactURI:  modelSvc.ArtifactURI,
				Type:         models.ModelTypePyFunc,
				Options: &models.ModelOption{
					PyFuncImageName: "gojek/project-model:1",
				},
				EnvVars:  models.EnvVars{models.EnvVar{Name: envOldDisableLivenessProbe, Value: "true"}},
				Metadata: modelSvc.Metadata,
				Protocol: protocol.HttpJson,
			},
			resourcePercentage: queueResourcePercentage,
			deploymentScale:    defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						"prometheus.io/scrape":                                "true",
						"prometheus.io/port":                                  "8080",
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testPredictorScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						PodSpec: kservev1beta1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  kserveconstant.InferenceServiceContainerName,
									Image: "gojek/project-model:1",
									Env: models.MergeEnvVars(models.EnvVars{
										{Name: defaultWorkersEnvVarName, Value: defaultWorkersEnvVarValue},
										{Name: envOldDisableLivenessProbe, Value: "true"},
									},
										createPyFuncDefaultEnvVarsWithProtocol(modelSvc, protocol.HttpJson),
									).ToKubernetesEnvVars(),
									Resources: expDefaultModelResourceRequests,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
		{
			name: "pyfunc spec with model observability enabled, overwrite the sampling ratio",
			modelSvc: &models.Service{
				Name:         modelSvc.Name,
				ModelName:    modelSvc.ModelName,
				ModelVersion: modelSvc.ModelVersion,
				Namespace:    project.Name,
				ArtifactURI:  modelSvc.ArtifactURI,
				Type:         models.ModelTypePyFunc,
				Options: &models.ModelOption{
					PyFuncImageName: "gojek/project-model:1",
				},
				EnvVars: models.EnvVars{
					{
						Name:  envPublisherSamplingRatio,
						Value: "0.5",
					},
				},
				Metadata:                  modelSvc.Metadata,
				Protocol:                  protocol.HttpJson,
				EnabledModelObservability: true,
				ModelSchema: &models.ModelSchema{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Spec: &models.SchemaSpec{
						SessionIDColumn: "session_id",
						RowIDColumn:     "row_id",
						TagColumns:      []string{"tags"},
						FeatureTypes: map[string]models.ValueType{
							"featureA": models.Float64,
							"featureB": models.Int64,
							"featureC": models.String,
							"featureD": models.Boolean,
						},
						FeatureOrders: []string{"featureA", "featureB", "featureC", "featureD"},
						ModelPredictionOutput: &models.ModelPredictionOutput{
							RankingOutput: &models.RankingOutput{
								RankScoreColumn:      "score",
								RelevanceScoreColumn: "relevance_score",
							},
						},
					},
				},
			},
			resourcePercentage: queueResourcePercentage,
			deploymentScale:    defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						"prometheus.io/scrape":                                "true",
						"prometheus.io/port":                                  "8080",
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testPredictorScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						PodSpec: kservev1beta1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  kserveconstant.InferenceServiceContainerName,
									Image: "gojek/project-model:1",
									Env: models.MergeEnvVars(
										createPyFuncPublisherEnvVars(modelSvcWithSchema, pyfuncPublisherConfig),
										models.EnvVars{
											{Name: defaultWorkersEnvVarName, Value: defaultWorkersEnvVarValue},
											{Name: envPublisherSamplingRatio, Value: "0.5"},
										},
										createPyFuncDefaultEnvVarsWithProtocol(modelSvcWithSchema, protocol.HttpJson),
									).ToKubernetesEnvVars(),
									Resources:     expDefaultModelResourceRequests,
									LivenessProbe: probeConfig,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
		{
			name: "pyfunc spec with liveness probe disabled (new)",
			modelSvc: &models.Service{
				Name:         modelSvc.Name,
				ModelName:    modelSvc.ModelName,
				ModelVersion: modelSvc.ModelVersion,
				Namespace:    project.Name,
				ArtifactURI:  modelSvc.ArtifactURI,
				Type:         models.ModelTypePyFunc,
				Options: &models.ModelOption{
					PyFuncImageName: "gojek/project-model:1",
				},
				EnvVars:  models.EnvVars{models.EnvVar{Name: envDisableLivenessProbe, Value: "true"}},
				Metadata: modelSvc.Metadata,
				Protocol: protocol.HttpJson,
			},
			resourcePercentage: queueResourcePercentage,
			deploymentScale:    defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						"prometheus.io/scrape":                                "true",
						"prometheus.io/port":                                  "8080",
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testPredictorScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						PodSpec: kservev1beta1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  kserveconstant.InferenceServiceContainerName,
									Image: "gojek/project-model:1",
									Env: models.MergeEnvVars(
										models.EnvVars{
											{Name: defaultWorkersEnvVarName, Value: defaultWorkersEnvVarValue},
											{Name: envDisableLivenessProbe, Value: "true"},
										},
										createPyFuncDefaultEnvVarsWithProtocol(modelSvc, protocol.HttpJson)).ToKubernetesEnvVars(),
									Resources: expDefaultModelResourceRequests,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
		{
			name: "tensorflow spec with user resource request",
			modelSvc: &models.Service{
				Name:            modelSvc.Name,
				Namespace:       project.Name,
				ArtifactURI:     modelSvc.ArtifactURI,
				Type:            models.ModelTypeTensorflow,
				Options:         &models.ModelOption{},
				Metadata:        modelSvc.Metadata,
				ResourceRequest: userResourceRequests,
				Protocol:        protocol.HttpJson,
			},
			resourcePercentage: queueResourcePercentage,
			deploymentScale:    defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testPredictorScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						Tensorflow: &kservev1beta1.TFServingSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     expUserResourceRequests,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{defaultEnvVarWithoutCPULimits},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &userResourceRequests.MinReplica,
							MaxReplicas: userResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
		{
			name: "custom spec with default resource request",
			modelSvc: &models.Service{
				Name:         modelSvc.Name,
				ModelName:    modelSvc.ModelName,
				ModelVersion: modelSvc.ModelVersion,
				Namespace:    project.Name,
				ArtifactURI:  modelSvc.ArtifactURI,
				Type:         models.ModelTypeCustom,
				Options: &models.ModelOption{
					CustomPredictor: &models.CustomPredictor{
						Image: "gcr.io/custom-model:v0.1",
					},
				},
				// Env var below will be overwritten by default values to prevent user overwrite
				EnvVars: models.EnvVars{
					models.EnvVar{Name: "MERLIN_PREDICTOR_PORT", Value: "1234"},
					models.EnvVar{Name: "MERLIN_MODEL_NAME", Value: "rubbish-model"},
					models.EnvVar{Name: "MERLIN_ARTIFACT_LOCATION", Value: "/mnt/models/wrong-path"},
					models.EnvVar{Name: "STORAGE_URI", Value: "invalid_uri"},
				},
				Metadata: modelSvc.Metadata,
				Protocol: protocol.HttpJson,
			},
			resourcePercentage: queueResourcePercentage,
			deploymentScale:    defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testPredictorScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						PodSpec: kservev1beta1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  kserveconstant.InferenceServiceContainerName,
									Image: "gcr.io/custom-model:v0.1",
									Env: models.MergeEnvVars(
										models.EnvVars{
											{Name: defaultWorkersEnvVarName, Value: defaultWorkersEnvVarValue},
										},
										createDefaultPredictorEnvVars(modelSvc),
									).ToKubernetesEnvVars(),
									Resources: expDefaultModelResourceRequests,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
		{
			name: "custom spec with resource request",
			modelSvc: &models.Service{
				Name:         modelSvc.Name,
				ModelName:    modelSvc.ModelName,
				ModelVersion: modelSvc.ModelVersion,
				Namespace:    project.Name,
				ArtifactURI:  modelSvc.ArtifactURI,
				Type:         models.ModelTypeCustom,
				Options: &models.ModelOption{
					CustomPredictor: &models.CustomPredictor{
						Image:   "gcr.io/custom-model:v0.1",
						Command: "./run.sh",
						Args:    "firstArg secondArg",
					},
				},
				Metadata:        modelSvc.Metadata,
				ResourceRequest: userResourceRequests,
				Protocol:        protocol.HttpJson,
			},
			resourcePercentage: queueResourcePercentage,
			deploymentScale:    defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testPredictorScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						PodSpec: kservev1beta1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  kserveconstant.InferenceServiceContainerName,
									Image: "gcr.io/custom-model:v0.1",
									Env: models.MergeEnvVars(
										models.EnvVars{
											{Name: defaultWorkersEnvVarName, Value: defaultWorkersEnvVarValue},
										},
										createDefaultPredictorEnvVars(modelSvc),
									).ToKubernetesEnvVars(),
									Resources: expUserResourceRequests,
									Command: []string{
										"./run.sh",
									},
									Args: []string{
										"firstArg",
										"secondArg",
									},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &userResourceRequests.MinReplica,
							MaxReplicas: userResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
		{
			name: "raw deployment using CPU autoscaling",
			modelSvc: &models.Service{
				Name:           modelSvc.Name,
				ModelName:      modelSvc.ModelName,
				ModelVersion:   modelSvc.ModelVersion,
				Namespace:      project.Name,
				ArtifactURI:    modelSvc.ArtifactURI,
				Type:           models.ModelTypeTensorflow,
				Options:        &models.ModelOption{},
				Metadata:       modelSvc.Metadata,
				DeploymentMode: deployment.RawDeploymentMode,
				AutoscalingPolicy: &autoscaling.AutoscalingPolicy{
					MetricsType: autoscaling.CPUUtilization,
					TargetValue: 30,
				},
				Protocol: protocol.HttpJson,
			},
			resourcePercentage: queueResourcePercentage,
			deploymentScale:    defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.RawDeployment),
						kserveconstant.AutoscalerClass:                        string(kserveconstant.AutoscalerClassHPA),
						kserveconstant.AutoscalerMetrics:                      string(kserveconstant.AutoScalerMetricsCPU),
						kserveconstant.TargetUtilizationPercentage:            "30",
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						Tensorflow: &kservev1beta1.TFServingSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     expDefaultModelResourceRequests,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{defaultEnvVarWithoutCPULimits},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
		{
			name: "raw deployment using not supported autoscaling metrics",
			modelSvc: &models.Service{
				Name:           modelSvc.Name,
				ModelName:      modelSvc.ModelName,
				ModelVersion:   modelSvc.ModelVersion,
				Namespace:      project.Name,
				ArtifactURI:    modelSvc.ArtifactURI,
				Type:           models.ModelTypeTensorflow,
				Options:        &models.ModelOption{},
				Metadata:       modelSvc.Metadata,
				DeploymentMode: deployment.RawDeploymentMode,
				AutoscalingPolicy: &autoscaling.AutoscalingPolicy{
					MetricsType: autoscaling.MemoryUtilization,
					TargetValue: 30,
				},
				Protocol: protocol.HttpJson,
			},
			resourcePercentage: queueResourcePercentage,
			wantErr:            true,
		},
		{
			name: "serverless deployment using CPU autoscaling",
			modelSvc: &models.Service{
				Name:           modelSvc.Name,
				ModelName:      modelSvc.ModelName,
				ModelVersion:   modelSvc.ModelVersion,
				Namespace:      project.Name,
				ArtifactURI:    modelSvc.ArtifactURI,
				Type:           models.ModelTypeTensorflow,
				Options:        &models.ModelOption{},
				Metadata:       modelSvc.Metadata,
				DeploymentMode: deployment.ServerlessDeploymentMode,
				AutoscalingPolicy: &autoscaling.AutoscalingPolicy{
					MetricsType: autoscaling.CPUUtilization,
					TargetValue: 30,
				},
				Protocol: protocol.HttpJson,
			},
			resourcePercentage: queueResourcePercentage,
			deploymentScale:    defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.ClassAnnotationKey:                      knautoscaling.HPA,
						knautoscaling.MetricAnnotationKey:                     knautoscaling.CPU,
						knautoscaling.TargetAnnotationKey:                     "30",
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testPredictorScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						Tensorflow: &kservev1beta1.TFServingSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     expDefaultModelResourceRequests,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{defaultEnvVarWithoutCPULimits},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
		{
			name: "serverless deployment using memory autoscaling",
			modelSvc: &models.Service{
				Name:           modelSvc.Name,
				ModelName:      modelSvc.ModelName,
				ModelVersion:   modelSvc.ModelVersion,
				Namespace:      project.Name,
				ArtifactURI:    modelSvc.ArtifactURI,
				Type:           models.ModelTypeTensorflow,
				Options:        &models.ModelOption{},
				Metadata:       modelSvc.Metadata,
				DeploymentMode: deployment.ServerlessDeploymentMode,
				AutoscalingPolicy: &autoscaling.AutoscalingPolicy{
					MetricsType: autoscaling.MemoryUtilization,
					TargetValue: 30,
				},
				Protocol: protocol.HttpJson,
			},
			resourcePercentage: queueResourcePercentage,
			deploymentScale:    defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.ClassAnnotationKey:                      knautoscaling.HPA,
						knautoscaling.MetricAnnotationKey:                     knautoscaling.Memory,
						knautoscaling.TargetAnnotationKey:                     "150", // 30% * default memory request (500Mi) = 150Mi
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testPredictorScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						Tensorflow: &kservev1beta1.TFServingSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     expDefaultModelResourceRequests,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{defaultEnvVarWithoutCPULimits},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
		{
			name: "serverless deployment using memory autoscaling when memory request is specified",
			modelSvc: &models.Service{
				Name:           modelSvc.Name,
				ModelName:      modelSvc.ModelName,
				ModelVersion:   modelSvc.ModelVersion,
				Namespace:      project.Name,
				ArtifactURI:    modelSvc.ArtifactURI,
				Type:           models.ModelTypeTensorflow,
				Options:        &models.ModelOption{},
				Metadata:       modelSvc.Metadata,
				DeploymentMode: deployment.ServerlessDeploymentMode,
				AutoscalingPolicy: &autoscaling.AutoscalingPolicy{
					MetricsType: autoscaling.MemoryUtilization,
					TargetValue: 20,
				},
				Protocol:        protocol.HttpJson,
				ResourceRequest: userResourceRequests,
			},
			resourcePercentage: queueResourcePercentage,
			deploymentScale:    defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.ClassAnnotationKey:                      knautoscaling.HPA,
						knautoscaling.MetricAnnotationKey:                     knautoscaling.Memory,
						knautoscaling.TargetAnnotationKey:                     "205", // 20% * (1Gi) ~= 205Mi
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testPredictorScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						Tensorflow: &kservev1beta1.TFServingSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     expUserResourceRequests,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{defaultEnvVarWithoutCPULimits},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &userResourceRequests.MinReplica,
							MaxReplicas: userResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
		{
			name: "serverless deployment using concurrency autoscaling",
			modelSvc: &models.Service{
				Name:           modelSvc.Name,
				ModelName:      modelSvc.ModelName,
				ModelVersion:   modelSvc.ModelVersion,
				Namespace:      project.Name,
				ArtifactURI:    modelSvc.ArtifactURI,
				Type:           models.ModelTypeTensorflow,
				Options:        &models.ModelOption{},
				Metadata:       modelSvc.Metadata,
				DeploymentMode: deployment.ServerlessDeploymentMode,
				AutoscalingPolicy: &autoscaling.AutoscalingPolicy{
					MetricsType: autoscaling.Concurrency,
					TargetValue: 2,
				},
				Protocol: protocol.HttpJson,
			},
			resourcePercentage: queueResourcePercentage,
			deploymentScale:    defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.ClassAnnotationKey:                      knautoscaling.KPA,
						knautoscaling.MetricAnnotationKey:                     knautoscaling.Concurrency,
						knautoscaling.TargetAnnotationKey:                     "2.00",
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testPredictorScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						Tensorflow: &kservev1beta1.TFServingSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     expDefaultModelResourceRequests,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{defaultEnvVarWithoutCPULimits},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
		{
			name: "serverless deployment using concurrency autoscaling where target after rounding < 0.01",
			modelSvc: &models.Service{
				Name:           modelSvc.Name,
				ModelName:      modelSvc.ModelName,
				ModelVersion:   modelSvc.ModelVersion,
				Namespace:      project.Name,
				ArtifactURI:    modelSvc.ArtifactURI,
				Type:           models.ModelTypeTensorflow,
				Options:        &models.ModelOption{},
				Metadata:       modelSvc.Metadata,
				DeploymentMode: deployment.ServerlessDeploymentMode,
				AutoscalingPolicy: &autoscaling.AutoscalingPolicy{
					MetricsType: autoscaling.Concurrency,
					TargetValue: 0.004,
				},
				Protocol: protocol.HttpJson,
			},
			resourcePercentage: queueResourcePercentage,
			deploymentScale:    defaultDeploymentScale,
			wantErr:            true,
		},
		{
			name: "serverless deployment using rps autoscaling",
			modelSvc: &models.Service{
				Name:           modelSvc.Name,
				ModelName:      modelSvc.ModelName,
				ModelVersion:   modelSvc.ModelVersion,
				Namespace:      project.Name,
				ArtifactURI:    modelSvc.ArtifactURI,
				Type:           models.ModelTypeTensorflow,
				Options:        &models.ModelOption{},
				Metadata:       modelSvc.Metadata,
				DeploymentMode: deployment.ServerlessDeploymentMode,
				AutoscalingPolicy: &autoscaling.AutoscalingPolicy{
					MetricsType: autoscaling.RPS,
					TargetValue: 10,
				},
				Protocol: protocol.HttpJson,
			},
			resourcePercentage: queueResourcePercentage,
			deploymentScale:    defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.ClassAnnotationKey:                      knautoscaling.KPA,
						knautoscaling.MetricAnnotationKey:                     knautoscaling.RPS,
						knautoscaling.TargetAnnotationKey:                     "10",
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testPredictorScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						Tensorflow: &kservev1beta1.TFServingSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     expDefaultModelResourceRequests,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{defaultEnvVarWithoutCPULimits},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
		{
			name: "tensorflow upi v1",
			modelSvc: &models.Service{
				Name:         modelSvc.Name,
				ModelName:    modelSvc.ModelName,
				ModelVersion: modelSvc.ModelVersion,
				Namespace:    project.Name,
				ArtifactURI:  modelSvc.ArtifactURI,
				Type:         models.ModelTypeTensorflow,
				Options:      &models.ModelOption{},
				Metadata:     modelSvc.Metadata,
				Protocol:     protocol.UpiV1,
			},
			resourcePercentage: queueResourcePercentage,
			deploymentScale:    defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testPredictorScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						Tensorflow: &kservev1beta1.TFServingSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     expDefaultModelResourceRequests,
									Ports:         grpcServerlessContainerPorts,
									Env:           []corev1.EnvVar{defaultEnvVarWithoutCPULimits},
									LivenessProbe: probeConfigUPI,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
		{
			name: "pyfunc upi v1",
			modelSvc: &models.Service{
				Name:         modelSvc.Name,
				ModelName:    modelSvc.ModelName,
				ModelVersion: modelSvc.ModelVersion,
				Namespace:    project.Name,
				ArtifactURI:  modelSvc.ArtifactURI,
				Type:         models.ModelTypePyFunc,
				Options: &models.ModelOption{
					PyFuncImageName: "gojek/project-model:1",
				},
				Metadata: modelSvc.Metadata,
				Protocol: protocol.UpiV1,
			},
			resourcePercentage: queueResourcePercentage,
			deploymentScale:    defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						"prometheus.io/scrape":                                "true",
						"prometheus.io/port":                                  "8080",
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testPredictorScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						PodSpec: kservev1beta1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:      kserveconstant.InferenceServiceContainerName,
									Image:     "gojek/project-model:1",
									Resources: expDefaultModelResourceRequests,
									Ports:     grpcServerlessContainerPorts,
									Env: models.MergeEnvVars(
										models.EnvVars{
											{Name: defaultWorkersEnvVarName, Value: defaultWorkersEnvVarValue},
											{Name: envGRPCOptions, Value: "{}"},
										},
										createPyFuncDefaultEnvVarsWithProtocol(modelSvc, protocol.UpiV1),
									).ToKubernetesEnvVars(),
									LivenessProbe: probeConfigUPI,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
		{
			name: "xgboost upi v1",
			modelSvc: &models.Service{
				Name:         modelSvc.Name,
				ModelName:    modelSvc.ModelName,
				ModelVersion: modelSvc.ModelVersion,
				Namespace:    project.Name,
				ArtifactURI:  modelSvc.ArtifactURI,
				Type:         models.ModelTypeXgboost,
				Options:      &models.ModelOption{},
				Metadata:     modelSvc.Metadata,
				Protocol:     protocol.UpiV1,
			},
			resourcePercentage: queueResourcePercentage,
			deploymentScale:    defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testPredictorScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						XGBoost: &kservev1beta1.XGBoostSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     expDefaultModelResourceRequests,
									Ports:         grpcServerlessContainerPorts,
									Env:           []corev1.EnvVar{defaultEnvVarWithoutCPULimits},
									LivenessProbe: probeConfigUPI,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
		{
			name: "custom modelSvc upi v1",
			modelSvc: &models.Service{
				Name:         modelSvc.Name,
				ModelName:    modelSvc.ModelName,
				ModelVersion: modelSvc.ModelVersion,
				Namespace:    project.Name,
				ArtifactURI:  modelSvc.ArtifactURI,
				Type:         models.ModelTypeCustom,
				Options: &models.ModelOption{
					CustomPredictor: &models.CustomPredictor{
						Image: "gcr.io/custom-model:v0.1",
					},
				},
				// Env var below will be overwritten by default values to prevent user overwrite
				EnvVars: models.EnvVars{
					models.EnvVar{Name: "MERLIN_PREDICTOR_PORT", Value: "1234"},
					models.EnvVar{Name: "MERLIN_MODEL_NAME", Value: "rubbish-model"},
					models.EnvVar{Name: "MERLIN_ARTIFACT_LOCATION", Value: "/mnt/models/wrong-path"},
					models.EnvVar{Name: "STORAGE_URI", Value: "invalid_uri"},
				},
				Metadata: modelSvc.Metadata,
				Protocol: protocol.UpiV1,
			},
			resourcePercentage: queueResourcePercentage,
			deploymentScale:    defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testPredictorScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						PodSpec: kservev1beta1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  kserveconstant.InferenceServiceContainerName,
									Image: "gcr.io/custom-model:v0.1",
									Env: models.MergeEnvVars(
										models.EnvVars{
											{Name: defaultWorkersEnvVarName, Value: defaultWorkersEnvVarValue},
										},
										createDefaultPredictorEnvVars(modelSvc),
									).ToKubernetesEnvVars(),
									Resources: expDefaultModelResourceRequests,
									Ports:     grpcServerlessContainerPorts,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deployConfig := &config.DeploymentConfig{
				DefaultModelResourceRequests:          defaultModelResourceRequests,
				DefaultTransformerResourceRequests:    defaultTransformerResourceRequests,
				QueueResourcePercentage:               tt.resourcePercentage,
				PyfuncGRPCOptions:                     "{}",
				StandardTransformer:                   standardTransformerConfig,
				PyFuncPublisher:                       pyfuncPublisherConfig,
				UserContainerCPUDefaultLimit:          userContainerCPUDefaultLimit,
				UserContainerCPULimitRequestFactor:    userContainerCPULimitRequestFactor,
				UserContainerMemoryLimitRequestFactor: userContainerMemoryLimitRequestFactor,
				DefaultEnvVarsWithoutCPULimits:        []corev1.EnvVar{defaultEnvVarWithoutCPULimits},
			}

			tpl := NewInferenceServiceTemplater(*deployConfig)
			infSvcSpec, err := tpl.CreateInferenceServiceSpec(tt.modelSvc, tt.deploymentScale)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// Sort all env vars in both expected and actual inference service specs before comparing
			sortInferenceServiceSpecEnvVars(tt.exp.Spec)
			sortInferenceServiceSpecEnvVars(infSvcSpec.Spec)
			assert.Equal(t, tt.exp, infSvcSpec)
		})
	}
}

func TestCreateInferenceServiceSpecWithTransformer(t *testing.T) {
	err := labeller.InitKubernetesLabeller("gojek.com/", "caraml.dev/", testEnvironmentName)
	assert.NoError(t, err)

	defer func() {
		_ = labeller.InitKubernetesLabeller("", "", "")
	}()

	project := mlp.Project{
		Name: "project",
	}

	loggerDestinationURL := "http://destination.default"
	modelSvc := &models.Service{
		Name:         "model-1",
		ModelName:    "model",
		ModelVersion: "1",
		Namespace:    "project",
		ArtifactURI:  "gs://my-artifacet",
		Metadata: models.Metadata{
			App:       "model",
			Component: models.ComponentModelVersion,
			Stream:    "dsp",
			Team:      "dsp",
			Labels: mlp.Labels{
				{
					Key:   "sample",
					Value: "true",
				},
			},
		},
		Protocol: protocol.HttpJson,
	}

	modelSvcGRPC := &models.Service{
		Name:         "model-1",
		ModelName:    "model",
		ModelVersion: "1",
		Namespace:    "project",
		ArtifactURI:  "gs://my-artifacet",
		Metadata: models.Metadata{
			App:       "model",
			Component: models.ComponentModelVersion,
			Stream:    "dsp",
			Team:      "dsp",
			Labels: mlp.Labels{
				{
					Key:   "sample",
					Value: "true",
				},
			},
		},
		Protocol: protocol.UpiV1,
	}

	queueResourcePercentage := "2"
	storageUri := fmt.Sprintf("%s/model", modelSvc.ArtifactURI)

	// Liveness probe config for the model containers
	probeConfig := createLivenessProbeSpec(protocol.HttpJson, fmt.Sprintf("/v1/models/%s", modelSvc.Name))
	probeConfigUPI := createLivenessProbeSpec(protocol.UpiV1, fmt.Sprintf("/v1/models/%s", modelSvc.Name))

	// Liveness probe config for the transformers
	transformerProbeConfig := createLivenessProbeSpec(protocol.HttpJson, "/")
	transformerProbeConfigUPI := createLivenessProbeSpec(protocol.UpiV1, "/")
	tests := []struct {
		name            string
		modelSvc        *models.Service
		deploymentScale DeploymentScale
		exp             *kservev1beta1.InferenceService
		wantErr         bool
	}{
		{
			name: "custom transformer with default resource request",
			modelSvc: &models.Service{
				Name:         modelSvc.Name,
				ModelName:    modelSvc.ModelName,
				ModelVersion: modelSvc.ModelVersion,
				Namespace:    project.Name,
				ArtifactURI:  modelSvc.ArtifactURI,
				Type:         models.ModelTypeTensorflow,
				Options:      &models.ModelOption{},
				Metadata:     modelSvc.Metadata,
				Transformer: &models.Transformer{
					Enabled: true,
					Image:   "ghcr.io/gojek/merlin-transformer-test",
					Command: "python",
					Args:    "main.py",
					EnvVars: models.EnvVars{
						{Name: envTransformerPort, Value: "1234"},                              // should be replace by default
						{Name: envTransformerModelName, Value: "model-1234"},                   // should be replace by default
						{Name: envTransformerPredictURL, Value: "model-112-predictor.project"}, // should be replace by default
					},
				},
				Protocol: protocol.HttpJson,
			},
			deploymentScale: defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testTransformerScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						Tensorflow: &kservev1beta1.TFServingSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     expDefaultModelResourceRequests,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{defaultEnvVarWithoutCPULimits},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
					},
					Transformer: &kservev1beta1.TransformerSpec{
						PodSpec: kservev1beta1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:    "transformer",
									Image:   "ghcr.io/gojek/merlin-transformer-test",
									Command: []string{"python"},
									Args:    []string{"main.py"},
									Env: MergeEnvVars(
										[]corev1.EnvVar{
											{Name: defaultWorkersEnvVarName, Value: defaultWorkersEnvVarValue},
										},
										createDefaultTransformerEnvVars(modelSvc),
									),
									Resources:     expDefaultTransformerResourceRequests,
									LivenessProbe: transformerProbeConfig,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultTransformerResourceRequests.MinReplica,
							MaxReplicas: defaultTransformerResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
		{
			name: "custom transformer with user resource request",
			modelSvc: &models.Service{
				Name:         modelSvc.Name,
				ModelName:    modelSvc.ModelName,
				ModelVersion: modelSvc.ModelVersion,
				Namespace:    project.Name,
				ArtifactURI:  modelSvc.ArtifactURI,
				Type:         models.ModelTypeTensorflow,
				Options:      &models.ModelOption{},
				Metadata:     modelSvc.Metadata,
				Transformer: &models.Transformer{
					Enabled:         true,
					Image:           "ghcr.io/gojek/merlin-transformer-test",
					Command:         "python",
					Args:            "main.py",
					ResourceRequest: userResourceRequests,
				},
				Logger: &models.Logger{
					DestinationURL: loggerDestinationURL,
					Transformer: &models.LoggerConfig{
						Enabled: false,
						Mode:    models.LogRequest,
					},
				},
				Protocol: protocol.HttpJson,
			},
			deploymentScale: defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testTransformerScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						Tensorflow: &kservev1beta1.TFServingSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     expDefaultModelResourceRequests,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{defaultEnvVarWithoutCPULimits},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
					},
					Transformer: &kservev1beta1.TransformerSpec{
						PodSpec: kservev1beta1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:    "transformer",
									Image:   "ghcr.io/gojek/merlin-transformer-test",
									Command: []string{"python"},
									Args:    []string{"main.py"},
									Env: MergeEnvVars(
										[]corev1.EnvVar{
											{Name: defaultWorkersEnvVarName, Value: defaultWorkersEnvVarValue},
										},
										createDefaultTransformerEnvVars(modelSvc),
									),
									Resources:     expUserResourceRequests,
									LivenessProbe: transformerProbeConfig,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &userResourceRequests.MinReplica,
							MaxReplicas: userResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
		{
			name: "custom transformer with cpu limits specified",
			modelSvc: &models.Service{
				Name:         modelSvc.Name,
				ModelName:    modelSvc.ModelName,
				ModelVersion: modelSvc.ModelVersion,
				Namespace:    project.Name,
				ArtifactURI:  modelSvc.ArtifactURI,
				Type:         models.ModelTypeTensorflow,
				Options:      &models.ModelOption{},
				Metadata:     modelSvc.Metadata,
				Transformer: &models.Transformer{
					Enabled:         true,
					Image:           "ghcr.io/gojek/merlin-transformer-test",
					Command:         "python",
					Args:            "main.py",
					ResourceRequest: userResourceRequestsWithCPULimits,
				},
				Logger: &models.Logger{
					DestinationURL: loggerDestinationURL,
					Transformer: &models.LoggerConfig{
						Enabled: false,
						Mode:    models.LogRequest,
					},
				},
				Protocol: protocol.HttpJson,
			},
			deploymentScale: defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testTransformerScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						Tensorflow: &kservev1beta1.TFServingSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     expDefaultModelResourceRequests,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{defaultEnvVarWithoutCPULimits},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
					},
					Transformer: &kservev1beta1.TransformerSpec{
						PodSpec: kservev1beta1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:          "transformer",
									Image:         "ghcr.io/gojek/merlin-transformer-test",
									Command:       []string{"python"},
									Args:          []string{"main.py"},
									Env:           createDefaultTransformerEnvVars(modelSvc),
									Resources:     expUserResourceRequestsWithCPULimits,
									LivenessProbe: transformerProbeConfig,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &userResourceRequests.MinReplica,
							MaxReplicas: userResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
		{
			name: "custom transformer upi v1",
			modelSvc: &models.Service{
				Name:         modelSvc.Name,
				ModelName:    modelSvc.ModelName,
				ModelVersion: modelSvc.ModelVersion,
				Namespace:    project.Name,
				ArtifactURI:  modelSvc.ArtifactURI,
				Type:         models.ModelTypeTensorflow,
				Options:      &models.ModelOption{},
				Metadata:     modelSvc.Metadata,
				Transformer: &models.Transformer{
					Enabled: true,
					Image:   "ghcr.io/gojek/merlin-transformer-test",
					Command: "python",
					Args:    "main.py",
					EnvVars: models.EnvVars{
						{Name: envTransformerPort, Value: "1234"},                              // should be replace by default
						{Name: envTransformerModelName, Value: "model-1234"},                   // should be replace by default
						{Name: envTransformerPredictURL, Value: "model-112-predictor.project"}, // should be replace by default
					},
				},
				Protocol: protocol.UpiV1,
			},
			deploymentScale: defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testTransformerScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						Tensorflow: &kservev1beta1.TFServingSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     expDefaultModelResourceRequests,
									Env:           []corev1.EnvVar{defaultEnvVarWithoutCPULimits},
									Ports:         grpcServerlessContainerPorts,
									LivenessProbe: probeConfigUPI,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
					},
					Transformer: &kservev1beta1.TransformerSpec{
						PodSpec: kservev1beta1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:    "transformer",
									Image:   "ghcr.io/gojek/merlin-transformer-test",
									Command: []string{"python"},
									Args:    []string{"main.py"},
									Env: MergeEnvVars(
										[]corev1.EnvVar{
											{Name: defaultWorkersEnvVarName, Value: defaultWorkersEnvVarValue},
										},
										createDefaultTransformerEnvVars(modelSvcGRPC),
									),
									Resources:     expDefaultTransformerResourceRequests,
									LivenessProbe: transformerProbeConfigUPI,
									Ports:         grpcServerlessContainerPorts,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultTransformerResourceRequests.MinReplica,
							MaxReplicas: defaultTransformerResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
		{
			name: "standard transformer",
			modelSvc: &models.Service{
				Name:         modelSvc.Name,
				ModelName:    modelSvc.ModelName,
				ModelVersion: modelSvc.ModelVersion,
				Namespace:    project.Name,
				ArtifactURI:  modelSvc.ArtifactURI,
				Type:         models.ModelTypeTensorflow,
				Options:      &models.ModelOption{},
				Metadata:     modelSvc.Metadata,
				Transformer: &models.Transformer{
					Enabled:         true,
					TransformerType: models.StandardTransformerType,
					EnvVars: models.EnvVars{
						{
							Name:  transformerpkg.StandardTransformerConfigEnvName,
							Value: `{"standard_transformer": null}`,
						},
					},
				},
				Logger: &models.Logger{
					DestinationURL: loggerDestinationURL,
					Transformer: &models.LoggerConfig{
						Enabled: false,
						Mode:    models.LogRequest,
					},
				},
				Protocol: protocol.HttpJson,
			},
			deploymentScale: defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testTransformerScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						Tensorflow: &kservev1beta1.TFServingSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     expDefaultModelResourceRequests,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{defaultEnvVarWithoutCPULimits},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
					},
					Transformer: &kservev1beta1.TransformerSpec{
						PodSpec: kservev1beta1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "transformer",
									Image: standardTransformerConfig.ImageName,
									Env: MergeEnvVars([]corev1.EnvVar{
										{
											Name:  defaultWorkersEnvVarName,
											Value: defaultWorkersEnvVarValue,
										},
										{
											Name:  transformerpkg.DefaultFeastSource,
											Value: standardTransformerConfig.DefaultFeastSource.String(),
										},
										{
											Name:  transformerpkg.FeastStorageConfigs,
											Value: `{"1":{"redisCluster":{"feastServingUrl":"localhost:6866","redisAddress":["10.1.1.2","10.1.1.3"],"option":{"poolSize":5,"minIdleConnections":2}}},"2":{"bigtable":{"feastServingUrl":"localhost:6867","project":"gcp-project","instance":"instance","appProfile":"default","option":{"grpcConnectionPool":4,"keepAliveInterval":"120s","keepAliveTimeout":"60s","credentialJson":"eyJrZXkiOiJ2YWx1ZSJ9"}}}}`,
										},
										{Name: transformerpkg.FeastServingKeepAliveEnabled, Value: "true"},
										{Name: transformerpkg.FeastServingKeepAliveTime, Value: "30s"},
										{Name: transformerpkg.FeastServingKeepAliveTimeout, Value: "1s"},
										{Name: transformerpkg.FeastGRPCConnCount, Value: "5"},
										{Name: transformerpkg.JaegerCollectorURL, Value: standardTransformerConfig.Jaeger.CollectorURL},
										{Name: transformerpkg.JaegerSamplerParam, Value: standardTransformerConfig.Jaeger.SamplerParam},
										{Name: transformerpkg.JaegerDisabled, Value: standardTransformerConfig.Jaeger.Disabled},
										{Name: transformerpkg.StandardTransformerConfigEnvName, Value: `{"standard_transformer":null}`},
									}, createDefaultTransformerEnvVars(modelSvc)),
									Resources:     expDefaultTransformerResourceRequests,
									LivenessProbe: transformerProbeConfig,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultTransformerResourceRequests.MinReplica,
							MaxReplicas: defaultTransformerResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
		{
			name: "standard transformer upi v1 raw",
			modelSvc: &models.Service{
				Name:         modelSvc.Name,
				ModelName:    modelSvc.ModelName,
				ModelVersion: modelSvc.ModelVersion,
				Namespace:    project.Name,
				ArtifactURI:  modelSvc.ArtifactURI,
				Type:         models.ModelTypeTensorflow,
				Options:      &models.ModelOption{},
				Metadata:     modelSvc.Metadata,
				Transformer: &models.Transformer{
					Enabled:         true,
					TransformerType: models.StandardTransformerType,
					EnvVars: models.EnvVars{
						{
							Name:  transformerpkg.StandardTransformerConfigEnvName,
							Value: `{"standard_transformer": null}`,
						},
					},
				},
				Logger: &models.Logger{
					DestinationURL: loggerDestinationURL,
					Transformer: &models.LoggerConfig{
						Enabled: false,
						Mode:    models.LogRequest,
					},
				},
				Protocol:       protocol.UpiV1,
				DeploymentMode: deployment.RawDeploymentMode,
			},
			deploymentScale: defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.RawDeployment),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						Tensorflow: &kservev1beta1.TFServingSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     expDefaultModelResourceRequests,
									Env:           []corev1.EnvVar{defaultEnvVarWithoutCPULimits},
									LivenessProbe: probeConfigUPI,
									Ports:         grpcRawContainerPorts,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
					},
					Transformer: &kservev1beta1.TransformerSpec{
						PodSpec: kservev1beta1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "transformer",
									Image: standardTransformerConfig.ImageName,
									Env: MergeEnvVars([]corev1.EnvVar{
										{
											Name:  defaultWorkersEnvVarName,
											Value: defaultWorkersEnvVarValue,
										},
										{
											Name:  transformerpkg.DefaultFeastSource,
											Value: standardTransformerConfig.DefaultFeastSource.String(),
										},
										{
											Name:  transformerpkg.FeastStorageConfigs,
											Value: `{"1":{"redisCluster":{"feastServingUrl":"localhost:6866","redisAddress":["10.1.1.2","10.1.1.3"],"option":{"poolSize":5,"minIdleConnections":2}}},"2":{"bigtable":{"feastServingUrl":"localhost:6867","project":"gcp-project","instance":"instance","appProfile":"default","option":{"grpcConnectionPool":4,"keepAliveInterval":"120s","keepAliveTimeout":"60s","credentialJson":"eyJrZXkiOiJ2YWx1ZSJ9"}}}}`,
										},
										{Name: transformerpkg.FeastServingKeepAliveEnabled, Value: "true"},
										{Name: transformerpkg.FeastServingKeepAliveTime, Value: "30s"},
										{Name: transformerpkg.FeastServingKeepAliveTimeout, Value: "1s"},
										{Name: transformerpkg.FeastGRPCConnCount, Value: "5"},
										{Name: transformerpkg.KafkaTopic, Value: "caraml-project-model-prediction-log"},
										{Name: transformerpkg.KafkaBrokers, Value: standardTransformerConfig.Kafka.Brokers},
										{Name: transformerpkg.KafkaMaxMessageSizeBytes, Value: fmt.Sprintf("%v", standardTransformerConfig.Kafka.MaxMessageSizeBytes)},
										{Name: transformerpkg.KafkaConnectTimeoutMS, Value: fmt.Sprintf("%v", standardTransformerConfig.Kafka.ConnectTimeoutMS)},
										{Name: transformerpkg.KafkaSerialization, Value: string(standardTransformerConfig.Kafka.SerializationFmt)},
										{Name: transformerpkg.ModelServerConnCount, Value: "3"},
										{Name: transformerpkg.JaegerCollectorURL, Value: standardTransformerConfig.Jaeger.CollectorURL},
										{Name: transformerpkg.JaegerSamplerParam, Value: standardTransformerConfig.Jaeger.SamplerParam},
										{Name: transformerpkg.JaegerDisabled, Value: standardTransformerConfig.Jaeger.Disabled},
										{Name: transformerpkg.StandardTransformerConfigEnvName, Value: `{"standard_transformer":null}`},
									}, createDefaultTransformerEnvVars(modelSvcGRPC)),
									Resources:     expDefaultTransformerResourceRequests,
									Ports:         grpcRawContainerPorts,
									LivenessProbe: transformerProbeConfigUPI,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultTransformerResourceRequests.MinReplica,
							MaxReplicas: defaultTransformerResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
		{
			name: "standard transformer upi v1; pyfunc raw",
			modelSvc: &models.Service{
				Name:         modelSvc.Name,
				ModelName:    modelSvc.ModelName,
				ModelVersion: modelSvc.ModelVersion,
				Namespace:    project.Name,
				ArtifactURI:  modelSvc.ArtifactURI,
				Type:         models.ModelTypePyFunc,
				Options: &models.ModelOption{
					PyFuncImageName: "gojek/project-model:1",
				},
				Metadata: modelSvc.Metadata,
				Transformer: &models.Transformer{
					Enabled:         true,
					TransformerType: models.StandardTransformerType,
					EnvVars: models.EnvVars{
						{
							Name:  transformerpkg.StandardTransformerConfigEnvName,
							Value: `{"standard_transformer": null}`,
						},
					},
				},
				Logger: &models.Logger{
					DestinationURL: loggerDestinationURL,
					Transformer: &models.LoggerConfig{
						Enabled: false,
						Mode:    models.LogRequest,
					},
				},
				Protocol:       protocol.UpiV1,
				DeploymentMode: deployment.RawDeploymentMode,
			},
			deploymentScale: defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.RawDeployment),
						annotationPrometheusScrapeFlag:                        "true",
						annotationPrometheusScrapePort:                        "8080",
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						PodSpec: kservev1beta1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  kserveconstant.InferenceServiceContainerName,
									Image: "gojek/project-model:1",
									Env: models.MergeEnvVars(
										models.EnvVars{
											{
												Name:  defaultWorkersEnvVarName,
												Value: defaultWorkersEnvVarValue,
											},
											{
												Name:  envGRPCOptions,
												Value: "{}",
											},
										},
										createPyFuncDefaultEnvVarsWithProtocol(modelSvc, protocol.UpiV1),
									).ToKubernetesEnvVars(),
									Resources:     expDefaultModelResourceRequests,
									LivenessProbe: probeConfigUPI,
									Ports:         grpcRawContainerPorts,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
					},
					Transformer: &kservev1beta1.TransformerSpec{
						PodSpec: kservev1beta1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "transformer",
									Image: standardTransformerConfig.ImageName,
									Env: MergeEnvVars([]corev1.EnvVar{
										{
											Name:  defaultWorkersEnvVarName,
											Value: defaultWorkersEnvVarValue,
										},
										{
											Name:  transformerpkg.DefaultFeastSource,
											Value: standardTransformerConfig.DefaultFeastSource.String(),
										},
										{
											Name:  transformerpkg.FeastStorageConfigs,
											Value: `{"1":{"redisCluster":{"feastServingUrl":"localhost:6866","redisAddress":["10.1.1.2","10.1.1.3"],"option":{"poolSize":5,"minIdleConnections":2}}},"2":{"bigtable":{"feastServingUrl":"localhost:6867","project":"gcp-project","instance":"instance","appProfile":"default","option":{"grpcConnectionPool":4,"keepAliveInterval":"120s","keepAliveTimeout":"60s","credentialJson":"eyJrZXkiOiJ2YWx1ZSJ9"}}}}`,
										},
										{Name: transformerpkg.ModelGRPCKeepAliveEnabled, Value: "true"},
										{Name: transformerpkg.ModelGRPCKeepAliveTime, Value: "45s"},
										{Name: transformerpkg.ModelGRPCKeepAliveTimeout, Value: "5s"},
										{Name: transformerpkg.FeastServingKeepAliveEnabled, Value: "true"},
										{Name: transformerpkg.FeastServingKeepAliveTime, Value: "30s"},
										{Name: transformerpkg.FeastServingKeepAliveTimeout, Value: "1s"},
										{Name: transformerpkg.FeastGRPCConnCount, Value: "5"},
										{Name: transformerpkg.KafkaTopic, Value: "caraml-project-model-prediction-log"},
										{Name: transformerpkg.KafkaBrokers, Value: standardTransformerConfig.Kafka.Brokers},
										{Name: transformerpkg.KafkaMaxMessageSizeBytes, Value: fmt.Sprintf("%v", standardTransformerConfig.Kafka.MaxMessageSizeBytes)},
										{Name: transformerpkg.KafkaConnectTimeoutMS, Value: fmt.Sprintf("%v", standardTransformerConfig.Kafka.ConnectTimeoutMS)},
										{Name: transformerpkg.KafkaSerialization, Value: string(standardTransformerConfig.Kafka.SerializationFmt)},
										{Name: transformerpkg.ModelServerConnCount, Value: "3"},
										{Name: transformerpkg.JaegerCollectorURL, Value: standardTransformerConfig.Jaeger.CollectorURL},
										{Name: transformerpkg.JaegerSamplerParam, Value: standardTransformerConfig.Jaeger.SamplerParam},
										{Name: transformerpkg.JaegerDisabled, Value: standardTransformerConfig.Jaeger.Disabled},
										{Name: transformerpkg.StandardTransformerConfigEnvName, Value: `{"standard_transformer":null}`},
									}, createDefaultTransformerEnvVars(modelSvcGRPC)),
									Resources:     expDefaultTransformerResourceRequests,
									Ports:         grpcRawContainerPorts,
									LivenessProbe: transformerProbeConfigUPI,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultTransformerResourceRequests.MinReplica,
							MaxReplicas: defaultTransformerResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
		{
			name: "standard transformer upi v1; pyfunc serverless; predictor UPI http enabled ",
			modelSvc: &models.Service{
				Name:         modelSvc.Name,
				ModelName:    modelSvc.ModelName,
				ModelVersion: modelSvc.ModelVersion,
				Namespace:    project.Name,
				ArtifactURI:  modelSvc.ArtifactURI,
				Type:         models.ModelTypePyFunc,
				Options: &models.ModelOption{
					PyFuncImageName: "gojek/project-model:1",
				},
				Metadata: modelSvc.Metadata,
				Transformer: &models.Transformer{
					Enabled:         true,
					TransformerType: models.StandardTransformerType,
					EnvVars: models.EnvVars{
						{
							Name:  transformerpkg.StandardTransformerConfigEnvName,
							Value: `{"standard_transformer": null}`,
						},
						{
							Name:  transformerpkg.PredictorUPIHTTPEnabled,
							Value: "true",
						},
					},
				},
				PredictorUPIOverHTTPEnabled: true,
				Logger: &models.Logger{
					DestinationURL: loggerDestinationURL,
					Transformer: &models.LoggerConfig{
						Enabled: false,
						Mode:    models.LogRequest,
					},
				},
				Protocol:       protocol.UpiV1,
				DeploymentMode: deployment.ServerlessDeploymentMode,
			},
			deploymentScale: defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						annotationPrometheusScrapeFlag:                        "true",
						annotationPrometheusScrapePort:                        "8080",
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testTransformerScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						PodSpec: kservev1beta1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  kserveconstant.InferenceServiceContainerName,
									Image: "gojek/project-model:1",
									Env: models.MergeEnvVars(
										models.EnvVars{
											{Name: defaultWorkersEnvVarName, Value: defaultWorkersEnvVarValue},
											{Name: envGRPCOptions, Value: "{}"},
										},
										createPyFuncDefaultEnvVarsWithProtocol(modelSvc, protocol.HttpJson),
									).ToKubernetesEnvVars(),
									Resources:     expDefaultModelResourceRequests,
									LivenessProbe: probeConfig,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
					},
					Transformer: &kservev1beta1.TransformerSpec{
						PodSpec: kservev1beta1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "transformer",
									Image: standardTransformerConfig.ImageName,
									Env: MergeEnvVars([]corev1.EnvVar{
										{
											Name:  defaultWorkersEnvVarName,
											Value: defaultWorkersEnvVarValue,
										},
										{
											Name:  transformerpkg.DefaultFeastSource,
											Value: standardTransformerConfig.DefaultFeastSource.String(),
										},
										{
											Name:  transformerpkg.FeastStorageConfigs,
											Value: `{"1":{"redisCluster":{"feastServingUrl":"localhost:6866","redisAddress":["10.1.1.2","10.1.1.3"],"option":{"poolSize":5,"minIdleConnections":2}}},"2":{"bigtable":{"feastServingUrl":"localhost:6867","project":"gcp-project","instance":"instance","appProfile":"default","option":{"grpcConnectionPool":4,"keepAliveInterval":"120s","keepAliveTimeout":"60s","credentialJson":"eyJrZXkiOiJ2YWx1ZSJ9"}}}}`,
										},
										{Name: transformerpkg.ModelGRPCKeepAliveEnabled, Value: "true"},
										{Name: transformerpkg.ModelGRPCKeepAliveTime, Value: "45s"},
										{Name: transformerpkg.ModelGRPCKeepAliveTimeout, Value: "5s"},
										{Name: transformerpkg.FeastServingKeepAliveEnabled, Value: "true"},
										{Name: transformerpkg.FeastServingKeepAliveTime, Value: "30s"},
										{Name: transformerpkg.FeastServingKeepAliveTimeout, Value: "1s"},
										{Name: transformerpkg.FeastGRPCConnCount, Value: "5"},
										{Name: transformerpkg.KafkaTopic, Value: "caraml-project-model-prediction-log"},
										{Name: transformerpkg.KafkaBrokers, Value: standardTransformerConfig.Kafka.Brokers},
										{Name: transformerpkg.KafkaMaxMessageSizeBytes, Value: fmt.Sprintf("%v", standardTransformerConfig.Kafka.MaxMessageSizeBytes)},
										{Name: transformerpkg.KafkaConnectTimeoutMS, Value: fmt.Sprintf("%v", standardTransformerConfig.Kafka.ConnectTimeoutMS)},
										{Name: transformerpkg.KafkaSerialization, Value: string(standardTransformerConfig.Kafka.SerializationFmt)},
										{Name: transformerpkg.ModelServerConnCount, Value: "3"},
										{Name: transformerpkg.JaegerCollectorURL, Value: standardTransformerConfig.Jaeger.CollectorURL},
										{Name: transformerpkg.JaegerSamplerParam, Value: standardTransformerConfig.Jaeger.SamplerParam},
										{Name: transformerpkg.JaegerDisabled, Value: standardTransformerConfig.Jaeger.Disabled},
										{Name: transformerpkg.StandardTransformerConfigEnvName, Value: `{"standard_transformer":null}`},
										{Name: transformerpkg.PredictorUPIHTTPEnabled, Value: "true"},
									}, createDefaultTransformerEnvVars(modelSvcGRPC)),
									Resources:     expDefaultTransformerResourceRequests,
									Ports:         grpcServerlessContainerPorts,
									LivenessProbe: transformerProbeConfigUPI,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultTransformerResourceRequests.MinReplica,
							MaxReplicas: defaultTransformerResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deployConfig := &config.DeploymentConfig{
				DefaultModelResourceRequests:          defaultModelResourceRequests,
				DefaultTransformerResourceRequests:    defaultTransformerResourceRequests,
				QueueResourcePercentage:               queueResourcePercentage,
				PyfuncGRPCOptions:                     "{}",
				StandardTransformer:                   standardTransformerConfig,
				UserContainerCPUDefaultLimit:          userContainerCPUDefaultLimit,
				UserContainerCPULimitRequestFactor:    userContainerCPULimitRequestFactor,
				UserContainerMemoryLimitRequestFactor: userContainerMemoryLimitRequestFactor,
				DefaultEnvVarsWithoutCPULimits:        []corev1.EnvVar{defaultEnvVarWithoutCPULimits},
			}

			tpl := NewInferenceServiceTemplater(*deployConfig)
			infSvcSpec, err := tpl.CreateInferenceServiceSpec(tt.modelSvc, tt.deploymentScale)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// Sort all env vars in both expected and actual inference service specs before comparing
			sortInferenceServiceSpecEnvVars(tt.exp.Spec)
			sortInferenceServiceSpecEnvVars(infSvcSpec.Spec)
			assert.Equal(t, tt.exp, infSvcSpec)
		})
	}
}

func TestCreateInferenceServiceSpecWithLogger(t *testing.T) {
	err := labeller.InitKubernetesLabeller("gojek.com/", "caraml.dev/", testEnvironmentName)
	assert.NoError(t, err)

	defer func() {
		_ = labeller.InitKubernetesLabeller("", "", "")
	}()

	project := mlp.Project{
		Name: "project",
	}

	loggerDestinationURL := "http://destination.default"
	modelSvc := &models.Service{
		Name:         "model-1",
		ModelName:    "model",
		Namespace:    project.Name,
		ModelVersion: "1",
		ArtifactURI:  "gs://my-artifacet",
		Metadata: models.Metadata{
			App:       "model",
			Component: models.ComponentModelVersion,
			Stream:    "dsp",
			Team:      "dsp",
			Labels: mlp.Labels{
				{
					Key:   "sample",
					Value: "true",
				},
			},
		},
		Protocol: protocol.HttpJson,
	}

	queueResourcePercentage := "2"
	storageUri := fmt.Sprintf("%s/model", modelSvc.ArtifactURI)

	// Liveness probe config for the model containers
	probeConfig := createLivenessProbeSpec(protocol.HttpJson, fmt.Sprintf("/v1/models/%s", modelSvc.Name))

	// Liveness probe config for the transformers
	transformerProbeConfig := createLivenessProbeSpec(protocol.HttpJson, "/")

	tests := []struct {
		name            string
		modelSvc        *models.Service
		deploymentScale DeploymentScale
		exp             *kservev1beta1.InferenceService
		wantErr         bool
	}{
		{
			name: "model logger enabled",
			modelSvc: &models.Service{
				Name:         modelSvc.Name,
				ModelName:    modelSvc.ModelName,
				ModelVersion: modelSvc.ModelVersion,
				Namespace:    project.Name,
				ArtifactURI:  modelSvc.ArtifactURI,
				Type:         models.ModelTypeTensorflow,
				Options:      &models.ModelOption{},
				Metadata:     modelSvc.Metadata,
				Logger: &models.Logger{
					DestinationURL: loggerDestinationURL,
					Model: &models.LoggerConfig{
						Enabled: true,
						Mode:    models.LogAll,
					},
				},
				Protocol: protocol.HttpJson,
			},
			deploymentScale: defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testPredictorScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						Tensorflow: &kservev1beta1.TFServingSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     expDefaultModelResourceRequests,
									LivenessProbe: probeConfig,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
							Logger: &kservev1beta1.LoggerSpec{
								URL:  &loggerDestinationURL,
								Mode: kservev1beta1.LogAll,
							},
						},
					},
				},
			},
		},
		{
			name: "model logger enabled with transformer",
			modelSvc: &models.Service{
				Name:         modelSvc.Name,
				ModelName:    modelSvc.ModelName,
				ModelVersion: modelSvc.ModelVersion,
				Namespace:    project.Name,
				ArtifactURI:  modelSvc.ArtifactURI,
				Type:         models.ModelTypeTensorflow,
				Options:      &models.ModelOption{},
				Metadata:     modelSvc.Metadata,
				Transformer: &models.Transformer{
					Enabled: true,
					Image:   "ghcr.io/gojek/merlin-transformer-test",
					Command: "python",
					Args:    "main.py",
				},
				Logger: &models.Logger{
					DestinationURL: loggerDestinationURL,
					Model: &models.LoggerConfig{
						Enabled: true,
						Mode:    models.LogAll,
					},
				},
				Protocol: protocol.HttpJson,
			},
			deploymentScale: defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testTransformerScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						Tensorflow: &kservev1beta1.TFServingSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     expDefaultModelResourceRequests,
									LivenessProbe: probeConfig,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
							Logger: &kservev1beta1.LoggerSpec{
								URL:  &loggerDestinationURL,
								Mode: kservev1beta1.LogAll,
							},
						},
					},
					Transformer: &kservev1beta1.TransformerSpec{
						PodSpec: kservev1beta1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:          "transformer",
									Image:         "ghcr.io/gojek/merlin-transformer-test",
									Command:       []string{"python"},
									Args:          []string{"main.py"},
									Env:           createDefaultTransformerEnvVars(modelSvc),
									Resources:     expDefaultTransformerResourceRequests,
									LivenessProbe: transformerProbeConfig,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultTransformerResourceRequests.MinReplica,
							MaxReplicas: defaultTransformerResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
		{
			name: "model logger disabled with transformer",
			modelSvc: &models.Service{
				Name:         modelSvc.Name,
				ModelName:    modelSvc.ModelName,
				ModelVersion: modelSvc.ModelVersion,
				Namespace:    project.Name,
				ArtifactURI:  modelSvc.ArtifactURI,
				Type:         models.ModelTypeTensorflow,
				Options:      &models.ModelOption{},
				Metadata:     modelSvc.Metadata,
				Transformer: &models.Transformer{
					Enabled: true,
					Image:   "ghcr.io/gojek/merlin-transformer-test",
					Command: "python",
					Args:    "main.py",
				},
				Logger: &models.Logger{
					DestinationURL: loggerDestinationURL,
					Model: &models.LoggerConfig{
						Enabled: false,
						Mode:    models.LogAll,
					},
				},
				Protocol: protocol.HttpJson,
			},
			deploymentScale: defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testTransformerScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						Tensorflow: &kservev1beta1.TFServingSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     expDefaultModelResourceRequests,
									LivenessProbe: probeConfig,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
					},
					Transformer: &kservev1beta1.TransformerSpec{
						PodSpec: kservev1beta1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:          "transformer",
									Image:         "ghcr.io/gojek/merlin-transformer-test",
									Command:       []string{"python"},
									Args:          []string{"main.py"},
									Env:           createDefaultTransformerEnvVars(modelSvc),
									Resources:     expDefaultTransformerResourceRequests,
									LivenessProbe: transformerProbeConfig,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultTransformerResourceRequests.MinReplica,
							MaxReplicas: defaultTransformerResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
		{
			name: "transformer logger enabled",
			modelSvc: &models.Service{
				Name:         modelSvc.Name,
				ModelName:    modelSvc.ModelName,
				ModelVersion: modelSvc.ModelVersion,
				Namespace:    project.Name,
				ArtifactURI:  modelSvc.ArtifactURI,
				Type:         models.ModelTypeTensorflow,
				Options:      &models.ModelOption{},
				Metadata:     modelSvc.Metadata,
				Transformer: &models.Transformer{
					Enabled: true,
					Image:   "ghcr.io/gojek/merlin-transformer-test",
					Command: "python",
					Args:    "main.py",
				},
				Logger: &models.Logger{
					DestinationURL: loggerDestinationURL,
					Transformer: &models.LoggerConfig{
						Enabled: true,
						Mode:    models.LogRequest,
					},
				},
				Protocol: protocol.HttpJson,
			},
			deploymentScale: defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testTransformerScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						Tensorflow: &kservev1beta1.TFServingSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     expDefaultModelResourceRequests,
									LivenessProbe: probeConfig,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
					},
					Transformer: &kservev1beta1.TransformerSpec{
						PodSpec: kservev1beta1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:          "transformer",
									Image:         "ghcr.io/gojek/merlin-transformer-test",
									Command:       []string{"python"},
									Args:          []string{"main.py"},
									Env:           createDefaultTransformerEnvVars(modelSvc),
									Resources:     expDefaultTransformerResourceRequests,
									LivenessProbe: transformerProbeConfig,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultTransformerResourceRequests.MinReplica,
							MaxReplicas: defaultTransformerResourceRequests.MaxReplica,
							Logger: &kservev1beta1.LoggerSpec{
								URL:  &loggerDestinationURL,
								Mode: kservev1beta1.LogRequest,
							},
						},
					},
				},
			},
		},
		{
			name: "transformer logger disabled",
			modelSvc: &models.Service{
				Name:         modelSvc.Name,
				ModelName:    modelSvc.ModelName,
				ModelVersion: modelSvc.ModelVersion,
				Namespace:    project.Name,
				ArtifactURI:  modelSvc.ArtifactURI,
				Type:         models.ModelTypeTensorflow,
				Options:      &models.ModelOption{},
				Metadata:     modelSvc.Metadata,
				Transformer: &models.Transformer{
					Enabled: true,
					Image:   "ghcr.io/gojek/merlin-transformer-test",
					Command: "python",
					Args:    "main.py",
				},
				Logger: &models.Logger{
					DestinationURL: loggerDestinationURL,
					Transformer: &models.LoggerConfig{
						Enabled: false,
						Mode:    models.LogRequest,
					},
				},
				Protocol: protocol.HttpJson,
			},
			deploymentScale: defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testTransformerScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						Tensorflow: &kservev1beta1.TFServingSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     expDefaultModelResourceRequests,
									LivenessProbe: probeConfig,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
					},
					Transformer: &kservev1beta1.TransformerSpec{
						PodSpec: kservev1beta1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:          "transformer",
									Image:         "ghcr.io/gojek/merlin-transformer-test",
									Command:       []string{"python"},
									Args:          []string{"main.py"},
									Env:           createDefaultTransformerEnvVars(modelSvc),
									Resources:     expDefaultTransformerResourceRequests,
									LivenessProbe: transformerProbeConfig,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultTransformerResourceRequests.MinReplica,
							MaxReplicas: defaultTransformerResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deployConfig := &config.DeploymentConfig{
				DefaultModelResourceRequests:          defaultModelResourceRequests,
				DefaultTransformerResourceRequests:    defaultTransformerResourceRequests,
				QueueResourcePercentage:               queueResourcePercentage,
				StandardTransformer:                   standardTransformerConfig,
				UserContainerCPUDefaultLimit:          userContainerCPUDefaultLimit,
				UserContainerCPULimitRequestFactor:    userContainerCPULimitRequestFactor,
				UserContainerMemoryLimitRequestFactor: userContainerMemoryLimitRequestFactor,
			}

			tpl := NewInferenceServiceTemplater(*deployConfig)
			infSvcSpec, err := tpl.CreateInferenceServiceSpec(tt.modelSvc, tt.deploymentScale)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.exp, infSvcSpec)
		})
	}
}

func TestCreateInferenceServiceSpecWithTopologySpreadConstraints(t *testing.T) {
	err := labeller.InitKubernetesLabeller("gojek.com/", "caraml.dev/", testEnvironmentName)
	assert.NoError(t, err)

	defer func() {
		_ = labeller.InitKubernetesLabeller("", "", "")
	}()

	project := mlp.Project{
		Name: "project",
	}

	modelSvc := &models.Service{
		Name:         "model-1",
		ModelName:    "model",
		Namespace:    project.Name,
		ModelVersion: "1",
		ArtifactURI:  "gs://my-artifacet",
		Metadata: models.Metadata{
			App:       "model",
			Component: models.ComponentModelVersion,
			Stream:    "dsp",
			Team:      "dsp",
			Labels: mlp.Labels{
				{
					Key:   "sample",
					Value: "true",
				},
			},
		},
		Protocol: protocol.HttpJson,
	}

	queueResourcePercentage := "2"
	storageUri := fmt.Sprintf("%s/model", modelSvc.ArtifactURI)

	// Liveness probe config for the model containers
	probeConfig := createLivenessProbeSpec(protocol.HttpJson, fmt.Sprintf("/v1/models/%s", modelSvc.Name))

	// Liveness probe config for the transformers
	transformerProbeConfig := createLivenessProbeSpec(protocol.HttpJson, "/")

	tests := []struct {
		name            string
		modelSvc        *models.Service
		deploymentScale DeploymentScale
		exp             *kservev1beta1.InferenceService
		wantErr         bool
	}{
		{
			name: "predictor with unspecified deployment mode (serverless)",
			modelSvc: &models.Service{
				Name:         modelSvc.Name,
				ModelName:    modelSvc.ModelName,
				ModelVersion: modelSvc.ModelVersion,
				Namespace:    project.Name,
				ArtifactURI:  modelSvc.ArtifactURI,
				Type:         models.ModelTypeTensorflow,
				Options:      &models.ModelOption{},
				Metadata:     modelSvc.Metadata,
				Protocol:     protocol.HttpJson,
			},
			deploymentScale: defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testPredictorScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						Tensorflow: &kservev1beta1.TFServingSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     expDefaultModelResourceRequests,
									LivenessProbe: probeConfig,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
						PodSpec: kservev1beta1.PodSpec{
							TopologySpreadConstraints: []corev1.TopologySpreadConstraint{
								{
									MaxSkew:           1,
									TopologyKey:       "kubernetes.io/hostname",
									WhenUnsatisfiable: corev1.ScheduleAnyway,
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"app": "model-1-predictor-00001",
										},
									},
								},
								{
									MaxSkew:           2,
									TopologyKey:       "kubernetes.io/hostname",
									WhenUnsatisfiable: corev1.DoNotSchedule,
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"app": "model-1-predictor-00001",
										},
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "app-expression",
												Operator: metav1.LabelSelectorOpIn,
												Values:   []string{"1"},
											},
										},
									},
								},
								{
									MaxSkew:           3,
									TopologyKey:       "kubernetes.io/hostname",
									WhenUnsatisfiable: corev1.DoNotSchedule,
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"app-label": "spread",
											"app":       "model-1-predictor-00001",
										},
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "app-expression",
												Operator: metav1.LabelSelectorOpIn,
												Values:   []string{"1"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "predictor with serverless deployment mode",
			modelSvc: &models.Service{
				Name:           modelSvc.Name,
				ModelName:      modelSvc.ModelName,
				ModelVersion:   modelSvc.ModelVersion,
				Namespace:      project.Name,
				ArtifactURI:    modelSvc.ArtifactURI,
				Type:           models.ModelTypeTensorflow,
				Options:        &models.ModelOption{},
				Metadata:       modelSvc.Metadata,
				DeploymentMode: deployment.ServerlessDeploymentMode,
				Protocol:       protocol.HttpJson,
			},
			deploymentScale: defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testPredictorScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						Tensorflow: &kservev1beta1.TFServingSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     expDefaultModelResourceRequests,
									LivenessProbe: probeConfig,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
						PodSpec: kservev1beta1.PodSpec{
							TopologySpreadConstraints: []corev1.TopologySpreadConstraint{
								{
									MaxSkew:           1,
									TopologyKey:       "kubernetes.io/hostname",
									WhenUnsatisfiable: corev1.ScheduleAnyway,
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"app": "model-1-predictor-00001",
										},
									},
								},
								{
									MaxSkew:           2,
									TopologyKey:       "kubernetes.io/hostname",
									WhenUnsatisfiable: corev1.DoNotSchedule,
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"app": "model-1-predictor-00001",
										},
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "app-expression",
												Operator: metav1.LabelSelectorOpIn,
												Values:   []string{"1"},
											},
										},
									},
								},
								{
									MaxSkew:           3,
									TopologyKey:       "kubernetes.io/hostname",
									WhenUnsatisfiable: corev1.DoNotSchedule,
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"app-label": "spread",
											"app":       "model-1-predictor-00001",
										},
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "app-expression",
												Operator: metav1.LabelSelectorOpIn,
												Values:   []string{"1"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "predictor with raw deployment mode",
			modelSvc: &models.Service{
				Name:           modelSvc.Name,
				ModelName:      modelSvc.ModelName,
				ModelVersion:   modelSvc.ModelVersion,
				Namespace:      project.Name,
				ArtifactURI:    modelSvc.ArtifactURI,
				Type:           models.ModelTypeTensorflow,
				Options:        &models.ModelOption{},
				Metadata:       modelSvc.Metadata,
				DeploymentMode: deployment.RawDeploymentMode,
				Protocol:       protocol.HttpJson,
			},
			deploymentScale: defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.RawDeployment),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						Tensorflow: &kservev1beta1.TFServingSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     expDefaultModelResourceRequests,
									LivenessProbe: probeConfig,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
						PodSpec: kservev1beta1.PodSpec{
							TopologySpreadConstraints: []corev1.TopologySpreadConstraint{
								{
									MaxSkew:           1,
									TopologyKey:       "kubernetes.io/hostname",
									WhenUnsatisfiable: corev1.ScheduleAnyway,
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"app": "isvc.model-1-predictor",
										},
									},
								},
								{
									MaxSkew:           2,
									TopologyKey:       "kubernetes.io/hostname",
									WhenUnsatisfiable: corev1.DoNotSchedule,
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"app": "isvc.model-1-predictor",
										},
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "app-expression",
												Operator: metav1.LabelSelectorOpIn,
												Values:   []string{"1"},
											},
										},
									},
								},
								{
									MaxSkew:           3,
									TopologyKey:       "kubernetes.io/hostname",
									WhenUnsatisfiable: corev1.DoNotSchedule,
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"app-label": "spread",
											"app":       "isvc.model-1-predictor",
										},
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "app-expression",
												Operator: metav1.LabelSelectorOpIn,
												Values:   []string{"1"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "predictor and transformer with unspecified deployment mode (serverless)",
			modelSvc: &models.Service{
				Name:         modelSvc.Name,
				ModelName:    modelSvc.ModelName,
				ModelVersion: modelSvc.ModelVersion,
				Namespace:    project.Name,
				ArtifactURI:  modelSvc.ArtifactURI,
				Type:         models.ModelTypeTensorflow,
				Options:      &models.ModelOption{},
				Metadata:     modelSvc.Metadata,
				Transformer: &models.Transformer{
					Enabled: true,
					Image:   "ghcr.io/gojek/merlin-transformer-test",
					Command: "python",
					Args:    "main.py",
				},
				Protocol: protocol.HttpJson,
			},
			deploymentScale: defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testTransformerScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						Tensorflow: &kservev1beta1.TFServingSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     expDefaultModelResourceRequests,
									LivenessProbe: probeConfig,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
						PodSpec: kservev1beta1.PodSpec{
							TopologySpreadConstraints: []corev1.TopologySpreadConstraint{
								{
									MaxSkew:           1,
									TopologyKey:       "kubernetes.io/hostname",
									WhenUnsatisfiable: corev1.ScheduleAnyway,
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"app": "model-1-predictor-00001",
										},
									},
								},
								{
									MaxSkew:           2,
									TopologyKey:       "kubernetes.io/hostname",
									WhenUnsatisfiable: corev1.DoNotSchedule,
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"app": "model-1-predictor-00001",
										},
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "app-expression",
												Operator: metav1.LabelSelectorOpIn,
												Values:   []string{"1"},
											},
										},
									},
								},
								{
									MaxSkew:           3,
									TopologyKey:       "kubernetes.io/hostname",
									WhenUnsatisfiable: corev1.DoNotSchedule,
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"app-label": "spread",
											"app":       "model-1-predictor-00001",
										},
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "app-expression",
												Operator: metav1.LabelSelectorOpIn,
												Values:   []string{"1"},
											},
										},
									},
								},
							},
						},
					},
					Transformer: &kservev1beta1.TransformerSpec{
						PodSpec: kservev1beta1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:          "transformer",
									Image:         "ghcr.io/gojek/merlin-transformer-test",
									Command:       []string{"python"},
									Args:          []string{"main.py"},
									Env:           createDefaultTransformerEnvVars(modelSvc),
									Resources:     expDefaultTransformerResourceRequests,
									LivenessProbe: transformerProbeConfig,
								},
							},
							TopologySpreadConstraints: []corev1.TopologySpreadConstraint{
								{
									MaxSkew:           1,
									TopologyKey:       "kubernetes.io/hostname",
									WhenUnsatisfiable: corev1.ScheduleAnyway,
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"app": "model-1-transformer-00001",
										},
									},
								},
								{
									MaxSkew:           2,
									TopologyKey:       "kubernetes.io/hostname",
									WhenUnsatisfiable: corev1.DoNotSchedule,
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"app": "model-1-transformer-00001",
										},
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "app-expression",
												Operator: metav1.LabelSelectorOpIn,
												Values:   []string{"1"},
											},
										},
									},
								},
								{
									MaxSkew:           3,
									TopologyKey:       "kubernetes.io/hostname",
									WhenUnsatisfiable: corev1.DoNotSchedule,
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"app-label": "spread",
											"app":       "model-1-transformer-00001",
										},
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "app-expression",
												Operator: metav1.LabelSelectorOpIn,
												Values:   []string{"1"},
											},
										},
									},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultTransformerResourceRequests.MinReplica,
							MaxReplicas: defaultTransformerResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
		{
			name: "predictor and transformer with serverless deployment mode",
			modelSvc: &models.Service{
				Name:         modelSvc.Name,
				ModelName:    modelSvc.ModelName,
				ModelVersion: modelSvc.ModelVersion,
				Namespace:    project.Name,
				ArtifactURI:  modelSvc.ArtifactURI,
				Type:         models.ModelTypeTensorflow,
				Options:      &models.ModelOption{},
				Metadata:     modelSvc.Metadata,
				Transformer: &models.Transformer{
					Enabled: true,
					Image:   "ghcr.io/gojek/merlin-transformer-test",
					Command: "python",
					Args:    "main.py",
				},
				DeploymentMode: deployment.ServerlessDeploymentMode,
				Protocol:       protocol.HttpJson,
			},
			deploymentScale: defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.InitialScaleAnnotationKey:               fmt.Sprint(testTransformerScale),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						Tensorflow: &kservev1beta1.TFServingSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     expDefaultModelResourceRequests,
									LivenessProbe: probeConfig,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
						PodSpec: kservev1beta1.PodSpec{
							TopologySpreadConstraints: []corev1.TopologySpreadConstraint{
								{
									MaxSkew:           1,
									TopologyKey:       "kubernetes.io/hostname",
									WhenUnsatisfiable: corev1.ScheduleAnyway,
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"app": "model-1-predictor-00001",
										},
									},
								},
								{
									MaxSkew:           2,
									TopologyKey:       "kubernetes.io/hostname",
									WhenUnsatisfiable: corev1.DoNotSchedule,
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"app": "model-1-predictor-00001",
										},
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "app-expression",
												Operator: metav1.LabelSelectorOpIn,
												Values:   []string{"1"},
											},
										},
									},
								},
								{
									MaxSkew:           3,
									TopologyKey:       "kubernetes.io/hostname",
									WhenUnsatisfiable: corev1.DoNotSchedule,
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"app-label": "spread",
											"app":       "model-1-predictor-00001",
										},
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "app-expression",
												Operator: metav1.LabelSelectorOpIn,
												Values:   []string{"1"},
											},
										},
									},
								},
							},
						},
					},
					Transformer: &kservev1beta1.TransformerSpec{
						PodSpec: kservev1beta1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:          "transformer",
									Image:         "ghcr.io/gojek/merlin-transformer-test",
									Command:       []string{"python"},
									Args:          []string{"main.py"},
									Env:           createDefaultTransformerEnvVars(modelSvc),
									Resources:     expDefaultTransformerResourceRequests,
									LivenessProbe: transformerProbeConfig,
								},
							},
							TopologySpreadConstraints: []corev1.TopologySpreadConstraint{
								{
									MaxSkew:           1,
									TopologyKey:       "kubernetes.io/hostname",
									WhenUnsatisfiable: corev1.ScheduleAnyway,
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"app": "model-1-transformer-00001",
										},
									},
								},
								{
									MaxSkew:           2,
									TopologyKey:       "kubernetes.io/hostname",
									WhenUnsatisfiable: corev1.DoNotSchedule,
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"app": "model-1-transformer-00001",
										},
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "app-expression",
												Operator: metav1.LabelSelectorOpIn,
												Values:   []string{"1"},
											},
										},
									},
								},
								{
									MaxSkew:           3,
									TopologyKey:       "kubernetes.io/hostname",
									WhenUnsatisfiable: corev1.DoNotSchedule,
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"app-label": "spread",
											"app":       "model-1-transformer-00001",
										},
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "app-expression",
												Operator: metav1.LabelSelectorOpIn,
												Values:   []string{"1"},
											},
										},
									},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultTransformerResourceRequests.MinReplica,
							MaxReplicas: defaultTransformerResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
		{
			name: "predictor and transformer with raw deployment mode",
			modelSvc: &models.Service{
				Name:         modelSvc.Name,
				ModelName:    modelSvc.ModelName,
				ModelVersion: modelSvc.ModelVersion,
				Namespace:    project.Name,
				ArtifactURI:  modelSvc.ArtifactURI,
				Type:         models.ModelTypeTensorflow,
				Options:      &models.ModelOption{},
				Metadata:     modelSvc.Metadata,
				Transformer: &models.Transformer{
					Enabled: true,
					Image:   "ghcr.io/gojek/merlin-transformer-test",
					Command: "python",
					Args:    "main.py",
				},
				DeploymentMode: deployment.RawDeploymentMode,
				Protocol:       protocol.HttpJson,
			},
			deploymentScale: defaultDeploymentScale,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.RawDeployment),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  testEnvironmentName,
						"gojek.com/orchestrator": testOrchestratorName,
						"gojek.com/stream":       modelSvc.Metadata.Stream,
						"gojek.com/team":         modelSvc.Metadata.Team,
						"sample":                 "true",
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						Tensorflow: &kservev1beta1.TFServingSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     expDefaultModelResourceRequests,
									LivenessProbe: probeConfig,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
						PodSpec: kservev1beta1.PodSpec{
							TopologySpreadConstraints: []corev1.TopologySpreadConstraint{
								{
									MaxSkew:           1,
									TopologyKey:       "kubernetes.io/hostname",
									WhenUnsatisfiable: corev1.ScheduleAnyway,
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"app": "isvc.model-1-predictor",
										},
									},
								},
								{
									MaxSkew:           2,
									TopologyKey:       "kubernetes.io/hostname",
									WhenUnsatisfiable: corev1.DoNotSchedule,
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"app": "isvc.model-1-predictor",
										},
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "app-expression",
												Operator: metav1.LabelSelectorOpIn,
												Values:   []string{"1"},
											},
										},
									},
								},
								{
									MaxSkew:           3,
									TopologyKey:       "kubernetes.io/hostname",
									WhenUnsatisfiable: corev1.DoNotSchedule,
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"app-label": "spread",
											"app":       "isvc.model-1-predictor",
										},
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "app-expression",
												Operator: metav1.LabelSelectorOpIn,
												Values:   []string{"1"},
											},
										},
									},
								},
							},
						},
					},
					Transformer: &kservev1beta1.TransformerSpec{
						PodSpec: kservev1beta1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:          "transformer",
									Image:         "ghcr.io/gojek/merlin-transformer-test",
									Command:       []string{"python"},
									Args:          []string{"main.py"},
									Env:           createDefaultTransformerEnvVars(modelSvc),
									Resources:     expDefaultTransformerResourceRequests,
									LivenessProbe: transformerProbeConfig,
								},
							},
							TopologySpreadConstraints: []corev1.TopologySpreadConstraint{
								{
									MaxSkew:           1,
									TopologyKey:       "kubernetes.io/hostname",
									WhenUnsatisfiable: corev1.ScheduleAnyway,
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"app": "isvc.model-1-transformer",
										},
									},
								},
								{
									MaxSkew:           2,
									TopologyKey:       "kubernetes.io/hostname",
									WhenUnsatisfiable: corev1.DoNotSchedule,
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"app": "isvc.model-1-transformer",
										},
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "app-expression",
												Operator: metav1.LabelSelectorOpIn,
												Values:   []string{"1"},
											},
										},
									},
								},
								{
									MaxSkew:           3,
									TopologyKey:       "kubernetes.io/hostname",
									WhenUnsatisfiable: corev1.DoNotSchedule,
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"app-label": "spread",
											"app":       "isvc.model-1-transformer",
										},
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "app-expression",
												Operator: metav1.LabelSelectorOpIn,
												Values:   []string{"1"},
											},
										},
									},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultTransformerResourceRequests.MinReplica,
							MaxReplicas: defaultTransformerResourceRequests.MaxReplica,
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deployConfig := &config.DeploymentConfig{
				DefaultModelResourceRequests:       defaultModelResourceRequests,
				DefaultTransformerResourceRequests: defaultTransformerResourceRequests,
				QueueResourcePercentage:            queueResourcePercentage,
				TopologySpreadConstraints: []corev1.TopologySpreadConstraint{
					{
						MaxSkew:           1,
						TopologyKey:       "kubernetes.io/hostname",
						WhenUnsatisfiable: corev1.ScheduleAnyway,
					},
					{
						MaxSkew:           2,
						TopologyKey:       "kubernetes.io/hostname",
						WhenUnsatisfiable: corev1.DoNotSchedule,
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "app-expression",
									Operator: metav1.LabelSelectorOpIn,
									Values:   []string{"1"},
								},
							},
						},
					},
					{
						MaxSkew:           3,
						TopologyKey:       "kubernetes.io/hostname",
						WhenUnsatisfiable: corev1.DoNotSchedule,
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app-label": "spread",
							},
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "app-expression",
									Operator: metav1.LabelSelectorOpIn,
									Values:   []string{"1"},
								},
							},
						},
					},
				},
				StandardTransformer:                   standardTransformerConfig,
				UserContainerCPUDefaultLimit:          userContainerCPUDefaultLimit,
				UserContainerCPULimitRequestFactor:    userContainerCPULimitRequestFactor,
				UserContainerMemoryLimitRequestFactor: userContainerMemoryLimitRequestFactor,
			}

			tpl := NewInferenceServiceTemplater(*deployConfig)
			infSvcSpec, err := tpl.CreateInferenceServiceSpec(tt.modelSvc, tt.deploymentScale)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.exp, infSvcSpec)
		})
	}
}

func TestCreateTransformerSpec(t *testing.T) {
	one := 1
	cpuRequest := resource.MustParse("1")
	memoryRequest := resource.MustParse("1Gi")

	cpuLimit := resource.MustParse("10")
	memoryLimit := ScaleQuantity(memoryRequest, 2)
	customCPULimit := resource.MustParse("8")

	// Liveness probe config for the transformers
	transformerProbeConfig := createLivenessProbeSpec(protocol.HttpJson, "/")

	modelSvc := &models.Service{
		Name:         "model-1",
		ModelName:    "model",
		Namespace:    "test",
		ModelVersion: "1",
		ArtifactURI:  "gs://my-artifacet",
		Metadata: models.Metadata{
			Team:   "dsp",
			Stream: "dsp",
			App:    "model",
			Labels: mlp.Labels{
				{
					Key:   "sample",
					Value: "true",
				},
			},
		},
		Protocol: protocol.HttpJson,
	}

	type args struct {
		modelService *models.Service
		transformer  *models.Transformer
		config       *config.DeploymentConfig
	}
	tests := []struct {
		name string
		args args
		want *kservev1beta1.TransformerSpec
	}{
		{
			"standard transformer",
			args{
				&models.Service{
					Name:         modelSvc.Name,
					ModelName:    modelSvc.ModelName,
					ModelVersion: modelSvc.ModelVersion,
					Namespace:    modelSvc.Namespace,
					Protocol:     protocol.HttpJson,
				},
				&models.Transformer{
					TransformerType: models.StandardTransformerType,
					Image:           standardTransformerConfig.ImageName,
					Command:         "python",
					Args:            "main.py",
					ResourceRequest: &models.ResourceRequest{
						MinReplica:    1,
						MaxReplica:    1,
						CPURequest:    cpuRequest,
						MemoryRequest: memoryRequest,
					},
					EnvVars: models.EnvVars{
						{Name: transformerpkg.JaegerCollectorURL, Value: "NEW_HOST"}, // test user overwrite
					},
				},
				&config.DeploymentConfig{
					StandardTransformer:                   standardTransformerConfig,
					UserContainerCPUDefaultLimit:          userContainerCPUDefaultLimit,
					UserContainerCPULimitRequestFactor:    userContainerCPULimitRequestFactor,
					UserContainerMemoryLimitRequestFactor: userContainerMemoryLimitRequestFactor,
				},
			},
			&kservev1beta1.TransformerSpec{
				PodSpec: kservev1beta1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "transformer",
							Image:   standardTransformerConfig.ImageName,
							Command: []string{"python"},
							Args:    []string{"main.py"},
							Env: MergeEnvVars([]corev1.EnvVar{
								{Name: transformerpkg.DefaultFeastSource, Value: standardTransformerConfig.DefaultFeastSource.String()},
								{
									Name:  transformerpkg.FeastStorageConfigs,
									Value: `{"1":{"redisCluster":{"feastServingUrl":"localhost:6866","redisAddress":["10.1.1.2","10.1.1.3"],"option":{"poolSize":5,"minIdleConnections":2}}},"2":{"bigtable":{"feastServingUrl":"localhost:6867","project":"gcp-project","instance":"instance","appProfile":"default","option":{"grpcConnectionPool":4,"keepAliveInterval":"120s","keepAliveTimeout":"60s","credentialJson":"eyJrZXkiOiJ2YWx1ZSJ9"}}}}`,
								},
								{Name: transformerpkg.FeastServingKeepAliveEnabled, Value: "true"},
								{Name: transformerpkg.FeastServingKeepAliveTime, Value: "30s"},
								{Name: transformerpkg.FeastServingKeepAliveTimeout, Value: "1s"},
								{Name: transformerpkg.FeastGRPCConnCount, Value: "5"},
								{Name: transformerpkg.JaegerCollectorURL, Value: "NEW_HOST"},
								{Name: transformerpkg.JaegerSamplerParam, Value: standardTransformerConfig.Jaeger.SamplerParam},
								{Name: transformerpkg.JaegerDisabled, Value: standardTransformerConfig.Jaeger.Disabled},
							}, createDefaultTransformerEnvVars(modelSvc)),
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    cpuRequest,
									corev1.ResourceMemory: memoryRequest,
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    cpuLimit,
									corev1.ResourceMemory: memoryLimit,
								},
							},
							LivenessProbe: transformerProbeConfig,
						},
					},
				},
				ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
					MinReplicas: &one,
					MaxReplicas: one,
				},
			},
		},
		{
			"standard transformer with cpu limits specified",
			args{
				&models.Service{
					Name:         modelSvc.Name,
					ModelName:    modelSvc.ModelName,
					ModelVersion: modelSvc.ModelVersion,
					Namespace:    modelSvc.Namespace,
					Protocol:     protocol.HttpJson,
				},
				&models.Transformer{
					TransformerType: models.StandardTransformerType,
					Image:           standardTransformerConfig.ImageName,
					Command:         "python",
					Args:            "main.py",
					ResourceRequest: &models.ResourceRequest{
						MinReplica:    1,
						MaxReplica:    1,
						CPURequest:    cpuRequest,
						MemoryRequest: memoryRequest,
						CPULimit:      &customCPULimit,
					},
					EnvVars: models.EnvVars{
						{Name: transformerpkg.JaegerCollectorURL, Value: "NEW_HOST"}, // test user overwrite
					},
				},
				&config.DeploymentConfig{
					StandardTransformer:                   standardTransformerConfig,
					UserContainerCPUDefaultLimit:          userContainerCPUDefaultLimit,
					UserContainerCPULimitRequestFactor:    userContainerCPULimitRequestFactor,
					UserContainerMemoryLimitRequestFactor: userContainerMemoryLimitRequestFactor,
				},
			},
			&kservev1beta1.TransformerSpec{
				PodSpec: kservev1beta1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "transformer",
							Image:   standardTransformerConfig.ImageName,
							Command: []string{"python"},
							Args:    []string{"main.py"},
							Env: MergeEnvVars([]corev1.EnvVar{
								{Name: transformerpkg.DefaultFeastSource, Value: standardTransformerConfig.DefaultFeastSource.String()},
								{
									Name:  transformerpkg.FeastStorageConfigs,
									Value: `{"1":{"redisCluster":{"feastServingUrl":"localhost:6866","redisAddress":["10.1.1.2","10.1.1.3"],"option":{"poolSize":5,"minIdleConnections":2}}},"2":{"bigtable":{"feastServingUrl":"localhost:6867","project":"gcp-project","instance":"instance","appProfile":"default","option":{"grpcConnectionPool":4,"keepAliveInterval":"120s","keepAliveTimeout":"60s","credentialJson":"eyJrZXkiOiJ2YWx1ZSJ9"}}}}`,
								},
								{Name: transformerpkg.FeastServingKeepAliveEnabled, Value: "true"},
								{Name: transformerpkg.FeastServingKeepAliveTime, Value: "30s"},
								{Name: transformerpkg.FeastServingKeepAliveTimeout, Value: "1s"},
								{Name: transformerpkg.FeastGRPCConnCount, Value: "5"},
								{Name: transformerpkg.JaegerCollectorURL, Value: "NEW_HOST"},
								{Name: transformerpkg.JaegerSamplerParam, Value: standardTransformerConfig.Jaeger.SamplerParam},
								{Name: transformerpkg.JaegerDisabled, Value: standardTransformerConfig.Jaeger.Disabled},
							}, createDefaultTransformerEnvVars(modelSvc)),
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    cpuRequest,
									corev1.ResourceMemory: memoryRequest,
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    customCPULimit,
									corev1.ResourceMemory: memoryLimit,
								},
							},
							LivenessProbe: transformerProbeConfig,
						},
					},
				},
				ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
					MinReplicas: &one,
					MaxReplicas: one,
				},
			},
		},
		{
			"custom transformer",
			args{
				&models.Service{
					Name:         modelSvc.Name,
					ModelName:    modelSvc.ModelName,
					ModelVersion: modelSvc.ModelVersion,
					Namespace:    modelSvc.Namespace,
					Protocol:     protocol.HttpJson,
				},
				&models.Transformer{
					TransformerType: models.CustomTransformerType,
					Image:           "ghcr.io/gojek/merlin-transformer-test",
					Command:         "python",
					Args:            "main.py",
					ResourceRequest: &models.ResourceRequest{
						MinReplica:    1,
						MaxReplica:    1,
						CPURequest:    cpuRequest,
						MemoryRequest: memoryRequest,
					},
				},
				&config.DeploymentConfig{
					UserContainerCPUDefaultLimit:          userContainerCPUDefaultLimit,
					UserContainerCPULimitRequestFactor:    userContainerCPULimitRequestFactor,
					UserContainerMemoryLimitRequestFactor: userContainerMemoryLimitRequestFactor,
				},
			},
			&kservev1beta1.TransformerSpec{
				PodSpec: kservev1beta1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "transformer",
							Image:   "ghcr.io/gojek/merlin-transformer-test",
							Command: []string{"python"},
							Args:    []string{"main.py"},
							Env:     createDefaultTransformerEnvVars(modelSvc),
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    cpuRequest,
									corev1.ResourceMemory: memoryRequest,
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    cpuLimit,
									corev1.ResourceMemory: memoryLimit,
								},
							},
							LivenessProbe: transformerProbeConfig,
						},
					},
				},
				ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
					MinReplicas: &one,
					MaxReplicas: one,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl := NewInferenceServiceTemplater(*tt.args.config)
			got, err := tpl.createTransformerSpec(tt.args.modelService, tt.args.transformer)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func createPyFuncDefaultEnvVarsWithProtocol(svc *models.Service, protocolValue protocol.Protocol) models.EnvVars {
	envVars := models.EnvVars{
		models.EnvVar{
			Name:  envPyFuncModelName,
			Value: models.CreateInferenceServiceName(svc.ModelName, svc.ModelVersion, svc.RevisionID.String()),
		},
		models.EnvVar{
			Name:  envModelName,
			Value: svc.ModelName,
		},
		models.EnvVar{
			Name:  envModelVersion,
			Value: svc.ModelVersion,
		},
		models.EnvVar{
			Name:  envModelFullName,
			Value: models.CreateInferenceServiceName(svc.ModelName, svc.ModelVersion, svc.RevisionID.String()),
		},
		models.EnvVar{
			Name:  envHTTPPort,
			Value: fmt.Sprint(defaultHTTPPort),
		},
		models.EnvVar{
			Name:  envGRPCPort,
			Value: fmt.Sprint(defaultGRPCPort),
		},
		models.EnvVar{
			Name:  envProtocol,
			Value: fmt.Sprint(protocolValue),
		},
		models.EnvVar{
			Name:  envProject,
			Value: svc.Namespace,
		},
	}

	if svc.ModelSchema != nil {
		compactedfJsonBuffer := new(bytes.Buffer)
		marshalledSchema, _ := json.Marshal(svc.ModelSchema)
		if err := json.Compact(compactedfJsonBuffer, []byte(marshalledSchema)); err == nil {
			envVars = append(envVars, models.EnvVar{
				Name:  envModelSchema,
				Value: compactedfJsonBuffer.String(),
			})
		}
	}
	return envVars
}

func createPyFuncPublisherEnvVars(svc *models.Service, pyfuncPublisher config.PyFuncPublisherConfig) models.EnvVars {
	envVars := models.EnvVars{
		models.EnvVar{
			Name:  envPublisherEnabled,
			Value: strconv.FormatBool(svc.EnabledModelObservability),
		},
		models.EnvVar{
			Name:  envPublisherKafkaTopic,
			Value: svc.GetPredictionLogTopicForVersion(),
		},
		models.EnvVar{
			Name:  envPublisherNumPartitions,
			Value: fmt.Sprintf("%d", pyfuncPublisher.Kafka.NumPartitions),
		},
		models.EnvVar{
			Name:  envPublisherReplicationFactor,
			Value: fmt.Sprintf("%d", pyfuncPublisher.Kafka.ReplicationFactor),
		},
		models.EnvVar{
			Name:  envPublisherKafkaBrokers,
			Value: pyfuncPublisher.Kafka.Brokers,
		},
		models.EnvVar{
			Name:  envPublisherKafkaLinger,
			Value: fmt.Sprintf("%d", pyfuncPublisher.Kafka.LingerMS),
		},
		models.EnvVar{
			Name:  envPublisherKafkaAck,
			Value: fmt.Sprintf("%d", pyfuncPublisher.Kafka.Acks),
		},
		models.EnvVar{
			Name:  envPublisherSamplingRatio,
			Value: fmt.Sprintf("%f", pyfuncPublisher.SamplingRatioRate),
		},
		models.EnvVar{
			Name:  envPublisherKafkaConfig,
			Value: pyfuncPublisher.Kafka.AdditionalConfig,
		},
	}
	return envVars
}

func sortInferenceServiceSpecEnvVars(isvc kservev1beta1.InferenceServiceSpec) {
	// Sort env vars in predictor
	if isvc.Predictor.SKLearn != nil {
		sort.Slice(isvc.Predictor.SKLearn.Env, func(i, j int) bool {
			return isvc.Predictor.SKLearn.Env[i].Name < isvc.Predictor.SKLearn.Env[j].Name
		})
	} else if isvc.Predictor.XGBoost != nil {
		sort.Slice(isvc.Predictor.XGBoost.Env, func(i, j int) bool {
			return isvc.Predictor.XGBoost.Env[i].Name < isvc.Predictor.XGBoost.Env[j].Name
		})
	} else if isvc.Predictor.Tensorflow != nil {
		sort.Slice(isvc.Predictor.Tensorflow.Env, func(i, j int) bool {
			return isvc.Predictor.Tensorflow.Env[i].Name < isvc.Predictor.Tensorflow.Env[j].Name
		})
	} else if isvc.Predictor.PyTorch != nil {
		sort.Slice(isvc.Predictor.PyTorch.Env, func(i, j int) bool {
			return isvc.Predictor.PyTorch.Env[i].Name < isvc.Predictor.PyTorch.Env[j].Name
		})
	} else if isvc.Predictor.PodSpec.Containers != nil {
		for _, c := range isvc.Predictor.PodSpec.Containers {
			sort.Slice(c.Env, func(i, j int) bool {
				return c.Env[i].Name < c.Env[j].Name
			})
		}
	}

	// Sort env vars in transformer
	if isvc.Transformer != nil && isvc.Transformer.PodSpec.Containers != nil {
		for _, c := range isvc.Transformer.PodSpec.Containers {
			sort.Slice(c.Env, func(i, j int) bool {
				return c.Env[i].Name < c.Env[j].Name
			})
		}
	}
}
