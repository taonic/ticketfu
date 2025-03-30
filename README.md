# TicketFu

[![Go Tests](https://github.com/taonic/ticketfu/workflows/Go%20Tests/badge.svg)](https://github.com/taonic/ticketfu/actions)
[![codecov](https://codecov.io/gh/taonic/ticketfu/branch/main/graph/badge.svg)](https://codecov.io/gh/taonic/ticketfu)

TicketFu is a system that enhances support ticket management with AI-powered insights. It integrates with Zendesk and uses LLMs (Gemini or OpenAI) to provide ticket summaries and organization-level analytics.

Deploy TicketFu with one click on Render:

[![Deploy to Render](https://render.com/images/deploy-to-render-button.svg)](https://render.com/deploy?repo=https://github.com/taonic/ticketfu)

## Features

- Automatically generates ticket summaries using AI
- Provides organization-level insights across all tickets
- RESTful API for ticket and organization management
- Uses Temporal workflows for reliable background processing
- Worker-based architecture for scalable processing
- Zendesk integration to pull ticket data
- Secure API authentication

## Architecture

TicketFu consists of two main components:

1. **Server**: Handles HTTP requests, exposes API endpoints, and triggers workflows
2. **Worker**: Processes tickets and organizations in the background using Temporal

## Installation

```bash
go install github.com/taonic/ticketfu/cmd/ticketfu@latest
```

## Usage

### Starting the server

```bash
ticketfu server start --server-api-token YOUR_API_TOKEN --temporal-address localhost:7233
```

### Starting a worker

```bash
ticketfu worker start \
  --zendesk-subdomain your-zendesk \
  --zendesk-email user@example.com \
  --zendesk-token YOUR_ZENDESK_TOKEN \
  --gemini-api-key YOUR_GEMINI_API_KEY \
  --temporal-address localhost:7233
```

## Configuration

### API Keys

#### Generating a Gemini API Key

1. Go to the [Google AI Studio](https://makersuite.google.com/app/apikey)
2. Log in with your Google account
3. Click on "Create API Key" button
4. Copy your new API key to use with TicketFu

Note: OpenAI support is currently in development and will be available in an upcoming release.

### Server Configuration

- `--host` or `HOST`: Server host address (default: "0.0.0.0")
- `--port` or `PORT`: Server port (default: 8080)
- `--server-api-token` or `SERVER_API_TOKEN`: API token for authenticating requests (required)

### Worker Configuration

- `--queue` or `WORKER_QUEUE`: Worker queue name (default: "default")
- `--threads` or `WORKER_THREADS`: Number of worker threads (default: 4)

### Zendesk Configuration

- `--zendesk-subdomain` or `ZENDESK_SUBDOMAIN`: Zendesk subdomain (required)
- `--zendesk-email` or `ZENDESK_EMAIL`: Zendesk email (required)
- `--zendesk-token` or `ZENDESK_TOKEN`: Zendesk API token (required)

### AI Configuration

- `--gemini-api-key` or `GEMINI_API_KEY`: Google Gemini API key (required)
- `--gemini-model` or `GEMINI_MODEL`: Gemini model (default: "gemini-2.0-flash")

### Temporal Configuration

- `--temporal-address` or `TEMPORAL_ADDRESS`: Temporal service address (default: "127.0.0.1:7233")
- `--temporal-namespace` or `TEMPORAL_NAMESPACE`: Temporal namespace (default: "default")
- `--temporal-api-key` or `TEMPORAL_API_KEY`: Temporal API key (optional)

## API Endpoints

- `GET /health`: Health check endpoint
- `POST /api/v1/ticket`: Process a new ticket or update an existing one
- `GET /api/v1/ticket/summary`: Get a specific ticket summary
- `GET /api/v1/organization/summary`: Get organization-level summary

## Development

### Running tests

```bash
go test ./...
```



## Using Temporal

TicketFu uses Temporal for workflow orchestration, implementing the "Entity Workflow" pattern. In this pattern:

- Each entity (ticket or organization) gets its own long-running workflow instance
- The workflow maintains the entity's state and handles all operations for that entity
- External events trigger operations via signals
- Queries allow reading the current state without interrupting workflow execution

Key workflow implementations:

- `ticket.TicketWorkflow`: Processes individual tickets, generates summaries, and maintains ticket state
  - Uses `SignalWithStartWorkflow` to ensure a single workflow instance per ticket
  - Signals organization workflows when ticket summaries change
  
- `org.OrganizationWorkflow`: Aggregates ticket data and generates organization-level insights
  - Maintains a map of ticket summaries for each organization
  - Generates AI-powered insights across all tickets
  - Uses `continue-as-new` for handling long histories

This pattern provides strong consistency, exactly-once semantics, and automatic retry handling for each entity.

## License

[MIT](LICENSE)
