package resource

import (
	"fmt"

	"github.com/gojek/merlin/pkg/autoscaling"
	"github.com/gojek/merlin/pkg/deployment"
	kserveconstant "github.com/kserve/kserve/pkg/constants"
	knautoscaling "knative.dev/serving/pkg/apis/autoscaling"
)

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
