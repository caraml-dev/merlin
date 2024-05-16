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

	"github.com/caraml-dev/mlp/api/pkg/pagination"

	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/service"
)

// PredictionJobController controls prediction job API.
type PredictionJobController struct {
	*AppContext
}

type ListJobsPaginatedResponse struct {
	Results []*models.PredictionJob `json:"results"`
	Paging  pagination.Paging       `json:"paging"`
}

// Create method creates a prediction job.
func (c *PredictionJobController) Create(r *http.Request, vars map[string]string, body interface{}) *Response {
	ctx := r.Context()

	modelID, _ := models.ParseID(vars["model_id"])
	versionID, _ := models.ParseID(vars["version_id"])

	model, version, err := c.getModelAndVersion(ctx, modelID, versionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFound(fmt.Sprintf("Model / version not found: %v", err))
		}
		return InternalServerError(fmt.Sprintf("Error getting model / version: %v", err))
	}

	data, ok := body.(*models.PredictionJob)
	if !ok {
		return BadRequest("Unable to parse body as prediction job")
	}

	env, err := c.AppContext.EnvironmentService.GetDefaultPredictionJobEnvironment()
	if err != nil {
		return InternalServerError(fmt.Sprintf("Unable to find default environment, specify environment target for deployment: %v", err))
	}

	predictionJob, err := c.PredictionJobService.CreatePredictionJob(ctx, env, model, version, data)
	if err != nil {
		return BadRequest(fmt.Sprintf("Error creating prediction job: %v", err))
	}

	return Ok(predictionJob)
}

// List method lists all prediction jobs of a model and version ID.
func (c *PredictionJobController) List(r *http.Request, vars map[string]string, _ interface{}) *Response {
	ctx := r.Context()

	versionID, _ := models.ParseID(vars["version_id"])
	modelID, _ := models.ParseID(vars["model_id"])

	model, _, err := c.getModelAndVersion(ctx, modelID, versionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFound(fmt.Sprintf("Model / version not found: %v", err))
		}
		return InternalServerError(fmt.Sprintf("Error getting model / version: %v", err))
	}

	query := &service.ListPredictionJobQuery{
		ModelID:   modelID,
		VersionID: versionID,
	}
	jobs, _, err := c.PredictionJobService.ListPredictionJobs(ctx, model.Project, query, false)
	if err != nil {
		return InternalServerError(fmt.Sprintf("Error listing prediction jobs: %v", err))
	}

	return Ok(jobs)
}

// ListInPage method lists all prediction jobs of a model and version ID, with pagination.
func (c *PredictionJobController) ListByPage(r *http.Request, vars map[string]string, _ interface{}) *Response {
	ctx := r.Context()

	versionID, _ := models.ParseID(vars["version_id"])
	modelID, _ := models.ParseID(vars["model_id"])

	model, _, err := c.getModelAndVersion(ctx, modelID, versionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFound(fmt.Sprintf("Model / version not found: %v", err))
		}
		return InternalServerError(fmt.Sprintf("Error getting model / version: %v", err))
	}

	var query service.ListPredictionJobQuery
	err = decoder.Decode(&query, r.URL.Query())
	if err != nil {
		return BadRequest(fmt.Sprintf("Bad query %s", r.URL.Query()))
	}
	query.ModelID = modelID
	query.VersionID = versionID

	jobs, paging, err := c.PredictionJobService.ListPredictionJobs(ctx, model.Project, &query, true)
	if err != nil {
		return InternalServerError(fmt.Sprintf("Error listing prediction jobs: %v", err))
	}

	return Ok(ListJobsPaginatedResponse{
		Results: jobs,
		Paging:  *paging,
	})
}

// Get method gets a prediction job.
func (c *PredictionJobController) Get(r *http.Request, vars map[string]string, _ interface{}) *Response {
	ctx := r.Context()

	modelID, _ := models.ParseID(vars["model_id"])
	versionID, _ := models.ParseID(vars["version_id"])
	id, _ := models.ParseID(vars["job_id"])

	model, version, err := c.getModelAndVersion(ctx, modelID, versionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFound(fmt.Sprintf("Model / version not found: %v", err))
		}
		return InternalServerError(fmt.Sprintf("Error getting model / version: %v", err))
	}

	env, err := c.AppContext.EnvironmentService.GetDefaultPredictionJobEnvironment()
	if err != nil {
		return InternalServerError(fmt.Sprintf("Unable to find default environment, specify environment target for deployment: %v", err))
	}

	job, err := c.PredictionJobService.GetPredictionJob(ctx, env, model, version, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFound(fmt.Sprintf("Prediction job not found: %v", err))
		}
		return InternalServerError(fmt.Sprintf("Error getting prediction job: %v", err))
	}

	return Ok(job)
}

