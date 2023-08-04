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

package config

import (
	"fmt"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	sigyaml "sigs.k8s.io/yaml"

	mlpcluster "github.com/caraml-dev/mlp/api/pkg/cluster"
)

type EnvironmentConfig struct {
	Name                   string `validate:"required" yaml:"name"`
	Cluster                string `yaml:"cluster"`
	IsDefault              bool   `yaml:"is_default"`
	IsPredictionJobEnabled bool   `yaml:"is_prediction_job_enabled"`
	IsDefaultPredictionJob bool   `yaml:"is_default_prediction_job"`

	Region            string        `yaml:"region"`
	GcpProject        string        `yaml:"gcp_project"`
	DeploymentTimeout time.Duration `yaml:"deployment_timeout"`
	NamespaceTimeout  time.Duration `yaml:"namespace_timeout"`

	Gpus []GpuConfig `yaml:"gpus"`

	MaxCPU                    string                    `yaml:"max_cpu"`
	MaxMemory                 string                    `yaml:"max_memory"`
	TopologySpreadConstraints TopologySpreadConstraints `yaml:"topology_spread_constraints"`
	PodDisruptionBudget       PodDisruptionBudgetConfig `yaml:"pod_disruption_budget"`

	QueueResourcePercentage string `yaml:"queue_resource_percentage"`

	DefaultPredictionJobConfig *PredictionJobResourceRequestConfig `yaml:"default_prediction_job_config"`
	DefaultDeploymentConfig    *ResourceRequestConfig              `yaml:"default_deployment_config"`
	DefaultTransformerConfig   *ResourceRequestConfig              `yaml:"default_transformer_config"`
	K8sConfig                  *mlpcluster.K8sConfig               `validate:"required" yaml:"k8s_config"`
}

func (e *EnvironmentConfig) Validate() error {
	v := validator.New()

	// Use struct level validation for PodDisruptionBudgetConfig
	v.RegisterStructValidation(func(sl validator.StructLevel) {
		field := sl.Current().Interface().(PodDisruptionBudgetConfig)
		// If PDB is enabled, one of max unavailable or min available shall be set
		if field.Enabled &&
			(field.MaxUnavailablePercentage == nil && field.MinAvailablePercentage == nil) ||
			(field.MaxUnavailablePercentage != nil && field.MinAvailablePercentage != nil) {
			sl.ReportError(field.MaxUnavailablePercentage, "max_unavailable_percentage", "int", "choose_one[max_unavailable_percentage,min_available_percentage]", "")
			sl.ReportError(field.MinAvailablePercentage, "min_available_percentage", "int", "choose_one[max_unavailable_percentage,min_available_percentage]", "")
		}
	}, PodDisruptionBudgetConfig{})

	return v.Struct(e)
}

type TopologySpreadConstraints []corev1.TopologySpreadConstraint

// UnmarshalYAML implements Unmarshal interface
// Since TopologySpreadConstraint fields only have json tags, sigyaml.Unmarshal needs to be used
// to unmarshal all the fields. This method reads TopologySpreadConstraint into a map[string]interface{},
// marshals it into a byte for, before passing to sigyaml.Unmarshal
func (t *TopologySpreadConstraints) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var topologySpreadConstraints []map[string]interface{}
	// Unmarshal into map[string]interface{}
	if err := unmarshal(&topologySpreadConstraints); err != nil {
		return err
	}
	// convert back to byte string
	byteForm, err := yaml.Marshal(topologySpreadConstraints)
	if err != nil {
		return err
	}
	// use sigyaml.Unmarshal to convert to json object then unmarshal
	if err := sigyaml.Unmarshal(byteForm, t); err != nil {
		return err
	}
	return nil
}

type PredictionJobResourceRequestConfig struct {
	ExecutorReplica       int32  `yaml:"executor_replica"`
	DriverCPURequest      string `yaml:"driver_cpu_request"`
	DriverMemoryRequest   string `yaml:"driver_memory_request"`
	ExecutorCPURequest    string `yaml:"executor_cpu_request"`
	ExecutorMemoryRequest string `yaml:"executor_memory_request"`
}

type ResourceRequestConfig struct {
	MinReplica    int    `yaml:"min_replica"`
	MaxReplica    int    `yaml:"max_replica"`
	CPURequest    string `yaml:"cpu_request"`
	MemoryRequest string `yaml:"memory_request"`
}

type GpuConfig struct {
	Values            []string          `yaml:"values"`
	DisplayName       string            `yaml:"display_name"`
	ResourceType      string            `yaml:"resource_type"`
	NodeSelector      map[string]string `yaml:"node_selector"`
	MonthlyCostPerGpu float64           `yaml:"monthly_cost_per_gpu"`
}

func InitEnvironmentConfigs(path string) ([]*EnvironmentConfig, error) {
	cfgFile, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read deployment config file: %s", path)
	}

	var configs []*EnvironmentConfig
	if err = yaml.Unmarshal(cfgFile, &configs); err != nil {
		return nil, fmt.Errorf("unable to unmarshall deployment config file:\n %s,\ndue to: %w", cfgFile, err)
	}

	for _, env := range configs {
		if err := env.Validate(); err != nil {
			return nil, fmt.Errorf("invalid environment config: %w", err)
		}

		if env.K8sConfig == nil {
			return nil, fmt.Errorf("k8sConfig for %s is nil", env.Name)
		}
	}

	return configs, nil
}

func ParseDeploymentConfig(cfg *EnvironmentConfig, pyfuncGRPCOptions string) DeploymentConfig {
	return DeploymentConfig{
		DeploymentTimeout: cfg.DeploymentTimeout,
		NamespaceTimeout:  cfg.NamespaceTimeout,
		DefaultModelResourceRequests: &ResourceRequests{
			MinReplica:    cfg.DefaultDeploymentConfig.MinReplica,
			MaxReplica:    cfg.DefaultDeploymentConfig.MaxReplica,
			CPURequest:    resource.MustParse(cfg.DefaultDeploymentConfig.CPURequest),
			MemoryRequest: resource.MustParse(cfg.DefaultDeploymentConfig.MemoryRequest),
			GPURequest:    resource.MustParse(cfg.DefaultDeploymentConfig.Gpus),
		},
		DefaultTransformerResourceRequests: &ResourceRequests{
			MinReplica:    cfg.DefaultTransformerConfig.MinReplica,
			MaxReplica:    cfg.DefaultTransformerConfig.MaxReplica,
			CPURequest:    resource.MustParse(cfg.DefaultTransformerConfig.CPURequest),
			MemoryRequest: resource.MustParse(cfg.DefaultTransformerConfig.MemoryRequest),
			GPURequest:    resource.MustParse(cfg.DefaultDeploymentConfig.Gpus),
		},
		MaxCPU:                    resource.MustParse(cfg.MaxCPU),
		MaxMemory:                 resource.MustParse(cfg.MaxMemory),
		TopologySpreadConstraints: cfg.TopologySpreadConstraints,
		QueueResourcePercentage:   cfg.QueueResourcePercentage,
		PyfuncGRPCOptions:         pyfuncGRPCOptions,
		PodDisruptionBudget:       cfg.PodDisruptionBudget,
	}
}
