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
		OpenAIAPIKey string
		OpenAIModel  string
		GeminiAPIKey string
		GeminiModel  string

		TicketSummaryPrompt string
		OrgSummaryPrompt    string
	}

	ServerConfig struct {
		Temporal TemporalClientConfig
		Host     string
		Port     int
		APIKey   string
	}

	WorkerConfig struct {
		Temporal  TemporalClientConfig
		Zendesk   ZendeskConfig
		OpenAI    AIConfig
		QueueName string
		Threads   int
	}
)
