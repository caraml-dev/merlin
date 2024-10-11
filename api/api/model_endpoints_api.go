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

	"github.com/caraml-dev/merlin/log"
	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/webhook"
	"gorm.io/gorm"
)

// ModelEndpointsController controls model endpoints API
type ModelEndpointsController struct {
	*AppContext
}

// ListModelEndpointInProject list all model endpoints under a project
func (c *ModelEndpointsController) ListModelEndpointInProject(r *http.Request, vars map[string]string, _ interface{}) *Response {
	ctx := r.Context()

	projectID, _ := models.ParseID(vars["project_id"])
	region := vars["region"]

	modelEndpoints, err := c.ModelEndpointsService.ListModelEndpointsInProject(ctx, projectID, region)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFound(fmt.Sprintf("Model endpoints not found: %v", err))
		}
		return InternalServerError(fmt.Sprintf("Error listing model endpoints: %v", err))
	}

	return Ok(modelEndpoints)
}

// ListModelEndpoints lists all model endpoints under a model.
func (c *ModelEndpointsController) ListModelEndpoints(r *http.Request, vars map[string]string, body interface{}) *Response {
	ctx := r.Context()

	modelID, _ := models.ParseID(vars["model_id"])
	modelEndpoints, err := c.ModelEndpointsService.ListModelEndpoints(ctx, modelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFound(fmt.Sprintf("Model endpoints not found: %v", err))
		}
		return InternalServerError(fmt.Sprintf("Error listing model endpoints: %v", err))
	}
	return Ok(modelEndpoints)
}

// GetModelEndpoint get model endpoint given an ID.
func (c *ModelEndpointsController) GetModelEndpoint(r *http.Request, vars map[string]string, _ interface{}) *Response {
	ctx := r.Context()

	modelEndpointID, _ := models.ParseID(vars["model_endpoint_id"])
	modelEndpoint, err := c.ModelEndpointsService.FindByID(ctx, modelEndpointID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFound(fmt.Sprintf("Model endpoint not found: %v", err))
		}
		return InternalServerError(fmt.Sprintf("Error getting model endpoint: %v", err))
	}

	return Ok(modelEndpoint)
}

// CreateModelEndpoint does following actions:
// 1. Deploy an Istio's VirtualService specifaction that maps from ModelEndpoint's rule
// 2. Insert a record on `model_endpoints` table
// 3. Update `endpoint_id` column on associated `model` record
func (c *ModelEndpointsController) CreateModelEndpoint(r *http.Request, vars map[string]string, body interface{}) *Response {
	ctx := r.Context()

	modelID, _ := models.ParseID(vars["model_id"])
	model, err := c.ModelsService.FindByID(ctx, modelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFound(fmt.Sprintf("Model not found: %v", err))
		}
		return InternalServerError(fmt.Sprintf("Error getting model: %v", err))
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
			return InternalServerError(fmt.Sprintf("Default environment not found, please specify one: %v", err))
		}
		endpoint.EnvironmentName = env.Name
	} else {
		// Check environment exists
		env, err = c.AppContext.EnvironmentService.GetEnvironment(endpoint.EnvironmentName)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return NotFound(fmt.Sprintf("Environment not found: %v", err))
			}
			return InternalServerError(fmt.Sprintf("Error getting environment: %v", err))
		}
	}
	endpoint.Environment = env

	// Deploy model endpoint as Istio's VirtualService
	endpoint, err = c.ModelEndpointsService.DeployEndpoint(ctx, model, endpoint)
	if err != nil {
		return InternalServerError(fmt.Sprintf("Error creating model endpoint: %v", err))
	}

	// trigger webhook call
	if err = c.Webhook.TriggerWebhooks(ctx, webhook.OnModelEndpointCreated, webhook.SetBody(endpoint)); err != nil {
		log.Warnf("unable to invoke webhook for event type: %s, model: %s, endpoint: %d, error: %v", webhook.OnModelEndpointCreated, model.Name, endpoint.ID, err)
	}

	// Success. Return endpoint as response.
	return Created(endpoint)
}

