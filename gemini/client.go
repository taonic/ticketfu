package gemini

import (
	"context"

	"github.com/taonic/ticketfu/config"
	"google.golang.org/genai"
)

func NewClient(config config.AIConfig) (*genai.Client, error) {
	return genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  config.GeminiAPIKey,
		Backend: genai.BackendGeminiAPI,
	})
}
