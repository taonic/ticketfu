package ticket

import (
	"fmt"
	"time"

	"github.com/taonic/ticketfu/zendesk"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// UpsertTicketInput is the input for the summarize ticket workflow
type UpsertTicketInput struct {
	TicketID string
}

type UpdateTicketOutput struct {
	Summary string
}

type ticketWorkflow struct {
	workflow.Context
	signalCh                   workflow.ReceiveChannel
	updatesBeforeContinueAsNew int
	activityOptions            workflow.ActivityOptions

	// Ticket state
	ticket zendesk.Ticket
}

func newTicketWorkflow(ctx workflow.Context, ticket zendesk.Ticket) *ticketWorkflow {
	return &ticketWorkflow{
		Context: workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			StartToCloseTimeout: 30 * time.Second,
			RetryPolicy: &temporal.RetryPolicy{
				InitialInterval:    time.Second,
				BackoffCoefficient: 2.0,
				MaximumInterval:    time.Minute,
				MaximumAttempts:    10,
			},
		}),
		signalCh:                   workflow.GetSignalChannel(ctx, UpsertSignal),
		updatesBeforeContinueAsNew: 500,
		ticket:                     ticket,
	}
}

// Define the workflow
func TicketWorkflow(ctx workflow.Context, ticket zendesk.Ticket) error {
	t := newTicketWorkflow(ctx, ticket)
	return t.run()
}

func (s *ticketWorkflow) run() error {
	selector := workflow.NewSelector(s)

	// Listen for cancelled
	var cancelled bool
	selector.AddReceive(s.Done(), func(workflow.ReceiveChannel, bool) {
		cancelled = true
	})

	// Listen for upsert signals
	var updateCount int
	var pendingUpsert *UpsertTicketInput
	selector.AddReceive(s.signalCh, func(ch workflow.ReceiveChannel, _ bool) {
		updateCount++
		ch.Receive(s.Context, &pendingUpsert)
	})

	// Continually select until there are too many requests and no pending
	// selects.
	//
	// The reason we check selector.HasPending even when we've reached the request
	// limit is to make sure no events get lost. HasPending will continually
	// return true while an unresolved future or a buffered signal exists. If, for
	// example, we did not check this and there was an unhandled signal buffered
	// locally, continue-as-new would be returned without it being handled and the
	// new workflow wouldn't get the signal either. So it'd be lost.
	for updateCount < s.updatesBeforeContinueAsNew || selector.HasPending() {
		selector.Select(s)

		if pendingUpsert != nil {
			s.processPendingUpsert(pendingUpsert)
			pendingUpsert = nil
		}

		fmt.Println(pendingUpsert)

		if cancelled {
			return temporal.NewCanceledError()
		}
	}

	return workflow.NewContinueAsNewError(s, TicketWorkflow, s.ticket)
}

func (s *ticketWorkflow) processPendingUpsert(pendingUpsert *UpsertTicketInput) {
	a := Activity{}

	// fetch ticket if it hasn't been assigned
	if s.ticket.ID == 0 {
		workflow.ExecuteActivity(s.Context, a.FetchTicket, pendingUpsert.TicketID).Get(s.Context, &s.ticket)
	}

	// fetch comments with cursor
	fetchCommentsResponse := zendesk.FetchCommentsResponse{}
	workflow.ExecuteActivity(s.Context, a.FetchComments, pendingUpsert.TicketID, s.ticket.AfterCursor).Get(s.Context, &fetchCommentsResponse)
	if len(fetchCommentsResponse.Comments) != 0 {
		s.ticket.Comments = fetchCommentsResponse.Comments
		s.ticket.AfterCursor = fetchCommentsResponse.AfterCursor
	}
}
