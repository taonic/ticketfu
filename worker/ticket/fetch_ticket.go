package ticket

import (
	"context"
	"strconv"
)

type (
	FetchTicketInput struct {
		ID string
	}

	FetchTicketOutput struct {
		Ticket Ticket
	}
)

func (a *Activity) FetchTicket(ctx context.Context, input FetchTicketInput) (*FetchTicketOutput, error) {
	num, err := strconv.ParseInt(input.ID, 10, 64)
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

	ticket := Ticket{
		ID:             rawTicket.ID,
		Subject:        rawTicket.Subject,
		Description:    rawTicket.Description,
		Priority:       rawTicket.Priority,
		Status:         rawTicket.Status,
		OrganizationID: rawTicket.OrganizationID,
		Requester:      requester.Name,
		Assignee:       assignee.Name,
		CreatedAt:      rawTicket.CreatedAt,
		UpdatedAt:      rawTicket.UpdatedAt,
	}

	if ticket.OrganizationID != 0 {
		organization, err := a.zClient.GetOrganization(ctx, rawTicket.OrganizationID)
		if err != nil {
			return nil, err
		}
		ticket.OrganizationName = organization.Name
	}

	return &FetchTicketOutput{Ticket: ticket}, nil
}
