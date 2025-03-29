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
	"google.golang.org/genai"
)

// MockGeminiAPI mocks the gemini.API
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
	// Define test cases
	testCases := []struct {
		name           string
		ticket         Ticket
		setupMock      func(*MockGeminiAPI)
		expectedOutput string
		expectedError  string
	}{
		{
			name:   "Successful Summary Generation",
			ticket: createTestTicket(),
			setupMock: func(m *MockGeminiAPI) {
				m.On("GetConfig").Return(config.AIConfig{
					GeminiModel:         "gemini-2.0-flash",
					TicketSummaryPrompt: "test",
				})

				response := &genai.GenerateContentResponse{
					Candidates: []*genai.Candidate{{
						Content: &genai.Content{
							Parts: []*genai.Part{
								&genai.Part{Text: "This is a generated summary"},
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
			expectedOutput: "This is a generated summary",
		},
		{
			name:   "Generation API Error",
			ticket: createTestTicket(),
			setupMock: func(m *MockGeminiAPI) {
				m.On("GetConfig").Return(config.AIConfig{
					GeminiModel:         "gemini-2.0-flash",
					TicketSummaryPrompt: "test",
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
			name:   "Empty Response",
			ticket: createTestTicket(),
			setupMock: func(m *MockGeminiAPI) {
				m.On("GetConfig").Return(config.AIConfig{
					GeminiModel:         "gemini-2.0-flash",
					TicketSummaryPrompt: "test",
				})

				emptyResponse := &genai.GenerateContentResponse{
					Candidates: []*genai.Candidate{{
						Content: &genai.Content{
							Parts: []*genai.Part{&genai.Part{Text: ""}},
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
			input := GenSummaryInput{Ticket: tc.ticket}
			output, err := activity.GenTicketSummary(ctx, input)

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
