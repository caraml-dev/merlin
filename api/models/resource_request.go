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

package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"k8s.io/apimachinery/pkg/api/resource"
)

type ResourceRequest struct {
	// Minimum number of replica of inference service
	MinReplica int `json:"min_replica"`
	// Maximum number of replica of inference service
	MaxReplica int `json:"max_replica"`
	// CPU request of inference service
	CPURequest resource.Quantity `json:"cpu_request"`
	// CPU limit of inference service
	CPULimit *resource.Quantity `json:"cpu_limit,omitempty"`
	// Memory request of inference service
	MemoryRequest resource.Quantity `json:"memory_request"`
	// GPU name
	GPUName string `json:"gpu_name,omitempty"`
	// GPU Quantity requests
	GPURequest resource.Quantity `json:"gpu_request,omitempty"`
	// Liveness probe configuration
	LivenessProbe *ProbeConfig `json:"liveness_probe,omitempty"`
	// Readiness probe configuration
	ReadinessProbe *ProbeConfig `json:"readiness_probe,omitempty"`
	// Startup probe configuration
	StartupProbe *ProbeConfig `json:"startup_probe,omitempty"`
}

// ProbeConfig represents the configuration for Kubernetes liveness/readiness probes
type ProbeConfig struct {
	// Path for HTTP probe (for HTTP-based probes)
	Path string `json:"path,omitempty"`
	// Port for the probe
	Port int32 `json:"port,omitempty"`
	// Scheme for HTTP probe (HTTP or HTTPS)
	Scheme string `json:"scheme,omitempty"`
	// Initial delay before starting the probe (seconds)
	InitialDelaySeconds int32 `json:"initial_delay_seconds,omitempty"`
	// Timeout for the probe (seconds)
	TimeoutSeconds int32 `json:"timeout_seconds,omitempty"`
	// Period between probe checks (seconds)
	PeriodSeconds int32 `json:"period_seconds,omitempty"`
	// Number of successes required to be considered healthy
	SuccessThreshold int32 `json:"success_threshold,omitempty"`
	// Number of failures before considered unhealthy
	FailureThreshold int32 `json:"failure_threshold,omitempty"`
}

func (r ResourceRequest) Value() (driver.Value, error) {
	return json.Marshal(r)
}

func (r *ResourceRequest) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &r)
}
