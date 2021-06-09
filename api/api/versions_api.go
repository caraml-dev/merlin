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
	"regexp"

	"github.com/jinzhu/gorm"

	"github.com/gojek/merlin/log"
	"github.com/gojek/merlin/models"
	"github.com/gojek/merlin/service"
)

var (
	labelsCheck = regexp.MustCompile("^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$")
)

type VersionsController struct {
	*AppContext
}

func (c *VersionsController) GetVersion(r *http.Request, vars map[string]string, _ interface{}) *Response {
	ctx := r.Context()

	modelID, _ := models.ParseID(vars["model_id"])
	versionID, _ := models.ParseID(vars["version_id"])

	v, err := c.VersionsService.FindByID(ctx, modelID, versionID, c.MonitoringConfig)
	if err != nil {
		log.Errorf("error getting model version for given model %s version %s", modelID, versionID)
		if gorm.IsRecordNotFoundError(err) {
			return NotFound(fmt.Sprintf("Model version %s for version %s", modelID, versionID))
		}
		return InternalServerError(fmt.Sprintf("Error getting model version for given model %s version %s", modelID, versionID))
	}

	return Ok(v)
}

func (c *VersionsController) PatchVersion(r *http.Request, vars map[string]string, body interface{}) *Response {
	ctx := r.Context()

	modelID, _ := models.ParseID(vars["model_id"])
	versionID, _ := models.ParseID(vars["version_id"])

	v, err := c.VersionsService.FindByID(ctx, modelID, versionID, c.MonitoringConfig)
	if err != nil {
		log.Errorf("error getting model version for given model %s version %s", modelID, versionID)
		if gorm.IsRecordNotFoundError(err) {
			return NotFound(fmt.Sprintf("Model version %s for version %s", modelID, versionID))
		}
		return InternalServerError(fmt.Sprintf("Error getting model version for given model %s version %s", modelID, versionID))
	}

	versionPatch, ok := body.(*models.VersionPatch)
	if !ok {
		return InternalServerError("Unable to parse request body")
	}

	if err := v.Patch(versionPatch); err != nil {
		return BadRequest(err.Error())
	}

	patchedVersion, err := c.VersionsService.Save(ctx, v, c.MonitoringConfig)
	if err != nil {
		return InternalServerError(fmt.Sprintf("Error patching model version for given model %s version %s", modelID, versionID))
	}

	return Ok(patchedVersion)
}

func (c *VersionsController) ListVersions(r *http.Request, vars map[string]string, _ interface{}) *Response {
	ctx := r.Context()

	var query service.VersionQuery
	if err := decoder.Decode(&query, r.URL.Query()); err != nil {
		log.Errorf("Error while parsing query string %v", err)
		return BadRequest(fmt.Sprintf("Unable to parse query string: %s", err))
	}

	modelID, _ := models.ParseID(vars["model_id"])
	versions, nextCursor, err := c.VersionsService.ListVersions(ctx, modelID, c.MonitoringConfig, query)
	if err != nil {
		return InternalServerError(err.Error())
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

	modelID, _ := models.ParseID(vars["model_id"])

	model, err := c.ModelsService.FindByID(ctx, modelID)
	if err != nil {
		return NotFound(fmt.Sprintf("Model with given `model_id: %d` not found", modelID))
	}

	run, err := c.MlflowClient.CreateRun(fmt.Sprintf("%d", model.ExperimentID))
	if err != nil {
		return InternalServerError(fmt.Sprintf("Unable to create mlflow run: %s", err.Error()))
	}

	version := &models.Version{
		ModelID:     modelID,
		RunID:       run.Info.RunID,
		ArtifactURI: run.Info.ArtifactURI,
		Labels:      versionPost.Labels,
	}

	version, _ = c.VersionsService.Save(ctx, version, c.MonitoringConfig)
	return Created(version)
}

func validateLabels(labels models.KV) bool {
	for key, element := range labels {
		value, ok := element.(string)
		if !ok {
			return false
		}

		lengthNameCheckResult := len(key) < 64
		lengthValueCheckResult := len(value) < 64
		if !(lengthNameCheckResult && lengthValueCheckResult) {
			return false
		}

		labelNameCheckResult := labelsCheck.MatchString(key)
		labelValueCheckResult := labelsCheck.MatchString(value)
		if !(labelNameCheckResult && labelValueCheckResult) {
			return false
		}
	}
	return true
}
