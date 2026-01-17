package tool

import (
	"context"
	"encoding/json"

	"github.com/adevcorn/ensemble/internal/protocol"
)

// Tool interface for all tools (server and client)
type Tool interface {
	// Name returns the tool's unique identifier
	Name() string

	// Description returns a human-readable description of the tool
	Description() string

	// Parameters returns the JSON Schema for tool parameters
	Parameters() json.RawMessage

	// Execute runs the tool with the given input
	Execute(ctx context.Context, input json.RawMessage) (json.RawMessage, error)

	// ExecutionLocation indicates where this tool should execute
	ExecutionLocation() protocol.ExecutionLocation
}

// Func is a function-based tool implementation
type Func struct {
	name        string
	description string
	parameters  json.RawMessage
	location    protocol.ExecutionLocation
	handler     func(context.Context, json.RawMessage) (json.RawMessage, error)
}

// NewFunc creates a new function-based tool
func NewFunc(
	name, description string,
	params json.RawMessage,
	location protocol.ExecutionLocation,
	handler func(context.Context, json.RawMessage) (json.RawMessage, error),
) *Func {
	return &Func{
		name:        name,
		description: description,
		parameters:  params,
		location:    location,
		handler:     handler,
	}
}

// Name returns the tool's name
func (f *Func) Name() string {
	return f.name
}

// Description returns the tool's description
func (f *Func) Description() string {
	return f.description
}

// Parameters returns the tool's parameter schema
func (f *Func) Parameters() json.RawMessage {
	return f.parameters
}

// Execute runs the tool's handler function
func (f *Func) Execute(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	return f.handler(ctx, input)
}

// ExecutionLocation returns where this tool should execute
func (f *Func) ExecutionLocation() protocol.ExecutionLocation {
	return f.location
}
