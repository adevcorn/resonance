package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/adevcorn/ensemble/internal/protocol"
)

// WriteFileTool is a client-side tool for writing files
type WriteFileTool struct{}

// NewWriteFileTool creates a new write_file tool
// This is a client-side tool - it doesn't execute on the server,
// but needs to be registered so agents know it's available
func NewWriteFileTool() *WriteFileTool {
	return &WriteFileTool{}
}

// Name returns the tool name
func (w *WriteFileTool) Name() string {
	return "write_file"
}

// Description returns the tool description
func (w *WriteFileTool) Description() string {
	return "Write content to a file. Creates the file if it doesn't exist, overwrites if it does. Creates parent directories as needed."
}

// Parameters returns the JSON Schema for write_file parameters
func (w *WriteFileTool) Parameters() json.RawMessage {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "File path relative to project root (e.g., 'README.md' or 'src/main.go')",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "Content to write to the file",
			},
		},
		"required": []string{"path", "content"},
	}

	data, _ := json.Marshal(schema)
	return data
}

// Execute should never be called on server - this is a client-side tool
func (w *WriteFileTool) Execute(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	// This should never be called because ExecutionLocation is Client
	// The client will handle execution
	return nil, fmt.Errorf("write_file is a client-side tool and should not be executed on server")
}

// ExecutionLocation returns where this tool executes
func (w *WriteFileTool) ExecutionLocation() protocol.ExecutionLocation {
	return protocol.ExecutionLocationClient
}
