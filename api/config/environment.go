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
	"io/ioutil"
	"log"
	"time"

	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/api/resource"
)

type EnvironmentConfig struct {
	Name                    string        `yaml:"name"`
	Cluster                 string        `yaml:"cluster"`
	IsDefault               bool          `yaml:"is_default"`
	Region                  string        `yaml:"region"`
	GcpProject              string        `yaml:"gcp_project"`
	DeploymentTimeout       time.Duration `yaml:"deployment_timeout"`
	NamespaceTimeout        time.Duration `yaml:"namespace_timeout"`
	MinReplica              int           `yaml:"min_replica"`
	MaxReplica              int           `yaml:"max_replica"`
	CpuRequest              string        `yaml:"cpu_request"`
	MaxCpu                  string        `yaml:"max_cpu"`
	MaxMemory               string        `yaml:"max_memory"`
	MemoryRequest           string        `yaml:"memory_request"`
	CpuLimit                string        `yaml:"cpu_limit"`
	MemoryLimit             string        `yaml:"memory_limit"`
	QueueResourcePercentage string        `yaml:"queue_resource_percentage"`

	IsPredictionJobEnabled bool                 `yaml:"is_prediction_job_enabled"`
	IsDefaultPredictionJob bool                 `yaml:"is_default_prediction_job"`
	PredictionJobConfig    *PredictionJobConfig `yaml:"prediction_job_config"`
}

type PredictionJobConfig struct {
	ExecutorReplica       int32  `yaml:"executor_replica"`
	DriverCpuRequest      string `yaml:"driver_cpu_request"`
	DriverMemoryRequest   string `yaml:"driver_memory_request"`
	ExecutorCpuRequest    string `yaml:"executor_cpu_request"`
	ExecutorMemoryRequest string `yaml:"executor_memory_request"`
}

func initEnvironmentConfigs(path string) []EnvironmentConfig {
	cfgFile, err := ioutil.ReadFile(path)
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

func ParseDeploymentConfig(cfg EnvironmentConfig) DeploymentConfig {
	return DeploymentConfig{
		DeploymentTimeout:       cfg.DeploymentTimeout,
		NamespaceTimeout:        cfg.NamespaceTimeout,
		MinReplica:              cfg.MinReplica,
		MaxReplica:              cfg.MaxReplica,
		CpuRequest:              resource.MustParse(cfg.CpuRequest),
		MemoryRequest:           resource.MustParse(cfg.MemoryRequest),
		CpuLimit:                resource.MustParse(cfg.CpuLimit), //Deprecated
		MaxCpu:                  resource.MustParse(cfg.MaxCpu),
		MaxMemory:               resource.MustParse(cfg.MaxMemory),
		MemoryLimit:             resource.MustParse(cfg.MemoryLimit),
		QueueResourcePercentage: cfg.QueueResourcePercentage,
	}
}
