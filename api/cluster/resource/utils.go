package resource

import (
	"fmt"
	"math"

	"github.com/gojek/merlin/models"
	"github.com/gojek/merlin/pkg/autoscaling"
	"github.com/gojek/merlin/pkg/deployment"
	kserveconstant "github.com/kserve/kserve/pkg/constants"
	"k8s.io/apimachinery/pkg/api/resource"
	knautoscaling "knative.dev/serving/pkg/apis/autoscaling"
)

func toKNativeAutoscalerMetrics(metricsType autoscaling.MetricsType, metricsValue float64, resourceReq *models.ResourceRequest) (string, string, error) {
	switch metricsType {
	case autoscaling.CPUUtilization:
		targetValue := fmt.Sprintf("%.0f", metricsValue)
		return knautoscaling.CPU, targetValue, nil
	case autoscaling.MemoryUtilization:
		// Calculate memory target value based on memory request * metricsValue
		// ref: https://github.com/krithika369/turing/blob/6cc3b5f673ccf66e9c1351b2d91382315e86f8ce/api/turing/cluster/knative_service.go#L166
		memoryTarget := computeResource(resourceReq.MemoryRequest, metricsValue/100)
		return knautoscaling.Memory, fmt.Sprintf("%.0f", float64(memoryTarget.Value())/math.Pow(1024, 2)), nil
	case autoscaling.RPS:
		targetValue := fmt.Sprintf("%.0f", metricsValue)
		return knautoscaling.RPS, targetValue, nil
	case autoscaling.Concurrency:
		targetValue := fmt.Sprintf("%.0f", metricsValue)
		return knautoscaling.Concurrency, targetValue, nil
	default:
		return "", "", fmt.Errorf("unsuppported autoscaler metrics on serverless deployment: %s", metricsType)
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

func computeResource(resourceQuantity resource.Quantity, fraction float64) resource.Quantity {
	scaledValue := resourceQuantity.Value()

	scaledMilliValue := int64(math.MaxInt64 - 1)
	if scaledValue < (math.MaxInt64 / 1000) {
		scaledMilliValue = resourceQuantity.MilliValue()
	}
	percentageValue := float64(scaledMilliValue) * fraction
	newValue := int64(math.MaxInt64)
	if percentageValue < math.MaxInt64 {
		newValue = int64(percentageValue)
	}
	newquantity := resource.NewMilliQuantity(newValue, resource.BinarySI)
	return *newquantity
}
