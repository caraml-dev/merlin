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

	"github.com/caraml-dev/merlin/models"
)

// AlertsController controls alerts API.
type AlertsController struct {
	*AppContext
}

// ListTeams lists registered teams for alerts notification.
func (c *AlertsController) ListTeams(r *http.Request, vars map[string]string, _ interface{}) *Response {
	teams, err := c.ModelEndpointAlertService.ListTeams()
	if err != nil {
		return InternalServerError(fmt.Sprintf("Error listing teams: %v", err))
	}

	return Ok(teams)
}

// ListModelEndpointAlerts lists alerts for model endpoints.
func (c *AlertsController) ListModelEndpointAlerts(r *http.Request, vars map[string]string, _ interface{}) *Response {
	modelID, _ := models.ParseID(vars["model_id"])

	modelEndpointAlerts, err := c.ModelEndpointAlertService.ListModelAlerts(modelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFound(fmt.Sprintf("Model endpoint alert not found: %v", err))
		}
		return InternalServerError(fmt.Sprintf("Error listing alerts for model: %v", err))
	}

	return Ok(modelEndpointAlerts)
}

// GetModelEndpointAlert gets alert for given model endpoint.
func (c *AlertsController) GetModelEndpointAlert(r *http.Request, vars map[string]string, _ interface{}) *Response {
	modelID, _ := models.ParseID(vars["model_id"])
	modelEndpointID, _ := models.ParseID(vars["model_endpoint_id"])

	modelEndpointAlert, err := c.ModelEndpointAlertService.GetModelEndpointAlert(modelID, modelEndpointID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFound(fmt.Sprintf("Model endpoint alert not found: %v", err))
		}
		return InternalServerError(fmt.Sprintf("Error getting alert for model endpoint: %v", err))
	}

	return Ok(modelEndpointAlert)
}

// CreateModelEndpointAlert creates alert for given model endpoint.
func (c *AlertsController) CreateModelEndpointAlert(r *http.Request, vars map[string]string, body interface{}) *Response {
	ctx := r.Context()

	user := vars["user"]
	modelID, _ := models.ParseID(vars["model_id"])
	modelEndpointID, _ := models.ParseID(vars["model_endpoint_id"])

	alert, ok := body.(*models.ModelEndpointAlert)
	if !ok {
		return BadRequest("Unable to parse body as model endpoint alert")
	}

	model, err := c.ModelsService.FindByID(ctx, modelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFound(fmt.Sprintf("Model not found: %v", err))
		}
		return InternalServerError(fmt.Sprintf("Error getting model: %v", err))
	}

	alert.Model = model

	modelEndpoint, err := c.ModelEndpointsService.FindByID(ctx, modelEndpointID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFound(fmt.Sprintf("Model endpoint not found: %v", err))
		}
		return InternalServerError(fmt.Sprintf("Error getting model endpoint: %v", err))
	}
	alert.ModelEndpoint = modelEndpoint

	alert, err = c.ModelEndpointAlertService.CreateModelEndpointAlert(user, alert)
	if err != nil {
		return InternalServerError(fmt.Sprintf("Error creating alert: %v", err))
	}

	return Created(alert)
}

// UpdateModelEndpointAlert updates model endpoint alert.
func (c *AlertsController) UpdateModelEndpointAlert(r *http.Request, vars map[string]string, body interface{}) *Response {
	ctx := r.Context()

	user := vars["user"]
	modelID, _ := models.ParseID(vars["model_id"])
	modelEndpointID, _ := models.ParseID(vars["model_endpoint_id"])

	newAlert, ok := body.(*models.ModelEndpointAlert)
	if !ok {
		return BadRequest("Unable to parse body as model endpoint alert")
	}

	model, err := c.ModelsService.FindByID(ctx, modelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFound(fmt.Sprintf("Model not found: %v", err))
		}
		return InternalServerError(fmt.Sprintf("Error getting model: %v", err))
	}

	oldAlert, err := c.ModelEndpointAlertService.GetModelEndpointAlert(modelID, modelEndpointID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFound(fmt.Sprintf("Model endpoint alert not found: %v", err))
		}
		return InternalServerError(fmt.Sprintf("Error getting alert for model endpoint: %v", err))
	}

	newAlert.ID = oldAlert.ID
	newAlert.Model = model
	newAlert.ModelEndpoint = oldAlert.ModelEndpoint

	newAlert, err = c.ModelEndpointAlertService.UpdateModelEndpointAlert(user, newAlert)
	if err != nil {
		return InternalServerError(fmt.Sprintf("Error updating model endpoint alert: %v", err))
	}

	return Created(newAlert)
}
