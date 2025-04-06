package webhook

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

func TestCreateTrigger(t *testing.T) {
	testSuite := testsuite.WorkflowTestSuite{}
	testEnv := testSuite.NewTestActivityEnvironment()

	testCases := []struct {
		name           string
		webhookID      string
		setupMock      func(*zd.MockZendeskClient)
		expectedOutput *CreateTriggerOutput
		expectedError  string
	}{
		{
			name:      "Successful Trigger Creation",
			webhookID: "webhook-123",
			setupMock: func(m *zd.MockZendeskClient) {
				// Expect CreateTrigger to be called with the expected parameters
				m.On("CreateTrigger", mock.Anything, mock.MatchedBy(func(trigger zendesk.Trigger) bool {
					// Verify the trigger is configured correctly
					if trigger.Title != "Notify TicketFu" || !trigger.Active {
						return false
					}

					// Check that we have at least one action
					if len(trigger.Actions) != 1 {
						return false
					}

					// Check the action is set to notification_webhook
					action := trigger.Actions[0]
					if action.Field != "notification_webhook" {
						return false
					}

					// Verify the webhook ID is in the action value
					webhookValues, ok := action.Value.([]interface{})
					if !ok || len(webhookValues) != 2 {
						return false
					}

					return webhookValues[0] == "webhook-123"
				})).Return(zendesk.Trigger{
					ID:     12345,
					Title:  "Notify TicketFu",
					Active: true,
				}, nil).Once()
			},
			expectedOutput: &CreateTriggerOutput{
				TriggerID: "12345",
			},
		},
		{
			name:      "API Error",
			webhookID: "webhook-123",
			setupMock: func(m *zd.MockZendeskClient) {
				m.On("CreateTrigger", mock.Anything, mock.Anything).
					Return(zendesk.Trigger{}, errors.New("API error")).Once()
			},
			expectedError: "failed to create trigger",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock client
			mockClient := new(zd.MockZendeskClient)
			tc.setupMock(mockClient)

			// Create the activity
			activity := &Activity{
				zClient: mockClient,
			}

			// Register the activity with the test environment
			testEnv.RegisterActivity(activity.CreateTrigger)

			// Create the input
			input := CreateTriggerInput{
				WebhookID: tc.webhookID,
			}

			// Execute the activity
			future, err := testEnv.ExecuteActivity(activity.CreateTrigger, input)

			// Check for errors or success
			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				require.NoError(t, err)

				var output CreateTriggerOutput
				err := future.Get(&output)
				require.NoError(t, err)

				assert.Equal(t, tc.expectedOutput.TriggerID, output.TriggerID)
			}

			// Verify that the mock was called as expected
			mockClient.AssertExpectations(t)
		})
	}
}
