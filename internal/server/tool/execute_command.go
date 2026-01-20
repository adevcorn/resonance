package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/adevcorn/ensemble/internal/protocol"
)

// ExecuteCommandTool is a client-side tool for executing shell commands
type ExecuteCommandTool struct{}

// NewExecuteCommandTool creates a new execute_command tool
// This is a client-side tool - it doesn't execute on the server,
// but needs to be registered so agents know it's available
func NewExecuteCommandTool() *ExecuteCommandTool {
	return &ExecuteCommandTool{}
}

// Name returns the tool name
func (e *ExecuteCommandTool) Name() string {
	return "execute_command"
}

// Description returns the tool description
func (e *ExecuteCommandTool) Description() string {
	return "Execute a shell command in the project directory. Use this to run builds, tests, git commands, etc. Returns stdout, stderr, and exit code."
}

// Parameters returns the JSON Schema for execute_command parameters
func (e *ExecuteCommandTool) Parameters() json.RawMessage {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "Shell command to execute (e.g., 'go test ./...' or 'git status')",
			},
			"workdir": map[string]interface{}{
				"type":        "string",
				"description": "Working directory relative to project root (optional, defaults to project root)",
			},
		},
		"required": []string{"command"},
	}

	data, _ := json.Marshal(schema)
	return data
}

// Execute should never be called on server - this is a client-side tool
func (e *ExecuteCommandTool) Execute(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	// This should never be called because ExecutionLocation is Client
	// The client will handle execution
	return nil, fmt.Errorf("execute_command is a client-side tool and should not be executed on server")
}

// ExecutionLocation returns where this tool executes
func (e *ExecuteCommandTool) ExecutionLocation() protocol.ExecutionLocation {
	return protocol.ExecutionLocationClient
}
