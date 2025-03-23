package server

import (
	"github.com/taonic/ticketfu/config"
	"github.com/taonic/ticketfu/temporal"
	"go.temporal.io/sdk/client"
)

// NewTemporalClient creates a Temporal client for the server
func NewTemporalClient(config config.ServerConfig) (client.Client, error) {
	return temporal.NewClient(config.Temporal)
}
