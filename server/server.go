package server

import (
	"context"
	"fmt"

	"github.com/taonic/ticketfu/config"
	"github.com/taonic/ticketfu/temporal"
	"go.temporal.io/sdk/client"
	"go.uber.org/fx"
)

type Server struct {
	config         config.ServerConfig
	httpServer     *HTTPServer
	temporalClient client.Client
}

func NewServer(config config.ServerConfig, httpServer *HTTPServer, temporalClient client.Client) *Server {
	return &Server{
		config:         config,
		httpServer:     httpServer,
		temporalClient: temporalClient,
	}
}

// Start initializes and starts the server
func (s *Server) Start(ctx context.Context) error {
	fmt.Println("Starting server")

	// Start the HTTP server
	if err := s.httpServer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	return nil
}

// Stop gracefully shuts down the server
func (s *Server) Stop(ctx context.Context) error {
	fmt.Println("Stopping server...")

	// Stop the HTTP server
	if err := s.httpServer.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop HTTP server: %w", err)
	}

	return nil
}

// Module registers server components with fx
var Module = fx.Options(
	fx.Provide(temporal.NewClient),
	fx.Provide(NewHTTPServer),
	fx.Provide(NewServer),
	fx.Invoke(func(lc fx.Lifecycle, server *Server) {
		lc.Append(fx.Hook{
			OnStart: server.Start,
			OnStop:  server.Stop,
		})
	}),
)
