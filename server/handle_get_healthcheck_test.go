package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/taonic/ticketfu/config"
	"go.temporal.io/sdk/mocks"
	"go.temporal.io/server/common/log"
)

func TestHandleHealthCheck(t *testing.T) {
	testCases := []struct {
		name           string
		setupMock      func(*mocks.Client)
		expectedStatus int
		expectedResp   HealthResponse
	}{
		{
			name: "Healthy Service",
			setupMock: func(m *mocks.Client) {
				m.On("CheckHealth", mock.Anything, mock.Anything).Return(nil, nil)
			},
			expectedStatus: http.StatusOK,
			expectedResp: HealthResponse{
				Status:     "OK",
				TemporalOK: true,
			},
		},
		{
			name: "Unhealthy Temporal Service",
			setupMock: func(m *mocks.Client) {
				m.On("CheckHealth", mock.Anything, mock.Anything).Return(nil, errors.New("temporal service unavailable"))
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedResp: HealthResponse{
				Status:      "Degraded",
				TemporalOK:  false,
				TemporalMsg: "temporal service unavailable",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var mockClient *mocks.Client

			if tc.name != "Nil Temporal Client" {
				mockClient = &mocks.Client{}
				tc.setupMock(mockClient)
			}

			server := NewHTTPServer(config.ServerConfig{
				APIToken: "test-api-key",
			}, mockClient, log.NewTestLogger())

			// Create request
			req := httptest.NewRequest("GET", "/health", nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Call the handler
			server.handleHealthCheck(w, req)

			// Check status code
			assert.Equal(t, tc.expectedStatus, w.Code)

			// Parse response
			var resp HealthResponse
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)

			assert.Equal(t, tc.expectedResp.Status, resp.Status)
			assert.Equal(t, tc.expectedResp.TemporalOK, resp.TemporalOK)

			if tc.expectedResp.TemporalMsg != "" {
				assert.Equal(t, tc.expectedResp.TemporalMsg, resp.TemporalMsg)
			}

			// Assert expectations for mock
			if mockClient != nil {
				mockClient.AssertExpectations(t)
			}
		})
	}
}

func TestHandleHealthCheckContextTimeout(t *testing.T) {
	mockClient := &mocks.Client{}
	mockClient.On("CheckHealth", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		ctx := args.Get(0).(context.Context)
		<-ctx.Done()
	}).Return(nil, context.DeadlineExceeded)

	server := NewHTTPServer(config.ServerConfig{
		APIToken: "test-api-key",
	}, mockClient, log.NewTestLogger())

	// Create request
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.handleHealthCheck(w, req)

	// Check status code
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	// Parse response
	var resp HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	// Check response values
	assert.Equal(t, "Degraded", resp.Status)
	assert.False(t, resp.TemporalOK)
	assert.Contains(t, resp.TemporalMsg, "context deadline exceeded")

	mockClient.AssertExpectations(t)
}
