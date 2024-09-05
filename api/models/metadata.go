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

	"github.com/caraml-dev/merlin/cluster/labeller"
	"github.com/caraml-dev/merlin/mlp"
	"github.com/caraml-dev/merlin/utils"
)

const (
	// orchestratorValue is the value of the orchestrator (which is Merlin)
	orchestratorValue = "merlin"

	ComponentBatchJob      = "batch-job"
	ComponentImageBuilder  = "image-builder"
	ComponentModelEndpoint = "model-endpoint"
	ComponentModelVersion  = "model-version"
)

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
		labeller.GetLabelName(labeller.LabelAppName):          metadata.App,
		labeller.GetLabelName(labeller.LabelComponent):        metadata.Component,
		labeller.GetLabelName(labeller.LabelEnvironment):      labeller.GetEnvironment(),
		labeller.GetLabelName(labeller.LabelOrchestratorName): orchestratorValue,
		labeller.GetLabelName(labeller.LabelStreamName):       metadata.Stream,
		labeller.GetLabelName(labeller.LabelTeamName):         metadata.Team,
	}

	for _, label := range metadata.Labels {
		// skip label that is trying to override reserved key
		if labeller.IsReservedKey(labeller.GetPrefix() + label.Key) {
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
