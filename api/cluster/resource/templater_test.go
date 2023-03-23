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
	"fmt"
	"testing"
	"time"

	"github.com/gojek/merlin/pkg/autoscaling"
	"github.com/gojek/merlin/pkg/deployment"
	"github.com/gojek/merlin/pkg/protocol"
	"github.com/gojek/merlin/pkg/transformer"
	kservev1beta1 "github.com/kserve/kserve/pkg/apis/serving/v1beta1"
	kserveconstant "github.com/kserve/kserve/pkg/constants"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	knautoscaling "knative.dev/serving/pkg/apis/autoscaling"
	knserving "knative.dev/serving/pkg/apis/serving"

	"github.com/gojek/merlin/config"
	"github.com/gojek/merlin/mlp"
	"github.com/gojek/merlin/models"
)

const (
	environmentName  = "dev"
	orchestratorName = "merlin"
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
			corev1.ResourceCPU:    getLimit(defaultModelResourceRequests.CPURequest),
			corev1.ResourceMemory: getLimit(defaultModelResourceRequests.MemoryRequest),
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
			corev1.ResourceCPU:    getLimit(defaultTransformerResourceRequests.CPURequest),
			corev1.ResourceMemory: getLimit(defaultTransformerResourceRequests.MemoryRequest),
		},
	}

	userResourceRequests = &models.ResourceRequest{
		MinReplica:    1,
		MaxReplica:    10,
		CPURequest:    resource.MustParse("1"),
		MemoryRequest: resource.MustParse("1Gi"),
	}

	expUserResourceRequests = corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    userResourceRequests.CPURequest,
			corev1.ResourceMemory: userResourceRequests.MemoryRequest,
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    getLimit(userResourceRequests.CPURequest),
			corev1.ResourceMemory: getLimit(userResourceRequests.MemoryRequest),
		},
	}

	oneMinuteDuration         = config.Duration(time.Minute * 1)
	twoMinuteDuration         = config.Duration(time.Minute * 2)
	standardTransformerConfig = config.StandardTransformerConfig{
		ImageName:    "merlin-standard-transformer",
		FeastCoreURL: "core.feast.dev:8081",
		Jaeger: config.JaegerConfig{
			AgentHost:    "localhost",
			AgentPort:    "6831",
			SamplerType:  "const",
			SamplerParam: "1",
			Disabled:     "false",
		},
		FeastRedisConfig: &config.FeastRedisConfig{
			IsRedisCluster: true,
			ServingURL:     "localhost:6866",
			RedisAddresses: []string{"10.1.1.2", "10.1.1.3"},
			PoolSize:       5,
			MinIdleConn:    2,
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
	}
)

