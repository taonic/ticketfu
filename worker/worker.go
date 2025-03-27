package worker

import (
	"context"
	"fmt"
	"log"

	"github.com/taonic/ticketfu/config"
	"github.com/taonic/ticketfu/openai"
	"github.com/taonic/ticketfu/temporal"
	"github.com/taonic/ticketfu/temporal/activities"
	"github.com/taonic/ticketfu/temporal/workflows"
	"github.com/taonic/ticketfu/zendesk"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

type Worker struct {
	worker.Worker
	config   config.WorkerConfig
	activity *activities.Activity
	tClient  client.Client
}

func NewWorker(config config.WorkerConfig, activity *activities.Activity, tClient client.Client) *Worker {
	worker := worker.New(tClient, temporal.TaskQueue, worker.Options{})
	worker.RegisterWorkflow(workflows.TicketWorkflow)
	worker.RegisterActivity(activity.FetchTicket)
	return &Worker{
		Worker:   worker,
		config:   config,
		activity: activity,
		tClient:  tClient,
	}
}

// Start initializes and starts the worker
func (w *Worker) Start(ctx context.Context) error {
	fmt.Println("Starting worker")
	err := w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
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
	fx.Provide(temporal.NewClient),
	fx.Provide(zendesk.NewClient),
	fx.Provide(openai.NewClient),
	fx.Provide(activities.NewActivity),
	fx.Invoke(func(lc fx.Lifecycle, worker *Worker) {
		lc.Append(fx.Hook{
			OnStart: worker.Start,
			OnStop:  worker.Stop,
		})
	}),
)
