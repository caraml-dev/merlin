package service

import (
	"github.com/gojek/merlin/mlp"
	"github.com/gojek/merlin/models"
	"github.com/gojek/merlin/pkg/protocol"
	"github.com/google/uuid"
)

var (
	uuid1, _  = uuid.NewUUID()
	isDefault = true
	env       = &models.Environment{
		Name:      "env1",
		Cluster:   "cluster1",
		IsDefault: &isDefault,
	}
	labels = mlp.Labels{
		{
			Key:   "sample",
			Value: "true",
		},
	}
	model1 = &models.Model{
		ID:        1,
		ProjectID: 1,
		Project: mlp.Project{
			Id:     1,
			Name:   "project-1",
			Team:   "dsp",
			Stream: "dsp",
			Labels: labels,
		},
		ExperimentID: 1,

		Name: "model-1",
		Type: models.ModelTypeTensorflow,
	}

	versionEndpoint1 = &models.VersionEndpoint{
		ID:                   uuid1,
		Status:               models.EndpointRunning,
		URL:                  "http://version-1.project-1.mlp.io/v1/models/version-1:predict",
		ServiceName:          "version-1-abcde",
		InferenceServiceName: "version-1",
		Namespace:            "project-1",
		Protocol:             protocol.HttpJson,
	}

	modelEndpointRequest1 = &models.ModelEndpoint{
		ModelID: 1,
		Rule: &models.ModelEndpointRule{
			Destination: []*models.ModelEndpointRuleDestination{
				{
					VersionEndpointID: uuid1,
					VersionEndpoint:   versionEndpoint1,
					Weight:            int32(100),
				},
			},
		},
		EnvironmentName: env.Name,
	}

	modelEndpointRequestWrongEnvironment = &models.ModelEndpoint{
		ModelID: 1,
		Rule: &models.ModelEndpointRule{
			Destination: []*models.ModelEndpointRuleDestination{
				{
					VersionEndpointID: uuid1,
					VersionEndpoint:   versionEndpoint1,
					Weight:            int32(100),
				},
			},
		},
		EnvironmentName: "staging-env",
	}

	modelEndpointResponse1 = &models.ModelEndpoint{
		ModelID: 1,
		URL:     "model-1.project-1.mlp.io",
		Status:  models.EndpointServing,
		Rule: &models.ModelEndpointRule{
			Destination: []*models.ModelEndpointRuleDestination{
				&models.ModelEndpointRuleDestination{
					VersionEndpointID: uuid1,
					VersionEndpoint:   versionEndpoint1,
					Weight:            int32(100),
				},
			},
		},
		EnvironmentName: env.Name,
	}
)
