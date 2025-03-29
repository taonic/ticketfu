package ticket

import (
	"context"
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
	result, err := a.geminiAPI.Client.Models.GenerateContent(
		ctx,
		a.geminiAPI.Model,
		genai.Text(input.Ticket.Description),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate %w", err)
	}
	output := GenSummaryOutput{Summary: result.Text()}

	return &output, nil
}