func TestCreateInferenceServiceSpec(t *testing.T) {
	err := models.InitKubernetesLabeller("gojek.com/", environmentName)
	assert.NoError(t, err)

	defer func() {
		_ = models.InitKubernetesLabeller("", "")
		assert.NoError(t, err)
	}()

	project := mlp.Project{
		Name: "project",
	}

	modelSvc := &models.Service{
		Name:         "my-model-1",
		ModelName:    "my-model",
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
	probeConfigUPI := createLivenessProbeSpec(protocol.UpiV1, fmt.Sprintf("/v1/models/%s", modelSvc.Name))

	tests := []struct {
		name               string
		modelSvc           *models.Service
		resourcePercentage string
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
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Env:           []corev1.EnvVar{},
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
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Env:           []corev1.EnvVar{},
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
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Env:           []corev1.EnvVar{},
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
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						kserveconstant.DeploymentMode: string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Env:           []corev1.EnvVar{},
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
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Env:           []corev1.EnvVar{},
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
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Env:           []corev1.EnvVar{},
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
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Env:           []corev1.EnvVar{},
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
			name: "pyfunc spec with liveness probe disabled",
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
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						"prometheus.io/scrape":                                "true",
						"prometheus.io/port":                                  "8080",
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Env: models.MergeEnvVars(models.EnvVars{models.EnvVar{Name: envOldDisableLivenessProbe, Value: "true"}},
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
			name: "pyfunc with liveness probe disabled",
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
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						"prometheus.io/scrape":                                "true",
						"prometheus.io/port":                                  "8080",
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Env: models.MergeEnvVars(models.EnvVars{models.EnvVar{Name: envDisableLivenessProbe, Value: "true"}},
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
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Env:           []corev1.EnvVar{},
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
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Image:     "gcr.io/custom-model:v0.1",
									Env:       createDefaultPredictorEnvVars(modelSvc).ToKubernetesEnvVars(),
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
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Image:     "gcr.io/custom-model:v0.1",
									Env:       createDefaultPredictorEnvVars(modelSvc).ToKubernetesEnvVars(),
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
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Env:           []corev1.EnvVar{},
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
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Env:           []corev1.EnvVar{},
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
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Env:           []corev1.EnvVar{},
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
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Env:           []corev1.EnvVar{},
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
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.ClassAnnotationKey:                      knautoscaling.KPA,
						knautoscaling.MetricAnnotationKey:                     knautoscaling.Concurrency,
						knautoscaling.TargetAnnotationKey:                     "2",
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Env:           []corev1.EnvVar{},
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
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Env:           []corev1.EnvVar{},
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
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Ports:         grpcContainerPorts,
									Env:           []corev1.EnvVar{},
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
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						"prometheus.io/scrape":                                "true",
						"prometheus.io/port":                                  "8080",
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Resources:     expDefaultModelResourceRequests,
									Ports:         grpcContainerPorts,
									Env:           createPyFuncDefaultEnvVarsWithProtocol(modelSvc, protocol.UpiV1).ToKubernetesEnvVars(),
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
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Ports:         grpcContainerPorts,
									Env:           []corev1.EnvVar{},
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
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Image:     "gcr.io/custom-model:v0.1",
									Env:       createDefaultPredictorEnvVars(modelSvc).ToKubernetesEnvVars(),
									Resources: expDefaultModelResourceRequests,
									Ports:     grpcContainerPorts,
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
				DefaultModelResourceRequests:       defaultModelResourceRequests,
				DefaultTransformerResourceRequests: defaultTransformerResourceRequests,
				QueueResourcePercentage:            tt.resourcePercentage,
			}

			tpl := NewInferenceServiceTemplater(standardTransformerConfig)
			infSvcSpec, err := tpl.CreateInferenceServiceSpec(tt.modelSvc, deployConfig)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.exp, infSvcSpec)
		})
	}
}

