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

	// validate custom predictor
	if model.Type == models.ModelTypeCustom {
		err := c.validateCustomPredictor(ctx, version)
		if err != nil {
			return BadRequest(fmt.Sprintf("Error validating custom predictor: %v", err))
		}
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
	}

	// check that UPI is supported
	if !isModelSupportUPI(model) && newEndpoint.Protocol == protocol.UpiV1 {
		return BadRequest(
			fmt.Sprintf("%s model is not supported by UPI", model.Type))
	}

	// check that the endpoint is not deployed nor deploying
	endpoint, ok := version.GetEndpointByEnvironmentName(env.Name)
	if ok && (endpoint.IsRunning() || endpoint.IsServing()) {
		return BadRequest(
			fmt.Sprintf("There is `%s` deployment for the model version", endpoint.Status))
	}

	// check that the model version quota
	deployedModelVersionCount, err := c.EndpointsService.CountEndpoints(ctx, env, model)
	if err != nil {
		return InternalServerError(fmt.Sprintf("Unable to count number of endpoints in env %s: %v", env.Name, err))
	}

	if deployedModelVersionCount >= config.MaxDeployedVersion {
		return BadRequest(fmt.Sprintf("Max deployed endpoint reached. Max: %d Current: %d, undeploy existing endpoint before continuing", config.MaxDeployedVersion, deployedModelVersionCount))
	}

	// validate transformer
	if newEndpoint.Transformer != nil && newEndpoint.Transformer.Enabled {
		err := c.validateTransformer(ctx, newEndpoint.Transformer, newEndpoint.Protocol, newEndpoint.Logger)
		if err != nil {
			return BadRequest(fmt.Sprintf("Error validating transformer: %v", err))
		}
	}

	endpoint, err = c.EndpointsService.DeployEndpoint(ctx, env, model, version, newEndpoint)
	if err != nil {
		if errors.Is(err, merror.InvalidInputError) {
			return BadRequest(fmt.Sprintf("Unable to process model version input: %v", err))
		}
		return InternalServerError(fmt.Sprintf("Unable to deploy model version: %v", err))
	}

	return Created(endpoint)
}

var supportedUPIModelTypes = map[string]bool{
	models.ModelTypePyFunc: true,
	models.ModelTypeCustom: true,
}

