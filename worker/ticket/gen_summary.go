package ticket

import (
	"context"
	"encoding/json"
	"fmt"
)

type (
	GenSummaryInput struct {
		Ticket Ticket
	}

	GenSummaryOutput struct {
		Summary string
	}
)

func (a *Activity) GenTicketSummary(ctx context.Context, input GenSummaryInput) (*GenSummaryOutput, error) {
	ticket := cleanse(input.Ticket)
	ticketJSON, err := json.Marshal(ticket)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ticket to JSON: %w", err)
	}

	result, err := a.genAPI.GenerateContent(ctx, a.genAPI.GetConfig().TicketSummaryPrompt, string(ticketJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to generate %w", err)
	}
	output := GenSummaryOutput{Summary: result}

	return &output, nil
}

func cleanse(ticket Ticket) Ticket {
	ticket.Summary = ""
	ticket.NextCursor = ""
	return ticket
}
