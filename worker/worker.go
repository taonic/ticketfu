package worker

import (
	"context"
	"fmt"
	"log"

	"github.com/taonic/ticketfu/config"
	"github.com/taonic/ticketfu/genai"
	"github.com/taonic/ticketfu/temporal"
	"github.com/taonic/ticketfu/worker/org"
	"github.com/taonic/ticketfu/worker/ticket"
	"github.com/taonic/ticketfu/zendesk"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

const (
	TaskQueue = "ticketfu-queue"
)

type Worker struct {
	worker.Worker
	config               config.WorkerConfig
	ticketActivity       *ticket.Activity
	organizationActivity *org.Activity
	tClient              client.Client
}

func NewWorker(config config.WorkerConfig, ticketActivity *ticket.Activity, organizationActivity *org.Activity, tClient client.Client) *Worker {
	worker := worker.New(tClient, TaskQueue, worker.Options{})

	// register ticket workflow and activities
	worker.RegisterWorkflow(ticket.TicketWorkflow)
	worker.RegisterActivity(ticketActivity.FetchTicket)
	worker.RegisterActivity(ticketActivity.FetchComments)
	worker.RegisterActivity(ticketActivity.GenTicketSummary)
	worker.RegisterActivity(ticketActivity.SignalOrganization)

	// register org workflow and activities
	worker.RegisterWorkflow(org.OrganizationWorkflow)
	worker.RegisterActivity(organizationActivity.FetchOrganization)
	worker.RegisterActivity(organizationActivity.GenOrgSummary)

	return &Worker{
		Worker:               worker,
		config:               config,
		ticketActivity:       ticketActivity,
		organizationActivity: organizationActivity,
		tClient:              tClient,
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
	fx.Provide(genai.NewAPI),
	fx.Provide(ticket.NewActivity),
	fx.Provide(org.NewActivity),
	fx.Invoke(func(lc fx.Lifecycle, worker *Worker) {
		lc.Append(fx.Hook{
			OnStart: worker.Start,
			OnStop:  worker.Stop,
		})
	}),
)
