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
	"strings"

	kfsv1alpha2 "github.com/kubeflow/kfserving/pkg/apis/serving/v1alpha2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gojek/merlin/config"
	"github.com/gojek/merlin/models"
	transformerpkg "github.com/gojek/merlin/pkg/transformer"
	"github.com/gojek/merlin/utils"
)

const (
	envTransformerPort       = "MERLIN_TRANSFORMER_PORT"
	envTransformerModelName  = "MERLIN_TRANSFORMER_MODEL_NAME"
	envTransformerPredictURL = "MERLIN_TRANSFORMER_MODEL_PREDICT_URL"

	defaultTransformerPort = "8080"

	annotationQueueProxyResource   = "queue.sidecar.serving.knative.dev/resourcePercentage"
	annotationPrometheusScrapeFlag = "prometheus.io/scrape"
	annotationPrometheusScrapePort = "prometheus.io/port"

	labelTeamName         = "gojek.com/team"
	labelStreamName       = "gojek.com/stream"
	labelAppName          = "gojek.com/app"
	labelOrchestratorName = "gojek.com/orchestrator"
	labelEnvironment      = "gojek.com/environment"
	labelUsersHeading     = "gojek.com/user-labels/%s"

	prometheusPort = "8080"
)

type KFServingResourceTemplater struct {
	standardTransformerConfig config.StandardTransformerConfig
}

func NewKFServingResourceTemplater(standardTransformerConfig config.StandardTransformerConfig) *KFServingResourceTemplater {
	return &KFServingResourceTemplater{standardTransformerConfig: standardTransformerConfig}
}

func (t *KFServingResourceTemplater) CreateInferenceServiceSpec(modelService *models.Service, config *config.DeploymentConfig) *kfsv1alpha2.InferenceService {
	labels := createLabels(modelService)

	objectMeta := metav1.ObjectMeta{
		Name:      modelService.Name,
		Namespace: modelService.Namespace,
		Annotations: map[string]string{
			annotationQueueProxyResource: config.QueueResourcePercentage,
		},
		Labels: labels,
	}

	if modelService.Type == models.ModelTypePyFunc {
		objectMeta.Annotations[annotationPrometheusScrapeFlag] = "true"
		objectMeta.Annotations[annotationPrometheusScrapePort] = prometheusPort
	}

	inferenceService := &kfsv1alpha2.InferenceService{
		ObjectMeta: objectMeta,
		Spec: kfsv1alpha2.InferenceServiceSpec{
			Default: kfsv1alpha2.EndpointSpec{
				Predictor: createPredictorSpec(modelService, config),
			},
		},
	}

	if modelService.Transformer != nil && modelService.Transformer.Enabled {
		inferenceService.Spec.Default.Transformer = t.createTransformerSpec(modelService, modelService.Transformer, config)
	}

	return inferenceService
}

func (t *KFServingResourceTemplater) PatchInferenceServiceSpec(orig *kfsv1alpha2.InferenceService, modelService *models.Service, config *config.DeploymentConfig) *kfsv1alpha2.InferenceService {
	labels := createLabels(modelService)
	orig.ObjectMeta.Labels = labels
	orig.Spec.Default.Predictor = createPredictorSpec(modelService, config)

	orig.Spec.Default.Transformer = nil
	if modelService.Transformer != nil && modelService.Transformer.Enabled {
		orig.Spec.Default.Transformer = t.createTransformerSpec(modelService, modelService.Transformer, config)
	}
	return orig
}

