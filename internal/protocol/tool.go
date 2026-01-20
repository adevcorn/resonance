package protocol

import "encoding/json"

// ExecutionLocation defines where a tool should be executed
type ExecutionLocation string

const (
	ExecutionLocationServer ExecutionLocation = "server"
	ExecutionLocationClient ExecutionLocation = "client"
)

// ToolCall represents a request to execute a tool
type ToolCall struct {
	ID        string          `json:"id"`
	ToolName  string          `json:"tool_name"`
	Arguments json.RawMessage `json:"arguments"`
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	CallID string          `json:"call_id"`
	Result json.RawMessage `json:"result"`
	Error  string          `json:"error,omitempty"`
}

// ToolDefinition defines a tool's interface for LLM consumption
type ToolDefinition struct {
	Name              string            `json:"name"`
	Description       string            `json:"description"`
	Parameters        json.RawMessage   `json:"parameters"` // JSON Schema
	ExecutionLocation ExecutionLocation `json:"execution_location"`
}
