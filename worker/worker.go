package worker

import (
	"context"
	"fmt"

	gozendesk "github.com/nukosuke/go-zendesk/zendesk"
	goopenai "github.com/sashabaranov/go-openai"
	"github.com/taonic/ticketfu/config"
	"github.com/taonic/ticketfu/openai"
	"github.com/taonic/ticketfu/zendesk"
	"go.uber.org/fx"
)

type Worker struct {
	config  config.WorkerConfig
	zClient *gozendesk.Client
	oClient *goopenai.Client
}

func NewWorker(config config.WorkerConfig, zClient *gozendesk.Client, oClient *goopenai.Client) *Worker {
	return &Worker{
		config:  config,
		zClient: zClient,
		oClient: oClient,
	}
}

// Start initializes and starts the worker
func (w *Worker) Start(ctx context.Context) error {
	fmt.Println("Testing openai client")
	resp, err := w.oClient.CreateChatCompletion(
		context.Background(),
		goopenai.ChatCompletionRequest{
			Model: goopenai.GPT3Dot5Turbo,
			Messages: []goopenai.ChatCompletionMessage{
				{
					Role:    goopenai.ChatMessageRoleUser,
					Content: "What is Go programming best known for?",
				},
			},
		},
	)

	if err != nil {
		panic(err)
	}

	fmt.Println("🤖", resp.Choices[0].Message.Content)

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
	fx.Provide(openai.NewClient),
	fx.Invoke(func(lc fx.Lifecycle, worker *Worker) {
		lc.Append(fx.Hook{
			OnStart: worker.Start,
			OnStop:  worker.Stop,
		})
	}),
)
