package org

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/taonic/ticketfu/config"
	"go.temporal.io/sdk/testsuite"
)

// MockGeminiAPI mocks the gemini.API interface
type MockGeminiAPI struct {
	mock.Mock
}

func (m *MockGeminiAPI) GenerateContent(ctx context.Context, instruction, content string) (string, error) {
	args := m.Called(ctx, instruction, content)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockGeminiAPI) GetConfig() config.AIConfig {
	args := m.Called()
	return args.Get(0).(config.AIConfig)
}

// Helper method to create a test organization
func createTestOrganization() Organization {
	return Organization{
		ID:      123,
		Name:    "Test Organization",
		Notes:   "Important client",
		Details: "Enterprise account",
		TicketSummaries: map[int64]string{
			1001: "Ticket about feature request",
			1002: "Ticket about billing issue",
			1003: "Ticket about technical support",
		},
	}
}

func TestActivity_GenOrgSummary(t *testing.T) {
	testSuite := testsuite.WorkflowTestSuite{}
	testEnv := testSuite.NewTestActivityEnvironment()

	// Define test cases
	testCases := []struct {
		name           string
		organization   Organization
		setupMock      func(*MockGeminiAPI)
		expectedOutput string
		expectedError  string
	}{
		{
			name:         "Successful Summary Generation",
			organization: createTestOrganization(),
			setupMock: func(m *MockGeminiAPI) {
				m.On("GetConfig").Return(config.AIConfig{
					LLMModel:         "gemini-2.0-flash",
					OrgSummaryPrompt: "Analyze organization tickets",
				})

				response := `{"overview": "Test Organization has multiple support issues"}`

				m.On("GenerateContent",
					mock.Anything,
					mock.Anything,
					mock.Anything).Return(response, nil)
			},
			expectedOutput: `{"overview": "Test Organization has multiple support issues"}`,
		},
		{
			name:         "Generation API Error",
			organization: createTestOrganization(),
			setupMock: func(m *MockGeminiAPI) {
				m.On("GetConfig").Return(config.AIConfig{
					LLMModel:         "gemini-2.0-flash",
					OrgSummaryPrompt: "Analyze organization tickets",
				})

				m.On("GenerateContent",
					mock.Anything,
					mock.Anything,
					mock.Anything).Return("", errors.New("API failure"))
			},
			expectedError: "failed to generate",
		},
		{
			name:         "Empty Response",
			organization: createTestOrganization(),
			setupMock: func(m *MockGeminiAPI) {
				m.On("GetConfig").Return(config.AIConfig{
					LLMModel:         "gemini-2.0-flash",
					OrgSummaryPrompt: "Analyze organization tickets",
				})

				emptyResponse := ""

				m.On("GenerateContent",
					mock.Anything,
					mock.Anything,
					mock.Anything).Return(emptyResponse, nil)
			},
			expectedOutput: "",
		},
		{
			name:         "Organization with No Ticket Summaries",
			organization: Organization{ID: 123, Name: "Test Organization"},
			setupMock: func(m *MockGeminiAPI) {
				m.On("GetConfig").Return(config.AIConfig{
					LLMModel:         "gemini-2.0-flash",
					OrgSummaryPrompt: "Analyze organization tickets",
				})

				response := `{"overview": "No ticket data available", "main_topics": []}`

				m.On("GenerateContent",
					mock.Anything,
					mock.Anything,
					mock.Anything).Return(response, nil)
			},
			expectedOutput: `{"overview": "No ticket data available", "main_topics": []}`,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			mockAPI := new(MockGeminiAPI)
			tc.setupMock(mockAPI)

			activity := &Activity{genAPI: mockAPI}

			// Register the activity with the test environment
			testEnv.RegisterActivity(activity.GenOrgSummary)

			// Execute
			input := GenSummaryInput{Organization: tc.organization}
			future, err := testEnv.ExecuteActivity(activity.GenOrgSummary, input)

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
