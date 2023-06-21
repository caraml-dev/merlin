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
	"strconv"

	"github.com/caraml-dev/merlin/log"
	"github.com/caraml-dev/merlin/service"

	"github.com/jinzhu/gorm"

	"github.com/caraml-dev/merlin/mlflow"
	"github.com/caraml-dev/merlin/models"
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
		if gorm.IsRecordNotFoundError(err) {
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
		return NotFound(err.Error())
	}

	model, err := c.ModelsService.FindByID(ctx, modelID)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return NotFound(fmt.Sprintf("Model is not found: %s", err.Error()))
		}
		return InternalServerError(err.Error())
	}

	// get all model version for this model
	versions, _, err := c.VersionsService.ListVersions(ctx, modelID, c.MonitoringConfig, service.VersionQuery{})
	if err != nil {
		return InternalServerError(err.Error())
	}

	var allJobInModel []*models.PredictionJob
	var allEndpointInModel []*models.VersionEndpoint

	// iterate for every version, check the active prediction job / active endpoint
	for _, version := range versions {
		if model.Type == "pyfunc_v2" {
			// check active prediction job
			// if there are any active prediction job using the model version, deletion of the model version are prohibited
			jobs, response := c.VersionsController.checkActivePredictionJobs(ctx, model, version)
			if response != nil {
				return response
			}

			// save all prediction jobs for deletion process later on
			allJobInModel = append(allJobInModel, jobs...)

		} else {
			// handle for model with type non pyfunc
			// check active endpoints
			// if there are any active endpoint using the model version, deletion of the model version are prohibited
			endpoints, response := c.VersionsController.checkActiveEndpoints(ctx, model, version)
			if response != nil {
				return response
			}

			// save all endpoint for deletion process later on
			allEndpointInModel = append(allEndpointInModel, endpoints...)

		}
	}

	// DELETING ALL THE RELATED ENTITY
	for _, version := range versions {
		if model.Type == "pyfunc_v2" {
			// DELETE PREDICTION JOBS
			response := c.VersionsController.deleteInactivePredictionJobs(ctx, allJobInModel, model, version)
			if response != nil {
				log.Errorf("failed to stop prediction job %v", response.data)
				//return InternalServerError(fmt.Sprintf("Failed stopping prediction job: %s", err))
				return response
			}
		} else {
			// DELETE ENDPOINTS
			response := c.VersionsController.deleteInactiveVersionEndpoints(allEndpointInModel, version)
			if response != nil {
				log.Errorf("failed to delete version endpoints %v", response.data)
				//return InternalServerError(fmt.Sprintf("Failed deleting version endpoints: %s", err))
				return response
			}
		}
		// DELETE VERSION
		err = c.VersionsService.Delete(version)
		if err != nil {
			log.Errorf("failed to delete model version %v", err)
			return InternalServerError(fmt.Sprintf("Delete Failed: %s", err.Error()))
		}
	}

	if model.Type != "pyfunc_v2" {
		// undeploy all existing model endpoint
		modelEndpoints, err := c.ModelEndpointsService.ListModelEndpoints(ctx, modelID)
		if err != nil {
			return InternalServerError(err.Error())
		}

		for _, modelEndpoint := range modelEndpoints {
			_, err := c.ModelEndpointsService.UndeployEndpoint(ctx, model, modelEndpoint)
			if err != nil {
				return InternalServerError(fmt.Sprintf("Unable to delete model endpoint: %s", err.Error()))
			}
		}
	}

	// Delete mlflow experiment and artifact if there are model version
	if len(versions) != 0 {
		s := strconv.FormatUint(uint64(model.ExperimentID), 10)
		err = c.MlflowDeleteService.DeleteExperiment(ctx, s, true)
		if err != nil {
			return InternalServerError(fmt.Sprintf("Delete Failed: %s", err.Error()))
		}
	}

	// Delete model data from DB
	err = c.ModelsService.Delete(model)
	if err != nil {
		return InternalServerError(fmt.Sprintf("Delete Failed: %s", err.Error()))
	}

	return Ok(model.ID)
}
