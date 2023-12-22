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
	"strconv"
	"strings"

	kservev1beta1 "github.com/kserve/kserve/pkg/apis/serving/v1beta1"
	kserveconstant "github.com/kserve/kserve/pkg/constants"
	"github.com/mitchellh/copystructure"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	knautoscaling "knative.dev/serving/pkg/apis/autoscaling"
	knserving "knative.dev/serving/pkg/apis/serving"

	"github.com/caraml-dev/merlin/config"
	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/pkg/autoscaling"
	"github.com/caraml-dev/merlin/pkg/deployment"
	"github.com/caraml-dev/merlin/pkg/protocol"
	prt "github.com/caraml-dev/merlin/pkg/protocol"
	transformerpkg "github.com/caraml-dev/merlin/pkg/transformer"
	"github.com/caraml-dev/merlin/utils"
)

const (
	// TODO: Deprecate following env vars
	envTransformerPort       = "MERLIN_TRANSFORMER_PORT"
	envTransformerModelName  = "MERLIN_TRANSFORMER_MODEL_NAME"
	envTransformerPredictURL = "MERLIN_TRANSFORMER_MODEL_PREDICT_URL"

	envPyFuncModelName           = "MODEL_NAME"
	envPredictorPort             = "MERLIN_PREDICTOR_PORT"
	envPredictorModelName        = "MERLIN_MODEL_NAME"
	envPredictorArtifactLocation = "MERLIN_ARTIFACT_LOCATION"
	envOldDisableLivenessProbe   = "MERLIN_DISABLE_LIVENESS_PROBE"

	envProtocol             = "CARAML_PROTOCOL"
	envHTTPPort             = "CARAML_HTTP_PORT"
	envGRPCPort             = "CARAML_GRPC_PORT"
	envModelName            = "CARAML_MODEL_NAME"
	envModelVersion         = "CARAML_MODEL_VERSION"
	envModelFullName        = "CARAML_MODEL_FULL_NAME"
	envProject              = "CARAML_PROJECT"
	envPredictorHost        = "CARAML_PREDICTOR_HOST"
	envArtifactLocation     = "CARAML_ARTIFACT_LOCATION"
	envDisableLivenessProbe = "CARAML_DISABLE_LIVENESS_PROBE"
	envStorageURI           = "STORAGE_URI"
	envGRPCOptions          = "GRPC_OPTIONS"

	annotationPrometheusScrapeFlag = "prometheus.io/scrape"
	annotationPrometheusScrapePort = "prometheus.io/port"

	liveProbeInitialDelaySec  = 30
	liveProbeTimeoutSec       = 5
	liveProbePeriodSec        = 10
	liveProbeSuccessThreshold = 1
	liveProbeFailureThreshold = 3

	defaultPredictorArtifactLocation = "/mnt/models"
	defaultHTTPPort                  = 8080
	defaultGRPCPort                  = 9000
	defaultPredictorPort             = 80

	envPublisherKafkaTopic    = "PUBLISHER_KAFKA_TOPIC"
	envPublisherKafkaBrokers  = "PUBLISHER_KAFKA_BROKERS"
	envPublisherEnabled       = "PUBLISHER_ENABLED"
	envPublisherKafkaLinger   = "PUBLISHER_KAFKA_LINGER_MS"
	envPublisherKafkaAck      = "PUBLISHER_KAFKA_ACKS"
	envPublisherSamplingRatio = "PUBLISHER_SAMPLING_RATIO"
	envPublisherKafkaConfig   = "PUBLISHER_KAFKA_CONFIG"

	grpcHealthProbeCommand = "grpc_health_probe"
)

var grpcKPAContainerPorts = []corev1.ContainerPort{
	{
		ContainerPort: defaultGRPCPort,
		Name:          "h2c",
		Protocol:      corev1.ProtocolTCP,
	},
}

var grpcHPAContainerPorts = []corev1.ContainerPort{
	{
		ContainerPort: defaultGRPCPort,
		Name:          "grpc",
		Protocol:      corev1.ProtocolTCP,
	},
}

