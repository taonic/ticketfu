package org

import (
	"errors"
	"testing"

	"github.com/nukosuke/go-zendesk/zendesk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	zd "github.com/taonic/ticketfu/zendesk"
	"go.temporal.io/sdk/testsuite"
)

func TestActivity_FetchOrganization(t *testing.T) {
	testSuite := testsuite.WorkflowTestSuite{}
	testEnv := testSuite.NewTestActivityEnvironment()

	// Define test cases
	testCases := []struct {
		name          string
		orgID         int64
		setupMock     func(*zd.MockZendeskClient)
		expected      *FetchOrganizationOutput
		expectedError string
	}{
		{
			name:  "Successful Organization Fetch",
			orgID: 123,
			setupMock: func(m *zd.MockZendeskClient) {
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
			setupMock: func(m *zd.MockZendeskClient) {
				m.On("GetOrganization", mock.Anything, int64(456)).Return(zendesk.Organization{}, errors.New("zendesk API error"))
			},
			expectedError: "zendesk API error",
		},
		{
			name:  "Organization Not Found",
			orgID: 789,
			setupMock: func(m *zd.MockZendeskClient) {
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
			mockClient := new(zd.MockZendeskClient)
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
