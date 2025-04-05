package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/taonic/ticketfu/config"
	"github.com/taonic/ticketfu/worker/ticket"
	"go.temporal.io/sdk/mocks"
	"go.temporal.io/server/common/log"
)

func TestHandleGetTicket(t *testing.T) {
	testCases := []struct {
		name           string
		ticketID       string
		setupMock      func(*mocks.Client)
		expectedStatus int
		expectedResp   *ticket.QueryTicketOutput
		expectedError  string
	}{
		{
			name:     "Successful Query",
			ticketID: "12345",
			setupMock: func(m *mocks.Client) {
				mockFuture := &mocks.Value{}
				mockFuture.On("Get", mock.Anything).Run(func(args mock.Arguments) {
					output := args.Get(0).(*ticket.QueryTicketOutput)
					output.Summary = "Test ticket summary for ID 12345"
				}).Return(nil)

				m.On("QueryWorkflow", mock.Anything, "ticket-workflow-12345", "", ticket.QueryTicketSummary, "").
					Return(mockFuture, nil)
			},
			expectedStatus: http.StatusOK,
			expectedResp: &ticket.QueryTicketOutput{
				Summary: "Test ticket summary for ID 12345",
			},
		},
		{
			name:     "Workflow Not Found",
			ticketID: "99999",
			setupMock: func(m *mocks.Client) {
				m.On("QueryWorkflow", mock.Anything, "ticket-workflow-99999", "", ticket.QueryTicketSummary, "").
					Return(nil, errors.New("workflow not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Failed to query workflow",
		},
		{
			name:     "Query Result Error",
			ticketID: "54321",
			setupMock: func(m *mocks.Client) {
				mockFuture := &mocks.Value{}
				mockFuture.On("Get", mock.Anything).Return(errors.New("failed to get query result"))

				m.On("QueryWorkflow", mock.Anything, "ticket-workflow-54321", "", ticket.QueryTicketSummary, "").
					Return(mockFuture, nil)
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Failed to query workflow",
		},
		{
			name:     "Empty Ticket ID",
			ticketID: "",
			setupMock: func(m *mocks.Client) {
				// No mocks needed as handler should validate ID before querying
			},
			expectedStatus: http.StatusMovedPermanently,
			expectedError:  "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock client
			mockClient := &mocks.Client{}
			tc.setupMock(mockClient)

			// Create server
			server := NewHTTPServer(config.ServerConfig{
				APIToken: "test-api-key",
			}, mockClient, log.NewTestLogger())

			// Create request
			url := "/api/v1/ticket/" + tc.ticketID + "/summary"
			req := httptest.NewRequest("GET", url, nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set(APIKeyHeader, "test-api-key")

			// Create response recorder
			w := httptest.NewRecorder()

			// Create router and register the handler
			router := mux.NewRouter()
			router.HandleFunc("/api/v1/ticket/{ticketId}/summary", server.handleGetTicket)

			// Serve the request
			router.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, tc.expectedStatus, w.Code)

			// Check response
			if tc.expectedError != "" {
				assert.Contains(t, w.Body.String(), tc.expectedError)
			} else if tc.expectedResp != nil {
				var resp ticket.QueryTicketOutput
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResp.Summary, resp.Summary)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestHandleGetTicketWithMissingWorkflow(t *testing.T) {
	// Create mock client
	mockClient := &mocks.Client{}

	// Set up mock to simulate a specific Temporal error
	mockClient.On("QueryWorkflow", mock.Anything, "ticket-workflow-45678", "", ticket.QueryTicketSummary, "").
		Return(nil, errors.New("workflow execution not found"))

	// Create server
	server := NewHTTPServer(config.ServerConfig{
		APIToken: "test-api-key",
	}, mockClient, log.NewTestLogger())

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/ticket/45678/summary", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(APIKeyHeader, "test-api-key")

	// Create response recorder
	w := httptest.NewRecorder()

	// Create router and register the handler
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/ticket/{ticketId}/summary", server.handleGetTicket)

	// Serve the request
	router.ServeHTTP(w, req)

	// Check status code - should be 404 Not Found
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to query workflow")

	mockClient.AssertExpectations(t)
}
