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

var Events = []webhooks.EventType{
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
	WebhookManager webhookManager.WebhookManager
}

type Client interface {
	TriggerModelEvent(ctx context.Context, event webhooks.EventType, model *models.Model) (err error)
	TriggerModelVersionEvent(ctx context.Context, event webhooks.EventType, version *models.Version) (err error)
	TriggerModelEndpointEvent(ctx context.Context, event webhooks.EventType, modelEndpoint *models.ModelEndpoint) (err error)
	TriggerVersionEndpointEvent(ctx context.Context, event webhooks.EventType, versionEndpoint *models.VersionEndpoint) (err error)
}

func NewWebhook(cfg *webhookManager.Config) *Webhook {
	manager, err := webhookManager.InitializeWebhooks(cfg, Events)
	if err != nil {
		log.Panicf("failed to initialize webhook: %s", err)
	}

	return &Webhook{
		WebhookManager: manager,
	}
}

func (w Webhook) TriggerModelEvent(ctx context.Context, event webhooks.EventType, model *models.Model) (err error) {
	if !w.isEventConfigured(event) {
		return
	}

	body := &ModelRequest{
		EventType: event,
		Model:     model,
	}

	err = w.WebhookManager.InvokeWebhooks(ctx, event, body, webhookManager.NoOpCallback, webhookManager.NoOpErrorHandler)
	if err != nil {
		log.Warnf("unable to invoke webhook for event type: %s, model: %s, error: %v", event, model.Name, err)
		return err
	}

	return nil
}

func (w Webhook) TriggerModelVersionEvent(ctx context.Context, event webhooks.EventType, version *models.Version) (err error) {
	if !w.isEventConfigured(event) {
		return
	}

	body := &ModelVersionRequest{
		EventType: event,
		Version:   version,
	}

	err = w.WebhookManager.InvokeWebhooks(ctx, event, body, webhookManager.NoOpCallback, webhookManager.NoOpErrorHandler)
	if err != nil {
		log.Warnf("unable to invoke webhook for event type: %s, model: %s, version: %d, error: %v", event, version.ModelID, version.ID, err)
		return err
	}

	return nil
}

func (w Webhook) TriggerModelEndpointEvent(ctx context.Context, event webhooks.EventType, modelEndpoint *models.ModelEndpoint) (err error) {
	if !w.isEventConfigured(event) {
		return
	}

	body := &ModelEndpointRequest{
		EventType:     event,
		ModelEndpoint: modelEndpoint,
	}

	err = w.WebhookManager.InvokeWebhooks(ctx, event, body, webhookManager.NoOpCallback, webhookManager.NoOpErrorHandler)
	if err != nil {
		log.Warnf("unable to invoke webhook for event type: %s, model id: %s, endpoint: %s, error: %v", event, modelEndpoint.ModelID, modelEndpoint.ID, err)
		return err
	}

	return nil
}

func (w Webhook) TriggerVersionEndpointEvent(ctx context.Context, event webhooks.EventType, versionEndpoint *models.VersionEndpoint) (err error) {
	if !w.isEventConfigured(event) {
		return
	}

	body := &VersionEndpointRequest{
		EventType:       event,
		VersionEndpoint: versionEndpoint,
	}

	err = w.WebhookManager.InvokeWebhooks(ctx, event, body, webhookManager.NoOpCallback, webhookManager.NoOpErrorHandler)
	if err != nil {
		log.Warnf("unable to invoke webhook for event type: %s, model id: %s, version: %d, error: %v", event, versionEndpoint.VersionModelID, versionEndpoint.ID, err)
		return err
	}

	return nil
}

func (w Webhook) isEventConfigured(event webhooks.EventType) bool {
	return w.WebhookManager != nil && w.WebhookManager.IsEventConfigured(event)
}
