# TicketFu - WIP

[![Go Tests](https://github.com/taonic/ticketfu/workflows/Go%20Tests/badge.svg)](https://github.com/taonic/ticketfu/actions)
[![codecov](https://codecov.io/gh/taonic/ticketfu/branch/main/graph/badge.svg)](https://codecov.io/gh/taonic/ticketfu)

Deploy TicketFu with one click on Render.

[![Deploy to Render](https://render.com/images/deploy-to-render-button.svg)](https://render.com/deploy?repo=https://github.com/taonic/ticketfu)

## Features

- RESTful API for ticket management
- Worker-based architecture for background processing
- CLI for easy management and operation

## Installation

```bash
go install github.com/taonic/ticketfu/cmd/ticketfu@latest
```

## Usage

### Starting the server

```bash
ticketfu server start --api-key YOUR_API_KEY
```

### Starting a worker

```bash
ticketfu worker start --queue default --threads 4
```

## Configuration

The application can be configured using command line flags or environment variables:

### Server Configuration

- `--host` or `SERVER_HOST`: Server host address (default: "localhost")
- `--port` or `SERVER_PORT`: Server port (default: 8080)
- `--api-key` or `API_KEY`: API key for authenticating requests (required)

### Worker Configuration

- `--queue` or `WORKER_QUEUE`: Worker queue name (default: "default")
- `--threads` or `WORKER_THREADS`: Number of worker threads (default: 4)

### Common Configuration

- `--log-level` or `LOG_LEVEL`: Set log level (debug, info, warn, error)

## Development

### Running tests

```bash
go test ./...
```

### Running linters

```bash
golangci-lint run
```

## License

[MIT](LICENSE)
