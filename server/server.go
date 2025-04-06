package server

import (
	"context"
	"fmt"

	"github.com/taonic/ticketfu/config"
	"github.com/taonic/ticketfu/temporal"
	"go.temporal.io/sdk/client"
	"go.temporal.io/server/common/log"
	"go.uber.org/fx"
)

type Server struct {
	config         config.ServerConfig
	httpServer     *HTTPServer
	temporalClient client.Client
	logger         log.Logger
}

func NewServer(config config.ServerConfig, httpServer *HTTPServer, temporalClient client.Client, logger log.Logger) *Server {
	return &Server{
		config:         config,
		httpServer:     httpServer,
		temporalClient: temporalClient,
		logger:         logger,
	}
}

// OnStart initializes and starts the server
func (s *Server) OnStart(ctx context.Context) error {
	s.logger.Info("Starting server")

	// Try to create Zendesk webhook on each server start.
	s.BootstrapZendeskWebhook(ctx)

	if err := s.httpServer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	return nil
}

// OnStop gracefully shuts down the server
func (s *Server) OnStop(ctx context.Context) error {
	s.logger.Info("Stopping server...")

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
			OnStart: server.OnStart,
			OnStop:  server.OnStop,
		})
	}),
)
