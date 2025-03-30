package cli

import (
	"github.com/urfave/cli/v2"
)

const (
	// Common flag names
	FlagLogLevel  = "log-level"
	FlagLogFormat = "log-format"

	// Temporal-specific flags
	FlagTemporalAddress   = "temporal-address"
	FlagTemporalNamespace = "temporal-namespace"
	FlagTemporalAPIKey    = "temporal-api-key"
	FlagTemporalTLSCert   = "temporal-tls-cert"
	FlagTemporalTLSKey    = "temporal-tls-key"

	// Zendesk-specific flags
	FlagZendeskSubdomain = "zendesk-subdomain"
	FlagZendeskEmail     = "zendesk-email"
	FlagZendeskToken     = "zendesk-token"

	// AI-specific flags
	FlagOpenAIAPIKey        = "openai-api-key"
	FlagOpenAIModel         = "openai-model"
	FlagGeminiAPIKey        = "gemini-api-key"
	FlagGeminiModel         = "gemini-model"
	FlagTicketSummaryPrompt = "ticket-summary-prompt"
	FlagOrgSummaryPrompt    = "org-summary-prompt"
)

// Temporal flags shared across commands
var temporalFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    FlagTemporalAddress,
		EnvVars: []string{"TEMPORAL_ADDRESS"},
		Usage:   "Temporal service address",
		Value:   "127.0.0.1:7233",
	},
	&cli.StringFlag{
		Name:    FlagTemporalNamespace,
		EnvVars: []string{"TEMPORAL_NAMESPACE"},
		Usage:   "Temporal namespace",
		Value:   "default",
	},
	&cli.StringFlag{
		Name:    FlagTemporalAPIKey,
		EnvVars: []string{"TEMPORAL_API_KEY"},
		Usage:   "Temporal API key for authentication",
	},
	&cli.StringFlag{
		Name:    FlagTemporalTLSCert,
		EnvVars: []string{"TEMPORAL_TLS_CERT"},
		Usage:   "Path to Temporal TLS certificate file",
	},
	&cli.StringFlag{
		Name:    FlagTemporalTLSKey,
		EnvVars: []string{"TEMPORAL_TLS_KEY"},
		Usage:   "Path to Temporal TLS key file",
	},
}

// Zendesk flags shared across commands
var zendeskFlags = []cli.Flag{
	&cli.StringFlag{
		Name:     FlagZendeskSubdomain,
		EnvVars:  []string{"ZENDESK_SUBDOMAIN"},
		Usage:    "Zendesk subdomain",
		Required: true,
	},
	&cli.StringFlag{
		Name:     FlagZendeskEmail,
		EnvVars:  []string{"ZENDESK_EMAIL"},
		Usage:    "Zendesk email",
		Required: true,
	},
	&cli.StringFlag{
		Name:     FlagZendeskToken,
		EnvVars:  []string{"ZENDESK_TOKEN"},
		Usage:    "Zendesk API token",
		Required: true,
	},
}

// AI flags shared across commands
var aiFlags = []cli.Flag{
	&cli.StringFlag{
		Name:     FlagOpenAIAPIKey,
		EnvVars:  []string{"OPENAI_API_KEY"},
		Usage:    "OpenAI API Key",
		Required: true,
	},
	&cli.StringFlag{
		Name:     FlagOpenAIModel,
		EnvVars:  []string{"OPENAI_MODEL"},
		Usage:    "OpenAI's LLM Model",
		Required: false,
		Value:    "gpt-4o-mini",
	},
	&cli.StringFlag{
		Name:     FlagGeminiAPIKey,
		EnvVars:  []string{"GEMINI_API_KEY"},
		Usage:    "Gemini API Key",
		Required: false,
	},
	&cli.StringFlag{
		Name:     FlagGeminiModel,
		EnvVars:  []string{"GEMINI_MODEL"},
		Usage:    "Gemini's LLM Model",
		Required: false,
		Value:    "gemini-2.0-flash",
	},
	&cli.StringFlag{
		Name:     FlagTicketSummaryPrompt,
		EnvVars:  []string{"TICKET_SUMMARY_PROMPT"},
		Usage:    "Prompt used for generating ticket summary",
		Required: false,
		Value: `
		* you are a support engineer \n
		* include participant names in the below summary
		* brief the intent of the ticket \n
		* summarize the ticket \n
		* brief the next step \n
		* return the result as json object with fields: intent, summary and next_step
		`,
	},
	&cli.StringFlag{
		Name:     FlagOrgSummaryPrompt,
		EnvVars:  []string{"ORG_SUMMARY_PROMPT"},
		Usage:    "Prompt used for generating organization summary",
		Required: false,
		Value: `
    You are an expert support analyst synthesizing multiple support tickets for an organization.

    Analyze the provided ticket summaries and create a comprehensive organization-level summary.
    Return the analysis as a valid JSON object with the following structure:
    {
        "overview": "Brief overview of the organization's support landscape",
        "main_topics": ["List of main topics identified"],
        "key_people": ["List of key people and their role from the customer"],
        "key_insights": "Key insights about organizational challenges and needs",
        "trending_topics": [
            {
                "topic": "Specific topic",
                "frequency": number of occurrences,
                "importance": "high/medium/low"
            }
        ],
        "recommended_actions": ["List of recommended actions"]
    }

    Guidelines:
    1. Overview: Provide a concise summary of the organization's support patterns. Name the organization.
    2. Main Topics: List key themes found across tickets. Name the organization.
    3. Key Insights: Extract meaningful patterns about challenges and needs
    4. Trending Topics: Identify recurring issues with their frequency and importance
    5. Recommended Actions: Suggest concrete steps based on the analysis

    Ensure the response is a valid JSON object that can be parsed programmatically.
    Focus on identifying patterns and insights that would be valuable for understanding the organization's overall support needs.
    Keep the analysis professional and actionable.
		`,
	},
}

// Common flags that apply to multiple commands
var commonFlags = []cli.Flag{
	&cli.StringFlag{
		Name:     FlagLogLevel,
		Aliases:  []string{"l"},
		EnvVars:  []string{"LOG_LEVEL"},
		Usage:    "Set log level(debug, info, warn, error). Default level is info",
		Required: false,
		Value:    "info",
	},
	&cli.StringFlag{
		Name:     FlagLogFormat,
		Aliases:  []string{"f"},
		EnvVars:  []string{"LOG_FORMAT"},
		Usage:    "Set log format(json, console). Default format is json",
		Required: false,
		Value:    "json",
	},
}
