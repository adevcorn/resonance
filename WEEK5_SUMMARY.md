# Week 5: Server Implementation

This week implements the HTTP API server that exposes the orchestration engine to clients.

## Implemented Components

### 1. Storage Layer (`internal/server/storage/`)

**storage.go** - Storage interface
- `Storage` interface defining CRUD operations for sessions

**json.go** - JSON file-based storage
- Thread-safe file operations with mutex
- Atomic writes (write to temp file, then rename)
- Sessions stored in `{basePath}/sessions/{id}.json`
- Proper error handling for concurrent access

**session.go** - Session manager
- Session lifecycle management (Create, Get, Update, Delete)
- UUID-based session IDs
- Helper methods:
  - `AddMessage` - Add message to session
  - `SetActiveTeam` - Set active team members
  - `SetState` - Update session state
- Automatic timestamp management (CreatedAt, UpdatedAt)

### 2. API Layer (`internal/server/api/`)

**router.go** - HTTP router
- Gorilla Mux-based routing
- Middleware stack (logging, recovery, CORS, request ID)
- Routes:
  - `POST /api/sessions` - Create session
  - `GET /api/sessions` - List sessions
  - `GET /api/sessions/:id` - Get session
  - `DELETE /api/sessions/:id` - Delete session
  - `GET /api/agents` - List agents
  - `GET /api/agents/:name` - Get agent details
  - `GET /api/sessions/:id/ws` - WebSocket for tasks
  - `GET /api/health` - Health check

**middleware.go** - HTTP middleware
- Request logging with zerolog
- Panic recovery with stack traces
- CORS headers
- Request ID tracking (X-Request-ID header)
- Response status code capture
- Helper functions for JSON responses

**sessions.go** - Session endpoints
- CRUD operations for sessions
- Query parameter support (`project_path` filter)
- Proper HTTP status codes (201 Created, 204 No Content, 404 Not Found)
- Request/response DTOs

**agents.go** - Agent endpoints
- List all agents with summaries
- Get detailed agent information
- AgentSummary and AgentDetail response types

**websocket.go** - WebSocket handler
- Connection upgrade
- Message types:
  - Client → Server: start, tool_result, cancel, ping
  - Server → Client: agent_message, tool_call, complete, error, pong
- Session validation
- Basic orchestration integration (simplified for MVP)

### 3. Server Binary (`cmd/ensemble-server/main.go`)

Complete server startup process:
1. **Configuration loading** - YAML config with environment variable expansion
2. **Logging setup** - Zerolog with configurable level
3. **Provider registration** - Anthropic and OpenAI providers
4. **Agent loading** - Load agents from YAML files
5. **Hot-reload watcher** - Optional fsnotify-based agent reloading
6. **Tool registry** - Register collaborate and assemble_team tools
7. **Storage initialization** - JSON file-based storage
8. **Engine creation** - Orchestration engine setup
9. **HTTP server** - Start HTTP server with proper timeouts
10. **Graceful shutdown** - Handle SIGINT/SIGTERM with 30s grace period

### 4. Comprehensive Tests

**json_test.go** - JSON storage tests
- Directory creation
- Session CRUD operations
- Concurrent access handling
- List and filter operations

**session_test.go** - Session manager tests
- Session lifecycle
- Message management
- Team management
- State management

**server_test.go** - API integration tests
- Session endpoints (create, get, update, delete, list)
- Agent endpoints (list, get details)
- Health endpoint
- HTTP status code validation
- Error handling

## Key Features

✅ **Thread-safe storage** - Mutex-based concurrency control  
✅ **Atomic writes** - Prevents file corruption  
✅ **Structured logging** - Zerolog with request tracking  
✅ **Graceful shutdown** - 30s timeout for in-flight requests  
✅ **CORS support** - Cross-origin requests enabled  
✅ **Error recovery** - Panic recovery middleware  
✅ **Request ID tracking** - End-to-end request tracing  
✅ **WebSocket support** - Real-time streaming (basic)  
✅ **Health checks** - `/api/health` endpoint  

## Testing

All tests pass:
```bash
go test ./internal/server/storage/... -v  # Storage tests
go test ./internal/server/api/... -v      # API tests
go test ./... -short                       # All tests
```

## Build

```bash
go build ./cmd/ensemble-server    # Build server binary
go build ./...                     # Build all packages
```

## Running the Server

```bash
./ensemble-server --config config/server.yaml
```

The server will:
1. Load configuration from `config/server.yaml`
2. Register LLM providers (Anthropic, OpenAI)
3. Load agent definitions from `./agents`
4. Start HTTP server on configured port
5. Handle graceful shutdown on SIGINT/SIGTERM

## Configuration

See `config/server.yaml` for configuration options:
- Server host/port
- Storage path
- Agent definitions path
- Provider API keys (via environment variables)
- Logging level and format

## Next Steps (Week 6: Client)

- [ ] CLI client implementation
- [ ] Local tool executor
- [ ] Permission system
- [ ] Project context detection
- [ ] WebSocket client
- [ ] Integration testing
- [ ] End-to-end demo

## Notes

### WebSocket Implementation

The current WebSocket implementation is simplified for the MVP. In production, we would:
- Create per-session engine instances with custom callbacks
- Stream agent messages in real-time
- Route tool calls to client and wait for results
- Handle cancellation properly
- Implement proper error recovery

This requires refactoring the orchestration engine to support per-session callbacks, which is planned for future iterations.

### Storage Improvements

Future improvements could include:
- Database backend (PostgreSQL, SQLite)
- Session pagination
- Session search/filtering
- Session archiving
- Backup/restore functionality
