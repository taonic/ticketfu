package org

import (
	"context"
)

type (
	FetchOrganizationInput struct {
		ID int64
	}

	FetchOrganizationOutput struct {
		Organization Organization
	}
)

func (a *Activity) FetchOrganization(ctx context.Context, input FetchOrganizationInput) (*FetchOrganizationOutput, error) {
	rawOrganization, err := a.zClient.GetOrganization(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	organization := Organization{
		ID:      rawOrganization.ID,
		Name:    rawOrganization.Name,
		Details: rawOrganization.Details,
		Notes:   rawOrganization.Notes,
	}

	return &FetchOrganizationOutput{Organization: organization}, nil
}
