package org

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
)

type OrgWorkflowTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
}

func (s *OrgWorkflowTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

func (s *OrgWorkflowTestSuite) TearDownTest() {
	s.env.AssertExpectations(s.T())
}

func (s *OrgWorkflowTestSuite) TestSuccessfulInitialFetch() {
	// Initial empty organization with just an ID
	org := Organization{ID: 303}

	// Mock successful fetch activity
	s.env.OnActivity((*Activity)(nil).FetchOrganization, mock.Anything, FetchOrganizationInput{ID: 303}).
		Return(&FetchOrganizationOutput{
			Organization: Organization{
				ID:              303,
				Name:            "Fetched Organization",
				Details:         "Organization details",
				Notes:           "Important notes",
				TicketSummaries: make(map[int64]string),
			},
		}, nil).Once()

	// Mock successful summary generation
	s.env.OnActivity((*Activity)(nil).GenOrgSummary, mock.Anything, mock.Anything).
		Return(&GenSummaryOutput{
			Summary: "Generated organization summary",
		}, nil).Once()

	// Send signal to trigger activities
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(UpsertOrganizationSignal, UpsertOrganizationInput{
			OrganizationID: 303,
			TicketID:       3001,
			TicketSummary:  "First ticket summary",
		})
	}, time.Millisecond*100)

	// Add cancellation to complete the test
	s.env.RegisterDelayedCallback(func() {
		s.env.CancelWorkflow()
	}, time.Millisecond*300)

	// Execute workflow
	s.env.ExecuteWorkflow(OrganizationWorkflow, org)

	// Verify workflow completed with cancellation
	s.True(s.env.IsWorkflowCompleted())
	var canceledErr *temporal.CanceledError
	s.ErrorAs(s.env.GetWorkflowError(), &canceledErr)

	// Query the workflow state
	var summary string
	future, err := s.env.QueryWorkflow(QueryOrganizationSummary, nil)
	future.Get(&summary)
	s.NoError(err)
	s.Equal("Generated organization summary", summary)
}

func (s *OrgWorkflowTestSuite) TestDuplicateTicketSummaries() {
	// Organization with existing ticket summary
	org := Organization{
		ID:   404,
		Name: "Duplicate Test Org",
		TicketSummaries: map[int64]string{
			4001: "Existing summary for ticket 4001",
		},
	}

	// Send signal with the same ticket ID but different summary
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(UpsertOrganizationSignal, UpsertOrganizationInput{
			OrganizationID: 404,
			TicketID:       4001,
			TicketSummary:  "Updated summary for ticket 4001",
		})
	}, time.Millisecond*100)

	// Mock summary generation (should be called because summary changed)
	s.env.OnActivity((*Activity)(nil).GenOrgSummary, mock.Anything, mock.Anything).
		Return(&GenSummaryOutput{
			Summary: "Updated org summary after ticket change",
		}, nil).Once()

	// Send another signal with same ticket ID and same summary (shouldn't trigger update)
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(UpsertOrganizationSignal, UpsertOrganizationInput{
			OrganizationID: 404,
			TicketID:       4001,
			TicketSummary:  "Updated summary for ticket 4001", // Same as previous signal
		})
	}, time.Millisecond*200)

	// Add cancellation to complete the test
	s.env.RegisterDelayedCallback(func() {
		s.env.CancelWorkflow()
	}, time.Millisecond*300)

	// Execute workflow
	s.env.ExecuteWorkflow(OrganizationWorkflow, org)

	// Verify workflow completed with cancellation
	s.True(s.env.IsWorkflowCompleted())
	var canceledErr *temporal.CanceledError
	s.ErrorAs(s.env.GetWorkflowError(), &canceledErr)

	// Query the workflow state
	var summary string
	future, err := s.env.QueryWorkflow(QueryOrganizationSummary, nil)
	future.Get(&summary)
	s.NoError(err)
	s.Equal("Updated org summary after ticket change", summary)
}

func (s *OrgWorkflowTestSuite) TestTicketTruncation() {
	// Create a map with test ticket summaries
	ticketMap := make(map[int64]string)
	for i := int64(1); i <= MaxTicketSummaries+10; i++ {
		ticketMap[i] = "Summary for ticket " + string(rune(i))
	}

	// Create organization with more than the maximum allowed tickets
	org := Organization{
		ID:              505,
		Name:            "Truncation Test Org",
		TicketSummaries: ticketMap,
	}

	// Add a new ticket that should trigger truncation
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(UpsertOrganizationSignal, UpsertOrganizationInput{
			OrganizationID: 505,
			TicketID:       MaxTicketSummaries + 11,
			TicketSummary:  "New ticket that triggers truncation",
		})
	}, time.Millisecond*100)

	// Mock GenOrgSummary - verify input contains truncated map
	s.env.OnActivity((*Activity)(nil).GenOrgSummary, mock.Anything, mock.MatchedBy(func(input GenSummaryInput) bool {
		// Verify ticket map was truncated to MaxTicketSummaries
		return len(input.Organization.TicketSummaries) == MaxTicketSummaries
	})).Return(&GenSummaryOutput{
		Summary: "Summary after truncation",
	}, nil).Once()

	// Add cancellation to complete the test
	s.env.RegisterDelayedCallback(func() {
		s.env.CancelWorkflow()
	}, time.Millisecond*200)

	// Execute workflow
	s.env.ExecuteWorkflow(OrganizationWorkflow, org)

	// Verify workflow completed
	s.True(s.env.IsWorkflowCompleted())
}

func (s *OrgWorkflowTestSuite) TestConcurrentSignals() {
	// Initial organization
	org := Organization{
		ID:              707,
		Name:            "Concurrent Signals Test Org",
		TicketSummaries: make(map[int64]string),
	}

	// Send multiple signals concurrently (same timestamp)
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(UpsertOrganizationSignal, UpsertOrganizationInput{
			OrganizationID: 707,
			TicketID:       7001,
			TicketSummary:  "First concurrent ticket",
		})

		s.env.SignalWorkflow(UpsertOrganizationSignal, UpsertOrganizationInput{
			OrganizationID: 707,
			TicketID:       7002,
			TicketSummary:  "Second concurrent ticket",
		})

		s.env.SignalWorkflow(UpsertOrganizationSignal, UpsertOrganizationInput{
			OrganizationID: 707,
			TicketID:       7003,
			TicketSummary:  "Third concurrent ticket",
		})
	}, time.Millisecond*100)

	// Mock GenOrgSummary - should be called for each update
	s.env.OnActivity((*Activity)(nil).GenOrgSummary, mock.Anything, mock.Anything).
		Return(&GenSummaryOutput{
			Summary: "Summary after concurrent signals",
		}, nil).Times(3)

	// Add cancellation to complete the test
	s.env.RegisterDelayedCallback(func() {
		s.env.CancelWorkflow()
	}, time.Millisecond*200)

	// Execute workflow
	s.env.ExecuteWorkflow(OrganizationWorkflow, org)

	// Verify workflow completed
	s.True(s.env.IsWorkflowCompleted())
	var canceledErr *temporal.CanceledError
	s.ErrorAs(s.env.GetWorkflowError(), &canceledErr)

	// Query final state
	var summary string
	future, err := s.env.QueryWorkflow(QueryOrganizationSummary, nil)
	future.Get(&summary)
	s.NoError(err)
	s.Equal("Summary after concurrent signals", summary)
}

func TestOrgWorkflowSuite(t *testing.T) {
	suite.Run(t, new(OrgWorkflowTestSuite))
}