func TestCreateInferenceServiceSpecWithTransformer(t *testing.T) {
	err := models.InitKubernetesLabeller("gojek.com/", environmentName)
	assert.NoError(t, err)

	defer func() {
		_ = models.InitKubernetesLabeller("", "")
		assert.NoError(t, err)
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
		name     string
		modelSvc *models.Service
		exp      *kservev1beta1.InferenceService
		wantErr  bool
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
						{Name: envTransformerPort, Value: "1234"},                                      // should be replace by default
						{Name: envTransformerModelName, Value: "model-1234"},                           // should be replace by default
						{Name: envTransformerPredictURL, Value: "model-112-predictor-default.project"}, // should be replace by default
					},
				},
				Protocol: protocol.HttpJson,
			},
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Env:           []corev1.EnvVar{},
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
									Env:           createDefaultTransformerEnvVars(modelSvc).ToKubernetesEnvVars(),
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
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Env:           []corev1.EnvVar{},
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
									Env:           createDefaultTransformerEnvVars(modelSvc).ToKubernetesEnvVars(),
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
						{Name: envTransformerPort, Value: "1234"},                                      // should be replace by default
						{Name: envTransformerModelName, Value: "model-1234"},                           // should be replace by default
						{Name: envTransformerPredictURL, Value: "model-112-predictor-default.project"}, // should be replace by default
					},
				},
				Protocol: protocol.UpiV1,
			},
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Env:           []corev1.EnvVar{},
									Ports:         grpcContainerPorts,
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
									Name:          "transformer",
									Image:         "ghcr.io/gojek/merlin-transformer-test",
									Command:       []string{"python"},
									Args:          []string{"main.py"},
									Env:           createDefaultTransformerEnvVars(modelSvcGRPC).ToKubernetesEnvVars(),
									Resources:     expDefaultTransformerResourceRequests,
									LivenessProbe: transformerProbeConfigUPI,
									Ports:         grpcContainerPorts,
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
							Name:  transformer.StandardTransformerConfigEnvName,
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
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Env:           []corev1.EnvVar{},
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
									Env: models.MergeEnvVars(models.EnvVars{
										{
											Name:  transformer.DefaultFeastSource,
											Value: standardTransformerConfig.DefaultFeastSource.String(),
										},
										{
											Name:  transformer.FeastStorageConfigs,
											Value: `{"1":{"redisCluster":{"feastServingUrl":"localhost:6866","redisAddress":["10.1.1.2","10.1.1.3"],"option":{"poolSize":5,"minIdleConnections":2}}},"2":{"bigtable":{"feastServingUrl":"localhost:6867","project":"gcp-project","instance":"instance","appProfile":"default","option":{"grpcConnectionPool":4,"keepAliveInterval":"120s","keepAliveTimeout":"60s","credentialJson":"eyJrZXkiOiJ2YWx1ZSJ9"}}}}`,
										},
										{Name: transformer.JaegerAgentHost, Value: standardTransformerConfig.Jaeger.AgentHost},
										{Name: transformer.JaegerAgentPort, Value: standardTransformerConfig.Jaeger.AgentPort},
										{Name: transformer.JaegerSamplerParam, Value: standardTransformerConfig.Jaeger.SamplerParam},
										{Name: transformer.JaegerSamplerType, Value: standardTransformerConfig.Jaeger.SamplerType},
										{Name: transformer.JaegerDisabled, Value: standardTransformerConfig.Jaeger.Disabled},
										{Name: transformer.StandardTransformerConfigEnvName, Value: `{"standard_transformer":null}`},
									}, createDefaultTransformerEnvVars(modelSvc)).ToKubernetesEnvVars(),
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
			name: "standard transformer upi v1",
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
							Name:  transformer.StandardTransformerConfigEnvName,
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
				Protocol: protocol.UpiV1,
			},
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Env:           []corev1.EnvVar{},
									LivenessProbe: probeConfigUPI,
									Ports:         grpcContainerPorts,
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
									Env: models.MergeEnvVars(models.EnvVars{
										{
											Name:  transformer.DefaultFeastSource,
											Value: standardTransformerConfig.DefaultFeastSource.String(),
										},
										{
											Name:  transformer.FeastStorageConfigs,
											Value: `{"1":{"redisCluster":{"feastServingUrl":"localhost:6866","redisAddress":["10.1.1.2","10.1.1.3"],"option":{"poolSize":5,"minIdleConnections":2}}},"2":{"bigtable":{"feastServingUrl":"localhost:6867","project":"gcp-project","instance":"instance","appProfile":"default","option":{"grpcConnectionPool":4,"keepAliveInterval":"120s","keepAliveTimeout":"60s","credentialJson":"eyJrZXkiOiJ2YWx1ZSJ9"}}}}`,
										},
										{Name: transformer.JaegerAgentHost, Value: standardTransformerConfig.Jaeger.AgentHost},
										{Name: transformer.JaegerAgentPort, Value: standardTransformerConfig.Jaeger.AgentPort},
										{Name: transformer.JaegerSamplerParam, Value: standardTransformerConfig.Jaeger.SamplerParam},
										{Name: transformer.JaegerSamplerType, Value: standardTransformerConfig.Jaeger.SamplerType},
										{Name: transformer.JaegerDisabled, Value: standardTransformerConfig.Jaeger.Disabled},
										{Name: transformer.StandardTransformerConfigEnvName, Value: `{"standard_transformer":null}`},
									}, createDefaultTransformerEnvVars(modelSvcGRPC)).ToKubernetesEnvVars(),
									Resources:     expDefaultTransformerResourceRequests,
									Ports:         grpcContainerPorts,
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
				DefaultModelResourceRequests:       defaultModelResourceRequests,
				DefaultTransformerResourceRequests: defaultTransformerResourceRequests,
				QueueResourcePercentage:            queueResourcePercentage,
			}

			tpl := NewInferenceServiceTemplater(standardTransformerConfig)
			infSvcSpec, err := tpl.CreateInferenceServiceSpec(tt.modelSvc, deployConfig)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.exp, infSvcSpec)
		})
	}
}

