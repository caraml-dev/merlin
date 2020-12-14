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

// Deployment is a struct representing an attempt to deploy a version_endpoint
type Deployment struct {
	ID                ID             `json:"id"`
	ProjectID         ID             `json:"project_id"`
	VersionID         ID             `json:"version_id"`
	VersionModelID    ID             `json:"model_id"`
	VersionEndpointID uuid.UUID      `json:"version_endpoint_id"`
	Status            EndpointStatus `json:"status"`
	Error             string         `json:"error"`
	CreatedUpdated
}
