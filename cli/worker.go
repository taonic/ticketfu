package cli

import (
	"context"
	"fmt"

	"github.com/taonic/ticketiq/config"
	"github.com/taonic/ticketiq/worker"
	"github.com/urfave/cli/v2"
	"go.temporal.io/server/common/log"
	"go.uber.org/fx"
)

// Worker-specific flags
var workerFlags = append([]cli.Flag{
	&cli.StringFlag{
		Name:     FlagConfig,
		Aliases:  []string{"c"},
		EnvVars:  []string{"CONFIG_PATH"},
		Usage:    "path to worker config file",
		Required: false,
	},
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
}, commonFlags...)

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
	// Otherwise, build config from CLI flags
	// This is just a simple example - extend as needed
	return config.WorkerConfig{
		QueueName: ctx.String(FlagWorkerQueue),
		Threads:   ctx.Int(FlagWorkerThreads),
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
