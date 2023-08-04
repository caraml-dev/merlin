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

	"github.com/caraml-dev/merlin/mlp"
	"github.com/caraml-dev/merlin/models"
)

type SecretsController struct {
	*AppContext
}

func (c *SecretsController) CreateSecret(r *http.Request, vars map[string]string, body interface{}) *Response {
	ctx := r.Context()

	projectID, _ := models.ParseID(vars["project_id"])
	_, err := c.ProjectsService.GetByID(ctx, int32(projectID))
	if err != nil {
		return NotFound(fmt.Sprintf("Project not found: %v", err))
	}

	secret, ok := body.(*mlp.Secret)
	if !ok {
		return BadRequest("Invalid request body")
	}

	newSecret, err := c.SecretService.Create(ctx, int32(projectID), *secret)
	if err != nil {
		return InternalServerError(fmt.Sprintf("Error creating secret: %v", err))
	}

	return Created(newSecret)
}

func (c *SecretsController) UpdateSecret(r *http.Request, vars map[string]string, body interface{}) *Response {
	ctx := r.Context()

	projectID, _ := models.ParseID(vars["project_id"])
	secretID, _ := models.ParseID(vars["secret_id"])
	if projectID <= 0 || secretID <= 0 {
		return BadRequest("project_id and secret_id is not valid")
	}

	_, err := c.ProjectsService.GetByID(ctx, int32(projectID))
	if err != nil {
		return NotFound(fmt.Sprintf("Project not found: %v", err))
	}

	secret, ok := body.(*mlp.Secret)
	if !ok {
		return BadRequest("Invalid request body")
	}

	updatedSecret, err := c.SecretService.Update(ctx, int32(projectID), *secret)
	if err != nil {
		return InternalServerError(fmt.Sprintf("Error updating secret: %v", err))
	}
	return Ok(updatedSecret)
}

func (c *SecretsController) DeleteSecret(r *http.Request, vars map[string]string, _ interface{}) *Response {
	ctx := r.Context()

	projectID, _ := models.ParseID(vars["project_id"])
	secretID, _ := models.ParseID(vars["secret_id"])
	if projectID <= 0 || secretID <= 0 {
		return BadRequest("project_id and secret_id is not valid")
	}

	_, err := c.ProjectsService.GetByID(ctx, int32(projectID))
	if err != nil {
		return NotFound(fmt.Sprintf("Project not found: %v", err))
	}

	if err := c.SecretService.Delete(ctx, int32(secretID), int32(projectID)); err != nil {
		return InternalServerError(fmt.Sprintf("Error deleting secret: %v", err))
	}

	return NoContent()
}

func (c *SecretsController) ListSecret(r *http.Request, vars map[string]string, body interface{}) *Response {
	ctx := r.Context()

	projectID, _ := models.ParseID(vars["project_id"])
	_, err := c.ProjectsService.GetByID(ctx, int32(projectID))
	if err != nil {
		return NotFound(fmt.Sprintf("Project not found: %v", err))
	}

	secrets, err := c.SecretService.List(ctx, int32(projectID))
	if err != nil {
		return InternalServerError(fmt.Sprintf("Error listing secrets: %v", err))
	}

	return Ok(secrets)
}
