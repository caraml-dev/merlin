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
	"context"
	"encoding/json"
	"fmt"

	"github.com/feast-dev/feast/sdk/go/protos/feast/core"
	"github.com/gojek/merlin/cluster"
	"github.com/gojek/merlin/config"
	"github.com/gojek/merlin/log"
	"github.com/gojek/merlin/models"
	"github.com/gojek/merlin/pkg/autoscaling"
	"github.com/gojek/merlin/pkg/imagebuilder"
	"github.com/gojek/merlin/pkg/transformer"
	"github.com/gojek/merlin/pkg/transformer/feast"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/queue"
	"github.com/gojek/merlin/queue/work"
	"github.com/gojek/merlin/storage"
	"github.com/golang/protobuf/jsonpb"
	"github.com/google/uuid"
)

type EndpointsService interface {
	// ListEndpoints list all endpoint created from a model version
	ListEndpoints(ctx context.Context, model *models.Model, version *models.Version) ([]*models.VersionEndpoint, error)
	// FindByID find specific endpoint using the given uuid
	FindByID(ctx context.Context, endpointUuid uuid.UUID) (*models.VersionEndpoint, error)
	// DeployEndpoint update or create an endpoint given a model version in the specified deployment environment
	DeployEndpoint(ctx context.Context, environment *models.Environment, model *models.Model, version *models.Version, endpoint *models.VersionEndpoint) (*models.VersionEndpoint, error)
	// UndeployEndpoint delete an endpoint given a model version in the specified deployment environment
	UndeployEndpoint(ctx context.Context, environment *models.Environment, model *models.Model, version *models.Version, endpoint *models.VersionEndpoint) (*models.VersionEndpoint, error)
	// CountEndpoints count number of endpoint created from a model in an environment
	CountEndpoints(ctx context.Context, environment *models.Environment, model *models.Model) (int, error)
	// ListContainers list all container associated with an endpoint
	ListContainers(ctx context.Context, model *models.Model, version *models.Version, endpointUuid uuid.UUID) ([]*models.Container, error)
}

const defaultWorkers = 1

type EndpointServiceParams struct {
	ClusterControllers        map[string]cluster.Controller
	ImageBuilder              imagebuilder.ImageBuilder
	Storage                   storage.VersionEndpointStorage
	DeploymentStorage         storage.DeploymentStorage
	Environment               string
	MonitoringConfig          config.MonitoringConfig
	LoggerDestinationURL      string
	JobProducer               queue.Producer
	FeastCoreClient           core.CoreServiceClient
	StandardTransformerConfig config.StandardTransformerConfig
}

type endpointService struct {
	clusterControllers        map[string]cluster.Controller
	imageBuilder              imagebuilder.ImageBuilder
	storage                   storage.VersionEndpointStorage
	deploymentStorage         storage.DeploymentStorage
	environment               string
	monitoringConfig          config.MonitoringConfig
	loggerDestinationURL      string
	jobProducer               queue.Producer
	feastCoreClient           core.CoreServiceClient
	standardTransformerConfig config.StandardTransformerConfig
}

func NewEndpointService(params EndpointServiceParams) EndpointsService {
	return &endpointService{
		clusterControllers:        params.ClusterControllers,
		imageBuilder:              params.ImageBuilder,
		storage:                   params.Storage,
		deploymentStorage:         params.DeploymentStorage,
		environment:               params.Environment,
		monitoringConfig:          params.MonitoringConfig,
		loggerDestinationURL:      params.LoggerDestinationURL,
		jobProducer:               params.JobProducer,
		feastCoreClient:           params.FeastCoreClient,
		standardTransformerConfig: params.StandardTransformerConfig,
	}
}

func (k *endpointService) ListEndpoints(ctx context.Context, model *models.Model, version *models.Version) ([]*models.VersionEndpoint, error) {
	endpoints, err := k.storage.ListEndpoints(model, version)
	if err != nil {
		return nil, err
	}

	return endpoints, nil
}

func (k *endpointService) FindByID(ctx context.Context, endpointUuid uuid.UUID) (*models.VersionEndpoint, error) {
	return k.storage.Get(endpointUuid)
}