func TestCreateInferenceServiceSpecWithLogger(t *testing.T) {
	err := models.InitKubernetesLabeller("gojek.com/", environmentName)
	assert.NoError(t, err)

	defer func() {
		_ = models.InitKubernetesLabeller("", "")
		assert.NoError(t, err)
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
		name     string
		modelSvc *models.Service
		exp      *kservev1beta1.InferenceService
		wantErr  bool
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
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Env:           []corev1.EnvVar{},
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
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Env:           []corev1.EnvVar{},
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
									Env:           createDefaultTransformerEnvVars(modelSvc).ToKubernetesEnvVars(),
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
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Env:           []corev1.EnvVar{},
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
									Env:           createDefaultTransformerEnvVars(modelSvc).ToKubernetesEnvVars(),
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
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Env:           []corev1.EnvVar{},
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
									Env:           createDefaultTransformerEnvVars(modelSvc).ToKubernetesEnvVars(),
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
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Env:           []corev1.EnvVar{},
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
									Env:           createDefaultTransformerEnvVars(modelSvc).ToKubernetesEnvVars(),
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
				DefaultModelResourceRequests:       defaultModelResourceRequests,
				DefaultTransformerResourceRequests: defaultTransformerResourceRequests,
				QueueResourcePercentage:            queueResourcePercentage,
			}

			tpl := NewInferenceServiceTemplater(standardTransformerConfig)
			infSvcSpec, err := tpl.CreateInferenceServiceSpec(tt.modelSvc, deployConfig)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.exp, infSvcSpec)
		})
	}
}