// UpdateModelEndpoint updates model endpoint.
func (c *ModelEndpointsController) UpdateModelEndpoint(r *http.Request, vars map[string]string, body interface{}) *Response {
	ctx := r.Context()

	modelID, _ := models.ParseID(vars["model_id"])
	modelEndpointID, _ := models.ParseID(vars["model_endpoint_id"])
	model, err := c.ModelsService.FindByID(ctx, modelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFound(fmt.Sprintf("Model not found: %v", err))
		}
		return InternalServerError(fmt.Sprintf("Error getting model: %v", err))
	}

	newEndpoint, ok := body.(*models.ModelEndpoint)
	if !ok {
		return BadRequest("Invalid request body")
	}

	// Check environment exists or assign default environment
	var env *models.Environment
	if newEndpoint.EnvironmentName == "" {
		// Use default environment if not specified
		env, err = c.EnvironmentService.GetDefaultEnvironment()
		if err != nil {
			return InternalServerError(fmt.Sprintf("Default environment not found, please specify one: %v", err))
		}
		newEndpoint.EnvironmentName = env.Name
	} else {
		// Check environment exists
		env, err = c.AppContext.EnvironmentService.GetEnvironment(newEndpoint.EnvironmentName)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return NotFound(fmt.Sprintf("Environment not found: %v", err))
			}
			return InternalServerError(fmt.Sprintf("Error getting environment: %v", err))

		}
	}
	newEndpoint.Environment = env

	currentEndpoint, err := c.ModelEndpointsService.FindByID(ctx, modelEndpointID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFound(fmt.Sprintf("Model endpoint not found: %v", err))
		}
		return InternalServerError(fmt.Sprintf("Error getting model endpoint: %v", err))
	}

	if newEndpoint.ID != currentEndpoint.ID {
		return BadRequest("Invalid request model endpoint id")
	}

	if currentEndpoint.Status == models.EndpointTerminated {
		newEndpoint, err = c.ModelEndpointsService.DeployEndpoint(ctx, model, newEndpoint)
	} else {
		// Update model endpoint's VirtualService
		newEndpoint, err = c.ModelEndpointsService.UpdateEndpoint(ctx, model, currentEndpoint, newEndpoint)
	}

	if err != nil {
		return InternalServerError(fmt.Sprintf("Error updating model endpoint: %v", err))
	}

	// trigger webhook call
	if err = c.Webhook.TriggerWebhooks(ctx, webhook.OnModelEndpointUpdated, webhook.SetBody(newEndpoint)); err != nil {
		log.Warnf("unable to invoke webhook for event type: %s, model: %s, error: %v", webhook.OnModelEndpointUpdated, model.Name, err)
	}

	return Ok(newEndpoint)
}

// DeleteModelEndpoint stops model endpoint for serving.
// To be more precise, it will do the following:
// 1. Delete the corresponding Istio's VirtualService.
// 2. Update the model endpoint database. Update the model endpoint status to "terminated".
// 3. Update the version endpoint database. Update the version endpoint status from serving to "running".
func (c *ModelEndpointsController) DeleteModelEndpoint(r *http.Request, vars map[string]string, _ interface{}) *Response {
	ctx := r.Context()

	modelID, _ := models.ParseID(vars["model_id"])
	modelEndpointID, _ := models.ParseID(vars["model_endpoint_id"])

	model, err := c.ModelsService.FindByID(ctx, modelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFound(fmt.Sprintf("Model not found: %v", err))
		}
		return InternalServerError(fmt.Sprintf("Error getting model: %v", err))
	}

	modelEndpoint, err := c.ModelEndpointsService.FindByID(ctx, modelEndpointID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFound(fmt.Sprintf("Model endpoint not found: %v", err))
		}
		return InternalServerError(fmt.Sprintf("Error getting model endpoint: %v", err))
	}

	_, err = c.ModelEndpointsService.UndeployEndpoint(ctx, model, modelEndpoint)
	if err != nil {
		return InternalServerError(fmt.Sprintf("Error deleting model endpoint: %v", err))
	}

	// trigger webhook call
	if err = c.Webhook.TriggerWebhooks(ctx, webhook.OnModelEndpointDeleted, webhook.SetBody(modelEndpoint)); err != nil {
		log.Warnf("unable to invoke webhook for event type: %s, model: %s, error: %v", webhook.OnModelEndpointDeleted, model.Name, err)
	}

	return Ok(nil)
}
