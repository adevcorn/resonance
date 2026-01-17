# Ensemble - Multi-Agent Coordination Tool

[![Go Version](https://img.shields.io/badge/Go-1.23-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

Ensemble is a multi-agent developer tool where a coordinating agent dynamically assembles teams of specialized agents from a pool to collaboratively accomplish software development tasks.

## Project Status

**Phase 1: COMPLETE ✓**

All Phase 1 deliverables have been fully implemented:

**Week 1 - Project Setup**: ✓ COMPLETE
- Go module initialized
- Complete directory structure
- Core protocol types
- Configuration system

**Week 2 - LLM Providers**: ✓ COMPLETE
- Provider interface
- Anthropic provider with streaming
- OpenAI provider with streaming
- Provider registry

**Week 3 - Agent System**: ✓ COMPLETE
- Agent YAML loader
- Agent pool management
- Hot-reload watcher with fsnotify
- 9 default agent definitions

**Week 4 - Orchestration**: ✓ COMPLETE
- Tool registry
- Collaborate tool
- Assemble team tool
- Coordinator, Moderator, Synthesizer
- Orchestration engine

**Week 5 - Server**: ✓ COMPLETE
- JSON storage
- Session manager
- HTTP API endpoints
- WebSocket handler
- Server binary

**Week 6 - Client**: ✓ COMPLETE
- Server connection (HTTP client)
- WebSocket connection
- Project detection
- Permission system with sandboxing
- Local tool executor
- File tools (read, write, list)
- Exec tools
- CLI with all commands
- Integration tests

## Architecture

Ensemble uses a **client-server architecture**:

- **Server**: Central backend managing agent orchestration, LLM provider communication, and session management
- **Client**: CLI tool that connects to the server and executes local tools (file operations, command execution) within the user's project

### Key Features (Planned)

- **Dynamic Team Assembly**: Coordinator AI analyzes tasks and selects the right specialists
- **Free-Form Collaboration**: Agents discuss naturally with shared context, coordinator moderates
- **Hybrid Tool Execution**: File/exec operations on client, search/fetch on server
- **Hot-Reloading Agents**: YAML agent definitions reload automatically on change
- **Multi-Provider Support**: OpenAI, Anthropic, Google AI, Ollama

## Project Structure

```
ensemble/
├── cmd/
│   ├── ensemble/           # CLI client binary
│   └── ensemble-server/    # Server binary
├── internal/
│   ├── server/            # Server-side code
│   │   ├── api/          # HTTP/WebSocket handlers
│   │   ├── orchestration/ # Core orchestration logic
│   │   ├── agent/        # Agent management
│   │   ├── provider/     # LLM providers
│   │   ├── tool/         # Server-side tools
│   │   └── storage/      # Persistence layer
│   ├── client/           # Client-side code
│   │   └── tools/        # Client-executed tools
│   ├── protocol/         # Shared types (client/server)
│   └── config/           # Configuration loaders
├── pkg/ensemble/         # Public SDK (future)
├── agents/              # Default agent definitions (YAML)
├── config/              # Configuration files
│   ├── server.yaml
│   └── client.yaml
├── go.mod
├── go.sum
└── PLAN.md             # Complete implementation plan
```

## Quick Start

### Prerequisites

- Go 1.23 or later
- API keys for LLM providers (Anthropic, OpenAI, etc.)

### Build

```bash
# Build both binaries
go build -o bin/ensemble ./cmd/ensemble
go build -o bin/ensemble-server ./cmd/ensemble-server

# Or use make (coming soon)
```

### Configuration

1. Copy sample configurations:
   ```bash
   cp config/server.yaml config/server.local.yaml
   cp config/client.yaml config/client.local.yaml
   ```

2. Set environment variables:
   ```bash
   export ANTHROPIC_API_KEY="your-key"
   export OPENAI_API_KEY="your-key"
   export GOOGLE_AI_API_KEY="your-key"
   ```

3. Update configurations as needed

### Usage

```bash
# Start the server
./bin/ensemble-server

# In another terminal, use the CLI client

# View available agents
./bin/ensemble agents list
./bin/ensemble agents show developer

# Run a task
./bin/ensemble run "implement user authentication"

# Manage sessions
./bin/ensemble sessions list
./bin/ensemble sessions show <session-id>
./bin/ensemble sessions delete <session-id>

# View configuration
./bin/ensemble config

# Show version
./bin/ensemble version
```

### CLI Commands

- `ensemble run [task]` - Run a task with multi-agent collaboration
- `ensemble agents list` - List all available agents
- `ensemble agents show [name]` - Show detailed agent information
- `ensemble sessions list` - List all sessions
- `ensemble sessions show [id]` - Show session details
- `ensemble sessions delete [id]` - Delete a session
- `ensemble config` - Display current configuration
- `ensemble version` - Display version information

## Development

### Dependencies

Core dependencies:
- `github.com/spf13/viper` - Configuration management
- `github.com/spf13/cobra` - CLI framework
- `gopkg.in/yaml.v3` - YAML parsing
- `github.com/rs/zerolog` - Structured logging
- `github.com/google/uuid` - UUID generation
- `github.com/fsnotify/fsnotify` - File watching for hot-reload

### Build & Test

```bash
# Build both binaries
make build

# Or build individually
go build -o bin/ensemble ./cmd/ensemble
go build -o bin/ensemble-server ./cmd/ensemble-server

# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test packages
go test ./internal/client/...
go test ./internal/server/...

# Tidy dependencies
go mod tidy
```

## Implementation Roadmap

### Phase 1: Core Foundation (Weeks 1-6) ✓ COMPLETE
- [x] **Week 1**: Project setup, protocol types, configuration
- [x] **Week 2**: LLM providers (Anthropic, OpenAI)
- [x] **Week 3**: Agent system with hot-reload
- [x] **Week 4**: Orchestration engine
- [x] **Week 5**: Server implementation
- [x] **Week 6**: Client CLI and integration

**🎉 Phase 1 MVP is now complete and functional!**

### Phase 2: Enhanced Features (Weeks 7-10)
- Google AI + Ollama providers
- Server tools (web_search, fetch_url)
- MCP client integration
- Session resume capability

### Phase 3: Production Hardening (Weeks 11-14)
- Authentication & authorization
- Rate limiting
- Observability (metrics, tracing)
- Performance optimization

### Phase 4: Frontends (Weeks 15-20)
- TUI with bubbletea
- Web frontend
- VS Code extension (stretch)

## Documentation

- [PLAN.md](PLAN.md) - Complete implementation plan and technical specifications
- [AGENTS.md](AGENTS.md) - Agent development guide
- Configuration schemas in PLAN.md

## Contributing

This project is in active development. Contributions will be welcome once Phase 1 is complete.

## License

MIT License (to be added)

## Acknowledgments

Built with Go and powered by modern LLM providers.
