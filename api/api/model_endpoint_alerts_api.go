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
)

// AlertsController controls alerts API.
type AlertsController struct {
	*AppContext
}

// ListTeams lists registered teams for alerts notification.
func (c *AlertsController) ListTeams(r *http.Request, vars map[string]string, _ interface{}) *ApiResponse {
	teams, err := c.ModelEndpointAlertService.ListTeams()
	if err != nil {
		log.Errorf("ListTeams: %s", err)
		return InternalServerError("ListTeams: Error while getting list of teams for alert notification")
	}

	return Ok(teams)
}

// ListModelEndpointAlerts lists alerts for model endpoints.
func (c *AlertsController) ListModelEndpointAlerts(r *http.Request, vars map[string]string, _ interface{}) *ApiResponse {
	modelId, _ := models.ParseId(vars["model_id"])

	modelEndpointAlerts, err := c.ModelEndpointAlertService.ListModelAlerts(modelId)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return NotFound(fmt.Sprintf("ListModelAlerts: Alerts for Model ID %s not found", modelId))
		}
		return InternalServerError(fmt.Sprintf("ListModelAlerts: Error while getting alerts for Model ID %s", modelId))
	}

	return Ok(modelEndpointAlerts)
}

// GetModelEndpointAlert gets alert for given model endpoint.
func (c *AlertsController) GetModelEndpointAlert(r *http.Request, vars map[string]string, _ interface{}) *ApiResponse {
	modelId, _ := models.ParseId(vars["model_id"])
	modelEndpointId, _ := models.ParseId(vars["model_endpoint_id"])

	modelEndpointAlert, err := c.ModelEndpointAlertService.GetModelEndpointAlert(modelId, modelEndpointId)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return NotFound(fmt.Sprintf("GetModelEndpointAlert: Alert for model endpoint with id %s not found", modelEndpointId))
		}
		return InternalServerError(fmt.Sprintf("GetModelEndpointAlert: Error while getting alert for model endpoint with id %s", modelEndpointId))
	}

	return Ok(modelEndpointAlert)
}

// CreateModelEndpointAlert creates alert for given model endpoint.
func (c *AlertsController) CreateModelEndpointAlert(r *http.Request, vars map[string]string, body interface{}) *ApiResponse {
	ctx := r.Context()

	user := vars["user"]
	modelId, _ := models.ParseId(vars["model_id"])
	modelEndpointId, _ := models.ParseId(vars["model_endpoint_id"])

	alert, ok := body.(*models.ModelEndpointAlert)
	if !ok {
		return BadRequest("Unable to parse body as model endpoint alert")
	}

	model, err := c.ModelsService.FindById(ctx, modelId)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return NotFound(fmt.Sprintf("Model with id %s not found", modelId))
		}
		return InternalServerError(fmt.Sprintf("Error while getting model with id %s", modelId))
	}

	alert.Model = model

	modelEndpoint, err := c.ModelEndpointsService.FindById(ctx, modelEndpointId)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return NotFound(fmt.Sprintf("Model endpoint with id %s not found", modelEndpointId))
		}
		return InternalServerError(fmt.Sprintf("Error while getting model endpoint with id %s", modelEndpointId))
	}
	alert.ModelEndpoint = modelEndpoint

	alert, err = c.ModelEndpointAlertService.CreateModelEndpointAlert(user, alert)
	if err != nil {
		log.Errorf("CreateModelEndpointAlert: %s", err)
		return InternalServerError(fmt.Sprintf("Error while creating model endpoint alert for Model %s, Endpoint %s", modelId, modelEndpointId))
	}

	return Created(alert)
}

func (c *AlertsController) UpdateModelEndpointAlert(r *http.Request, vars map[string]string, body interface{}) *ApiResponse {
	ctx := r.Context()

	user := vars["user"]
	modelId, _ := models.ParseId(vars["model_id"])
	modelEndpointId, _ := models.ParseId(vars["model_endpoint_id"])

	newAlert, ok := body.(*models.ModelEndpointAlert)
	if !ok {
		return BadRequest("Unable to parse body as model endpoint alert")
	}

	model, err := c.ModelsService.FindById(ctx, modelId)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return NotFound(fmt.Sprintf("Model with id %s not found", modelId))
		}
		return InternalServerError(fmt.Sprintf("Error while getting model with id %s", modelId))
	}

	oldAlert, err := c.ModelEndpointAlertService.GetModelEndpointAlert(modelId, modelEndpointId)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return NotFound(fmt.Sprintf("Alert for Model ID %s and Model Endpoint ID %s not found", modelId, modelEndpointId))
		}
		return InternalServerError(fmt.Sprintf("Error while getting alert for Model ID %s and Model Endpoint ID %s", modelId, modelEndpointId))
	}

	newAlert.Id = oldAlert.Id
	newAlert.Model = model
	newAlert.ModelEndpoint = oldAlert.ModelEndpoint

	newAlert, err = c.ModelEndpointAlertService.UpdateModelEndpointAlert(user, newAlert)
	if err != nil {
		log.Errorf("UpdateModelEndpointAlert: %s", err)
		return InternalServerError(fmt.Sprintf("Error while updating model endpoint alert for Model %s, Endpoint %s", modelId, modelEndpointId))
	}

	return Created(newAlert)
}
