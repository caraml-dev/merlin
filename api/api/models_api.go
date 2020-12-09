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

	"github.com/jinzhu/gorm"

	"github.com/gojek/merlin/mlflow"
	"github.com/gojek/merlin/models"
)

// ModelsController controls models API.
type ModelsController struct {
	*AppContext
}

// ListModels list all models of a project.
func (c *ModelsController) ListModels(r *http.Request, vars map[string]string, _ interface{}) *APIResponse {
	ctx := r.Context()

	projectID, _ := models.ParseId(vars["project_id"])

	models, err := c.ModelsService.ListModels(ctx, projectID, vars["name"])
	if err != nil {
		return InternalServerError(err.Error())
	}

	return Ok(models)
}

// CreateModel creates a new model in an existing project.
func (c *ModelsController) CreateModel(r *http.Request, vars map[string]string, body interface{}) *APIResponse {
	ctx := r.Context()

	model := body.(*models.Model)

	projectID, _ := models.ParseId(vars["project_id"])
	project, err := c.ProjectsService.GetByID(ctx, int32(projectID))
	if err != nil {
		return NotFound(err.Error())
	}

	mlflowClient := mlflow.NewClient(nil, project.MlflowTrackingUrl)
	experimentName := fmt.Sprintf("%s/%s", project.Name, model.Name)

	experimentID, err := mlflowClient.CreateExperiment(experimentName)
	if err != nil {
		switch err.Error() {
		case mlflow.ResourceAlreadyExists:
			return BadRequest(fmt.Sprintf("MLflow experiment for model with name `%s` already exists", model.Name))
		default:
			return InternalServerError(fmt.Sprintf("Failed to create mlflow experiment: %s", err.Error()))
		}
	}

	model.ProjectId = projectID
	model.ExperimentId, _ = models.ParseId(experimentID)

	model, err = c.ModelsService.Save(ctx, model)
	if err != nil {
		return InternalServerError(err.Error())
	}

	return Created(model)
}

// GetModel gets model given a project and model ID.
func (c *ModelsController) GetModel(r *http.Request, vars map[string]string, body interface{}) *APIResponse {
	ctx := r.Context()

	projectID, _ := models.ParseId(vars["project_id"])
	modelID, _ := models.ParseId(vars["model_id"])

	_, err := c.ProjectsService.GetByID(ctx, int32(projectID))
	if err != nil {
		return NotFound(err.Error())
	}

	model, err := c.ModelsService.FindById(ctx, modelID)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return NotFound(fmt.Sprintf("Model id %s not found", modelID))
		}

		return InternalServerError(err.Error())
	}

	return Ok(model)
}
