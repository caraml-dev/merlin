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

	"github.com/gojek/merlin/mlp"
	"github.com/gojek/merlin/utils"
)

const (
	LabelAppName          = "app"
	LabelComponent        = "component"
	LabelEnvironment      = "environment"
	LabelOrchestratorName = "orchestrator"
	LabelStreamName       = "stream"
	LabelTeamName         = "team"

	ComponentBatchJob      = "batch-job"
	ComponentImageBuilder  = "image-builder"
	ComponentModelEndpoint = "model-endpoint"
	ComponentModelVersion  = "model-version"
)

var reservedKeys = map[string]bool{
	LabelAppName:          true,
	LabelComponent:        true,
	LabelEnvironment:      true,
	LabelOrchestratorName: true,
	LabelStreamName:       true,
	LabelTeamName:         true,
}

var prefix string

// InitKubernetesLabeller builds a new KubernetesLabeller Singleton
func InitKubernetesLabeller(p string) {
	prefix = p
}

type Metadata struct {
	App         string
	Component   string
	Environment string
	Stream      string
	Team        string
	Labels      mlp.Labels
}

func (metadata Metadata) Value() (driver.Value, error) {
	return json.Marshal(metadata)
}

func (metadata *Metadata) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &metadata)
}

func (metadata *Metadata) ToLabel() map[string]string {
	labels := map[string]string{
		prefix + LabelAppName:          metadata.App,
		prefix + LabelComponent:        metadata.Component,
		prefix + LabelEnvironment:      metadata.Environment,
		prefix + LabelOrchestratorName: "merlin",
		prefix + LabelStreamName:       metadata.Stream,
		prefix + LabelTeamName:         metadata.Team,
	}

	for _, label := range metadata.Labels {
		// skip label that is trying to override reserved key
		if _, usingReservedKeys := reservedKeys[label.Key]; usingReservedKeys {
			continue
		}

		// skip label that has invalid key name
		if err := utils.IsValidLabel(label.Key); err != nil {
			continue
		}

		// skip label that has invalid value name
		if err := utils.IsValidLabel(label.Value); err != nil {
			continue
		}

		key := prefix + label.Key
		labels[key] = label.Value
	}

	return labels
}
