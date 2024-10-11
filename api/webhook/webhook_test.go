package webhook

import (
	"context"
	"fmt"
	"testing"

	"github.com/caraml-dev/merlin/models"
	webhookManager "github.com/caraml-dev/mlp/api/pkg/webhooks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTriggerWebhooks(t *testing.T) {
	defaultModel := &models.Model{
		ID:        2,
		ProjectID: 1,
		Name:      "my-model",
	}

	testCases := []struct {
		name      string
		event     webhookManager.EventType
		webhook   func() *webhookManager.MockWebhookManager
		call      func(w *Webhook) error
		wantError bool
	}{
		{
			name:  "Success: event is configured",
			event: OnModelCreated,
			webhook: func() *webhookManager.MockWebhookManager {
				w := webhookManager.NewMockWebhookManager(t)
				w.On("IsEventConfigured", OnModelCreated).Return(true)
				w.On("InvokeWebhooks", mock.Anything, OnModelCreated, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				return w
			},
			wantError: false,
		},
		{
			name:  "Success: event is not configured",
			event: OnModelCreated,
			webhook: func() *webhookManager.MockWebhookManager {
				w := webhookManager.NewMockWebhookManager(t)
				w.On("IsEventConfigured", OnModelCreated).Return(false)
				return w
			},
			wantError: false,
		},
		{
			name:  "Success: with body set up",
			event: OnModelCreated,
			webhook: func() *webhookManager.MockWebhookManager {
				body := &WebhookRequest{
					EventType: OnModelCreated,
					Data: map[string]interface{}{
						"model": defaultModel,
					},
				}
				w := webhookManager.NewMockWebhookManager(t)
				w.On("IsEventConfigured", OnModelCreated).Return(true)
				w.On("InvokeWebhooks", mock.Anything, OnModelCreated, body, mock.Anything, mock.Anything).Return(nil)
				return w
			},
			call: func(w *Webhook) error {
				return w.TriggerWebhooks(context.Background(), OnModelCreated, SetBody(defaultModel))
			},
			wantError: false,
		},
		{
			name:  "Fail: there was a webhook error",
			event: OnModelCreated,
			webhook: func() *webhookManager.MockWebhookManager {
				w := webhookManager.NewMockWebhookManager(t)
				w.On("IsEventConfigured", OnModelCreated).Return(true)
				w.On("InvokeWebhooks", mock.Anything, OnModelCreated, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("there was an error"))
				return w
			},
			wantError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := &Webhook{
				webhookManager: tc.webhook(),
			}

			var err error
			if tc.call != nil {
				err = tc.call(w)
			} else {
				err = w.TriggerWebhooks(context.Background(), tc.event)
			}

			if tc.wantError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}
