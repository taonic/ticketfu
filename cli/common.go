package cli

import (
	"github.com/urfave/cli/v2"
)

const (
	// Common flag names
	FlagLogLevel = "log-level"
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