type DeploymentScale struct {
	Predictor   *int
	Transformer *int
}

type InferenceServiceTemplater struct {
	deploymentConfig config.DeploymentConfig
}

func NewInferenceServiceTemplater(deploymentCfg config.DeploymentConfig) *InferenceServiceTemplater {
	return &InferenceServiceTemplater{deploymentConfig: deploymentCfg}
}

func (t *InferenceServiceTemplater) CreateInferenceServiceSpec(modelService *models.Service, currentReplicas DeploymentScale) (*kservev1beta1.InferenceService, error) {
	t.applyDefaults(modelService)

	// Identify the desired initial scale of the new deployment
	var initialScale *int
	if currentReplicas.Predictor != nil {
		// The desired scale of the new deployment is a single value, applicable to both the predictor and the transformer.
		// Set the desired scale of the new deployment by taking the max of the 2 values.
		// Consider the transformer's scale only if it is also enabled in the new spec.
		if modelService.Transformer != nil &&
			modelService.Transformer.Enabled &&
			currentReplicas.Transformer != nil &&
			*currentReplicas.Transformer > *currentReplicas.Predictor {
			initialScale = currentReplicas.Transformer
		} else {
			initialScale = currentReplicas.Predictor
		}
	}

	annotations, err := t.createAnnotations(modelService, initialScale)
	if err != nil {
		return nil, fmt.Errorf("unable to create inference service spec: %w", err)
	}

	objectMeta := metav1.ObjectMeta{
		Name:        modelService.Name,
		Namespace:   modelService.Namespace,
		Labels:      modelService.Metadata.ToLabel(),
		Annotations: annotations,
	}

	predictorSpec := t.createPredictorSpec(modelService)
	predictorSpec.TopologySpreadConstraints, err = t.createNewInferenceServiceTopologySpreadConstraints(
		modelService,
		kservev1beta1.PredictorComponent,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create predictor topology spread constraints: %w", err)
	}

	inferenceService := &kservev1beta1.InferenceService{
		ObjectMeta: objectMeta,
		Spec: kservev1beta1.InferenceServiceSpec{
			Predictor: predictorSpec,
		},
	}

	if modelService.Transformer != nil && modelService.Transformer.Enabled {
		inferenceService.Spec.Transformer = t.createTransformerSpec(modelService, modelService.Transformer)
		if err != nil {
			return nil, fmt.Errorf("unable to create transformer spec: %w", err)
		}
		inferenceService.Spec.Transformer.TopologySpreadConstraints, err = t.createNewInferenceServiceTopologySpreadConstraints(
			modelService,
			kservev1beta1.TransformerComponent,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to create transformer topology spread constraints: %w", err)
		}
	}

	return inferenceService, nil
}

