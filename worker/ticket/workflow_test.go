package ticket

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
)

type TicketWorkflowTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
}

func (s *TicketWorkflowTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

func (s *TicketWorkflowTestSuite) TearDownTest() {
	s.env.AssertExpectations(s.T())
}

func (s *TicketWorkflowTestSuite) TestBasicTicketWorkflow() {
	// Create initial empty ticket
	ticket := Ticket{ID: 0}

	// Mock the activities
	s.env.OnActivity((*Activity)(nil).FetchTicket, mock.Anything, FetchTicketInput{
		ID: "12345",
	}).Return(&FetchTicketOutput{
		Ticket: Ticket{
			ID:               12345,
			Subject:          "Test Subject",
			Description:      "Test Description",
			Priority:         "high",
			Status:           "open",
			Requester:        "Test User",
			Assignee:         "Test Agent",
			OrganizationID:   101,
			OrganizationName: "Test Organization",
			CreatedAt:        &time.Time{},
			UpdatedAt:        &time.Time{},
		},
	}, nil).Once()

	s.env.OnActivity((*Activity)(nil).FetchComments, mock.Anything, FetchCommentsInput{
		ID:     "12345",
		Cursor: "",
	}).Return(&FetchCommentsOutput{
		Comments:   []string{"First comment", "Second comment"},
		NextCursor: "next-page-token",
	}, nil).Once()

	s.env.OnActivity((*Activity)(nil).GenTicketSummary, mock.Anything, mock.MatchedBy(func(input GenSummaryInput) bool {
		return input.Ticket.ID == 12345 && len(input.Ticket.Comments) == 2
	})).Return(&GenSummaryOutput{
		Summary: "Test ticket summary",
	}, nil).Once()

	s.env.OnActivity((*Activity)(nil).SignalOrganization, mock.Anything, mock.MatchedBy(func(input SignalOrganizationInput) bool {
		return input.OrganizationID == 101 &&
			input.TicketID == 12345 &&
			input.TicketSummary == "Test ticket summary"
	})).Return(nil).Once()

	// Send signal to start processing
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(UpsertTicketSignal, UpsertTicketInput{
			TicketID: "12345",
		})
	}, time.Millisecond*100)

	// Execute workflow
	s.env.ExecuteWorkflow(TicketWorkflow, ticket)

	// Verify workflow completes successfully
	s.True(s.env.IsWorkflowCompleted())

	// Verify the query handler returns the expected summary
	var output QueryTicketOutput
	future, err := s.env.QueryWorkflow(QueryTicketSummary, nil)
	future.Get(&output)
	s.NoError(err)
	s.Equal("Test ticket summary", output.Summary)
}

func (s *TicketWorkflowTestSuite) TestTicketWithoutOrganization() {
	// Create initial empty ticket
	ticket := Ticket{}

	// Mock the activities
	s.env.OnActivity((*Activity)(nil).FetchTicket, mock.Anything, FetchTicketInput{
		ID: "12345",
	}).Return(&FetchTicketOutput{
		Ticket: Ticket{
			ID:               12345,
			Subject:          "Test Subject",
			Description:      "Test Description",
			Priority:         "high",
			Status:           "open",
			Requester:        "Test User",
			Assignee:         "Test Agent",
			OrganizationID:   0, // No organization
			OrganizationName: "",
			CreatedAt:        &time.Time{},
			UpdatedAt:        &time.Time{},
		},
	}, nil).Once()

	s.env.OnActivity((*Activity)(nil).FetchComments, mock.Anything, mock.Anything).
		Return(&FetchCommentsOutput{
			Comments:   []string{"First comment", "Second comment"},
			NextCursor: "",
		}, nil).Once()

	s.env.OnActivity((*Activity)(nil).GenTicketSummary, mock.Anything, mock.Anything).
		Return(&GenSummaryOutput{
			Summary: "Test ticket summary",
		}, nil).Once()

	// SignalOrganization should NOT be called since OrganizationID is 0

	// Send signal to start processing
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(UpsertTicketSignal, UpsertTicketInput{
			TicketID: "12345",
		})
	}, time.Millisecond*100)

	// Execute workflow
	s.env.ExecuteWorkflow(TicketWorkflow, ticket)

	// Verify workflow completes successfully
	s.True(s.env.IsWorkflowCompleted())
}

func (s *TicketWorkflowTestSuite) TestMultipleUpdates() {
	// Create initial empty ticket
	ticket := Ticket{ID: 0}

	// Mock the activities for first update
	s.env.OnActivity((*Activity)(nil).FetchTicket, mock.Anything, FetchTicketInput{
		ID: "12345",
	}).Return(&FetchTicketOutput{
		Ticket: Ticket{
			ID:               12345,
			Subject:          "Test Subject",
			Description:      "Test Description",
			OrganizationID:   101,
			OrganizationName: "Test Organization",
		},
	}, nil).Once()

	s.env.OnActivity((*Activity)(nil).FetchComments, mock.Anything, FetchCommentsInput{
		ID:     "12345",
		Cursor: "",
	}).Return(&FetchCommentsOutput{
		Comments:   []string{"First comment"},
		NextCursor: "cursor1",
	}, nil).Once()

	s.env.OnActivity((*Activity)(nil).GenTicketSummary, mock.Anything, mock.MatchedBy(func(input GenSummaryInput) bool {
		return len(input.Ticket.Comments) == 1
	})).Return(&GenSummaryOutput{
		Summary: "First summary",
	}, nil).Once()

	s.env.OnActivity((*Activity)(nil).SignalOrganization, mock.Anything, mock.Anything).
		Return(nil).Once()

	// Mock the activities for second update (with the cursor from first update)
	s.env.OnActivity((*Activity)(nil).FetchComments, mock.Anything, FetchCommentsInput{
		ID:     "12345",
		Cursor: "cursor1",
	}).Return(&FetchCommentsOutput{
		Comments:   []string{"Second comment", "Third comment"},
		NextCursor: "cursor2",
	}, nil).Once()

	s.env.OnActivity((*Activity)(nil).GenTicketSummary, mock.Anything, mock.MatchedBy(func(input GenSummaryInput) bool {
		return len(input.Ticket.Comments) == 2 // Only the new comments, not the old ones
	})).Return(&GenSummaryOutput{
		Summary: "Updated summary",
	}, nil).Once()

	s.env.OnActivity((*Activity)(nil).SignalOrganization, mock.Anything, mock.Anything).
		Return(nil).Once()

	// Send first signal
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(UpsertTicketSignal, UpsertTicketInput{
			TicketID: "12345",
		})
	}, time.Millisecond*100)

	// Send second signal
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(UpsertTicketSignal, UpsertTicketInput{
			TicketID: "12345",
		})
	}, time.Millisecond*200)

	// Cancel the workflow after processing both signals
	s.env.RegisterDelayedCallback(func() {
		s.env.CancelWorkflow()
	}, time.Millisecond*300)

	// Execute workflow
	s.env.ExecuteWorkflow(TicketWorkflow, ticket)

	// Verify workflow is cancelled as expected
	s.True(s.env.IsWorkflowCompleted())
	var canceledErr *temporal.CanceledError
	s.ErrorAs(s.env.GetWorkflowError(), &canceledErr)

	// Check the final state via query
	var output QueryTicketOutput
	future, err := s.env.QueryWorkflow(QueryTicketSummary, nil)
	future.Get(&output)
	s.NoError(err)
	s.Equal("Updated summary", output.Summary)
}

func TestTicketWorkflowSuite(t *testing.T) {
	suite.Run(t, new(TicketWorkflowTestSuite))
}
