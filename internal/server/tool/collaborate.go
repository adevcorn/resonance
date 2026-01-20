package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/adevcorn/ensemble/internal/server/agent"
)

// CollaborateTool enables agent-to-agent communication
type CollaborateTool struct {
	pool      *agent.Pool
	onMessage func(from string, input *protocol.CollaborateInput) error
}

// NewCollaborateTool creates a new collaborate tool
func NewCollaborateTool(pool *agent.Pool, onMessage func(string, *protocol.CollaborateInput) error) *CollaborateTool {
	return &CollaborateTool{
		pool:      pool,
		onMessage: onMessage,
	}
}

// Name returns the tool name
func (c *CollaborateTool) Name() string {
	return "collaborate"
}

// Description returns the tool description
func (c *CollaborateTool) Description() string {
	return "Send messages to other agents in the team. Use this to broadcast updates, ask specific agents for help, or signal task completion."
}

// Parameters returns the JSON Schema for collaborate parameters
func (c *CollaborateTool) Parameters() json.RawMessage {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"broadcast", "direct", "help", "complete"},
				"description": "Type of collaboration action: broadcast (all team), direct (specific agent), help (request help), complete (task done)",
			},
			"message": map[string]interface{}{
				"type":        "string",
				"description": "The message to send",
			},
			"to_agent": map[string]interface{}{
				"type":        "string",
				"description": "Target agent name (required for 'direct' and 'help' actions)",
			},
			"artifacts": map[string]interface{}{
				"type":        "array",
				"description": "File paths, code snippets, or other artifacts to share",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
		},
		"required": []string{"action", "message"},
	}

	data, _ := json.Marshal(schema)
	return data
}

// Execute handles the collaborate action
func (c *CollaborateTool) Execute(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var colInput protocol.CollaborateInput
	if err := json.Unmarshal(input, &colInput); err != nil {
		return nil, fmt.Errorf("invalid collaborate input: %w", err)
	}

	// Validate action
	switch colInput.Action {
	case protocol.CollaborateBroadcast:
		// No additional validation needed
	case protocol.CollaborateDirect, protocol.CollaborateHelp:
		if colInput.ToAgent == "" {
			return nil, fmt.Errorf("to_agent is required for %s action", colInput.Action)
		}
		// Validate that the target agent exists
		if c.pool != nil {
			_, err := c.pool.Get(colInput.ToAgent)
			if err != nil {
				// Return error with smart suggestion
				return nil, fmt.Errorf("invalid target agent: %w", err)
			}
		}
	case protocol.CollaborateComplete:
		// No additional validation needed
	default:
		return nil, fmt.Errorf("invalid action: %s", colInput.Action)
	}

	if colInput.Message == "" {
		return nil, fmt.Errorf("message cannot be empty")
	}

	// Determine the sender from context (will be set by orchestration engine)
	from := "unknown"
	if ctx.Value("agent_name") != nil {
		from = ctx.Value("agent_name").(string)
	}

	// Call the onMessage handler
	if err := c.onMessage(from, &colInput); err != nil {
		return nil, fmt.Errorf("collaboration failed: %w", err)
	}

	// Build response
	recipients := []string{}
	switch colInput.Action {
	case protocol.CollaborateBroadcast:
		recipients = []string{"all team members"}
	case protocol.CollaborateDirect, protocol.CollaborateHelp:
		recipients = []string{colInput.ToAgent}
	case protocol.CollaborateComplete:
		recipients = []string{"coordinator"}
	}

	output := protocol.CollaborateOutput{
		Delivered:  true,
		Recipients: recipients,
	}

	data, err := json.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output: %w", err)
	}

	return data, nil
}

// ExecutionLocation returns where this tool executes
func (c *CollaborateTool) ExecutionLocation() protocol.ExecutionLocation {
	return protocol.ExecutionLocationServer
}
