package cli

import (
	"github.com/urfave/cli/v2"
)

const (
	Version = "0.0.1"
)

// Run parses CLI args and starts the appropriate fx application
func Run(args []string) error {
	app := newCliApp()
	return app.Run(args)
}

// newCliApp creates a new CLI app with all commands registered
func newCliApp() *cli.App {
	app := cli.NewApp()
	app.Name = "ticketiq"
	app.Version = Version
	app.Commands = []*cli.Command{
		NewWorkerCommand(),
		NewServerCommand(),
	}

	// Default action if no command is provided
	app.Action = func(c *cli.Context) error {
		return cli.ShowAppHelp(c)
	}

	return app
}
