package cli

import (
	"github.com/urfave/cli/v2"
)

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

// GetServerFlags returns the flags for server commands
func GetServerFlags() []cli.Flag {
	return serverFlags
}

// GetWorkerFlags returns the flags for worker commands
func GetWorkerFlags() []cli.Flag {
	return workerFlags
}

// GetFlag gets a flag value from the CLI context with proper type conversion
func GetFlag[T any](cliCtx *cli.Context, flagName string) (T, bool) {
	var zero T

	if !cliCtx.IsSet(flagName) {
		return zero, false
	}

	value := cliCtx.Value(flagName)
	if typedValue, ok := value.(T); ok {
		return typedValue, true
	}

	return zero, false
}
