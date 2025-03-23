package cli

import (
	"context"
	"fmt"

	"github.com/taonic/ticketiq/config"
	"github.com/taonic/ticketiq/server"
	"github.com/urfave/cli/v2"
	"go.temporal.io/server/common/log"
	"go.uber.org/fx"
)

// Server-specific flags
var serverFlags = append([]cli.Flag{
	&cli.StringFlag{
		Name:     FlagConfig,
		Aliases:  []string{"c"},
		EnvVars:  []string{"CONFIG_PATH"},
		Usage:    "path to server config file",
		Required: false,
	},
	&cli.StringFlag{
		Name:    FlagServerHost,
		EnvVars: []string{"SERVER_HOST"},
		Usage:   "server host address",
		Value:   "localhost",
	},
	&cli.IntFlag{
		Name:    FlagServerPort,
		EnvVars: []string{"SERVER_PORT"},
		Usage:   "server port",
		Value:   8080,
	},
}, commonFlags...)

// NewServerCommand creates a new server command with subcommands
func NewServerCommand() *cli.Command {
	return &cli.Command{
		Name:  "server",
		Usage: "Server commands",
		Subcommands: []*cli.Command{
			{
				Name:    "start",
				Aliases: []string{"s"},
				Flags:   serverFlags,
				Action:  startServer,
			},
		},
	}
}

// startServer is the action for the server start command
func startServer(c *cli.Context) error {
	app, err := NewServerApp(c)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start and wait for app to complete
	if err := app.Start(ctx); err != nil {
		return err
	}

	<-app.Done()
	return nil
}

// NewServerConfig creates a ServerConfig from CLI context
func NewServerConfig(ctx *cli.Context) (config.ServerConfig, error) {
	// Otherwise, build config from CLI flags
	// This is just a simple example - extend as needed
	return config.ServerConfig{
		Host: ctx.String(FlagServerHost),
	}, nil
}

// NewServerApp creates an fx application for the server command
func NewServerApp(ctx *cli.Context) (*fx.App, error) {
	var logCfg log.Config
	if logLevel := ctx.String(FlagLogLevel); len(logLevel) != 0 {
		logCfg.Level = logLevel
	}

	// Create server config from CLI context
	serverConfig, err := NewServerConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create server config: %w", err)
	}

	app := fx.New(
		// Provide logger
		fx.Provide(func() log.Logger {
			return log.NewZapLogger(log.BuildZapLogger(logCfg))
		}),

		// Provide the server config directly
		fx.Supply(serverConfig),

		// Include modules
		server.Module,
	)

	return app, nil
}
