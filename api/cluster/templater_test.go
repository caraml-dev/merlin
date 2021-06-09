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

// +build unit

package cluster

import (
	"fmt"
	"testing"

	"github.com/kubeflow/kfserving/pkg/apis/serving/v1alpha2"
	kfsv1alpha2 "github.com/kubeflow/kfserving/pkg/apis/serving/v1alpha2"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gojek/merlin/config"
	"github.com/gojek/merlin/mlp"
	"github.com/gojek/merlin/models"
	"github.com/gojek/merlin/pkg/transformer"
)

var (
	defaultModelResourceRequests = &config.ResourceRequests{
		MinReplica:    1,
		MaxReplica:    2,
		CPURequest:    resource.MustParse("500m"),
		MemoryRequest: resource.MustParse("100Mi"),
	}

	expDefaultModelResourceRequests = v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    defaultModelResourceRequests.CPURequest,
			v1.ResourceMemory: defaultModelResourceRequests.MemoryRequest,
		},
		Limits: v1.ResourceList{
			v1.ResourceCPU:    getLimit(defaultModelResourceRequests.CPURequest),
			v1.ResourceMemory: getLimit(defaultModelResourceRequests.MemoryRequest),
		},
	}

	defaultTransformerResourceRequests = &config.ResourceRequests{
		MinReplica:    1,
		MaxReplica:    2,
		CPURequest:    resource.MustParse("100m"),
		MemoryRequest: resource.MustParse("500Mi"),
	}

	expDefaultTransformerResourceRequests = v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    defaultTransformerResourceRequests.CPURequest,
			v1.ResourceMemory: defaultTransformerResourceRequests.MemoryRequest,
		},
		Limits: v1.ResourceList{
			v1.ResourceCPU:    getLimit(defaultTransformerResourceRequests.CPURequest),
			v1.ResourceMemory: getLimit(defaultTransformerResourceRequests.MemoryRequest),
		},
	}

	userResourceRequests = &models.ResourceRequest{
		MinReplica:    1,
		MaxReplica:    10,
		CPURequest:    resource.MustParse("1"),
		MemoryRequest: resource.MustParse("1Gi"),
	}

	expUserResourceRequests = v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    userResourceRequests.CPURequest,
			v1.ResourceMemory: userResourceRequests.MemoryRequest,
		},
		Limits: v1.ResourceList{
			v1.ResourceCPU:    getLimit(userResourceRequests.CPURequest),
			v1.ResourceMemory: getLimit(userResourceRequests.MemoryRequest),
		},
	}

	standardTransformerConfig = config.StandardTransformerConfig{
		ImageName:       "merlin-standard-transformer",
		FeastServingURL: "serving.feast.dev:8081",
		FeastCoreURL:    "core.feast.dev:8081",
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

	tests := []struct {
		name     string
		modelSvc *models.Service
		exp      *v1alpha2.InferenceService
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
			exp: &v1alpha2.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						"queue.sidecar.serving.knative.dev/resourcePercentage": queueResourcePercentage,
					},
					Labels: map[string]string{
						"gojek.com/app":                model.Metadata.App,
						"gojek.com/orchestrator":       "merlin",
						"gojek.com/stream":             model.Metadata.Stream,
						"gojek.com/team":               model.Metadata.Team,
						"gojek.com/user-labels/sample": "true",
						"gojek.com/environment":        model.Metadata.Environment,
					},
				},
				Spec: v1alpha2.InferenceServiceSpec{
					Default: v1alpha2.EndpointSpec{
						Predictor: v1alpha2.PredictorSpec{
							Tensorflow: &v1alpha2.TensorflowSpec{
								StorageURI: fmt.Sprintf("%s/model", model.ArtifactURI),
								Resources:  expDefaultModelResourceRequests,
							},
							DeploymentSpec: v1alpha2.DeploymentSpec{
								MinReplicas: &defaultModelResourceRequests.MinReplica,
								MaxReplicas: defaultModelResourceRequests.MaxReplica,
							},
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
			exp: &v1alpha2.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						"queue.sidecar.serving.knative.dev/resourcePercentage": queueResourcePercentage,
					},
					Labels: map[string]string{
						"gojek.com/app":                model.Metadata.App,
						"gojek.com/orchestrator":       "merlin",
						"gojek.com/stream":             model.Metadata.Stream,
						"gojek.com/team":               model.Metadata.Team,
						"gojek.com/user-labels/sample": "true",
						"gojek.com/environment":        model.Metadata.Environment,
					},
				},
				Spec: v1alpha2.InferenceServiceSpec{
					Default: v1alpha2.EndpointSpec{
						Predictor: v1alpha2.PredictorSpec{
							XGBoost: &v1alpha2.XGBoostSpec{
								StorageURI: fmt.Sprintf("%s/model", model.ArtifactURI),
								Resources:  expDefaultModelResourceRequests,
							},
							DeploymentSpec: v1alpha2.DeploymentSpec{
								MinReplicas: &defaultModelResourceRequests.MinReplica,
								MaxReplicas: defaultModelResourceRequests.MaxReplica,
							},
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
			exp: &v1alpha2.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						"queue.sidecar.serving.knative.dev/resourcePercentage": queueResourcePercentage,
					},
					Labels: map[string]string{
						"gojek.com/app":                model.Metadata.App,
						"gojek.com/orchestrator":       "merlin",
						"gojek.com/stream":             model.Metadata.Stream,
						"gojek.com/team":               model.Metadata.Team,
						"gojek.com/user-labels/sample": "true",
						"gojek.com/environment":        model.Metadata.Environment,
					},
				},
				Spec: v1alpha2.InferenceServiceSpec{
					Default: v1alpha2.EndpointSpec{
						Predictor: v1alpha2.PredictorSpec{
							SKLearn: &v1alpha2.SKLearnSpec{
								StorageURI: fmt.Sprintf("%s/model", model.ArtifactURI),
								Resources:  expDefaultModelResourceRequests,
							},
							DeploymentSpec: v1alpha2.DeploymentSpec{
								MinReplicas: &defaultModelResourceRequests.MinReplica,
								MaxReplicas: defaultModelResourceRequests.MaxReplica,
							},
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
				Options: &models.ModelOption{
					PyTorchModelClassName: "MyModel",
				},
				Metadata: model.Metadata,
			},
			exp: &v1alpha2.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						"queue.sidecar.serving.knative.dev/resourcePercentage": queueResourcePercentage,
					},
					Labels: map[string]string{
						"gojek.com/app":                model.Metadata.App,
						"gojek.com/orchestrator":       "merlin",
						"gojek.com/stream":             model.Metadata.Stream,
						"gojek.com/team":               model.Metadata.Team,
						"gojek.com/user-labels/sample": "true",
						"gojek.com/environment":        model.Metadata.Environment,
					},
				},
				Spec: v1alpha2.InferenceServiceSpec{
					Default: v1alpha2.EndpointSpec{
						Predictor: v1alpha2.PredictorSpec{
							PyTorch: &v1alpha2.PyTorchSpec{
								StorageURI:     fmt.Sprintf("%s/model", model.ArtifactURI),
								ModelClassName: "MyModel",
								Resources:      expDefaultModelResourceRequests,
							},
							DeploymentSpec: v1alpha2.DeploymentSpec{
								MinReplicas: &defaultModelResourceRequests.MinReplica,
								MaxReplicas: defaultModelResourceRequests.MaxReplica,
							},
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
			exp: &v1alpha2.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						"queue.sidecar.serving.knative.dev/resourcePercentage": queueResourcePercentage,
					},
					Labels: map[string]string{
						"gojek.com/app":                model.Metadata.App,
						"gojek.com/orchestrator":       "merlin",
						"gojek.com/stream":             model.Metadata.Stream,
						"gojek.com/team":               model.Metadata.Team,
						"gojek.com/user-labels/sample": "true",
						"gojek.com/environment":        model.Metadata.Environment,
					},
				},
				Spec: v1alpha2.InferenceServiceSpec{
					Default: v1alpha2.EndpointSpec{
						Predictor: v1alpha2.PredictorSpec{
							ONNX: &v1alpha2.ONNXSpec{
								StorageURI: fmt.Sprintf("%s/model", model.ArtifactURI),
								Resources:  expDefaultModelResourceRequests,
							},
							DeploymentSpec: v1alpha2.DeploymentSpec{
								MinReplicas: &defaultModelResourceRequests.MinReplica,
								MaxReplicas: defaultModelResourceRequests.MaxReplica,
							},
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
			exp: &v1alpha2.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						"queue.sidecar.serving.knative.dev/resourcePercentage": queueResourcePercentage,
						"prometheus.io/scrape":                                 "true",
						"prometheus.io/port":                                   "8080",
					},
					Labels: map[string]string{
						"gojek.com/app":                model.Metadata.App,
						"gojek.com/orchestrator":       "merlin",
						"gojek.com/stream":             model.Metadata.Stream,
						"gojek.com/team":               model.Metadata.Team,
						"gojek.com/user-labels/sample": "true",
						"gojek.com/environment":        model.Metadata.Environment,
					},
				},
				Spec: v1alpha2.InferenceServiceSpec{
					Default: v1alpha2.EndpointSpec{
						Predictor: v1alpha2.PredictorSpec{
							Custom: &v1alpha2.CustomSpec{
								Container: v1.Container{
									Image:     "gojek/project-model:1",
									Env:       models.PyfuncDefaultEnvVars(models.Model{Name: model.Name}, models.Version{ID: models.ID(1), ArtifactURI: model.ArtifactURI}, defaultModelResourceRequests.CPURequest.Value()).ToKubernetesEnvVars(),
									Resources: expDefaultModelResourceRequests,
								},
							},
							DeploymentSpec: v1alpha2.DeploymentSpec{
								MinReplicas: &defaultModelResourceRequests.MinReplica,
								MaxReplicas: defaultModelResourceRequests.MaxReplica,
							},
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
			exp: &v1alpha2.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						"queue.sidecar.serving.knative.dev/resourcePercentage": queueResourcePercentage,
					},
					Labels: map[string]string{
						"gojek.com/app":                model.Metadata.App,
						"gojek.com/orchestrator":       "merlin",
						"gojek.com/stream":             model.Metadata.Stream,
						"gojek.com/team":               model.Metadata.Team,
						"gojek.com/user-labels/sample": "true",
						"gojek.com/environment":        model.Metadata.Environment,
					},
				},
				Spec: v1alpha2.InferenceServiceSpec{
					Default: v1alpha2.EndpointSpec{
						Predictor: v1alpha2.PredictorSpec{
							Tensorflow: &v1alpha2.TensorflowSpec{
								StorageURI: fmt.Sprintf("%s/model", model.ArtifactURI),
								Resources:  expUserResourceRequests,
							},
							DeploymentSpec: v1alpha2.DeploymentSpec{
								MinReplicas: &userResourceRequests.MinReplica,
								MaxReplicas: userResourceRequests.MaxReplica,
							},
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
			exp: &v1alpha2.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						"queue.sidecar.serving.knative.dev/resourcePercentage": queueResourcePercentage,
					},
					Labels: map[string]string{
						"gojek.com/app":                model.Metadata.App,
						"gojek.com/orchestrator":       "merlin",
						"gojek.com/stream":             model.Metadata.Stream,
						"gojek.com/team":               model.Metadata.Team,
						"gojek.com/user-labels/sample": "true",
						"gojek.com/environment":        model.Metadata.Environment,
					},
				},
				Spec: v1alpha2.InferenceServiceSpec{
					Default: v1alpha2.EndpointSpec{
						Predictor: v1alpha2.PredictorSpec{
							Custom: &kfsv1alpha2.CustomSpec{
								Container: v1.Container{
									Image: "gcr.io/custom-model:v0.1",
									Env: models.EnvVars{
										models.EnvVar{Name: "MERLIN_PREDICTOR_PORT", Value: "8080"},
										models.EnvVar{Name: "MERLIN_MODEL_NAME", Value: models.CreateInferenceServiceName(model.Name, "1")},
										models.EnvVar{Name: "MERLIN_ARTIFACT_LOCATION", Value: "/mnt/models"},
										models.EnvVar{Name: "STORAGE_URI", Value: model.ArtifactURI},
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
							DeploymentSpec: v1alpha2.DeploymentSpec{
								MinReplicas: &userResourceRequests.MinReplica,
								MaxReplicas: userResourceRequests.MaxReplica,
							},
						},
					},
				},
			},
		},
		{
			name: "custom spec without resource request",
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
			exp: &v1alpha2.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						"queue.sidecar.serving.knative.dev/resourcePercentage": queueResourcePercentage,
					},
					Labels: map[string]string{
						"gojek.com/app":                model.Metadata.App,
						"gojek.com/orchestrator":       "merlin",
						"gojek.com/stream":             model.Metadata.Stream,
						"gojek.com/team":               model.Metadata.Team,
						"gojek.com/user-labels/sample": "true",
						"gojek.com/environment":        model.Metadata.Environment,
					},
				},
				Spec: v1alpha2.InferenceServiceSpec{
					Default: v1alpha2.EndpointSpec{
						Predictor: v1alpha2.PredictorSpec{
							Custom: &kfsv1alpha2.CustomSpec{
								Container: v1.Container{
									Image: "gcr.io/custom-model:v0.1",
									Env: models.EnvVars{
										models.EnvVar{Name: "MERLIN_PREDICTOR_PORT", Value: "8080"},
										models.EnvVar{Name: "MERLIN_MODEL_NAME", Value: models.CreateInferenceServiceName(model.Name, "1")},
										models.EnvVar{Name: "MERLIN_ARTIFACT_LOCATION", Value: "/mnt/models"},
										models.EnvVar{Name: "STORAGE_URI", Value: model.ArtifactURI},
									}.ToKubernetesEnvVars(),
									Resources: expDefaultModelResourceRequests,
									Command:   nil,
									Args:      nil,
								},
							},
							DeploymentSpec: v1alpha2.DeploymentSpec{
								MinReplicas: &defaultModelResourceRequests.MinReplica,
								MaxReplicas: defaultModelResourceRequests.MaxReplica,
							},
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
			infSvcSpec := tpl.CreateInferenceServiceSpec(tt.modelSvc, deployConfig)
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

	tests := []struct {
		name     string
		modelSvc *models.Service
		exp      *v1alpha2.InferenceService
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
			exp: &v1alpha2.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						"queue.sidecar.serving.knative.dev/resourcePercentage": queueResourcePercentage,
					},
					Labels: map[string]string{
						"gojek.com/app":                model.Metadata.App,
						"gojek.com/orchestrator":       "merlin",
						"gojek.com/stream":             model.Metadata.Stream,
						"gojek.com/team":               model.Metadata.Team,
						"gojek.com/user-labels/sample": "true",
						"gojek.com/environment":        model.Metadata.Environment,
					},
				},
				Spec: v1alpha2.InferenceServiceSpec{
					Default: v1alpha2.EndpointSpec{
						Predictor: v1alpha2.PredictorSpec{
							Tensorflow: &v1alpha2.TensorflowSpec{
								StorageURI: fmt.Sprintf("%s/model", model.ArtifactURI),
								Resources:  expDefaultModelResourceRequests,
							},
							DeploymentSpec: v1alpha2.DeploymentSpec{
								MinReplicas: &defaultModelResourceRequests.MinReplica,
								MaxReplicas: defaultModelResourceRequests.MaxReplica,
							},
						},
						Transformer: &kfsv1alpha2.TransformerSpec{
							Custom: &kfsv1alpha2.CustomSpec{
								Container: v1.Container{
									Name:    "transformer",
									Image:   "ghcr.io/gojek/merlin-transformer-test",
									Command: []string{"python"},
									Args:    []string{"main.py"},
									Env: []v1.EnvVar{
										{Name: envTransformerPort, Value: defaultTransformerPort},
										{Name: envTransformerModelName, Value: "model-1"},
										{Name: envTransformerPredictURL, Value: "model-1-predictor-default.project"},
									},
									Resources: expDefaultTransformerResourceRequests,
								},
							},
							DeploymentSpec: kfsv1alpha2.DeploymentSpec{
								MinReplicas: &defaultTransformerResourceRequests.MinReplica,
								MaxReplicas: defaultTransformerResourceRequests.MaxReplica,
							},
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
			exp: &v1alpha2.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						"queue.sidecar.serving.knative.dev/resourcePercentage": queueResourcePercentage,
					},
					Labels: map[string]string{
						"gojek.com/app":                model.Metadata.App,
						"gojek.com/orchestrator":       "merlin",
						"gojek.com/stream":             model.Metadata.Stream,
						"gojek.com/team":               model.Metadata.Team,
						"gojek.com/user-labels/sample": "true",
						"gojek.com/environment":        model.Metadata.Environment,
					},
				},
				Spec: v1alpha2.InferenceServiceSpec{
					Default: v1alpha2.EndpointSpec{
						Predictor: v1alpha2.PredictorSpec{
							Tensorflow: &v1alpha2.TensorflowSpec{
								StorageURI: fmt.Sprintf("%s/model", model.ArtifactURI),
								Resources:  expDefaultModelResourceRequests,
							},
							DeploymentSpec: v1alpha2.DeploymentSpec{
								MinReplicas: &defaultModelResourceRequests.MinReplica,
								MaxReplicas: defaultModelResourceRequests.MaxReplica,
							},
						},
						Transformer: &kfsv1alpha2.TransformerSpec{
							Custom: &kfsv1alpha2.CustomSpec{
								Container: v1.Container{
									Name:    "transformer",
									Image:   "ghcr.io/gojek/merlin-transformer-test",
									Command: []string{"python"},
									Args:    []string{"main.py"},
									Env: []v1.EnvVar{
										{Name: envTransformerPort, Value: defaultTransformerPort},
										{Name: envTransformerModelName, Value: "model-1"},
										{Name: envTransformerPredictURL, Value: "model-1-predictor-default.project"},
									},
									Resources: expUserResourceRequests,
								},
							},
							DeploymentSpec: kfsv1alpha2.DeploymentSpec{
								MinReplicas: &userResourceRequests.MinReplica,
								MaxReplicas: userResourceRequests.MaxReplica,
							},
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
							Value: `  { "standard_transformer": null}`,
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
			exp: &v1alpha2.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						"queue.sidecar.serving.knative.dev/resourcePercentage": queueResourcePercentage,
					},
					Labels: map[string]string{
						"gojek.com/app":                model.Metadata.App,
						"gojek.com/orchestrator":       "merlin",
						"gojek.com/stream":             model.Metadata.Stream,
						"gojek.com/team":               model.Metadata.Team,
						"gojek.com/user-labels/sample": "true",
						"gojek.com/environment":        model.Metadata.Environment,
					},
				},
				Spec: v1alpha2.InferenceServiceSpec{
					Default: v1alpha2.EndpointSpec{
						Predictor: v1alpha2.PredictorSpec{
							Tensorflow: &v1alpha2.TensorflowSpec{
								StorageURI: fmt.Sprintf("%s/model", model.ArtifactURI),
								Resources:  expDefaultModelResourceRequests,
							},
							DeploymentSpec: v1alpha2.DeploymentSpec{
								MinReplicas: &defaultModelResourceRequests.MinReplica,
								MaxReplicas: defaultModelResourceRequests.MaxReplica,
							},
						},
						Transformer: &kfsv1alpha2.TransformerSpec{
							Custom: &kfsv1alpha2.CustomSpec{
								Container: v1.Container{
									Name:  "transformer",
									Image: standardTransformerConfig.ImageName,
									Env: []v1.EnvVar{
										{Name: transformer.StandardTransformerConfigEnvName, Value: `{"standard_transformer":null}`},
										{Name: transformer.FeastServingURLEnvName, Value: standardTransformerConfig.FeastServingURL},
										{Name: envTransformerPort, Value: defaultTransformerPort},
										{Name: envTransformerModelName, Value: "model-1"},
										{Name: envTransformerPredictURL, Value: "model-1-predictor-default.project"},
									},
									Resources: expDefaultTransformerResourceRequests,
								},
							},
							DeploymentSpec: kfsv1alpha2.DeploymentSpec{
								MinReplicas: &defaultTransformerResourceRequests.MinReplica,
								MaxReplicas: defaultTransformerResourceRequests.MaxReplica,
							},
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
			infSvcSpec := tpl.CreateInferenceServiceSpec(tt.modelSvc, deployConfig)
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

	tests := []struct {
		name     string
		modelSvc *models.Service
		exp      *v1alpha2.InferenceService
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
			exp: &v1alpha2.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						"queue.sidecar.serving.knative.dev/resourcePercentage": queueResourcePercentage,
					},
					Labels: map[string]string{
						"gojek.com/app":                model.Metadata.App,
						"gojek.com/orchestrator":       "merlin",
						"gojek.com/stream":             model.Metadata.Stream,
						"gojek.com/team":               model.Metadata.Team,
						"gojek.com/user-labels/sample": "true",
						"gojek.com/environment":        model.Metadata.Environment,
					},
				},
				Spec: v1alpha2.InferenceServiceSpec{
					Default: v1alpha2.EndpointSpec{
						Predictor: v1alpha2.PredictorSpec{
							Tensorflow: &v1alpha2.TensorflowSpec{
								StorageURI: fmt.Sprintf("%s/model", model.ArtifactURI),
								Resources:  expDefaultModelResourceRequests,
							},
							DeploymentSpec: v1alpha2.DeploymentSpec{
								MinReplicas: &defaultModelResourceRequests.MinReplica,
								MaxReplicas: defaultModelResourceRequests.MaxReplica,
								Logger: &kfsv1alpha2.Logger{
									Url:  &loggerDestinationURL,
									Mode: kfsv1alpha2.LogAll,
								},
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
			exp: &v1alpha2.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						"queue.sidecar.serving.knative.dev/resourcePercentage": queueResourcePercentage,
					},
					Labels: map[string]string{
						"gojek.com/app":                model.Metadata.App,
						"gojek.com/orchestrator":       "merlin",
						"gojek.com/stream":             model.Metadata.Stream,
						"gojek.com/team":               model.Metadata.Team,
						"gojek.com/user-labels/sample": "true",
						"gojek.com/environment":        model.Metadata.Environment,
					},
				},
				Spec: v1alpha2.InferenceServiceSpec{
					Default: v1alpha2.EndpointSpec{
						Predictor: v1alpha2.PredictorSpec{
							Tensorflow: &v1alpha2.TensorflowSpec{
								StorageURI: fmt.Sprintf("%s/model", model.ArtifactURI),
								Resources:  expDefaultModelResourceRequests,
							},
							DeploymentSpec: v1alpha2.DeploymentSpec{
								MinReplicas: &defaultModelResourceRequests.MinReplica,
								MaxReplicas: defaultModelResourceRequests.MaxReplica,
								Logger: &kfsv1alpha2.Logger{
									Url:  &loggerDestinationURL,
									Mode: kfsv1alpha2.LogAll,
								},
							},
						},
						Transformer: &kfsv1alpha2.TransformerSpec{
							Custom: &kfsv1alpha2.CustomSpec{
								Container: v1.Container{
									Name:    "transformer",
									Image:   "ghcr.io/gojek/merlin-transformer-test",
									Command: []string{"python"},
									Args:    []string{"main.py"},
									Env: []v1.EnvVar{
										{Name: envTransformerPort, Value: defaultTransformerPort},
										{Name: envTransformerModelName, Value: "model-1"},
										{Name: envTransformerPredictURL, Value: "model-1-predictor-default.project"},
									},
									Resources: expDefaultTransformerResourceRequests,
								},
							},
							DeploymentSpec: kfsv1alpha2.DeploymentSpec{
								MinReplicas: &defaultTransformerResourceRequests.MinReplica,
								MaxReplicas: defaultTransformerResourceRequests.MaxReplica,
							},
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
			exp: &v1alpha2.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						"queue.sidecar.serving.knative.dev/resourcePercentage": queueResourcePercentage,
					},
					Labels: map[string]string{
						"gojek.com/app":                model.Metadata.App,
						"gojek.com/orchestrator":       "merlin",
						"gojek.com/stream":             model.Metadata.Stream,
						"gojek.com/team":               model.Metadata.Team,
						"gojek.com/user-labels/sample": "true",
						"gojek.com/environment":        model.Metadata.Environment,
					},
				},
				Spec: v1alpha2.InferenceServiceSpec{
					Default: v1alpha2.EndpointSpec{
						Predictor: v1alpha2.PredictorSpec{
							Tensorflow: &v1alpha2.TensorflowSpec{
								StorageURI: fmt.Sprintf("%s/model", model.ArtifactURI),
								Resources:  expDefaultModelResourceRequests,
							},
							DeploymentSpec: v1alpha2.DeploymentSpec{
								MinReplicas: &defaultModelResourceRequests.MinReplica,
								MaxReplicas: defaultModelResourceRequests.MaxReplica,
							},
						},
						Transformer: &kfsv1alpha2.TransformerSpec{
							Custom: &kfsv1alpha2.CustomSpec{
								Container: v1.Container{
									Name:    "transformer",
									Image:   "ghcr.io/gojek/merlin-transformer-test",
									Command: []string{"python"},
									Args:    []string{"main.py"},
									Env: []v1.EnvVar{
										{Name: envTransformerPort, Value: defaultTransformerPort},
										{Name: envTransformerModelName, Value: "model-1"},
										{Name: envTransformerPredictURL, Value: "model-1-predictor-default.project"},
									},
									Resources: expDefaultTransformerResourceRequests,
								},
							},
							DeploymentSpec: kfsv1alpha2.DeploymentSpec{
								MinReplicas: &defaultTransformerResourceRequests.MinReplica,
								MaxReplicas: defaultTransformerResourceRequests.MaxReplica,
							},
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
			exp: &v1alpha2.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						"queue.sidecar.serving.knative.dev/resourcePercentage": queueResourcePercentage,
					},
					Labels: map[string]string{
						"gojek.com/app":                model.Metadata.App,
						"gojek.com/orchestrator":       "merlin",
						"gojek.com/stream":             model.Metadata.Stream,
						"gojek.com/team":               model.Metadata.Team,
						"gojek.com/user-labels/sample": "true",
						"gojek.com/environment":        model.Metadata.Environment,
					},
				},
				Spec: v1alpha2.InferenceServiceSpec{
					Default: v1alpha2.EndpointSpec{
						Predictor: v1alpha2.PredictorSpec{
							Tensorflow: &v1alpha2.TensorflowSpec{
								StorageURI: fmt.Sprintf("%s/model", model.ArtifactURI),
								Resources:  expDefaultModelResourceRequests,
							},
							DeploymentSpec: v1alpha2.DeploymentSpec{
								MinReplicas: &defaultModelResourceRequests.MinReplica,
								MaxReplicas: defaultModelResourceRequests.MaxReplica,
							},
						},
						Transformer: &kfsv1alpha2.TransformerSpec{
							Custom: &kfsv1alpha2.CustomSpec{
								Container: v1.Container{
									Name:    "transformer",
									Image:   "ghcr.io/gojek/merlin-transformer-test",
									Command: []string{"python"},
									Args:    []string{"main.py"},
									Env: []v1.EnvVar{
										{Name: envTransformerPort, Value: defaultTransformerPort},
										{Name: envTransformerModelName, Value: "model-1"},
										{Name: envTransformerPredictURL, Value: "model-1-predictor-default.project"},
									},
									Resources: expDefaultTransformerResourceRequests,
								},
							},
							DeploymentSpec: kfsv1alpha2.DeploymentSpec{
								MinReplicas: &defaultTransformerResourceRequests.MinReplica,
								MaxReplicas: defaultTransformerResourceRequests.MaxReplica,
								Logger: &kfsv1alpha2.Logger{
									Url:  &loggerDestinationURL,
									Mode: kfsv1alpha2.LogRequest,
								},
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
			exp: &v1alpha2.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						"queue.sidecar.serving.knative.dev/resourcePercentage": queueResourcePercentage,
					},
					Labels: map[string]string{
						"gojek.com/app":                model.Metadata.App,
						"gojek.com/orchestrator":       "merlin",
						"gojek.com/stream":             model.Metadata.Stream,
						"gojek.com/team":               model.Metadata.Team,
						"gojek.com/user-labels/sample": "true",
						"gojek.com/environment":        model.Metadata.Environment,
					},
				},
				Spec: v1alpha2.InferenceServiceSpec{
					Default: v1alpha2.EndpointSpec{
						Predictor: v1alpha2.PredictorSpec{
							Tensorflow: &v1alpha2.TensorflowSpec{
								StorageURI: fmt.Sprintf("%s/model", model.ArtifactURI),
								Resources:  expDefaultModelResourceRequests,
							},
							DeploymentSpec: v1alpha2.DeploymentSpec{
								MinReplicas: &defaultModelResourceRequests.MinReplica,
								MaxReplicas: defaultModelResourceRequests.MaxReplica,
							},
						},
						Transformer: &kfsv1alpha2.TransformerSpec{
							Custom: &kfsv1alpha2.CustomSpec{
								Container: v1.Container{
									Name:    "transformer",
									Image:   "ghcr.io/gojek/merlin-transformer-test",
									Command: []string{"python"},
									Args:    []string{"main.py"},
									Env: []v1.EnvVar{
										{Name: envTransformerPort, Value: defaultTransformerPort},
										{Name: envTransformerModelName, Value: "model-1"},
										{Name: envTransformerPredictURL, Value: "model-1-predictor-default.project"},
									},
									Resources: expDefaultTransformerResourceRequests,
								},
							},
							DeploymentSpec: kfsv1alpha2.DeploymentSpec{
								MinReplicas: &defaultTransformerResourceRequests.MinReplica,
								MaxReplicas: defaultTransformerResourceRequests.MaxReplica,
							},
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
			infSvcSpec := tpl.CreateInferenceServiceSpec(tt.modelSvc, deployConfig)
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

	resourceRequests := v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    cpuRequest,
			v1.ResourceMemory: memoryRequest,
		},
		Limits: v1.ResourceList{
			v1.ResourceCPU:    cpuLimit,
			v1.ResourceMemory: memoryLimit,
		},
	}

	tests := []struct {
		name     string
		modelSvc *models.Service
		original *v1alpha2.InferenceService
		exp      *v1alpha2.InferenceService
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
			original: &v1alpha2.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						"queue.sidecar.serving.knative.dev/resourcePercentage": queueResourcePercentage,
					},
				},
				Spec: v1alpha2.InferenceServiceSpec{
					Default: v1alpha2.EndpointSpec{
						Predictor: v1alpha2.PredictorSpec{
							Tensorflow: &v1alpha2.TensorflowSpec{
								StorageURI: fmt.Sprintf("%s/model", model.ArtifactURI),
								Resources:  resourceRequests,
							},
							DeploymentSpec: v1alpha2.DeploymentSpec{
								MinReplicas: &minReplica,
								MaxReplicas: maxReplica,
							},
						},
					},
				},
			},
			exp: &v1alpha2.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						"queue.sidecar.serving.knative.dev/resourcePercentage": queueResourcePercentage,
					},
					Labels: map[string]string{
						"gojek.com/app":                model.Metadata.App,
						"gojek.com/orchestrator":       "merlin",
						"gojek.com/stream":             model.Metadata.Stream,
						"gojek.com/team":               model.Metadata.Team,
						"gojek.com/user-labels/sample": "true",
						"gojek.com/environment":        model.Metadata.Environment,
					},
				},
				Spec: v1alpha2.InferenceServiceSpec{
					Default: v1alpha2.EndpointSpec{
						Predictor: v1alpha2.PredictorSpec{
							Tensorflow: &v1alpha2.TensorflowSpec{
								StorageURI: fmt.Sprintf("%s/model", model.ArtifactURI),
								Resources:  resourceRequests,
							},
							DeploymentSpec: v1alpha2.DeploymentSpec{
								MinReplicas: &minReplica,
								MaxReplicas: maxReplica,
							},
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
			original: &v1alpha2.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						"queue.sidecar.serving.knative.dev/resourcePercentage": queueResourcePercentage,
					},
				},
				Spec: v1alpha2.InferenceServiceSpec{
					Default: v1alpha2.EndpointSpec{
						Predictor: v1alpha2.PredictorSpec{
							Tensorflow: &v1alpha2.TensorflowSpec{
								StorageURI: fmt.Sprintf("%s/model", model.ArtifactURI),
								Resources:  resourceRequests,
							},
							DeploymentSpec: v1alpha2.DeploymentSpec{
								MinReplicas: &minReplica,
								MaxReplicas: maxReplica,
							},
						},
					},
				},
			},
			exp: &v1alpha2.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						"queue.sidecar.serving.knative.dev/resourcePercentage": queueResourcePercentage,
					},
					Labels: map[string]string{
						"gojek.com/app":                model.Metadata.App,
						"gojek.com/orchestrator":       "merlin",
						"gojek.com/stream":             model.Metadata.Stream,
						"gojek.com/team":               model.Metadata.Team,
						"gojek.com/user-labels/sample": "true",
						"gojek.com/environment":        model.Metadata.Environment,
					},
				},
				Spec: v1alpha2.InferenceServiceSpec{
					Default: v1alpha2.EndpointSpec{
						Predictor: v1alpha2.PredictorSpec{
							Tensorflow: &v1alpha2.TensorflowSpec{
								StorageURI: fmt.Sprintf("%s/model", model.ArtifactURI),
								Resources:  resourceRequests,
							},
							DeploymentSpec: v1alpha2.DeploymentSpec{
								MinReplicas: &minReplica,
								MaxReplicas: maxReplica,
							},
						},
						Transformer: &kfsv1alpha2.TransformerSpec{
							Custom: &kfsv1alpha2.CustomSpec{
								Container: v1.Container{
									Name:    "transformer",
									Image:   "ghcr.io/gojek/merlin-transformer-test",
									Command: []string{"python"},
									Args:    []string{"main.py"},
									Env: []v1.EnvVar{
										{Name: envTransformerPort, Value: defaultTransformerPort},
										{Name: envTransformerModelName, Value: "model-1"},
										{Name: envTransformerPredictURL, Value: "model-1-predictor-default.project"},
									},
									Resources: v1.ResourceRequirements{
										Requests: v1.ResourceList{
											v1.ResourceCPU:    cpuRequest,
											v1.ResourceMemory: memoryRequest,
										},
										Limits: v1.ResourceList{
											v1.ResourceCPU:    cpuLimit,
											v1.ResourceMemory: memoryLimit,
										},
									},
								},
							},
							DeploymentSpec: kfsv1alpha2.DeploymentSpec{
								MinReplicas: &one,
								MaxReplicas: one,
							},
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
			original: &v1alpha2.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						"queue.sidecar.serving.knative.dev/resourcePercentage": queueResourcePercentage,
					},
				},
				Spec: v1alpha2.InferenceServiceSpec{
					Default: v1alpha2.EndpointSpec{
						Predictor: v1alpha2.PredictorSpec{
							Tensorflow: &v1alpha2.TensorflowSpec{
								StorageURI: fmt.Sprintf("%s/model", model.ArtifactURI),
								Resources:  resourceRequests,
							},
							DeploymentSpec: v1alpha2.DeploymentSpec{
								MinReplicas: &minReplica,
								MaxReplicas: maxReplica,
							},
						},
						Transformer: &kfsv1alpha2.TransformerSpec{
							Custom: &kfsv1alpha2.CustomSpec{
								Container: v1.Container{
									Name:    "transformer",
									Image:   "ghcr.io/gojek/merlin-transformer-test",
									Command: []string{"python"},
									Args:    []string{"main.py"},
									Env: []v1.EnvVar{
										{Name: envTransformerPort, Value: defaultTransformerPort},
										{Name: envTransformerModelName, Value: "model-1"},
										{Name: envTransformerPredictURL, Value: "model-1-predictor-default.project"},
									},
									Resources: v1.ResourceRequirements{
										Requests: v1.ResourceList{
											v1.ResourceCPU:    cpuRequest,
											v1.ResourceMemory: memoryRequest,
										},
										Limits: v1.ResourceList{
											v1.ResourceCPU:    cpuLimit,
											v1.ResourceMemory: memoryLimit,
										},
									},
								},
							},
							DeploymentSpec: kfsv1alpha2.DeploymentSpec{
								MinReplicas: &one,
								MaxReplicas: one,
							},
						},
					},
				},
			},
			exp: &v1alpha2.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						"queue.sidecar.serving.knative.dev/resourcePercentage": queueResourcePercentage,
					},
					Labels: map[string]string{
						"gojek.com/app":                model.Metadata.App,
						"gojek.com/orchestrator":       "merlin",
						"gojek.com/stream":             model.Metadata.Stream,
						"gojek.com/team":               model.Metadata.Team,
						"gojek.com/user-labels/sample": "true",
						"gojek.com/environment":        model.Metadata.Environment,
					},
				},
				Spec: v1alpha2.InferenceServiceSpec{
					Default: v1alpha2.EndpointSpec{
						Predictor: v1alpha2.PredictorSpec{
							Tensorflow: &v1alpha2.TensorflowSpec{
								StorageURI: fmt.Sprintf("%s/model", model.ArtifactURI),
								Resources:  resourceRequests,
							},
							DeploymentSpec: v1alpha2.DeploymentSpec{
								MinReplicas: &minReplica,
								MaxReplicas: maxReplica,
							},
						},
						Transformer: nil,
					},
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
			original: &v1alpha2.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						"queue.sidecar.serving.knative.dev/resourcePercentage": queueResourcePercentage,
					},
					Labels: map[string]string{
						"gojek.com/app":                model.Metadata.App,
						"gojek.com/orchestrator":       "merlin",
						"gojek.com/stream":             model.Metadata.Stream,
						"gojek.com/team":               model.Metadata.Team,
						"gojek.com/user-labels/sample": "true",
						"gojek.com/environment":        model.Metadata.Environment,
					},
				},
				Spec: v1alpha2.InferenceServiceSpec{
					Default: v1alpha2.EndpointSpec{
						Predictor: v1alpha2.PredictorSpec{
							Custom: &kfsv1alpha2.CustomSpec{
								Container: v1.Container{
									Image: "gcr.io/custom-model:v0.1",
									Env: models.EnvVars{
										models.EnvVar{Name: "MERLIN_PREDICTOR_PORT", Value: "8080"},
										models.EnvVar{Name: "MERLIN_MODEL_NAME", Value: models.CreateInferenceServiceName(model.Name, "1")},
										models.EnvVar{Name: "MERLIN_ARTIFACT_LOCATION", Value: "/mnt/models"},
										models.EnvVar{Name: "STORAGE_URI", Value: model.ArtifactURI},
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
							DeploymentSpec: v1alpha2.DeploymentSpec{
								MinReplicas: &userResourceRequests.MinReplica,
								MaxReplicas: userResourceRequests.MaxReplica,
							},
						},
					},
				},
			},
			exp: &v1alpha2.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionID),
					Namespace: project.Name,
					Annotations: map[string]string{
						"queue.sidecar.serving.knative.dev/resourcePercentage": queueResourcePercentage,
					},
					Labels: map[string]string{
						"gojek.com/app":                model.Metadata.App,
						"gojek.com/orchestrator":       "merlin",
						"gojek.com/stream":             model.Metadata.Stream,
						"gojek.com/team":               model.Metadata.Team,
						"gojek.com/user-labels/sample": "true",
						"gojek.com/environment":        model.Metadata.Environment,
					},
				},
				Spec: v1alpha2.InferenceServiceSpec{
					Default: v1alpha2.EndpointSpec{
						Predictor: v1alpha2.PredictorSpec{
							Custom: &kfsv1alpha2.CustomSpec{
								Container: v1.Container{
									Image: "gcr.io/custom-model:v0.2",
									Env: models.EnvVars{
										models.EnvVar{Name: "MERLIN_PREDICTOR_PORT", Value: "8080"},
										models.EnvVar{Name: "MERLIN_MODEL_NAME", Value: models.CreateInferenceServiceName(model.Name, "1")},
										models.EnvVar{Name: "MERLIN_ARTIFACT_LOCATION", Value: "/mnt/models"},
										models.EnvVar{Name: "STORAGE_URI", Value: model.ArtifactURI},
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
							DeploymentSpec: v1alpha2.DeploymentSpec{
								MinReplicas: &userResourceRequests.MinReplica,
								MaxReplicas: userResourceRequests.MaxReplica,
							},
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
			infSvcSpec := tpl.PatchInferenceServiceSpec(tt.original, tt.modelSvc, deployConfig)
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
		want *kfsv1alpha2.TransformerSpec
	}{
		{
			"complete",
			args{
				&models.Service{
					Name:      "test-1",
					Namespace: "test",
				},
				&models.Transformer{
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
				&config.DeploymentConfig{},
			},
			&kfsv1alpha2.TransformerSpec{
				Custom: &kfsv1alpha2.CustomSpec{
					Container: v1.Container{
						Name:    "transformer",
						Image:   "ghcr.io/gojek/merlin-transformer-test",
						Command: []string{"python"},
						Args:    []string{"main.py"},
						Env: []v1.EnvVar{
							{Name: envTransformerPort, Value: defaultTransformerPort},
							{Name: envTransformerModelName, Value: "test-1"},
							{Name: envTransformerPredictURL, Value: "test-1-predictor-default.test"},
						},
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceCPU:    cpuRequest,
								v1.ResourceMemory: memoryRequest,
							},
							Limits: v1.ResourceList{
								v1.ResourceCPU:    cpuLimit,
								v1.ResourceMemory: memoryLimit,
							},
						},
					},
				},
				DeploymentSpec: kfsv1alpha2.DeploymentSpec{
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
