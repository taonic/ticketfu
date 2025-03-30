package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/taonic/ticketfu/config"
	"go.temporal.io/sdk/client"
	"go.temporal.io/server/common/log"
	"go.temporal.io/server/common/log/tag"
)

// HTTPServer encapsulates the HTTP server functionality
type HTTPServer struct {
	logger         log.Logger
	server         *http.Server
	mux            *http.ServeMux
	config         config.ServerConfig
	temporalClient client.Client
}

// NewHTTPServer creates a new HTTP server with configured mux router
func NewHTTPServer(config config.ServerConfig, temporalClient client.Client, logger log.Logger) *HTTPServer {
	mux := http.NewServeMux()

	// Create the server with the mux as the handler
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &HTTPServer{
		logger:         logger,
		server:         server,
		mux:            mux,
		config:         config,
		temporalClient: temporalClient,
	}
}

// Start begins listening and serving HTTP requests
func (h *HTTPServer) Start(ctx context.Context) error {
	// Register routes
	h.registerRoutes()

	// Start the server in a goroutine
	go func() {
		if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			h.logger.Error("HTTP server encountered error", tag.Error(err))
		}
	}()

	h.logger.Info("HTTP server started", tag.Address(h.server.Addr))
	return nil
}

// Stop gracefully shuts down the HTTP server
func (h *HTTPServer) Stop(ctx context.Context) error {
	// Create a deadline for shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// Close the Temporal client
	if h.temporalClient != nil {
		h.temporalClient.Close()
	}

	// Attempt graceful shutdown
	if err := h.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("HTTP server shutdown error: %w", err)
	}

	h.logger.Info("HTTP server stopped")
	return nil
}

// registerRoutes sets up all the HTTP routes for the server
func (h *HTTPServer) registerRoutes() {
	// Health check endpoint
	h.mux.HandleFunc("GET /health", h.handleHealthCheck)

	// API routes
	verifyAPIKey := APIKeyMiddleware(h.config.APIKey)
	h.mux.HandleFunc("GET  /api/v1/ticket/summary", verifyAPIKey(h.handleGetTicket))
	h.mux.HandleFunc("POST /api/v1/ticket", verifyAPIKey(h.handleUpdateTicket))
	h.mux.HandleFunc("GET /api/v1/organization/summary", verifyAPIKey(h.handleGetOrganization))
}

func (h *HTTPServer) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	response := HealthResponse{
		Status:     "OK",
		TemporalOK: true,
	}

	// Check Temporal service health
	if h.temporalClient != nil {
		if _, err := h.temporalClient.CheckHealth(ctx, nil); err != nil {
			response.Status = "Degraded"
			response.TemporalOK = false
			response.TemporalMsg = err.Error()
		}
	} else {
		response.Status = "Degraded"
		response.TemporalOK = false
		response.TemporalMsg = "Temporal client not initialized"
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if response.Status != "OK" {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(response)
}

func (h *HTTPServer) handleGetTicket(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement fetching a specific ticket
	id := r.PathValue("id")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"id":"%s","status":"open"}`, id)))
}

func (h *HTTPServer) handleGetOrganization(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement get organization
	w.WriteHeader(http.StatusNoContent)
}
