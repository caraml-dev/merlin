package resource

import (
	"fmt"
	"testing"

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
)

var (
	defaultGPUNodeSelector = map[string]string{"cloud.google.com/gke-accelerator": "nvidia-tesla-p4"}

	defaultGPUTolerations = []corev1.Toleration{
		{
			Key:      "caraml/nvidia-tesla-p4",
			Operator: corev1.TolerationOpEqual,
			Value:    "enabled",
			Effect:   corev1.TaintEffectNoSchedule,
		},
		{
			Key:      "nvidia.com/gpu",
			Operator: corev1.TolerationOpEqual,
			Value:    "present",
			Effect:   corev1.TaintEffectNoSchedule,
		},
	}

	defaultGPUsConfig = []config.GPUConfig{
		{
			Name:                 "NVIDIA P4",
			Values:               []string{"None", "1", "2", "4"},
			ResourceType:         "nvidia.com/gpu",
			NodeSelector:         defaultGPUNodeSelector,
			Tolerations:          defaultGPUTolerations,
			MinMonthlyCostPerGPU: 332.15,
			MaxMonthlyCostPerGPU: 332.15,
		},
	}

	expDefaultModelResourceRequestsWithGPU = corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    defaultModelResourceRequests.CPURequest,
			corev1.ResourceMemory: defaultModelResourceRequests.MemoryRequest,
			"nvidia.com/gpu":      resource.MustParse("1"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("10"),
			corev1.ResourceMemory: ScaleQuantity(defaultModelResourceRequests.MemoryRequest, 2),
			"nvidia.com/gpu":      resource.MustParse("1"),
		},
	}
)

