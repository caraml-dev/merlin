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
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gojek/merlin/config"
	"github.com/gojek/merlin/mlp"
	"github.com/gojek/merlin/models"
)

func TestCreateInferenceServiceSpec(t *testing.T) {
	project := mlp.Project{
		Name: "project",
	}

	model := &models.Service{
		Name:        "model",
		ArtifactUri: "gs://my-artifacet",
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
	versionId := 1

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
		exp      *v1alpha2.InferenceService
	}{
		{
			name: "tensorflow spec",
			modelSvc: &models.Service{
				Name:        models.CreateInferenceServiceName(model.Name, "1"),
				Namespace:   project.Name,
				ArtifactUri: model.ArtifactUri,
				Type:        models.ModelTypeTensorflow,
				Options:     &models.ModelOption{},
				Metadata:    model.Metadata,
			},
			exp: &v1alpha2.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionId),
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
								StorageURI: fmt.Sprintf("%s/model", model.ArtifactUri),
								Resources:  resourceRequests,
							},
							DeploymentSpec: v1alpha2.DeploymentSpec{
								MinReplicas: minReplica,
								MaxReplicas: maxReplica,
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
				ArtifactUri: model.ArtifactUri,
				Type:        models.ModelTypeXgboost,
				Options:     &models.ModelOption{},
				Metadata:    model.Metadata,
			},
			exp: &v1alpha2.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionId),
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
								StorageURI: fmt.Sprintf("%s/model", model.ArtifactUri),
								Resources:  resourceRequests,
							},
							DeploymentSpec: v1alpha2.DeploymentSpec{
								MinReplicas: minReplica,
								MaxReplicas: maxReplica,
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
				ArtifactUri: model.ArtifactUri,
				Type:        models.ModelTypeSkLearn,
				Options:     &models.ModelOption{},
				Metadata:    model.Metadata,
			},
			exp: &v1alpha2.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionId),
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
								StorageURI: fmt.Sprintf("%s/model", model.ArtifactUri),
								Resources:  resourceRequests,
							},
							DeploymentSpec: v1alpha2.DeploymentSpec{
								MinReplicas: minReplica,
								MaxReplicas: maxReplica,
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
				ArtifactUri: model.ArtifactUri,
				Type:        models.ModelTypePyTorch,
				Options: &models.ModelOption{
					PyTorchModelClassName: "MyModel",
				},
				Metadata: model.Metadata,
			},
			exp: &v1alpha2.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionId),
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
								StorageURI:     fmt.Sprintf("%s/model", model.ArtifactUri),
								ModelClassName: "MyModel",
								Resources:      resourceRequests,
							},
							DeploymentSpec: v1alpha2.DeploymentSpec{
								MinReplicas: minReplica,
								MaxReplicas: maxReplica,
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
				ArtifactUri: model.ArtifactUri,
				Type:        models.ModelTypeOnnx,
				Options:     &models.ModelOption{},
				Metadata:    model.Metadata,
			},
			exp: &v1alpha2.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionId),
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
								StorageURI: fmt.Sprintf("%s/model", model.ArtifactUri),
								Resources:  resourceRequests,
							},
							DeploymentSpec: v1alpha2.DeploymentSpec{
								MinReplicas: minReplica,
								MaxReplicas: maxReplica,
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
				ArtifactUri: model.ArtifactUri,
				Type:        models.ModelTypePyFunc,
				Options: &models.ModelOption{
					PyFuncImageName: "gojek/project-model:1",
				},
				EnvVars:  models.PyfuncDefaultEnvVars(models.Model{Name: model.Name}, models.Version{Id: models.Id(1), ArtifactUri: model.ArtifactUri}, cpuRequest.Value()),
				Metadata: model.Metadata,
			},

			exp: &v1alpha2.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionId),
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
									Env:       models.PyfuncDefaultEnvVars(models.Model{Name: model.Name}, models.Version{Id: models.Id(1), ArtifactUri: model.ArtifactUri}, cpuRequest.Value()).ToKubernetesEnvVars(),
									Resources: resourceRequests,
								},
							},
							DeploymentSpec: v1alpha2.DeploymentSpec{
								MinReplicas: minReplica,
								MaxReplicas: maxReplica,
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
				MinReplica:              minReplica,
				MaxReplica:              maxReplica,
				CpuRequest:              cpuRequest,
				CpuLimit:                cpuLimit,
				MemoryRequest:           memoryRequest,
				MemoryLimit:             memoryLimit,
				QueueResourcePercentage: queueResourcePercentage,
			}

			infSvcSpec := createInferenceServiceSpec(tt.modelSvc, deployConfig)
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
		ArtifactUri: "gs://my-artifacet",
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
	versionId := 1

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
				ArtifactUri: model.ArtifactUri,
				Type:        models.ModelTypeTensorflow,
				Options:     &models.ModelOption{},
				Metadata:    model.Metadata,
			},
			original: &v1alpha2.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionId),
					Namespace: project.Name,
					Annotations: map[string]string{
						"queue.sidecar.serving.knative.dev/resourcePercentage": queueResourcePercentage,
					},
				},
				Spec: v1alpha2.InferenceServiceSpec{
					Default: v1alpha2.EndpointSpec{
						Predictor: v1alpha2.PredictorSpec{
							Tensorflow: &v1alpha2.TensorflowSpec{
								StorageURI: fmt.Sprintf("%s/model", model.ArtifactUri),
								Resources:  resourceRequests,
							},
							DeploymentSpec: v1alpha2.DeploymentSpec{
								MinReplicas: minReplica,
								MaxReplicas: maxReplica,
							},
						},
					},
				},
			},
			exp: &v1alpha2.InferenceService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", model.Name, versionId),
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
								StorageURI: fmt.Sprintf("%s/model", model.ArtifactUri),
								Resources:  resourceRequests,
							},
							DeploymentSpec: v1alpha2.DeploymentSpec{
								MinReplicas: minReplica,
								MaxReplicas: maxReplica,
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
				MinReplica:              minReplica,
				MaxReplica:              maxReplica,
				CpuRequest:              cpuRequest,
				CpuLimit:                cpuLimit,
				MemoryRequest:           memoryRequest,
				MemoryLimit:             memoryLimit,
				QueueResourcePercentage: queueResourcePercentage,
			}

			infSvcSpec := patchInferenceServiceSpec(tt.original, tt.modelSvc, deployConfig)
			assert.Equal(t, tt.exp, infSvcSpec)
		})
	}
}
