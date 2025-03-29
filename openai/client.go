package openai

import (
	"github.com/sashabaranov/go-openai"
	"github.com/taonic/ticketfu/config"
)

func NewClient(config config.AIConfig) *openai.Client {
	return openai.NewClient(config.OpenAIAPIKey)
}
