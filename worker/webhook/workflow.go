package webhook

import (
	"fmt"
	"time"

	sdklog "go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.temporal.io/server/common/log/tag"
)

type (
	Webhook struct {
		ID             string
		BaseURL        string
		ServerAPIToken string
	}

	webhookWorkflow struct {
		workflow.Context
		activity     Activity
		logger       sdklog.Logger
		upsertSignal workflow.ReceiveChannel
		selector     workflow.Selector
		upsertCount  int
		cancelled    bool

		// webhook is the entity maintained by the workflow
		webhook Webhook
	}

	UpsertWebhookInput struct {
		Webhook Webhook
	}
)

const (
	UpsertWebhookSignal = "UpsertWebhookSignal"
)

var (
	upsertBeforeCAN = 500
)

func newWebhookWorkflow(ctx workflow.Context, webhook Webhook) *webhookWorkflow {
	return &webhookWorkflow{
		Context: workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			StartToCloseTimeout: 30 * time.Second,
		}),
		logger:       sdklog.With(workflow.GetLogger(ctx)),
		selector:     workflow.NewSelector(ctx),
		upsertSignal: workflow.GetSignalChannel(ctx, UpsertWebhookSignal),
		webhook:      webhook,
	}
}

func WebhookWorkflow(ctx workflow.Context, webhook Webhook) error {
	t := newWebhookWorkflow(ctx, webhook)
	return t.run()
}

func (s *webhookWorkflow) run() error {
	// Listen for cancellation
	s.selector.AddReceive(s.Done(), func(workflow.ReceiveChannel, bool) {
		s.cancelled = true
	})

	// Listen for upsert signals
	var pendingUpsert *UpsertWebhookInput
	s.selector.AddReceive(s.upsertSignal, func(ch workflow.ReceiveChannel, _ bool) {
		ch.Receive(s.Context, &pendingUpsert)
	})

	// Keep the workflow running to receive further upserts
	for s.upsertCount < upsertBeforeCAN || s.selector.HasPending() {
		s.selector.Select(s)

		if pendingUpsert != nil {
			s.processPendingUpsert(pendingUpsert)
			pendingUpsert = nil
			s.upsertCount++
		}

		if s.cancelled {
			return temporal.NewCanceledError()
		}
	}

	return workflow.NewContinueAsNewError(s, WebhookWorkflow, s.webhook)
}

func (s *webhookWorkflow) processPendingUpsert(upsert *UpsertWebhookInput) error {
	s.logger.Debug("Processing upsert signal", tag.Value(upsert.Webhook))

	// Create Zendesk webhook
	createWebhookInput := CreateWebhookInput{
		Webhook: upsert.Webhook,
	}

	// Assign the webhook ID here to make the activity idempotent:
	// Check if the wehbook still exist before creating a new one.
	createWebhookInput.Webhook.ID = s.webhook.ID

	var createWebhookOutput CreateWebhookOutput
	err := workflow.ExecuteActivity(s.Context, s.activity.CreateWebhook, createWebhookInput).
		Get(s.Context, &createWebhookOutput)
	if err != nil {
		return fmt.Errorf("failed to create webhook %w", err)
	}

	// Only create trigger if new webhook is created
	if createWebhookOutput.WebhookID != s.webhook.ID {
		s.webhook.ID = createWebhookOutput.WebhookID
		s.logger.Debug("Created a new webhook with ID", tag.Value(createWebhookOutput.WebhookID))

		// Create Zendesk trigger based on the webhook ID
		createTriggerInput := CreateTriggerInput{
			WebhookID: s.webhook.ID,
		}
		var createTriggerOutput CreateTriggerOutput
		err = workflow.ExecuteActivity(s.Context, s.activity.CreateTrigger, createTriggerInput).
			Get(s.Context, &createTriggerOutput)
		if err != nil {
			return fmt.Errorf("failed to create trigger %w", err)
		}

		s.logger.Debug("Created trigger with ID", tag.Value(createTriggerOutput.TriggerID))
	} else {
		s.logger.Debug("Skipping trigger creation as webhook already exist", tag.Value(createWebhookOutput.WebhookID))
	}

	return nil
}