// Stop method stops a prediction job.
func (c *PredictionJobController) Stop(r *http.Request, vars map[string]string, _ interface{}) *Response {
	ctx := r.Context()

	modelID, _ := models.ParseID(vars["model_id"])
	versionID, _ := models.ParseID(vars["version_id"])
	id, _ := models.ParseID(vars["job_id"])

	model, version, err := c.getModelAndVersion(ctx, modelID, versionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFound(fmt.Sprintf("Model / version not found: %v", err))
		}
		return InternalServerError(fmt.Sprintf("Error getting model / version: %v", err))
	}

	env, err := c.AppContext.EnvironmentService.GetDefaultPredictionJobEnvironment()
	if err != nil {
		return InternalServerError(fmt.Sprintf("Unable to find default environment, specify environment target for deployment: %v", err))
	}

	_, err = c.PredictionJobService.StopPredictionJob(ctx, env, model, version, id)
	if err != nil {
		return BadRequest(fmt.Sprintf("Error stopping prediction job: %v", err))
	}

	return NoContent()
}

// ListContainers method lists all containers of a prediction job.
func (c *PredictionJobController) ListContainers(r *http.Request, vars map[string]string, body interface{}) *Response {
	ctx := r.Context()

	versionID, _ := models.ParseID(vars["version_id"])
	modelID, _ := models.ParseID(vars["model_id"])
	id, _ := models.ParseID(vars["job_id"])

	model, version, err := c.getModelAndVersion(ctx, modelID, versionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFound(fmt.Sprintf("Model / version not found: %v", err))
		}
		return InternalServerError(fmt.Sprintf("Error getting model / version: %v", err))
	}

	env, err := c.AppContext.EnvironmentService.GetDefaultPredictionJobEnvironment()
	if err != nil {
		return InternalServerError(fmt.Sprintf("Unable to find default environment, specify environment target for deployment: %v", err))
	}

	job, err := c.PredictionJobService.GetPredictionJob(ctx, env, model, version, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFound(fmt.Sprintf("Prediction job not found: %v", err))
		}
		return InternalServerError(fmt.Sprintf("Error getting prediction job: %v", err))
	}

	containers, err := c.PredictionJobService.ListContainers(ctx, env, model, version, job)
	if err != nil {
		return InternalServerError(fmt.Sprintf("Error while getting containers for endpoint: %v", err))
	}
	return Ok(containers)
}

// ListAllInProject lists all prediction jobs of a project.
func (c *PredictionJobController) ListAllInProject(r *http.Request, vars map[string]string, body interface{}) *Response {
	ctx := r.Context()

	var query service.ListPredictionJobQuery
	err := decoder.Decode(&query, r.URL.Query())
	if err != nil {
		return BadRequest(fmt.Sprintf("Bad query %s", r.URL.Query()))
	}

	projectID, _ := models.ParseID(vars["project_id"])

	project, err := c.ProjectsService.GetByID(ctx, int32(projectID))
	if err != nil {
		return NotFound(fmt.Sprintf("Project not found: %v", err))
	}

	jobs, _, err := c.PredictionJobService.ListPredictionJobs(ctx, project, &query, false)
	if err != nil {
		return InternalServerError(fmt.Sprintf("Error listing prediction jobs: %v", err))
	}

	return Ok(jobs)
}

// ListAllInProject lists all prediction jobs of a project, with pagination
func (c *PredictionJobController) ListAllInProjectByPage(r *http.Request, vars map[string]string, body interface{}) *Response {
	ctx := r.Context()

	var query service.ListPredictionJobQuery
	err := decoder.Decode(&query, r.URL.Query())
	if err != nil {
		return BadRequest(fmt.Sprintf("Bad query %s", r.URL.Query()))
	}

	projectID, _ := models.ParseID(vars["project_id"])

	project, err := c.ProjectsService.GetByID(ctx, int32(projectID))
	if err != nil {
		return NotFound(fmt.Sprintf("Project not found: %v", err))
	}

	jobs, paging, err := c.PredictionJobService.ListPredictionJobs(ctx, project, &query, true)
	if err != nil {
		return InternalServerError(fmt.Sprintf("Error listing prediction jobs: %v", err))
	}

	return Ok(ListJobsPaginatedResponse{
		Results: jobs,
		Paging:  *paging,
	})
}
