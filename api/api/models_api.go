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
	"strconv"

	"gorm.io/gorm"

	"github.com/caraml-dev/merlin/mlflow"
	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/service"
)

// ModelsController controls models API.
type ModelsController struct {
	*AppContext
	*VersionsController
}

// ListModels list all models of a project.
func (c *ModelsController) ListModels(r *http.Request, vars map[string]string, _ interface{}) *Response {
	ctx := r.Context()

	projectID, _ := models.ParseID(vars["project_id"])

	models, err := c.ModelsService.ListModels(ctx, projectID, vars["name"])
	if err != nil {
		return InternalServerError(fmt.Sprintf("Error listing models: %v", err))
	}

	return Ok(models)
}

// CreateModel creates a new model in an existing project.
func (c *ModelsController) CreateModel(r *http.Request, vars map[string]string, body interface{}) *Response {
	ctx := r.Context()

	model := body.(*models.Model)

	projectID, _ := models.ParseID(vars["project_id"])
	project, err := c.ProjectsService.GetByID(ctx, int32(projectID))
	if err != nil {
		return NotFound(fmt.Sprintf("Project not found: %v", err))
	}

	experimentName := fmt.Sprintf("%s/%s", project.Name, model.Name)
	experimentID, err := c.MlflowClient.CreateExperiment(experimentName)
	if err != nil {
		switch err.Error() {
		case mlflow.ResourceAlreadyExists:
			return BadRequest(fmt.Sprintf("MLflow experiment for model with name `%s` already exists", model.Name))
		default:
			return InternalServerError(fmt.Sprintf("Failed to create mlflow experiment: %v", err))
		}
	}

	model.ProjectID = projectID
	model.ExperimentID, _ = models.ParseID(experimentID)

	model, err = c.ModelsService.Save(ctx, model)
	if err != nil {
		return InternalServerError(fmt.Sprintf("Error saving model: %v", err))
	}

	return Created(model)
}

// GetModel gets model given a project and model ID.
func (c *ModelsController) GetModel(r *http.Request, vars map[string]string, body interface{}) *Response {
	ctx := r.Context()

	projectID, _ := models.ParseID(vars["project_id"])
	modelID, _ := models.ParseID(vars["model_id"])

	_, err := c.ProjectsService.GetByID(ctx, int32(projectID))
	if err != nil {
		return NotFound(fmt.Sprintf("Model not found: %v", err))
	}

	model, err := c.ModelsService.FindByID(ctx, modelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFound(fmt.Sprintf("Model not found: %v", err))
		}
		return InternalServerError(fmt.Sprintf("Error getting model: %v", err))
	}

	return Ok(model)
}

// DeleteModel delete a model given a project and model ID
func (c *ModelsController) DeleteModel(r *http.Request, vars map[string]string, body interface{}) *Response {
	ctx := r.Context()

	projectID, err := models.ParseID(vars["project_id"])
	if err != nil {
		return BadRequest("Unable to parse project_id")
	}
	modelID, err := models.ParseID(vars["model_id"])
	if err != nil {
		return BadRequest("Unable to parse model_id")
	}

	_, err = c.ProjectsService.GetByID(ctx, int32(projectID))
	if err != nil {
		return NotFound(fmt.Sprintf("Project not found: %v", err))
	}

	model, err := c.ModelsService.FindByID(ctx, modelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFound(fmt.Sprintf("Model is not found: %s", err.Error()))
		}
		return InternalServerError(fmt.Sprintf("Error getting model: %v", err.Error()))
	}

	// get all model version for this model
	versions, _, err := c.VersionsService.ListVersions(ctx, modelID, c.MonitoringConfig, service.VersionQuery{})
	if err != nil {
		return InternalServerError(fmt.Sprintf("Error getting list of versions: %v", err.Error()))
	}

	var inactiveJobsInModel [][]*models.PredictionJob
	var inactiveEndpointsInModel [][]*models.VersionEndpoint

	// iterate for every version, check the active prediction job / active endpoint
	for _, version := range versions {
		if model.Type == "pyfunc_v2" {
			// check active prediction job
			// if there are any active prediction job using the model version, deletion of the model version are prohibited
			jobs, errResponse := c.VersionsController.getInactivePredictionJobsForDeletion(ctx, model, version)
			if errResponse != nil {
				return errResponse
			}

			// save all prediction jobs for deletion process later on
			inactiveJobsInModel = append(inactiveJobsInModel, jobs)

		}

		// check active endpoints for all model type
		// if there are any active endpoint using the model version, deletion of the model version are prohibited
		endpoints, errResponse := c.VersionsController.getInactiveEndpointsForDeletion(ctx, model, version)
		if errResponse != nil {
			return errResponse
		}

		// save all endpoint for deletion process later on
		inactiveEndpointsInModel = append(inactiveEndpointsInModel, endpoints)

	}

	// DELETING ALL THE RELATED ENTITY
	for index, version := range versions {
		if model.Type == "pyfunc_v2" {
			// delete inactive prediction jobs for model with type pyfunc_v2
			errResponse := c.VersionsController.deletePredictionJobs(ctx, inactiveJobsInModel[index], model, version)
			if errResponse != nil {
				return errResponse
			}
		}

		// delete inactive endpoints for all model type
		errResponse := c.VersionsController.deleteVersionEndpoints(inactiveEndpointsInModel[index], version)
		if errResponse != nil {
			return errResponse
		}

		// DELETE VERSION
		err = c.VersionsService.Delete(version)
		if err != nil {
			return InternalServerError(fmt.Sprintf("Delete version id %d failed: %s", version.ID, err.Error()))
		}
	}

	// undeploy all existing model endpoint
	modelEndpoints, err := c.ModelEndpointsService.ListModelEndpoints(ctx, modelID)
	if err != nil {
		return InternalServerError(fmt.Sprintf("Error getting list of model endpoint: %v", err.Error()))
	}

	for _, modelEndpoint := range modelEndpoints {
		err := c.ModelEndpointsService.DeleteModelEndpoint(modelEndpoint)
		if err != nil {
			return InternalServerError(fmt.Sprintf("Unable to delete model endpoint: %s", err.Error()))
		}
	}

	// Delete mlflow experiment and artifact if there are model version
	if len(versions) != 0 {
		s := strconv.FormatUint(uint64(model.ExperimentID), 10)
		err = c.MlflowDeleteService.DeleteExperiment(ctx, s, true)
		if err != nil {
			return InternalServerError(fmt.Sprintf("Delete mlflow experiment failed: %s", err.Error()))
		}
	}

	// Delete model data from DB
	err = c.ModelsService.Delete(model)
	if err != nil {
		return InternalServerError(fmt.Sprintf("Delete model failed: %s", err.Error()))
	}

	return Ok(model.ID)
}