func TestCreateInferenceServiceSpecWithGPU(t *testing.T) {
	err := models.InitKubernetesLabeller("gojek.com/", testEnvironmentName)
	assert.NoError(t, err)

	defer func() {
		_ = models.InitKubernetesLabeller("", "")
	}()

	project := mlp.Project{
		Name: "project",
	}

	modelSvc := &models.Service{
		Name:         "my-model-1",
		ModelName:    "my-model",
		Namespace:    "project",
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
		ResourceRequest: &models.ResourceRequest{
			MinReplica:    1,
			MaxReplica:    2,
			CPURequest:    resource.MustParse("500m"),
			MemoryRequest: resource.MustParse("500Mi"),
			GPUName:       "NVIDIA P4",
			GPURequest:    resource.MustParse("1"),
		},
	}

	invalidResourceRequest := &models.ResourceRequest{
		MinReplica:    1,
		MaxReplica:    2,
		CPURequest:    resource.MustParse("500m"),
		MemoryRequest: resource.MustParse("500Mi"),
		GPUName:       "NVIDIA P4",
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
				Name:            modelSvc.Name,
				ModelName:       modelSvc.ModelName,
				ModelVersion:    modelSvc.ModelVersion,
				Namespace:       project.Name,
				ArtifactURI:     modelSvc.ArtifactURI,
				Type:            models.ModelTypeTensorflow,
				Options:         &models.ModelOption{},
				Metadata:        modelSvc.Metadata,
				Protocol:        protocol.HttpJson,
				ResourceRequest: modelSvc.ResourceRequest,
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
									Resources:     expDefaultModelResourceRequestsWithGPU,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
						PodSpec: kservev1beta1.PodSpec{
							NodeSelector: defaultGPUNodeSelector,
							Tolerations:  defaultGPUTolerations,
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
				ResourceRequest: modelSvc.ResourceRequest,
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
									Resources:     expDefaultModelResourceRequestsWithGPU,
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
						PodSpec: kservev1beta1.PodSpec{
							NodeSelector: defaultGPUNodeSelector,
							Tolerations:  defaultGPUTolerations,
						},
					},
				},
			},
		},
		{
			name: "tensorflow spec as raw deployment",
			modelSvc: &models.Service{
				Name:            modelSvc.Name,
				ModelName:       modelSvc.ModelName,
				ModelVersion:    modelSvc.ModelVersion,
				Namespace:       project.Name,
				ArtifactURI:     modelSvc.ArtifactURI,
				Type:            models.ModelTypeTensorflow,
				Options:         &models.ModelOption{},
				Metadata:        modelSvc.Metadata,
				DeploymentMode:  deployment.RawDeploymentMode,
				Protocol:        protocol.HttpJson,
				ResourceRequest: modelSvc.ResourceRequest,
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
									Resources:     expDefaultModelResourceRequestsWithGPU,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
						PodSpec: kservev1beta1.PodSpec{
							NodeSelector: defaultGPUNodeSelector,
							Tolerations:  defaultGPUTolerations,
						},
					},
				},
			},
		},
		{
			name: "tensorflow spec as serverless",
			modelSvc: &models.Service{
				Name:            modelSvc.Name,
				ModelName:       modelSvc.ModelName,
				ModelVersion:    modelSvc.ModelVersion,
				Namespace:       project.Name,
				ArtifactURI:     modelSvc.ArtifactURI,
				Type:            models.ModelTypeTensorflow,
				Options:         &models.ModelOption{},
				Metadata:        modelSvc.Metadata,
				DeploymentMode:  deployment.ServerlessDeploymentMode,
				Protocol:        protocol.HttpJson,
				ResourceRequest: modelSvc.ResourceRequest,
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
									Resources:     expDefaultModelResourceRequestsWithGPU,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
						PodSpec: kservev1beta1.PodSpec{
							NodeSelector: defaultGPUNodeSelector,
							Tolerations:  defaultGPUTolerations,
						},
					},
				},
			},
		},
		{
			name: "tensorflow spec without queue resource percentage",
			modelSvc: &models.Service{
				Name:            modelSvc.Name,
				ModelName:       modelSvc.ModelName,
				ModelVersion:    modelSvc.ModelVersion,
				Namespace:       project.Name,
				ArtifactURI:     modelSvc.ArtifactURI,
				Type:            models.ModelTypeTensorflow,
				Options:         &models.ModelOption{},
				Metadata:        modelSvc.Metadata,
				Protocol:        protocol.HttpJson,
				ResourceRequest: modelSvc.ResourceRequest,
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
									Resources:     expDefaultModelResourceRequestsWithGPU,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
						PodSpec: kservev1beta1.PodSpec{
							NodeSelector: defaultGPUNodeSelector,
							Tolerations:  defaultGPUTolerations,
						},
					},
				},
			},
		},
		{
			name: "xgboost spec",
			modelSvc: &models.Service{
				Name:            modelSvc.Name,
				ModelName:       modelSvc.ModelName,
				ModelVersion:    modelSvc.ModelVersion,
				Namespace:       project.Name,
				ArtifactURI:     modelSvc.ArtifactURI,
				Type:            models.ModelTypeXgboost,
				Options:         &models.ModelOption{},
				Metadata:        modelSvc.Metadata,
				Protocol:        protocol.HttpJson,
				ResourceRequest: modelSvc.ResourceRequest,
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
									Resources:     expDefaultModelResourceRequestsWithGPU,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
						PodSpec: kservev1beta1.PodSpec{
							NodeSelector: defaultGPUNodeSelector,
							Tolerations:  defaultGPUTolerations,
						},
					},
				},
			},
		},
		{
			name: "sklearn spec",
			modelSvc: &models.Service{
				Name:            modelSvc.Name,
				ModelName:       modelSvc.ModelName,
				ModelVersion:    modelSvc.ModelVersion,
				Namespace:       project.Name,
				ArtifactURI:     modelSvc.ArtifactURI,
				Type:            models.ModelTypeSkLearn,
				Options:         &models.ModelOption{},
				Metadata:        modelSvc.Metadata,
				Protocol:        protocol.HttpJson,
				ResourceRequest: modelSvc.ResourceRequest,
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
									Resources:     expDefaultModelResourceRequestsWithGPU,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
						PodSpec: kservev1beta1.PodSpec{
							NodeSelector: defaultGPUNodeSelector,
							Tolerations:  defaultGPUTolerations,
						},
					},
				},
			},
		},
		{
			name: "pytorch spec",
			modelSvc: &models.Service{
				Name:            modelSvc.Name,
				ModelName:       modelSvc.ModelName,
				ModelVersion:    modelSvc.ModelVersion,
				Namespace:       project.Name,
				ArtifactURI:     modelSvc.ArtifactURI,
				Type:            models.ModelTypePyTorch,
				Metadata:        modelSvc.Metadata,
				Protocol:        protocol.HttpJson,
				ResourceRequest: modelSvc.ResourceRequest,
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
									Resources:     expDefaultModelResourceRequestsWithGPU,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
						PodSpec: kservev1beta1.PodSpec{
							NodeSelector: defaultGPUNodeSelector,
							Tolerations:  defaultGPUTolerations,
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
				EnvVars:         models.EnvVars{models.EnvVar{Name: envOldDisableLivenessProbe, Value: "true"}},
				Metadata:        modelSvc.Metadata,
				Protocol:        protocol.HttpJson,
				ResourceRequest: modelSvc.ResourceRequest,
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
									Env: models.MergeEnvVars(models.EnvVars{models.EnvVar{Name: envOldDisableLivenessProbe, Value: "true"}},
										createPyFuncDefaultEnvVarsWithProtocol(modelSvc, protocol.HttpJson)).ToKubernetesEnvVars(),
									Resources: expDefaultModelResourceRequestsWithGPU,
								},
							},
							NodeSelector: defaultGPUNodeSelector,
							Tolerations:  defaultGPUTolerations,
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
				EnvVars:         models.EnvVars{models.EnvVar{Name: envDisableLivenessProbe, Value: "true"}},
				Metadata:        modelSvc.Metadata,
				Protocol:        protocol.HttpJson,
				ResourceRequest: modelSvc.ResourceRequest,
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
									Env: models.MergeEnvVars(models.EnvVars{models.EnvVar{Name: envDisableLivenessProbe, Value: "true"}},
										createPyFuncDefaultEnvVarsWithProtocol(modelSvc, protocol.HttpJson)).ToKubernetesEnvVars(),
									Resources: expDefaultModelResourceRequestsWithGPU,
								},
							},
							NodeSelector: defaultGPUNodeSelector,
							Tolerations:  defaultGPUTolerations,
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
				Metadata:        modelSvc.Metadata,
				Protocol:        protocol.HttpJson,
				ResourceRequest: modelSvc.ResourceRequest,
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
									Name:      kserveconstant.InferenceServiceContainerName,
									Image:     "gcr.io/custom-model:v0.1",
									Env:       createDefaultPredictorEnvVars(modelSvc).ToKubernetesEnvVars(),
									Resources: expDefaultModelResourceRequestsWithGPU,
								},
							},
							NodeSelector: defaultGPUNodeSelector,
							Tolerations:  defaultGPUTolerations,
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
				ResourceRequest: modelSvc.ResourceRequest,
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
									Name:      kserveconstant.InferenceServiceContainerName,
									Image:     "gcr.io/custom-model:v0.1",
									Env:       createDefaultPredictorEnvVars(modelSvc).ToKubernetesEnvVars(),
									Resources: expDefaultModelResourceRequestsWithGPU,
									Command: []string{
										"./run.sh",
									},
									Args: []string{
										"firstArg",
										"secondArg",
									},
								},
							},
							NodeSelector: defaultGPUNodeSelector,
							Tolerations:  defaultGPUTolerations,
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &modelSvc.ResourceRequest.MinReplica,
							MaxReplicas: modelSvc.ResourceRequest.MaxReplica,
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
				Protocol:        protocol.HttpJson,
				ResourceRequest: modelSvc.ResourceRequest,
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
									Resources:     expDefaultModelResourceRequestsWithGPU,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
						PodSpec: kservev1beta1.PodSpec{
							NodeSelector: defaultGPUNodeSelector,
							Tolerations:  defaultGPUTolerations,
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
				Protocol:        protocol.HttpJson,
				ResourceRequest: modelSvc.ResourceRequest,
			},
			resourcePercentage: queueResourcePercentage,
			deploymentScale:    defaultDeploymentScale,
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
				Protocol:        protocol.HttpJson,
				ResourceRequest: modelSvc.ResourceRequest,
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
						knautoscaling.ClassAnnotationKey:                      knautoscaling.HPA,
						knautoscaling.MetricAnnotationKey:                     knautoscaling.CPU,
						knautoscaling.TargetAnnotationKey:                     "30",
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
									Resources:     expDefaultModelResourceRequestsWithGPU,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
						PodSpec: kservev1beta1.PodSpec{
							NodeSelector: defaultGPUNodeSelector,
							Tolerations:  defaultGPUTolerations,
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
				Protocol:        protocol.HttpJson,
				ResourceRequest: modelSvc.ResourceRequest,
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
						knautoscaling.ClassAnnotationKey:                      knautoscaling.HPA,
						knautoscaling.MetricAnnotationKey:                     knautoscaling.Memory,
						knautoscaling.TargetAnnotationKey:                     "150",
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
									Resources:     expDefaultModelResourceRequestsWithGPU,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
						PodSpec: kservev1beta1.PodSpec{
							NodeSelector: defaultGPUNodeSelector,
							Tolerations:  defaultGPUTolerations,
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
				ResourceRequest: modelSvc.ResourceRequest,
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
						knautoscaling.ClassAnnotationKey:                      knautoscaling.HPA,
						knautoscaling.MetricAnnotationKey:                     knautoscaling.Memory,
						knautoscaling.TargetAnnotationKey:                     "100",
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
									Resources:     expDefaultModelResourceRequestsWithGPU,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &modelSvc.ResourceRequest.MinReplica,
							MaxReplicas: modelSvc.ResourceRequest.MaxReplica,
						},
						PodSpec: kservev1beta1.PodSpec{
							NodeSelector: defaultGPUNodeSelector,
							Tolerations:  defaultGPUTolerations,
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
				Protocol:        protocol.HttpJson,
				ResourceRequest: modelSvc.ResourceRequest,
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
						knautoscaling.ClassAnnotationKey:                      knautoscaling.KPA,
						knautoscaling.MetricAnnotationKey:                     knautoscaling.Concurrency,
						knautoscaling.TargetAnnotationKey:                     "2.00",
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
									Resources:     expDefaultModelResourceRequestsWithGPU,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
						PodSpec: kservev1beta1.PodSpec{
							NodeSelector: defaultGPUNodeSelector,
							Tolerations:  defaultGPUTolerations,
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
				Protocol:        protocol.HttpJson,
				ResourceRequest: modelSvc.ResourceRequest,
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
						knautoscaling.ClassAnnotationKey:                      knautoscaling.KPA,
						knautoscaling.MetricAnnotationKey:                     knautoscaling.RPS,
						knautoscaling.TargetAnnotationKey:                     "10",
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
									Resources:     expDefaultModelResourceRequestsWithGPU,
									LivenessProbe: probeConfig,
									Env:           []corev1.EnvVar{},
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
						PodSpec: kservev1beta1.PodSpec{
							NodeSelector: defaultGPUNodeSelector,
							Tolerations:  defaultGPUTolerations,
						},
					},
				},
			},
		},
		{
			name: "tensorflow upi v1",
			modelSvc: &models.Service{
				Name:            modelSvc.Name,
				ModelName:       modelSvc.ModelName,
				ModelVersion:    modelSvc.ModelVersion,
				Namespace:       project.Name,
				ArtifactURI:     modelSvc.ArtifactURI,
				Type:            models.ModelTypeTensorflow,
				Options:         &models.ModelOption{},
				Metadata:        modelSvc.Metadata,
				Protocol:        protocol.UpiV1,
				ResourceRequest: modelSvc.ResourceRequest,
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
									Resources:     expDefaultModelResourceRequestsWithGPU,
									Ports:         grpcServerlessContainerPorts,
									Env:           []corev1.EnvVar{},
									LivenessProbe: probeConfigUPI,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
						PodSpec: kservev1beta1.PodSpec{
							NodeSelector: defaultGPUNodeSelector,
							Tolerations:  defaultGPUTolerations,
						},
					},
				},
			},
		},
		{
			name: "pyfunc upi v1 raw",
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
				Metadata:        modelSvc.Metadata,
				Protocol:        protocol.UpiV1,
				ResourceRequest: modelSvc.ResourceRequest,
				DeploymentMode:  deployment.RawDeploymentMode,
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
						PodSpec: kservev1beta1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:          kserveconstant.InferenceServiceContainerName,
									Image:         "gojek/project-model:1",
									Resources:     expDefaultModelResourceRequestsWithGPU,
									Ports:         grpcRawContainerPorts,
									Env:           models.MergeEnvVars(models.EnvVars{models.EnvVar{Name: envGRPCOptions, Value: "{}"}}, createPyFuncDefaultEnvVarsWithProtocol(modelSvc, protocol.UpiV1)).ToKubernetesEnvVars(),
									LivenessProbe: probeConfigUPI,
								},
							},
							NodeSelector: defaultGPUNodeSelector,
							Tolerations:  defaultGPUTolerations,
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
				Name:            modelSvc.Name,
				ModelName:       modelSvc.ModelName,
				ModelVersion:    modelSvc.ModelVersion,
				Namespace:       project.Name,
				ArtifactURI:     modelSvc.ArtifactURI,
				Type:            models.ModelTypeXgboost,
				Options:         &models.ModelOption{},
				Metadata:        modelSvc.Metadata,
				Protocol:        protocol.UpiV1,
				ResourceRequest: modelSvc.ResourceRequest,
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
									Resources:     expDefaultModelResourceRequestsWithGPU,
									Ports:         grpcServerlessContainerPorts,
									Env:           []corev1.EnvVar{},
									LivenessProbe: probeConfigUPI,
								},
							},
						},
						ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
							MinReplicas: &defaultModelResourceRequests.MinReplica,
							MaxReplicas: defaultModelResourceRequests.MaxReplica,
						},
						PodSpec: kservev1beta1.PodSpec{
							NodeSelector: defaultGPUNodeSelector,
							Tolerations:  defaultGPUTolerations,
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
				Metadata:        modelSvc.Metadata,
				Protocol:        protocol.UpiV1,
				ResourceRequest: modelSvc.ResourceRequest,
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
									Name:      kserveconstant.InferenceServiceContainerName,
									Image:     "gcr.io/custom-model:v0.1",
									Env:       createDefaultPredictorEnvVars(modelSvc).ToKubernetesEnvVars(),
									Resources: expDefaultModelResourceRequestsWithGPU,
									Ports:     grpcServerlessContainerPorts,
								},
							},
							NodeSelector: defaultGPUNodeSelector,
							Tolerations:  defaultGPUTolerations,
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
			name: "invalid resource request with 0 GPU requested",
			modelSvc: &models.Service{
				Name:            modelSvc.Name,
				ModelName:       modelSvc.ModelName,
				ModelVersion:    modelSvc.ModelVersion,
				Namespace:       project.Name,
				ArtifactURI:     modelSvc.ArtifactURI,
				Type:            models.ModelTypeTensorflow,
				Options:         &models.ModelOption{},
				Metadata:        modelSvc.Metadata,
				Protocol:        protocol.HttpJson,
				ResourceRequest: invalidResourceRequest,
			},
			resourcePercentage: queueResourcePercentage,
			deploymentScale:    defaultDeploymentScale,
			exp:                &kservev1beta1.InferenceService{},
			wantErr:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deployConfig := &config.DeploymentConfig{
				DefaultModelResourceRequests:          defaultModelResourceRequests,
				DefaultTransformerResourceRequests:    defaultTransformerResourceRequests,
				QueueResourcePercentage:               tt.resourcePercentage,
				PyfuncGRPCOptions:                     "{}",
				GPUs:                                  defaultGPUsConfig,
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
