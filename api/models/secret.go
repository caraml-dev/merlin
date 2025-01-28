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

// Secret represents an MLP secret present in a container that is mounted as an environment variable
type Secret struct {
	// Name of the secret as stored in MLP
	MLPSecretName string `json:"name"`

	// Name of the environment variable when the secret is mounted
	EnvVarSecretName string `json:"value"`
}

type Secrets []Secret

func (sec Secrets) Value() (driver.Value, error) {
	return json.Marshal(sec)
}

func (sec Secrets) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &sec)
}

// ToKubernetesEnvVars returns the representation of Kubernetes'
// v1.EnvVars.
func (sec Secrets) ToKubernetesEnvVars() []v1.EnvVar {
	kubeEnvVars := make([]v1.EnvVar, len(sec))

	for k, secret := range sec {
		kubeEnvVars[k] = v1.EnvVar{
			Name: secret.EnvVarSecretName,
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					Key: secret.MLPSecretName,
				},
			},
		}
	}

	return kubeEnvVars
}
