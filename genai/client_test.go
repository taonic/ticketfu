package genai

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/taonic/ticketfu/config"
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
