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
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// DeploymentConfig specify the configuration of inference service deployment
type DeploymentConfig struct {
	// Duration to wait for inference service creation
	DeploymentTimeout time.Duration
	// Duration to wait for namespaceResource creation
	NamespaceTimeout time.Duration
	// Default resource request for model deployment
	DefaultModelResourceRequests *ResourceRequests
	// Default resource request for transformer deployment
	DefaultTransformerResourceRequests *ResourceRequests
	// Max CPU of machine
	MaxCPU resource.Quantity
	// Max Memory of machine
	MaxMemory resource.Quantity
	// TopologySpreadConstraints to be applied on the pods of each model deployment
	TopologySpreadConstraints []corev1.TopologySpreadConstraint
	// Percentage of knative's queue proxy resource request from the inference service resource request
	QueueResourcePercentage string
	// GRPC Options for Pyfunc server
	PyfuncGRPCOptions string
	// PDB config to be applied on models and transformers
	PodDisruptionBudget PodDisruptionBudgetConfig
	// GPU Config
	Gpus string
}

type ResourceRequests struct {
	// Minimum number of replica of inference service
	MinReplica int
	// Maximum number of replica of inference service
	MaxReplica int
	// CPU request of inference service
	CPURequest resource.Quantity
	// Memory request of inference service
	MemoryRequest resource.Quantity
}

// PodDisruptionBudgetConfig are the configuration for PodDisruptionBudgetConfig for
// Turing services.
type PodDisruptionBudgetConfig struct {
	Enabled bool `yaml:"enabled"`
	// Can specify only one of maxUnavailable and minAvailable
	MaxUnavailablePercentage *int `yaml:"max_unavailable_percentage"`
	MinAvailablePercentage   *int `yaml:"min_available_percentage"`
}
