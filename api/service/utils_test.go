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
			ID:     1,
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
				{
					VersionEndpointID: uuid1,
					VersionEndpoint:   versionEndpoint1,
					Weight:            int32(100),
				},
			},
		},
		EnvironmentName: env.Name,
		Protocol:        protocol.HttpJson,
	}

	upiV1Model = &models.Model{
		ID:        2,
		ProjectID: 1,
		Project: mlp.Project{
			ID:     1,
			Name:   "project-1",
			Team:   "dsp",
			Stream: "dsp",
			Labels: labels,
		},
		ExperimentID: 1,

		Name: "model-upi",
		Type: models.ModelTypeTensorflow,
	}

	upiV1VersionEndpoint1UUID, _ = uuid.NewUUID()
	upiV1VersionEndpoint1        = &models.VersionEndpoint{
		ID:                   upiV1VersionEndpoint1UUID,
		Status:               models.EndpointRunning,
		URL:                  "model-upi-1.project-1.mlp.io",
		ServiceName:          "model-upi-1-abcde",
		InferenceServiceName: "model-upi-1",
		Namespace:            "project-1",
		Protocol:             protocol.UpiV1,
	}

	upiV1ModelEndpointRequest1 = &models.ModelEndpoint{
		ModelID: upiV1Model.ID,
		Rule: &models.ModelEndpointRule{
			Destination: []*models.ModelEndpointRuleDestination{
				{
					VersionEndpointID: upiV1VersionEndpoint1.ID,
					VersionEndpoint:   upiV1VersionEndpoint1,
					Weight:            int32(100),
				},
			},
		},
		EnvironmentName: env.Name,
	}

	upiV1ModelEndpointResponse1 = &models.ModelEndpoint{
		ModelID: upiV1Model.ID,
		URL:     "model-upi.project-1.mlp.io",
		Status:  models.EndpointServing,
		Rule: &models.ModelEndpointRule{
			Destination: []*models.ModelEndpointRuleDestination{
				{
					VersionEndpointID: upiV1VersionEndpoint1.ID,
					VersionEndpoint:   upiV1VersionEndpoint1,
					Weight:            int32(100),
				},
			},
		},
		EnvironmentName: env.Name,
		Protocol:        protocol.UpiV1,
	}
)
