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
	"bytes"
	"encoding/json"
	"strings"

	kservev1beta1 "github.com/kserve/kserve/pkg/apis/serving/v1beta1"
	"github.com/kserve/kserve/pkg/constants"
	corev1 "k8s.io/api/core/v1"
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

	envPredictorPort             = "MERLIN_PREDICTOR_PORT"
	envPredictorModelName        = "MERLIN_MODEL_NAME"
	envPredictorArtifactLocation = "MERLIN_ARTIFACT_LOCATION"
	envPredictorStorageURI       = "STORAGE_URI"

	defaultPredictorArtifactLocation = "/mnt/models"
	defaultPredictorPort             = "8080"
	defaultPredictorContainerName    = "kfserving-container"

	defaultTransformerPort = "8080"

	annotationQueueProxyResource   = "queue.sidecar.serving.knative.dev/resourcePercentage"
	annotationPrometheusScrapeFlag = "prometheus.io/scrape"
	annotationPrometheusScrapePort = "prometheus.io/port"

	prometheusPort = "8080"
)

type KFServingResourceTemplater struct {
	standardTransformerConfig config.StandardTransformerConfig
}

func NewKFServingResourceTemplater(standardTransformerConfig config.StandardTransformerConfig) *KFServingResourceTemplater {
	return &KFServingResourceTemplater{standardTransformerConfig: standardTransformerConfig}
}

func (t *KFServingResourceTemplater) CreateInferenceServiceSpec(modelService *models.Service, config *config.DeploymentConfig) *kservev1beta1.InferenceService {
	labels := modelService.Metadata.ToLabel()

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

	inferenceService := &kservev1beta1.InferenceService{
		ObjectMeta: objectMeta,
		Spec: kservev1beta1.InferenceServiceSpec{
			Predictor: createPredictorSpec(modelService, config),
		},
	}

	if modelService.Transformer != nil && modelService.Transformer.Enabled {
		inferenceService.Spec.Transformer = t.createTransformerSpec(modelService, modelService.Transformer, config)
	}

	return inferenceService
}

func (t *KFServingResourceTemplater) PatchInferenceServiceSpec(orig *kservev1beta1.InferenceService, modelService *models.Service, config *config.DeploymentConfig) *kservev1beta1.InferenceService {
	labels := modelService.Metadata.ToLabel()
	orig.ObjectMeta.Labels = labels
	orig.Spec.Predictor = createPredictorSpec(modelService, config)

	orig.Spec.Transformer = nil
	if modelService.Transformer != nil && modelService.Transformer.Enabled {
		orig.Spec.Transformer = t.createTransformerSpec(modelService, modelService.Transformer, config)
	}
	return orig
}

