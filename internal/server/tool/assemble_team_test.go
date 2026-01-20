package tool

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/adevcorn/ensemble/internal/server/agent"
	"github.com/adevcorn/ensemble/internal/server/provider"
	"github.com/adevcorn/ensemble/internal/server/provider/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestPool() *agent.Pool {
	mockProvider := mock.NewMockProvider("mock", []string{})
	registry := provider.NewRegistry()
	registry.Register(mockProvider)

	pool := agent.NewPool(registry)

	// Load test agents
	definitions := []*protocol.AgentDefinition{
		{
			Name:        "coordinator",
			DisplayName: "Coordinator",
			Description: "Coordinates team",
			Model: protocol.ModelConfig{
				Provider: "mock",
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
		{
			Name:        "writer",
			DisplayName: "Writer",
			Description: "Documentation writer",
			Model: protocol.ModelConfig{
				Provider: "mock",
			},
		},
	}

	_ = pool.Load(definitions)
	return pool
}

func TestAssembleTeamTool_Name(t *testing.T) {
	pool := createTestPool()
	tool := NewAssembleTeamTool(pool, nil)
	assert.Equal(t, "assemble_team", tool.Name())
}

func TestAssembleTeamTool_Description(t *testing.T) {
	pool := createTestPool()
	tool := NewAssembleTeamTool(pool, nil)
	assert.NotEmpty(t, tool.Description())
}

func TestAssembleTeamTool_Parameters(t *testing.T) {
	pool := createTestPool()
	tool := NewAssembleTeamTool(pool, nil)
	params := tool.Parameters()
	assert.NotNil(t, params)

	// Parse schema
	var schema map[string]interface{}
	err := json.Unmarshal(params, &schema)
	require.NoError(t, err)

	props := schema["properties"].(map[string]interface{})
	assert.Contains(t, props, "agents")
	assert.Contains(t, props, "reason")
}

func TestAssembleTeamTool_ExecuteSuccess(t *testing.T) {
	pool := createTestPool()

	var assembledAgents []string
	var assembledReason string

	onAssemble := func(agents []string, reason string) error {
		assembledAgents = agents
		assembledReason = reason
		return nil
	}

	tool := NewAssembleTeamTool(pool, onAssemble)

	input := AssembleTeamInput{
		Agents: []string{"coordinator", "developer", "architect"},
		Reason: "Need full team for complex task",
	}
	inputJSON, _ := json.Marshal(input)

	ctx := context.Background()
	result, err := tool.Execute(ctx, inputJSON)
	require.NoError(t, err)

	// Check callback was called
	assert.Equal(t, []string{"coordinator", "developer", "architect"}, assembledAgents)
	assert.Equal(t, "Need full team for complex task", assembledReason)

	// Check result
	var output AssembleTeamOutput
	err = json.Unmarshal(result, &output)
	require.NoError(t, err)
	assert.True(t, output.Success)
	assert.Equal(t, []string{"coordinator", "developer", "architect"}, output.Team)
	assert.Contains(t, output.Message, "Team assembled")
}

func TestAssembleTeamTool_ExecuteInvalidAgent(t *testing.T) {
	pool := createTestPool()
	tool := NewAssembleTeamTool(pool, nil)

	input := AssembleTeamInput{
		Agents: []string{"coordinator", "nonexistent", "developer"},
		Reason: "Testing invalid agent",
	}
	inputJSON, _ := json.Marshal(input)

	ctx := context.Background()
	result, err := tool.Execute(ctx, inputJSON)
	require.NoError(t, err)

	// Should return unsuccessful result, not error
	var output AssembleTeamOutput
	err = json.Unmarshal(result, &output)
	require.NoError(t, err)
	assert.False(t, output.Success)
	assert.Contains(t, output.Message, "Team assembly failed")
	assert.Contains(t, output.Message, "nonexistent")
	assert.Contains(t, output.Message, "available agents")
}

func TestAssembleTeamTool_ExecuteEmptyAgents(t *testing.T) {
	pool := createTestPool()
	tool := NewAssembleTeamTool(pool, nil)

	input := AssembleTeamInput{
		Agents: []string{},
		Reason: "Empty team",
	}
	inputJSON, _ := json.Marshal(input)

	ctx := context.Background()
	_, err := tool.Execute(ctx, inputJSON)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one agent")
}

func TestAssembleTeamTool_ExecuteMissingReason(t *testing.T) {
	pool := createTestPool()
	tool := NewAssembleTeamTool(pool, nil)

	input := AssembleTeamInput{
		Agents: []string{"developer"},
		Reason: "",
	}
	inputJSON, _ := json.Marshal(input)

	ctx := context.Background()
	_, err := tool.Execute(ctx, inputJSON)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reason")
}

func TestAssembleTeamTool_ExecutionLocation(t *testing.T) {
	pool := createTestPool()
	tool := NewAssembleTeamTool(pool, nil)
	assert.Equal(t, protocol.ExecutionLocationServer, tool.ExecutionLocation())
}

func TestAssembleTeamTool_ExecuteWithSuggestion(t *testing.T) {
	pool := createTestPool()
	tool := NewAssembleTeamTool(pool, nil)

	// Try to use "documentation" agent which should suggest "writer"
	input := AssembleTeamInput{
		Agents: []string{"developer", "documentation"},
		Reason: "Need developer and documentation support",
	}
	inputJSON, _ := json.Marshal(input)

	ctx := context.Background()
	result, err := tool.Execute(ctx, inputJSON)
	require.NoError(t, err)

	// Should return unsuccessful result with suggestion
	var output AssembleTeamOutput
	err = json.Unmarshal(result, &output)
	require.NoError(t, err)
	assert.False(t, output.Success)
	assert.Contains(t, output.Message, "Team assembly failed")
	assert.Contains(t, output.Message, "documentation")
	// Should suggest "writer" as alternative
	assert.Contains(t, output.Message, "writer")
	assert.Contains(t, output.Message, "did you mean")
}
