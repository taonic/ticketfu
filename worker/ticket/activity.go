package ticket

import (
	"github.com/nukosuke/go-zendesk/zendesk"
	"github.com/sashabaranov/go-openai"
)

type Activity struct {
	zClient *zendesk.Client
	oClient *openai.Client
}

func NewActivity(zClient *zendesk.Client, oClient *openai.Client) *Activity {
	return &Activity{
		zClient: zClient,
		oClient: oClient,
	}
}
