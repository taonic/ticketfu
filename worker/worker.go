package worker

import (
	"context"
	"fmt"

	gozendesk "github.com/nukosuke/go-zendesk/zendesk"
	"github.com/taonic/ticketfu/config"
	"github.com/taonic/ticketfu/zendesk"
	"go.uber.org/fx"
)

type Worker struct {
	config  config.WorkerConfig
	zClient *gozendesk.Client
}

func NewWorker(config config.WorkerConfig, zClient *gozendesk.Client) *Worker {
	return &Worker{
		config:  config,
		zClient: zClient,
	}
}

// Start initializes and starts the worker
func (w *Worker) Start(ctx context.Context) error {
	fmt.Println("Starting worker")
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
	fx.Provide(zendesk.NewClient),
	fx.Invoke(func(lc fx.Lifecycle, worker *Worker) {
		lc.Append(fx.Hook{
			OnStart: worker.Start,
			OnStop:  worker.Stop,
		})
	}),
)
