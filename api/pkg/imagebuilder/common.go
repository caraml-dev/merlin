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

package imagebuilder

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// RequestLimitResources is a Kubernetes resource request and limits
type RequestLimitResources struct {
	Request Resource
	Limit   Resource
}

// Build converts the spec into a Kubernetes spec
func (r *RequestLimitResources) Build() corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Requests: r.Request.Build(),
		Limits:   r.Limit.Build(),
	}
}

// Resource is a Kubernetes resource
type Resource struct {
	CPU    resource.Quantity
	Memory resource.Quantity
}

// Build converts the spec into a Kubernetes spec
func (r *Resource) Build() corev1.ResourceList {
	return corev1.ResourceList{
		corev1.ResourceCPU:    r.CPU,
		corev1.ResourceMemory: r.Memory,
	}
}
