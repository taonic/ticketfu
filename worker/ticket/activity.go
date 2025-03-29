package ticket

import (
	"github.com/taonic/ticketfu/genai"
	"github.com/taonic/ticketfu/zendesk"
	"go.temporal.io/sdk/client"
)

type Activity struct {
	tClient client.Client
	zClient zendesk.Client
	genAPI  genai.API
}

func NewActivity(tClient client.Client, zClient zendesk.Client, genAPI genai.API) *Activity {
	return &Activity{
		tClient: tClient,
		zClient: zClient,
		genAPI:  genAPI,
	}
}
