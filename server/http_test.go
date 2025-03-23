package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/taonic/ticketfu/config"
)

func TestHTTPServer_RegisterRoutes(t *testing.T) {
	// Create a test config
	cfg := config.ServerConfig{
		Host:   "localhost",
		Port:   8080,
		APIKey: "test-api-key",
	}

	// Create a new HTTP server
	server := NewHTTPServer(cfg)

	// Register routes
	server.registerRoutes()

	// Test the health check endpoint
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	server.mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

func TestHTTPServer_StartStop(t *testing.T) {
	// Create a test config with a random available port
	cfg := config.ServerConfig{
		Host:   "localhost",
		Port:   0, // Use port 0 to get a random available port
		APIKey: "test-api-key",
	}

	// Create a new HTTP server
	server := NewHTTPServer(cfg)
	server.server.Addr = "localhost:0" // Ensure using a random available port

	// Create a context
	ctx := context.Background()

	// Start the server
	err := server.Start(ctx)
	assert.NoError(t, err)

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Stop the server
	err = server.Stop(ctx)
	assert.NoError(t, err)
}

func TestHTTPServer_HealthEndpoint(t *testing.T) {
	// Create a test config
	cfg := config.ServerConfig{
		Host:   "localhost",
		Port:   8080,
		APIKey: "test-api-key",
	}

	// Create a new HTTP server
	server := NewHTTPServer(cfg)
	server.registerRoutes()

	// Create a test request for the health endpoint
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Serve the request
	server.mux.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

func TestHTTPServer_APIEndpoints(t *testing.T) {
	// Create a test config
	cfg := config.ServerConfig{
		Host:   "localhost",
		Port:   8080,
		APIKey: "test-api-key",
	}

	// Create a new HTTP server
	server := NewHTTPServer(cfg)
	server.registerRoutes()

	// Test cases for each API endpoint
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
			// Create a test request
			req := httptest.NewRequest(tc.method, tc.path, nil)
			if tc.apiKey != "" {
				req.Header.Set(APIKeyHeader, tc.apiKey)
			}
			w := httptest.NewRecorder()

			// Serve the request
			server.mux.ServeHTTP(w, req)

			// Check the response
			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.Equal(t, tc.expectedBody, w.Body.String())
		})
	}
}