func (t *InferenceServiceTemplater) createPredictorSpec(modelService *models.Service) kservev1beta1.PredictorSpec {
	envVars := modelService.EnvVars

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

	nodeSelector := map[string]string{}
	tolerations := []corev1.Toleration{}
	if modelService.ResourceRequest.GPUName != "" && !modelService.ResourceRequest.GPURequest.IsZero() {
		// Look up to the GPU resource type and quantity from DeploymentConfig
		for _, gpuConfig := range t.deploymentConfig.GPUs {
			if gpuConfig.Name == modelService.ResourceRequest.GPUName {
				// Declare and initialize resourceType and resourceQuantity variables
				resourceType := corev1.ResourceName(gpuConfig.ResourceType)
				resourceQuantity := modelService.ResourceRequest.GPURequest

				// Set the resourceType as the key in the maps, with resourceQuantity as the value
				resources.Requests[resourceType] = resourceQuantity
				resources.Limits[resourceType] = resourceQuantity

				nodeSelector = gpuConfig.NodeSelector
				tolerations = gpuConfig.Tolerations
			}
		}
	}

	// liveness probe config. if env var to disable != true or not set, it will default to enabled
	var livenessProbeConfig *corev1.Probe = nil
	envVarsMap := envVars.ToMap()
	if !strings.EqualFold(envVarsMap[envOldDisableLivenessProbe], "true") &&
		!strings.EqualFold(envVarsMap[envDisableLivenessProbe], "true") {
		livenessProbeConfig = createLivenessProbeSpec(modelService.Protocol, fmt.Sprintf("/v1/models/%s", modelService.Name))
	}

	containerPorts := createContainerPorts(modelService.Protocol, modelService.DeploymentMode)
	storageUri := utils.CreateModelLocation(modelService.ArtifactURI)
	var predictorSpec kservev1beta1.PredictorSpec
	switch modelService.Type {
	case models.ModelTypeTensorflow:
		predictorSpec = kservev1beta1.PredictorSpec{
			Tensorflow: &kservev1beta1.TFServingSpec{
				PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
					StorageURI: &storageUri,
					Container: corev1.Container{
						Name:          kserveconstant.InferenceServiceContainerName,
						Resources:     resources,
						LivenessProbe: livenessProbeConfig,
						Ports:         containerPorts,
						Env:           envVars.ToKubernetesEnvVars(),
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
						Name:          kserveconstant.InferenceServiceContainerName,
						Resources:     resources,
						LivenessProbe: livenessProbeConfig,
						Ports:         containerPorts,
						Env:           envVars.ToKubernetesEnvVars(),
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
						Name:          kserveconstant.InferenceServiceContainerName,
						Resources:     resources,
						LivenessProbe: livenessProbeConfig,
						Ports:         containerPorts,
						Env:           envVars.ToKubernetesEnvVars(),
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
						Name:          kserveconstant.InferenceServiceContainerName,
						Resources:     resources,
						LivenessProbe: livenessProbeConfig,
						Ports:         containerPorts,
						Env:           envVars.ToKubernetesEnvVars(),
					},
				},
			},
		}
	case models.ModelTypePyFunc, models.ModelTypePyFuncV3:
		envVars := models.MergeEnvVars(modelService.EnvVars, createPyFuncDefaultEnvVars(modelService))
		if modelService.Protocol == protocol.UpiV1 {
			envVars = append(envVars, models.EnvVar{Name: envGRPCOptions, Value: t.deploymentConfig.PyfuncGRPCOptions})
		}
		if modelService.EnabledModelObservability && modelService.Type == models.ModelTypePyFuncV3 {
			pyfuncPublisherCfg := t.deploymentConfig.PyFuncPublisher
			envVars = append(envVars, models.EnvVar{Name: envPublisherEnabled, Value: strconv.FormatBool(modelService.EnabledModelObservability)})
			envVars = append(envVars, models.EnvVar{Name: envPublisherKafkaTopic, Value: modelService.GetPredictionLogTopicForVersion()})
			envVars = append(envVars, models.EnvVar{Name: envPublisherKafkaBrokers, Value: pyfuncPublisherCfg.Kafka.Brokers})
			envVars = append(envVars, models.EnvVar{Name: envPublisherKafkaLinger, Value: fmt.Sprintf("%d", pyfuncPublisherCfg.Kafka.LingerMS)})
			envVars = append(envVars, models.EnvVar{Name: envPublisherKafkaAck, Value: fmt.Sprintf("%d", pyfuncPublisherCfg.Kafka.Acks)})
			envVars = append(envVars, models.EnvVar{Name: envPublisherSamplingRatio, Value: fmt.Sprintf("%f", pyfuncPublisherCfg.SamplingRatioRate)})
			envVars = append(envVars, models.EnvVar{Name: envPublisherKafkaConfig, Value: pyfuncPublisherCfg.Kafka.AdditionalConfig})
		}
		predictorSpec = kservev1beta1.PredictorSpec{
			PodSpec: kservev1beta1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:          kserveconstant.InferenceServiceContainerName,
						Image:         modelService.Options.PyFuncImageName,
						Env:           envVars.ToKubernetesEnvVars(),
						Resources:     resources,
						LivenessProbe: livenessProbeConfig,
						Ports:         containerPorts,
					},
				},
			},
		}

	case models.ModelTypeCustom:
		predictorSpec = createCustomPredictorSpec(modelService, resources, nodeSelector, tolerations)
	}

	if len(nodeSelector) > 0 {
		predictorSpec.NodeSelector = nodeSelector
	}

	if len(tolerations) > 0 {
		predictorSpec.Tolerations = tolerations
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

func (t *InferenceServiceTemplater) createTransformerSpec(modelService *models.Service, transformer *models.Transformer) *kservev1beta1.TransformerSpec {
	// Set cpu limit and memory limit to be 2x of the requests
	cpuLimit := transformer.ResourceRequest.CPURequest.DeepCopy()
	cpuLimit.Add(transformer.ResourceRequest.CPURequest)
	memoryLimit := transformer.ResourceRequest.MemoryRequest.DeepCopy()
	memoryLimit.Add(transformer.ResourceRequest.MemoryRequest)

	envVars := transformer.EnvVars

	// Put in defaults if not provided by users (user's input is used)
	if transformer.TransformerType == models.StandardTransformerType {
		transformer.Image = t.deploymentConfig.StandardTransformer.ImageName
		envVars = t.enrichStandardTransformerEnvVars(modelService, envVars)
	}

	// Overwrite user's values with defaults, if provided (overwrite by user not allowed)
	defaultEnvVars := createDefaultTransformerEnvVars(modelService)
	envVars = models.MergeEnvVars(envVars, defaultEnvVars)

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

	// liveness probe config. if env var to disable != true or not set, it will default to enabled
	var livenessProbeConfig *corev1.Probe = nil
	envVarsMap := envVars.ToMap()
	if !strings.EqualFold(envVarsMap[envOldDisableLivenessProbe], "true") &&
		!strings.EqualFold(envVarsMap[envDisableLivenessProbe], "true") {
		livenessProbeConfig = createLivenessProbeSpec(modelService.Protocol, "/")
	}

	containerPorts := createContainerPorts(modelService.Protocol, modelService.DeploymentMode)
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
					Command:       transformerCommand,
					Args:          transformerArgs,
					LivenessProbe: livenessProbeConfig,
					Ports:         containerPorts,
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

func (t *InferenceServiceTemplater) enrichStandardTransformerEnvVars(modelService *models.Service, envVars models.EnvVars) models.EnvVars {
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

	standardTransformerCfg := t.deploymentConfig.StandardTransformer

	// Additional env var to add
	addEnvVar := models.EnvVars{}
	addEnvVar = append(addEnvVar, models.EnvVar{Name: transformerpkg.DefaultFeastSource, Value: standardTransformerCfg.DefaultFeastSource.String()})

	// adding feast related config env variable
	feastStorageConfig := standardTransformerCfg.ToFeastStorageConfigs()
	if feastStorageConfigJsonByte, err := json.Marshal(feastStorageConfig); err == nil {
		addEnvVar = append(addEnvVar, models.EnvVar{Name: transformerpkg.FeastStorageConfigs, Value: string(feastStorageConfigJsonByte)})
	}
	// Add keepalive configuration for predictor
	// only pyfunc config that enforced by merlin
	keepAliveModelCfg := standardTransformerCfg.ModelClientKeepAlive
	if modelService.Protocol == protocol.UpiV1 && (modelService.Type == models.ModelTypePyFunc || modelService.Type == models.ModelTypePyFuncV3) {
		addEnvVar = append(addEnvVar, models.EnvVar{Name: transformerpkg.ModelGRPCKeepAliveEnabled, Value: strconv.FormatBool(keepAliveModelCfg.Enabled)})
		addEnvVar = append(addEnvVar, models.EnvVar{Name: transformerpkg.ModelGRPCKeepAliveTime, Value: keepAliveModelCfg.Time.String()})
		addEnvVar = append(addEnvVar, models.EnvVar{Name: transformerpkg.ModelGRPCKeepAliveTimeout, Value: keepAliveModelCfg.Timeout.String()})
	}

	keepaliveCfg := standardTransformerCfg.FeastServingKeepAlive
	addEnvVar = append(addEnvVar, models.EnvVar{Name: transformerpkg.FeastServingKeepAliveEnabled, Value: strconv.FormatBool(keepaliveCfg.Enabled)})
	addEnvVar = append(addEnvVar, models.EnvVar{Name: transformerpkg.FeastServingKeepAliveTime, Value: keepaliveCfg.Time.String()})
	addEnvVar = append(addEnvVar, models.EnvVar{Name: transformerpkg.FeastServingKeepAliveTimeout, Value: keepaliveCfg.Timeout.String()})
	addEnvVar = append(addEnvVar, models.EnvVar{Name: transformerpkg.FeastGRPCConnCount, Value: fmt.Sprintf("%d", standardTransformerCfg.FeastGPRCConnCount)})

	if modelService.Protocol == protocol.UpiV1 {
		// add kafka configuration
		kafkaCfg := standardTransformerCfg.Kafka
		addEnvVar = append(addEnvVar, models.EnvVar{Name: transformerpkg.KafkaTopic, Value: modelService.GetPredictionLogTopic()})
		addEnvVar = append(addEnvVar, models.EnvVar{Name: transformerpkg.KafkaBrokers, Value: kafkaCfg.Brokers})
		addEnvVar = append(addEnvVar, models.EnvVar{Name: transformerpkg.KafkaMaxMessageSizeBytes, Value: fmt.Sprintf("%v", kafkaCfg.MaxMessageSizeBytes)})
		addEnvVar = append(addEnvVar, models.EnvVar{Name: transformerpkg.KafkaConnectTimeoutMS, Value: fmt.Sprintf("%v", kafkaCfg.ConnectTimeoutMS)})
		addEnvVar = append(addEnvVar, models.EnvVar{Name: transformerpkg.KafkaSerialization, Value: string(kafkaCfg.SerializationFmt)})

		addEnvVar = append(addEnvVar, models.EnvVar{Name: transformerpkg.ModelServerConnCount, Value: fmt.Sprintf("%d", standardTransformerCfg.ModelServerConnCount)})
	}

	jaegerCfg := standardTransformerCfg.Jaeger
	jaegerEnvVars := []models.EnvVar{
		{Name: transformerpkg.JaegerCollectorURL, Value: jaegerCfg.CollectorURL},
		{Name: transformerpkg.JaegerSamplerParam, Value: jaegerCfg.SamplerParam},
		{Name: transformerpkg.JaegerDisabled, Value: jaegerCfg.Disabled},
	}

	addEnvVar = append(addEnvVar, jaegerEnvVars...)

	// merge back to envVars (default above will be overriden by users if provided)
	envVars = models.MergeEnvVars(addEnvVar, envVars)

	return envVars
}

func createHTTPGetLivenessProbe(httpPath string, port int) *corev1.Probe {
	return &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path:   httpPath,
				Scheme: "HTTP",
				Port: intstr.IntOrString{
					IntVal: int32(port),
				},
			},
		},
		InitialDelaySeconds: liveProbeInitialDelaySec,
		TimeoutSeconds:      liveProbeTimeoutSec,
		PeriodSeconds:       liveProbePeriodSec,
		SuccessThreshold:    liveProbeSuccessThreshold,
		FailureThreshold:    liveProbeFailureThreshold,
	}
}

