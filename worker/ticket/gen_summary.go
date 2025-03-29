package ticket

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/genai"
)

type (
	GenSummaryInput struct {
		Ticket Ticket
	}

	GenSummaryOutput struct {
		Summary string
	}
)

func (a *Activity) GenSummary(ctx context.Context, input GenSummaryInput) (*GenSummaryOutput, error) {
	ticket := cleanse(input.Ticket)
	ticketJSON, err := json.Marshal(ticket)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ticket to JSON: %w", err)
	}

	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{Parts: []*genai.Part{{Text: a.geminiAPI.TicketSummaryPrompt}}},
	}

	result, err := a.geminiAPI.Client.Models.GenerateContent(
		ctx,
		a.geminiAPI.Model,
		genai.Text(string(ticketJSON)),
		config,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate %w", err)
	}
	output := GenSummaryOutput{Summary: result.Text()}

	return &output, nil
}

func cleanse(ticket Ticket) Ticket {
	ticket.Summary = ""
	ticket.NextCursor = ""
	return ticket
}
