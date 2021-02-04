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

package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/gojek/merlin/cluster"
	"github.com/gojek/merlin/config"
	"github.com/gojek/merlin/imagebuilder"
	"github.com/gojek/merlin/log"
	"github.com/gojek/merlin/models"
	"github.com/gojek/merlin/storage"
)

var deploymentCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name:      "deploy_count",
		Namespace: "merlin_api",
		Help:      "Number of deployment",
	},
	[]string{"project", "model", "status"},
)

func init() {
	prometheus.MustRegister(deploymentCounter)
}

type EndpointsService interface {
	ListEndpoints(model *models.Model, version *models.Version) ([]*models.VersionEndpoint, error)
	FindByID(uuid2 uuid.UUID) (*models.VersionEndpoint, error)
	DeployEndpoint(environment *models.Environment, model *models.Model, version *models.Version, endpoint *models.VersionEndpoint) (*models.VersionEndpoint, error)
	UndeployEndpoint(environment *models.Environment, model *models.Model, version *models.Version, endpoint *models.VersionEndpoint) (*models.VersionEndpoint, error)
	CountEndpoints(environment *models.Environment, model *models.Model) (int, error)
	ListContainers(model *models.Model, version *models.Version, id uuid.UUID) ([]*models.Container, error)
}

const defaultWorkers = 1

type endpointService struct {
	clusterControllers   map[string]cluster.Controller
	imageBuilder         imagebuilder.ImageBuilder
	storage              storage.VersionEndpointStorage
	deploymentStorage    storage.DeploymentStorage
	environment          string
	monitoringConfig     config.MonitoringConfig
	loggerDestinationURL string
}

func NewEndpointService(clusterControllers map[string]cluster.Controller,
	imageBuilder imagebuilder.ImageBuilder,
	storage storage.VersionEndpointStorage,
	deploymentStorage storage.DeploymentStorage,
	environment string,
	monitoringConfig config.MonitoringConfig,
	loggerDestinationURL string) EndpointsService {
	return &endpointService{
		clusterControllers:   clusterControllers,
		imageBuilder:         imageBuilder,
		storage:              storage,
		deploymentStorage:    deploymentStorage,
		environment:          environment,
		monitoringConfig:     monitoringConfig,
		loggerDestinationURL: loggerDestinationURL,
	}
}

func (k *endpointService) ListEndpoints(model *models.Model, version *models.Version) ([]*models.VersionEndpoint, error) {
	endpoints, err := k.storage.ListEndpoints(model, version)
	if err != nil {
		return nil, err
	}

	return endpoints, nil
}

func (k *endpointService) FindByID(uuid uuid.UUID) (*models.VersionEndpoint, error) {
	return k.storage.Get(uuid)
}

