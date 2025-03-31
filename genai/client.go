package genai

import (
	"context"
	"fmt"
	"strings"

	"github.com/taonic/ticketfu/config"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/openai"
	"go.temporal.io/server/common/log"
)

type genAI struct {
	logger log.Logger
	model  llms.Model
	Config config.AIConfig
}

type API interface {
	GenerateContent(ctx context.Context, instruction, content string) (string, error)
	GetConfig() config.AIConfig
}

func NewAPI(logger log.Logger, config config.AIConfig) (API, error) {
	ctx := context.Background()
	genAI := genAI{
		logger: logger,
		Config: config,
	}

	// configure at least one model based on the provided API key
	if config.OpenAIAPIKey != "" {
		model, err := openai.New(openai.WithToken(config.OpenAIAPIKey), openai.WithModel(config.OpenAIModel))
		if err != nil {
			return nil, err
		}
		genAI.model = model
		logger.Debug("Configured OpenAI model.")
	} else if config.GeminiAPIKey != "" {
		model, err := googleai.New(ctx, googleai.WithAPIKey(config.GeminiAPIKey), googleai.WithDefaultModel(config.GeminiModel))
		if err != nil {
			return nil, err
		}
		genAI.model = model
		logger.Debug("Configured Gemini model.")
	}

	if genAI.model == nil {
		return nil, fmt.Errorf("One of the OpenAPI or Gemini API key should be provided")
	}

	return &genAI, nil
}

func (a *genAI) GenerateContent(ctx context.Context, instruction, content string) (string, error) {
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, instruction),
		llms.TextParts(llms.ChatMessageTypeHuman, content),
	}
	var result strings.Builder
	_, err := a.model.GenerateContent(ctx, messages, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
		result.WriteString(string(chunk))
		return nil
	}))
	if err != nil {
		return "", fmt.Errorf("failed to generate content %w", err)
	}
	return result.String(), nil
}

func (a *genAI) GetConfig() config.AIConfig {
	return a.Config
}