func createPredictorSpec(modelService *models.Service, config *config.DeploymentConfig) kfsv1alpha2.PredictorSpec {
	var predictorSpec kfsv1alpha2.PredictorSpec

	if modelService.ResourceRequest == nil {
		modelService.ResourceRequest = &models.ResourceRequest{
			MinReplica:    config.DefaultModelResourceRequests.MinReplica,
			MaxReplica:    config.DefaultModelResourceRequests.MaxReplica,
			CPURequest:    config.DefaultModelResourceRequests.CPURequest,
			MemoryRequest: config.DefaultModelResourceRequests.MemoryRequest,
		}
	}

	// Set cpu limit and memory limit to be 2x of the requests
	cpuLimit := modelService.ResourceRequest.CPURequest.DeepCopy()
	cpuLimit.Add(modelService.ResourceRequest.CPURequest)
	memoryLimit := modelService.ResourceRequest.MemoryRequest.DeepCopy()
	memoryLimit.Add(modelService.ResourceRequest.MemoryRequest)

	Resources := v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    modelService.ResourceRequest.CPURequest,
			v1.ResourceMemory: modelService.ResourceRequest.MemoryRequest,
		},
		Limits: v1.ResourceList{
			v1.ResourceCPU:    cpuLimit,
			v1.ResourceMemory: memoryLimit,
		},
	}

	switch modelService.Type {
	case models.ModelTypeTensorflow:
		predictorSpec = kfsv1alpha2.PredictorSpec{
			Tensorflow: &kfsv1alpha2.TensorflowSpec{
				StorageURI: utils.CreateModelLocation(modelService.ArtifactURI),
				Resources:  Resources,
			},
		}
	case models.ModelTypeOnnx:
		predictorSpec = kfsv1alpha2.PredictorSpec{
			ONNX: &kfsv1alpha2.ONNXSpec{
				StorageURI: utils.CreateModelLocation(modelService.ArtifactURI),
				Resources:  Resources,
			},
		}
	case models.ModelTypeSkLearn:
		predictorSpec = kfsv1alpha2.PredictorSpec{
			SKLearn: &kfsv1alpha2.SKLearnSpec{
				StorageURI: utils.CreateModelLocation(modelService.ArtifactURI),
				Resources:  Resources,
			},
		}
	case models.ModelTypeXgboost:
		predictorSpec = kfsv1alpha2.PredictorSpec{
			XGBoost: &kfsv1alpha2.XGBoostSpec{
				StorageURI: utils.CreateModelLocation(modelService.ArtifactURI),
				Resources:  Resources,
			},
		}
	case models.ModelTypePyTorch:
		predictorSpec = kfsv1alpha2.PredictorSpec{
			PyTorch: &kfsv1alpha2.PyTorchSpec{
				StorageURI:     utils.CreateModelLocation(modelService.ArtifactURI),
				ModelClassName: modelService.Options.PyTorchModelClassName,
				Resources:      Resources,
			},
		}
	case models.ModelTypePyFunc:
		predictorSpec = kfsv1alpha2.PredictorSpec{
			Custom: &kfsv1alpha2.CustomSpec{
				Container: v1.Container{
					Image:     modelService.Options.PyFuncImageName,
					Env:       modelService.EnvVars.ToKubernetesEnvVars(),
					Resources: Resources,
				},
			},
		}
	}

	var loggerSpec *kfsv1alpha2.Logger
	if modelService.Logger != nil && modelService.Logger.Model != nil && modelService.Logger.Model.Enabled {
		logger := modelService.Logger
		loggerSpec = createLoggerSpec(logger.DestinationURL, *logger.Model)
	}

	predictorSpec.DeploymentSpec = kfsv1alpha2.DeploymentSpec{
		MinReplicas: &(modelService.ResourceRequest.MinReplica),
		MaxReplicas: modelService.ResourceRequest.MaxReplica,
		Logger:      loggerSpec,
	}

	return predictorSpec
}

func createLoggerSpec(loggerURL string, loggerConfig models.LoggerConfig) *kfsv1alpha2.Logger {
	loggerMode := models.ToKFServingLoggerMode(loggerConfig.Mode)
	return &kfsv1alpha2.Logger{
		Url:  &loggerURL,
		Mode: loggerMode,
	}
}

