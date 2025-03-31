package org

import (
	"time"

	"github.com/taonic/ticketfu/worker/util"
	sdklog "go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	UpsertOrganizationSignal       = "upsert-organization-signal"
	QueryOrganizationSummary       = "query-organization-summary"
	OrganizationWorkflowIDTemplate = "organization-workflow-%s" // e.g. organization-workflow-123
	MaxTicketSummaries             = 500
)

var (
	updatesBeforeContinueAsNew = 500
)

type (
	Organization struct {
		ID      int64
		Name    string
		Notes   string
		Details string

		TicketSummaries map[int64]string

		// LLM generated summary
		Summary string
	}

	UpsertOrganizationInput struct {
		OrganizationID int64
		TicketID       int64
		TicketSummary  string
	}

	QueryOrganizationOutput struct {
		Summary string `json:"summary"`
	}

	organizationWorkflow struct {
		workflow.Context
		logger                     sdklog.Logger
		signalCh                   workflow.ReceiveChannel
		updatesBeforeContinueAsNew int
		activityOptions            workflow.ActivityOptions
		activity                   Activity

		// Organization state
		organization Organization
	}
)

func newOrganizationWorkflow(ctx workflow.Context, organization Organization) *organizationWorkflow {
	return &organizationWorkflow{
		Context: workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			StartToCloseTimeout: 30 * time.Second,
			RetryPolicy: &temporal.RetryPolicy{
				InitialInterval:    time.Second,
				BackoffCoefficient: 2.0,
				MaximumInterval:    time.Minute,
				MaximumAttempts:    10,
			},
		}),
		logger:                     sdklog.With(workflow.GetLogger(ctx)),
		signalCh:                   workflow.GetSignalChannel(ctx, UpsertOrganizationSignal),
		updatesBeforeContinueAsNew: updatesBeforeContinueAsNew,
		organization:               organization,
	}
}

// Define the workflow
func OrganizationWorkflow(ctx workflow.Context, organization Organization) error {
	t := newOrganizationWorkflow(ctx, organization)
	return t.run()
}

func (s *organizationWorkflow) run() error {
	selector := workflow.NewSelector(s)

	// Listen for cancellation
	var cancelled bool
	selector.AddReceive(s.Done(), func(workflow.ReceiveChannel, bool) {
		cancelled = true
	})

	// Listen for upsert signals
	var updateCount int
	var pendingUpsert *UpsertOrganizationInput
	selector.AddReceive(s.signalCh, func(ch workflow.ReceiveChannel, _ bool) {
		ch.Receive(s.Context, &pendingUpsert)
	})

	// Set query summary handler
	if err := workflow.SetQueryHandler(s.Context, QueryOrganizationSummary, s.handleQuerySummary); err != nil {
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
			updateCount++
		}

		if cancelled {
			return temporal.NewCanceledError()
		}
	}

	return workflow.NewContinueAsNewError(s, OrganizationWorkflow, s.organization)
}

func (s *organizationWorkflow) processPendingUpsert(pendingUpsert *UpsertOrganizationInput) {
	// fetch organization if it hasn't been fetched
	if s.organization.Name == "" {
		fetchOrganizationInput := FetchOrganizationInput{ID: pendingUpsert.OrganizationID}
		fetchOrganizationOutput := FetchOrganizationOutput{}

		workflow.ExecuteActivity(s.Context, s.activity.FetchOrganization, fetchOrganizationInput).
			Get(s.Context, &fetchOrganizationOutput)

		s.organization = fetchOrganizationOutput.Organization
	}

	// Initialize ticket summary map
	if s.organization.TicketSummaries == nil {
		s.organization.TicketSummaries = make(map[int64]string)
	}

	// Generate summary if needed
	targetSummary, exist := s.organization.TicketSummaries[pendingUpsert.TicketID]
	if !exist || targetSummary != pendingUpsert.TicketSummary {
		s.logger.Debug("Updating org summary", "org-id", s.organization.ID, "ticket-id", pendingUpsert.TicketID)
		s.organization.TicketSummaries[pendingUpsert.TicketID] = pendingUpsert.TicketSummary

		// Truncate by keeping up to 500 most recent tickets
		// todo: make it configurable
		var truncated bool
		s.organization.TicketSummaries, truncated = util.TruncateStringMap(s.organization.TicketSummaries, MaxTicketSummaries)
		if truncated {
			s.logger.Debug("Truncated ticket summaries to the limit: ", MaxTicketSummaries)
		}

		// Generate org summary
		genSummaryInput := GenSummaryInput{Organization: s.organization}
		genSummaryOutput := GenSummaryOutput{}

		workflow.ExecuteActivity(s.Context, s.activity.GenOrgSummary, genSummaryInput).
			Get(s.Context, &genSummaryOutput)

		if genSummaryOutput.Summary != "" {
			s.organization.Summary = genSummaryOutput.Summary
		}
	}
}

func (s *organizationWorkflow) handleQuerySummary() (QueryOrganizationOutput, error) {
	return QueryOrganizationOutput{Summary: s.organization.Summary}, nil
}
