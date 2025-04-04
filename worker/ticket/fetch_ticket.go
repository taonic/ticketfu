package ticket

import (
	"context"
	"strconv"

	"golang.org/x/sync/errgroup"
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

	ticket := Ticket{
		ID:             rawTicket.ID,
		Subject:        rawTicket.Subject,
		Description:    rawTicket.Description,
		Priority:       rawTicket.Priority,
		Status:         rawTicket.Status,
		OrganizationID: rawTicket.OrganizationID,
		CreatedAt:      rawTicket.CreatedAt,
		UpdatedAt:      rawTicket.UpdatedAt,
	}

	g, ctx := errgroup.WithContext(ctx)

	var requesterName string
	g.Go(func() error {
		requester, err := a.zClient.GetUser(ctx, rawTicket.RequesterID)
		if err != nil {
			return err
		}
		requesterName = requester.Name
		return nil
	})

	var assigneeName string
	g.Go(func() error {
		assignee, err := a.zClient.GetUser(ctx, rawTicket.AssigneeID)
		if err != nil {
			return err
		}
		assigneeName = assignee.Name
		return nil
	})

	var organizationName string
	if rawTicket.OrganizationID != 0 {
		g.Go(func() error {
			organization, err := a.zClient.GetOrganization(ctx, rawTicket.OrganizationID)
			if err != nil {
				return err
			}
			organizationName = organization.Name
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	ticket.Requester = requesterName
	ticket.Assignee = assigneeName
	ticket.OrganizationName = organizationName

	return &FetchTicketOutput{Ticket: ticket}, nil
}
