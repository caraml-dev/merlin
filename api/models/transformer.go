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

import "github.com/google/uuid"

type TransformerType string

const (
	DefaultTransformerType  = ""
	CustomTransformerType   = "custom"
	StandardTransformerType = "standard"
)

// Transformer is a service for pre/post-processing steps.
type Transformer struct {
	ID                string          `json:"id"`
	Enabled           bool            `json:"enabled"`
	VersionEndpointID uuid.UUID       `json:"version_endpoint_id"`
	TransformerType   TransformerType `json:"transformer_type"`

	// Docker image name.
	Image   string `json:"image"`
	Command string `json:"command,omitempty"`
	Args    string `json:"args,omitempty"`

	ResourceRequest *ResourceRequest `json:"resource_request"`
	EnvVars         EnvVars          `json:"env_vars"`
	CreatedUpdated
}
