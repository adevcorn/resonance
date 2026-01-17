package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/adevcorn/ensemble/internal/server/agent"
)

// AssembleTeamInput is the input for team assembly
type AssembleTeamInput struct {
	Agents []string `json:"agents"`
	Reason string   `json:"reason"`
}

// AssembleTeamOutput is the result of team assembly
type AssembleTeamOutput struct {
	Team    []string `json:"team"`
	Success bool     `json:"success"`
	Message string   `json:"message"`
}

// AssembleTeamTool allows coordinator to select team members
type AssembleTeamTool struct {
	pool       *agent.Pool
	onAssemble func(agents []string, reason string) error
}

// NewAssembleTeamTool creates a new assemble team tool
func NewAssembleTeamTool(pool *agent.Pool, onAssemble func([]string, string) error) *AssembleTeamTool {
	return &AssembleTeamTool{
		pool:       pool,
		onAssemble: onAssemble,
	}
}

// Name returns the tool name
func (a *AssembleTeamTool) Name() string {
	return "assemble_team"
}

// Description returns the tool description
func (a *AssembleTeamTool) Description() string {
	return "Assemble a team of agents to work on the task. Select agents based on their specialties and the requirements of the task."
}

// Parameters returns the JSON Schema for assemble_team parameters
func (a *AssembleTeamTool) Parameters() json.RawMessage {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"agents": map[string]interface{}{
				"type":        "array",
				"description": "List of agent names to include in the team",
				"items": map[string]interface{}{
					"type": "string",
				},
				"minItems": 1,
			},
			"reason": map[string]interface{}{
				"type":        "string",
				"description": "Explanation of why these agents were selected",
			},
		},
		"required": []string{"agents", "reason"},
	}

	data, _ := json.Marshal(schema)
	return data
}

// Execute assembles the team
func (a *AssembleTeamTool) Execute(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var teamInput AssembleTeamInput
	if err := json.Unmarshal(input, &teamInput); err != nil {
		return nil, fmt.Errorf("invalid assemble_team input: %w", err)
	}

	if len(teamInput.Agents) == 0 {
		return nil, fmt.Errorf("at least one agent must be specified")
	}

	if teamInput.Reason == "" {
		return nil, fmt.Errorf("reason for team assembly must be provided")
	}

	// Validate that all requested agents exist in the pool
	var invalidAgents []string
	for _, agentName := range teamInput.Agents {
		if !a.pool.Has(agentName) {
			invalidAgents = append(invalidAgents, agentName)
		}
	}

	if len(invalidAgents) > 0 {
		output := AssembleTeamOutput{
			Team:    nil,
			Success: false,
			Message: fmt.Sprintf("Invalid agents: %v. Available agents: %v", invalidAgents, a.pool.List()),
		}
		data, _ := json.Marshal(output)
		return data, nil
	}

	// Call the onAssemble callback
	if err := a.onAssemble(teamInput.Agents, teamInput.Reason); err != nil {
		output := AssembleTeamOutput{
			Team:    nil,
			Success: false,
			Message: fmt.Sprintf("Failed to assemble team: %v", err),
		}
		data, _ := json.Marshal(output)
		return data, nil
	}

	// Success
	output := AssembleTeamOutput{
		Team:    teamInput.Agents,
		Success: true,
		Message: fmt.Sprintf("Team assembled: %v. Reason: %s", teamInput.Agents, teamInput.Reason),
	}

	data, err := json.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output: %w", err)
	}

	return data, nil
}

// ExecutionLocation returns where this tool executes
func (a *AssembleTeamTool) ExecutionLocation() protocol.ExecutionLocation {
	return protocol.ExecutionLocationServer
}
