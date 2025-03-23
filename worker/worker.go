package worker

import (
	"context"
	"fmt"

	"github.com/taonic/ticketiq/config"
	"go.uber.org/fx"
)

type Worker struct {
	config config.WorkerConfig
}

func NewWorker(config config.WorkerConfig) *Worker {
	return &Worker{
		config: config,
	}
}

// Start initializes and starts the worker
func (w *Worker) Start(ctx context.Context) error {
	fmt.Println("Starting worker with config:", w.config)
	// Actual worker implementation goes here
	return nil
}

// Stop gracefully shuts down the worker
func (w *Worker) Stop(ctx context.Context) error {
	fmt.Println("Stopping worker...")
	// Graceful shutdown implementation goes here
	return nil
}

// Module registers the worker with fx
var Module = fx.Options(
	fx.Provide(NewWorker),
	fx.Invoke(func(lc fx.Lifecycle, worker *Worker) {
		lc.Append(fx.Hook{
			OnStart: worker.Start,
			OnStop:  worker.Stop,
		})
	}),
)
