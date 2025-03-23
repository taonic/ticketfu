package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIKeyMiddleware(t *testing.T) {
	// Define the API key to use for testing
	const testAPIKey = "test-api-key"

	// Create a simple handler for testing
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}

	// Create middleware with the test API key
	middleware := APIKeyMiddleware(testAPIKey)

	// Wrap the test handler with the middleware
	wrappedHandler := middleware(testHandler)

	// Test cases
	testCases := []struct {
		name           string
		apiKey         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Valid API Key",
			apiKey:         testAPIKey,
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
		{
			name:           "Invalid API Key",
			apiKey:         "wrong-api-key",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error": "Invalid or missing API key"}`,
		},
		{
			name:           "Missing API Key",
			apiKey:         "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error": "Invalid or missing API key"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test request
			req := httptest.NewRequest("GET", "/test", nil)

			// Set API key if provided
			if tc.apiKey != "" {
				req.Header.Set(APIKeyHeader, tc.apiKey)
			}

			// Create a response recorder
			w := httptest.NewRecorder()

			// Call the wrapped handler
			wrappedHandler(w, req)

			// Check status code
			assert.Equal(t, tc.expectedStatus, w.Code)

			// Check response body
			assert.Equal(t, tc.expectedBody, w.Body.String())
		})
	}
}

func TestAPIKeyMiddleware_MultipleHandlers(t *testing.T) {
	// Define the API key to use for testing
	const testAPIKey = "test-api-key"

	// Create middleware with the test API key
	middleware := APIKeyMiddleware(testAPIKey)

	// Create two different handlers
	handler1 := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("handler1"))
	}

	handler2 := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("handler2"))
	}

	// Wrap both handlers with the middleware
	wrappedHandler1 := middleware(handler1)
	wrappedHandler2 := middleware(handler2)

	// Test both handlers
	req1 := httptest.NewRequest("GET", "/test1", nil)
	req1.Header.Set(APIKeyHeader, testAPIKey)
	w1 := httptest.NewRecorder()
	wrappedHandler1(w1, req1)

	req2 := httptest.NewRequest("GET", "/test2", nil)
	req2.Header.Set(APIKeyHeader, testAPIKey)
	w2 := httptest.NewRecorder()
	wrappedHandler2(w2, req2)

	// Check that both handlers were called correctly
	assert.Equal(t, http.StatusOK, w1.Code)
	assert.Equal(t, "handler1", w1.Body.String())

	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Equal(t, "handler2", w2.Body.String())
}
