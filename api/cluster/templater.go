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
	"fmt"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"

	"github.com/gojek/merlin/config"
	"github.com/gojek/merlin/models"
	"github.com/gojek/merlin/pkg/autoscaling"
	"github.com/gojek/merlin/pkg/deployment"
	transformerpkg "github.com/gojek/merlin/pkg/transformer"
	kservev1beta1 "github.com/kserve/kserve/pkg/apis/serving/v1beta1"
	kserveconstant "github.com/kserve/kserve/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	knautoscaling "knative.dev/serving/pkg/apis/autoscaling"
	knserving "knative.dev/serving/pkg/apis/serving"

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
	defaultTransformerPort           = "8080"

	annotationPrometheusScrapeFlag = "prometheus.io/scrape"
	annotationPrometheusScrapePort = "prometheus.io/port"

	prometheusPort = "8080"
)

var (
	// list of configuration stored as annotations
	configAnnotationkeys = []string{
		annotationPrometheusScrapeFlag,
		annotationPrometheusScrapePort,
		knserving.QueueSidecarResourcePercentageAnnotationKey,
		kserveconstant.AutoscalerClass,
		kserveconstant.AutoscalerMetrics,
		kserveconstant.TargetUtilizationPercentage,
		knautoscaling.ClassAnnotationKey,
		knautoscaling.MetricAnnotationKey,
		knautoscaling.TargetAnnotationKey,
	}
)

type KFServingResourceTemplater struct {
	standardTransformerConfig config.StandardTransformerConfig
}

func NewKFServingResourceTemplater(standardTransformerConfig config.StandardTransformerConfig) *KFServingResourceTemplater {
	return &KFServingResourceTemplater{standardTransformerConfig: standardTransformerConfig}
}

