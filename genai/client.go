package genai

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
	"github.com/taonic/ticketfu/config"
	"google.golang.org/genai"
)

type genAI struct {
	gClient *genai.Client
	oClient *openai.Client
	Config  config.AIConfig
}

type API interface {
	GenerateContent(context.Context, string, []*genai.Content, *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error)
	GetConfig() config.AIConfig
}

func NewAPI(config config.AIConfig) (API, error) {
	genAI := genAI{Config: config}

	if config.OpenAIAPIKey != "" {
		genAI.oClient = openai.NewClient(config.OpenAIAPIKey)
	}

	if config.GeminiAPIKey != "" {
		client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
			APIKey:  config.GeminiAPIKey,
			Backend: genai.BackendGeminiAPI,
		})
		if err != nil {
			return nil, err
		}

		genAI.gClient = client
	}

	if genAI.oClient == nil && genAI.gClient == nil {
		return nil, fmt.Errorf("One of OpenAPI or Gemini API key should be provided")
	}

	return &genAI, nil
}

func (a *genAI) GenerateContent(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
	return a.gClient.Models.GenerateContent(ctx, model, contents, config)
}

func (a *genAI) GetConfig() config.AIConfig {
	return a.Config
}