func createGRPCLivenessProbe(port int) *corev1.Probe {
	return &corev1.Probe{
		Handler: corev1.Handler{
			Exec: &corev1.ExecAction{
				Command: []string{grpcHealthProbeCommand, fmt.Sprintf("-addr=:%d", port)},
			},
		},
		InitialDelaySeconds: liveProbeInitialDelaySec,
		TimeoutSeconds:      liveProbeTimeoutSec,
		PeriodSeconds:       liveProbePeriodSec,
		SuccessThreshold:    liveProbeSuccessThreshold,
		FailureThreshold:    liveProbeFailureThreshold,
	}
}

func createLivenessProbeSpec(protocol prt.Protocol, httpPath string) *corev1.Probe {
	if protocol == prt.UpiV1 {
		return createGRPCLivenessProbe(defaultGRPCPort)
	}
	return createHTTPGetLivenessProbe(httpPath, defaultHTTPPort)
}

func createPredictorHost(modelService *models.Service) string {
	return fmt.Sprintf("%s-predictor.%s:%d", modelService.Name, modelService.Namespace, defaultPredictorPort)
}

func (t *InferenceServiceTemplater) createAnnotations(modelService *models.Service, initialScale *int) (map[string]string, error) {
	annotations := make(map[string]string)

	config := t.deploymentConfig
	if config.QueueResourcePercentage != "" {
		annotations[knserving.QueueSidecarResourcePercentageAnnotationKey] = config.QueueResourcePercentage
	}

	if modelService.Type == models.ModelTypePyFunc || modelService.Type == models.ModelTypePyFuncV3 {
		annotations[annotationPrometheusScrapeFlag] = "true"
		annotations[annotationPrometheusScrapePort] = fmt.Sprint(defaultHTTPPort)
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

			autoscalingMetrics, targetValue, err := toKNativeAutoscalerMetrics(modelService.AutoscalingPolicy.MetricsType, modelService.AutoscalingPolicy.TargetValue, modelService.ResourceRequest)
			if err != nil {
				return nil, err
			}

			annotations[knautoscaling.MetricAnnotationKey] = autoscalingMetrics
			annotations[knautoscaling.TargetAnnotationKey] = targetValue
		}
	}

	// For serverless deployments, set initial scale, if supplied.
	if deployMode == string(kserveconstant.Serverless) && initialScale != nil {
		annotations[knautoscaling.InitialScaleAnnotationKey] = strconv.Itoa(*initialScale)
	}

	return annotations, nil
}

