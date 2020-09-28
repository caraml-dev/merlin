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

	"k8s.io/apimachinery/pkg/api/resource"
)

// DeploymentConfig specify the configuration of inference service deployment
type DeploymentConfig struct {
	// Duration to wait for inference service creation
	DeploymentTimeout time.Duration
	// Duration to wait for namespaceResource creation
	NamespaceTimeout time.Duration
	// Minimum number of replica of inference service
	MinReplica int
	// Maximum number of replica of inference service
	MaxReplica int

	// CPU limit of inference service
	CpuLimit resource.Quantity //deprecated
	// Max CPU of machine
	MaxCpu resource.Quantity
	// Max Memory of machine
	MaxMemory resource.Quantity
	// Memory limit of inference service
	MemoryLimit resource.Quantity
	// CPU request of inference service
	CpuRequest resource.Quantity
	// Memory request of inference service
	MemoryRequest resource.Quantity

	// Percentage of knative's queue proxy resource request from the inference service resource request
	QueueResourcePercentage string
}
