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
	// Memory request of inference service
	MemoryRequest resource.Quantity `json:"memory_request"`
	// GPU name
	GPUName string `json:"gpu_name,omitempty"`
	// GPU Quantity requests
	GPURequest resource.Quantity `json:"gpu_request,omitempty"`
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
