package webhook

import (
	"github.com/taonic/ticketfu/zendesk"
)

type Activity struct {
	zClient zendesk.Client
}

func NewActivity(zClient zendesk.Client) *Activity {
	return &Activity{
		zClient: zClient,
	}
}
