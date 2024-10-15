package webhook

import (
	"github.com/caraml-dev/mlp/api/pkg/webhooks"
)

type WebhookRequest struct {
	EventType webhooks.EventType     `json:"event_type"`
	Data      map[string]interface{} `json:"data"`
}
