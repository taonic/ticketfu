package ticket

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	TaskQueue = "ticket-queue"

	UpsertTicketSignal = "upsert-ticket-signal"
	QueryTicketSummary = "query-ticket-summary"

	TicketWorkflowIDTemplate = "ticket-workflow-%s" // e.g. ticket-workflow-1234 where 1234 is the ticket ID

	QueryOrgSummary        = "QueryOrgSummary"
	UpdateOrgSummarySignal = "UpdateOrgSummary"

	OrgWorkflowIDTemplate = "summarize-org-%s-%s" // <account-id>-<ticket-id>
)

type Ticket struct {
	ID           int64
	Subject      string
	Description  string
	Priority     string
	Status       string
	Requester    string
	Assignee     string
	Organization string
	CreatedAt    *time.Time
	UpdatedAt    *time.Time

	// Comments and cursor
	Comments   []string
	NextCursor string

	// LLM generated summary
	Summary string
}

// UpsertTicketInput is the input for the summarize ticket workflow
type UpsertTicketInput struct {
	TicketID string
}

type ticketWorkflow struct {
	workflow.Context
	signalCh                   workflow.ReceiveChannel
	updatesBeforeContinueAsNew int
	activityOptions            workflow.ActivityOptions
	activity                   Activity

	// Ticket state
	ticket Ticket
}

func newTicketWorkflow(ctx workflow.Context, ticket Ticket) *ticketWorkflow {
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
		signalCh:                   workflow.GetSignalChannel(ctx, UpsertTicketSignal),
		updatesBeforeContinueAsNew: 500,
		ticket:                     ticket,
	}
}

// Define the workflow
func TicketWorkflow(ctx workflow.Context, ticket Ticket) error {
	t := newTicketWorkflow(ctx, ticket)
	return t.run()
}

func (s *ticketWorkflow) run() error {
	selector := workflow.NewSelector(s)

	// Listen for cancellation
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

	// Set query summary handler
	if err := workflow.SetQueryHandler(s.Context, QueryTicketSummary, s.handleQuerySummary); err != nil {
		return err
	}

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
	// fetch ticket if it hasn't been assigned
	fetchTicketInput := FetchTicketInput{ID: pendingUpsert.TicketID}
	fetchTicketOutput := FetchTicketOutput{}

	if s.ticket.ID == 0 {
		workflow.ExecuteActivity(s.Context, s.activity.FetchTicket, fetchTicketInput).
			Get(s.Context, &fetchTicketOutput)
	}

	s.ticket = fetchTicketOutput.Ticket

	// fetch comments with the cursor
	fetchCommentsInput := FetchCommentsInput{ID: pendingUpsert.TicketID, Cursor: s.ticket.NextCursor}
	fetchCommentsOutput := FetchCommentsOutput{}

	workflow.ExecuteActivity(s.Context, s.activity.FetchComments, fetchCommentsInput).
		Get(s.Context, &fetchCommentsOutput)

	if len(fetchCommentsOutput.Comments) != 0 {
		s.ticket.Comments = fetchCommentsOutput.Comments
		s.ticket.NextCursor = fetchCommentsOutput.NextCursor
	}

	// gen summary
	genSummaryInput := GenSummaryInput{Ticket: s.ticket}
	genSummaryOutput := GenSummaryOutput{}

	workflow.ExecuteActivity(s.Context, s.activity.GenSummary, genSummaryInput).
		Get(s.Context, &genSummaryOutput)

	if genSummaryOutput.Summary != "" {
		s.ticket.Summary = genSummaryOutput.Summary
	}
}

func (s *ticketWorkflow) handleQuerySummary() (string, error) {
	return s.ticket.Summary, nil
}