func createContainerPorts(protocolValue protocol.Protocol, deploymentMode deployment.Mode) []corev1.ContainerPort {
	var containerPorts []corev1.ContainerPort = nil
	switch protocolValue {
	case protocol.UpiV1:
		if deploymentMode == deployment.RawDeploymentMode {
			containerPorts = grpcHPAContainerPorts
		} else {
			containerPorts = grpcKPAContainerPorts
		}
	default:
	}
	return containerPorts
}

func createLoggerSpec(loggerURL string, loggerConfig models.LoggerConfig) *kservev1beta1.LoggerSpec {
	loggerMode := models.ToKFServingLoggerMode(loggerConfig.Mode)
	return &kservev1beta1.LoggerSpec{
		URL:  &loggerURL,
		Mode: loggerMode,
	}
}

// createNewInferenceServiceTopologySpreadConstraints creates topology spread constrains for a component of a new
// inference service
func (t *InferenceServiceTemplater) createNewInferenceServiceTopologySpreadConstraints(
	modelService *models.Service,
	component kservev1beta1.ComponentType,
) ([]corev1.TopologySpreadConstraint, error) {
	config := t.deploymentConfig
	if len(config.TopologySpreadConstraints) == 0 {
		var topologySpreadConstraints []corev1.TopologySpreadConstraint
		return topologySpreadConstraints, nil
	}
	var newRevisionName string
	if modelService.DeploymentMode == deployment.RawDeploymentMode {
		newRevisionName = fmt.Sprintf("isvc.%s-%s", modelService.Name, component)
	} else if modelService.DeploymentMode == deployment.ServerlessDeploymentMode ||
		modelService.DeploymentMode == deployment.EmptyDeploymentMode {
		newRevisionName = fmt.Sprintf("%s-%s-00001", modelService.Name, component)
	} else {
		return nil, fmt.Errorf("invalid deployment mode: %s", modelService.DeploymentMode)
	}
	return appendPodSpreadingLabelSelectorsToTopologySpreadConstraints(
		config.TopologySpreadConstraints,
		newRevisionName,
	)
}

