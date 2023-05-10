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

	projectID, _ := models.ParseID(vars["project_id"])
	modelID, _ := models.ParseID(vars["model_id"])

	_, err := c.ProjectsService.GetByID(ctx, int32(projectID))
	if err != nil {
		return NotFound(err.Error())
	}

	model, err := c.ModelsService.FindByID(ctx, modelID)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return NotFound(fmt.Sprintf("Model id %s not found", modelID))
		}
		return InternalServerError(err.Error())
	}

	// GET MODEL VERSION
	versions, _, err := c.VersionsService.ListVersions(ctx, modelID, c.MonitoringConfig, service.VersionQuery{})
	if err != nil {
		return InternalServerError(err.Error())
	}

	// CHECK CONDITION FOR EVERY VERSION
	for _, version := range versions {
		// CHECK PREDICTION JOBS
		query := &service.ListPredictionJobQuery{
			ModelID:   modelID,
			VersionID: version.ID,
		}

		jobs, err := c.PredictionJobService.ListPredictionJobs(ctx, model.Project, query)
		if err != nil {
			log.Errorf("failed to list all prediction job for model %s version %s: %v", model.Name, version.ID, err)
			return InternalServerError("Failed listing prediction job")
		}

		for _, item := range jobs {
			if item.Status == models.JobPending {
				return BadRequest("There are active prediction job using this model version")
			}
		}
		// CHECK IF THERE IS ANY ENDPOINT WITH STATUS NOT TERMINATED
		endpoints, err := c.EndpointsService.ListEndpoints(ctx, model, version)
		if err != nil {
			log.Errorf("failed to list all endpoint for model %s version %s: %v", model.Name, version.ID, err)
			return InternalServerError("Failed listing model version endpoint")
		}

		for _, item := range endpoints {
			if item.Status != models.EndpointTerminated {
				return BadRequest("There are endpoint that still using this model version")
			}
		}

	}

	// DELETING ALL THE RELATED ENTITY
	for _, version := range versions {

		query := &service.ListPredictionJobQuery{
			ModelID:   modelID,
			VersionID: version.ID,
		}

		jobs, err := c.PredictionJobService.ListPredictionJobs(ctx, model.Project, query)
		if err != nil {
			log.Errorf("failed to list all prediction job for model %s version %s: %v", model.Name, version.ID, err)
			return InternalServerError("Failed listing prediction job")
		}

		// DELETE PREDICTION JOBS
		for _, item := range jobs {
			_, err = c.PredictionJobService.StopPredictionJob(ctx, item.Environment, model, version, item.ID)
			if err != nil {
				log.Errorf("failed to stop prediction job %v", err)
				return BadRequest(fmt.Sprintf("Failed stopping prediction job %s", err))
			}
		}

		endpoints, err := c.EndpointsService.ListEndpoints(ctx, model, version)
		if err != nil {
			log.Errorf("failed to list all endpoint for model %s version %s: %v", model.Name, version.ID, err)
			return InternalServerError("Failed listing model version endpoint")
		}

		// DELETE ENDPOINTS
		for _, item := range endpoints {
			if item.Status != models.EndpointTerminated {
				err = c.EndpointsService.DeleteEndpoint(version, item)
				if err != nil {
					log.Errorf("failed to undeploy endpoint job %v", err)
					return InternalServerError(fmt.Sprintf("Failed to delete Endpoint %s", err))
				}
			}
		}
	}

	// UNDEPLOY MODEL ENDPOINT
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
	//Delete Data from DB
	err = c.ModelsService.Delete(model)
	if err != nil {
		return InternalServerError(err.Error())
	}

	// Delete Data from mlflow
	s := strconv.FormatUint(uint64(model.ExperimentID), 10)
	err = c.MlflowDeleteService.DeleteExperiment(ctx, s, true)
	if err != nil {
		return InternalServerError(err.Error())
	}

	return Ok(model)
}
