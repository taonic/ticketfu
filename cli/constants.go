package cli

// Flag name constants
const (
	// Common flag names
	FlagLogLevel = "log-level"

	// Config flags
	FlagConfig = "config"

	// Server-specific flags
	FlagServerHost = "host"
	FlagServerPort = "port"

	// Worker-specific flags
	FlagWorkerQueue   = "queue"
	FlagWorkerThreads = "threads"
)