func createPredictorSpec(modelService *models.Service, config *config.DeploymentConfig) kservev1beta1.PredictorSpec {
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

	resources := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    modelService.ResourceRequest.CPURequest,
			corev1.ResourceMemory: modelService.ResourceRequest.MemoryRequest,
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    cpuLimit,
			corev1.ResourceMemory: memoryLimit,
		},
	}

	var predictorSpec kservev1beta1.PredictorSpec

	storageUri := utils.CreateModelLocation(modelService.ArtifactURI)
	switch modelService.Type {
	case models.ModelTypeTensorflow:
		predictorSpec = kservev1beta1.PredictorSpec{
			Tensorflow: &kservev1beta1.TFServingSpec{
				PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
					StorageURI: &storageUri,
					Container: corev1.Container{
						Name:      constants.InferenceServiceContainerName,
						Resources: resources,
					},
				},
			},
		}
	case models.ModelTypeOnnx:
		predictorSpec = kservev1beta1.PredictorSpec{
			ONNX: &kservev1beta1.ONNXRuntimeSpec{
				PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
					StorageURI: &storageUri,
					Container: corev1.Container{
						Name:      constants.InferenceServiceContainerName,
						Resources: resources,
					},
				},
			},
		}
	case models.ModelTypeSkLearn:
		predictorSpec = kservev1beta1.PredictorSpec{
			SKLearn: &kservev1beta1.SKLearnSpec{
				PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
					StorageURI: &storageUri,
					Container: corev1.Container{
						Name:      constants.InferenceServiceContainerName,
						Resources: resources,
					},
				},
			},
		}
	case models.ModelTypeXgboost:
		predictorSpec = kservev1beta1.PredictorSpec{
			XGBoost: &kservev1beta1.XGBoostSpec{
				PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
					StorageURI: &storageUri,
					Container: corev1.Container{
						Name:      constants.InferenceServiceContainerName,
						Resources: resources,
					},
				},
			},
		}
	case models.ModelTypePyTorch:
		predictorSpec = kservev1beta1.PredictorSpec{
			PyTorch: &kservev1beta1.TorchServeSpec{
				PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
					StorageURI: &storageUri,
					Container: corev1.Container{
						Name:      constants.InferenceServiceContainerName,
						Resources: resources,
					},
				},
			},
		}
	case models.ModelTypePyFunc:
		predictorSpec = kservev1beta1.PredictorSpec{
			PodSpec: kservev1beta1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:      constants.InferenceServiceContainerName,
						Image:     modelService.Options.PyFuncImageName,
						Env:       modelService.EnvVars.ToKubernetesEnvVars(),
						Resources: resources,
					},
				},
			},
		}
	case models.ModelTypeCustom:
		predictorSpec = createCustomPredictorSpec(modelService, resources)
	}

	var loggerSpec *kservev1beta1.LoggerSpec
	if modelService.Logger != nil && modelService.Logger.Model != nil && modelService.Logger.Model.Enabled {
		logger := modelService.Logger
		loggerSpec = createLoggerSpec(logger.DestinationURL, *logger.Model)
	}

	predictorSpec.MinReplicas = &(modelService.ResourceRequest.MinReplica)
	predictorSpec.MaxReplicas = modelService.ResourceRequest.MaxReplica
	predictorSpec.Logger = loggerSpec

	return predictorSpec
}

func createLoggerSpec(loggerURL string, loggerConfig models.LoggerConfig) *kservev1beta1.LoggerSpec {
	loggerMode := models.ToKFServingLoggerMode(loggerConfig.Mode)
	return &kservev1beta1.LoggerSpec{
		URL:  &loggerURL,
		Mode: loggerMode,
	}
}

func createCustomPredictorSpec(modelService *models.Service, resources corev1.ResourceRequirements) kservev1beta1.PredictorSpec {
	envVars := modelService.EnvVars
	envVars = append(envVars, models.EnvVar{Name: envPredictorPort, Value: defaultPredictorPort})
	envVars = append(envVars, models.EnvVar{Name: envPredictorModelName, Value: modelService.Name})
	envVars = append(envVars, models.EnvVar{Name: envPredictorArtifactLocation, Value: defaultPredictorArtifactLocation})
	envVars = append(envVars, models.EnvVar{Name: envPredictorStorageURI, Value: utils.CreateModelLocation(modelService.ArtifactURI)})

	var containerCommand []string
	customPredictor := modelService.Options.CustomPredictor
	if customPredictor.Command != "" {
		command := strings.Split(customPredictor.Command, " ")
		if len(command) > 0 {
			containerCommand = command
		}
	}

	var containerArgs []string
	if customPredictor.Args != "" {
		args := strings.Split(customPredictor.Args, " ")
		if len(args) > 0 {
			containerArgs = args
		}
	}
	return kservev1beta1.PredictorSpec{
		PodSpec: kservev1beta1.PodSpec{
			Containers: []corev1.Container{
				{

					Image:     customPredictor.Image,
					Env:       envVars.ToKubernetesEnvVars(),
					Resources: resources,
					Command:   containerCommand,
					Args:      containerArgs,
					Name:      constants.InferenceServiceContainerName,
				},
			},
		},
	}
}

