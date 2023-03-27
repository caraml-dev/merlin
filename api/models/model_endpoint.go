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

	"github.com/caraml-dev/merlin/pkg/protocol"
	"github.com/google/uuid"
)

// ModelEndpoint stores model's endpoint data.
type ModelEndpoint struct {
	ID              ID                 `json:"id" gorm:"primary_key;"`
	ModelID         ID                 `json:"model_id"`
	Model           *Model             `json:"model"`
	Status          EndpointStatus     `json:"status"`
	URL             string             `json:"url" gorm:"url"`
	Rule            *ModelEndpointRule `json:"rule" gorm:"rule"`
	Environment     *Environment       `json:"environment" gorm:"association_foreignkey:Name"`
	EnvironmentName string             `json:"environment_name"`
	Protocol        protocol.Protocol  `json:"protocol" gorm:"protocol"`
	CreatedUpdated
}

// ModelEndpointRule describes model's endpoint traffic rule.
type ModelEndpointRule struct {
	Destination []*ModelEndpointRuleDestination `json:"destinations"`
	Mirror      *VersionEndpoint                `json:"mirror,omitempty"`
}

func (rule ModelEndpointRule) Value() (driver.Value, error) {
	return json.Marshal(rule)
}

func (rule *ModelEndpointRule) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &rule)
}

// ModelEndpointRuleDestination describes forwarding target.
type ModelEndpointRuleDestination struct {
	VersionEndpointID uuid.UUID        `json:"version_endpoint_id"`
	VersionEndpoint   *VersionEndpoint `json:"version_endpoint"`
	Weight            int32            `json:"weight"`
}
