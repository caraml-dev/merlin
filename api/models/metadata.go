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
	"fmt"
	"regexp"

	"github.com/gojek/merlin/mlp"
	"github.com/gojek/merlin/utils"
)

const (
	labelAppName          = "app"
	labelComponent        = "component"
	labelEnvironment      = "environment"
	LabelOrchestratorName = "orchestrator"
	labelStreamName       = "stream"
	labelTeamName         = "team"

	// orchestratorValue is the value of the orchestrator (which is Merlin)
	orchestratorValue = "merlin"

	ComponentBatchJob      = "batch-job"
	ComponentImageBuilder  = "image-builder"
	ComponentModelEndpoint = "model-endpoint"
	ComponentModelVersion  = "model-version"
)

var reservedKeys = map[string]bool{
	labelAppName:          true,
	labelComponent:        true,
	labelEnvironment:      true,
	LabelOrchestratorName: true,
	labelStreamName:       true,
	labelTeamName:         true,
}

var (
	prefix           string
	environment      string
	validPrefixRegex = regexp.MustCompile("^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?(/)$")
)

// InitKubernetesLabeller builds a new KubernetesLabeller Singleton
func InitKubernetesLabeller(p, e string) error {
	if len(p) > 253 {
		return fmt.Errorf("length of prefix is greater than 253 characters")
	}

	if isValidPrefix := validPrefixRegex.MatchString(p); !isValidPrefix {
		return fmt.Errorf("name violates kubernetes label's prefix constraint")
	}

	prefix = p
	environment = e
	return nil
}

type Metadata struct {
	App       string
	Component string
	Stream    string
	Team      string
	Labels    mlp.Labels
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
		prefix + labelAppName:          metadata.App,
		prefix + labelComponent:        metadata.Component,
		prefix + labelEnvironment:      environment,
		prefix + LabelOrchestratorName: orchestratorValue,
		prefix + labelStreamName:       metadata.Stream,
		prefix + labelTeamName:         metadata.Team,
	}

	for _, label := range metadata.Labels {
		// skip label that is trying to override reserved key
		if _, usingReservedKeys := reservedKeys[prefix+label.Key]; usingReservedKeys {
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

		labels[label.Key] = label.Value
	}

	return labels
}
