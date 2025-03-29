package ticket

import (
	"context"
	"strconv"

	"github.com/taonic/ticketfu/zendesk"
)

func (a *Activity) FetchTicket(ctx context.Context, id string) (*zendesk.Ticket, error) {
	num, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, err
	}

	rawTicket, err := a.zClient.GetTicket(ctx, num)
	if err != nil {
		return nil, err
	}

	requester, err := a.zClient.GetUser(ctx, rawTicket.RequesterID)
	if err != nil {
		return nil, err
	}

	assignee, err := a.zClient.GetUser(ctx, rawTicket.AssigneeID)
	if err != nil {
		return nil, err
	}

	orgnization, err := a.zClient.GetOrganization(ctx, rawTicket.OrganizationID)
	if err != nil {
		return nil, err
	}

	ticket := zendesk.Ticket{
		ID:           rawTicket.ID,
		Subject:      rawTicket.Subject,
		Description:  rawTicket.Description,
		Priority:     rawTicket.Priority,
		Status:       rawTicket.Status,
		Organization: orgnization.Name,
		Requester:    requester.Name,
		Assignee:     assignee.Name,
		CreatedAt:    rawTicket.CreatedAt,
		UpdatedAt:    rawTicket.UpdatedAt,
	}

	return &ticket, nil
}
