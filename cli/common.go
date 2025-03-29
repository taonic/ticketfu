package cli

import (
	"github.com/urfave/cli/v2"
)

const (
	// Common flag names
	FlagLogLevel = "log-level"

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
	FlagOpenAIAPIKey = "openai-api-key"
	FlagGeminiAPIKey = "gemini-api-key"
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
		Name:     FlagGeminiAPIKey,
		EnvVars:  []string{"GEMINI_API_KEY"},
		Usage:    "Gemini API Key",
		Required: true,
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
	},
}
