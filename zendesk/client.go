package zendesk

import (
	"context"
	"fmt"

	"github.com/taonic/ticketfu/config"

	"github.com/nukosuke/go-zendesk/zendesk"
)

type Client interface {
	GetTicket(ctx context.Context, id int64) (zendesk.Ticket, error)
	GetTicketCommentsCBP(ctx context.Context, opts *zendesk.CBPOptions) ([]zendesk.TicketComment, zendesk.CursorPaginationMeta, error)
	GetUser(ctx context.Context, userID int64) (zendesk.User, error)
	GetOrganization(ctx context.Context, orgID int64) (zendesk.Organization, error)
}

func NewClient(config config.ZendeskConfig) (Client, error) {
	client, err := zendesk.NewClient(nil)
	if err != nil {
		return nil, fmt.Errorf("Unable to create Zendesk client: %w", err)
	}
	client.SetSubdomain(config.ZendeskSubdomain)
	client.SetCredential(zendesk.NewAPITokenCredential(config.ZendeskEmail, config.ZendeskToken))

	return client, nil
}
