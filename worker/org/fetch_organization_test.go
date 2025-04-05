package org

import (
	"context"
	"errors"
	"testing"

	"github.com/nukosuke/go-zendesk/zendesk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

// MockZendeskClient implements the zendesk.Client interface needed for this test
type MockZendeskClient struct {
	mock.Mock
}

// Implement all required methods from the zendesk.Client interface
func (m *MockZendeskClient) GetTicket(ctx context.Context, id int64) (zendesk.Ticket, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(zendesk.Ticket), args.Error(1)
}

func (m *MockZendeskClient) GetTicketCommentsCBP(ctx context.Context, opts *zendesk.CBPOptions) ([]zendesk.TicketComment, zendesk.CursorPaginationMeta, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).([]zendesk.TicketComment), args.Get(1).(zendesk.CursorPaginationMeta), args.Error(2)
}

func (m *MockZendeskClient) GetUser(ctx context.Context, userID int64) (zendesk.User, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(zendesk.User), args.Error(1)
}

func (m *MockZendeskClient) GetOrganization(ctx context.Context, orgID int64) (zendesk.Organization, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0).(zendesk.Organization), args.Error(1)
}

func TestActivity_FetchOrganization(t *testing.T) {
	testSuite := testsuite.WorkflowTestSuite{}
	testEnv := testSuite.NewTestActivityEnvironment()

	// Define test cases
	testCases := []struct {
		name          string
		orgID         int64
		setupMock     func(*MockZendeskClient)
		expected      *FetchOrganizationOutput
		expectedError string
	}{
		{
			name:  "Successful Organization Fetch",
			orgID: 123,
			setupMock: func(m *MockZendeskClient) {
				zendeskOrg := zendesk.Organization{
					ID:      123,
					Name:    "Test Organization",
					Details: "Organization details",
					Notes:   "Important notes",
				}
				m.On("GetOrganization", mock.Anything, int64(123)).Return(zendeskOrg, nil)
			},
			expected: &FetchOrganizationOutput{
				Organization: Organization{
					ID:      123,
					Name:    "Test Organization",
					Details: "Organization details",
					Notes:   "Important notes",
				},
			},
		},
		{
			name:  "Zendesk API Error",
			orgID: 456,
			setupMock: func(m *MockZendeskClient) {
				m.On("GetOrganization", mock.Anything, int64(456)).Return(zendesk.Organization{}, errors.New("zendesk API error"))
			},
			expectedError: "zendesk API error",
		},
		{
			name:  "Organization Not Found",
			orgID: 789,
			setupMock: func(m *MockZendeskClient) {
				// Use a simple error since we can't directly create a Zendesk Error type
				notFoundErr := errors.New("404 Not Found")
				m.On("GetOrganization", mock.Anything, int64(789)).Return(zendesk.Organization{}, notFoundErr)
			},
			expectedError: "404",
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			mockClient := new(MockZendeskClient)
			tc.setupMock(mockClient)

			activity := &Activity{
				zClient: mockClient,
			}

			// Register the activity with the test environment
			testEnv.RegisterActivity(activity.FetchOrganization)

			// Execute the activity
			input := FetchOrganizationInput{ID: tc.orgID}
			future, err := testEnv.ExecuteActivity(activity.FetchOrganization, input)

			// Verify
			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				require.NoError(t, err)

				var output FetchOrganizationOutput
				err := future.Get(&output)
				require.NoError(t, err)

				assert.Equal(t, tc.expected.Organization.ID, output.Organization.ID)
				assert.Equal(t, tc.expected.Organization.Name, output.Organization.Name)
				assert.Equal(t, tc.expected.Organization.Details, output.Organization.Details)
				assert.Equal(t, tc.expected.Organization.Notes, output.Organization.Notes)
			}

			mockClient.AssertExpectations(t)
		})
	}
}