func (k *endpointService) DeployEndpoint(ctx context.Context, environment *models.Environment, model *models.Model, version *models.Version, newEndpoint *models.VersionEndpoint) (*models.VersionEndpoint, error) {
	endpoint, _ := version.GetEndpointByEnvironmentName(environment.Name)
	if endpoint == nil {
		endpoint = models.NewVersionEndpoint(environment, model.Project, model, version, k.monitoringConfig)
	}

	if newEndpoint.ResourceRequest != nil {
		endpoint.ResourceRequest = newEndpoint.ResourceRequest
	}

	if newEndpoint.DeploymentMode != "" {
		endpoint.DeploymentMode = newEndpoint.DeploymentMode
	}

	if newEndpoint.AutoscalingPolicy != nil {
		err := autoscaling.ValidateAutoscalingPolicy(newEndpoint.AutoscalingPolicy, endpoint.DeploymentMode)
		if err != nil {
			return nil, err
		}
		endpoint.AutoscalingPolicy = newEndpoint.AutoscalingPolicy
	}

	if newEndpoint.Transformer != nil {
		if newEndpoint.Transformer.ResourceRequest == nil {
			newEndpoint.Transformer.ResourceRequest = environment.DefaultTransformerResourceRequest
		}

		if newEndpoint.Transformer.TransformerType == models.DefaultTransformerType {
			newEndpoint.Transformer.TransformerType = models.CustomTransformerType
		}

		if newEndpoint.Transformer.TransformerType == models.StandardTransformerType {
			// update standard transformer config
			// 1. Add feature table metadata variables to transformer
			// 2. Update feature table source if empty
			updatedStandardTransformer, err := k.reconfigureStandardTransformer(newEndpoint.Transformer)
			if err != nil {
				return nil, err
			}
			endpoint.Transformer = updatedStandardTransformer
		}

		endpoint.Transformer = newEndpoint.Transformer
		endpoint.Transformer.VersionEndpointID = endpoint.ID
	}

	if newEndpoint.Logger != nil {
		endpoint.Logger = newEndpoint.Logger
		endpoint.Logger.DestinationURL = k.loggerDestinationURL
		modelLogger := endpoint.Logger.Model
		if modelLogger != nil {
			modelLogger.SanitizeMode()
			endpoint.Logger.Model = modelLogger
		}
		transformerLogger := endpoint.Logger.Transformer
		if transformerLogger != nil {
			transformerLogger.SanitizeMode()
			endpoint.Logger.Transformer = transformerLogger
		}
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
	} else {
		endpoint.EnvVars = newEndpoint.EnvVars
	}

	originalEndpoint := *endpoint

	endpoint.Status = models.EndpointPending
	err := k.storage.Save(endpoint)
	if err != nil {
		return nil, err
	}

	if err := k.jobProducer.EnqueueJob(&queue.Job{
		Name: ModelServiceDeployment,
		Arguments: queue.Arguments{
			dataArgKey: work.EndpointJob{
				Endpoint: &originalEndpoint,
				Version:  version,
				Model:    model,
				Project:  model.Project,
			},
		},
	}); err != nil {
		// if error enqueue job, mark endpoint status to failed
		endpoint.Status = models.EndpointFailed
		if err := k.storage.Save(endpoint); err != nil {
			log.Errorf("error to update endpoint %s status to failed: %v", endpoint.ID, err)
		}
		return nil, err
	}

	return endpoint, nil
}

