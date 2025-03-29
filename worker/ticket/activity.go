package ticket

import (
	"github.com/sashabaranov/go-openai"
	"github.com/taonic/ticketfu/gemini"
	"github.com/taonic/ticketfu/zendesk"
)

type Activity struct {
	zClient   zendesk.Client
	oClient   *openai.Client
	geminiAPI *gemini.API
}

func NewActivity(zClient zendesk.Client, oClient *openai.Client, geminiAPI *gemini.API) *Activity {
	return &Activity{
		zClient:   zClient,
		oClient:   oClient,
		geminiAPI: geminiAPI,
	}
}
