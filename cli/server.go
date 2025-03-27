package cli

import (
	"context"

	"github.com/taonic/ticketfu/config"
	"github.com/taonic/ticketfu/server"
	"github.com/urfave/cli/v2"
	"go.temporal.io/server/common/log"
	"go.uber.org/fx"
)

const (
	// Server-specific flags
	FlagServerHost = "host"
	FlagServerPort = "port"
	FlagAPIKey     = "api-key"
)

// Server-specific flags
var serverFlags = append(append([]cli.Flag{
	&cli.StringFlag{
		Name:    FlagServerHost,
		EnvVars: []string{"HOST"},
		Usage:   "server host address",
		Value:   "0.0.0.0",
	},
	&cli.IntFlag{
		Name:    FlagServerPort,
		EnvVars: []string{"PORT"},
		Usage:   "server port",
		Value:   8080,
	},
	&cli.StringFlag{
		Name:     FlagAPIKey,
		Aliases:  []string{"k"},
		EnvVars:  []string{"API_KEY"},
		Usage:    "API key for authenticating requests",
		Required: true,
	},
}, temporalFlags...), commonFlags...)

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

// NewServerApp creates an fx application for the server command
func NewServerApp(ctx *cli.Context) (*fx.App, error) {
	var logCfg log.Config
	if logLevel := ctx.String(FlagLogLevel); len(logLevel) != 0 {
		logCfg.Level = logLevel
	}

	serverConfig := config.ServerConfig{
		Host:   ctx.String(FlagServerHost),
		Port:   ctx.Int(FlagServerPort),
		APIKey: ctx.String(FlagAPIKey),
	}

	temporalClientConfig := config.TemporalClientConfig{
		Address:     ctx.String(FlagTemporalAddress),
		Namespace:   ctx.String(FlagTemporalNamespace),
		APIKey:      ctx.String(FlagTemporalAPIKey),
		TLSCertPath: ctx.String(FlagTemporalTLSCert),
		TLSKeyPath:  ctx.String(FlagTemporalTLSKey),
	}

	app := fx.New(
		fx.Provide(func() log.Logger {
			return log.NewZapLogger(log.BuildZapLogger(logCfg))
		}),
		fx.Supply(
			serverConfig,
			temporalClientConfig,
		),
		server.Module,
	)

	return app, nil
}
