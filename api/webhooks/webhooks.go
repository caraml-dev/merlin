package webhooks

import (
	"github.com/caraml-dev/mlp/api/pkg/webhooks"

	"github.com/caraml-dev/merlin/models"
)

var (
	OnModelVersionPredeployment = webhooks.EventType("on-model-version-predeployment")
	OnModelVersionDeployed      = webhooks.EventType("on-model-version-deployed")
	OnModelVersionUndeployed    = webhooks.EventType("on-model-version-undeployed")
)

var WebhookEvents = []webhooks.EventType{
	OnModelVersionPredeployment,
	OnModelVersionDeployed,
	OnModelVersionUndeployed,
}

type VersionEndpointRequest struct {
	EventType       webhooks.EventType      `json:"event_type"`
	VersionEndpoint *models.VersionEndpoint `json:"version_endpoint"`
}
