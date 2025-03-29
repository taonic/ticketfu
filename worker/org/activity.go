package org

import (
	"github.com/taonic/ticketfu/genai"
	"github.com/taonic/ticketfu/zendesk"
)

type Activity struct {
	zClient zendesk.Client
	genAPI  genai.API
}

func NewActivity(zClient zendesk.Client, genAPI genai.API) *Activity {
	return &Activity{
		zClient: zClient,
		genAPI:  genAPI,
	}
}