func (k *endpointService) DeployEndpoint(environment *models.Environment, model *models.Model, version *models.Version, newEndpoint *models.VersionEndpoint) (*models.VersionEndpoint, error) {
	ctl, ok := k.clusterControllers[environment.Name]
	if !ok {
		return nil, fmt.Errorf("unable to find cluster controller for environment %s", environment.Name)
	}

	endpoint, ok := version.GetEndpointByEnvironmentName(environment.Name)
	if endpoint == nil {
		endpoint = models.NewVersionEndpoint(environment, model.Project, model, version, k.monitoringConfig)
	}

	if endpoint.ResourceRequest == nil {
		endpoint.ResourceRequest = environment.DefaultResourceRequest
	}

	if newEndpoint.ResourceRequest != nil {
		endpoint.ResourceRequest = newEndpoint.ResourceRequest
	}

	if newEndpoint.Transformer != nil {
		if newEndpoint.Transformer.ResourceRequest == nil {
			newEndpoint.Transformer.ResourceRequest = environment.DefaultTransformerResourceRequest
		}

		if newEndpoint.Transformer.TransformerType == models.DefaultTransformerType {
			newEndpoint.Transformer.TransformerType = models.CustomTransformerType
		}

		endpoint.Transformer = newEndpoint.Transformer
		endpoint.Transformer.VersionEndpointID = endpoint.ID
	}

	if newEndpoint.Logger != nil {
		endpoint.Logger = newEndpoint.Logger
		endpoint.Logger.DestinationURL = k.loggerDestinationURL
	}

	// Configure environment variables for Pyfunc model
	if model.Type == models.ModelTypePyFunc {
		pyfuncDefaultEnvVars := models.PyfuncDefaultEnvVars(*model, *version, defaultWorkers)

		// This section is for:
		// 1. backward-compatibility
		// 2. when user didn't specify any env vars config at the first time
		if len(endpoint.EnvVars) == 0 {
			endpoint.EnvVars = pyfuncDefaultEnvVars
		}

		if len(newEndpoint.EnvVars) > 0 {
			if err := newEndpoint.EnvVars.CheckForProtectedEnvVars(); err != nil {
				return nil, err
			}
			endpoint.EnvVars = models.MergeEnvVars(pyfuncDefaultEnvVars, newEndpoint.EnvVars)
		}
	}

	previousStatus := endpoint.Status
	endpoint.Status = models.EndpointPending

	err := k.storage.Save(endpoint)
	if err != nil {
		return nil, err
	}

	go func() {
		log.Infof("creating deployment for model %s version %s with endpoint id: %s", model.Name, version.ID, endpoint.ID)

		// copy endpoint to avoid race condition
		ep := *endpoint
		ep.Status = models.EndpointFailed
		defer func() {
			deploymentCounter.WithLabelValues(model.Project.Name, model.Name, string(ep.Status)).Inc()

			// record the deployment result
			if _, err := k.deploymentStorage.Save(&models.Deployment{
				ProjectID:         model.ProjectID,
				VersionModelID:    model.ID,
				VersionID:         version.ID,
				VersionEndpointID: ep.ID,
				Status:            ep.Status,
				Error:             ep.Message,
			}); err != nil {
				log.Warnf("unable to insert deployment history", err)
			}

			if err := k.storage.Save(&ep); err != nil {
				log.Errorf("unable to update endpoint status for model: %s, version: %s, reason: %v", model.Name, version.ID, err)
			}
		}()

		modelOpt := &models.ModelOption{}
		if model.Type == models.ModelTypePyFunc {
			imageRef, err := k.imageBuilder.BuildImage(model.Project, model, version)
			modelOpt.PyFuncImageName = imageRef
			if err != nil {
				ep.Message = err.Error()
				return
			}
		} else if model.Type == models.ModelTypePyTorch {
			modelOpt = models.NewPyTorchModelOption(version)
		}

		modelService := models.NewService(model, version, modelOpt, endpoint.ResourceRequest, endpoint.EnvVars, k.environment, endpoint.Transformer, endpoint.Logger)
		svc, err := ctl.Deploy(modelService)
		if err != nil {
			log.Errorf("unable to deploy version endpoint for model: %s, version: %s, reason: %v", model.Name, version.ID, err)
			ep.Message = err.Error()
			return
		}

		ep.URL = svc.URL
		if previousStatus == models.EndpointServing {
			ep.Status = models.EndpointServing
		} else {
			ep.Status = models.EndpointRunning
		}
		ep.ServiceName = svc.ServiceName
	}()

	return endpoint, nil
}

func (k *endpointService) UndeployEndpoint(environment *models.Environment, model *models.Model, version *models.Version, endpoint *models.VersionEndpoint) (*models.VersionEndpoint, error) {
	ctl, ok := k.clusterControllers[environment.Name]
	if !ok {
		return nil, fmt.Errorf("unable to find cluster controller for environment %s", environment.Name)
	}

	modelService := &models.Service{
		Name:      models.CreateInferenceServiceName(model.Name, version.ID.String()),
		Namespace: model.Project.Name,
	}

	_, err := ctl.Delete(modelService)
	if err != nil {
		return nil, err
	}

	endpoint.Status = models.EndpointTerminated
	err = k.storage.Save(endpoint)
	if err != nil {
		return nil, err
	}

	return endpoint, nil
}

// CountEndpoints count number of running/pending version endpoint of a model within an environment
func (k *endpointService) CountEndpoints(environment *models.Environment, model *models.Model) (int, error) {
	return k.storage.CountEndpoints(environment, model)
}

// ListContainers list all containers belong to the given version endpoint
func (k *endpointService) ListContainers(model *models.Model, version *models.Version, id uuid.UUID) ([]*models.Container, error) {
	ve, err := k.storage.Get(id)
	if err != nil {
		return nil, err
	}

	ctl, ok := k.clusterControllers[ve.EnvironmentName]
	if !ok {
		return nil, fmt.Errorf("unable to find cluster controller for environment %s", ve.EnvironmentName)
	}

	containers := make([]*models.Container, 0)
	if model.Type == models.ModelTypePyFunc {
		imgBuilderContainers, err := k.imageBuilder.GetContainers(model.Project, model, version)
		if err != nil {
			return nil, err
		}

		containers = append(containers, imgBuilderContainers...)
	}

	modelContainers, err := ctl.GetContainers(model.Project.Name, models.OnlineInferencePodLabelSelector(model.Name, version.ID.String()))
	if err != nil {
		return nil, err
	}
	containers = append(containers, modelContainers...)

	for _, container := range containers {
		container.VersionEndpointID = id
	}

	return containers, nil
}
