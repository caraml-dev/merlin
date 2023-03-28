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
	"log"
	"os"
	"time"

	mlpcluster "github.com/gojek/mlp/api/pkg/cluster"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/api/resource"
)

type EnvironmentConfig struct {
	Name                   string `yaml:"name"`
	Cluster                string `yaml:"cluster"`
	IsDefault              bool   `yaml:"is_default"`
	IsPredictionJobEnabled bool   `yaml:"is_prediction_job_enabled"`
	IsDefaultPredictionJob bool   `yaml:"is_default_prediction_job"`

	Region            string        `yaml:"region"`
	GcpProject        string        `yaml:"gcp_project"`
	DeploymentTimeout time.Duration `yaml:"deployment_timeout"`
	NamespaceTimeout  time.Duration `yaml:"namespace_timeout"`

	MaxCPU    string `yaml:"max_cpu"`
	MaxMemory string `yaml:"max_memory"`

	QueueResourcePercentage string `yaml:"queue_resource_percentage"`

	DefaultPredictionJobConfig *PredictionJobResourceRequestConfig `yaml:"default_prediction_job_config"`
	DefaultDeploymentConfig    *ResourceRequestConfig              `yaml:"default_deployment_config"`
	DefaultTransformerConfig   *ResourceRequestConfig              `yaml:"default_transformer_config"`
	K8sConfig                  *mlpcluster.K8sConfig               `yaml:"k8s_config"`
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

func initEnvironmentConfigs(path string) []EnvironmentConfig {
	cfgFile, err := os.ReadFile(path)
	if err != nil {
		log.Panicf("unable to read deployment config file: %s", path)
	}
	var configs []EnvironmentConfig
	err = yaml.Unmarshal(cfgFile, &configs)
	if err != nil {
		log.Panicf("unable to unmarshall deployment config file:\n %s,\nDue to: %v", cfgFile, err)
	}
	return configs
}

func ParseDeploymentConfig(cfg EnvironmentConfig, pyfuncGRPCOptions string) DeploymentConfig {
	return DeploymentConfig{
		DeploymentTimeout: cfg.DeploymentTimeout,
		NamespaceTimeout:  cfg.NamespaceTimeout,
		DefaultModelResourceRequests: &ResourceRequests{
			MinReplica:    cfg.DefaultDeploymentConfig.MinReplica,
			MaxReplica:    cfg.DefaultDeploymentConfig.MaxReplica,
			CPURequest:    resource.MustParse(cfg.DefaultDeploymentConfig.CPURequest),
			MemoryRequest: resource.MustParse(cfg.DefaultDeploymentConfig.MemoryRequest),
		},
		DefaultTransformerResourceRequests: &ResourceRequests{
			MinReplica:    cfg.DefaultTransformerConfig.MinReplica,
			MaxReplica:    cfg.DefaultTransformerConfig.MaxReplica,
			CPURequest:    resource.MustParse(cfg.DefaultTransformerConfig.CPURequest),
			MemoryRequest: resource.MustParse(cfg.DefaultTransformerConfig.MemoryRequest),
		},
		MaxCPU:                  resource.MustParse(cfg.MaxCPU),
		MaxMemory:               resource.MustParse(cfg.MaxMemory),
		QueueResourcePercentage: cfg.QueueResourcePercentage,
		PyfuncGRPCOptions:       pyfuncGRPCOptions,
	}
}