// appendPodSpreadingLabelSelectorsToTopologySpreadConstraints makes a deep copy of the config topology spread
// constraints and then adds the given revisionName as a label to the match labels of each topology spread constraint
// to spread out all the pods across the specified topologyKey
func appendPodSpreadingLabelSelectorsToTopologySpreadConstraints(
	templateTopologySpreadConstraints []corev1.TopologySpreadConstraint,
	revisionName string,
) ([]corev1.TopologySpreadConstraint, error) {
	topologySpreadConstraints, err := copyTopologySpreadConstraints(templateTopologySpreadConstraints)
	if err != nil {
		return nil, err
	}
	for i := range topologySpreadConstraints {
		if topologySpreadConstraints[i].LabelSelector == nil {
			topologySpreadConstraints[i].LabelSelector = &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": revisionName},
			}
		} else {
			if topologySpreadConstraints[i].LabelSelector.MatchLabels == nil {
				topologySpreadConstraints[i].LabelSelector.MatchLabels = make(map[string]string)
			}
			topologySpreadConstraints[i].LabelSelector.MatchLabels["app"] = revisionName
		}
	}
	return topologySpreadConstraints, nil
}

// copyTopologySpreadConstraints copies the topology spread constraints using the service builder's as a template
func copyTopologySpreadConstraints(
	topologySpreadConstraints []corev1.TopologySpreadConstraint,
) ([]corev1.TopologySpreadConstraint, error) {
	topologySpreadConstraintsRaw, err := copystructure.Copy(topologySpreadConstraints)
	if err != nil {
		return nil, fmt.Errorf("Error copying topology spread constraints: %w", err)
	}
	topologySpreadConstraints, ok := topologySpreadConstraintsRaw.([]corev1.TopologySpreadConstraint)
	if !ok {
		return nil, fmt.Errorf("Error in type assertion of copied topology spread constraints interface: %w", err)
	}
	return topologySpreadConstraints, nil
}

