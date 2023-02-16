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
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var (
	cpu    = resource.MustParse("500m")
	memory = resource.MustParse("500Mi")
)

func CreateRequestLimitResources() RequestLimitResources {
	return RequestLimitResources{
		Request: Resource{
			CPU:    cpu,
			Memory: memory,
		},
		Limit: Resource{
			CPU:    cpu,
			Memory: memory,
		},
	}
}

func TestContainer(t *testing.T) {
	expected := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    cpu,
			corev1.ResourceMemory: memory,
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    cpu,
			corev1.ResourceMemory: memory,
		},
	}
	c := CreateRequestLimitResources()

	assert.Equal(t, expected, c.Build())
}
