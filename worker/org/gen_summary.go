package org

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/genai"
)

type (
	GenSummaryInput struct {
		Organization Organization
	}

	GenSummaryOutput struct {
		Summary string
	}
)

func (a *Activity) GenOrgSummary(ctx context.Context, input GenSummaryInput) (*GenSummaryOutput, error) {
	organizationJSON, err := json.Marshal(input.Organization)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal organization to JSON: %w", err)
	}

	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{Parts: []*genai.Part{{Text: a.genAPI.GetConfig().OrgSummaryPrompt}}},
	}

	result, err := a.genAPI.GenerateContent(
		ctx,
		a.genAPI.GetConfig().GeminiModel,
		genai.Text(string(organizationJSON)),
		config,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate %w", err)
	}
	output := GenSummaryOutput{Summary: result.Text()}

	return &output, nil
}
