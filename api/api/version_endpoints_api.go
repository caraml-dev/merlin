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
	"errors"
	"fmt"
	"net/http"

	merror "github.com/caraml-dev/merlin/pkg/errors"
	"github.com/caraml-dev/merlin/pkg/protocol"
	"github.com/feast-dev/feast/sdk/go/protos/feast/core"
	"github.com/google/uuid"
	"google.golang.org/protobuf/encoding/protojson"
	"gorm.io/gorm"

	"github.com/caraml-dev/merlin/config"
	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/pkg/transformer"
	"github.com/caraml-dev/merlin/pkg/transformer/feast"
	"github.com/caraml-dev/merlin/pkg/transformer/pipeline"
	"github.com/caraml-dev/merlin/pkg/transformer/spec"
)

type EndpointsController struct {
	*AppContext
}

// ListEndpoint list all endpoints created from certain model version
func (c *EndpointsController) ListEndpoint(r *http.Request, vars map[string]string, _ interface{}) *Response {
	ctx := r.Context()

	modelID, _ := models.ParseID(vars["model_id"])
	versionID, _ := models.ParseID(vars["version_id"])

	model, err := c.ModelsService.FindByID(ctx, modelID)
	if err != nil {
		return NotFound(fmt.Sprintf("Model with given `model_id: %d` not found", modelID))
	}

	version, err := c.VersionsService.FindByID(ctx, modelID, versionID, c.FeatureToggleConfig.MonitoringConfig)
	if err != nil {
		return NotFound(fmt.Sprintf("Version with given `version_id: %d` not found: %v", versionID, err))
	}

	endpoints, err := c.EndpointsService.ListEndpoints(ctx, model, version)
	if err != nil {
		return InternalServerError(err.Error())
	}

	if c.FeatureToggleConfig.MonitoringConfig.MonitoringEnabled {
		for k := range endpoints {
			endpoints[k].UpdateMonitoringURL(c.FeatureToggleConfig.MonitoringConfig.MonitoringBaseURL, models.EndpointMonitoringURLParams{
				Cluster:      endpoints[k].Environment.Cluster,
				Project:      model.Project.Name,
				Model:        model.Name,
				ModelVersion: model.Name + "-" + version.ID.String(),
			})
		}
	}
	return Ok(endpoints)
}

// GetEndpoint get model endpoint with certain ID
func (c *EndpointsController) GetEndpoint(r *http.Request, vars map[string]string, _ interface{}) *Response {
	ctx := r.Context()

	modelID, _ := models.ParseID(vars["model_id"])
	versionID, _ := models.ParseID(vars["version_id"])
	endpointID, _ := uuid.Parse(vars["endpoint_id"])

	model, err := c.ModelsService.FindByID(ctx, modelID)
	if err != nil {
		return NotFound(fmt.Sprintf("Model not found: %v", err))
	}

	version, err := c.VersionsService.FindByID(ctx, modelID, versionID, c.FeatureToggleConfig.MonitoringConfig)
	if err != nil {
		return NotFound(fmt.Sprintf("Version not found: %v", err))
	}

	endpoint, err := c.EndpointsService.FindByID(ctx, endpointID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFound(fmt.Sprintf("Version endpoint not found: %v", err))
		}
		return InternalServerError(fmt.Sprintf("Error getting version endpoint: %v", err))
	}

	if c.FeatureToggleConfig.MonitoringConfig.MonitoringEnabled {
		endpoint.UpdateMonitoringURL(c.FeatureToggleConfig.MonitoringConfig.MonitoringBaseURL, models.EndpointMonitoringURLParams{
			Cluster:      endpoint.Environment.Cluster,
			Project:      model.Project.Name,
			Model:        model.Name,
			ModelVersion: model.Name + "-" + version.ID.String(),
		})
	}

	return Ok(endpoint)
}

// CreateEndpoint create new endpoint from a model version and deploy to certain environment as specified by request
// If target environment is not set then fallback to default environment
func (c *EndpointsController) CreateEndpoint(r *http.Request, vars map[string]string, body interface{}) *Response {
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
	env, err := c.AppContext.EnvironmentService.GetDefaultEnvironment()
	if err != nil {
		return InternalServerError(fmt.Sprintf("Unable to find default environment, specify environment target for deployment: %v", err))
	}

	newEndpoint := &models.VersionEndpoint{}

	// SDK > v0.1.0 send request body
	if body != nil {
		data, ok := body.(*models.VersionEndpoint)
		if !ok {
			return BadRequest("Unable to parse body as version endpoint resource")
		}

		newEndpoint = data

		if newEndpoint.EnvironmentName != "" {
			env, err = c.AppContext.EnvironmentService.GetEnvironment(newEndpoint.EnvironmentName)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return NotFound(fmt.Sprintf("Environment not found: %v", err))
				}
				return InternalServerError(fmt.Sprintf("Error getting environment: %v", err))
			}
		}

		newEndpoint.EnvironmentName = env.Name
		newEndpoint.SetModeObservabilityIfNil()
	}

	validationRules := []requestValidator{
		resourceRequestValidation(newEndpoint),
		customModelValidation(model, version),
		upiModelValidation(model, newEndpoint.Protocol),
		newVersionEndpointValidation(version, env.Name),
		deploymentQuotaValidation(ctx, model, env, c.EndpointsService),
		transformerValidation(ctx, newEndpoint, c.StandardTransformerConfig, c.FeastCoreClient),
		modelObservabilityValidation(newEndpoint, model),
	}

	if err := validateRequest(validationRules...); err != nil {
		return BadRequest(fmt.Sprintf("Request validation failed: %v", err))
	}

	endpoint, err := c.EndpointsService.DeployEndpoint(ctx, env, model, version, newEndpoint)
	if err != nil {
		if errors.Is(err, merror.ErrInvalidInput) {
			return BadRequest(fmt.Sprintf("Unable to process model version input: %v", err))
		}
		return InternalServerError(fmt.Sprintf("Unable to deploy model version: %v", err))
	}

	return Created(endpoint)
}