func createDefaultTransformerEnvVars(modelService *models.Service) models.EnvVars {
	defaultEnvVars := models.EnvVars{}

	// These env vars are to be deprecated
	defaultEnvVars = append(defaultEnvVars, models.EnvVar{Name: envTransformerPort, Value: fmt.Sprint(defaultHTTPPort)})
	defaultEnvVars = append(defaultEnvVars, models.EnvVar{Name: envTransformerModelName, Value: modelService.Name})
	defaultEnvVars = append(defaultEnvVars, models.EnvVar{Name: envTransformerPredictURL, Value: createPredictorHost(modelService)})

	defaultEnvVars = append(defaultEnvVars, models.EnvVar{Name: envHTTPPort, Value: fmt.Sprint(defaultHTTPPort)})
	defaultEnvVars = append(defaultEnvVars, models.EnvVar{Name: envGRPCPort, Value: fmt.Sprint(defaultGRPCPort)})
	defaultEnvVars = append(defaultEnvVars, models.EnvVar{Name: envModelName, Value: modelService.ModelName})
	defaultEnvVars = append(defaultEnvVars, models.EnvVar{Name: envModelVersion, Value: modelService.ModelVersion})
	defaultEnvVars = append(defaultEnvVars, models.EnvVar{Name: envModelFullName, Value: modelService.Name})
	defaultEnvVars = append(defaultEnvVars, models.EnvVar{Name: envProject, Value: modelService.Namespace})
	defaultEnvVars = append(defaultEnvVars, models.EnvVar{Name: envPredictorHost, Value: createPredictorHost(modelService)})
	defaultEnvVars = append(defaultEnvVars, models.EnvVar{Name: envProtocol, Value: fmt.Sprint(modelService.Protocol)})
	return defaultEnvVars
}

