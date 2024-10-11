package webhook

import (
	"context"

	"github.com/caraml-dev/merlin/log"
	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/mlp/api/pkg/webhooks"
	webhookManager "github.com/caraml-dev/mlp/api/pkg/webhooks"
)

var (
	OnModelCreated = webhooks.EventType("on-model-created")

	OnModelEndpointCreated = webhooks.EventType("on-model-endpoint-created")
	OnModelEndpointUpdated = webhooks.EventType("on-model-endpoint-updated")
	OnModelEndpointDeleted = webhooks.EventType("on-model-endpoint-deleted")

	OnModelVersionCreated = webhooks.EventType("on-model-version-created")
	OnModelVersionUpdated = webhooks.EventType("on-model-version-updated")
	OnModelVersionDeleted = webhooks.EventType("on-model-version-deleted")

	OnVersionEndpointPredeployment = webhooks.EventType("on-version-endpoint-predeployment")
	OnVersionEndpointDeployed      = webhooks.EventType("on-version-endpoint-deployed")
	OnVersionEndpointUndeployed    = webhooks.EventType("on-version-endpoint-undeployed")
)

var events = []webhooks.EventType{
	OnModelCreated,
	OnModelEndpointCreated,
	OnModelEndpointUpdated,
	OnModelEndpointDeleted,
	OnModelVersionCreated,
	OnModelVersionUpdated,
	OnModelVersionDeleted,
	OnVersionEndpointPredeployment,
	OnVersionEndpointDeployed,
	OnVersionEndpointUndeployed,
}

type Webhook struct {
	webhookManager webhookManager.WebhookManager
}

type Option func(*config)

type config struct {
	successCallback func(payload []byte) error
	errorCallback   func(error) error
	body            map[string]interface{}
}

type Client interface {
	TriggerWebhooks(ctx context.Context, event webhooks.EventType, opts ...Option) error
}

func NewWebhook(cfg *webhookManager.Config) *Webhook {
	manager, err := webhookManager.InitializeWebhooks(cfg, events)
	if err != nil {
		log.Panicf("failed to initialize webhook: %s", err)
	}

	return &Webhook{
		webhookManager: manager,
	}
}

func newDefaultOption() *config {
	return &config{
		successCallback: webhooks.NoOpCallback,
		errorCallback:   webhooks.NoOpErrorHandler,
	}
}

func (w Webhook) TriggerWebhooks(ctx context.Context, event webhooks.EventType, opts ...Option) error {
	if !w.isEventConfigured(event) {
		return nil
	}

	conf := newDefaultOption()
	for _, opt := range opts {
		opt(conf)
	}

	b := &WebhookRequest{
		EventType: event,
		Data:      conf.body,
	}

	return w.webhookManager.InvokeWebhooks(ctx, event, b, conf.successCallback, conf.errorCallback)
}

func (w Webhook) isEventConfigured(event webhooks.EventType) bool {
	return w.webhookManager != nil && w.webhookManager.IsEventConfigured(event)
}

func SetSuccessCallBack(f func(payload []byte) error) Option {
	return func(c *config) {
		c.successCallback = f
	}
}

func SetErrorCallback(f func(error) error) Option {
	return func(c *config) {
		c.errorCallback = f
	}
}

func SetBody(items ...interface{}) Option {
	body := make(map[string]interface{})

	for _, item := range items {
		switch item.(type) {
		case *models.Model:
			body["model"] = item
		case *models.Version:
			body["model_version"] = item
		case *models.ModelEndpoint:
			body["model_endpoint"] = item
		case *models.VersionEndpoint:
			body["version_endpoint"] = item
		}
	}

	return func(c *config) {
		c.body = body
	}
}
