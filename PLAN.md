# Ensemble - Multi-Agent Coordination Tool

## Implementation Plan

**Project**: Ensemble  
**Language**: Go  
**Architecture**: Client-Server  
**Status**: Phase 1 Ready

---

## Table of Contents

1. [Overview](#overview)
2. [Key Design Decisions](#key-design-decisions)
3. [Architecture](#architecture)
4. [Project Structure](#project-structure)
5. [Core Concepts](#core-concepts)
6. [Implementation Phases](#implementation-phases)
7. [Phase 1 Detailed Breakdown](#phase-1-detailed-breakdown)
8. [Technical Specifications](#technical-specifications)
9. [Default Agents](#default-agents)
10. [Configuration Schemas](#configuration-schemas)
11. [API Specification](#api-specification)
12. [Dependencies](#dependencies)
13. [Session Continuation Prompt](#session-continuation-prompt)

---

## Overview

Ensemble is a multi-agent developer tool where a coordinating agent dynamically assembles teams of specialized agents from a pool to collaboratively accomplish software development tasks.

### Core Value Proposition

- **Dynamic Team Assembly**: Coordinator AI analyzes tasks and selects the right specialists
- **Free-Form Collaboration**: Agents discuss naturally with shared context, coordinator moderates
- **Hybrid Tool Execution**: File/exec operations on client, search/fetch on server
- **Hot-Reloading Agents**: YAML agent definitions reload automatically on change
- **Multi-Provider Support**: OpenAI, Anthropic, Google AI, Ollama

---

## Key Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Communication Pattern | Broadcast + Direct | Agents can message entire team or specific agents |
| Agent Selection | AI-Driven | Coordinator LLM selects appropriate agents for task |
| Context Management | Shared Context Window | All agents see full conversation history |
| Turn-Taking | Free-Form | Coordinator moderates, agents contribute naturally |
| Tool Permissions | Per-Agent Granular | Security and role-based access control |
| Model Flexibility | Per-Agent Models | Different agents can use different LLMs |
| Agent Hot-Reload | Automatic | fsnotify watches agent definitions |
| Default Agents | Ship with Defaults | 9 pre-configured specialist agents |
| Error Recovery | Coordinator Decides | Central error handling and retry logic |
| Deployment | Client-Server | Backend central, clients work locally |
| Sessions | Project-Scoped | Sessions tied to client's project directory |
| Tool Execution | Hybrid | File/exec on client, search/fetch on server |
| Persistence | File-Based JSON | Simple, portable, human-readable |

---

## Architecture

### High-Level Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         SERVER (Central)                         │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │   HTTP API  │  │  WebSocket  │  │    Session Manager      │  │
│  │  (REST)     │  │  (Streaming)│  │  (Project-Scoped)       │  │
│  └──────┬──────┘  └──────┬──────┘  └───────────┬─────────────┘  │
│         │                │                     │                 │
│  ┌──────▼────────────────▼─────────────────────▼─────────────┐  │
│  │                  Orchestration Engine                      │  │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐    │  │
│  │  │ Coordinator │  │  Moderator  │  │   Synthesizer   │    │  │
│  │  │ (Team Asm.) │  │ (Turn Mgmt) │  │ (Result Merge)  │    │  │
│  │  └─────────────┘  └─────────────┘  └─────────────────┘    │  │
│  └───────────────────────────┬───────────────────────────────┘  │
│                              │                                   │
│  ┌───────────────────────────▼───────────────────────────────┐  │
│  │                      Agent Pool                            │  │
│  │  ┌────────┐ ┌──────────┐ ┌────────┐ ┌──────────┐         │  │
│  │  │Developer│ │Architect │ │Reviewer│ │Researcher│ ...     │  │
│  │  └────────┘ └──────────┘ └────────┘ └──────────┘         │  │
│  └───────────────────────────┬───────────────────────────────┘  │
│                              │                                   │
│  ┌───────────────────────────▼───────────────────────────────┐  │
│  │                    LLM Providers                           │  │
│  │  ┌─────────┐ ┌────────┐ ┌────────┐ ┌──────┐              │  │
│  │  │Anthropic│ │ OpenAI │ │ Google │ │Ollama│              │  │
│  │  └─────────┘ └────────┘ └────────┘ └──────┘              │  │
│  └───────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌─────────────────┐  ┌─────────────────────────────────────┐  │
│  │  Server Tools   │  │         JSON Storage                 │  │
│  │  - web_search   │  │  - Sessions                          │  │
│  │  - fetch_url    │  │  - Agent Definitions                 │  │
│  └─────────────────┘  └─────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
                    WebSocket/HTTP
                              │
┌─────────────────────────────▼───────────────────────────────────┐
│                      CLIENT (User's Machine)                     │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────────────────────────┐  │
│  │    CLI / TUI    │  │         Project Context              │  │
│  │                 │  │  - Working Directory                 │  │
│  │  $ ensemble run │  │  - Git Info                          │  │
│  │    "implement   │  │  - File Tree                         │  │
│  │     feature X"  │  └─────────────────────────────────────┘  │
│  └────────┬────────┘                                            │
│           │                                                      │
│  ┌────────▼─────────────────────────────────────────────────┐  │
│  │                 Local Tool Executor                       │  │
│  │  ┌───────────┐ ┌─────────────┐ ┌───────────────────────┐ │  │
│  │  │ read_file │ │ write_file  │ │ execute_command       │ │  │
│  │  │ list_dir  │ │ (sandboxed) │ │ (within project)      │ │  │
│  │  └───────────┘ └─────────────┘ └───────────────────────┘ │  │
│  │                                                           │  │
│  │  ┌───────────────────────────────────────────────────┐   │  │
│  │  │              MCP Client (Optional)                 │   │  │
│  │  │  - Connect to local MCP servers                    │   │  │
│  │  │  - Expose MCP tools to agents                      │   │  │
│  │  └───────────────────────────────────────────────────┘   │  │
│  └───────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Component | Location | Responsibility |
|-----------|----------|----------------|
| HTTP API | Server | REST endpoints for sessions, agents, config |
| WebSocket | Server | Real-time streaming of collaboration |
| Session Manager | Server | Project-scoped session lifecycle |
| Coordinator | Server | AI-driven team assembly via `assemble_team` tool |
| Moderator | Server | Free-form turn management, decides who speaks next |
| Synthesizer | Server | Merges agent outputs into final response |
| Agent Pool | Server | Hot-reloadable agent definitions |
| LLM Providers | Server | Abstraction over Anthropic, OpenAI, Google, Ollama |
| Server Tools | Server | `web_search`, `fetch_url` |
| JSON Storage | Server | Sessions, agent definitions |
| CLI/TUI | Client | User interface |
| Project Context | Client | Working directory, git info, file tree |
| Local Executor | Client | Sandboxed file/exec operations |
| MCP Client | Client | Optional MCP server integration |

---

## Project Structure

```
ensemble/
├── cmd/
│   ├── ensemble/                    # CLI client binary
│   │   └── main.go
│   └── ensemble-server/             # Server binary
│       └── main.go
│
├── internal/
│   ├── server/                      # Server-side code
│   │   ├── api/                     # HTTP/WebSocket handlers
│   │   │   ├── router.go            # Route definitions
│   │   │   ├── sessions.go          # Session endpoints
│   │   │   ├── agents.go            # Agent endpoints
│   │   │   ├── websocket.go         # WebSocket handler
│   │   │   └── middleware.go        # Auth, logging, etc.
│   │   │
│   │   ├── orchestration/           # Core orchestration
│   │   │   ├── coordinator.go       # Team assembly logic
│   │   │   ├── moderator.go         # Turn management
│   │   │   ├── synthesizer.go       # Result merging
│   │   │   └── engine.go            # Main orchestration loop
│   │   │
│   │   ├── agent/                   # Agent management
│   │   │   ├── pool.go              # Agent pool
│   │   │   ├── loader.go            # YAML loader
│   │   │   ├── watcher.go           # Hot-reload (fsnotify)
│   │   │   └── agent.go             # Agent struct/methods
│   │   │
│   │   ├── provider/                # LLM providers
│   │   │   ├── provider.go          # Provider interface
│   │   │   ├── anthropic/
│   │   │   │   └── anthropic.go
│   │   │   ├── openai/
│   │   │   │   └── openai.go
│   │   │   ├── google/
│   │   │   │   └── google.go
│   │   │   └── ollama/
│   │   │       └── ollama.go
│   │   │
│   │   ├── tool/                    # Server-side tools
│   │   │   ├── registry.go          # Tool registry
│   │   │   ├── tool.go              # Tool interface
│   │   │   ├── web_search.go
│   │   │   └── fetch_url.go
│   │   │
│   │   └── storage/                 # Persistence
│   │       ├── storage.go           # Storage interface
│   │       ├── json.go              # JSON file storage
│   │       └── session.go           # Session storage
│   │
│   ├── client/                      # Client-side code
│   │   ├── client.go                # Server connection
│   │   ├── executor.go              # Local tool executor
│   │   ├── project.go               # Project context
│   │   ├── permissions.go           # Tool permission system
│   │   └── tools/                   # Client-executed tools
│   │       ├── file.go              # read_file, write_file, list_directory
│   │       ├── exec.go              # execute_command
│   │       └── mcp.go               # MCP client integration
│   │
│   ├── protocol/                    # Shared types (client/server)
│   │   ├── message.go               # Message types
│   │   ├── tool.go                  # Tool call/result types
│   │   ├── session.go               # Session types
│   │   ├── agent.go                 # Agent definition types
│   │   └── collaborate.go           # Collaborate tool types
│   │
│   └── config/                      # Configuration
│       ├── server.go                # Server config loader
│       └── client.go                # Client config loader
│
├── pkg/ensemble/                    # Public SDK (future)
│   └── client.go                    # Programmatic client
│
├── agents/                          # Default agent definitions
│   ├── coordinator.yaml
│   ├── developer.yaml
│   ├── architect.yaml
│   ├── reviewer.yaml
│   ├── researcher.yaml
│   ├── security.yaml
│   ├── writer.yaml
│   ├── tester.yaml
│   └── devops.yaml
│
├── config/                          # Configuration files
│   ├── server.yaml                  # Server configuration
│   └── client.yaml                  # Client configuration
│
├── go.mod
├── go.sum
├── Makefile
├── README.md
└── PLAN.md                          # This document
```

---

## Core Concepts

### 1. Agent Pool

Agents are defined in YAML files and hot-reloaded on change.

```yaml
# agents/developer.yaml
name: developer
display_name: "Developer"
description: "Expert software developer for code implementation"

system_prompt: |
  You are an expert software developer working as part of a collaborative team.
  Your role is to implement features, write clean code, and follow best practices.
  
  When collaborating:
  - Read existing code before making changes
  - Follow the project's coding conventions
  - Write clear, maintainable code
  - Add appropriate comments and documentation
  - Consider edge cases and error handling

capabilities:
  - code_implementation
  - refactoring
  - debugging
  - api_design

model:
  provider: anthropic
  name: claude-sonnet-4-20250514
  temperature: 0.3
  max_tokens: 8192

tools:
  allowed:
    - read_file
    - write_file
    - list_directory
    - execute_command
    - collaborate
  denied: []
```

### 2. Coordinator Agent

The coordinator is a special meta-agent that:

1. Analyzes incoming tasks
2. Assembles appropriate teams using `assemble_team` tool
3. Moderates free-form discussion
4. Synthesizes final results

```yaml
# agents/coordinator.yaml
name: coordinator
display_name: "Coordinator"
description: "Orchestrates multi-agent collaboration"

system_prompt: |
  You are the coordinator of a multi-agent development team. Your role is to:
  
  1. ANALYZE the user's request to understand what needs to be done
  2. ASSEMBLE the right team of specialists using the assemble_team tool
  3. MODERATE the discussion, ensuring productive collaboration
  4. SYNTHESIZE the final result from agent contributions
  
  Available agents and their specialties:
  - developer: Code implementation, refactoring, debugging
  - architect: System design, architecture decisions, technical planning
  - reviewer: Code review, quality assurance, best practices
  - researcher: Information gathering, documentation research
  - security: Security analysis, vulnerability assessment
  - writer: Documentation, technical writing
  - tester: Test creation, test strategy, QA
  - devops: CI/CD, deployment, infrastructure
  
  When assembling a team:
  - Select only the agents needed for the specific task
  - Consider task complexity when choosing team size
  - Include yourself in the team to moderate
  
  During collaboration:
  - Guide the discussion toward the goal
  - Ask clarifying questions when needed
  - Ensure all perspectives are considered
  - Resolve conflicts between agents
  - Synthesize a coherent final response

capabilities:
  - task_analysis
  - team_assembly
  - moderation
  - synthesis

model:
  provider: anthropic
  name: claude-sonnet-4-20250514
  temperature: 0.5
  max_tokens: 8192

tools:
  allowed:
    - assemble_team
    - collaborate
    - read_file
    - list_directory
```

### 3. Collaborate Tool

Unified mechanism for agent communication:

```go
// internal/protocol/collaborate.go

type CollaborateAction string

const (
    CollaborateBroadcast CollaborateAction = "broadcast"  // Message to all team members
    CollaborateDirect    CollaborateAction = "direct"     // Message to specific agent
    CollaborateHelp      CollaborateAction = "help"       // Request help from another agent
    CollaborateComplete  CollaborateAction = "complete"   // Signal task completion
)

type CollaborateInput struct {
    Action    CollaborateAction `json:"action"`
    Message   string            `json:"message"`
    ToAgent   string            `json:"to_agent,omitempty"`   // For direct/help
    Artifacts []string          `json:"artifacts,omitempty"`  // File paths, code snippets
}

type CollaborateOutput struct {
    Delivered bool     `json:"delivered"`
    Recipients []string `json:"recipients"`
}
```

### 4. Shared Context Window

All team members see the full conversation:

```go
// internal/protocol/message.go

type Message struct {
    ID        string            `json:"id"`
    SessionID string            `json:"session_id"`
    Role      MessageRole       `json:"role"`       // user, assistant, system, tool
    Agent     string            `json:"agent"`      // Which agent sent this
    Content   string            `json:"content"`
    ToolCalls []ToolCall        `json:"tool_calls,omitempty"`
    ToolResults []ToolResult    `json:"tool_results,omitempty"`
    Timestamp time.Time         `json:"timestamp"`
    Metadata  map[string]any    `json:"metadata,omitempty"`
}

type ConversationContext struct {
    Messages     []Message       `json:"messages"`
    ActiveTeam   []string        `json:"active_team"`
    CurrentTask  string          `json:"current_task"`
    ProjectInfo  *ProjectInfo    `json:"project_info"`
}
```

### 5. Tool Execution Model

```
┌─────────────────────────────────────────────────────────────┐
│                    Tool Call Flow                            │
└─────────────────────────────────────────────────────────────┘

Agent generates tool call
         │
         ▼
┌─────────────────────┐
│  Server receives    │
│  tool call          │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Check tool type    │
└──────────┬──────────┘
           │
     ┌─────┴─────┐
     │           │
     ▼           ▼
┌─────────┐ ┌─────────┐
│ Server  │ │ Client  │
│ Tool?   │ │ Tool?   │
└────┬────┘ └────┬────┘
     │           │
     ▼           ▼
┌─────────┐ ┌─────────────────────────┐
│ Execute │ │ Route to client via WS  │
│ on      │ │                         │
│ server  │ │ ┌─────────────────────┐ │
│         │ │ │ Check permissions   │ │
│ Return  │ │ │         │           │ │
│ result  │ │ │         ▼           │ │
└────┬────┘ │ │ Execute in sandbox  │ │
     │      │ │         │           │ │
     │      │ │         ▼           │ │
     │      │ │ Return result       │ │
     │      │ └─────────────────────┘ │
     │      └────────────┬────────────┘
     │                   │
     └───────┬───────────┘
             │
             ▼
┌─────────────────────┐
│  Return result to   │
│  agent              │
└─────────────────────┘

Server Tools:          Client Tools:
- web_search           - read_file
- fetch_url            - write_file
                       - list_directory
                       - execute_command
                       - mcp_* (from MCP servers)
```

---

## Implementation Phases

### Phase 1: Core Foundation (5-6 weeks)

**Goal**: Minimal viable multi-agent system with CLI

| Week | Focus | Deliverables |
|------|-------|--------------|
| 1 | Project Setup | Go modules, project structure, config system |
| 2 | Providers | Provider interface, Anthropic + OpenAI implementations |
| 3 | Agents | Agent loader, pool, hot-reload watcher |
| 4 | Orchestration | Coordinator, moderator, collaborate tool |
| 5 | Server | HTTP API, WebSocket, session management |
| 6 | Client | CLI, local executor, integration testing |

### Phase 2: Enhanced Features (3-4 weeks)

**Goal**: Full provider support, MCP integration, better UX

- Google AI + Ollama providers
- Server tools (web_search, fetch_url)
- MCP client integration
- Session resume capability
- Interactive mode for CLI

### Phase 3: Production Hardening (3-4 weeks)

**Goal**: Production-ready with auth, observability

- Authentication system
- Rate limiting
- Structured logging
- Metrics/tracing
- Comprehensive documentation
- Performance optimization

### Phase 4: Frontends (4-6 weeks)

**Goal**: Rich user interfaces

- TUI with bubbletea
- Web frontend
- VS Code extension (stretch)

---

## Phase 1 Detailed Breakdown

### Week 1: Project Setup

#### Tasks

1. **Initialize Go Module**
   ```bash
   go mod init github.com/adevcorn/ensemble
   ```

2. **Create Directory Structure**
   - All directories from project structure

3. **Configuration System**
   - `internal/config/server.go` - Server config loader
   - `internal/config/client.go` - Client config loader
   - Support YAML, env vars, CLI flags

4. **Core Protocol Types**
   - `internal/protocol/message.go`
   - `internal/protocol/tool.go`
   - `internal/protocol/session.go`
   - `internal/protocol/agent.go`
   - `internal/protocol/collaborate.go`

#### Deliverables
- [ ] Go module initialized
- [ ] Directory structure created
- [ ] Config loading works
- [ ] Protocol types defined

### Week 2: LLM Providers

#### Tasks

1. **Provider Interface**
   ```go
   // internal/server/provider/provider.go
   type Provider interface {
       Name() string
       Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
       Stream(ctx context.Context, req *CompletionRequest) (<-chan StreamEvent, error)
       SupportsTools() bool
   }
   
   type CompletionRequest struct {
       Model       string
       Messages    []Message
       Tools       []ToolDefinition
       Temperature float64
       MaxTokens   int
   }
   
   type CompletionResponse struct {
       Content   string
       ToolCalls []ToolCall
       Usage     Usage
   }
   ```

2. **Anthropic Provider**
   - Use `github.com/anthropics/anthropic-sdk-go`
   - Implement streaming
   - Handle tool calls

3. **OpenAI Provider**
   - Use `github.com/openai/openai-go/v3`
   - Implement streaming
   - Handle tool calls

4. **Provider Registry**
   - Register providers by name
   - Select provider for agent

#### Deliverables
- [ ] Provider interface defined
- [ ] Anthropic provider working
- [ ] OpenAI provider working
- [ ] Streaming works for both
- [ ] Tool calls work for both

### Week 3: Agent System

#### Tasks

1. **Agent Definition Types**
   ```go
   // internal/protocol/agent.go
   type AgentDefinition struct {
       Name         string       `yaml:"name"`
       DisplayName  string       `yaml:"display_name"`
       Description  string       `yaml:"description"`
       SystemPrompt string       `yaml:"system_prompt"`
       Capabilities []string     `yaml:"capabilities"`
       Model        ModelConfig  `yaml:"model"`
       Tools        ToolsConfig  `yaml:"tools"`
   }
   
   type ModelConfig struct {
       Provider    string  `yaml:"provider"`
       Name        string  `yaml:"name"`
       Temperature float64 `yaml:"temperature"`
       MaxTokens   int     `yaml:"max_tokens"`
   }
   
   type ToolsConfig struct {
       Allowed []string `yaml:"allowed"`
       Denied  []string `yaml:"denied"`
   }
   ```

2. **Agent Loader**
   - Parse YAML files
   - Validate definitions
   - Error handling for malformed files

3. **Agent Pool**
   - Store loaded agents
   - Lookup by name
   - List available agents

4. **Hot-Reload Watcher**
   - Use `fsnotify` to watch agents directory
   - Reload on file changes
   - Handle add/remove/modify

#### Deliverables
- [ ] Agent YAML parser working
- [ ] Agent pool manages agents
- [ ] Hot-reload works
- [ ] All 9 default agents defined

### Week 4: Orchestration

#### Tasks

1. **Tool Registry**
   ```go
   // internal/server/tool/registry.go
   type ToolRegistry struct {
       tools map[string]Tool
   }
   
   func (r *ToolRegistry) Register(tool Tool)
   func (r *ToolRegistry) Get(name string) (Tool, bool)
   func (r *ToolRegistry) GetAllowed(agentName string) []ToolDefinition
   ```

2. **Collaborate Tool**
   - Implement broadcast, direct, help, complete actions
   - Integration with orchestration engine

3. **Assemble Team Tool**
   - Coordinator uses this to select agents
   - Returns team composition

4. **Coordinator Logic**
   - Analyze task
   - Select appropriate agents
   - Initialize team context

5. **Moderator Logic**
   - Decide which agent speaks next
   - Handle free-form discussion
   - Detect completion conditions

6. **Synthesizer Logic**
   - Merge agent contributions
   - Generate final response

#### Deliverables
- [ ] Tool registry works
- [ ] Collaborate tool implemented
- [ ] Assemble team tool implemented
- [ ] Coordinator selects agents
- [ ] Moderator manages turns
- [ ] Synthesizer merges results

### Week 5: Server

#### Tasks

1. **Session Manager**
   ```go
   // internal/server/storage/session.go
   type Session struct {
       ID           string
       ProjectPath  string
       CreatedAt    time.Time
       UpdatedAt    time.Time
       Messages     []Message
       ActiveTeam   []string
       State        SessionState
   }
   
   type SessionManager interface {
       Create(projectPath string) (*Session, error)
       Get(id string) (*Session, error)
       Update(session *Session) error
       Delete(id string) error
       ListByProject(projectPath string) ([]*Session, error)
   }
   ```

2. **JSON Storage**
   - File-based session storage
   - Atomic writes
   - Proper locking

3. **HTTP API**
   ```
   POST   /api/sessions           - Create session
   GET    /api/sessions/:id       - Get session
   DELETE /api/sessions/:id       - Delete session
   GET    /api/agents             - List agents
   GET    /api/agents/:name       - Get agent details
   POST   /api/sessions/:id/run   - Run task (returns WS upgrade)
   ```

4. **WebSocket Handler**
   - Streaming collaboration output
   - Tool call routing to client
   - Bidirectional communication

5. **Error Recovery**
   - Coordinator handles errors
   - Retry logic
   - Graceful degradation

#### Deliverables
- [ ] Session CRUD works
- [ ] JSON storage persists sessions
- [ ] HTTP API endpoints work
- [ ] WebSocket streaming works
- [ ] Tool calls route correctly

### Week 6: Client

#### Tasks

1. **Server Connection**
   - HTTP client for REST endpoints
   - WebSocket client for streaming
   - Reconnection logic

2. **Local Tool Executor**
   ```go
   // internal/client/executor.go
   type LocalExecutor struct {
       projectPath string
       permissions *PermissionConfig
   }
   
   func (e *LocalExecutor) Execute(call ToolCall) (*ToolResult, error)
   ```

3. **Permission System**
   - Allow/deny tool execution
   - Path restrictions (project sandbox)
   - Command restrictions

4. **CLI Commands**
   ```bash
   ensemble run "implement feature X"     # Run task
   ensemble agents list                    # List available agents
   ensemble agents show developer          # Show agent details
   ensemble sessions list                  # List sessions
   ensemble sessions show <id>             # Show session
   ensemble config                         # Show/edit config
   ```

5. **Project Context**
   - Detect project root
   - Gather git info
   - Build file tree

6. **Integration Testing**
   - End-to-end tests
   - Mock providers for testing
   - Test all agent configurations

#### Deliverables
- [ ] CLI connects to server
- [ ] `run` command works end-to-end
- [ ] Local tools execute correctly
- [ ] Permissions enforce sandbox
- [ ] All CLI commands work
- [ ] Integration tests pass

---

## Technical Specifications

### Tool Interface

```go
// internal/server/tool/tool.go

type Tool interface {
    Name() string
    Description() string
    Parameters() json.RawMessage  // JSON Schema
    Execute(ctx context.Context, input json.RawMessage) (json.RawMessage, error)
    ExecutionLocation() ExecutionLocation
}

type ExecutionLocation string

const (
    ExecuteOnServer ExecutionLocation = "server"
    ExecuteOnClient ExecutionLocation = "client"
)
```

### Client Tool Definitions

```go
// internal/client/tools/file.go

// read_file - Read file contents
type ReadFileInput struct {
    Path string `json:"path"`
}

type ReadFileOutput struct {
    Content string `json:"content"`
    Error   string `json:"error,omitempty"`
}

// write_file - Write content to file
type WriteFileInput struct {
    Path    string `json:"path"`
    Content string `json:"content"`
}

type WriteFileOutput struct {
    Success bool   `json:"success"`
    Error   string `json:"error,omitempty"`
}

// list_directory - List directory contents
type ListDirectoryInput struct {
    Path      string `json:"path"`
    Recursive bool   `json:"recursive"`
}

type ListDirectoryOutput struct {
    Entries []DirEntry `json:"entries"`
    Error   string     `json:"error,omitempty"`
}

// internal/client/tools/exec.go

// execute_command - Run shell command
type ExecuteCommandInput struct {
    Command string   `json:"command"`
    Args    []string `json:"args"`
    Cwd     string   `json:"cwd,omitempty"`
}

type ExecuteCommandOutput struct {
    Stdout   string `json:"stdout"`
    Stderr   string `json:"stderr"`
    ExitCode int    `json:"exit_code"`
    Error    string `json:"error,omitempty"`
}
```

### WebSocket Protocol

```go
// Client -> Server
type WSClientMessage struct {
    Type    string          `json:"type"`
    Payload json.RawMessage `json:"payload"`
}

// Types: "tool_result", "cancel", "ping"

// Server -> Client
type WSServerMessage struct {
    Type    string          `json:"type"`
    Payload json.RawMessage `json:"payload"`
}

// Types: "agent_message", "tool_call", "complete", "error", "pong"

// Tool call routing
type ToolCallMessage struct {
    CallID    string          `json:"call_id"`
    ToolName  string          `json:"tool_name"`
    Arguments json.RawMessage `json:"arguments"`
}

type ToolResultMessage struct {
    CallID string          `json:"call_id"`
    Result json.RawMessage `json:"result"`
    Error  string          `json:"error,omitempty"`
}
```

---

## Default Agents

### 1. Coordinator (`agents/coordinator.yaml`)

**Role**: Orchestrates multi-agent collaboration

**Capabilities**: task_analysis, team_assembly, moderation, synthesis

**Tools**: assemble_team, collaborate, read_file, list_directory

### 2. Developer (`agents/developer.yaml`)

**Role**: Code implementation and debugging

**Capabilities**: code_implementation, refactoring, debugging, api_design

**Tools**: read_file, write_file, list_directory, execute_command, collaborate

### 3. Architect (`agents/architect.yaml`)

**Role**: System design and architecture decisions

**Capabilities**: system_design, architecture_review, technical_planning, trade_off_analysis

**Tools**: read_file, list_directory, collaborate

### 4. Reviewer (`agents/reviewer.yaml`)

**Role**: Code review and quality assurance

**Capabilities**: code_review, quality_assurance, best_practices, security_review

**Tools**: read_file, list_directory, collaborate

### 5. Researcher (`agents/researcher.yaml`)

**Role**: Information gathering and research

**Capabilities**: information_gathering, documentation_research, api_exploration

**Tools**: read_file, list_directory, web_search, fetch_url, collaborate

### 6. Security (`agents/security.yaml`)

**Role**: Security analysis and vulnerability assessment

**Capabilities**: security_analysis, vulnerability_assessment, threat_modeling

**Tools**: read_file, list_directory, collaborate

### 7. Writer (`agents/writer.yaml`)

**Role**: Documentation and technical writing

**Capabilities**: documentation, technical_writing, api_documentation

**Tools**: read_file, write_file, list_directory, collaborate

### 8. Tester (`agents/tester.yaml`)

**Role**: Test creation and QA

**Capabilities**: test_creation, test_strategy, quality_assurance

**Tools**: read_file, write_file, list_directory, execute_command, collaborate

### 9. DevOps (`agents/devops.yaml`)

**Role**: CI/CD, deployment, and infrastructure

**Capabilities**: ci_cd, deployment, infrastructure, containerization

**Tools**: read_file, write_file, list_directory, execute_command, collaborate

---

## Configuration Schemas

### Server Configuration (`config/server.yaml`)

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  
storage:
  type: "json"
  path: "./data"
  
agents:
  path: "./agents"
  watch: true  # Hot-reload enabled
  
providers:
  anthropic:
    api_key: "${ANTHROPIC_API_KEY}"
    default_model: "claude-sonnet-4-20250514"
  openai:
    api_key: "${OPENAI_API_KEY}"
    default_model: "gpt-4o"
  google:
    api_key: "${GOOGLE_AI_API_KEY}"
    default_model: "gemini-1.5-pro"
  ollama:
    host: "http://localhost:11434"
    default_model: "llama3"

defaults:
  provider: "anthropic"
  temperature: 0.3
  max_tokens: 8192
  
logging:
  level: "info"
  format: "json"
```

### Client Configuration (`config/client.yaml`)

```yaml
client:
  server_url: "http://localhost:8080"
  
project:
  auto_detect: true  # Auto-detect project root
  
permissions:
  file:
    allowed_paths:
      - "."  # Project root
    denied_paths:
      - ".git"
      - ".env"
      - "node_modules"
  exec:
    allowed_commands:
      - "go"
      - "npm"
      - "git"
      - "make"
    denied_commands:
      - "rm -rf"
      - "sudo"
      
mcp:
  servers: []  # MCP server configurations
  
logging:
  level: "info"
```

---

## API Specification

### REST Endpoints

#### Sessions

```
POST /api/sessions
Content-Type: application/json
{
  "project_path": "/path/to/project"
}

Response:
{
  "id": "session_abc123",
  "project_path": "/path/to/project",
  "created_at": "2025-01-17T10:00:00Z",
  "state": "active"
}
```

```
GET /api/sessions/:id

Response:
{
  "id": "session_abc123",
  "project_path": "/path/to/project",
  "created_at": "2025-01-17T10:00:00Z",
  "updated_at": "2025-01-17T10:05:00Z",
  "state": "active",
  "messages": [...],
  "active_team": ["coordinator", "developer", "reviewer"]
}
```

```
DELETE /api/sessions/:id

Response:
{
  "success": true
}
```

#### Agents

```
GET /api/agents

Response:
{
  "agents": [
    {
      "name": "coordinator",
      "display_name": "Coordinator",
      "description": "Orchestrates multi-agent collaboration",
      "capabilities": ["task_analysis", "team_assembly", ...]
    },
    ...
  ]
}
```

```
GET /api/agents/:name

Response:
{
  "name": "developer",
  "display_name": "Developer",
  "description": "Expert software developer",
  "capabilities": ["code_implementation", ...],
  "model": {
    "provider": "anthropic",
    "name": "claude-sonnet-4-20250514"
  },
  "tools": {
    "allowed": ["read_file", "write_file", ...]
  }
}
```

### WebSocket API

```
GET /api/sessions/:id/run
Upgrade: websocket

Client sends:
{
  "type": "start",
  "payload": {
    "task": "Implement user authentication"
  }
}

Server streams:
{
  "type": "agent_message",
  "payload": {
    "agent": "coordinator",
    "content": "I'll analyze this task...",
    "timestamp": "2025-01-17T10:00:00Z"
  }
}

{
  "type": "tool_call",
  "payload": {
    "call_id": "call_123",
    "tool_name": "read_file",
    "arguments": {"path": "src/auth.go"}
  }
}

Client responds:
{
  "type": "tool_result",
  "payload": {
    "call_id": "call_123",
    "result": {"content": "package auth..."}
  }
}

Server streams completion:
{
  "type": "complete",
  "payload": {
    "summary": "Authentication implemented successfully",
    "artifacts": ["src/auth.go", "src/auth_test.go"]
  }
}
```

---

## Dependencies

### Core Dependencies

```go
// go.mod

module github.com/adevcorn/ensemble

go 1.22

require (
    // LLM Providers
    github.com/anthropics/anthropic-sdk-go v0.2.0
    github.com/openai/openai-go/v3 v3.0.0
    
    // HTTP/WebSocket
    github.com/gorilla/mux v1.8.1
    github.com/gorilla/websocket v1.5.1
    
    // Configuration
    github.com/spf13/viper v1.18.2
    gopkg.in/yaml.v3 v3.0.1
    
    // File watching
    github.com/fsnotify/fsnotify v1.7.0
    
    // CLI
    github.com/spf13/cobra v1.8.0
    
    // Logging
    github.com/rs/zerolog v1.32.0
    
    // Testing
    github.com/stretchr/testify v1.9.0
)
```

### Optional Dependencies (Phase 2+)

```go
    // Google AI
    cloud.google.com/go/vertexai v0.6.0
    
    // TUI
    github.com/charmbracelet/bubbletea v0.25.0
    github.com/charmbracelet/lipgloss v0.10.0
    
    // MCP
    // (MCP Go SDK when available)
```

---

## Session Continuation Prompt

Use this prompt to continue implementation in a new session:

```
I'm building Ensemble, a multi-agent coordination tool in Go. The full design 
and implementation plan is in PLAN.md.

Key architecture:
- Client-server model: Server runs orchestration/LLM inference centrally, 
  CLI client connects and executes local tools (file/exec) in user's project
- Coordinator agent assembles teams dynamically using AI-driven selection
- Free-form collaboration with shared context window
- Hybrid tool execution: file/exec on client, search/fetch on server
- Project-scoped sessions with JSON persistence
- Hot-reloading agent definitions (YAML)
- Multi-provider support (Anthropic, OpenAI, Google, Ollama)

Current status: [UPDATE WITH CURRENT PROGRESS]

Next task: [SPECIFY NEXT TASK FROM PHASE 1]

Please read PLAN.md for full details and continue implementation.
```

---

## Checklist: Phase 1

### Week 1: Project Setup
- [ ] Go module initialized (`go mod init github.com/adevcorn/ensemble`)
- [ ] Directory structure created
- [ ] Server config loader (`internal/config/server.go`)
- [ ] Client config loader (`internal/config/client.go`)
- [ ] Protocol types: message.go, tool.go, session.go, agent.go, collaborate.go

### Week 2: LLM Providers
- [ ] Provider interface (`internal/server/provider/provider.go`)
- [ ] Anthropic provider with streaming
- [ ] OpenAI provider with streaming
- [ ] Provider registry
- [ ] Tool calls work for both providers

### Week 3: Agent System
- [ ] Agent definition YAML parser
- [ ] Agent pool (`internal/server/agent/pool.go`)
- [ ] Agent loader (`internal/server/agent/loader.go`)
- [ ] Hot-reload watcher (`internal/server/agent/watcher.go`)
- [ ] All 9 default agent YAML files

### Week 4: Orchestration
- [ ] Tool registry (`internal/server/tool/registry.go`)
- [ ] Collaborate tool
- [ ] Assemble team tool
- [ ] Coordinator (`internal/server/orchestration/coordinator.go`)
- [ ] Moderator (`internal/server/orchestration/moderator.go`)
- [ ] Synthesizer (`internal/server/orchestration/synthesizer.go`)
- [ ] Orchestration engine (`internal/server/orchestration/engine.go`)

### Week 5: Server
- [ ] JSON storage (`internal/server/storage/json.go`)
- [ ] Session manager (`internal/server/storage/session.go`)
- [ ] HTTP router (`internal/server/api/router.go`)
- [ ] Session endpoints (`internal/server/api/sessions.go`)
- [ ] Agent endpoints (`internal/server/api/agents.go`)
- [ ] WebSocket handler (`internal/server/api/websocket.go`)
- [ ] Server binary (`cmd/ensemble-server/main.go`)

### Week 6: Client
- [ ] Server connection (`internal/client/client.go`)
- [ ] Local tool executor (`internal/client/executor.go`)
- [ ] File tools (`internal/client/tools/file.go`)
- [ ] Exec tools (`internal/client/tools/exec.go`)
- [ ] Permission system (`internal/client/permissions.go`)
- [ ] Project context (`internal/client/project.go`)
- [ ] CLI commands (`cmd/ensemble/main.go`)
- [ ] Integration tests
- [ ] End-to-end demo working
