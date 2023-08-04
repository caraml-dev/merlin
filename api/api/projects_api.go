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
	"fmt"
	"net/http"

	"github.com/caraml-dev/merlin/models"
)

// ProjectsController controls projects API.
type ProjectsController struct {
	*AppContext
}

// ListProjects lists all projects.
func (c *ProjectsController) ListProjects(r *http.Request, vars map[string]string, _ interface{}) *Response {
	ctx := r.Context()

	projects, err := c.ProjectsService.List(ctx, vars["name"])
	if err != nil {
		return InternalServerError(fmt.Sprintf("Error listing projects: %v", err))
	}

	return Ok(projects)
}

// GetProject gets a project of a project ID.
func (c *ProjectsController) GetProject(r *http.Request, vars map[string]string, body interface{}) *Response {
	ctx := r.Context()
	projectID, _ := models.ParseID(vars["project_id"])
	project, err := c.ProjectsService.GetByID(ctx, int32(projectID))
	if err != nil {
		return NotFound(fmt.Sprintf("Project not found: %v", err))
	}

	return Ok(project)
}
