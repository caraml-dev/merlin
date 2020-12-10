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
	"database/sql"
	"fmt"
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"

	"github.com/gojek/merlin/log"
	"github.com/gojek/merlin/models"
)

// ModelEndpointsController controls model endpoints API
type ModelEndpointsController struct {
	*AppContext
}

// ListModelEndpointInProject list all model endpoints under a project
func (c *ModelEndpointsController) ListModelEndpointInProject(r *http.Request, vars map[string]string, _ interface{}) *APIResponse {
	ctx := r.Context()

	projectID, _ := models.ParseID(vars["project_id"])
	region := vars["region"]

	modelEndpoints, err := c.ModelEndpointsService.ListModelEndpointsInProject(ctx, projectID, region)
	if err != nil {
		log.Errorf("Error finding Model Endpoints for Project ID %s, reason: %v", projectID, err)

		if gorm.IsRecordNotFoundError(err) {
			return NotFound(fmt.Sprintf("Model Endpoints for Project ID %s not found", projectID))
		}

		return InternalServerError(fmt.Sprintf("Error while getting Model Endpoints for Project ID %s", projectID))
	}

	return Ok(modelEndpoints)
}

// ListModelEndpoints lists all model endpoints under a model.
func (c *ModelEndpointsController) ListModelEndpoints(r *http.Request, vars map[string]string, body interface{}) *APIResponse {
	ctx := r.Context()

	modelID, _ := models.ParseID(vars["model_id"])
	modelEndpoints, err := c.ModelEndpointsService.ListModelEndpoints(ctx, modelID)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return NotFound(fmt.Sprintf("Model Endpoints for Model ID %s not found", modelID))
		}
		return InternalServerError(fmt.Sprintf("Error while getting Model Endpoints for Model ID %s", modelID))
	}
	return Ok(modelEndpoints)
}

// GetModelEndpoint get model endpoint given an ID.
func (c *ModelEndpointsController) GetModelEndpoint(r *http.Request, vars map[string]string, _ interface{}) *APIResponse {
	ctx := r.Context()

	modelEndpointID, _ := models.ParseID(vars["model_endpoint_id"])
	modelEndpoint, err := c.ModelEndpointsService.FindByID(ctx, modelEndpointID)
	if err != nil {
		log.Errorf("Error finding model endpoint with id %s, reason: %v", modelEndpointID, err)

		if gorm.IsRecordNotFoundError(err) {
			return NotFound(fmt.Sprintf("Model endpoint with id %s not found", modelEndpointID))
		}

		return InternalServerError(fmt.Sprintf("Error while getting model endpoint with id %s", modelEndpointID))
	}

	return Ok(modelEndpoint)
}

// CreateModelEndpoint does following actions:
// 1. Deploy an Istio's VirtualService specifaction that maps from ModelEndpoint's rule
// 2. Insert a record on `model_endpoints` table
// 3. Update `endpoint_id` column on associated `model` record
func (c *ModelEndpointsController) CreateModelEndpoint(r *http.Request, vars map[string]string, body interface{}) *APIResponse {
	ctx := r.Context()

	modelID, _ := models.ParseID(vars["model_id"])
	model, err := c.ModelsService.FindByID(ctx, modelID)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return NotFound(fmt.Sprintf("Model ID %s not found", modelID))
		}

		return InternalServerError(fmt.Sprintf("Error while getting Model ID %s", modelID))
	}

	endpoint, ok := body.(*models.ModelEndpoint)
	if !ok {
		return BadRequest("Invalid request body")
	}

	var env *models.Environment
	if endpoint.EnvironmentName == "" {
		// Use default environment if not specified
		env, err = c.EnvironmentService.GetDefaultEnvironment()
		if err != nil {
			return InternalServerError("Default environment not found, please specify one")
		}
		endpoint.EnvironmentName = env.Name
	} else {
		// Check environment exists
		env, err = c.AppContext.EnvironmentService.GetEnvironment(endpoint.EnvironmentName)
		if err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				return InternalServerError("Unable to find the specified environment")
			}

			return NotFound(fmt.Sprintf("Environment not found: %s", endpoint.EnvironmentName))
		}
	}
	endpoint.Environment = env

	// Fetch version endpoint as model endpoint destination
	endpoint, err = c.assignVersionEndpoint(ctx, endpoint)
	if err != nil {
		return BadRequest(fmt.Sprintf("Invalid version endpoints destination: %s", err))
	}

	// Deploy model endpoint as Istio's VirtualService
	endpoint, err = c.ModelEndpointsService.DeployEndpoint(ctx, model, endpoint)
	if err != nil {
		return InternalServerError(fmt.Sprintf("Unable to create model endpoint: %s", err.Error()))
	}

	// Update DB
	err = c.saveDB(ctx, endpoint, nil)
	if err != nil {
		return InternalServerError(fmt.Sprintf("Unable to update model data: %s", err.Error()))
	}

	// Success. Return endpoint as response.
	return Created(endpoint)
}

