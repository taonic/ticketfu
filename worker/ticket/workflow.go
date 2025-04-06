package ticket

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	UpsertTicketSignal       = "upsert-ticket-signal"
	QueryTicketSummary       = "query-ticket-summary"
	TicketWorkflowIDTemplate = "ticket-workflow-%s" // e.g. ticket-workflow-1234 where 1234 is the ticket ID
)

type Ticket struct {
	ID               int64
	Subject          string
	Description      string
	Priority         string
	Status           string
	Requester        string
	Assignee         string
	OrganizationID   int64
	OrganizationName string
	CreatedAt        *time.Time
	UpdatedAt        *time.Time

	// Comments and cursor
	Comments   []string
	NextCursor string

	// LLM generated summary
	Summary string
}

type (
	UpsertTicketInput struct {
		TicketID string
	}

	QueryTicketOutput struct {
		Summary string `json:"summary"`
	}

	ticketWorkflow struct {
		workflow.Context
		signalCh                   workflow.ReceiveChannel
		updatesBeforeContinueAsNew int
		activity                   Activity

		// Ticket state
		ticket Ticket
	}
)

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
		ch.Receive(s.Context, &pendingUpsert)
	})

	// Set query summary handler
	if err := workflow.SetQueryHandler(s.Context, QueryTicketSummary, s.handleQuerySummary); err != nil {
		return err
	}

	// Continually select until there are too many requests and no pending
	// selects.
	for updateCount < s.updatesBeforeContinueAsNew || selector.HasPending() {
		selector.Select(s)

		if pendingUpsert != nil {
			if err := s.processPendingUpsert(pendingUpsert); err != nil {
				return err
			}
			pendingUpsert = nil
			updateCount++
		}

		if cancelled {
			return temporal.NewCanceledError()
		}
	}

	return workflow.NewContinueAsNewError(s, TicketWorkflow, s.ticket)
}

func (s *ticketWorkflow) processPendingUpsert(pendingUpsert *UpsertTicketInput) error {
	// fetch ticket if it hasn't been assigned
	if s.ticket.ID == 0 {
		fetchTicketInput := FetchTicketInput{ID: pendingUpsert.TicketID}
		fetchTicketOutput := FetchTicketOutput{}

		if err := workflow.ExecuteActivity(s.Context, s.activity.FetchTicket, fetchTicketInput).
			Get(s.Context, &fetchTicketOutput); err != nil {
			return err
		}

		s.ticket = fetchTicketOutput.Ticket
	}

	// fetch comments with the cursor
	fetchCommentsInput := FetchCommentsInput{ID: pendingUpsert.TicketID, Cursor: s.ticket.NextCursor}
	fetchCommentsOutput := FetchCommentsOutput{}

	if err := workflow.ExecuteActivity(s.Context, s.activity.FetchComments, fetchCommentsInput).
		Get(s.Context, &fetchCommentsOutput); err != nil {
		return err
	}

	if len(fetchCommentsOutput.Comments) != 0 {
		s.ticket.Comments = fetchCommentsOutput.Comments
		s.ticket.NextCursor = fetchCommentsOutput.NextCursor
	}

	// gen summary
	genSummaryInput := GenSummaryInput{Ticket: s.ticket}
	genSummaryOutput := GenSummaryOutput{}

	if err := workflow.ExecuteActivity(s.Context, s.activity.GenTicketSummary, genSummaryInput).
		Get(s.Context, &genSummaryOutput); err != nil {
		return err
	}

	if genSummaryOutput.Summary != "" {
		s.ticket.Summary = genSummaryOutput.Summary
	}

	// signal organization
	if s.ticket.OrganizationID != 0 {
		signalOrganizationInput := SignalOrganizationInput{
			OrganizationID: s.ticket.OrganizationID,
			TicketID:       s.ticket.ID,
			TicketSummary:  s.ticket.Summary,
		}

		if err := workflow.ExecuteActivity(s.Context, s.activity.SignalOrganization, signalOrganizationInput).
			Get(s.Context, nil); err != nil {
			return err
		}
	}

	return nil
}

func (s *ticketWorkflow) handleQuerySummary() (QueryTicketOutput, error) {
	return QueryTicketOutput{Summary: s.ticket.Summary}, nil
}
