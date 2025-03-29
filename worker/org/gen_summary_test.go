package org

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/taonic/ticketfu/config"
	"google.golang.org/genai"
)

// MockGeminiAPI mocks the gemini.API interface
type MockGeminiAPI struct {
	mock.Mock
}

func (m *MockGeminiAPI) GenerateContent(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
	args := m.Called(ctx, model, contents, config)
	return args.Get(0).(*genai.GenerateContentResponse), args.Error(1)
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
					GeminiModel:      "gemini-2.0-flash",
					OrgSummaryPrompt: "Analyze organization tickets",
				})

				response := &genai.GenerateContentResponse{
					Candidates: []*genai.Candidate{{
						Content: &genai.Content{
							Parts: []*genai.Part{
								&genai.Part{Text: `{"overview": "Test Organization has multiple support issues", "main_topics": ["feature request", "billing issues", "technical support"]}`},
							},
						},
					}},
				}

				m.On("GenerateContent",
					mock.Anything,
					"gemini-2.0-flash",
					mock.Anything,
					mock.Anything).Return(response, nil)
			},
			expectedOutput: `{"overview": "Test Organization has multiple support issues", "main_topics": ["feature request", "billing issues", "technical support"]}`,
		},
		{
			name:         "Generation API Error",
			organization: createTestOrganization(),
			setupMock: func(m *MockGeminiAPI) {
				m.On("GetConfig").Return(config.AIConfig{
					GeminiModel:      "gemini-2.0-flash",
					OrgSummaryPrompt: "Analyze organization tickets",
				})

				m.On("GenerateContent",
					mock.Anything,
					"gemini-2.0-flash",
					mock.Anything,
					mock.Anything).Return(&genai.GenerateContentResponse{}, errors.New("API failure"))
			},
			expectedError: "failed to generate",
		},
		{
			name:         "Empty Response",
			organization: createTestOrganization(),
			setupMock: func(m *MockGeminiAPI) {
				m.On("GetConfig").Return(config.AIConfig{
					GeminiModel:      "gemini-2.0-flash",
					OrgSummaryPrompt: "Analyze organization tickets",
				})

				emptyResponse := &genai.GenerateContentResponse{
					Candidates: []*genai.Candidate{{
						Content: &genai.Content{
							Parts: []*genai.Part{
								&genai.Part{Text: ""},
							},
						},
					}},
				}

				m.On("GenerateContent",
					mock.Anything,
					"gemini-2.0-flash",
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
					GeminiModel:      "gemini-2.0-flash",
					OrgSummaryPrompt: "Analyze organization tickets",
				})

				response := &genai.GenerateContentResponse{
					Candidates: []*genai.Candidate{{
						Content: &genai.Content{
							Parts: []*genai.Part{
								&genai.Part{Text: `{"overview": "No ticket data available", "main_topics": []}`},
							},
						},
					}},
				}

				m.On("GenerateContent",
					mock.Anything,
					"gemini-2.0-flash",
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
			ctx := context.Background()
			mockAPI := new(MockGeminiAPI)
			tc.setupMock(mockAPI)

			activity := &Activity{
				genAPI: mockAPI,
			}

			// Execute
			input := GenSummaryInput{Organization: tc.organization}
			output, err := activity.GenOrgSummary(ctx, input)

			// Verify
			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
				assert.Nil(t, output)
			} else {
				require.NoError(t, err)
				require.NotNil(t, output)
				assert.Equal(t, tc.expectedOutput, output.Summary)
			}

			mockAPI.AssertExpectations(t)
		})
	}
}
