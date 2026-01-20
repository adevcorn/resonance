package capability

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/adevcorn/ensemble/internal/protocol"
)

// ExecuteCommandCapability is a client-side capability for executing shell commands
type ExecuteCommandCapability struct{}

// NewExecuteCommandCapability creates a new execute command capability
func NewExecuteCommandCapability() *ExecuteCommandCapability {
	return &ExecuteCommandCapability{}
}

// Name returns the capability name
func (e *ExecuteCommandCapability) Name() string {
	return "execute_command"
}

// Execute should never be called on server - this is a client-side capability
func (e *ExecuteCommandCapability) Execute(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	// This should never be called because ExecutionLocation is Client
	return nil, fmt.Errorf("execute_command is a client-side capability and should not be executed on server")
}

// ExecutionLocation returns where this capability executes
func (e *ExecuteCommandCapability) ExecutionLocation() protocol.ExecutionLocation {
	return protocol.ExecutionLocationClient
}
