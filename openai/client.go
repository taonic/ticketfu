package openai

import (
	"github.com/sashabaranov/go-openai"
	"github.com/taonic/ticketfu/config"
)

func NewClient(config config.OpenAIConfig) *openai.Client {
	return openai.NewClient(config.OpenAIAPIKey)
}
