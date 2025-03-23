package config

type (
	ClientOptions struct {
		Address                    string
		Namespace                  string
		ApiKey                     string
		GrpcMeta                   []string
		Tls                        bool
		TlsCertPath                string
		TlsCertData                string
		TlsKeyPath                 string
		TlsKeyData                 string
		TlsCaPath                  string
		TlsCaData                  string
		TlsDisableHostVerification bool
		TlsServerName              string
		CodecEndpoint              string
		CodecAuth                  string
		CodecHeader                []string
	}

	ServerConfig struct {
		ClientOptions
		Host   string
		Port   int
		APIKey string // API key for authenticating requests
	}

	WorkerConfig struct {
		ClientOptions
		QueueName string
		Threads   int
	}
)
