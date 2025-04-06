package webhook

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/nukosuke/go-zendesk/zendesk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	zd "github.com/taonic/ticketfu/zendesk"
	"go.temporal.io/sdk/testsuite"
)

func TestCreateWebhook(t *testing.T) {
	testSuite := testsuite.WorkflowTestSuite{}
	testEnv := testSuite.NewTestActivityEnvironment()

	testCases := []struct {
		name           string
		webhook        Webhook
		setupMock      func(*zd.MockZendeskClient)
		expectedOutput *CreateWebhookOutput
		expectedError  string
	}{
		{
			name: "Create New Webhook",
			webhook: Webhook{
				BaseURL:        "https://example.com",
				ServerAPIToken: "test-api-token",
			},
			setupMock: func(m *zd.MockZendeskClient) {
				expectedWebhook := &zendesk.Webhook{
					Name:          WebhookName,
					Status:        "active",
					Endpoint:      "https://example.com/api/v1/ticket",
					HTTPMethod:    "POST",
					RequestFormat: "json",
					Authentication: &zendesk.WebhookAuthentication{
						Type:        "api_key",
						AddPosition: "header",
						Data: map[string]string{
							"name":  "X-Ticketfu-Key",
							"value": "test-api-token",
						},
					},
				}

				createdWebhook := &zendesk.Webhook{
					ID:            "new-webhook-123",
					Name:          WebhookName,
					Status:        "active",
					Endpoint:      "https://example.com/api/v1/ticket",
					HTTPMethod:    "POST",
					RequestFormat: "json",
				}

				m.On("CreateWebhook", mock.Anything, mock.MatchedBy(func(hook *zendesk.Webhook) bool {
					return hook.Name == expectedWebhook.Name &&
						hook.Endpoint == expectedWebhook.Endpoint &&
						hook.HTTPMethod == expectedWebhook.HTTPMethod &&
						hook.RequestFormat == expectedWebhook.RequestFormat &&
						hook.Authentication.Type == expectedWebhook.Authentication.Type &&
						hook.Authentication.AddPosition == expectedWebhook.Authentication.AddPosition
				})).Return(createdWebhook, nil).Once()
			},
			expectedOutput: &CreateWebhookOutput{
				WebhookID: "new-webhook-123",
			},
		},
		{
			name: "Check Existing Webhook",
			webhook: Webhook{
				ID:             "existing-webhook-123",
				BaseURL:        "https://example.com",
				ServerAPIToken: "test-api-token",
			},
			setupMock: func(m *zd.MockZendeskClient) {
				existingWebhook := &zendesk.Webhook{
					ID:            "existing-webhook-123",
					Name:          WebhookName,
					Status:        "active",
					Endpoint:      "https://example.com/api/v1/ticket",
					HTTPMethod:    "POST",
					RequestFormat: "json",
				}

				m.On("GetWebhook", mock.Anything, "existing-webhook-123").
					Return(existingWebhook, nil).Once()
			},
			expectedOutput: &CreateWebhookOutput{
				WebhookID: "existing-webhook-123",
			},
		},
		{
			name: "Webhook Not Found Creates New One",
			webhook: Webhook{
				ID:             "nonexistent-webhook-123",
				BaseURL:        "https://example.com",
				ServerAPIToken: "test-api-token",
			},
			setupMock: func(m *zd.MockZendeskClient) {
				// Return 404 for the GetWebhook call
				notFoundErr := zendesk.NewError(nil, &http.Response{StatusCode: 404})
				m.On("GetWebhook", mock.Anything, "nonexistent-webhook-123").
					Return(nil, notFoundErr).Once()

				// Expect CreateWebhook to be called with the right parameters
				createdWebhook := &zendesk.Webhook{
					ID:            "new-webhook-456",
					Name:          WebhookName,
					Status:        "active",
					Endpoint:      "https://example.com/api/v1/ticket",
					HTTPMethod:    "POST",
					RequestFormat: "json",
				}

				m.On("CreateWebhook", mock.Anything, mock.Anything).
					Return(createdWebhook, nil).Once()
			},
			expectedOutput: &CreateWebhookOutput{
				WebhookID: "new-webhook-456",
			},
		},
		{
			name: "GetWebhook API Error",
			webhook: Webhook{
				ID:             "error-webhook-123",
				BaseURL:        "https://example.com",
				ServerAPIToken: "test-api-token",
			},
			setupMock: func(m *zd.MockZendeskClient) {
				m.On("GetWebhook", mock.Anything, "error-webhook-123").
					Return(nil, errors.New("API error")).Once()
			},
			expectedError: "failed to get webhook from Zendesk",
		},
		{
			name: "CreateWebhook API Error",
			webhook: Webhook{
				BaseURL:        "https://example.com",
				ServerAPIToken: "test-api-token",
			},
			setupMock: func(m *zd.MockZendeskClient) {
				m.On("CreateWebhook", mock.Anything, mock.Anything).
					Return(nil, errors.New("API error")).Once()
			},
			expectedError: "failed to create webhook",
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
			testEnv.RegisterActivity(activity.CreateWebhook)

			// Create the input
			input := CreateWebhookInput{
				Webhook: tc.webhook,
			}

			// Execute the activity
			future, err := testEnv.ExecuteActivity(activity.CreateWebhook, input)

			// Check for errors or success
			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				require.NoError(t, err)

				var output CreateWebhookOutput
				err := future.Get(&output)
				require.NoError(t, err)

				assert.Equal(t, tc.expectedOutput.WebhookID, output.WebhookID)
			}

			// Verify that the mock was called as expected
			mockClient.AssertExpectations(t)
		})
	}
}

