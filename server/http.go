package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/taonic/ticketfu/config"
	"go.temporal.io/sdk/client"
)

// HTTPServer encapsulates the HTTP server functionality
type HTTPServer struct {
	server         *http.Server
	mux            *http.ServeMux
	config         config.ServerConfig
	temporalClient client.Client
}

// NewHTTPServer creates a new HTTP server with configured mux router
func NewHTTPServer(config config.ServerConfig, temporalClient client.Client) *HTTPServer {
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
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	fmt.Printf("HTTP server started on %s\n", h.server.Addr)
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

	fmt.Println("HTTP server stopped")
	return nil
}

// registerRoutes sets up all the HTTP routes for the server
func (h *HTTPServer) registerRoutes() {
	// Health check endpoint
	h.mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// API routes
	verifyAPIKey := APIKeyMiddleware(h.config.APIKey)
	h.mux.HandleFunc("GET  /api/v1/ticket/summary", verifyAPIKey(h.handleGetTicket))
	h.mux.HandleFunc("POST /api/v1/ticket", verifyAPIKey(h.handleUpdateTicket))
	h.mux.HandleFunc("GET /api/v1/organization/summary", verifyAPIKey(h.handleGetOrganization))
}

func (h *HTTPServer) handleGetTicket(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement fetching a specific ticket
	id := r.PathValue("id")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"id":"%s","status":"open"}`, id)))
}

func (h *HTTPServer) handleUpdateTicket(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement updating a ticket
	id := r.PathValue("id")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"id":"%s","status":"updated"}`, id)))
}

func (h *HTTPServer) handleGetOrganization(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement get organization
	w.WriteHeader(http.StatusNoContent)
}