func isModelSupportUPI(model *models.Model) bool {
	_, isSupported := supportedUPIModelTypes[model.Type]

	return isSupported
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

	// validate custom predictor
	if model.Type == models.ModelTypeCustom {
		err := c.validateCustomPredictor(ctx, version)
		if err != nil {
			return BadRequest(fmt.Sprintf("Error validating custom predictor: %v", err))
		}
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

	err = validateUpdateRequest(endpoint, newEndpoint)
	if err != nil {
		return BadRequest(fmt.Sprintf("Error validating request: %v", err))
	}

	env, err := c.AppContext.EnvironmentService.GetEnvironment(newEndpoint.EnvironmentName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFound(fmt.Sprintf("Environment not found: %v", err))
		}
		return InternalServerError(fmt.Sprintf("Error getting the specified environment: %v", err))
	}

	if newEndpoint.Status == models.EndpointRunning || newEndpoint.Status == models.EndpointServing {
		// validate transformer
		if newEndpoint.Transformer != nil && newEndpoint.Transformer.Enabled {
			err := c.validateTransformer(ctx, newEndpoint.Transformer, newEndpoint.Protocol, newEndpoint.Logger)
			if err != nil {
				return BadRequest(fmt.Sprintf("Error validating the transformer: %v", err))
			}
		}

		// Should not allow changing the deployment mode of a pending/running/serving model for 2 reasons:
		// * For "serving" models it's risky as, we can't guarantee graceful re-deployment
		// * Kserve uses slightly different deployment resource naming under the hood and doesn't clean up the older deployment
		if (endpoint.IsRunning() || endpoint.IsServing()) && newEndpoint.DeploymentMode != endpoint.DeploymentMode {
			return BadRequest(fmt.Sprintf("Changing deployment type of a %s model is not allowed, please terminate it first.",
				endpoint.Status))
		}

		endpoint, err = c.EndpointsService.DeployEndpoint(ctx, env, model, version, newEndpoint)
		if err != nil {
			if errors.Is(err, merror.InvalidInputError) {
				return BadRequest(fmt.Sprintf("Unable to deploy model version: %v", err))
			}

			return InternalServerError(fmt.Sprintf("Unable to deploy model version: %v", err))
		}
	} else if newEndpoint.Status == models.EndpointTerminated {
		endpoint, err = c.EndpointsService.UndeployEndpoint(ctx, env, model, version, endpoint)
		if err != nil {
			return InternalServerError(fmt.Sprintf("Unable to undeploy version endpoint %s: %v", endpointID, err))
		}
	} else {
		return InternalServerError(fmt.Sprintf("Updating endpoint status to %s is not allowed", newEndpoint.Status))
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

	endpoint, err := c.EndpointsService.ListContainers(ctx, model, version, endpointID)
	if err != nil {
		return InternalServerError(fmt.Sprintf("Error while getting container for endpoint: %v", err))
	}
	return Ok(endpoint)
}

func validateUpdateRequest(prev *models.VersionEndpoint, new *models.VersionEndpoint) error {
	if prev.EnvironmentName != new.EnvironmentName {
		return fmt.Errorf("Updating environment is not allowed, previous: %s, new: %s", prev.EnvironmentName, new.EnvironmentName)
	}

	if prev.Status == models.EndpointPending {
		return fmt.Errorf("Updating endpoint status to %s is not allowed when the endpoint is currently in the pending state", new.Status)
	}

	if new.Status != prev.Status {
		if prev.Status == models.EndpointServing {
			return fmt.Errorf("Updating endpoint status to %s is not allowed when the endpoint is currently in the serving state", new.Status)
		}

		if new.Status != models.EndpointRunning && new.Status != models.EndpointTerminated {
			return fmt.Errorf("Updating endpoint status to %s is not allowed", new.Status)
		}
	}

	return nil
}

func (c *EndpointsController) validateTransformer(ctx context.Context, trans *models.Transformer, protocol protocol.Protocol, logger *models.Logger) error {
	switch trans.TransformerType {
	case models.CustomTransformerType, models.DefaultTransformerType:
		if trans.Image == "" {
			return errors.New("Transformer image name is not specified")
		}
	case models.StandardTransformerType:
		envVars := trans.EnvVars.ToMap()
		cfg, ok := envVars[transformer.StandardTransformerConfigEnvName]
		if !ok {
			return errors.New("Standard transformer config is not specified")
		}

		var predictionLogCfg *spec.PredictionLogConfig
		if logger != nil && logger.Prediction != nil {
			predictionLogCfg = logger.Prediction.ToPredictionLogConfig()
		}

		return c.validateStandardTransformerConfig(ctx, cfg, protocol, predictionLogCfg)
	default:
		return fmt.Errorf("Unknown transformer type: %s", trans.TransformerType)
	}

	return nil
}

func (c *EndpointsController) validateCustomPredictor(ctx context.Context, version *models.Version) error {
	customPredictor := version.CustomPredictor
	if customPredictor == nil {
		return errors.New("custom predictor must be specified")
	}
	return customPredictor.IsValid()
}

func (c *EndpointsController) validateStandardTransformerConfig(ctx context.Context, cfg string, protocol protocol.Protocol, predictionLogConfig *spec.PredictionLogConfig) error {
	stdTransformerConfig := &spec.StandardTransformerConfig{}
	err := protojson.Unmarshal([]byte(cfg), stdTransformerConfig)
	if err != nil {
		return err
	}

	feastOptions := &feast.Options{
		StorageConfigs: c.StandardTransformerConfig.ToFeastStorageConfigs(),
	}

	stdTransformerConfig.PredictionLogConfig = predictionLogConfig

	return pipeline.ValidateTransformerConfig(ctx, c.FeastCoreClient, stdTransformerConfig, feastOptions, protocol)
}
