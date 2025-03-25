package zendesk

import (
	"fmt"

	"github.com/taonic/ticketfu/config"

	"github.com/nukosuke/go-zendesk/zendesk"
)

func NewClient(config config.WorkerConfig) (*zendesk.Client, error) {
	client, err := zendesk.NewClient(nil)
	if err != nil {
		return nil, fmt.Errorf("Unable to create Zendesk client: %w", err)
	}
	client.SetSubdomain(config.ZendeskSubdomain)
	client.SetCredential(zendesk.NewAPITokenCredential(config.ZendeskEmail, config.ZendeskToken))

	return client, nil
}
