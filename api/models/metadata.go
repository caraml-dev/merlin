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
	labelTeamName         = "gojek.com/team"
	labelStreamName       = "gojek.com/stream"
	labelAppName          = "gojek.com/app"
	labelOrchestratorName = "gojek.com/orchestrator"
	labelEnvironment      = "gojek.com/environment"
)

var reservedKeys = map[string]bool{
	labelTeamName:         true,
	labelStreamName:       true,
	labelAppName:          true,
	labelOrchestratorName: true,
	labelEnvironment:      true,
}

type Metadata struct {
	Team        string
	Stream      string
	App         string
	Environment string
	Labels      mlp.Labels
	LabelPrefix string
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
		labelOrchestratorName: "merlin",
		labelAppName:          metadata.App,
		labelEnvironment:      metadata.Environment,
		labelStreamName:       metadata.Stream,
		labelTeamName:         metadata.Team,
	}

	for _, label := range metadata.Labels {
		key := metadata.LabelPrefix + label.Key
		// skip label that is trying to override reserved key
		if _, usingReservedKeys := reservedKeys[key]; usingReservedKeys {
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
		labels[key] = label.Value
	}
	return labels
}
