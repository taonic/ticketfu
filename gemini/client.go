package gemini

import (
	"context"

	"github.com/taonic/ticketfu/config"
	"google.golang.org/genai"
)

type API struct {
	Client *genai.Client
	Model  string
}

func NewAPI(config config.AIConfig) (*API, error) {
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  config.GeminiAPIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, err
	}

	api := API{
		Client: client,
		Model:  config.GeminiModel,
	}

	return &api, nil
}
