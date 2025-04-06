package webhook

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

func TestWebhookWorkflowSuite(t *testing.T) {
	suite.Run(t, new(WebhookWorkflowTestSuite))
}

type WebhookWorkflowTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
}

func (s *WebhookWorkflowTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

func (s *WebhookWorkflowTestSuite) TearDownTest() {
	s.env.AssertExpectations(s.T())
}

func (s *WebhookWorkflowTestSuite) TestSuccessfulWebhookCreation() {
	// Initial webhook with just base URL and API token
	webhook := Webhook{
		BaseURL:        "https://example.com",
		ServerAPIToken: "test-api-token",
	}

	// Mock successful create webhook activity
	s.env.OnActivity((*Activity)(nil).CreateWebhook, mock.Anything, mock.MatchedBy(func(input CreateWebhookInput) bool {
		return input.Webhook.BaseURL == "https://example.com" &&
			input.Webhook.ServerAPIToken == "test-api-token"
	})).Return(&CreateWebhookOutput{
		WebhookID: "webhook-123",
	}, nil).Once()

	// Mock successful create trigger activity
	s.env.OnActivity((*Activity)(nil).CreateTrigger, mock.Anything, CreateTriggerInput{
		WebhookID: "webhook-123",
	}).Return(&CreateTriggerOutput{
		TriggerID: "456",
	}, nil).Once()

	// Send signal to trigger webhooks creation
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(UpsertWebhookSignal, UpsertWebhookInput{
			Webhook: Webhook{
				BaseURL:        "https://example.com",
				ServerAPIToken: "test-api-token",
			},
		})
	}, time.Millisecond*100)

	// Add cancellation to complete the test
	s.env.RegisterDelayedCallback(func() {
		s.env.CancelWorkflow()
	}, time.Millisecond*300)

	// Execute workflow
	s.env.ExecuteWorkflow(WebhookWorkflow, webhook)

	// Verify workflow completed with cancellation
	s.True(s.env.IsWorkflowCompleted())
	var canErr *temporal.CanceledError
	s.True(errors.As(s.env.GetWorkflowError(), &canErr))
}

func (s *WebhookWorkflowTestSuite) TestExistingWebhookNoTriggerCreation() {
	// Initial webhook with existing ID
	webhook := Webhook{
		ID:             "existing-webhook-123",
		BaseURL:        "https://example.com",
		ServerAPIToken: "test-api-token",
	}

	// Mock successful create webhook activity that returns the existing webhook ID
	s.env.OnActivity((*Activity)(nil).CreateWebhook, mock.Anything, mock.MatchedBy(func(input CreateWebhookInput) bool {
		return input.Webhook.ID == "existing-webhook-123"
	})).Return(&CreateWebhookOutput{
		WebhookID: "existing-webhook-123",
	}, nil).Once()

	// We should NOT see a call to create trigger since the webhook already exists
	// No need to mock CreateTrigger

	// Send signal to trigger webhook check
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(UpsertWebhookSignal, UpsertWebhookInput{
			Webhook: Webhook{
				BaseURL:        "https://example.com",
				ServerAPIToken: "test-api-token",
			},
		})
	}, time.Millisecond*100)

	// Add cancellation to complete the test
	s.env.RegisterDelayedCallback(func() {
		s.env.CancelWorkflow()
	}, time.Millisecond*300)

	// Execute workflow
	s.env.ExecuteWorkflow(WebhookWorkflow, webhook)

	// Verify workflow completed with cancellation
	s.True(s.env.IsWorkflowCompleted())
	var canceledErr *temporal.CanceledError
	s.ErrorAs(s.env.GetWorkflowError(), &canceledErr)
}

func (s *WebhookWorkflowTestSuite) TestMultipleUpsertSignals() {
	// Initial webhook
	webhook := Webhook{
		BaseURL:        "https://original.com",
		ServerAPIToken: "original-token",
	}

	// First API call with original webhook
	s.env.OnActivity((*Activity)(nil).CreateWebhook, mock.Anything, mock.MatchedBy(func(input CreateWebhookInput) bool {
		return input.Webhook.BaseURL == "https://original.com" &&
			input.Webhook.ServerAPIToken == "original-token"
	})).Return(&CreateWebhookOutput{
		WebhookID: "webhook-123",
	}, nil).Once()

	// First trigger creation
	s.env.OnActivity((*Activity)(nil).CreateTrigger, mock.Anything, CreateTriggerInput{
		WebhookID: "webhook-123",
	}).Return(&CreateTriggerOutput{
		TriggerID: "456",
	}, nil).Once()

	// Second API call with updated webhook, but returns same ID (so no new trigger)
	s.env.OnActivity((*Activity)(nil).CreateWebhook, mock.Anything, mock.MatchedBy(func(input CreateWebhookInput) bool {
		return input.Webhook.BaseURL == "https://updated.com" &&
			input.Webhook.ServerAPIToken == "updated-token" &&
			input.Webhook.ID == "webhook-123"
	})).Return(&CreateWebhookOutput{
		WebhookID: "webhook-123",
	}, nil).Once()

	// Don't expect a second trigger creation since webhook ID is the same

	// Send first signal
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(UpsertWebhookSignal, UpsertWebhookInput{
			Webhook: Webhook{
				BaseURL:        "https://original.com",
				ServerAPIToken: "original-token",
			},
		})
	}, time.Millisecond*100)

	// Send second signal with updated webhook data
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(UpsertWebhookSignal, UpsertWebhookInput{
			Webhook: Webhook{
				BaseURL:        "https://updated.com",
				ServerAPIToken: "updated-token",
			},
		})
	}, time.Millisecond*200)

	// Add cancellation to complete the test
	s.env.RegisterDelayedCallback(func() {
		s.env.CancelWorkflow()
	}, time.Millisecond*300)

	// Execute workflow
	s.env.ExecuteWorkflow(WebhookWorkflow, webhook)

	// Verify workflow completed with cancellation
	s.True(s.env.IsWorkflowCompleted())
	var canceledErr *temporal.CanceledError
	s.ErrorAs(s.env.GetWorkflowError(), &canceledErr)
}

func (s *WebhookWorkflowTestSuite) TestContinueAsNewAfterLimit() {
	// Set a low limit for testing continue-as-new
	oldLimit := upsertBeforeCAN
	upsertBeforeCAN = 2
	defer func() { upsertBeforeCAN = oldLimit }()

	webhook := Webhook{
		BaseURL:        "https://example.com",
		ServerAPIToken: "test-api-token",
	}

	// Mock WebhookActivity calls - allow any number of calls
	s.env.OnActivity((*Activity)(nil).CreateWebhook, mock.Anything, mock.Anything).
		Return(&CreateWebhookOutput{WebhookID: "webhook-123"}, nil).Maybe()
	s.env.OnActivity((*Activity)(nil).CreateTrigger, mock.Anything, mock.Anything).
		Return(&CreateTriggerOutput{TriggerID: "456"}, nil).Maybe()

	// Send enough signals to trigger a continue-as-new
	for i := 0; i < upsertBeforeCAN+1; i++ {
		s.env.RegisterDelayedCallback(func() {
			s.env.SignalWorkflow(UpsertWebhookSignal, UpsertWebhookInput{
				Webhook: webhook,
			})
		}, time.Millisecond*time.Duration(100+i*50))
	}

	// Execute workflow
	s.env.ExecuteWorkflow(WebhookWorkflow, webhook)

	// Verify workflow completed with continue-as-new
	var continueErr *workflow.ContinueAsNewError
	s.Error(s.env.GetWorkflowError())
	s.True(errors.As(s.env.GetWorkflowError(), &continueErr))
}
