package ticket

import (
	"errors"
	"testing"
	"time"

	"github.com/nukosuke/go-zendesk/zendesk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	zd "github.com/taonic/ticketfu/zendesk"
	"go.temporal.io/sdk/testsuite"
)

func TestFetchTicket(t *testing.T) {
	testSuite := testsuite.WorkflowTestSuite{}
	testEnv := testSuite.NewTestActivityEnvironment()

	testCases := []struct {
		name           string
		ticketID       string
		setupMock      func(*zd.MockZendeskClient)
		expectedErr    string
		expectedTicket *Ticket
	}{
		{
			name:     "Successful Fetch",
			ticketID: "12345",
			setupMock: func(m *zd.MockZendeskClient) {
				now := time.Now()
				m.On("GetTicket", mock.Anything, int64(12345)).Return(zendesk.Ticket{
					ID:             12345,
					Subject:        "Test Subject",
					Description:    "Test Description",
					Priority:       "high",
					Status:         "open",
					RequesterID:    101,
					AssigneeID:     102,
					OrganizationID: 201,
					CreatedAt:      &now,
					UpdatedAt:      &now,
				}, nil)
				m.On("GetUser", mock.Anything, int64(101)).Return(zendesk.User{
					ID:   101,
					Name: "Test Requester",
				}, nil)
				m.On("GetUser", mock.Anything, int64(102)).Return(zendesk.User{
					ID:   102,
					Name: "Test Assignee",
				}, nil)
				m.On("GetOrganization", mock.Anything, int64(201)).Return(zendesk.Organization{
					ID:   201,
					Name: "Test Organization",
				}, nil)
			},
			expectedTicket: &Ticket{
				ID:               12345,
				Subject:          "Test Subject",
				Description:      "Test Description",
				Priority:         "high",
				Status:           "open",
				Requester:        "Test Requester",
				Assignee:         "Test Assignee",
				OrganizationID:   201,
				OrganizationName: "Test Organization",
				// CreatedAt and UpdatedAt will be checked separately
			},
		},
		{
			name:     "Invalid ID",
			ticketID: "not-a-number",
			setupMock: func(m *zd.MockZendeskClient) {
				// No mock setup needed as it will fail parsing the ID
			},
			expectedErr: "strconv.ParseInt",
		},
		{
			name:     "Ticket API Error",
			ticketID: "12345",
			setupMock: func(m *zd.MockZendeskClient) {
				m.On("GetTicket", mock.Anything, int64(12345)).Return(zendesk.Ticket{}, errors.New("ticket API error"))
			},
			expectedErr: "ticket API error",
		},
		{
			name:     "Requester API Error",
			ticketID: "12345",
			setupMock: func(m *zd.MockZendeskClient) {
				now := time.Now()
				m.On("GetTicket", mock.Anything, int64(12345)).Return(zendesk.Ticket{
					ID:             12345,
					RequesterID:    101,
					AssigneeID:     102,
					OrganizationID: 201,
					CreatedAt:      &now,
					UpdatedAt:      &now,
				}, nil)
				m.On("GetUser", mock.Anything, int64(101)).Return(zendesk.User{}, errors.New("requester API error"))
				m.On("GetUser", mock.Anything, int64(102)).Return(zendesk.User{}, nil).Maybe()
				m.On("GetOrganization", mock.Anything, int64(201)).Return(zendesk.Organization{}, nil).Maybe()
			},
			expectedErr: "requester API error",
		},
		{
			name:     "Organization API Error",
			ticketID: "12345",
			setupMock: func(m *zd.MockZendeskClient) {
				now := time.Now()
				m.On("GetTicket", mock.Anything, int64(12345)).Return(zendesk.Ticket{
					ID:             12345,
					RequesterID:    101,
					AssigneeID:     102,
					OrganizationID: 201,
					CreatedAt:      &now,
					UpdatedAt:      &now,
				}, nil)
				m.On("GetUser", mock.Anything, int64(101)).Return(zendesk.User{
					ID:   101,
					Name: "Test Requester",
				}, nil)
				m.On("GetUser", mock.Anything, int64(102)).Return(zendesk.User{
					ID:   102,
					Name: "Test Assignee",
				}, nil)
				m.On("GetOrganization", mock.Anything, int64(201)).Return(zendesk.Organization{}, errors.New("org API error"))
			},
			expectedErr: "org API error",
		},
		{
			name:     "Ticket with No Organization",
			ticketID: "12345",
			setupMock: func(m *zd.MockZendeskClient) {
				now := time.Now()
				// Setup mock ticket response with no organization
				m.On("GetTicket", mock.Anything, int64(12345)).Return(zendesk.Ticket{
					ID:             12345,
					Subject:        "Test Subject",
					Description:    "Test Description",
					Priority:       "high",
					Status:         "open",
					RequesterID:    101,
					AssigneeID:     102,
					OrganizationID: 0, // No organization
					CreatedAt:      &now,
					UpdatedAt:      &now,
				}, nil)

				// Setup requester response
				m.On("GetUser", mock.Anything, int64(101)).Return(zendesk.User{
					ID:   101,
					Name: "Test Requester",
				}, nil)

				// Setup assignee response
				m.On("GetUser", mock.Anything, int64(102)).Return(zendesk.User{
					ID:   102,
					Name: "Test Assignee",
				}, nil)
			},
			expectedTicket: &Ticket{
				ID:               12345,
				Subject:          "Test Subject",
				Description:      "Test Description",
				Priority:         "high",
				Status:           "open",
				Requester:        "Test Requester",
				Assignee:         "Test Assignee",
				OrganizationID:   0,
				OrganizationName: "", // Should be empty since no org
				// CreatedAt and UpdatedAt will be checked separately
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create and setup mock client
			mockZClient := new(zd.MockZendeskClient)
			tc.setupMock(mockZClient)

			// Create activity instance
			activity := &Activity{
				zClient: mockZClient,
			}

			// Register the activity
			testEnv.RegisterActivity(activity.FetchTicket)

			// Create input
			input := FetchTicketInput{ID: tc.ticketID}

			// Execute the activity
			future, err := testEnv.ExecuteActivity(activity.FetchTicket, input)

			// Verify results
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr)
			} else {
				// Check for successful result
				var output FetchTicketOutput
				err := future.Get(&output)
				require.NoError(t, err)
				require.NotNil(t, output)

				// Test the ticket fields
				assert.Equal(t, tc.expectedTicket.ID, output.Ticket.ID)
				assert.Equal(t, tc.expectedTicket.Subject, output.Ticket.Subject)
				assert.Equal(t, tc.expectedTicket.Description, output.Ticket.Description)
				assert.Equal(t, tc.expectedTicket.Priority, output.Ticket.Priority)
				assert.Equal(t, tc.expectedTicket.Status, output.Ticket.Status)
				assert.Equal(t, tc.expectedTicket.Requester, output.Ticket.Requester)
				assert.Equal(t, tc.expectedTicket.Assignee, output.Ticket.Assignee)
				assert.Equal(t, tc.expectedTicket.OrganizationID, output.Ticket.OrganizationID)
				assert.Equal(t, tc.expectedTicket.OrganizationName, output.Ticket.OrganizationName)

				assert.NotNil(t, output.Ticket.CreatedAt)
				assert.NotNil(t, output.Ticket.UpdatedAt)
			}

			mockZClient.AssertExpectations(t)
		})
	}
}