// UpdateEndpoint update a an existing endpoint i.e. trigger redeployment
func (c *EndpointsController) UpdateEndpoint(r *http.Request, vars map[string]string, body interface{}) *Response {
	ctx := r.Context()

	modelID, _ := models.ParseID(vars["model_id"])
	versionID, _ := models.ParseID(vars["version_id"])
	endpointID, _ := uuid.Parse(vars["endpoint_id"])

	model, version, err := c.getModelAndVersion(ctx, modelID, versionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFound(fmt.Sprintf("Model / version not found: %v", err))
		}
		return InternalServerError(fmt.Sprintf("Error getting model / version: %v", err))
	}

	endpoint, err := c.EndpointsService.FindByID(ctx, endpointID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFound(fmt.Sprintf("Endpoint not found: %v", err))
		}
		return InternalServerError(fmt.Sprintf("Error getting endpoint: %v", err))
	}

	newEndpoint, ok := body.(*models.VersionEndpoint)
	if !ok {
		return BadRequest("Unable to parse body as version endpoint resource")
	}

	newEndpoint.SetModeObservabilityIfNil()

	env, err := c.AppContext.EnvironmentService.GetEnvironment(newEndpoint.EnvironmentName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFound(fmt.Sprintf("Environment not found: %v", err))
		}
		return InternalServerError(fmt.Sprintf("Error getting the specified environment: %v", err))
	}

	validationRules := []requestValidator{
		resourceRequestValidation(newEndpoint),
		customModelValidation(model, version),
		updateRequestValidation(endpoint, newEndpoint),
		modelObservabilityValidation(newEndpoint, model),
	}

	if newEndpoint.Status == models.EndpointRunning || newEndpoint.Status == models.EndpointServing {
		validationRules = append(
			validationRules,
			transformerValidation(ctx, newEndpoint, c.StandardTransformerConfig, c.FeastCoreClient),
			deploymentModeValidation(endpoint, newEndpoint),
		)

		if err := validateRequest(validationRules...); err != nil {
			return BadRequest(fmt.Sprintf("Request validation failed: %v", err))
		}

		endpoint, err = c.EndpointsService.DeployEndpoint(ctx, env, model, version, newEndpoint)
		if err != nil {
			if errors.Is(err, merror.ErrInvalidInput) {
				return BadRequest(fmt.Sprintf("Unable to deploy model version: %v", err))
			}

			return InternalServerError(fmt.Sprintf("Unable to deploy model version: %v", err))
		}
	} else if newEndpoint.Status == models.EndpointTerminated {
		if err := validateRequest(validationRules...); err != nil {
			return BadRequest(fmt.Sprintf("Request validation failed: %v", err))
		}

		endpoint, err = c.EndpointsService.UndeployEndpoint(ctx, env, model, version, endpoint)
		if err != nil {
			return InternalServerError(fmt.Sprintf("Unable to undeploy version endpoint %s: %v", endpointID, err))
		}
	} else {
		return BadRequest(fmt.Sprintf("Updating endpoint status to %s is not allowed", newEndpoint.Status))
	}

	return Ok(endpoint)
}