func createDefaultPredictorEnvVars(modelService *models.Service) models.EnvVars {
	defaultEnvVars := models.EnvVars{}

	// These env vars are to be deprecated
	defaultEnvVars = append(defaultEnvVars, models.EnvVar{Name: envPredictorPort, Value: fmt.Sprint(defaultHTTPPort)})
	defaultEnvVars = append(defaultEnvVars, models.EnvVar{Name: envPredictorModelName, Value: modelService.Name})
	defaultEnvVars = append(defaultEnvVars, models.EnvVar{Name: envPredictorArtifactLocation, Value: defaultPredictorArtifactLocation})

	defaultEnvVars = append(defaultEnvVars, models.EnvVar{Name: envStorageURI, Value: utils.CreateModelLocation(modelService.ArtifactURI)})
	defaultEnvVars = append(defaultEnvVars, models.EnvVar{Name: envHTTPPort, Value: fmt.Sprint(defaultHTTPPort)})
	defaultEnvVars = append(defaultEnvVars, models.EnvVar{Name: envGRPCPort, Value: fmt.Sprint(defaultGRPCPort)})
	defaultEnvVars = append(defaultEnvVars, models.EnvVar{Name: envModelName, Value: modelService.ModelName})
	defaultEnvVars = append(defaultEnvVars, models.EnvVar{Name: envModelVersion, Value: modelService.ModelVersion})
	defaultEnvVars = append(defaultEnvVars, models.EnvVar{Name: envModelFullName, Value: modelService.Name})
	defaultEnvVars = append(defaultEnvVars, models.EnvVar{Name: envArtifactLocation, Value: defaultPredictorArtifactLocation})
	return defaultEnvVars
}

func createCustomPredictorSpec(modelService *models.Service, resources corev1.ResourceRequirements, nodeSelector map[string]string, tolerations []corev1.Toleration) kservev1beta1.PredictorSpec {
	envVars := modelService.EnvVars

	// Add default env var (Overwrite by user not allowed)
	defaultEnvVar := createDefaultPredictorEnvVars(modelService)
	envVars = models.MergeEnvVars(envVars, defaultEnvVar)

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

	containerPorts := createContainerPorts(modelService.Protocol, modelService.DeploymentMode)
	spec := kservev1beta1.PredictorSpec{
		PodSpec: kservev1beta1.PodSpec{
			Containers: []corev1.Container{
				{
					Image:     customPredictor.Image,
					Env:       envVars.ToKubernetesEnvVars(),
					Resources: resources,
					Command:   containerCommand,
					Args:      containerArgs,
					Name:      kserveconstant.InferenceServiceContainerName,
					Ports:     containerPorts,
				},
			},
		},
	}
	if len(nodeSelector) > 0 {
		spec.NodeSelector = nodeSelector
	}

	if len(tolerations) > 0 {
		spec.Tolerations = tolerations
	}

	return spec
}

// createPyFuncDefaultEnvVars return default env vars for Pyfunc model.
func createPyFuncDefaultEnvVars(svc *models.Service) models.EnvVars {
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
			Value: fmt.Sprint(svc.Protocol),
		},
		models.EnvVar{
			Name:  envProject,
			Value: svc.Namespace,
		},
	}
	return envVars
}

func (t *InferenceServiceTemplater) applyDefaults(service *models.Service) {
	// apply default resource request for model
	config := t.deploymentConfig
	if service.ResourceRequest == nil {
		service.ResourceRequest = &models.ResourceRequest{
			MinReplica:    config.DefaultModelResourceRequests.MinReplica,
			MaxReplica:    config.DefaultModelResourceRequests.MaxReplica,
			CPURequest:    config.DefaultModelResourceRequests.CPURequest,
			MemoryRequest: config.DefaultModelResourceRequests.MemoryRequest,
		}
	}

	// apply default resource request for transformer
	if service.Transformer != nil && service.Transformer.Enabled {
		if service.Transformer.ResourceRequest == nil {
			service.Transformer.ResourceRequest = &models.ResourceRequest{
				MinReplica:    config.DefaultTransformerResourceRequests.MinReplica,
				MaxReplica:    config.DefaultTransformerResourceRequests.MaxReplica,
				CPURequest:    config.DefaultTransformerResourceRequests.CPURequest,
				MemoryRequest: config.DefaultTransformerResourceRequests.MemoryRequest,
			}
		}
	}
}
