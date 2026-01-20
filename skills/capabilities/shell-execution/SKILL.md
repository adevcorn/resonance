---
name: shell-execution
description: Execute shell commands for builds, tests, and git operations
category: capability
capabilities:
  - execute_command
---

# Shell Command Execution

Run any shell command in the project directory.

## Important Note

**This is a client-side capability** - commands execute on the client machine, not the server.

## How to Execute

```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {
    "command": "go test ./...",
    "workdir": "."
  }
}
```

### Parameters

- `command` (required): The shell command to execute
- `workdir` (optional): Working directory relative to project root (defaults to project root)

## Common Use Cases

### Running Tests

Execute all tests:
```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {"command": "go test ./..."}
}
```

Run tests with verbose output:
```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {"command": "go test -v ./..."}
}
```

Run specific test:
```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {"command": "go test -run TestSkillSearch ./internal/server/skill"}
}
```

### Building

Build the server:
```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {"command": "go build -o bin/ensemble-server ./cmd/ensemble-server"}
}
```

Build all packages:
```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {"command": "go build ./..."}
}
```

### Git Operations

**See the `git-workflow` skill for detailed git patterns and best practices.**

Check status:
```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {"command": "git status"}
}
```

View diff:
```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {"command": "git diff"}
}
```

Stage files:
```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {"command": "git add ."}
}
```

Commit changes:
```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {"command": "git commit -m 'feat: add new capability'"}
}
```

### Code Quality

Run linter:
```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {"command": "golangci-lint run"}
}
```

Format code:
```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {"command": "gofmt -w ."}
}
```

Check for errors:
```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {"command": "go vet ./..."}
}
```

### Dependency Management

Download dependencies:
```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {"command": "go mod download"}
}
```

Tidy dependencies:
```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {"command": "go mod tidy"}
}
```

Update dependency:
```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {"command": "go get -u github.com/pkg/errors"}
}
```

### File Operations

Search for pattern:
```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {"command": "grep -r 'TODO' internal/"}
}
```

Count lines of code:
```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {"command": "wc -l internal/**/*.go"}
}
```

Find files:
```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {"command": "find internal/ -name '*_test.go'"}
}
```

## Working Directory

Specify a different working directory:

```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {
    "command": "npm install",
    "workdir": "internal/server/provider/gemini/bridge"
  }
}
```

This is equivalent to:
```bash
cd internal/server/provider/gemini/bridge
npm install
```

## Security & Best Practices

### Security
- Commands run with the client's permissions
- No interactive commands (they will hang)
- Use non-interactive flags (e.g., `git add .` not `git add -i`)
- Be careful with destructive commands

### Best Practices
- Verify commands before running
- Check exit codes in responses
- Handle errors appropriately
- Use `--help` to understand command options
- Test commands in safe environments first

### Non-Interactive Mode

Always use non-interactive flags:

**Good:**
```json
{"command": "git add ."}
{"command": "apt-get install -y package"}
{"command": "npm install --yes"}
```

**Bad (will hang):**
```json
{"command": "git add -i"}
{"command": "apt-get install package"}
{"command": "npm install"}  // if prompts appear
```

## Workflow Example

### Complete Build & Test Cycle

1. **Search for this skill**:
   ```json
   {"action": "search_skills", "query": "run commands"}
   ```

2. **Load this skill**:
   ```json
   {"action": "load_skill", "skill_name": "shell-execution"}
   ```

3. **Build the project**:
   ```json
   {"action": "execute", "capability": "execute_command", "parameters": {"command": "go build ./..."}}
   ```

4. **Run tests**:
   ```json
   {"action": "execute", "capability": "execute_command", "parameters": {"command": "go test ./..."}}
   ```

5. **Check code quality**:
   ```json
   {"action": "execute", "capability": "execute_command", "parameters": {"command": "go vet ./..."}}
   ```

6. **If all pass, commit changes**:
   ```json
   {"action": "execute", "capability": "execute_command", "parameters": {"command": "git add ."}}
   ```
   ```json
   {"action": "execute", "capability": "execute_command", "parameters": {"command": "git commit -m 'feat: implement feature'"}}
   ```

## Response Format

The client will return stdout, stderr, and exit code. Parse these to understand command results:

- Exit code 0: Success
- Exit code != 0: Error (check stderr for details)
- stdout: Command output
- stderr: Error messages and warnings

## Troubleshooting

**Command hangs/never returns**
- Likely an interactive command
- Use non-interactive flags
- Ensure command doesn't wait for user input

**Permission denied**
- Command requires elevated permissions
- Check file/directory permissions
- May need to modify permissions first

**Command not found**
- Binary not in PATH
- Use full path to binary
- Install required tool first

## Related Skills

- **git-workflow**: Detailed git patterns and commit conventions
- **filesystem-operations**: For file reading/writing operations