func (k *endpointService) UndeployEndpoint(ctx context.Context, environment *models.Environment, model *models.Model, version *models.Version, endpoint *models.VersionEndpoint) (*models.VersionEndpoint, error) {
	ctl, ok := k.clusterControllers[environment.Name]
	if !ok {
		return nil, fmt.Errorf("unable to find cluster controller for environment %s", environment.Name)
	}

	modelService := &models.Service{
		Name:      models.CreateInferenceServiceName(model.Name, version.ID.String()),
		Namespace: model.Project.Name,
	}

	_, err := ctl.Delete(ctx, modelService)
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
func (k *endpointService) CountEndpoints(ctx context.Context, environment *models.Environment, model *models.Model) (int, error) {
	return k.storage.CountEndpoints(environment, model)
}

// ListContainers list all containers belong to the given version endpoint
func (k *endpointService) ListContainers(ctx context.Context, model *models.Model, version *models.Version, id uuid.UUID) ([]*models.Container, error) {
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
		imgBuilderContainers, err := k.imageBuilder.GetContainers(ctx, model.Project, model, version)
		if err != nil {
			return nil, err
		}

		containers = append(containers, imgBuilderContainers...)
	}

	modelContainers, err := ctl.GetContainers(ctx, model.Project.Name, models.OnlineInferencePodLabelSelector(model.Name, version.ID.String()))
	if err != nil {
		return nil, err
	}
	containers = append(containers, modelContainers...)

	for _, container := range containers {
		container.VersionEndpointID = id
	}

	return containers, nil
}

func (k *endpointService) reconfigureStandardTransformer(standardTransformer *models.Transformer) (*models.Transformer, error) {
	envVars := standardTransformer.EnvVars
	envVarsMap := envVars.ToMap()
	stdTransformerConfigString := envVarsMap[transformer.StandardTransformerConfigEnvName]
	stdTransformerConfig := &spec.StandardTransformerConfig{}
	err := jsonpb.UnmarshalString(stdTransformerConfigString, stdTransformerConfig)
	if err != nil {
		return nil, err
	}
	// add feature table metadata
	standardTransformer, err = k.addFeatureTableMetadata(standardTransformer, stdTransformerConfig)
	if err != nil {
		return nil, err
	}

	// modify standard transformer feature table source for backward compatibility
	standardTransformer, err = k.updateFeatureTableSource(standardTransformer, stdTransformerConfig)
	if err != nil {
		return nil, err
	}

	return standardTransformer, nil
}

func (k *endpointService) updateFeatureTableSource(standardTransformer *models.Transformer, standardTransformerConfig *spec.StandardTransformerConfig) (*models.Transformer, error) {
	sourceFromServingURL := make(map[string]spec.ServingSource)
	if bigTableCfg := k.standardTransformerConfig.FeastBigtableConfig; bigTableCfg != nil {
		sourceFromServingURL[bigTableCfg.ServingURL] = spec.ServingSource_BIGTABLE
	}
	if redisCfg := k.standardTransformerConfig.FeastRedisConfig; redisCfg != nil {
		sourceFromServingURL[redisCfg.ServingURL] = spec.ServingSource_REDIS
	}

	feast.UpdateFeatureTableSource(standardTransformerConfig, sourceFromServingURL, k.standardTransformerConfig.DefaultFeastSource)

	marshaler := jsonpb.Marshaler{}
	transformerCfgString, err := marshaler.MarshalToString(standardTransformerConfig)
	if err != nil {
		return nil, err
	}
	envVars := standardTransformer.EnvVars
	envVars = models.MergeEnvVars(envVars, models.EnvVars{
		{
			Name:  transformer.StandardTransformerConfigEnvName,
			Value: transformerCfgString,
		},
	})
	standardTransformer.EnvVars = envVars
	return standardTransformer, nil
}

func (k *endpointService) addFeatureTableMetadata(standardTransformer *models.Transformer, standardTransformerConfig *spec.StandardTransformerConfig) (*models.Transformer, error) {
	featureTableSpecs, err := feast.GetAllFeatureTableMetadata(context.TODO(), k.feastCoreClient, standardTransformerConfig)
	if err != nil {
		return nil, err
	}
	// early return if feature table spec is empty
	if len(featureTableSpecs) == 0 {
		return standardTransformer, nil
	}
	featureTableSpecsJSON, err := json.Marshal(featureTableSpecs)
	if err != nil {
		return nil, err
	}

	envVars := standardTransformer.EnvVars
	// to make sure doesn't create duplicate env vars with the same name
	envVars = models.MergeEnvVars(envVars, models.EnvVars{
		{
			Name:  transformer.FeastFeatureTableSpecsJSON,
			Value: string(featureTableSpecsJSON),
		},
	})
	standardTransformer.EnvVars = envVars
	return standardTransformer, nil
}
