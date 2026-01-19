# Ensemble - Multi-Agent Coordination Tool

[![Go Version](https://img.shields.io/badge/Go-1.23-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

Ensemble is a multi-agent developer tool where a coordinating agent dynamically assembles teams of specialized agents from a pool to collaboratively accomplish software development tasks.

## Project Status

**Phase 1: COMPLETE ✓** (January 17, 2026)

All Phase 1 deliverables have been fully implemented and tested:

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
- WebSocket handler with streaming
- Server binary

**Week 6 - Client**: ✓ COMPLETE
- Server connection (HTTP client)
- WebSocket connection with bidirectional tool execution
- Project detection
- Permission system with sandboxing
- Local tool executor
- File tools (read, write, list)
- Exec tools
- CLI with all commands
- Integration tests

**Post-Phase 1 - Tool Execution**: ✓ COMPLETE
- WebSocket connection fix (URL mismatch resolved)
- Per-session orchestration engines
- Bidirectional tool execution flow
- Real-time agent message streaming
- Tool result propagation with timeout handling
- End-to-end working system

**Post-Phase 1 - Additional Providers**: ✓ COMPLETE
- Z.ai provider integration (GLM models with OpenAI-compatible API)
- Google Gemini provider integration (Gemini 2.0 models with tool calling)
- Full tool calling and streaming support

## Architecture

Ensemble uses a **client-server architecture**:

- **Server**: Central backend managing agent orchestration, LLM provider communication, and session management
- **Client**: CLI tool that connects to the server and executes local tools (file operations, command execution) within the user's project

### Key Features

✅ **Dynamic Team Assembly**: Coordinator AI analyzes tasks and selects the right specialists
✅ **Free-Form Collaboration**: Agents discuss naturally with shared context, coordinator moderates
✅ **Bidirectional Tool Execution**: File/exec operations on client, full result propagation to agents
✅ **Real-time Streaming**: Agent messages and tool calls stream in real-time via WebSocket
✅ **Hot-Reloading Agents**: YAML agent definitions reload automatically on change
✅ **Multi-Provider Support**: OpenAI, Anthropic, Z.ai (GLM models), and Google Gemini - Ollama planned for Phase 2
✅ **Permission System**: Sandboxed tool execution with path validation and command allowlists
✅ **Session Management**: Persistent sessions with complete conversation history

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

**The easiest way to get started:**

```bash
# 1. Set your API key (choose one provider)
export ANTHROPIC_API_KEY="sk-ant-..."     # Get from https://console.anthropic.com/
# OR
export OPENAI_API_KEY="sk-..."            # Get from https://platform.openai.com/
# OR
export GEMINI_API_KEY="..."               # Get from https://aistudio.google.com/apikey
# OR
export ZAI_API_KEY="..."                  # Get from https://open.bigmodel.cn/

# 2. Run the interactive setup
./start.sh
```

### Manual Setup

If you prefer manual setup:

```bash
# 1. Build the binaries
make build

# 2. Start the server (in one terminal)
./bin/ensemble-server

# 3. Run a task (in another terminal)
./bin/ensemble run "analyze this project and tell me what it does"
```

### Example Commands

```bash
# View available agents
./bin/ensemble agents list
./bin/ensemble agents show developer

# Run tasks
./bin/ensemble run "review the project structure"
./bin/ensemble run "write a README explaining the architecture"
./bin/ensemble run "identify files that need unit tests"

# Manage sessions
./bin/ensemble sessions list
./bin/ensemble sessions show <session-id>
./bin/ensemble sessions delete <session-id>
```

### Configuration

Default configurations are in `config/`:
- `config/server.yaml` - Server settings
- `config/client.yaml` - Client permissions and settings

You can customize:
- LLM provider and model (Anthropic, OpenAI, Z.ai, or Google Gemini)
- Temperature settings per agent
- File and command permissions
- Agent system prompts (edit `agents/*.yaml` - hot-reloads!)

#### Gemini Provider Setup

Ensemble supports **two authentication methods** for Google Gemini:

**Option 1: API Key (Direct SDK)**
```bash
export GEMINI_API_KEY="..."  # Get from https://aistudio.google.com/apikey
```

**Option 2: OAuth via Gemini CLI (Recommended)**

This option uses OAuth authentication through the Gemini CLI, allowing you to use your existing Gemini Code Assist subscription without managing API keys.

```bash
# 1. Install Gemini CLI globally
npm install -g @google/gemini-cli

# 2. Authenticate with Google (opens browser)
gemini

# 3. Install and start the Node.js bridge
cd internal/server/provider/gemini/bridge
npm install
npm start  # Runs on port 3001

# 4. Configure Ensemble to use CLI mode
# Edit config/server.yaml:
#   providers:
#     gemini:
#       use_cli: true
#       bridge_url: "http://localhost:3001"

# 5. Start Ensemble server (in another terminal)
./bin/ensemble-server
```

**Benefits of CLI Mode:**
- OAuth authentication (no API key management)
- Use existing Gemini Code Assist subscription
- Full feature access (streaming, tool calling)

See `internal/server/provider/gemini/bridge/README.md` for detailed documentation.

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
