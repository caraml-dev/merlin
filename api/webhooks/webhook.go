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

type Request struct {
	EventType webhooks.EventType `json:"event_type"`

	Data interface{} `json:"data"`
}

type VersionEndpointData struct {
	VersionEndpoint *models.VersionEndpoint `json:"version_endpoint"`
}

func BuildRequest(eventType webhooks.EventType, d interface{}) *Request {
	return &Request{
		EventType: eventType,
		Data:      d,
	}
}
