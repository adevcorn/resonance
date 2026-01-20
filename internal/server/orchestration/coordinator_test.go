package orchestration

import (
	"context"
	"testing"

	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/adevcorn/ensemble/internal/server/agent"
	"github.com/adevcorn/ensemble/internal/server/provider"
	"github.com/adevcorn/ensemble/internal/server/provider/mock"
	"github.com/adevcorn/ensemble/internal/server/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestPoolForOrchestration() *agent.Pool {
	mockProvider := mock.NewMockProvider("mock", []string{"assemble_team"})
	registry := provider.NewRegistry()
	registry.Register(mockProvider)

	pool := agent.NewPool(registry)

	definitions := []*protocol.AgentDefinition{
		{
			Name:         "coordinator",
			DisplayName:  "Coordinator",
			Description:  "Coordinates multi-agent collaboration",
			SystemPrompt: "You are a coordinator",
			Model: protocol.ModelConfig{
				Provider: "mock",
			},
			Tools: protocol.ToolsConfig{
				Allowed: []string{"assemble_team", "collaborate"},
			},
		},
		{
			Name:        "developer",
			DisplayName: "Developer",
			Description: "Writes code",
			Model: protocol.ModelConfig{
				Provider: "mock",
			},
		},
		{
			Name:        "architect",
			DisplayName: "Architect",
			Description: "Designs systems",
			Model: protocol.ModelConfig{
				Provider: "mock",
			},
		},
	}

	_ = pool.Load(definitions)
	return pool
}

func TestNewCoordinator(t *testing.T) {
	pool := createTestPoolForOrchestration()
	registry := tool.NewRegistry()

	coord, err := NewCoordinator(pool, registry)
	require.NoError(t, err)
	assert.NotNil(t, coord)
	assert.NotNil(t, coord.coordinatorAgent)
}

func TestNewCoordinatorMissingAgent(t *testing.T) {
	// Create pool without coordinator agent
	mockProvider := mock.NewMockProvider("mock", []string{})
	providerRegistry := provider.NewRegistry()
	providerRegistry.Register(mockProvider)

	pool := agent.NewPool(providerRegistry)

	registry := tool.NewRegistry()

	_, err := NewCoordinator(pool, registry)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "coordinator agent not found")
}

func TestCoordinator_AnalyzeTask(t *testing.T) {
	pool := createTestPoolForOrchestration()
	registry := tool.NewRegistry()

	// Register assemble_team tool
	var capturedAgents []string
	assembleTool := tool.NewAssembleTeamTool(pool, func(agents []string, reason string) error {
		capturedAgents = agents
		return nil
	})
	_ = registry.Register(assembleTool)

	coord, err := NewCoordinator(pool, registry)
	require.NoError(t, err)

	ctx := context.Background()
	team, err := coord.AnalyzeTask(ctx, "Implement a new feature")
	require.NoError(t, err)

	// Should get a team (may be just coordinator if mock doesn't return tool call)
	assert.NotEmpty(t, team)
	_ = capturedAgents // May or may not be set depending on mock behavior
}

func TestCoordinator_GetSetActiveTeam(t *testing.T) {
	pool := createTestPoolForOrchestration()
	registry := tool.NewRegistry()

	coord, err := NewCoordinator(pool, registry)
	require.NoError(t, err)

	// Initially empty
	assert.Empty(t, coord.GetActiveTeam())

	// Set team
	team := []string{"coordinator", "developer"}
	coord.SetActiveTeam(team)

	// Get team
	retrieved := coord.GetActiveTeam()
	assert.Equal(t, team, retrieved)
}
