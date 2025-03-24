package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/taonic/ticketfu/config"
	"go.temporal.io/sdk/mocks"
)

func TestHTTPServer_RegisterRoutes(t *testing.T) {
	cfg := config.ServerConfig{
		Host:   "localhost",
		Port:   8080,
		APIKey: "test-api-key",
		Temporal: config.TemporalClientConfig{
			Address:   "localhost:7233",
			Namespace: "test-namespace",
		},
	}

	mockClient := &mocks.Client{}
	mockClient.On("CheckHealth", mock.Anything, mock.Anything).Return(nil, nil)

	server := NewHTTPServer(cfg, mockClient)
	server.registerRoutes()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	server.mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHTTPServer_HealthEndpoint(t *testing.T) {
	cfg := config.ServerConfig{
		Host:   "localhost",
		Port:   8080,
		APIKey: "test-api-key",
		Temporal: config.TemporalClientConfig{
			Address:   "localhost:7233",
			Namespace: "test-namespace",
		},
	}

	testCases := []struct {
		name           string
		checkHealthErr error
		expectedStatus int
		expectedOK     bool
	}{
		{
			name:           "Temporal service healthy",
			checkHealthErr: nil,
			expectedStatus: http.StatusOK,
			expectedOK:     true,
		},
		{
			name:           "Temporal service unhealthy",
			checkHealthErr: errors.New("temporal service unavailable"),
			expectedStatus: http.StatusServiceUnavailable,
			expectedOK:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &mocks.Client{}
			mockClient.On("CheckHealth", mock.Anything, mock.Anything).Return(nil, tc.checkHealthErr)

			server := NewHTTPServer(cfg, mockClient)
			server.registerRoutes()

			req := httptest.NewRequest("GET", "/health", nil)
			w := httptest.NewRecorder()
			server.mux.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			var response HealthResponse
			err := json.NewDecoder(w.Body).Decode(&response)
			assert.NoError(t, err)

			assert.Equal(t, tc.expectedOK, response.TemporalOK)
			if tc.checkHealthErr != nil {
				assert.Equal(t, tc.checkHealthErr.Error(), response.TemporalMsg)
				assert.Equal(t, "Degraded", response.Status)
			} else {
				assert.Equal(t, "OK", response.Status)
				assert.Empty(t, response.TemporalMsg)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

func TestHTTPServer_StartStop(t *testing.T) {
	cfg := config.ServerConfig{
		Host:   "localhost",
		Port:   0,
		APIKey: "test-api-key",
		Temporal: config.TemporalClientConfig{
			Address:   "localhost:7233",
			Namespace: "test-namespace",
		},
	}

	mockClient := &mocks.Client{}
	mockClient.On("Close").Return()

	server := NewHTTPServer(cfg, mockClient)
	server.server.Addr = "localhost:0"

	ctx := context.Background()

	err := server.Start(ctx)
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	err = server.Stop(ctx)
	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
}

func TestHTTPServer_APIEndpoints(t *testing.T) {
	cfg := config.ServerConfig{
		Host:   "localhost",
		Port:   8080,
		APIKey: "test-api-key",
		Temporal: config.TemporalClientConfig{
			Address:   "localhost:7233",
			Namespace: "test-namespace",
		},
	}

	mockClient := &mocks.Client{}

	server := NewHTTPServer(cfg, mockClient)
	server.registerRoutes()

	testCases := []struct {
		name           string
		method         string
		path           string
		apiKey         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Valid API key for ticket summary",
			method:         "GET",
			path:           "/api/v1/ticket/summary",
			apiKey:         "test-api-key",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":"","status":"open"}`,
		},
		{
			name:           "Invalid API key for ticket summary",
			method:         "GET",
			path:           "/api/v1/ticket/summary",
			apiKey:         "wrong-key",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error": "Invalid or missing API key"}`,
		},
		{
			name:           "Missing API key for ticket summary",
			method:         "GET",
			path:           "/api/v1/ticket/summary",
			apiKey:         "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error": "Invalid or missing API key"}`,
		},
		{
			name:           "Valid API key for update ticket",
			method:         "POST",
			path:           "/api/v1/ticket",
			apiKey:         "test-api-key",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":"","status":"updated"}`,
		},
		{
			name:           "Valid API key for organization summary",
			method:         "GET",
			path:           "/api/v1/organization/summary",
			apiKey:         "test-api-key",
			expectedStatus: http.StatusNoContent,
			expectedBody:   ``,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			if tc.apiKey != "" {
				req.Header.Set(APIKeyHeader, tc.apiKey)
			}
			w := httptest.NewRecorder()

			server.mux.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.Equal(t, tc.expectedBody, w.Body.String())
		})
	}
}