func (t *KFServingResourceTemplater) createTransformerSpec(modelService *models.Service, transformer *models.Transformer, config *config.DeploymentConfig) *kfsv1alpha2.TransformerSpec {
	if transformer.ResourceRequest == nil {
		transformer.ResourceRequest = &models.ResourceRequest{
			MinReplica:    config.DefaultTransformerResourceRequests.MinReplica,
			MaxReplica:    config.DefaultTransformerResourceRequests.MaxReplica,
			CPURequest:    config.DefaultTransformerResourceRequests.CPURequest,
			MemoryRequest: config.DefaultTransformerResourceRequests.MemoryRequest,
		}
	}

	// Set cpu limit and memory limit to be 2x of the requests
	cpuLimit := transformer.ResourceRequest.CPURequest.DeepCopy()
	cpuLimit.Add(transformer.ResourceRequest.CPURequest)
	memoryLimit := transformer.ResourceRequest.MemoryRequest.DeepCopy()
	memoryLimit.Add(transformer.ResourceRequest.MemoryRequest)

	envVars := transformer.EnvVars
	if transformer.TransformerType == models.StandardTransformerType {
		transformer.Image = t.standardTransformerConfig.ImageName
		envVars = append(envVars, models.EnvVar{Name: transformerpkg.FeastServingAddressEnvName, Value: t.standardTransformerConfig.FeastServingAddress})
		envVars = append(envVars, models.EnvVar{Name: transformerpkg.FeastServingPortEnvName, Value: string(t.standardTransformerConfig.FeastServingPort)})
	}

	envVars = append(envVars, models.EnvVar{Name: envTransformerPort, Value: defaultTransformerPort})
	envVars = append(envVars, models.EnvVar{Name: envTransformerModelName, Value: modelService.Name})
	envVars = append(envVars, models.EnvVar{Name: envTransformerPredictURL, Value: createPredictURL(modelService)})

	var loggerSpec *kfsv1alpha2.Logger
	if modelService.Logger != nil && modelService.Logger.Transformer != nil && modelService.Logger.Transformer.Enabled {
		logger := modelService.Logger
		loggerSpec = createLoggerSpec(logger.DestinationURL, *logger.Transformer)
	}

	transformerSpec := &kfsv1alpha2.TransformerSpec{
		Custom: &kfsv1alpha2.CustomSpec{
			Container: v1.Container{
				Name:  "transformer",
				Image: transformer.Image,
				Env:   envVars.ToKubernetesEnvVars(),
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceCPU:    transformer.ResourceRequest.CPURequest,
						v1.ResourceMemory: transformer.ResourceRequest.MemoryRequest,
					},
					Limits: v1.ResourceList{
						v1.ResourceCPU:    cpuLimit,
						v1.ResourceMemory: memoryLimit,
					},
				},
			},
		},
		DeploymentSpec: kfsv1alpha2.DeploymentSpec{
			MinReplicas: &(transformer.ResourceRequest.MinReplica),
			MaxReplicas: transformer.ResourceRequest.MaxReplica,
			Logger:      loggerSpec,
		},
	}

	if transformer.Command != "" {
		command := strings.Split(transformer.Command, " ")
		if len(command) > 0 {
			transformerSpec.Custom.Container.Command = command
		}
	}
	if transformer.Args != "" {
		args := strings.Split(transformer.Args, " ")
		if len(args) > 0 {
			transformerSpec.Custom.Container.Args = args
		}
	}

	return transformerSpec
}

func createLabels(modelService *models.Service) map[string]string {
	labels := map[string]string{
		labelTeamName:         modelService.Metadata.Team,
		labelStreamName:       modelService.Metadata.Stream,
		labelAppName:          modelService.Metadata.App,
		labelOrchestratorName: "merlin",
		labelEnvironment:      modelService.Metadata.Environment,
	}

	for _, label := range modelService.Metadata.Labels {
		labels[fmt.Sprintf(labelUsersHeading, label.Key)] = label.Value
	}

	return labels
}

func createPredictURL(modelService *models.Service) string {
	return modelService.Name + "-predictor-default." + modelService.Namespace
}
