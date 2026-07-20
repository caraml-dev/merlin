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

package api

import (
	"errors"
	"fmt"
	"net/http"

	"gorm.io/gorm"
)

// NodeController serves node listing for a deployment environment.
type NodeController struct {
	*AppContext
}

// ListNodes returns the nodes (name, status, labels, taints) of the cluster
// backing the given environment, so the UI can offer them for model placement.
func (c *NodeController) ListNodes(r *http.Request, vars map[string]string, _ interface{}) *Response {
	ctx := r.Context()

	environmentName := vars["environment_name"]
	env, err := c.EnvironmentService.GetEnvironment(environmentName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFound(fmt.Sprintf("Environment not found: %v", err))
		}
		return InternalServerError(fmt.Sprintf("Error getting environment: %v", err))
	}

	// Route by environment name — this selects the same per-environment controller
	// that deployment uses, so nodes are read from the exact cluster the model
	// deploys to (in-cluster SA locally, or mTLS client cert for a remote cluster).
	nodes, err := c.NodeService.ListNodes(ctx, env.Name)
	if err != nil {
		return InternalServerError(fmt.Sprintf("Error listing nodes: %v", err))
	}

	return Ok(nodes)
}
