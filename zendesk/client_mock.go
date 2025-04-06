package zendesk

import (
	"context"

	"github.com/nukosuke/go-zendesk/zendesk"
	"github.com/stretchr/testify/mock"
)

// MockZendeskClient is a mock of the Zendesk client interface
type MockZendeskClient struct {
	mock.Mock
}

func (m *MockZendeskClient) GetTicket(ctx context.Context, id int64) (zendesk.Ticket, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(zendesk.Ticket), args.Error(1)
}

func (m *MockZendeskClient) GetUser(ctx context.Context, id int64) (zendesk.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(zendesk.User), args.Error(1)
}

func (m *MockZendeskClient) GetOrganization(ctx context.Context, id int64) (zendesk.Organization, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(zendesk.Organization), args.Error(1)
}

func (m *MockZendeskClient) GetTicketCommentsCBP(ctx context.Context, opts *zendesk.CBPOptions) ([]zendesk.TicketComment, zendesk.CursorPaginationMeta, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).([]zendesk.TicketComment), args.Get(1).(zendesk.CursorPaginationMeta), args.Error(2)
}

func (m *MockZendeskClient) CreateWebhook(ctx context.Context, hook *zendesk.Webhook) (*zendesk.Webhook, error) {
	args := m.Called(ctx, hook)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*zendesk.Webhook), args.Error(1)
}

func (m *MockZendeskClient) GetWebhook(ctx context.Context, id string) (*zendesk.Webhook, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*zendesk.Webhook), args.Error(1)
}

func (m *MockZendeskClient) CreateTrigger(ctx context.Context, trigger zendesk.Trigger) (zendesk.Trigger, error) {
	args := m.Called(ctx, trigger)
	return args.Get(0).(zendesk.Trigger), args.Error(1)
}
