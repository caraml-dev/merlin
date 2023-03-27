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

	"github.com/caraml-dev/merlin/cluster"
	"github.com/caraml-dev/merlin/config"
	"github.com/caraml-dev/merlin/log"
	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/pkg/autoscaling"
	"github.com/caraml-dev/merlin/pkg/deployment"
	"github.com/caraml-dev/merlin/pkg/imagebuilder"
	"github.com/caraml-dev/merlin/pkg/protocol"
	"github.com/caraml-dev/merlin/pkg/transformer"
	"github.com/caraml-dev/merlin/pkg/transformer/feast"
	"github.com/caraml-dev/merlin/pkg/transformer/spec"
	"github.com/caraml-dev/merlin/queue"
	"github.com/caraml-dev/merlin/queue/work"
	"github.com/caraml-dev/merlin/storage"
	"github.com/feast-dev/feast/sdk/go/protos/feast/core"
	"github.com/google/uuid"
	"google.golang.org/protobuf/encoding/protojson"
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
	// get existing endpoint or create a new one with default config
	endpoint, _ := version.GetEndpointByEnvironmentName(environment.Name)
	if endpoint == nil {
		// create endpoint with default configurations
		endpoint = models.NewVersionEndpoint(environment, model.Project, model, version, k.monitoringConfig, newEndpoint.DeploymentMode)
	}

	// override existing endpoint configuration with the user request
	err := k.override(endpoint, newEndpoint, environment)
	if err != nil {
		return nil, err
	}

	// Copy to avoid race condition
	tobeDeployedEndpoint := *endpoint
	endpoint.Status = models.EndpointPending
	err = k.storage.Save(endpoint)
	if err != nil {
		return nil, err
	}

	if err := k.jobProducer.EnqueueJob(&queue.Job{
		Name: ModelServiceDeployment,
		Arguments: queue.Arguments{
			dataArgKey: work.EndpointJob{
				Endpoint: &tobeDeployedEndpoint,
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

// override left version endpoint with values on the right version endpoint
func (k *endpointService) override(left *models.VersionEndpoint, right *models.VersionEndpoint, environment *models.Environment) error {
	// override deployment mode
	if right.DeploymentMode != deployment.EmptyDeploymentMode {
		left.DeploymentMode = right.DeploymentMode
	}

	// override autoscaling policy
	if right.AutoscalingPolicy != nil {
		err := autoscaling.ValidateAutoscalingPolicy(right.AutoscalingPolicy, left.DeploymentMode)
		if err != nil {
			return err
		}
		left.AutoscalingPolicy = right.AutoscalingPolicy
	}

	// override resource request
	if right.ResourceRequest != nil {
		left.ResourceRequest = right.ResourceRequest
	}

	// override transformer config
	if right.Transformer != nil {
		if right.Transformer.ResourceRequest == nil {
			right.Transformer.ResourceRequest = environment.DefaultTransformerResourceRequest
		}

		if right.Transformer.TransformerType == models.DefaultTransformerType {
			right.Transformer.TransformerType = models.CustomTransformerType
		}

		if right.Transformer.TransformerType == models.StandardTransformerType {
			// update standard transformer config
			// 1. Add feature table metadata variables to transformer
			// 2. Update feature table source if empty
			updatedStandardTransformer, err := k.reconfigureStandardTransformer(right.Transformer)
			if err != nil {
				return err
			}
			left.Transformer = updatedStandardTransformer
		}

		left.Transformer = right.Transformer
		left.Transformer.VersionEndpointID = left.ID
	}

	// override logger
	if right.Logger != nil {
		left.Logger = right.Logger
		left.Logger.DestinationURL = k.loggerDestinationURL
		modelLogger := left.Logger.Model
		if modelLogger != nil {
			modelLogger.SanitizeMode()
			left.Logger.Model = modelLogger
		}
		transformerLogger := left.Logger.Transformer
		if transformerLogger != nil {
			transformerLogger.SanitizeMode()
			left.Logger.Transformer = transformerLogger
		}
	}

	// override env vars
	// Configure environment variables for Pyfunc model
	if len(right.EnvVars) > 0 {
		left.EnvVars = models.MergeEnvVars(left.EnvVars, right.EnvVars)
	}

	// default to HttpJson if not provided
	if right.Protocol == "" {
		right.Protocol = protocol.HttpJson
	}

	// override protocol
	left.Protocol = right.Protocol

	return nil
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
	err := protojson.Unmarshal([]byte(stdTransformerConfigString), stdTransformerConfig)
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

	tc, err := protojson.Marshal(standardTransformerConfig)
	if err != nil {
		return nil, err
	}
	envVars := standardTransformer.EnvVars
	envVars = models.MergeEnvVars(envVars, models.EnvVars{
		{
			Name:  transformer.StandardTransformerConfigEnvName,
			Value: string(tc),
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
