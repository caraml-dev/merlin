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
	"context"
	"fmt"
	"net/http"

	"github.com/caraml-dev/merlin/log"

	"github.com/jinzhu/gorm"

	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/service"
	"github.com/caraml-dev/merlin/utils"
)

const DEFAULT_PYTHON_VERSION = "3.7.*"

type VersionsController struct {
	*AppContext
}

func (c *VersionsController) GetVersion(r *http.Request, vars map[string]string, _ interface{}) *Response {
	ctx := r.Context()

	modelID, _ := models.ParseID(vars["model_id"])
	versionID, _ := models.ParseID(vars["version_id"])

	v, err := c.VersionsService.FindByID(ctx, modelID, versionID, c.MonitoringConfig)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return NotFound(fmt.Sprintf("Model version not found: %v", err))
		}
		return InternalServerError(fmt.Sprintf("Error getting model version: %sv", err))
	}

	return Ok(v)
}

func (c *VersionsController) PatchVersion(r *http.Request, vars map[string]string, body interface{}) *Response {
	ctx := r.Context()

	modelID, _ := models.ParseID(vars["model_id"])
	versionID, _ := models.ParseID(vars["version_id"])

	v, err := c.VersionsService.FindByID(ctx, modelID, versionID, c.MonitoringConfig)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return NotFound(fmt.Sprintf("Model version not found: %v", err))
		}
		return InternalServerError(fmt.Sprintf("Error getting model version: %v", err))
	}

	versionPatch, ok := body.(*models.VersionPatch)
	if !ok {
		return InternalServerError("Unable to parse request body")
	}

	if err := v.Patch(versionPatch); err != nil {
		return BadRequest(fmt.Sprintf("Error validating version: %v", err))
	}

	patchedVersion, err := c.VersionsService.Save(ctx, v, c.MonitoringConfig)
	if err != nil {
		return InternalServerError(fmt.Sprintf("Error patching model version: %v", err))
	}

	return Ok(patchedVersion)
}

func (c *VersionsController) ListVersions(r *http.Request, vars map[string]string, _ interface{}) *Response {
	ctx := r.Context()

	var query service.VersionQuery
	if err := decoder.Decode(&query, r.URL.Query()); err != nil {
		return BadRequest(fmt.Sprintf("Unable to parse query string: %v", err))
	}

	modelID, _ := models.ParseID(vars["model_id"])
	versions, nextCursor, err := c.VersionsService.ListVersions(ctx, modelID, c.MonitoringConfig, query)
	if err != nil {
		return InternalServerError(fmt.Sprintf("Error listing versions for model: %v", err))
	}

	responseHeaders := make(map[string]string)
	if nextCursor != "" {
		responseHeaders["Next-Cursor"] = nextCursor
	}

	return OkWithHeaders(versions, responseHeaders)
}

func (c *VersionsController) CreateVersion(r *http.Request, vars map[string]string, body interface{}) *Response {
	ctx := r.Context()

	versionPost, ok := body.(*models.VersionPost)
	if !ok {
		return BadRequest("Unable to parse request body")
	}
	validLabel := validateLabels(versionPost.Labels)
	if !validLabel {
		return BadRequest("Valid label key/values must be 63 characters or less and must be empty or begin and end with an alphanumeric character ([a-z0-9A-Z]) with dashes (-), underscores (_), dots (.), and alphanumerics between.")
	}

	if versionPost.PythonVersion == "" { // if using older sdk, this will be blank
		versionPost.PythonVersion = DEFAULT_PYTHON_VERSION
	}

	modelID, _ := models.ParseID(vars["model_id"])

	model, err := c.ModelsService.FindByID(ctx, modelID)
	if err != nil {
		return NotFound(fmt.Sprintf("Model not found: %v", err))
	}

	run, err := c.MlflowClient.CreateRun(fmt.Sprintf("%d", model.ExperimentID))
	if err != nil {
		return InternalServerError(fmt.Sprintf("Unable to create mlflow run: %v", err))
	}

	version := &models.Version{
		ModelID:       modelID,
		RunID:         run.Info.RunID,
		ArtifactURI:   run.Info.ArtifactURI,
		Labels:        versionPost.Labels,
		PythonVersion: versionPost.PythonVersion,
	}

	version, _ = c.VersionsService.Save(ctx, version, c.MonitoringConfig)
	return Created(version)
}

func (c *VersionsController) DeleteVersion(r *http.Request, vars map[string]string, _ interface{}) *Response {
	ctx := r.Context()

	modelID, err := models.ParseID(vars["model_id"])
	if err != nil {
		return BadRequest("Unable to parse model_id")
	}
	versionID, err := models.ParseID(vars["version_id"])
	if err != nil {
		return BadRequest("Unable to parse version_id")
	}

	model, version, err := c.getModelAndVersion(ctx, modelID, versionID)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return NotFound(fmt.Sprintf("Model / version not found: %v", err))
		}
		return InternalServerError(fmt.Sprintf("Error getting model / version: %v", err))
	}

	// handle for pyfunc v2, since the prediction job feature is only available for pyfunc_v2 model
	if model.Type == "pyfunc_v2" {
		// check active prediction job
		// if there are any active prediction job using the model version, deletion of the model version are prohibited
		inactiveJobs, errResponse := c.getInactivePredictionJobsForDeletion(ctx, model, version)
		if errResponse != nil {
			return errResponse
		}

		// delete inactive prediction job
		errResponse = c.deletePredictionJobs(ctx, inactiveJobs, model, version)
		if errResponse != nil {
			return errResponse
		}
	}

	// check active endpoints for all model
	// if there are any active endpoint using the model version, deletion of the model version are prohibited
	inactiveEndpoints, errResponse := c.getInactiveEndpointsForDeletion(ctx, model, version)
	if errResponse != nil {
		return errResponse
	}
	// delete inactive endpoint
	errResponse = c.deleteVersionEndpoints(inactiveEndpoints, version)
	if errResponse != nil {
		return errResponse
	}

	// delete mlflow run and artifact
	err = c.MlflowDeleteService.DeleteRun(ctx, version.RunID, version.ArtifactURI, true)
	if err != nil {
		return InternalServerError(fmt.Sprintf("Delete mlflow run failed: %s", err.Error()))
	}

	// delete model version from db
	err = c.VersionsService.Delete(version)
	if err != nil {
		return InternalServerError(fmt.Sprintf("Delete model version failed: %s", err.Error()))
	}

	return Ok(versionID)
}

