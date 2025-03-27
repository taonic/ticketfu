package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/taonic/ticketfu/config"
	"github.com/taonic/ticketfu/temporal"
	"github.com/taonic/ticketfu/temporal/workflows"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/mocks"
)

func TestHandleUpdateTicket(t *testing.T) {
	mockRun := &mocks.WorkflowRun{}
	mockRun.On("GetID").Return("test-workflow-id")

	testCases := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*mocks.Client)
		expectedStatus int
		expectedResp   *response
		expectedError  string
	}{
		{
			name: "Success",
			requestBody: UpdateTicketRequest{
				TicketURL:      "company.zendesk.com/tickets/12345",
				OrganizationID: "org123",
				RequesterID:    "user123",
				RequesterEmail: "user@example.com",
			},
			setupMock: func(m *mocks.Client) {
				workflowID := "ticket-workflow-12345"
				m.On("SignalWithStartWorkflow",
					mock.Anything,
					workflowID,
					workflows.UpdateTicketSummarySignal,
					mock.Anything,
					mock.MatchedBy(func(options client.StartWorkflowOptions) bool {
						return options.ID == workflowID && options.TaskQueue == temporal.TaskQueue
					}),
					mock.AnythingOfType("func(internal.Context, workflows.UpsertTicketInput) error"),
					mock.MatchedBy(func(input workflows.UpsertTicketInput) bool {
						return input.TicketID == "12345"
					}),
				).Return(mockRun, nil)
			},
			expectedStatus: http.StatusOK,
			expectedResp: &response{
				Message:    "Ticket update workflow started or signaled",
				WorkflowID: "",
			},
		},
		{
			name: "Invalid URL",
			requestBody: UpdateTicketRequest{
				TicketURL:      "invalid_url",
				OrganizationID: "org123",
				RequesterID:    "user123",
				RequesterEmail: "user@example.com",
			},
			setupMock: func(m *mocks.Client) {
				// No mock setup needed
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid Zendesk subdomain in URL: http://invalid_url",
		},
		{
			name:        "Invalid JSON",
			requestBody: "not-a-json",
			setupMock: func(m *mocks.Client) {
				// No mock setup needed
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid JSON payload",
		},
		{
			name: "Temporal Service Error",
			requestBody: UpdateTicketRequest{
				TicketURL:      "company.zendesk.com/tickets/12345",
				OrganizationID: "org123",
				RequesterID:    "user123",
				RequesterEmail: "user@example.com",
			},
			setupMock: func(m *mocks.Client) {
				m.On("SignalWithStartWorkflow",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(mockRun, errors.New("temporal service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to start or signal workflow",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock client
			mockClient := &mocks.Client{}
			tc.setupMock(mockClient)

			// Create server
			server := NewHTTPServer(config.ServerConfig{
				APIKey: "test-api-key",
			}, mockClient)

			// Create request
			var reqBody []byte
			var err error
			if str, ok := tc.requestBody.(string); ok {
				reqBody = []byte(str)
			} else {
				reqBody, err = json.Marshal(tc.requestBody)
				assert.NoError(t, err)
			}
			req := httptest.NewRequest("POST", "/api/v1/ticket", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set(APIKeyHeader, "test-api-key")

			w := httptest.NewRecorder()

			server.handleUpdateTicket(w, req)

			// Check response
			assert.Equal(t, tc.expectedStatus, w.Code)
			if tc.expectedError != "" {
				assert.Contains(t, w.Body.String(), tc.expectedError)
			} else if tc.expectedResp != nil {
				var resp response
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResp.Message, resp.Message)
				// Don't check WorkflowID directly as it's generated by the mock
				assert.NotEmpty(t, resp.WorkflowID)
			}

			mockClient.AssertExpectations(t)
		})
	}
}