func (t *KFServingResourceTemplater) CreateInferenceServiceSpec(modelService *models.Service, config *config.DeploymentConfig) (*kservev1beta1.InferenceService, error) {
	annotations, err := createAnnotations(modelService, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create inference service spec: %w", err)
	}

	objectMeta := metav1.ObjectMeta{
		Name:        modelService.Name,
		Namespace:   modelService.Namespace,
		Labels:      modelService.Metadata.ToLabel(),
		Annotations: annotations,
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

	return inferenceService, nil
}

func (t *KFServingResourceTemplater) PatchInferenceServiceSpec(orig *kservev1beta1.InferenceService, modelService *models.Service, config *config.DeploymentConfig) (*kservev1beta1.InferenceService, error) {
	orig.ObjectMeta.Labels = modelService.Metadata.ToLabel()
	annotations, err := createAnnotations(modelService, config)
	if err != nil {
		return nil, fmt.Errorf("unable to patch inference service spec: %w", err)
	}
	orig.ObjectMeta.Annotations = utils.MergeMaps(utils.ExcludeKeys(orig.ObjectMeta.Annotations, configAnnotationkeys), annotations)
	orig.Spec.Predictor = createPredictorSpec(modelService, config)

	orig.Spec.Transformer = nil
	if modelService.Transformer != nil && modelService.Transformer.Enabled {
		orig.Spec.Transformer = t.createTransformerSpec(modelService, modelService.Transformer, config)
	}
	return orig, nil
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
						Name:      kserveconstant.InferenceServiceContainerName,
						Resources: resources,
						LivenessProbe: &corev1.Probe{
							Handler: corev1.Handler{
								TCPSocket: &corev1.TCPSocketAction{
									Port: intstr.IntOrString{
										IntVal: 9000,
									},
								},
							},
							InitialDelaySeconds: 10,
							TimeoutSeconds:      5,
							PeriodSeconds:       10,
							SuccessThreshold:    1,
						},
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
						Name:      kserveconstant.InferenceServiceContainerName,
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
						Name:      kserveconstant.InferenceServiceContainerName,
						Resources: resources,
						LivenessProbe: &corev1.Probe{
							Handler: corev1.Handler{
								HTTPGet: &corev1.HTTPGetAction{
									Path:   fmt.Sprintf("/v1/models/%s", modelService.Name),
									Scheme: "HTTP",
								},
							},
							InitialDelaySeconds: 10,
							TimeoutSeconds:      5,
							PeriodSeconds:       10,
							SuccessThreshold:    1,
						},
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
						Name:      kserveconstant.InferenceServiceContainerName,
						Resources: resources,
						LivenessProbe: &corev1.Probe{
							Handler: corev1.Handler{
								HTTPGet: &corev1.HTTPGetAction{
									Path:   fmt.Sprintf("/v1/models/%s", modelService.Name),
									Scheme: "HTTP",
								},
							},
							InitialDelaySeconds: 10,
							TimeoutSeconds:      5,
							PeriodSeconds:       10,
							SuccessThreshold:    1,
						},
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
						Name:      kserveconstant.InferenceServiceContainerName,
						Resources: resources,
						LivenessProbe: &corev1.Probe{
							Handler: corev1.Handler{
								HTTPGet: &corev1.HTTPGetAction{
									Path:   fmt.Sprintf("/v1/models/%s", modelService.Name),
									Scheme: "HTTP",
								},
							},
							InitialDelaySeconds: 10,
							TimeoutSeconds:      5,
							PeriodSeconds:       10,
							SuccessThreshold:    1,
						},
					},
				},
			},
		}
	case models.ModelTypePyFunc:
		predictorSpec = kservev1beta1.PredictorSpec{
			PodSpec: kservev1beta1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:      kserveconstant.InferenceServiceContainerName,
						Image:     modelService.Options.PyFuncImageName,
						Env:       modelService.EnvVars.ToKubernetesEnvVars(),
						Resources: resources,
						LivenessProbe: &corev1.Probe{
							Handler: corev1.Handler{
								HTTPGet: &corev1.HTTPGetAction{
									Path:   fmt.Sprintf("/v1/models/%s", modelService.Name),
									Scheme: "HTTP",
								},
							},
							InitialDelaySeconds: 10,
							TimeoutSeconds:      5,
							PeriodSeconds:       10,
							SuccessThreshold:    1,
						},
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
					Name:      kserveconstant.InferenceServiceContainerName,
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

func createAnnotations(modelService *models.Service, config *config.DeploymentConfig) (map[string]string, error) {
	annotations := make(map[string]string)

	if config.QueueResourcePercentage != "" {
		annotations[knserving.QueueSidecarResourcePercentageAnnotationKey] = config.QueueResourcePercentage
	}

	if modelService.Type == models.ModelTypePyFunc {
		annotations[annotationPrometheusScrapeFlag] = "true"
		annotations[annotationPrometheusScrapePort] = prometheusPort
	}

	deployMode, err := toKServeDeploymentMode(modelService.DeploymentMode)
	if err != nil {
		return nil, err
	}
	annotations[kserveconstant.DeploymentMode] = deployMode

	if modelService.AutoscalingPolicy != nil {
		if modelService.DeploymentMode == deployment.RawDeploymentMode {
			annotations[kserveconstant.AutoscalerClass] = string(kserveconstant.AutoscalerClassHPA)
			autoscalingMetrics, err := toKServeAutoscalerMetrics(modelService.AutoscalingPolicy.MetricsType)
			if err != nil {
				return nil, err
			}

			annotations[kserveconstant.AutoscalerMetrics] = autoscalingMetrics
			annotations[kserveconstant.TargetUtilizationPercentage] = fmt.Sprintf("%.0f", modelService.AutoscalingPolicy.TargetValue)
		} else if modelService.DeploymentMode == deployment.ServerlessDeploymentMode {
			if modelService.AutoscalingPolicy.MetricsType == autoscaling.CPUUtilization ||
				modelService.AutoscalingPolicy.MetricsType == autoscaling.MemoryUtilization {
				annotations[knautoscaling.ClassAnnotationKey] = knautoscaling.HPA
			} else {
				annotations[knautoscaling.ClassAnnotationKey] = knautoscaling.KPA
			}

			autoscalingMetrics, err := toKNativeAutoscalerMetrics(modelService.AutoscalingPolicy.MetricsType)
			if err != nil {
				return nil, err
			}

			annotations[knautoscaling.MetricAnnotationKey] = autoscalingMetrics
			annotations[knautoscaling.TargetAnnotationKey] = fmt.Sprintf("%.0f", modelService.AutoscalingPolicy.TargetValue)
		}
	}

	return annotations, nil
}

func toKNativeAutoscalerMetrics(metricsType autoscaling.MetricsType) (string, error) {
	switch metricsType {
	case autoscaling.CPUUtilization:
		return knautoscaling.CPU, nil
	case autoscaling.MemoryUtilization:
		return knautoscaling.Memory, nil
	case autoscaling.RPS:
		return knautoscaling.RPS, nil
	case autoscaling.Concurrency:
		return knautoscaling.Concurrency, nil
	default:
		return "", fmt.Errorf("unsuppported autoscaler metrics on serverless deployment: %s", metricsType)
	}
}

func toKServeAutoscalerMetrics(metricsType autoscaling.MetricsType) (string, error) {
	switch metricsType {
	case autoscaling.CPUUtilization:
		return string(kserveconstant.AutoScalerMetricsCPU), nil
	default:
		return "", fmt.Errorf("unsupported autoscaler metrics on raw deployment: %s", metricsType)
	}
}

func toKServeDeploymentMode(deploymentMode deployment.Mode) (string, error) {
	switch deploymentMode {
	case deployment.RawDeploymentMode:
		return string(kserveconstant.RawDeployment), nil
	case deployment.ServerlessDeploymentMode, deployment.EmptyDeploymentMode:
		return string(kserveconstant.Serverless), nil
	default:
		return "", fmt.Errorf("unsupported deployment mode: %s", deploymentMode)
	}
}
