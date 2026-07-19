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

// NodePoolController serves node pool discovery for a deployment environment.
type NodePoolController struct {
	*AppContext
}

// ListNodePools returns the schedulable node pools (taint + shared labels) of the
// cluster backing the given environment, so the UI can offer them for placement.
func (c *NodePoolController) ListNodePools(r *http.Request, vars map[string]string, _ interface{}) *Response {
	ctx := r.Context()

	environmentName := vars["environment_name"]
	env, err := c.EnvironmentService.GetEnvironment(environmentName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFound(fmt.Sprintf("Environment not found: %v", err))
		}
		return InternalServerError(fmt.Sprintf("Error getting environment: %v", err))
	}

	nodePools, err := c.NodePoolService.ListNodePools(ctx, env.Cluster)
	if err != nil {
		return InternalServerError(fmt.Sprintf("Error listing node pools: %v", err))
	}

	return Ok(nodePools)
}
