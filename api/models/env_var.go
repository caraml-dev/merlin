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

	v1 "k8s.io/api/core/v1"
)

// EnvVar represents an environment variable present in a container.
type EnvVar struct {
	// Name of the environment variable.
	Name string `json:"name"`

	// Value of the environment variable.
	// Defaults to "".
	Value string `json:"value"`
}

// EnvVars is a list of environment variables to set in the container.
type EnvVars []EnvVar

// ToMap convert EnvVars into map of strings
func (evs EnvVars) ToMap() map[string]string {
	maps := make(map[string]string)
	for _, envVar := range evs {
		maps[envVar.Name] = envVar.Value
	}
	return maps
}

func (evs EnvVars) Value() (driver.Value, error) {
	return json.Marshal(evs)
}

func (evs *EnvVars) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &evs)
}

// ToKubernetesEnvVars returns the representation of Kubernetes'
// v1.EnvVars.
func (evs EnvVars) ToKubernetesEnvVars() []v1.EnvVar {
	kubeEnvVars := make([]v1.EnvVar, len(evs))

	for k, ev := range evs {
		kubeEnvVars[k] = v1.EnvVar{Name: ev.Name, Value: ev.Value}
	}

	return kubeEnvVars
}

// MergeEnvVars merges multiple sets of environment variables and return the merging result.
// All the EnvVars passed as arguments will be not mutated.
// EnvVars to the right have higher precedence.
func MergeEnvVars(left EnvVars, rightEnvVars ...EnvVars) EnvVars {
	for _, right := range rightEnvVars {
		envIndexMap := make(map[string]int, len(left)+len(right))
		for index, ev := range left {
			envIndexMap[ev.Name] = index
		}
		for _, add := range right {
			if index, exist := envIndexMap[add.Name]; exist {
				left[index].Value = add.Value
			} else {
				left = append(left, add)
			}
		}
	}
	return left
}
