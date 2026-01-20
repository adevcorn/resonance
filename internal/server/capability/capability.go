package capability

import (
	"context"
	"encoding/json"

	"github.com/adevcorn/ensemble/internal/protocol"
)

// Capability represents an executable action that agents can perform
type Capability interface {
	// Name returns the unique identifier for this capability
	Name() string

	// Execute runs the capability with the given parameters
	Execute(ctx context.Context, params json.RawMessage) (json.RawMessage, error)

	// ExecutionLocation indicates where this capability should execute
	ExecutionLocation() protocol.ExecutionLocation
}
