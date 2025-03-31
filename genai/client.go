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
	"go.temporal.io/server/common/log/tag"
)

const (
	OpenAI   = "openai"
	GoogleAI = "googleai"
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
	if config.LLMAPIKey == "" {
		return nil, fmt.Errorf("llm-api-key is not provided")
	}

	ctx := context.Background()

	var model llms.Model
	var err error
	switch config.LLMProvider {
	case OpenAI:
		model, err = openai.New(openai.WithToken(config.LLMAPIKey), openai.WithModel(config.LLMModel))
	case GoogleAI:
		model, err = googleai.New(ctx, googleai.WithAPIKey(config.LLMAPIKey), googleai.WithDefaultModel(config.LLMModel))
	default:
		return nil, fmt.Errorf("unknown LLM provider: %s", config.LLMModel)
	}
	if model == nil || err != nil {
		return nil, fmt.Errorf("failed to initialize LLM for provider: %s %w", model, err)
	}

	logger.Info("Configured LLM", tag.Value(config.LLMModel))

	genAI := genAI{
		logger: logger,
		Config: config,
		model:  model,
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
