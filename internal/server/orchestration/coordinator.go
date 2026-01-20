package orchestration

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/adevcorn/ensemble/internal/server/agent"
	"github.com/adevcorn/ensemble/internal/server/provider"
	"github.com/adevcorn/ensemble/internal/server/tool"
)

// Coordinator orchestrates multi-agent collaboration
type Coordinator struct {
	coordinatorAgent *agent.Agent
	pool             *agent.Pool
	registry         *tool.Registry
	activeTeam       []string
}

// NewCoordinator creates a new coordinator
func NewCoordinator(pool *agent.Pool, registry *tool.Registry) (*Coordinator, error) {
	// Get the coordinator agent from the pool
	coordinatorAgent, err := pool.Get("coordinator")
	if err != nil {
		return nil, fmt.Errorf("coordinator agent not found in pool: %w", err)
	}

	return &Coordinator{
		coordinatorAgent: coordinatorAgent,
		pool:             pool,
		registry:         registry,
		activeTeam:       []string{},
	}, nil
}

// AnalyzeTask analyzes the user's task and determines the team
func (c *Coordinator) AnalyzeTask(ctx context.Context, task string) ([]string, error) {
	// Build prompt for coordinator to analyze task
	availableAgents := c.pool.List()
	var agentDescriptions []string
	for _, agentName := range availableAgents {
		if agentName == "coordinator" {
			continue // Skip coordinator itself in the list
		}
		agent, err := c.pool.Get(agentName)
		if err != nil {
			continue
		}
		agentDescriptions = append(agentDescriptions,
			fmt.Sprintf("- %s: %s", agentName, agent.Description()))
	}

	prompt := fmt.Sprintf(`You are tasked with analyzing a user request and assembling the right team of agents to accomplish it.

Available agents:
%s

User's task:
%s

Please analyze this task and use the assemble_team tool to select the most appropriate agents. Consider:
1. What type of work is required (coding, documentation, testing, etc.)
2. The complexity of the task
3. Which specialists would be most helpful

Always include yourself (coordinator) in the team to moderate the discussion.`,
		strings.Join(agentDescriptions, "\n"),
		task)

	// Build messages for the coordinator
	messages := []protocol.Message{
		{
			Role:    protocol.MessageRoleUser,
			Content: prompt,
		},
	}

	// Get tools allowed for coordinator (should include assemble_team)
	tools := c.registry.GetAllowed(c.coordinatorAgent)

	// Create completion request
	req := &provider.CompletionRequest{
		Messages: messages,
		Tools:    tools,
	}

	// Call coordinator agent
	resp, err := c.coordinatorAgent.Complete(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("coordinator failed to analyze task: %w", err)
	}

	// Look for assemble_team tool call in response
	var team []string
	for _, toolCall := range resp.ToolCalls {
		if toolCall.ToolName == "assemble_team" {
			var input tool.AssembleTeamInput
			if err := json.Unmarshal(toolCall.Arguments, &input); err != nil {
				return nil, fmt.Errorf("failed to parse assemble_team arguments: %w", err)
			}
			team = input.Agents
			break
		}
	}

	if len(team) == 0 {
		// If coordinator didn't use assemble_team, fallback to just coordinator
		team = []string{"coordinator"}
	}

	return team, nil
}

// GetActiveTeam returns the current team
func (c *Coordinator) GetActiveTeam() []string {
	return c.activeTeam
}

// SetActiveTeam sets the active team
func (c *Coordinator) SetActiveTeam(team []string) {
	c.activeTeam = team
}