func (t *KFServingResourceTemplater) createTransformerSpec(modelService *models.Service, transformer *models.Transformer, config *config.DeploymentConfig) *kservev1beta1.TransformerSpec {
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
		envVars = t.enrichStandardTransformerEnvVars(envVars)
	}

	envVars = append(envVars, models.EnvVar{Name: envTransformerPort, Value: defaultTransformerPort})
	envVars = append(envVars, models.EnvVar{Name: envTransformerModelName, Value: modelService.Name})
	envVars = append(envVars, models.EnvVar{Name: envTransformerPredictURL, Value: createPredictURL(modelService)})

	var loggerSpec *kservev1beta1.LoggerSpec
	if modelService.Logger != nil && modelService.Logger.Transformer != nil && modelService.Logger.Transformer.Enabled {
		logger := modelService.Logger
		loggerSpec = createLoggerSpec(logger.DestinationURL, *logger.Transformer)
	}

	var transformerCommand []string
	var transformerArgs []string
	if transformer.Command != "" {
		command := strings.Split(transformer.Command, " ")
		if len(command) > 0 {
			transformerCommand = command
		}
	}
	if transformer.Args != "" {
		args := strings.Split(transformer.Args, " ")
		if len(args) > 0 {
			transformerArgs = args
		}
	}

	transformerSpec := &kservev1beta1.TransformerSpec{
		PodSpec: kservev1beta1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "transformer",
					Image: transformer.Image,
					Env:   envVars.ToKubernetesEnvVars(),
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    transformer.ResourceRequest.CPURequest,
							corev1.ResourceMemory: transformer.ResourceRequest.MemoryRequest,
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    cpuLimit,
							corev1.ResourceMemory: memoryLimit,
						},
					},
					Command: transformerCommand,
					Args:    transformerArgs,
				},
			},
		},
		ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
			MinReplicas: &(transformer.ResourceRequest.MinReplica),
			MaxReplicas: transformer.ResourceRequest.MaxReplica,
			Logger:      loggerSpec,
		},
	}

	return transformerSpec
}

func (t *KFServingResourceTemplater) enrichStandardTransformerEnvVars(envVars models.EnvVars) models.EnvVars {
	// compact standard transformer config
	envVarsMap := envVars.ToMap()
	standardTransformerSpec := envVarsMap[transformerpkg.StandardTransformerConfigEnvName]
	if standardTransformerSpec != "" {
		compactedfJsonBuffer := new(bytes.Buffer)
		if err := json.Compact(compactedfJsonBuffer, []byte(standardTransformerSpec)); err == nil {
			models.MergeEnvVars(envVars, models.EnvVars{
				{
					Name:  transformerpkg.StandardTransformerConfigEnvName,
					Value: compactedfJsonBuffer.String(),
				},
			})
		}
	}

	envVars = append(envVars, models.EnvVar{Name: transformerpkg.DefaultFeastSource, Value: t.standardTransformerConfig.DefaultFeastSource.String()})

	// adding feast storage config env variable
	feastStorageConfig := t.standardTransformerConfig.ToFeastStorageConfigs()

	if feastStorageConfigJsonByte, err := json.Marshal(feastStorageConfig); err == nil {
		envVars = append(envVars, models.EnvVar{Name: transformerpkg.FeastStorageConfigs, Value: string(feastStorageConfigJsonByte)})
	}

	jaegerCfg := t.standardTransformerConfig.Jaeger
	jaegerEnvVars := []models.EnvVar{
		{Name: transformerpkg.JaegerAgentHost, Value: jaegerCfg.AgentHost},
		{Name: transformerpkg.JaegerAgentPort, Value: jaegerCfg.AgentPort},
		{Name: transformerpkg.JaegerSamplerParam, Value: jaegerCfg.SamplerParam},
		{Name: transformerpkg.JaegerSamplerType, Value: jaegerCfg.SamplerType},
		{Name: transformerpkg.JaegerDisabled, Value: jaegerCfg.Disabled},
	}
	// We want Jaeger's env vars to be on the first order so these values could be overriden by users.
	envVars = append(jaegerEnvVars, envVars...)
	return envVars
}

func createPredictURL(modelService *models.Service) string {
	return modelService.Name + "-predictor-default." + modelService.Namespace
}