// DeleteEndpoint undeploys running model version endpoint.
func (c *EndpointsController) DeleteEndpoint(r *http.Request, vars map[string]string, _ interface{}) *Response {
	ctx := r.Context()

	versionID, _ := models.ParseID(vars["version_id"])
	modelID, _ := models.ParseID(vars["model_id"])
	rawEndpointID, ok := vars["endpoint_id"]
	endpointID, _ := uuid.Parse(rawEndpointID)

	model, version, err := c.getModelAndVersion(ctx, modelID, versionID)
	if err != nil {
		return NotFound(fmt.Sprintf("Error getting model / version: %v", err))
	}

	var env *models.Environment
	var endpoint *models.VersionEndpoint
	if !ok {
		// SDK v0.1.0 doesn't send endpoint id to undeploy
		// Fetch endpoint in default environment and undeploy it
		env, err = c.AppContext.EnvironmentService.GetDefaultEnvironment()
		if err != nil {
			return NotFound(fmt.Sprintf("Unable to find default environment: %v", err))
		}

		endpoint, ok = version.GetEndpointByEnvironmentName(env.Name)
		if !ok {
			return NotFound(fmt.Sprintf("Endpoint with given `environment_name: %s` not found: %v", env.Name, err))
		}
	} else {

		if err != nil {
			return BadRequest(fmt.Sprintf("Unable to parse endpoint_id %s: %v", rawEndpointID, err))
		}

		endpoint, err = c.EndpointsService.FindByID(ctx, endpointID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Record does not existing, nothing to delete
				return Ok(fmt.Sprintf("Endpoint not found: %v", err))
			}
			return InternalServerError(fmt.Sprintf("Error while finding endpoint: %v", err))
		}

		env, err = c.EnvironmentService.GetEnvironment(endpoint.EnvironmentName)
		if err != nil {
			return InternalServerError(fmt.Sprintf("Unable to find environment %s: %v", endpoint.EnvironmentName, err))
		}
	}

	if endpoint.IsServing() {
		return BadRequest(fmt.Sprintf("Version Endpoints %s is still serving traffic. Please route the traffic to another model version first", rawEndpointID))
	}

	if endpoint.IsPending() {
		return BadRequest(fmt.Sprintf("Version Endpoint %s is still pending and cannot be undeployed", rawEndpointID))
	}

	_, err = c.EndpointsService.UndeployEndpoint(ctx, env, model, version, endpoint)
	if err != nil {
		return InternalServerError(fmt.Sprintf("Unable to undeploy version endpoint %s: %v", rawEndpointID, err))
	}
	return Ok(nil)
}

func (c *EndpointsController) ListContainers(r *http.Request, vars map[string]string, body interface{}) *Response {
	ctx := r.Context()

	versionID, _ := models.ParseID(vars["version_id"])
	endpointID, _ := uuid.Parse(vars["endpoint_id"])
	modelID, _ := models.ParseID(vars["model_id"])

	model, err := c.ModelsService.FindByID(ctx, modelID)
	if err != nil {
		return NotFound(fmt.Sprintf("Model not found: %v", err))
	}
	version, err := c.VersionsService.FindByID(ctx, modelID, versionID, c.FeatureToggleConfig.MonitoringConfig)
	if err != nil {
		return NotFound(fmt.Sprintf("Version not found: %v", err))
	}
	endpoint, err := c.EndpointsService.FindByID(ctx, endpointID)
	if err != nil {
		return NotFound(fmt.Sprintf("Endpoint not found: %v", err))
	}

	containers, err := c.EndpointsService.ListContainers(ctx, model, version, endpoint)
	if err != nil {
		return InternalServerError(fmt.Sprintf("Error while getting container for endpoint: %v", err))
	}
	return Ok(containers)
}

func validateTransformer(ctx context.Context, endpoint *models.VersionEndpoint, stdTransformerConfig config.StandardTransformerConfig, feastCore core.CoreServiceClient) error {
	trans := endpoint.Transformer
	protocol := endpoint.Protocol
	logger := endpoint.Logger
	switch trans.TransformerType {
	case models.CustomTransformerType, models.DefaultTransformerType:
		if trans.Image == "" {
			return errors.New("transformer image name is not specified")
		}
	case models.StandardTransformerType:
		envVars := trans.EnvVars.ToMap()
		cfg, ok := envVars[transformer.StandardTransformerConfigEnvName]
		if !ok {
			return errors.New("standard transformer config is not specified")
		}

		var predictionLogCfg *spec.PredictionLogConfig
		if logger != nil && logger.Prediction != nil {
			predictionLogCfg = logger.Prediction.ToPredictionLogConfig()
		}

		feastOptions := &feast.Options{
			StorageConfigs: stdTransformerConfig.ToFeastStorageConfigs(),
		}
		return validateStandardTransformerConfig(ctx, cfg, protocol, predictionLogCfg, feastOptions, feastCore)
	default:
		return fmt.Errorf("unknown transformer type: %s", trans.TransformerType)
	}

	return nil
}

func validateCustomPredictor(version *models.Version) error {
	customPredictor := version.CustomPredictor
	if customPredictor == nil {
		return errors.New("custom predictor must be specified")
	}
	return customPredictor.IsValid()
}

func validateStandardTransformerConfig(ctx context.Context, cfg string, protocol protocol.Protocol, predictionLogConfig *spec.PredictionLogConfig, feastOpts *feast.Options, feastCore core.CoreServiceClient) error {
	stdTransformerConfig := &spec.StandardTransformerConfig{}
	err := protojson.Unmarshal([]byte(cfg), stdTransformerConfig)
	if err != nil {
		return err
	}

	stdTransformerConfig.PredictionLogConfig = predictionLogConfig

	return pipeline.ValidateTransformerConfig(ctx, feastCore, stdTransformerConfig, feastOpts, protocol)
}
