package server

import (
	"context"
	"fmt"

	"github.com/taonic/ticketiq/config"
	"go.uber.org/fx"
)

type Server struct {
	config config.ServerConfig
}

func NewServer(config config.ServerConfig) *Server {
	return &Server{
		config: config,
	}
}

// Start initializes and starts the server
func (s *Server) Start(ctx context.Context) error {
	fmt.Println("Starting server with config:", s.config)
	// Actual server implementation goes here
	return nil
}

// Stop gracefully shuts down the server
func (s *Server) Stop(ctx context.Context) error {
	fmt.Println("Stopping server...")
	// Graceful shutdown implementation goes here
	return nil
}

// Module registers lifecycle hooks with fx
var Module = fx.Options(
	fx.Provide(NewServer),
	fx.Invoke(func(lc fx.Lifecycle, server *Server) {
		lc.Append(fx.Hook{
			OnStart: server.Start,
			OnStop:  server.Stop,
		})
	}),
)
