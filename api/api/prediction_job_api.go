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

	"github.com/gojek/merlin/log"
	"github.com/gojek/merlin/models"
	"github.com/gojek/merlin/service"
)

// PredictionJobController controls prediction job API.
type PredictionJobController struct {
	*AppContext
}

// Create method creates a prediction job.
func (c *PredictionJobController) Create(r *http.Request, vars map[string]string, body interface{}) *APIResponse {
	ctx := r.Context()

	modelID, _ := models.ParseId(vars["model_id"])
	versionID, _ := models.ParseId(vars["version_id"])

	model, version, err := c.getModelAndVersion(ctx, modelID, versionID)
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return InternalServerError(err.Error())
		}
		return NotFound(err.Error())
	}

	data, ok := body.(*models.PredictionJob)
	if !ok {
		return BadRequest("Unable to parse body as prediction job")
	}

	env, err := c.AppContext.EnvironmentService.GetDefaultPredictionJobEnvironment()
	if err != nil {
		return InternalServerError("Unable to find default environment, specify environment target for deployment")
	}

	predictionJob, err := c.PredictionJobService.CreatePredictionJob(env, model, version, data)
	if err != nil {
		log.Errorf("failed creating prediction job %v", err)
		return BadRequest(fmt.Sprintf("Failed creating prediction job %s", err))
	}

	return Ok(predictionJob)
}

// List method lists all prediction jobs of a model and version ID.
func (c *PredictionJobController) List(r *http.Request, vars map[string]string, _ interface{}) *APIResponse {
	ctx := r.Context()

	modelID, _ := models.ParseId(vars["model_id"])
	versionID, _ := models.ParseId(vars["version_id"])

	model, version, err := c.getModelAndVersion(ctx, modelID, versionID)
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return InternalServerError(err.Error())
		}
		return NotFound(err.Error())
	}

	query := &service.ListPredictionJobQuery{
		ModelId:   modelID,
		VersionId: versionID,
	}

	jobs, err := c.PredictionJobService.ListPredictionJobs(model.Project, query)
	if err != nil {
		log.Errorf("failed to list all prediction job for model %s version %s: %v", model.Name, version.Id, err)
		return InternalServerError("Failed listing prediction job")
	}

	return Ok(jobs)
}

// Get method gets a prediction job.
func (c *PredictionJobController) Get(r *http.Request, vars map[string]string, _ interface{}) *APIResponse {
	ctx := r.Context()

	modelID, _ := models.ParseId(vars["model_id"])
	versionID, _ := models.ParseId(vars["version_id"])
	id, _ := models.ParseId(vars["job_id"])

	model, version, err := c.getModelAndVersion(ctx, modelID, versionID)
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return InternalServerError(err.Error())
		}
		return NotFound(err.Error())
	}

	env, err := c.AppContext.EnvironmentService.GetDefaultPredictionJobEnvironment()
	if err != nil {
		return InternalServerError("Unable to find default environment, specify environment target for deployment")
	}

	job, err := c.PredictionJobService.GetPredictionJob(env, model, version, id)
	if err != nil {
		log.Errorf("failed to get prediction job %s for model %s version %s: %v", id, model.Name, version.Id, err)
		return InternalServerError("Failed reading prediction job")
	}

	return Ok(job)
}

// Stop method stops a prediction job.
func (c *PredictionJobController) Stop(r *http.Request, vars map[string]string, _ interface{}) *APIResponse {
	ctx := r.Context()

	modelID, _ := models.ParseId(vars["model_id"])
	versionID, _ := models.ParseId(vars["version_id"])
	id, _ := models.ParseId(vars["job_id"])

	model, version, err := c.getModelAndVersion(ctx, modelID, versionID)
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return InternalServerError(err.Error())
		}
		return NotFound(err.Error())
	}

	env, err := c.AppContext.EnvironmentService.GetDefaultPredictionJobEnvironment()
	if err != nil {
		return InternalServerError("Unable to find default environment, specify environment target for deployment")
	}

	_, err = c.PredictionJobService.StopPredictionJob(env, model, version, id)
	if err != nil {
		log.Errorf("failed to stop prediction job %v", err)
		return BadRequest(fmt.Sprintf("Failed stopping prediction job %s", err))
	}

	return NoContent()
}

// ListContainers method lists all containers of a prediction job.
func (c *PredictionJobController) ListContainers(r *http.Request, vars map[string]string, body interface{}) *APIResponse {
	ctx := r.Context()

	versionID, _ := models.ParseId(vars["version_id"])
	modelID, _ := models.ParseId(vars["model_id"])
	id, _ := models.ParseId(vars["job_id"])

	model, version, err := c.getModelAndVersion(ctx, modelID, versionID)
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return InternalServerError(err.Error())
		}
		return NotFound(err.Error())
	}

	env, err := c.AppContext.EnvironmentService.GetDefaultPredictionJobEnvironment()
	if err != nil {
		return InternalServerError("Unable to find default environment, specify environment target for deployment")
	}

	job, err := c.PredictionJobService.GetPredictionJob(env, model, version, id)
	if err != nil {
		log.Errorf("failed to get prediction job %s for model %s version %s: %v", id, model.Name, version.Id, err)
		return InternalServerError("Failed reading prediction job")
	}

	containers, err := c.PredictionJobService.ListContainers(env, model, version, job)
	if err != nil {
		log.Errorf("Error finding containers for model %v, version %v, reason: %v", model, version, err)
		return InternalServerError(fmt.Sprintf("Error while getting container for endpoint with model %v and version %v", model.Id, version.Id))
	}
	return Ok(containers)
}

// ListAllInProject lists all prediction jobs of a project.
func (c *PredictionJobController) ListAllInProject(r *http.Request, vars map[string]string, body interface{}) *APIResponse {
	ctx := r.Context()

	var query service.ListPredictionJobQuery
	err := decoder.Decode(&query, r.URL.Query())
	projectID, _ := models.ParseId(vars["project_id"])

	project, err := c.ProjectsService.GetByID(ctx, int32(projectID))
	if err != nil {
		return NotFound(err.Error())
	}

	jobs, err := c.PredictionJobService.ListPredictionJobs(project, &query)
	if err != nil {
		log.Errorf("failed to list all prediction job for model ")
		return InternalServerError("Failed listing prediction job")
	}

	return Ok(jobs)
}
