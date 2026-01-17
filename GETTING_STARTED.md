# Getting Started with Ensemble

This guide will help you get started with Ensemble, the multi-agent coordination tool.

## Prerequisites

- **Go 1.23 or later** - [Install Go](https://golang.org/doc/install)
- **API Keys** for LLM providers:
  - Anthropic API key (recommended)
  - OpenAI API key (optional)
  - Google AI API key (optional)

## Installation

### 1. Clone the Repository

```bash
git clone https://github.com/adevcorn/ensemble.git
cd ensemble
```

### 2. Build the Binaries

```bash
# Build both client and server
make build

# Or build individually
make build-server
make build-client
```

This will create two binaries in the `bin/` directory:
- `bin/ensemble-server` - The server component
- `bin/ensemble` - The CLI client

### 3. Set Up Configuration

#### Server Configuration

Create a local server configuration:

```bash
cp config/server.yaml config/server.local.yaml
```

Edit `config/server.local.yaml` and configure your settings. The important settings are:

```yaml
server:
  host: "0.0.0.0"
  port: 8080

providers:
  anthropic:
    api_key: "${ANTHROPIC_API_KEY}"
    default_model: "claude-sonnet-4-20250514"
  openai:
    api_key: "${OPENAI_API_KEY}"
    default_model: "gpt-4o"
```

#### Client Configuration

Create a local client configuration:

```bash
cp config/client.yaml config/client.local.yaml
```

The default client configuration should work fine:

```yaml
client:
  server_url: "http://localhost:8080"

project:
  auto_detect: true

permissions:
  file:
    allowed_paths:
      - "."
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
```

### 4. Set Environment Variables

Set your API keys:

```bash
export ANTHROPIC_API_KEY="your-anthropic-api-key"
export OPENAI_API_KEY="your-openai-api-key"  # Optional
```

Add these to your `~/.bashrc`, `~/.zshrc`, or equivalent for persistence.

## Running Ensemble

### Start the Server

In one terminal, start the server:

```bash
./bin/ensemble-server
```

You should see output indicating the server is running:

```
INFO Server starting on 0.0.0.0:8080
INFO Loaded 9 agents from agents/
```

### Use the CLI Client

In another terminal, navigate to your project directory and run tasks:

```bash
# View available agents
./bin/ensemble agents list

# Show details for a specific agent
./bin/ensemble agents show developer

# Run a task
./bin/ensemble run "analyze this codebase and suggest improvements"

# List sessions
./bin/ensemble sessions list

# View configuration
./bin/ensemble config
```

## Your First Task

Let's run a simple task to see Ensemble in action:

```bash
cd /path/to/your/project
/path/to/ensemble/bin/ensemble run "read the README file and summarize it"
```

You should see:
1. Project detection
2. Session creation
3. Task execution with agent collaboration
4. Real-time streaming of agent messages
5. Tool calls being executed locally
6. Final summary

## Understanding the Architecture

### Server Component

The server (`ensemble-server`) runs centrally and handles:
- **Orchestration**: Coordinates agent collaboration
- **LLM Communication**: Talks to Anthropic, OpenAI, etc.
- **Session Management**: Tracks conversation state
- **Agent Pool**: Manages 9 specialized agents

### Client Component

The CLI client (`ensemble`) runs on your machine and:
- **Project Detection**: Automatically finds your project root
- **Local Tool Execution**: Runs file and command tools safely
- **Permission Enforcement**: Sandboxes operations to your project
- **Real-time Streaming**: Shows agent collaboration as it happens

### The 9 Default Agents

1. **Coordinator** - Orchestrates collaboration, assembles teams
2. **Developer** - Implements code, refactors, debugs
3. **Architect** - System design, technical planning
4. **Reviewer** - Code review, quality assurance
5. **Researcher** - Information gathering, documentation
6. **Security** - Security analysis, vulnerability assessment
7. **Writer** - Documentation, technical writing
8. **Tester** - Test creation, QA strategy
9. **DevOps** - CI/CD, deployment, infrastructure

## Example Tasks

Here are some example tasks you can try:

### Code Analysis
```bash
./bin/ensemble run "analyze the main.go file for potential improvements"
```

### Feature Implementation
```bash
./bin/ensemble run "add logging to all API endpoints"
```

### Documentation
```bash
./bin/ensemble run "create API documentation for all exported functions"
```

### Testing
```bash
./bin/ensemble run "write unit tests for the user service"
```

### Code Review
```bash
./bin/ensemble run "review the authentication code for security issues"
```

## Configuration Options

### Permissions

The client enforces strict permissions for security:

**File Permissions:**
- `allowed_paths`: Paths that can be accessed (default: project root)
- `denied_paths`: Paths that are blocked (e.g., `.git`, `.env`)

**Exec Permissions:**
- `allowed_commands`: Commands that can be executed
- `denied_commands`: Commands that are blocked

Agents can only access files within your project directory and execute allowed commands.

### Tool Permissions Per Agent

Each agent has specific tools they can use, defined in their YAML file:

```yaml
# agents/developer.yaml
tools:
  allowed:
    - read_file
    - write_file
    - list_directory
    - execute_command
    - collaborate
```

## Troubleshooting

### Server Won't Start

**Issue**: Server fails to start with "address already in use"

**Solution**: Another process is using port 8080. Change the port in `config/server.local.yaml`:

```yaml
server:
  port: 8081
```

And update client configuration accordingly.

### API Key Errors

**Issue**: "API key not found" or authentication errors

**Solution**: Verify your environment variables are set:

```bash
echo $ANTHROPIC_API_KEY
```

### Permission Denied

**Issue**: "access denied: path not allowed"

**Solution**: Check your client configuration's `permissions.file.allowed_paths` includes the path you're trying to access.

### Command Not Allowed

**Issue**: "access denied: command not allowed"

**Solution**: Add the command to `permissions.exec.allowed_commands` in `config/client.local.yaml`.

## Next Steps

- Read [PLAN.md](PLAN.md) for the complete technical specification
- Explore [agents/](agents/) to see how agents are defined
- Check out [ARCHITECTURE.md](ARCHITECTURE.md) for system design details
- Experiment with different tasks and agent combinations

## Getting Help

- Check the [README.md](README.md) for overview
- Review [PLAN.md](PLAN.md) for technical details
- Open an issue on GitHub (coming soon)

## Development

If you want to contribute or modify Ensemble:

```bash
# Run tests
make test

# Format code
make fmt

# Run linter
make lint

# Build and run server in one command
make run-server
```

Happy collaborating with AI agents! 🤖🎉
