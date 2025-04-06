package server

import (
	"context"

	"github.com/taonic/ticketfu/worker"
	"github.com/taonic/ticketfu/worker/webhook"
	"go.temporal.io/sdk/client"
	"go.temporal.io/server/common/log/tag"
)

const WebhookWorkflow = "WebhookWorkflow"

// BootstrapZendeskWebhook invokes a workflow for creating Zendesk webhook
// It's conditioned on whether ZendeskWebhookBaseURL is configured.
// It also checks if the webhook exists before creating a new one.
func (s *Server) BootstrapZendeskWebhook(ctx context.Context) error {
	if len(s.config.ZendeskWebhookBaseURL) == 0 {
		s.logger.Debug("Skipping bootstraping webhook as webhook endpoint is not configured.")
		return nil
	}

	s.logger.Info("Creating Zendesk webhook idempotently", tag.WorkflowID(WebhookWorkflow))
	workflowOptions := client.StartWorkflowOptions{
		ID:        WebhookWorkflow,
		TaskQueue: worker.TaskQueue,
	}
	hook := webhook.Webhook{
		BaseURL:        s.config.ZendeskWebhookBaseURL,
		ServerAPIToken: s.config.APIToken,
	}
	upsertInput := webhook.UpsertWebhookInput{
		Webhook: hook,
	}
	_, err := s.temporalClient.SignalWithStartWorkflow(ctx,
		WebhookWorkflow,
		webhook.UpsertWebhookSignal,
		upsertInput,
		workflowOptions,
		webhook.WebhookWorkflow,
		hook,
	)

	if err != nil {
		s.logger.Debug("Failed to start the workflow to create Zendesk webhook", tag.Error(err))
		return err
	}
	return nil
}
