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

	"github.com/gojek/mlp/api/client"
	"github.com/gojek/mlp/api/pkg/authz/enforcer"

	"github.com/gojek/merlin/mlp"
	"github.com/gojek/merlin/models"
)

type ProjectsController struct {
	*AppContext
}

func (c *ProjectsController) ListProjects(r *http.Request, vars map[string]string, _ interface{}) *ApiResponse {
	ctx := r.Context()

	projects, err := c.ProjectsService.List(ctx, vars["name"])
	if err != nil {
		return InternalServerError(err.Error())
	}

	user := vars["user"]
	projects, err = c.filterAuthorizedProjects(user, projects, enforcer.ActionRead)
	if err != nil {
		return InternalServerError(err.Error())
	}

	return Ok(projects)
}

func (c *ProjectsController) GetProject(r *http.Request, vars map[string]string, body interface{}) *ApiResponse {
	ctx := r.Context()
	projectId, _ := models.ParseId(vars["project_id"])
	project, err := c.ProjectsService.GetByID(ctx, int32(projectId))
	if err != nil {
		return NotFound(err.Error())
	}

	return Ok(project)
}

func (c *ProjectsController) filterAuthorizedProjects(user string, projects mlp.Projects, action string) (mlp.Projects, error) {
	if c.AuthorizationEnabled {
		projectIds := make([]string, 0, 0)
		allowedProjects := mlp.Projects{}
		projectMap := make(map[string]mlp.Project)
		for _, project := range projects {
			projectId := fmt.Sprintf("projects:%d", project.Id)
			projectIds = append(projectIds, projectId)
			projectMap[projectId] = mlp.Project(project)
		}

		allowedProjectIds, err := c.Enforcer.FilterAuthorizedResource(user, projectIds, action)
		if err != nil {
			return nil, err
		}

		for _, projectId := range allowedProjectIds {
			allowedProjects = append(allowedProjects, client.Project(projectMap[projectId]))
		}

		return allowedProjects, nil
	} else {
		return projects, nil
	}
}
