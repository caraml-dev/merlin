package autoscaling

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/gojek/merlin/pkg/deployment"
	merror "github.com/gojek/merlin/pkg/errors"
)

// MetricsType are supported metrics for autoscaling
type MetricsType string

const (
	// CPUUtilization autoscaling based on cpu utilization (0-100 percent)
	CPUUtilization MetricsType = "cpu_utilization"
	// MemoryUtilization autoscaling based on memory utilization (0-100 percent)
	MemoryUtilization MetricsType = "memory_utilization"
	// Concurrency autoscaling based on number of concurrent request that the model should handle
	Concurrency MetricsType = "concurrency"
	// RPS autoscaling based on throughput (request per seconds)
	RPS MetricsType = "rps"
)

var (
	DefaultRawDeploymentAutoscalingTarget = &AutoscalingTarget{
		MetricsType: CPUUtilization,
		TargetValue: 50,
	}

	DefaultServerlessAutoscalingTarget = &AutoscalingTarget{
		MetricsType: Concurrency,
		TargetValue: 1,
	}
)

// AutoscalingTarget specify the metrics type and target value for autoscaling
type AutoscalingTarget struct {
	// MetricsType type of metrics to be used
	MetricsType MetricsType `json:"metrics_type"`
	// TargetValue specifies the target value
	TargetValue float64 `json:"target_value"`
}

func (r AutoscalingTarget) Value() (driver.Value, error) {
	return json.Marshal(r)
}

func (r *AutoscalingTarget) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &r)
}

// ValidateAutoscalingTarget check autoscaling target is valid and supported by the given deployment mode
func ValidateAutoscalingTarget(target *AutoscalingTarget, mode deployment.Mode) error {
	// raw deployment only support cpu utilization
	if mode == deployment.RawDeploymentMode && target.MetricsType != CPUUtilization {
		return merror.NewInvalidInputErrorf("raw_deployment doesn't support %v autoscaling metrics", target)
	}

	// boundary check for cpu and memory utilization
	if (target.MetricsType == CPUUtilization || target.MetricsType == MemoryUtilization) &&
		(target.TargetValue <= 0 || target.TargetValue > 100) {
		return merror.NewInvalidInputErrorf("target %v is outside 0-100 range", target.MetricsType)
	}

	// boundary check for rps and concurrency
	if (target.MetricsType == Concurrency || target.MetricsType == RPS) &&
		target.TargetValue <= 0 {
		return merror.NewInvalidInputErrorf("target %v is less than or equal to 0", target.MetricsType)
	}

	return nil
}
