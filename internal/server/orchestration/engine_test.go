package orchestration

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/adevcorn/ensemble/internal/server/agent"
	"github.com/adevcorn/ensemble/internal/server/provider"
	"github.com/adevcorn/ensemble/internal/server/provider/mock"
	"github.com/adevcorn/ensemble/internal/server/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestEngine() (*Engine, *agent.Pool, *tool.Registry, error) {
	// Create mock provider with tool support and enough responses for the test
	// The engine will call the mock multiple times: team assembly, each agent turn, moderator decisions, synthesis
	mockProvider := mock.NewMockProvider("mock", []string{
		"coordinator",                 // For team assembly
		"Let's implement the feature", // Coordinator response
		"developer",                   // Moderator decision
		"I'll write the code",         // Developer response
		"complete",                    // Moderator decision to complete
		"Summary of work completed",   // Synthesis
	})
	providerRegistry := provider.NewRegistry()
	providerRegistry.Register(mockProvider)

	// Create agent pool
	pool := agent.NewPool(providerRegistry)

	// Load test agents
	definitions := []*protocol.AgentDefinition{
		{
			Name:         "coordinator",
			DisplayName:  "Coordinator",
			Description:  "Coordinates multi-agent collaboration",
			SystemPrompt: "You are a coordinator",
			Model: protocol.ModelConfig{
				Provider:    "mock",
				Name:        "mock-model",
				Temperature: 0.7,
				MaxTokens:   2048,
			},
			Tools: protocol.ToolsConfig{
				Allowed: []string{"assemble_team", "collaborate"},
			},
		},
		{
			Name:         "developer",
			DisplayName:  "Developer",
			Description:  "Expert software developer",
			SystemPrompt: "You are a developer",
			Model: protocol.ModelConfig{
				Provider:    "mock",
				Name:        "mock-model",
				Temperature: 0.3,
				MaxTokens:   4096,
			},
			Tools: protocol.ToolsConfig{
				Allowed: []string{"collaborate", "read_file", "write_file"},
			},
		},
		{
			Name:         "architect",
			DisplayName:  "Architect",
			Description:  "System architect",
			SystemPrompt: "You are an architect",
			Model: protocol.ModelConfig{
				Provider:    "mock",
				Name:        "mock-model",
				Temperature: 0.5,
				MaxTokens:   2048,
			},
			Tools: protocol.ToolsConfig{
				Allowed: []string{"collaborate", "read_file"},
			},
		},
	}

	if err := pool.Load(definitions); err != nil {
		return nil, nil, nil, err
	}

	// Create tool registry
	toolRegistry := tool.NewRegistry()

	// Register collaborate tool
	collaborateTool := tool.NewCollaborateTool(pool, func(from string, input *protocol.CollaborateInput) error {
		// Just accept all collaborations in tests
		return nil
	})
	_ = toolRegistry.Register(collaborateTool)

	// Register assemble_team tool
	assembleTeamTool := tool.NewAssembleTeamTool(pool, func(agents []string, reason string) error {
		// Accept all team assemblies in tests
		return nil
	})
	_ = toolRegistry.Register(assembleTeamTool)

	// Callbacks for engine
	onMessage := func(msg protocol.Message) error {
		return nil
	}

	onToolCall := func(tc protocol.ToolCall) (protocol.ToolResult, error) {
		// Mock client tool results
		return protocol.ToolResult{
			CallID: tc.ID,
			Result: []byte(`{"success": true}`),
		}, nil
	}

	// Create engine
	engine, err := NewEngine(pool, toolRegistry, onMessage, onToolCall)
	if err != nil {
		return nil, nil, nil, err
	}

	return engine, pool, toolRegistry, nil
}

func TestNewEngine(t *testing.T) {
	engine, _, _, err := createTestEngine()
	require.NoError(t, err)
	assert.NotNil(t, engine)
	assert.NotNil(t, engine.coordinator)
	assert.NotNil(t, engine.moderator)
	assert.NotNil(t, engine.synthesizer)
}

func TestNewEngineMissingCoordinator(t *testing.T) {
	// Create pool without coordinator
	mockProvider := mock.NewMockProvider("mock", []string{})
	providerRegistry := provider.NewRegistry()
	providerRegistry.Register(mockProvider)

	pool := agent.NewPool(providerRegistry)
	toolRegistry := tool.NewRegistry()

	_, err := NewEngine(pool, toolRegistry, nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "coordinator")
}

func TestEngine_Run(t *testing.T) {
	t.Skip("Integration test - complex mock setup required")
	// This test requires careful orchestration of many mock responses
	// The engine calls: AnalyzeTask (Complete), agent turns (Stream), SelectNextAgent (Complete), Synthesize (Complete)
	// For now, we test individual components separately
}

func TestEngine_ExecuteToolCalls(t *testing.T) {
	engine, _, _, err := createTestEngine()
	require.NoError(t, err)

	ctx := context.Background()

	// Test server-side tool (collaborate)
	collaborateInput := protocol.CollaborateInput{
		Action:  protocol.CollaborateBroadcast,
		Message: "Hello team",
	}
	inputJSON, _ := json.Marshal(collaborateInput)

	toolCalls := []protocol.ToolCall{
		{
			ID:        "call_1",
			ToolName:  "collaborate",
			Arguments: inputJSON,
		},
	}

	results, err := engine.executeToolCalls(ctx, "developer", toolCalls)
	require.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "call_1", results[0].CallID)
	assert.Empty(t, results[0].Error)
}

func TestEngine_ExecuteToolCallsClientTool(t *testing.T) {
	// Create engine with onToolCall callback
	var capturedToolCall protocol.ToolCall

	engine, pool, toolRegistry, err := createTestEngine()
	require.NoError(t, err)

	// Replace onToolCall
	engine.onToolCall = func(tc protocol.ToolCall) (protocol.ToolResult, error) {
		capturedToolCall = tc
		return protocol.ToolResult{
			CallID: tc.ID,
			Result: []byte(`{"content": "file contents"}`),
		}, nil
	}

	ctx := context.Background()

	// Test client-side tool (read_file)
	readInput := map[string]string{"path": "test.go"}
	inputJSON, _ := json.Marshal(readInput)

	toolCalls := []protocol.ToolCall{
		{
			ID:        "call_2",
			ToolName:  "read_file",
			Arguments: inputJSON,
		},
	}

	results, err := engine.executeToolCalls(ctx, "developer", toolCalls)
	require.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "call_2", results[0].CallID)
	assert.Empty(t, results[0].Error)

	// Check callback was invoked
	assert.Equal(t, "read_file", capturedToolCall.ToolName)

	_ = pool
	_ = toolRegistry
}

func TestEngine_GenerateID(t *testing.T) {
	id1 := generateID()
	time.Sleep(10 * time.Millisecond) // Ensure different timestamp
	id2 := generateID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2) // IDs should be unique
}
