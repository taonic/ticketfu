package ticket

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