func TestGetWebhook(t *testing.T) {
	testCases := []struct {
		name           string
		webhookID      string
		setupMock      func(*zd.MockZendeskClient)
		expectedOutput *zendesk.Webhook
		expectedError  string
	}{
		{
			name:      "Existing Webhook",
			webhookID: "existing-webhook-123",
			setupMock: func(m *zd.MockZendeskClient) {
				existingWebhook := &zendesk.Webhook{
					ID:            "existing-webhook-123",
					Name:          WebhookName,
					Status:        "active",
					Endpoint:      "https://example.com/api/v1/ticket",
					HTTPMethod:    "POST",
					RequestFormat: "json",
				}

				m.On("GetWebhook", mock.Anything, "existing-webhook-123").
					Return(existingWebhook, nil).Once()
			},
			expectedOutput: &zendesk.Webhook{
				ID:            "existing-webhook-123",
				Name:          WebhookName,
				Status:        "active",
				Endpoint:      "https://example.com/api/v1/ticket",
				HTTPMethod:    "POST",
				RequestFormat: "json",
			},
		},
		{
			name:      "Webhook Not Found",
			webhookID: "nonexistent-webhook-123",
			setupMock: func(m *zd.MockZendeskClient) {
				// Create a Zendesk Error with 404 status
				notFoundErr := zendesk.NewError(nil, &http.Response{StatusCode: 404})
				m.On("GetWebhook", mock.Anything, "nonexistent-webhook-123").
					Return(nil, notFoundErr).Once()
			},
			expectedOutput: nil,
		},
		{
			name:      "API Error",
			webhookID: "error-webhook-123",
			setupMock: func(m *zd.MockZendeskClient) {
				m.On("GetWebhook", mock.Anything, "error-webhook-123").
					Return(nil, errors.New("API error")).Once()
			},
			expectedError: "API error",
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

			// Call getWebhook directly
			result, err := activity.getWebhook(context.Background(), tc.webhookID)

			// Check for errors or success
			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedOutput, result)
			}

			// Verify that the mock was called as expected
			mockClient.AssertExpectations(t)
		})
	}
}
