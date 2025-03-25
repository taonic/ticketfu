package cli

import (
	"context"
	"fmt"

	"github.com/taonic/ticketfu/config"
	"github.com/taonic/ticketfu/worker"
	"github.com/urfave/cli/v2"
	"go.temporal.io/server/common/log"
	"go.uber.org/fx"
)

const (
	// Worker-specific flags
	FlagWorkerQueue   = "queue"
	FlagWorkerThreads = "threads"
)

// Worker-specific flags
var workerFlags = append(append(append([]cli.Flag{
	&cli.StringFlag{
		Name:    FlagWorkerQueue,
		EnvVars: []string{"WORKER_QUEUE"},
		Usage:   "worker queue name",
		Value:   "default",
	},
	&cli.IntFlag{
		Name:    FlagWorkerThreads,
		EnvVars: []string{"WORKER_THREADS"},
		Usage:   "number of worker threads",
		Value:   4,
	},
}, temporalFlags...), commonFlags...), zendeskFlags...)

// NewWorkerCommand creates a new worker command with subcommands
func NewWorkerCommand() *cli.Command {
	return &cli.Command{
		Name:  "worker",
		Usage: "Worker commands",
		Subcommands: []*cli.Command{
			{
				Name:    "start",
				Aliases: []string{"s"},
				Flags:   workerFlags,
				Action:  startWorker,
			},
		},
	}
}

// startWorker is the action for the worker start command
func startWorker(c *cli.Context) error {
	app, err := NewWorkerApp(c)
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

// NewWorkerConfig creates a WorkerConfig from CLI context
func NewWorkerConfig(ctx *cli.Context) (config.WorkerConfig, error) {
	// Build config from CLI flags
	return config.WorkerConfig{
		QueueName: ctx.String(FlagWorkerQueue),
		Threads:   ctx.Int(FlagWorkerThreads),
		Temporal: config.TemporalClientConfig{
			Address:     ctx.String(FlagTemporalAddress),
			Namespace:   ctx.String(FlagTemporalNamespace),
			APIKey:      ctx.String(FlagTemporalAPIKey),
			TLSCertPath: ctx.String(FlagTemporalTLSCert),
			TLSKeyPath:  ctx.String(FlagTemporalTLSKey),
		},
		ZendeskSubdomain: ctx.String(FlagZendeskSubdomain),
		ZendeskEmail:     ctx.String(FlagZendeskEmail),
		ZendeskToken:     ctx.String(FlagZendeskToken),
	}, nil
}

// NewWorkerApp creates an fx application for the worker command
func NewWorkerApp(ctx *cli.Context) (*fx.App, error) {
	var logCfg log.Config
	if logLevel := ctx.String(FlagLogLevel); len(logLevel) != 0 {
		logCfg.Level = logLevel
	}

	// Create worker config from CLI context
	workerConfig, err := NewWorkerConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create worker config: %w", err)
	}

	app := fx.New(
		// Provide logger
		fx.Provide(func() log.Logger {
			return log.NewZapLogger(log.BuildZapLogger(logCfg))
		}),

		// Provide the worker config directly
		fx.Supply(workerConfig),

		// Include modules
		worker.Module,
	)

	return app, nil
}