// UpdateModelEndpoint updates model endpoint.
func (c *ModelEndpointsController) UpdateModelEndpoint(r *http.Request, vars map[string]string, body interface{}) *APIResponse {
	ctx := r.Context()

	modelID, _ := models.ParseID(vars["model_id"])
	modelEndpointID, _ := models.ParseID(vars["model_endpoint_id"])
	model, err := c.ModelsService.FindByID(ctx, modelID)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return NotFound(fmt.Sprintf("Model ID %s not found", modelID))
		}

		return InternalServerError(fmt.Sprintf("Error while getting Model ID %s", modelID))
	}

	newEndpoint, ok := body.(*models.ModelEndpoint)
	if !ok {
		return BadRequest("Invalid request body")
	}

	// Check environment exists or assign default environment
	env, err := c.AppContext.EnvironmentService.GetEnvironment(newEndpoint.EnvironmentName)
	if err != nil {
		return InternalServerError("Unable to find the specified environment")
	}
	newEndpoint.EnvironmentName = env.Name
	newEndpoint.Environment = env

	// Fetch version endpoint as model endpoint destination
	newEndpoint, err = c.assignVersionEndpoint(ctx, newEndpoint)
	if err != nil {
		return BadRequest(fmt.Sprintf("Invalid version endpoints destination: %s", err))
	}

	currentEndpoint, err := c.ModelEndpointsService.FindByID(ctx, modelEndpointID)
	if err != nil {
		return NotFound(fmt.Sprintf("Model Endpoint with given `model_endpoint_id: %d` not found", modelEndpointID))
	}

	if newEndpoint.ID != currentEndpoint.ID {
		return BadRequest("Invalid request model endpoint id")
	}

	if currentEndpoint.Status == models.EndpointTerminated {
		newEndpoint, err = c.ModelEndpointsService.DeployEndpoint(ctx, model, newEndpoint)
	} else {
		// Update model endpoint's VirtualService
		newEndpoint, err = c.ModelEndpointsService.UpdateEndpoint(ctx, model, newEndpoint)
	}
	if err != nil {
		return InternalServerError(fmt.Sprintf("Unable to update model endpoint: %s", err.Error()))
	}

	// Update DB
	err = c.saveDB(ctx, newEndpoint, currentEndpoint)
	if err != nil {
		return InternalServerError(fmt.Sprintf("Unable to update model data: %s", err.Error()))
	}

	return Ok(newEndpoint)
}