func (c *VersionsController) getInactivePredictionJobsForDeletion(ctx context.Context, model *models.Model, version *models.Version) ([]*models.PredictionJob, *Response) {
	jobQuery := &service.ListPredictionJobQuery{
		ModelID:   model.ID,
		VersionID: version.ID,
	}

	jobs, err := c.PredictionJobService.ListPredictionJobs(ctx, model.Project, jobQuery)
	if err != nil {
		return nil, InternalServerError("Failed listing prediction job")
	}

	for _, item := range jobs {
		if item.Status == models.JobPending || item.Status == models.JobRunning {
			return nil, BadRequest("There are active prediction job that still using this model version")
		}
	}

	return jobs, nil
}

func (c *VersionsController) getInactiveEndpointsForDeletion(ctx context.Context, model *models.Model, version *models.Version) ([]*models.VersionEndpoint, *Response) {

	endpoints, err := c.EndpointsService.ListEndpoints(ctx, model, version)
	if err != nil {
		return nil, InternalServerError("Failed listing model version endpoint")
	}

	for _, item := range endpoints {
		if item.Status != models.EndpointTerminated && item.Status != models.EndpointFailed {
			return nil, BadRequest("There are active endpoint that still using this model version")
		}

	return endpoints, nil
}

func (c *VersionsController) deleteVersionEndpoints(endpoints []*models.VersionEndpoint, version *models.Version) *Response {
	for _, item := range endpoints {
		err := c.EndpointsService.DeleteEndpoint(version, item)
		if err != nil {
			return InternalServerError(fmt.Sprintf("Failed to delete endpoint: %s", err))
		}
	}
	return nil
}

func (c *VersionsController) deletePredictionJobs(ctx context.Context, jobs []*models.PredictionJob, model *models.Model, version *models.Version) *Response {
	for _, item := range jobs {
		_, err := c.PredictionJobService.StopPredictionJob(ctx, item.Environment, model, version, item.ID)
		if err != nil {
			return InternalServerError(fmt.Sprintf("Failed stopping prediction job: %s", err))
		}
	}
	return nil
}
func (c *VersionsController) checkActivePredictionJobs(ctx context.Context, model *models.Model, version *models.Version) (int, []*models.PredictionJob, error) {
	jobQuery := &service.ListPredictionJobQuery{
		ModelID:   model.ID,
		VersionID: version.ID,
	}

	jobs, err := c.PredictionJobService.ListPredictionJobs(ctx, model.Project, jobQuery)
	if err != nil {
		log.Errorf("failed to list all prediction job for model %s version %s: %v", model.Name, version.ID, err)
		return nil, InternalServerError("Failed listing prediction job")
	}

	for _, item := range jobs {
		if item.Status == models.JobPending || item.Status == models.JobRunning {
			return nil, BadRequest("There are active prediction job that still using this model version")
		}
	}

	return jobs, nil
}

func (c *VersionsController) getInactiveEndpointsForDeletion(ctx context.Context, model *models.Model, version *models.Version) ([]*models.VersionEndpoint, *Response) {

	endpoints, err := c.EndpointsService.ListEndpoints(ctx, model, version)
	if err != nil {
		log.Errorf("failed to list all endpoint for model %s version %s: %v", model.Name, version.ID, err)
		return nil, InternalServerError("Failed listing model version endpoint")
	}

	for _, item := range endpoints {
		if item.Status != models.EndpointTerminated && item.Status != models.EndpointFailed {
			return nil, BadRequest("There are active endpoint that still using this model version")
		}
	}

	return endpoints, nil
}

func (c *VersionsController) deleteVersionEndpoints(endpoints []*models.VersionEndpoint, version *models.Version) *Response {
	for _, item := range endpoints {
		err := c.EndpointsService.DeleteEndpoint(version, item)
		if err != nil {
			log.Errorf("failed to undeploy endpoint job %v", err)
			return InternalServerError(fmt.Sprintf("Failed to delete endpoint: %s", err))
		}
	}
	return nil
}

func (c *VersionsController) deletePredictionJobs(ctx context.Context, jobs []*models.PredictionJob, model *models.Model, version *models.Version) *Response {
	for _, item := range jobs {
		_, err := c.PredictionJobService.StopPredictionJob(ctx, item.Environment, model, version, item.ID)
		if err != nil {
			log.Errorf("failed to stop prediction job %v", err)
			return InternalServerError(fmt.Sprintf("Failed stopping prediction job: %s", err))
		}
	}
	return nil
}

func validateLabels(labels models.KV) bool {
	for key, element := range labels {
		value, ok := element.(string)
		if !ok {
			return false
		}

		if err := utils.IsValidLabel(key); err != nil {
			return false
		}

		if err := utils.IsValidLabel(value); err != nil {
			return false
		}
	}
	return true
}
