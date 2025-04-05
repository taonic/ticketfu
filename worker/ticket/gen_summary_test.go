package ticket

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/taonic/ticketfu/config"
	"go.temporal.io/sdk/testsuite"
)

// MockGenAIAPI mocks the gemini.API
type MockGenAIAPI struct {
	mock.Mock
}

func (m *MockGenAIAPI) GenerateContent(ctx context.Context, instruction, content string) (string, error) {
	args := m.Called(ctx, instruction, content)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockGenAIAPI) GetConfig() config.AIConfig {
	args := m.Called()
	return args.Get(0).(config.AIConfig)
}

// Helper method to create a test ticket
func createTestTicket() Ticket {
	now := time.Now()
	return Ticket{
		ID:               12345,
		Subject:          "Test Issue",
		Description:      "This is a test description",
		Priority:         "high",
		Status:           "open",
		Requester:        "John Doe",
		Assignee:         "Support Agent",
		OrganizationName: "Test Org",
		CreatedAt:        &now,
		UpdatedAt:        &now,
		Comments:         []string{"Comment 1", "Comment 2"},
	}
}

func TestActivity_GenSummary(t *testing.T) {
	testSuite := testsuite.WorkflowTestSuite{}
	testEnv := testSuite.NewTestActivityEnvironment()

	// Define test cases
	testCases := []struct {
		name           string
		ticket         Ticket
		setupMock      func(*MockGenAIAPI)
		expectedOutput string
		expectedError  string
	}{
		{
			name:   "Successful Summary Generation",
			ticket: createTestTicket(),
			setupMock: func(m *MockGenAIAPI) {
				m.On("GetConfig").Return(config.AIConfig{
					LLMModel:            "gemini-2.0-flash",
					TicketSummaryPrompt: "test",
				})

				response := "This is a generated summary"

				m.On("GenerateContent",
					mock.Anything,
					mock.Anything,
					mock.Anything).Return(response, nil)
			},
			expectedOutput: "This is a generated summary",
		},
		{
			name:   "Generation API Error",
			ticket: createTestTicket(),
			setupMock: func(m *MockGenAIAPI) {
				m.On("GetConfig").Return(config.AIConfig{
					LLMModel:            "gemini-2.0-flash",
					TicketSummaryPrompt: "test",
				})

				m.On("GenerateContent",
					mock.Anything,
					mock.Anything,
					mock.Anything).Return("", errors.New("API failure"))
			},
			expectedError: "failed to generate",
		},
		{
			name:   "Empty Response",
			ticket: createTestTicket(),
			setupMock: func(m *MockGenAIAPI) {
				m.On("GetConfig").Return(config.AIConfig{
					LLMModel:            "gemini-2.0-flash",
					TicketSummaryPrompt: "test",
				})

				emptyResponse := ""

				m.On("GenerateContent",
					mock.Anything,
					mock.Anything,
					mock.Anything).Return(emptyResponse, nil)
			},
			expectedOutput: "",
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			mockAPI := new(MockGenAIAPI)
			tc.setupMock(mockAPI)

			activity := &Activity{genAPI: mockAPI}
			testEnv.RegisterActivity(activity.GenTicketSummary)

			// Execute
			input := GenSummaryInput{Ticket: tc.ticket}
			future, err := testEnv.ExecuteActivity(activity.GenTicketSummary, input)

			// Verify
			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				require.NoError(t, err)
				var output GenSummaryOutput
				err := future.Get(&output)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedOutput, output.Summary)
			}

			mockAPI.AssertExpectations(t)
		})
	}
}
