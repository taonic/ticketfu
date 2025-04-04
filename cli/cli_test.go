package cli

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestCliApp(t *testing.T) {
	app := newCliApp()

	assert.Equal(t, "ticketfu", app.Name)
	assert.Equal(t, Version, app.Version)
	assert.NotNil(t, app.Action)

	require.Len(t, app.Commands, 2)

	var workerCmd, serverCmd *cli.Command
	for _, cmd := range app.Commands {
		if cmd.Name == "worker" {
			workerCmd = cmd
		} else if cmd.Name == "server" {
			serverCmd = cmd
		}
	}

	require.NotNil(t, workerCmd, "Worker command missing")
	require.NotNil(t, serverCmd, "Server command missing")

	// Check subcommands
	require.Len(t, workerCmd.Subcommands, 1)
	require.Len(t, serverCmd.Subcommands, 1)
	assert.Equal(t, "start", workerCmd.Subcommands[0].Name)
	assert.Equal(t, "start", serverCmd.Subcommands[0].Name)
}

// TestServerApp tests fx application creation from CLI context
func TestServerApp(t *testing.T) {
	// Create CLI context with minimal required values
	app := cli.NewApp()
	set := flag.NewFlagSet("test", 0)

	set.String(FlagLogLevel, "fatal", "")
	set.String(FlagLogFormat, "json", "")

	set.String(FlagServerHost, "localhost", "")
	set.Int(FlagServerPort, 8080, "")
	set.String(FlagServerAPIToken, "test-token", "")

	set.String(FlagTemporalAddress, "localhost:7233", "")
	set.String(FlagTemporalNamespace, "default", "")
	ctx := cli.NewContext(app, set, nil)

	// Test app creation
	fxApp, err := NewServerApp(ctx)

	require.NoError(t, err)
	assert.NotNil(t, fxApp)
}

// TestWorkerApp tests fx application creation from CLI context
func TestWorkerApp(t *testing.T) {
	// Create CLI context with minimal required values
	app := cli.NewApp()
	set := flag.NewFlagSet("test", 0)
	// Worker config
	set.String(FlagWorkerQueue, "test-queue", "")
	// Zendesk config
	set.String(FlagZendeskSubdomain, "test-subdomain", "")
	set.String(FlagZendeskEmail, "test@example.com", "")
	set.String(FlagZendeskToken, "test-token", "")
	// AI config
	set.String(FlagLLMProvider, "openai", "")
	set.String(FlagLLMModel, "gpt-4", "")
	set.String(FlagLLMAPIKey, "test-key", "")
	// Temporal config
	set.String(FlagTemporalAddress, "localhost:7233", "")
	set.String(FlagTemporalNamespace, "default", "")
	// Log config
	set.String(FlagLogLevel, "fatal", "")
	set.String(FlagLogFormat, "json", "")

	ctx := cli.NewContext(app, set, nil)

	// Test app creation
	fxApp, err := NewWorkerApp(ctx)

	require.NoError(t, err)
	assert.NotNil(t, fxApp)
}