func TestPatchInferenceServiceSpec(t *testing.T) {
	err := models.InitKubernetesLabeller("gojek.com/", environmentName)
	assert.NoError(t, err)

	defer func() {
		_ = models.InitKubernetesLabeller("", "")
		assert.NoError(t, err)
	}()

	project := mlp.Project{
		Name: "project",
	}

	modelSvc := &models.Service{
		Name:         "model-1",
		ModelName:    "model",
		ModelVersion: "1",
		Namespace:    project.Name,
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

	storageUri := fmt.Sprintf("%s/model", modelSvc.ArtifactURI)

	// Liveness probe config for the model containers
	probeConfig := createLivenessProbeSpec(protocol.HttpJson, fmt.Sprintf("/v1/models/%s", modelSvc.Name))

	// Liveness probe config for the transformers
	transformerProbeConfig := createLivenessProbeSpec(protocol.HttpJson, "/")

	one := 1
	minReplica := 1
	maxReplica := 10
	cpuRequest := resource.MustParse("1")
	memoryRequest := resource.MustParse("1Gi")
	cpuLimit := cpuRequest.DeepCopy()
	cpuLimit.Add(cpuRequest)
	memoryLimit := memoryRequest.DeepCopy()
	memoryLimit.Add(memoryRequest)
	queueResourcePercentage := "2"

	resourceRequests := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    cpuRequest,
			corev1.ResourceMemory: memoryRequest,
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    cpuLimit,
			corev1.ResourceMemory: memoryLimit,
		},
	}

	tests := []struct {
		name     string
		modelSvc *models.Service
		original *kservev1beta1.InferenceService
		exp      *kservev1beta1.InferenceService
		wantErr  bool
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
			original: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						Tensorflow: &kservev1beta1.TFServingSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     resourceRequests,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &minReplica,
							MaxReplicas: maxReplica,
						},
					},
				},
			},
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Resources:     resourceRequests,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &minReplica,
							MaxReplicas: maxReplica,
						},
					},
				},
			},
		},
		{
			name: "tensorflow + transformer spec",
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
					ResourceRequest: &models.ResourceRequest{
						MinReplica:    1,
						MaxReplica:    1,
						CPURequest:    cpuRequest,
						MemoryRequest: memoryRequest,
					},
				},
				Protocol: protocol.HttpJson,
			},
			original: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						Tensorflow: &kservev1beta1.TFServingSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     resourceRequests,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &minReplica,
							MaxReplicas: maxReplica,
						},
					},
				},
			},
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Resources:     resourceRequests,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &minReplica,
							MaxReplicas: maxReplica,
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
									Env:     createDefaultTransformerEnvVars(modelSvc).ToKubernetesEnvVars(),
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
			},
		},
		{
			name: "tensorflow + transformer spec top tensorflow spec only",
			modelSvc: &models.Service{
				Name:        modelSvc.Name,
				Namespace:   project.Name,
				ArtifactURI: modelSvc.ArtifactURI,
				Type:        models.ModelTypeTensorflow,
				Options:     &models.ModelOption{},
				Metadata:    modelSvc.Metadata,
				Transformer: &models.Transformer{
					Enabled: false,
				},
				Protocol: protocol.HttpJson,
			},
			original: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						Tensorflow: &kservev1beta1.TFServingSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     resourceRequests,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &minReplica,
							MaxReplicas: maxReplica,
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
									Env: []corev1.EnvVar{
										{Name: envTransformerPort, Value: fmt.Sprint(defaultHTTPPort)},
										{Name: envTransformerModelName, Value: "model-1"},
										{Name: envTransformerPredictURL, Value: "model-1-predictor-default.project"},
									},
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
			},
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Resources:     resourceRequests,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &minReplica,
							MaxReplicas: maxReplica,
						},
					},
					Transformer: nil,
				},
			},
		},
		{
			name: "custom spec",
			modelSvc: &models.Service{
				Name:         modelSvc.Name,
				ModelName:    modelSvc.ModelName,
				ModelVersion: modelSvc.ModelVersion,
				Namespace:    project.Name,
				ArtifactURI:  modelSvc.ArtifactURI,
				Type:         models.ModelTypeCustom,
				Options: &models.ModelOption{
					CustomPredictor: &models.CustomPredictor{
						Image:   "gcr.io/custom-model:v0.2",
						Command: "./run-1.sh",
						Args:    "firstArg secondArg",
					},
				},
				Metadata:        modelSvc.Metadata,
				ResourceRequest: userResourceRequests,
				Protocol:        protocol.HttpJson,
			},
			original: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Image:     "gcr.io/custom-model:v0.1",
									Env:       createDefaultPredictorEnvVars(modelSvc).ToKubernetesEnvVars(),
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
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Image:     "gcr.io/custom-model:v0.2",
									Env:       createDefaultPredictorEnvVars(modelSvc).ToKubernetesEnvVars(),
									Resources: expUserResourceRequests,
									Command: []string{
										"./run-1.sh",
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
			name: "patch deployment mode from serverless to raw_deployment",
			modelSvc: &models.Service{
				Name:           modelSvc.Name,
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
			original: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						Tensorflow: &kservev1beta1.TFServingSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     resourceRequests,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &minReplica,
							MaxReplicas: maxReplica,
						},
					},
				},
			},
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
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Resources:     resourceRequests,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &minReplica,
							MaxReplicas: maxReplica,
						},
					},
				},
			},
		},
		{
			name: "patch deployment mode from raw_deployment to serverless_deployment",
			modelSvc: &models.Service{
				Name:           modelSvc.Name,
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
			original: &kservev1beta1.InferenceService{
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
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						Tensorflow: &kservev1beta1.TFServingSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:          kserveconstant.InferenceServiceContainerName,
									Resources:     resourceRequests,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &minReplica,
							MaxReplicas: maxReplica,
						},
					},
				},
			},
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      modelSvc.Name,
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.ClassAnnotationKey:                      knautoscaling.KPA,
						knautoscaling.MetricAnnotationKey:                     knautoscaling.Concurrency,
						knautoscaling.TargetAnnotationKey:                     "2",
					},
					Labels: map[string]string{
						"gojek.com/app":          modelSvc.Metadata.App,
						"gojek.com/component":    models.ComponentModelVersion,
						"gojek.com/environment":  environmentName,
						"gojek.com/orchestrator": orchestratorName,
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
									Resources:     resourceRequests,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &minReplica,
							MaxReplicas: maxReplica,
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deployConfig := &config.DeploymentConfig{
				DefaultModelResourceRequests: &config.ResourceRequests{
					MinReplica:    minReplica,
					MaxReplica:    maxReplica,
					CPURequest:    cpuRequest,
					MemoryRequest: memoryRequest,
				},
				QueueResourcePercentage: queueResourcePercentage,
			}

			tpl := NewInferenceServiceTemplater(standardTransformerConfig)
			infSvcSpec, err := tpl.PatchInferenceServiceSpec(tt.original, tt.modelSvc, deployConfig)
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
	cpuLimit := cpuRequest.DeepCopy()
	cpuLimit.Add(cpuRequest)
	memoryLimit := memoryRequest.DeepCopy()
	memoryLimit.Add(memoryRequest)

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
						{Name: transformer.JaegerAgentHost, Value: "NEW_HOST"}, // test user overwrite
					},
				},
				&config.DeploymentConfig{},
			},
			&kservev1beta1.TransformerSpec{
				PodSpec: kservev1beta1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "transformer",
							Image:   standardTransformerConfig.ImageName,
							Command: []string{"python"},
							Args:    []string{"main.py"},
							Env: models.MergeEnvVars(models.EnvVars{
								{Name: transformer.DefaultFeastSource, Value: standardTransformerConfig.DefaultFeastSource.String()},
								{
									Name:  transformer.FeastStorageConfigs,
									Value: `{"1":{"redisCluster":{"feastServingUrl":"localhost:6866","redisAddress":["10.1.1.2","10.1.1.3"],"option":{"poolSize":5,"minIdleConnections":2}}},"2":{"bigtable":{"feastServingUrl":"localhost:6867","project":"gcp-project","instance":"instance","appProfile":"default","option":{"grpcConnectionPool":4,"keepAliveInterval":"120s","keepAliveTimeout":"60s","credentialJson":"eyJrZXkiOiJ2YWx1ZSJ9"}}}}`,
								},
								{Name: transformer.JaegerAgentHost, Value: "NEW_HOST"},
								{Name: transformer.JaegerAgentPort, Value: standardTransformerConfig.Jaeger.AgentPort},
								{Name: transformer.JaegerSamplerParam, Value: standardTransformerConfig.Jaeger.SamplerParam},
								{Name: transformer.JaegerSamplerType, Value: standardTransformerConfig.Jaeger.SamplerType},
								{Name: transformer.JaegerDisabled, Value: standardTransformerConfig.Jaeger.Disabled},
							}, createDefaultTransformerEnvVars(modelSvc)).ToKubernetesEnvVars(),
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
				&config.DeploymentConfig{},
			},
			&kservev1beta1.TransformerSpec{
				PodSpec: kservev1beta1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "transformer",
							Image:   "ghcr.io/gojek/merlin-transformer-test",
							Command: []string{"python"},
							Args:    []string{"main.py"},
							Env:     createDefaultTransformerEnvVars(modelSvc).ToKubernetesEnvVars(),
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
			tpl := NewInferenceServiceTemplater(standardTransformerConfig)
			got := tpl.createTransformerSpec(tt.args.modelService, tt.args.transformer, tt.args.config)
			assert.Equal(t, tt.want, got)
		})
	}
}

func getLimit(quantity resource.Quantity) resource.Quantity {
	limit := quantity.DeepCopy()
	limit.Add(quantity)
	return limit
}

func createPyFuncDefaultEnvVarsWithProtocol(svc *models.Service, protocolValue protocol.Protocol) models.EnvVars {
	envVars := models.EnvVars{
		models.EnvVar{
			Name:  envPyFuncModelName,
			Value: models.CreateInferenceServiceName(svc.ModelName, svc.ModelVersion),
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
			Value: models.CreateInferenceServiceName(svc.ModelName, svc.ModelVersion),
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
	}
	return envVars
}
