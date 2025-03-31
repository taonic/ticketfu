package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
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
	h.server.Handler = h.registerRoutes()

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
func (h *HTTPServer) registerRoutes() *mux.Router {
	r := mux.NewRouter()
	// Health check endpoint
	r.HandleFunc("/health", h.handleHealthCheck).Methods("GET")

	// API routes
	verifyAPIKey := APIKeyMiddleware(h.config.APIToken)
	r.HandleFunc("/api/v1/ticket/{ticketId}/summary", verifyAPIKey(h.handleGetTicket)).Methods("GET")
	r.HandleFunc("/api/v1/ticket", verifyAPIKey(h.handleUpdateTicket)).Methods("POST")
	r.HandleFunc("/api/v1/organization/{orgId}/summary", verifyAPIKey(h.handleGetOrganization)).Methods("GET")

	return r
}
