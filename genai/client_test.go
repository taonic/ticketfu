package genai

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/taonic/ticketfu/config"
	"github.com/tmc/langchaingo/llms"
	"go.temporal.io/server/common/log"
)

// Test specific configs to improve readability
var (
	validGeminiConfig = config.AIConfig{
		LLMProvider: "googleai",
		LLMModel:    "gemini-2.0-flash",
		LLMAPIKey:   "test-key",
	}

	validOpenAIConfig = config.AIConfig{
		LLMProvider: "openai",
		LLMModel:    "gpt-4o-mini",
		LLMAPIKey:   "test-key",
	}

	invalidProviderConfig = config.AIConfig{
		LLMProvider: "invalid",
		LLMModel:    "model",
		LLMAPIKey:   "test-key",
	}

	emptyAPIKeyConfig = config.AIConfig{
		LLMProvider: "openai",
		LLMModel:    "gpt-4o-mini",
		LLMAPIKey:   "",
	}
)

func TestNewAPI(t *testing.T) {
	testCases := []struct {
		name        string
		config      config.AIConfig
		expectError bool
		errorText   string
	}{
		{
			name:        "Valid Gemini Config",
			config:      validGeminiConfig,
			expectError: false,
		},
		{
			name:        "Valid OpenAI Config",
			config:      validOpenAIConfig,
			expectError: false,
		},
		{
			name:        "Invalid Provider",
			config:      invalidProviderConfig,
			expectError: true,
			errorText:   "unknown LLM provider",
		},
		{
			name:        "Empty API Key",
			config:      emptyAPIKeyConfig,
			expectError: true,
			errorText:   "llm-api-key is not provided",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testLogger := log.NewTestLogger()

			// We can't actually test the creation of real clients without valid API keys,
			// so we'll just check for expected errors
			api, err := NewAPI(testLogger, tc.config)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, api)
				if tc.errorText != "" {
					assert.Contains(t, err.Error(), tc.errorText)
				}
			} else if err != nil {
				// If we have a valid API key, we'll get a real client
				// If not, we'll get an error which is normal when testing
				assert.Contains(t, err.Error(), "failed to initialize LLM")
			}
		})
	}
}

// MockLLMModel is a mock of the llms.Model interface
type MockLLMModel struct {
	mock.Mock
}

func (m *MockLLMModel) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	args := m.Called(ctx, messages, options)
	return args.Get(0).(*llms.ContentResponse), args.Error(1)
}
func (m *MockLLMModel) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	args := m.Called(ctx, prompt, options)
	return args.Get(0).(string), args.Error(1)
}

func TestGenerateContent(t *testing.T) {
	testCases := []struct {
		name        string
		instruction string
		content     string
		setupMock   func(*MockLLMModel)
		expected    string
		expectError bool
	}{
		{
			name:        "Successful Generation",
			instruction: "Summarize this",
			content:     "Long text to summarize",
			setupMock: func(m *MockLLMModel) {
				m.On("GenerateContent",
					mock.Anything,
					mock.MatchedBy(func(messages []llms.MessageContent) bool {
						// Verify that messages have the correct structure
						if len(messages) != 2 {
							return false
						}
						// Check system instruction
						if messages[0].Role != llms.ChatMessageTypeSystem ||
							partToString(t, messages[0].Parts[0]) != "Summarize this" {
							return false
						}
						// Check user content
						if messages[1].Role != llms.ChatMessageTypeHuman ||
							partToString(t, messages[1].Parts[0]) != "Long text to summarize" {
							return false
						}
						return true
					}),
					mock.Anything).Return(&llms.ContentResponse{}, nil)
			},
			expectError: false,
		},
		{
			name:        "Model Error",
			instruction: "Summarize this",
			content:     "Text with error",
			setupMock: func(m *MockLLMModel) {
				m.On("GenerateContent",
					mock.Anything,
					mock.Anything,
					mock.Anything).Return(&llms.ContentResponse{}, errors.New("model error"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test logger
			testLogger := log.NewTestLogger()

			// Create a mock LLM model
			mockModel := new(MockLLMModel)
			tc.setupMock(mockModel)

			// Create the genAI instance
			genAIInstance := &genAI{
				logger: testLogger,
				model:  mockModel,
				Config: config.AIConfig{
					LLMProvider: "test-provider",
					LLMModel:    "test-model",
					LLMAPIKey:   "test-key",
				},
			}

			// Call the method being tested
			_, err := genAIInstance.GenerateContent(context.Background(), tc.instruction, tc.content)

			// Check the results
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify that the mock was called as expected
			mockModel.AssertExpectations(t)
		})
	}
}

func partToString(t *testing.T, cp llms.ContentPart) string {
	var part strings.Builder
	if tc, ok := cp.(llms.TextContent); ok {
		part.WriteString(tc.Text)
	} else {
		t.Error("Cannot convert ContentPart to string")
	}
	return part.String()
}
