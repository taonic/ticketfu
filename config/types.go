package config

type (
	TemporalClientConfig struct {
		Address     string // Temporal service address
		Namespace   string // Temporal namespace
		APIKey      string // Temporal API key (if using secured service)
		TLSCertPath string // Path to TLS cert file
		TLSKeyPath  string // Path to TLS key file
	}

	ZendeskConfig struct {
		ZendeskSubdomain string
		ZendeskEmail     string
		ZendeskToken     string
	}

	AIConfig struct {
		LLMProvider string
		LLMModel    string
		LLMAPIKey   string

		TicketSummaryPrompt string
		OrgSummaryPrompt    string
	}

	ServerConfig struct {
		Temporal              TemporalClientConfig
		Host                  string
		Port                  int
		ZendeskWebhookBaseURL string
		APIToken              string
	}

	WorkerConfig struct {
		QueueName string
	}
)
