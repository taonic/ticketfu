package webhook

import (
	"context"
	"fmt"

	"github.com/nukosuke/go-zendesk/zendesk"
)

type (
	CreateTriggerInput struct {
		WebhookID string
	}

	CreateTriggerOutput struct {
		TriggerID string
	}
)

func (a *Activity) CreateTrigger(ctx context.Context, input CreateTriggerInput) (*CreateTriggerOutput, error) {
	trigger := zendesk.Trigger{
		Title:    "Notify TicketFu",
		Active:   true,
		Position: 1,
		Actions: []zendesk.TriggerAction{
			{
				Field: "notification_webhook",
				Value: []interface{}{
					input.WebhookID,
					`{
						"ticket_url": "{{ticket.url}}"
					}`,
				},
			},
		},
		Description: "Used to notify TicketFu to generat ticket summary",
	}

	trigger.Conditions.Any = []zendesk.TriggerCondition{
		{
			Field:    "update_type",
			Operator: "is",
			Value:    "Create",
		},
		{
			Field:    "update_type",
			Operator: "is",
			Value:    "Change",
		},
	}

	createdTrigger, err := a.zClient.CreateTrigger(ctx, trigger)
	if err != nil {
		return nil, fmt.Errorf("failed to create trigger: %w", err)
	}

	output := CreateTriggerOutput{
		TriggerID: fmt.Sprintf("%d", createdTrigger.ID),
	}

	return &output, nil
}
