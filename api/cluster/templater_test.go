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

package cluster

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/intstr"
	"testing"
	"time"

	"github.com/gojek/merlin/pkg/autoscaling"
	"github.com/gojek/merlin/pkg/deployment"
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
	"github.com/gojek/merlin/utils"
)

var (
	defaultModelResourceRequests = &config.ResourceRequests{
		MinReplica:    1,
		MaxReplica:    2,
		CPURequest:    resource.MustParse("500m"),
		MemoryRequest: resource.MustParse("100Mi"),
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
	project := mlp.Project{
		Name: "project",
	}

	model := &models.Service{
		Name:        "model",
		ArtifactURI: "gs://my-artifacet",
		Metadata: models.Metadata{
			Team:        "dsp",
			Stream:      "dsp",
			App:         "model",
			Environment: "dev",
			Labels: mlp.Labels{
				{
					Key:   "sample",
					Value: "true",
				},
			},
		},
	}
	versionID := 1
	queueResourcePercentage := "2"
	storageUri := fmt.Sprintf("%s/model", model.ArtifactURI)

	// Liveness probe config for the model containers
	probeConfig := &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path:   fmt.Sprintf("/v1/models/%s-%d", model.Name, versionID),
				Scheme: "HTTP",
				Port: intstr.IntOrString{
					IntVal: 8080,
				},
			},
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      5,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
	}

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
				Name:        models.CreateInferenceServiceName(model.Name, "1"),
				Namespace:   project.Name,
				ArtifactURI: model.ArtifactURI,
				Type:        models.ModelTypeTensorflow,
				Options:     &models.ModelOption{},
				Metadata:    model.Metadata,
			},
			resourcePercentage: queueResourcePercentage,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          model.Metadata.App,
						"gojek.com/orchestrator": "merlin",
						"gojek.com/stream":       model.Metadata.Stream,
						"gojek.com/team":         model.Metadata.Team,
						"gojek.com/sample":       "true",
						"gojek.com/environment":  model.Metadata.Environment,
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
				},
			},
		},
		{
			name: "tensorflow spec as raw deployment",
			modelSvc: &models.Service{
				Name:           models.CreateInferenceServiceName(model.Name, "1"),
				Namespace:      project.Name,
				ArtifactURI:    model.ArtifactURI,
				Type:           models.ModelTypeTensorflow,
				Options:        &models.ModelOption{},
				Metadata:       model.Metadata,
				DeploymentMode: deployment.RawDeploymentMode,
			},
			resourcePercentage: queueResourcePercentage,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.RawDeployment),
					},
					Labels: map[string]string{
						"gojek.com/app":          model.Metadata.App,
						"gojek.com/orchestrator": "merlin",
						"gojek.com/stream":       model.Metadata.Stream,
						"gojek.com/team":         model.Metadata.Team,
						"gojek.com/sample":       "true",
						"gojek.com/environment":  model.Metadata.Environment,
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
				},
			},
		},
		{
			name: "tensorflow spec as serverless",
			modelSvc: &models.Service{
				Name:           models.CreateInferenceServiceName(model.Name, "1"),
				Namespace:      project.Name,
				ArtifactURI:    model.ArtifactURI,
				Type:           models.ModelTypeTensorflow,
				Options:        &models.ModelOption{},
				Metadata:       model.Metadata,
				DeploymentMode: deployment.ServerlessDeploymentMode,
			},
			resourcePercentage: queueResourcePercentage,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          model.Metadata.App,
						"gojek.com/orchestrator": "merlin",
						"gojek.com/stream":       model.Metadata.Stream,
						"gojek.com/team":         model.Metadata.Team,
						"gojek.com/sample":       "true",
						"gojek.com/environment":  model.Metadata.Environment,
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
				},
			},
		},
		{
			name: "tensorflow spec without queue resource percentage",
			modelSvc: &models.Service{
				Name:        models.CreateInferenceServiceName(model.Name, "1"),
				Namespace:   project.Name,
				ArtifactURI: model.ArtifactURI,
				Type:        models.ModelTypeTensorflow,
				Options:     &models.ModelOption{},
				Metadata:    model.Metadata,
			},
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						kserveconstant.DeploymentMode: string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          model.Metadata.App,
						"gojek.com/orchestrator": "merlin",
						"gojek.com/stream":       model.Metadata.Stream,
						"gojek.com/team":         model.Metadata.Team,
						"gojek.com/sample":       "true",
						"gojek.com/environment":  model.Metadata.Environment,
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
				},
			},
		},
		{
			name: "xgboost spec",
			modelSvc: &models.Service{
				Name:        models.CreateInferenceServiceName(model.Name, "1"),
				Namespace:   project.Name,
				ArtifactURI: model.ArtifactURI,
				Type:        models.ModelTypeXgboost,
				Options:     &models.ModelOption{},
				Metadata:    model.Metadata,
			},
			resourcePercentage: queueResourcePercentage,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          model.Metadata.App,
						"gojek.com/orchestrator": "merlin",
						"gojek.com/stream":       model.Metadata.Stream,
						"gojek.com/team":         model.Metadata.Team,
						"gojek.com/sample":       "true",
						"gojek.com/environment":  model.Metadata.Environment,
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
				Name:        models.CreateInferenceServiceName(model.Name, "1"),
				Namespace:   project.Name,
				ArtifactURI: model.ArtifactURI,
				Type:        models.ModelTypeSkLearn,
				Options:     &models.ModelOption{},
				Metadata:    model.Metadata,
			},
			resourcePercentage: queueResourcePercentage,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          model.Metadata.App,
						"gojek.com/orchestrator": "merlin",
						"gojek.com/stream":       model.Metadata.Stream,
						"gojek.com/team":         model.Metadata.Team,
						"gojek.com/sample":       "true",
						"gojek.com/environment":  model.Metadata.Environment,
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
				Name:        models.CreateInferenceServiceName(model.Name, "1"),
				Namespace:   project.Name,
				ArtifactURI: model.ArtifactURI,
				Type:        models.ModelTypePyTorch,
				Metadata:    model.Metadata,
			},
			resourcePercentage: queueResourcePercentage,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          model.Metadata.App,
						"gojek.com/orchestrator": "merlin",
						"gojek.com/stream":       model.Metadata.Stream,
						"gojek.com/team":         model.Metadata.Team,
						"gojek.com/sample":       "true",
						"gojek.com/environment":  model.Metadata.Environment,
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
			name: "onnx spec",
			modelSvc: &models.Service{
				Name:        models.CreateInferenceServiceName(model.Name, "1"),
				Namespace:   project.Name,
				ArtifactURI: model.ArtifactURI,
				Type:        models.ModelTypeOnnx,
				Options:     &models.ModelOption{},
				Metadata:    model.Metadata,
			},
			resourcePercentage: queueResourcePercentage,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          model.Metadata.App,
						"gojek.com/orchestrator": "merlin",
						"gojek.com/stream":       model.Metadata.Stream,
						"gojek.com/team":         model.Metadata.Team,
						"gojek.com/sample":       "true",
						"gojek.com/environment":  model.Metadata.Environment,
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						ONNX: &kservev1beta1.ONNXRuntimeSpec{
							PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
								StorageURI: &storageUri,
								Container: corev1.Container{
									Name:      kserveconstant.InferenceServiceContainerName,
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
			name: "pyfunc spec",
			modelSvc: &models.Service{
				Name:        models.CreateInferenceServiceName(model.Name, "1"),
				Namespace:   project.Name,
				ArtifactURI: model.ArtifactURI,
				Type:        models.ModelTypePyFunc,
				Options: &models.ModelOption{
					PyFuncImageName: "gojek/project-model:1",
				},
				EnvVars:  models.PyfuncDefaultEnvVars(models.Model{Name: model.Name}, models.Version{ID: models.ID(1), ArtifactURI: model.ArtifactURI}, defaultModelResourceRequests.CPURequest.Value()),
				Metadata: model.Metadata,
			},
			resourcePercentage: queueResourcePercentage,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						"prometheus.io/scrape":                                "true",
						"prometheus.io/port":                                  "8080",
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          model.Metadata.App,
						"gojek.com/orchestrator": "merlin",
						"gojek.com/stream":       model.Metadata.Stream,
						"gojek.com/team":         model.Metadata.Team,
						"gojek.com/sample":       "true",
						"gojek.com/environment":  model.Metadata.Environment,
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						PodSpec: kservev1beta1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:          kserveconstant.InferenceServiceContainerName,
									Image:         "gojek/project-model:1",
									Env:           models.PyfuncDefaultEnvVars(models.Model{Name: model.Name}, models.Version{ID: models.ID(1), ArtifactURI: model.ArtifactURI}, defaultModelResourceRequests.CPURequest.Value()).ToKubernetesEnvVars(),
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
			name: "tensorflow spec with user resource request",
			modelSvc: &models.Service{
				Name:            models.CreateInferenceServiceName(model.Name, "1"),
				Namespace:       project.Name,
				ArtifactURI:     model.ArtifactURI,
				Type:            models.ModelTypeTensorflow,
				Options:         &models.ModelOption{},
				Metadata:        model.Metadata,
				ResourceRequest: userResourceRequests,
			},
			resourcePercentage: queueResourcePercentage,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          model.Metadata.App,
						"gojek.com/orchestrator": "merlin",
						"gojek.com/stream":       model.Metadata.Stream,
						"gojek.com/team":         model.Metadata.Team,
						"gojek.com/sample":       "true",
						"gojek.com/environment":  model.Metadata.Environment,
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
				Name:        models.CreateInferenceServiceName(model.Name, "1"),
				Namespace:   project.Name,
				ArtifactURI: model.ArtifactURI,
				Type:        models.ModelTypeCustom,
				Options: &models.ModelOption{
					CustomPredictor: &models.CustomPredictor{
						Image: "gcr.io/custom-model:v0.1",
					},
				},
				Metadata: model.Metadata,
			},
			resourcePercentage: queueResourcePercentage,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          model.Metadata.App,
						"gojek.com/orchestrator": "merlin",
						"gojek.com/stream":       model.Metadata.Stream,
						"gojek.com/team":         model.Metadata.Team,
						"gojek.com/sample":       "true",
						"gojek.com/environment":  model.Metadata.Environment,
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						PodSpec: kservev1beta1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  kserveconstant.InferenceServiceContainerName,
									Image: "gcr.io/custom-model:v0.1",
									Env: models.EnvVars{
										models.EnvVar{Name: "MERLIN_PREDICTOR_PORT", Value: "8080"},
										models.EnvVar{Name: "MERLIN_MODEL_NAME", Value: models.CreateInferenceServiceName(model.Name, "1")},
										models.EnvVar{Name: "MERLIN_ARTIFACT_LOCATION", Value: "/mnt/models"},
										models.EnvVar{Name: "STORAGE_URI", Value: utils.CreateModelLocation(model.ArtifactURI)},
									}.ToKubernetesEnvVars(),
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
				Name:        models.CreateInferenceServiceName(model.Name, "1"),
				Namespace:   project.Name,
				ArtifactURI: model.ArtifactURI,
				Type:        models.ModelTypeCustom,
				Options: &models.ModelOption{
					CustomPredictor: &models.CustomPredictor{
						Image:   "gcr.io/custom-model:v0.1",
						Command: "./run.sh",
						Args:    "firstArg secondArg",
					},
				},
				Metadata:        model.Metadata,
				ResourceRequest: userResourceRequests,
			},
			resourcePercentage: queueResourcePercentage,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          model.Metadata.App,
						"gojek.com/orchestrator": "merlin",
						"gojek.com/stream":       model.Metadata.Stream,
						"gojek.com/team":         model.Metadata.Team,
						"gojek.com/sample":       "true",
						"gojek.com/environment":  model.Metadata.Environment,
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						PodSpec: kservev1beta1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  kserveconstant.InferenceServiceContainerName,
									Image: "gcr.io/custom-model:v0.1",
									Env: models.EnvVars{
										models.EnvVar{Name: "MERLIN_PREDICTOR_PORT", Value: "8080"},
										models.EnvVar{Name: "MERLIN_MODEL_NAME", Value: models.CreateInferenceServiceName(model.Name, "1")},
										models.EnvVar{Name: "MERLIN_ARTIFACT_LOCATION", Value: "/mnt/models"},
										models.EnvVar{Name: "STORAGE_URI", Value: utils.CreateModelLocation(model.ArtifactURI)},
									}.ToKubernetesEnvVars(),
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
				Name:           models.CreateInferenceServiceName(model.Name, "1"),
				Namespace:      project.Name,
				ArtifactURI:    model.ArtifactURI,
				Type:           models.ModelTypeTensorflow,
				Options:        &models.ModelOption{},
				Metadata:       model.Metadata,
				DeploymentMode: deployment.RawDeploymentMode,
				AutoscalingPolicy: &autoscaling.AutoscalingPolicy{
					MetricsType: autoscaling.CPUUtilization,
					TargetValue: 30,
				},
			},
			resourcePercentage: queueResourcePercentage,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.RawDeployment),
						kserveconstant.AutoscalerClass:                        string(kserveconstant.AutoscalerClassHPA),
						kserveconstant.AutoscalerMetrics:                      string(kserveconstant.AutoScalerMetricsCPU),
						kserveconstant.TargetUtilizationPercentage:            "30",
					},
					Labels: map[string]string{
						"gojek.com/app":          model.Metadata.App,
						"gojek.com/orchestrator": "merlin",
						"gojek.com/stream":       model.Metadata.Stream,
						"gojek.com/team":         model.Metadata.Team,
						"gojek.com/sample":       "true",
						"gojek.com/environment":  model.Metadata.Environment,
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
				},
			},
		},
		{
			name: "raw deployment using not supported autoscaling metrics",
			modelSvc: &models.Service{
				Name:           models.CreateInferenceServiceName(model.Name, "1"),
				Namespace:      project.Name,
				ArtifactURI:    model.ArtifactURI,
				Type:           models.ModelTypeTensorflow,
				Options:        &models.ModelOption{},
				Metadata:       model.Metadata,
				DeploymentMode: deployment.RawDeploymentMode,
				AutoscalingPolicy: &autoscaling.AutoscalingPolicy{
					MetricsType: autoscaling.MemoryUtilization,
					TargetValue: 30,
				},
			},
			resourcePercentage: queueResourcePercentage,
			wantErr:            true,
		},
		{
			name: "serverless deployment using CPU autoscaling",
			modelSvc: &models.Service{
				Name:           models.CreateInferenceServiceName(model.Name, "1"),
				Namespace:      project.Name,
				ArtifactURI:    model.ArtifactURI,
				Type:           models.ModelTypeTensorflow,
				Options:        &models.ModelOption{},
				Metadata:       model.Metadata,
				DeploymentMode: deployment.ServerlessDeploymentMode,
				AutoscalingPolicy: &autoscaling.AutoscalingPolicy{
					MetricsType: autoscaling.CPUUtilization,
					TargetValue: 30,
				},
			},
			resourcePercentage: queueResourcePercentage,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.ClassAnnotationKey:                      knautoscaling.HPA,
						knautoscaling.MetricAnnotationKey:                     knautoscaling.CPU,
						knautoscaling.TargetAnnotationKey:                     "30",
					},
					Labels: map[string]string{
						"gojek.com/app":          model.Metadata.App,
						"gojek.com/orchestrator": "merlin",
						"gojek.com/stream":       model.Metadata.Stream,
						"gojek.com/team":         model.Metadata.Team,
						"gojek.com/sample":       "true",
						"gojek.com/environment":  model.Metadata.Environment,
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
				},
			},
		},
		{
			name: "serverless deployment using memory autoscaling",
			modelSvc: &models.Service{
				Name:           models.CreateInferenceServiceName(model.Name, "1"),
				Namespace:      project.Name,
				ArtifactURI:    model.ArtifactURI,
				Type:           models.ModelTypeTensorflow,
				Options:        &models.ModelOption{},
				Metadata:       model.Metadata,
				DeploymentMode: deployment.ServerlessDeploymentMode,
				AutoscalingPolicy: &autoscaling.AutoscalingPolicy{
					MetricsType: autoscaling.MemoryUtilization,
					TargetValue: 30,
				},
			},
			resourcePercentage: queueResourcePercentage,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.ClassAnnotationKey:                      knautoscaling.HPA,
						knautoscaling.MetricAnnotationKey:                     knautoscaling.Memory,
						knautoscaling.TargetAnnotationKey:                     "30",
					},
					Labels: map[string]string{
						"gojek.com/app":          model.Metadata.App,
						"gojek.com/orchestrator": "merlin",
						"gojek.com/stream":       model.Metadata.Stream,
						"gojek.com/team":         model.Metadata.Team,
						"gojek.com/sample":       "true",
						"gojek.com/environment":  model.Metadata.Environment,
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
				},
			},
		},
		{
			name: "serverless deployment using concurrency autoscaling",
			modelSvc: &models.Service{
				Name:           models.CreateInferenceServiceName(model.Name, "1"),
				Namespace:      project.Name,
				ArtifactURI:    model.ArtifactURI,
				Type:           models.ModelTypeTensorflow,
				Options:        &models.ModelOption{},
				Metadata:       model.Metadata,
				DeploymentMode: deployment.ServerlessDeploymentMode,
				AutoscalingPolicy: &autoscaling.AutoscalingPolicy{
					MetricsType: autoscaling.Concurrency,
					TargetValue: 2,
				},
			},
			resourcePercentage: queueResourcePercentage,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.ClassAnnotationKey:                      knautoscaling.KPA,
						knautoscaling.MetricAnnotationKey:                     knautoscaling.Concurrency,
						knautoscaling.TargetAnnotationKey:                     "2",
					},
					Labels: map[string]string{
						"gojek.com/app":          model.Metadata.App,
						"gojek.com/orchestrator": "merlin",
						"gojek.com/stream":       model.Metadata.Stream,
						"gojek.com/team":         model.Metadata.Team,
						"gojek.com/sample":       "true",
						"gojek.com/environment":  model.Metadata.Environment,
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
				},
			},
		},
		{
			name: "serverless deployment using rps autoscaling",
			modelSvc: &models.Service{
				Name:           models.CreateInferenceServiceName(model.Name, "1"),
				Namespace:      project.Name,
				ArtifactURI:    model.ArtifactURI,
				Type:           models.ModelTypeTensorflow,
				Options:        &models.ModelOption{},
				Metadata:       model.Metadata,
				DeploymentMode: deployment.ServerlessDeploymentMode,
				AutoscalingPolicy: &autoscaling.AutoscalingPolicy{
					MetricsType: autoscaling.RPS,
					TargetValue: 10,
				},
			},
			resourcePercentage: queueResourcePercentage,
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.ClassAnnotationKey:                      knautoscaling.KPA,
						knautoscaling.MetricAnnotationKey:                     knautoscaling.RPS,
						knautoscaling.TargetAnnotationKey:                     "10",
					},
					Labels: map[string]string{
						"gojek.com/app":          model.Metadata.App,
						"gojek.com/orchestrator": "merlin",
						"gojek.com/stream":       model.Metadata.Stream,
						"gojek.com/team":         model.Metadata.Team,
						"gojek.com/sample":       "true",
						"gojek.com/environment":  model.Metadata.Environment,
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

			tpl := NewKFServingResourceTemplater(standardTransformerConfig)
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
	project := mlp.Project{
		Name: "project",
	}

	loggerDestinationURL := "http://destination.default"
	model := &models.Service{
		Name:        "model",
		ArtifactURI: "gs://my-artifacet",
		Metadata: models.Metadata{
			Team:        "dsp",
			Stream:      "dsp",
			App:         "model",
			Environment: "dev",
			Labels: mlp.Labels{
				{
					Key:   "sample",
					Value: "true",
				},
			},
		},
	}
	versionID := 1
	queueResourcePercentage := "2"
	storageUri := fmt.Sprintf("%s/model", model.ArtifactURI)

	// Liveness probe config for the model containers
	probeConfig := &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path:   fmt.Sprintf("/v1/models/%s-%d", model.Name, versionID),
				Scheme: "HTTP",
				Port: intstr.IntOrString{
					IntVal: 8080,
				},
			},
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      5,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
	}

	tests := []struct {
		name     string
		modelSvc *models.Service
		exp      *kservev1beta1.InferenceService
		wantErr  bool
	}{
		{
			name: "custom transformer with default resource request",
			modelSvc: &models.Service{
				Name:        models.CreateInferenceServiceName(model.Name, "1"),
				Namespace:   project.Name,
				ArtifactURI: model.ArtifactURI,
				Type:        models.ModelTypeTensorflow,
				Options:     &models.ModelOption{},
				Metadata:    model.Metadata,
				Transformer: &models.Transformer{
					Enabled: true,
					Image:   "ghcr.io/gojek/merlin-transformer-test",
					Command: "python",
					Args:    "main.py",
				},
			},
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          model.Metadata.App,
						"gojek.com/orchestrator": "merlin",
						"gojek.com/stream":       model.Metadata.Stream,
						"gojek.com/team":         model.Metadata.Team,
						"gojek.com/sample":       "true",
						"gojek.com/environment":  model.Metadata.Environment,
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
									Name:    "transformer",
									Image:   "ghcr.io/gojek/merlin-transformer-test",
									Command: []string{"python"},
									Args:    []string{"main.py"},
									Env: []corev1.EnvVar{
										{Name: envTransformerPort, Value: defaultTransformerPort},
										{Name: envTransformerModelName, Value: "model-1"},
										{Name: envTransformerPredictURL, Value: "model-1-predictor-default.project"},
									},
									Resources: expDefaultTransformerResourceRequests,
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
				Name:        models.CreateInferenceServiceName(model.Name, "1"),
				Namespace:   project.Name,
				ArtifactURI: model.ArtifactURI,
				Type:        models.ModelTypeTensorflow,
				Options:     &models.ModelOption{},
				Metadata:    model.Metadata,
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
			},
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          model.Metadata.App,
						"gojek.com/orchestrator": "merlin",
						"gojek.com/stream":       model.Metadata.Stream,
						"gojek.com/team":         model.Metadata.Team,
						"gojek.com/sample":       "true",
						"gojek.com/environment":  model.Metadata.Environment,
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
									Name:    "transformer",
									Image:   "ghcr.io/gojek/merlin-transformer-test",
									Command: []string{"python"},
									Args:    []string{"main.py"},
									Env: []corev1.EnvVar{
										{Name: envTransformerPort, Value: defaultTransformerPort},
										{Name: envTransformerModelName, Value: "model-1"},
										{Name: envTransformerPredictURL, Value: "model-1-predictor-default.project"},
									},
									Resources: expUserResourceRequests,
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
			name: "standard transformer",
			modelSvc: &models.Service{
				Name:        models.CreateInferenceServiceName(model.Name, "1"),
				Namespace:   project.Name,
				ArtifactURI: model.ArtifactURI,
				Type:        models.ModelTypeTensorflow,
				Options:     &models.ModelOption{},
				Metadata:    model.Metadata,
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
			},
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          model.Metadata.App,
						"gojek.com/orchestrator": "merlin",
						"gojek.com/stream":       model.Metadata.Stream,
						"gojek.com/team":         model.Metadata.Team,
						"gojek.com/sample":       "true",
						"gojek.com/environment":  model.Metadata.Environment,
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
									Name:  "transformer",
									Image: standardTransformerConfig.ImageName,
									Env: []corev1.EnvVar{
										{Name: transformer.JaegerAgentHost, Value: standardTransformerConfig.Jaeger.AgentHost},
										{Name: transformer.JaegerAgentPort, Value: standardTransformerConfig.Jaeger.AgentPort},
										{Name: transformer.JaegerSamplerParam, Value: standardTransformerConfig.Jaeger.SamplerParam},
										{Name: transformer.JaegerSamplerType, Value: standardTransformerConfig.Jaeger.SamplerType},
										{Name: transformer.JaegerDisabled, Value: standardTransformerConfig.Jaeger.Disabled},
										{Name: transformer.StandardTransformerConfigEnvName, Value: `{"standard_transformer":null}`},
										{
											Name:  transformer.DefaultFeastSource,
											Value: standardTransformerConfig.DefaultFeastSource.String(),
										},
										{
											Name:  transformer.FeastStorageConfigs,
											Value: `{"1":{"redisCluster":{"feastServingUrl":"localhost:6866","redisAddress":["10.1.1.2","10.1.1.3"],"option":{"poolSize":5,"minIdleConnections":2}}},"2":{"bigtable":{"feastServingUrl":"localhost:6867","project":"gcp-project","instance":"instance","appProfile":"default","option":{"grpcConnectionPool":4,"keepAliveInterval":"120s","keepAliveTimeout":"60s","credentialJson":"eyJrZXkiOiJ2YWx1ZSJ9"}}}}`,
										},
										{Name: envTransformerPort, Value: defaultTransformerPort},
										{Name: envTransformerModelName, Value: "model-1"},
										{Name: envTransformerPredictURL, Value: "model-1-predictor-default.project"},
									},
									Resources: expDefaultTransformerResourceRequests,
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

			tpl := NewKFServingResourceTemplater(standardTransformerConfig)
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
	project := mlp.Project{
		Name: "project",
	}

	loggerDestinationURL := "http://destination.default"
	model := &models.Service{
		Name:        "model",
		ArtifactURI: "gs://my-artifacet",
		Metadata: models.Metadata{
			Team:        "dsp",
			Stream:      "dsp",
			App:         "model",
			Environment: "dev",
			Labels: mlp.Labels{
				{
					Key:   "sample",
					Value: "true",
				},
			},
		},
	}
	versionID := 1
	queueResourcePercentage := "2"
	storageUri := fmt.Sprintf("%s/model", model.ArtifactURI)

	// Liveness probe config for the model containers
	probeConfig := &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path:   fmt.Sprintf("/v1/models/%s-%d", model.Name, versionID),
				Scheme: "HTTP",
				Port: intstr.IntOrString{
					IntVal: 8080,
				},
			},
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      5,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
	}

	tests := []struct {
		name     string
		modelSvc *models.Service
		exp      *kservev1beta1.InferenceService
		wantErr  bool
	}{
		{
			name: "model logger enabled",
			modelSvc: &models.Service{
				Name:        models.CreateInferenceServiceName(model.Name, "1"),
				Namespace:   project.Name,
				ArtifactURI: model.ArtifactURI,
				Type:        models.ModelTypeTensorflow,
				Options:     &models.ModelOption{},
				Metadata:    model.Metadata,
				Logger: &models.Logger{
					DestinationURL: loggerDestinationURL,
					Model: &models.LoggerConfig{
						Enabled: true,
						Mode:    models.LogAll,
					},
				},
			},
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          model.Metadata.App,
						"gojek.com/orchestrator": "merlin",
						"gojek.com/stream":       model.Metadata.Stream,
						"gojek.com/team":         model.Metadata.Team,
						"gojek.com/sample":       "true",
						"gojek.com/environment":  model.Metadata.Environment,
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
				Name:        models.CreateInferenceServiceName(model.Name, "1"),
				Namespace:   project.Name,
				ArtifactURI: model.ArtifactURI,
				Type:        models.ModelTypeTensorflow,
				Options:     &models.ModelOption{},
				Metadata:    model.Metadata,
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
			},
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          model.Metadata.App,
						"gojek.com/orchestrator": "merlin",
						"gojek.com/stream":       model.Metadata.Stream,
						"gojek.com/team":         model.Metadata.Team,
						"gojek.com/sample":       "true",
						"gojek.com/environment":  model.Metadata.Environment,
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
									Name:    "transformer",
									Image:   "ghcr.io/gojek/merlin-transformer-test",
									Command: []string{"python"},
									Args:    []string{"main.py"},
									Env: []corev1.EnvVar{
										{Name: envTransformerPort, Value: defaultTransformerPort},
										{Name: envTransformerModelName, Value: "model-1"},
										{Name: envTransformerPredictURL, Value: "model-1-predictor-default.project"},
									},
									Resources: expDefaultTransformerResourceRequests,
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
				Name:        models.CreateInferenceServiceName(model.Name, "1"),
				Namespace:   project.Name,
				ArtifactURI: model.ArtifactURI,
				Type:        models.ModelTypeTensorflow,
				Options:     &models.ModelOption{},
				Metadata:    model.Metadata,
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
			},
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          model.Metadata.App,
						"gojek.com/orchestrator": "merlin",
						"gojek.com/stream":       model.Metadata.Stream,
						"gojek.com/team":         model.Metadata.Team,
						"gojek.com/sample":       "true",
						"gojek.com/environment":  model.Metadata.Environment,
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
									Name:    "transformer",
									Image:   "ghcr.io/gojek/merlin-transformer-test",
									Command: []string{"python"},
									Args:    []string{"main.py"},
									Env: []corev1.EnvVar{
										{Name: envTransformerPort, Value: defaultTransformerPort},
										{Name: envTransformerModelName, Value: "model-1"},
										{Name: envTransformerPredictURL, Value: "model-1-predictor-default.project"},
									},
									Resources: expDefaultTransformerResourceRequests,
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
				Name:        models.CreateInferenceServiceName(model.Name, "1"),
				Namespace:   project.Name,
				ArtifactURI: model.ArtifactURI,
				Type:        models.ModelTypeTensorflow,
				Options:     &models.ModelOption{},
				Metadata:    model.Metadata,
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
			},
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          model.Metadata.App,
						"gojek.com/orchestrator": "merlin",
						"gojek.com/stream":       model.Metadata.Stream,
						"gojek.com/team":         model.Metadata.Team,
						"gojek.com/sample":       "true",
						"gojek.com/environment":  model.Metadata.Environment,
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
									Name:    "transformer",
									Image:   "ghcr.io/gojek/merlin-transformer-test",
									Command: []string{"python"},
									Args:    []string{"main.py"},
									Env: []corev1.EnvVar{
										{Name: envTransformerPort, Value: defaultTransformerPort},
										{Name: envTransformerModelName, Value: "model-1"},
										{Name: envTransformerPredictURL, Value: "model-1-predictor-default.project"},
									},
									Resources: expDefaultTransformerResourceRequests,
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
				Name:        models.CreateInferenceServiceName(model.Name, "1"),
				Namespace:   project.Name,
				ArtifactURI: model.ArtifactURI,
				Type:        models.ModelTypeTensorflow,
				Options:     &models.ModelOption{},
				Metadata:    model.Metadata,
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
			},
			exp: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          model.Metadata.App,
						"gojek.com/orchestrator": "merlin",
						"gojek.com/stream":       model.Metadata.Stream,
						"gojek.com/team":         model.Metadata.Team,
						"gojek.com/sample":       "true",
						"gojek.com/environment":  model.Metadata.Environment,
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
									Name:    "transformer",
									Image:   "ghcr.io/gojek/merlin-transformer-test",
									Command: []string{"python"},
									Args:    []string{"main.py"},
									Env: []corev1.EnvVar{
										{Name: envTransformerPort, Value: defaultTransformerPort},
										{Name: envTransformerModelName, Value: "model-1"},
										{Name: envTransformerPredictURL, Value: "model-1-predictor-default.project"},
									},
									Resources: expDefaultTransformerResourceRequests,
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

			tpl := NewKFServingResourceTemplater(standardTransformerConfig)
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
	project := mlp.Project{
		Name: "project",
	}

	model := &models.Service{
		Name:        "model",
		ArtifactURI: "gs://my-artifacet",
		Metadata: models.Metadata{
			Team:        "dsp",
			Stream:      "dsp",
			App:         "model",
			Environment: "dev",
			Labels: mlp.Labels{
				{
					Key:   "sample",
					Value: "true",
				},
			},
		},
	}
	versionID := 1
	storageUri := fmt.Sprintf("%s/model", model.ArtifactURI)

	// Liveness probe config for the model containers
	probeConfig := &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path:   fmt.Sprintf("/v1/models/%s-%d", model.Name, versionID),
				Scheme: "HTTP",
				Port: intstr.IntOrString{
					IntVal: 8080,
				},
			},
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      5,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
	}

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
				Name:        models.CreateInferenceServiceName(model.Name, "1"),
				Namespace:   project.Name,
				ArtifactURI: model.ArtifactURI,
				Type:        models.ModelTypeTensorflow,
				Options:     &models.ModelOption{},
				Metadata:    model.Metadata,
			},
			original: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
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
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          model.Metadata.App,
						"gojek.com/orchestrator": "merlin",
						"gojek.com/stream":       model.Metadata.Stream,
						"gojek.com/team":         model.Metadata.Team,
						"gojek.com/sample":       "true",
						"gojek.com/environment":  model.Metadata.Environment,
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
				Name:        models.CreateInferenceServiceName(model.Name, "1"),
				Namespace:   project.Name,
				ArtifactURI: model.ArtifactURI,
				Type:        models.ModelTypeTensorflow,
				Options:     &models.ModelOption{},
				Metadata:    model.Metadata,
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
			},
			original: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
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
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          model.Metadata.App,
						"gojek.com/orchestrator": "merlin",
						"gojek.com/stream":       model.Metadata.Stream,
						"gojek.com/team":         model.Metadata.Team,
						"gojek.com/sample":       "true",
						"gojek.com/environment":  model.Metadata.Environment,
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
										{Name: envTransformerPort, Value: defaultTransformerPort},
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
				Name:        models.CreateInferenceServiceName(model.Name, "1"),
				Namespace:   project.Name,
				ArtifactURI: model.ArtifactURI,
				Type:        models.ModelTypeTensorflow,
				Options:     &models.ModelOption{},
				Metadata:    model.Metadata,
				Transformer: &models.Transformer{
					Enabled: false,
				},
			},
			original: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
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
										{Name: envTransformerPort, Value: defaultTransformerPort},
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
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          model.Metadata.App,
						"gojek.com/orchestrator": "merlin",
						"gojek.com/stream":       model.Metadata.Stream,
						"gojek.com/team":         model.Metadata.Team,
						"gojek.com/sample":       "true",
						"gojek.com/environment":  model.Metadata.Environment,
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
				Name:        models.CreateInferenceServiceName(model.Name, "1"),
				Namespace:   project.Name,
				ArtifactURI: model.ArtifactURI,
				Type:        models.ModelTypeCustom,
				Options: &models.ModelOption{
					CustomPredictor: &models.CustomPredictor{
						Image:   "gcr.io/custom-model:v0.2",
						Command: "./run-1.sh",
						Args:    "firstArg secondArg",
					},
				},
				Metadata:        model.Metadata,
				ResourceRequest: userResourceRequests,
			},
			original: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          model.Metadata.App,
						"gojek.com/orchestrator": "merlin",
						"gojek.com/stream":       model.Metadata.Stream,
						"gojek.com/team":         model.Metadata.Team,
						"gojek.com/sample":       "true",
						"gojek.com/environment":  model.Metadata.Environment,
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						PodSpec: kservev1beta1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  kserveconstant.InferenceServiceContainerName,
									Image: "gcr.io/custom-model:v0.1",
									Env: models.EnvVars{
										models.EnvVar{Name: "MERLIN_PREDICTOR_PORT", Value: "8080"},
										models.EnvVar{Name: "MERLIN_MODEL_NAME", Value: models.CreateInferenceServiceName(model.Name, "1")},
										models.EnvVar{Name: "MERLIN_ARTIFACT_LOCATION", Value: "/mnt/models"},
										models.EnvVar{Name: "STORAGE_URI", Value: utils.CreateModelLocation(model.ArtifactURI)},
									}.ToKubernetesEnvVars(),
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
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
					},
					Labels: map[string]string{
						"gojek.com/app":          model.Metadata.App,
						"gojek.com/orchestrator": "merlin",
						"gojek.com/stream":       model.Metadata.Stream,
						"gojek.com/team":         model.Metadata.Team,
						"gojek.com/sample":       "true",
						"gojek.com/environment":  model.Metadata.Environment,
					},
				},
				Spec: kservev1beta1.InferenceServiceSpec{
					Predictor: kservev1beta1.PredictorSpec{
						PodSpec: kservev1beta1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  kserveconstant.InferenceServiceContainerName,
									Image: "gcr.io/custom-model:v0.2",
									Env: models.EnvVars{
										models.EnvVar{Name: "MERLIN_PREDICTOR_PORT", Value: "8080"},
										models.EnvVar{Name: "MERLIN_MODEL_NAME", Value: models.CreateInferenceServiceName(model.Name, "1")},
										models.EnvVar{Name: "MERLIN_ARTIFACT_LOCATION", Value: "/mnt/models"},
										models.EnvVar{Name: "STORAGE_URI", Value: utils.CreateModelLocation(model.ArtifactURI)},
									}.ToKubernetesEnvVars(),
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
				Name:           models.CreateInferenceServiceName(model.Name, "1"),
				Namespace:      project.Name,
				ArtifactURI:    model.ArtifactURI,
				Type:           models.ModelTypeTensorflow,
				Options:        &models.ModelOption{},
				Metadata:       model.Metadata,
				DeploymentMode: deployment.RawDeploymentMode,
				AutoscalingPolicy: &autoscaling.AutoscalingPolicy{
					MetricsType: autoscaling.CPUUtilization,
					TargetValue: 30,
				},
			},
			original: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
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
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.RawDeployment),
						kserveconstant.AutoscalerClass:                        string(kserveconstant.AutoscalerClassHPA),
						kserveconstant.AutoscalerMetrics:                      string(kserveconstant.AutoScalerMetricsCPU),
						kserveconstant.TargetUtilizationPercentage:            "30",
					},
					Labels: map[string]string{
						"gojek.com/app":          model.Metadata.App,
						"gojek.com/orchestrator": "merlin",
						"gojek.com/stream":       model.Metadata.Stream,
						"gojek.com/team":         model.Metadata.Team,
						"gojek.com/sample":       "true",
						"gojek.com/environment":  model.Metadata.Environment,
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
				Name:           models.CreateInferenceServiceName(model.Name, "1"),
				Namespace:      project.Name,
				ArtifactURI:    model.ArtifactURI,
				Type:           models.ModelTypeTensorflow,
				Options:        &models.ModelOption{},
				Metadata:       model.Metadata,
				DeploymentMode: deployment.ServerlessDeploymentMode,
				AutoscalingPolicy: &autoscaling.AutoscalingPolicy{
					MetricsType: autoscaling.Concurrency,
					TargetValue: 2,
				},
			},
			original: &kservev1beta1.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
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
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						knserving.QueueSidecarResourcePercentageAnnotationKey: queueResourcePercentage,
						kserveconstant.DeploymentMode:                         string(kserveconstant.Serverless),
						knautoscaling.ClassAnnotationKey:                      knautoscaling.KPA,
						knautoscaling.MetricAnnotationKey:                     knautoscaling.Concurrency,
						knautoscaling.TargetAnnotationKey:                     "2",
					},
					Labels: map[string]string{
						"gojek.com/app":          model.Metadata.App,
						"gojek.com/orchestrator": "merlin",
						"gojek.com/stream":       model.Metadata.Stream,
						"gojek.com/team":         model.Metadata.Team,
						"gojek.com/sample":       "true",
						"gojek.com/environment":  model.Metadata.Environment,
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

			tpl := NewKFServingResourceTemplater(standardTransformerConfig)
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
					Name:      "test-1",
					Namespace: "test",
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
							Env: []corev1.EnvVar{
								{Name: transformer.JaegerAgentHost, Value: standardTransformerConfig.Jaeger.AgentHost},
								{Name: transformer.JaegerAgentPort, Value: standardTransformerConfig.Jaeger.AgentPort},
								{Name: transformer.JaegerSamplerParam, Value: standardTransformerConfig.Jaeger.SamplerParam},
								{Name: transformer.JaegerSamplerType, Value: standardTransformerConfig.Jaeger.SamplerType},
								{Name: transformer.JaegerDisabled, Value: standardTransformerConfig.Jaeger.Disabled},
								{Name: transformer.DefaultFeastSource, Value: standardTransformerConfig.DefaultFeastSource.String()},
								{
									Name:  transformer.FeastStorageConfigs,
									Value: `{"1":{"redisCluster":{"feastServingUrl":"localhost:6866","redisAddress":["10.1.1.2","10.1.1.3"],"option":{"poolSize":5,"minIdleConnections":2}}},"2":{"bigtable":{"feastServingUrl":"localhost:6867","project":"gcp-project","instance":"instance","appProfile":"default","option":{"grpcConnectionPool":4,"keepAliveInterval":"120s","keepAliveTimeout":"60s","credentialJson":"eyJrZXkiOiJ2YWx1ZSJ9"}}}}`,
								},
								{Name: envTransformerPort, Value: defaultTransformerPort},
								{Name: envTransformerModelName, Value: "test-1"},
								{Name: envTransformerPredictURL, Value: "test-1-predictor-default.test"},
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
					Name:      "test-1",
					Namespace: "test",
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
							Env: []corev1.EnvVar{
								{Name: envTransformerPort, Value: defaultTransformerPort},
								{Name: envTransformerModelName, Value: "test-1"},
								{Name: envTransformerPredictURL, Value: "test-1-predictor-default.test"},
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
			tpl := NewKFServingResourceTemplater(standardTransformerConfig)
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
