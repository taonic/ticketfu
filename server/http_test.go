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
	"go.temporal.io/server/common/log"
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

	server := NewHTTPServer(cfg, mockClient, log.NewTestLogger())
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

			server := NewHTTPServer(cfg, mockClient, log.NewTestLogger())
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

	server := NewHTTPServer(cfg, mockClient, log.NewTestLogger())
	server.server.Addr = "localhost:0"

	ctx := context.Background()

	err := server.Start(ctx)
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	err = server.Stop(ctx)
	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
}
