package worker

import (
	"context"

	"github.com/taonic/ticketfu/config"
	"github.com/taonic/ticketfu/genai"
	"github.com/taonic/ticketfu/temporal"
	"github.com/taonic/ticketfu/worker/org"
	"github.com/taonic/ticketfu/worker/ticket"
	"github.com/taonic/ticketfu/worker/webhook"
	"github.com/taonic/ticketfu/zendesk"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/server/common/log"
	"go.temporal.io/server/common/log/tag"
	"go.uber.org/fx"
)

const (
	TaskQueue = "ticketfu-queue"
)

type Worker struct {
	worker.Worker
	logger               log.Logger
	config               config.WorkerConfig
	ticketActivity       *ticket.Activity
	organizationActivity *org.Activity
	webhookActivities    *webhook.Activity
	tClient              client.Client
}

func NewWorker(
	config config.WorkerConfig,
	logger log.Logger,
	webhookActivity *webhook.Activity,
	ticketActivity *ticket.Activity,
	organizationActivity *org.Activity,
	tClient client.Client,
) *Worker {
	worker := worker.New(tClient, TaskQueue, worker.Options{})

	// register webhook workflow and activities
	worker.RegisterWorkflow(webhook.WebhookWorkflow)
	worker.RegisterActivity(webhookActivity.CreateWebhook)
	worker.RegisterActivity(webhookActivity.CreateTrigger)

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
		logger:               logger,
		config:               config,
		ticketActivity:       ticketActivity,
		organizationActivity: organizationActivity,
		tClient:              tClient,
	}
}

// Start initializes and starts the worker
func (w *Worker) OnStart(ctx context.Context) error {
	w.logger.Info("Starting worker")
	err := w.Start()
	if err != nil {
		w.logger.Fatal("Unable to start worker", tag.Error(err))
	}
	return nil
}

// Stop gracefully shuts down the worker
func (w *Worker) OnStop(ctx context.Context) error {
	w.logger.Info("Stopping worker")
	w.Stop()
	return nil
}

// Module registers the worker with fx
var Module = fx.Options(
	fx.Provide(NewWorker),
	fx.Provide(temporal.NewClient),
	fx.Provide(zendesk.NewClient),
	fx.Provide(genai.NewAPI),
	fx.Provide(webhook.NewActivity),
	fx.Provide(ticket.NewActivity),
	fx.Provide(org.NewActivity),
	fx.Invoke(func(lc fx.Lifecycle, worker *Worker) {
		lc.Append(fx.Hook{
			OnStart: worker.OnStart,
			OnStop:  worker.OnStop,
		})
	}),
)
