# Ensemble - Multi-Agent Coordination Tool

[![Go Version](https://img.shields.io/badge/Go-1.23-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

Ensemble is a multi-agent developer tool where a coordinating agent dynamically assembles teams of specialized agents from a pool to collaboratively accomplish software development tasks.

## Project Status

**Phase 1 - Week 1: COMPLETE ✓**

Week 1 deliverables (Project Setup) have been fully implemented:
- ✓ Go module initialized
- ✓ Complete directory structure
- ✓ Core protocol types
- ✓ Configuration system
- ✓ Sample configuration files
- ✓ Initial dependencies
- ✓ Placeholder binaries

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

### Run (Coming Soon)

```bash
# Start the server
./bin/ensemble-server

# In another terminal, run a task
./bin/ensemble run "implement feature X"
```

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
# Build all packages
go build ./...

# Run tests (coming soon)
go test ./...

# Tidy dependencies
go mod tidy
```

## Implementation Roadmap

### Phase 1: Core Foundation (Weeks 1-6)
- [x] **Week 1**: Project setup, protocol types, configuration ✓
- [ ] **Week 2**: LLM providers (Anthropic, OpenAI)
- [ ] **Week 3**: Agent system with hot-reload
- [ ] **Week 4**: Orchestration engine
- [ ] **Week 5**: Server implementation
- [ ] **Week 6**: Client CLI and integration

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
