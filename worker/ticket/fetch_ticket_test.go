package ticket

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/nukosuke/go-zendesk/zendesk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockZendeskClient is a mock of the Zendesk client interface used by the activity
type MockZendeskClient struct {
	mock.Mock
}

func (m *MockZendeskClient) GetTicket(ctx context.Context, id int64) (zendesk.Ticket, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(zendesk.Ticket), args.Error(1)
}

func (m *MockZendeskClient) GetUser(ctx context.Context, id int64) (zendesk.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(zendesk.User), args.Error(1)
}

func (m *MockZendeskClient) GetOrganization(ctx context.Context, id int64) (zendesk.Organization, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(zendesk.Organization), args.Error(1)
}

func (m *MockZendeskClient) GetTicketCommentsCBP(ctx context.Context, opts *zendesk.CBPOptions) ([]zendesk.TicketComment, zendesk.CursorPaginationMeta, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).([]zendesk.TicketComment), args.Get(1).(zendesk.CursorPaginationMeta), args.Error(2)
}

func TestActivity_FetchTicket_Success(t *testing.T) {
	// Setup
	mockZClient := new(MockZendeskClient)
	activity := &Activity{
		zClient: mockZClient,
	}

	ctx := context.Background()
	ticketID := "12345"
	parsedID, _ := strconv.ParseInt(ticketID, 10, 64)

	// Create test data
	now := time.Now()
	testTicket := zendesk.Ticket{
		ID:             parsedID,
		Subject:        "Test Subject",
		Description:    "Test Description",
		Priority:       "high",
		Status:         "open",
		RequesterID:    101,
		AssigneeID:     102,
		OrganizationID: 201,
		CreatedAt:      &now,
		UpdatedAt:      &now,
	}

	testRequester := zendesk.User{
		ID:   101,
		Name: "Test Requester",
	}

	testAssignee := zendesk.User{
		ID:   102,
		Name: "Test Assignee",
	}

	testOrg := zendesk.Organization{
		ID:   201,
		Name: "Test Organization",
	}

	// Set up mock expectations
	mockZClient.On("GetTicket", ctx, parsedID).Return(testTicket, nil)
	mockZClient.On("GetUser", mock.Anything, int64(101)).Return(testRequester, nil)
	mockZClient.On("GetUser", mock.Anything, int64(102)).Return(testAssignee, nil)
	mockZClient.On("GetOrganization", mock.Anything, int64(201)).Return(testOrg, nil)

	// Execute
	input := FetchTicketInput{ID: ticketID}
	output, err := activity.FetchTicket(ctx, input)

	// Verify
	require.NoError(t, err)
	require.NotNil(t, output)

	assert.Equal(t, parsedID, output.Ticket.ID)
	assert.Equal(t, testTicket.Subject, output.Ticket.Subject)
	assert.Equal(t, testTicket.Description, output.Ticket.Description)
	assert.Equal(t, testTicket.Priority, output.Ticket.Priority)
	assert.Equal(t, testTicket.Status, output.Ticket.Status)
	assert.Equal(t, testRequester.Name, output.Ticket.Requester)
	assert.Equal(t, testAssignee.Name, output.Ticket.Assignee)
	assert.Equal(t, testOrg.Name, output.Ticket.OrganizationName)
	assert.Equal(t, testTicket.CreatedAt, output.Ticket.CreatedAt)
	assert.Equal(t, testTicket.UpdatedAt, output.Ticket.UpdatedAt)

	// Verify all mock expectations were met
	mockZClient.AssertExpectations(t)
}

func TestActivity_FetchTicket_InvalidID(t *testing.T) {
	// Setup
	mockZClient := new(MockZendeskClient)
	activity := &Activity{
		zClient: mockZClient,
	}

	ctx := context.Background()

	// Execute with non-numeric ID
	input := FetchTicketInput{ID: "not-a-number"}
	output, err := activity.FetchTicket(ctx, input)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "strconv.ParseInt")
}

func TestActivity_FetchTicket_GetTicketError(t *testing.T) {
	// Setup
	mockZClient := new(MockZendeskClient)
	activity := &Activity{
		zClient: mockZClient,
	}

	ctx := context.Background()
	ticketID := "12345"
	parsedID, _ := strconv.ParseInt(ticketID, 10, 64)

	// Setup mock to return error
	expectedError := errors.New("zendesk API error")
	mockZClient.On("GetTicket", ctx, parsedID).Return(zendesk.Ticket{}, expectedError)

	// Execute
	input := FetchTicketInput{ID: ticketID}
	output, err := activity.FetchTicket(ctx, input)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Equal(t, expectedError, err)

	mockZClient.AssertExpectations(t)
}

func TestActivity_FetchTicket_GetUserError(t *testing.T) {
	// Setup
	mockZClient := new(MockZendeskClient)
	activity := &Activity{
		zClient: mockZClient,
	}

	ctx := context.Background()
	ticketID := "12345"
	parsedID, _ := strconv.ParseInt(ticketID, 10, 64)

	// Create test data
	now := time.Now()
	testTicket := zendesk.Ticket{
		ID:             parsedID,
		Subject:        "Test Subject",
		Description:    "Test Description",
		RequesterID:    101,
		AssigneeID:     102,
		OrganizationID: 201,
		CreatedAt:      &now,
		UpdatedAt:      &now,
	}

	// Setup mocks
	mockZClient.On("GetTicket", ctx, parsedID).Return(testTicket, nil)
	mockZClient.On("GetUser", mock.Anything, int64(101)).Return(zendesk.User{}, errors.New("requester not found"))

	// For this test we don't expect the assignee or organization to be called,
	// since the requester call should fail first, but with errgroup we can't guarantee order
	// So we need to set up the expectations anyway, as they might be called before the error is returned
	mockZClient.On("GetUser", mock.Anything, int64(102)).Return(zendesk.User{}, nil).Maybe()
	mockZClient.On("GetOrganization", mock.Anything, int64(201)).Return(zendesk.Organization{}, nil).Maybe()

	// Execute
	input := FetchTicketInput{ID: ticketID}
	output, err := activity.FetchTicket(ctx, input)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "requester not found")

	mockZClient.AssertExpectations(t)
}

func TestActivity_FetchTicket_GetOrganizationError(t *testing.T) {
	// Setup
	mockZClient := MockZendeskClient{}
	activity := &Activity{
		zClient: &mockZClient,
	}

	ctx := context.Background()
	ticketID := "12345"
	parsedID, _ := strconv.ParseInt(ticketID, 10, 64)

	// Create test data
	now := time.Now()
	testTicket := zendesk.Ticket{
		ID:             parsedID,
		RequesterID:    101,
		AssigneeID:     102,
		OrganizationID: 201,
		CreatedAt:      &now,
		UpdatedAt:      &now,
	}

	testRequester := zendesk.User{
		ID:   101,
		Name: "Test Requester",
	}

	testAssignee := zendesk.User{
		ID:   102,
		Name: "Test Assignee",
	}

	// Setup mocks
	mockZClient.On("GetTicket", ctx, parsedID).Return(testTicket, nil)
	mockZClient.On("GetUser", mock.Anything, int64(101)).Return(testRequester, nil)
	mockZClient.On("GetUser", mock.Anything, int64(102)).Return(testAssignee, nil)
	mockZClient.On("GetOrganization", mock.Anything, int64(201)).Return(zendesk.Organization{}, errors.New("org not found"))

	// Execute
	input := FetchTicketInput{ID: ticketID}
	output, err := activity.FetchTicket(ctx, input)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "org not found")

	mockZClient.AssertExpectations(t)
}
