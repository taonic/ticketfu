package ticket

import (
	"context"
	"fmt"

	"github.com/taonic/ticketfu/worker/org"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
)

type SignalOrganizationInput struct {
	OrganizationID int64
	TicketID       int64
	TicketSummary  string
}

type UpdateOrganizationSignal struct {
	OrganizationID int64
	TicketID       int64
	TicketSummary  string
}

func (a *Activity) SignalOrganization(ctx context.Context, input SignalOrganizationInput) error {
	workflowID := fmt.Sprintf(org.OrganizationWorkflowIDTemplate, input.OrganizationID)

	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: activity.GetInfo(ctx).TaskQueue,
	}

	signalPayload := UpdateOrganizationSignal{
		OrganizationID: input.OrganizationID,
		TicketID:       input.TicketID,
		TicketSummary:  input.TicketSummary,
	}

	_, err := a.tClient.SignalWithStartWorkflow(ctx,
		workflowID,
		org.UpsertOrganizationSignal,
		signalPayload,
		workflowOptions,
		org.OrganizationWorkflow,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to signal org workflow: %w", err)
	}

	return nil
}
