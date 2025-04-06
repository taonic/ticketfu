package webhook

import (
	"context"
	"fmt"

	"github.com/nukosuke/go-zendesk/zendesk"
)

type (
	CreateWebhookInput struct {
		Webhook Webhook
	}

	CreateWebhookOutput struct {
		WebhookID string
	}
)

const (
	WebhookName = "TicketFu Webhook"
)

func (a *Activity) CreateWebhook(ctx context.Context, input CreateWebhookInput) (*CreateWebhookOutput, error) {
	if input.Webhook.ID != "" {
		// check if it still exists on Zendesk
		webhook, err := a.getWebhook(ctx, input.Webhook.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get webhook from Zendesk %w", err)
		}

		if webhook != nil {
			return &CreateWebhookOutput{WebhookID: webhook.ID}, nil
		}
	}

	webhook := zendesk.Webhook{
		Name:          WebhookName,
		Status:        "active",
		Endpoint:      fmt.Sprintf("%s/api/v1/ticket", input.Webhook.BaseURL),
		HTTPMethod:    "POST",
		RequestFormat: "json",
		Authentication: &zendesk.WebhookAuthentication{
			Type:        "api_key",
			AddPosition: "header",
			Data: map[string]string{
				"name":  "X-Ticketfu-Key",
				"value": input.Webhook.ServerAPIToken,
			},
		},
	}

	createdWebhook, err := a.zClient.CreateWebhook(ctx, &webhook)
	if err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}

	output := CreateWebhookOutput{
		WebhookID: createdWebhook.ID,
	}

	return &output, nil
}

func (a *Activity) getWebhook(ctx context.Context, id string) (*zendesk.Webhook, error) {
	webhook, err := a.zClient.GetWebhook(ctx, id)
	if err != nil {
		if zerr, ok := err.(zendesk.Error); ok {
			if zerr.Status() == 404 {
				return nil, nil
			}
		}
		return nil, err
	}

	return webhook, nil
}