func (c *ModelEndpointsController) saveDB(ctx context.Context, newModelEndpoint, prevModelEndpoint *models.ModelEndpoint) error {

	tx := c.DB.BeginTx(ctx, &sql.TxOptions{})
	defer tx.RollbackUnlessCommitted()

	var err error

	err = tx.Save(newModelEndpoint).Error
	if err != nil {
		return errors.Wrap(err, "Failed to save model endpoint")
	}

	// Update version and version endpoints from previous model endpoint
	if prevModelEndpoint != nil {
		for _, ruleDestination := range prevModelEndpoint.Rule.Destination {
			versionEndpoint, err := c.EndpointsService.FindByID(ruleDestination.VersionEndpointID)
			if err != nil {
				return err
			}

			versionEndpoint.Status = models.EndpointRunning
			err = tx.Save(versionEndpoint).Error
			if err != nil {
				return errors.Wrap(err, "Failed to save previous model version endpoint")
			}
		}
	}

	// Update version and version endpoints from new model endpoint
	for _, ruleDestination := range newModelEndpoint.Rule.Destination {
		versionEndpoint, err := c.EndpointsService.FindByID(ruleDestination.VersionEndpointID)
		if err != nil {
			return err
		}

		versionEndpoint.Status = models.EndpointServing
		if newModelEndpoint.Status == models.EndpointTerminated {
			versionEndpoint.Status = models.EndpointRunning
		}

		err = tx.Save(versionEndpoint).Error
		if err != nil {
			return errors.Wrap(err, "Failed to save new model version endpoint")
		}
	}

	err = tx.Commit().Error
	if err != nil {
		return err
	}

	return nil
}

// assignVersionEndpoint fetches destination version endpoints from database and assign to model endpoint.
// assignVersionEndpoint validates version endpoint status and returns error if find no running version endpoint.
func (c *ModelEndpointsController) assignVersionEndpoint(ctx context.Context, endpoint *models.ModelEndpoint) (*models.ModelEndpoint, error) {
	for k := range endpoint.Rule.Destination {
		versionEndpointID := endpoint.Rule.Destination[k].VersionEndpointID

		versionEndpoint, err := c.EndpointsService.FindByID(versionEndpointID)
		if err != nil {
			return nil, fmt.Errorf("Version Endpoint with given `version_endpoint_id: %s` not found", versionEndpointID)
		}

		if !versionEndpoint.IsRunning() {
			return nil, fmt.Errorf("Version Endpoint %s is not running, but %s", versionEndpoint.ID, versionEndpoint.Status)
		}

		endpoint.Rule.Destination[k].VersionEndpoint = versionEndpoint
	}

	return endpoint, nil
}

// DeleteModelEndpoint stops model endpoint for serving.
// To be more precise, it will do the following:
// 1. Delete the corresponding Istio's VirtualService
// 2. Update the model endpoint database
//    - Update the model endpoint status to terminated
// 3. Update the version endpoint database
//    - Update the version endpoint status from serving to running
func (c *ModelEndpointsController) DeleteModelEndpoint(r *http.Request, vars map[string]string, _ interface{}) *APIResponse {
	ctx := r.Context()

	modelID, _ := models.ParseID(vars["model_id"])
	modelEndpointID, _ := models.ParseID(vars["model_endpoint_id"])

	model, err := c.ModelsService.FindByID(ctx, modelID)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return NotFound(fmt.Sprintf("Model ID %s not found", modelID))
		}

		return InternalServerError(fmt.Sprintf("Error while getting Model ID %s", modelID))
	}

	modelEndpoint, err := c.ModelEndpointsService.FindByID(ctx, modelEndpointID)
	if err != nil {
		log.Errorf("Error finding model endpoint with id %s, reason: %v", modelEndpointID, err)

		if gorm.IsRecordNotFoundError(err) {
			return NotFound(fmt.Sprintf("Model endpoint with id %s not found", modelEndpointID))
		}

		return InternalServerError(fmt.Sprintf("Error while getting model endpoint with id %s", modelEndpointID))
	}

	modelEndpoint, err = c.ModelEndpointsService.UndeployEndpoint(ctx, model, modelEndpoint)
	if err != nil {
		return InternalServerError(fmt.Sprintf("Unable to delete model endpoint: %s", err.Error()))
	}

	// Update DB
	err = c.saveDB(ctx, modelEndpoint, nil)
	if err != nil {
		return InternalServerError(fmt.Sprintf("Unable to update model data: %s", err.Error()))
	}

	return Ok(nil)
}
