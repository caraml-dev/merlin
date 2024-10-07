package webhook

import (
	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/mlp/api/pkg/webhooks"
)

type VersionRequest struct {
	EventType webhooks.EventType `json:"event_type"`
	Data      interface{}        `json:"data"`
}

type VersionEndpointRequest struct {
	EventType       webhooks.EventType      `json:"event_type"`
	VersionEndpoint *models.VersionEndpoint `json:"version_endpoint"`
}

type ModelVersionRequest struct {
	EventType webhooks.EventType `json:"event_type"`
	Version   *models.Version    `json:"version"`
}

type ModelEndpointRequest struct {
	EventType     webhooks.EventType    `json:"event_type"`
	ModelEndpoint *models.ModelEndpoint `json:"model_endpoint"`
}

type ModelRequest struct {
	EventType webhooks.EventType `json:"event_type"`
	Model     *models.Model      `json:"model"`
}
