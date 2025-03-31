package org

import (
	"context"
	"encoding/json"
	"fmt"
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

	result, err := a.genAPI.GenerateContent(ctx, a.genAPI.GetConfig().OrgSummaryPrompt, string(organizationJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to generate %w", err)
	}
	output := GenSummaryOutput{Summary: result}

	return &output, nil
}
