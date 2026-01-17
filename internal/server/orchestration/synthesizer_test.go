package orchestration

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/adevcorn/ensemble/internal/server/agent"
	"github.com/adevcorn/ensemble/internal/server/provider/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSynthesizer(t *testing.T) {
	coord := createCoordinatorAgent()
	synth := NewSynthesizer(coord)
	assert.NotNil(t, synth)
}

func TestSynthesizer_Synthesize(t *testing.T) {
	mockProvider := mock.NewMockProvider("mock", []string{"Task completed successfully. Implemented authentication middleware and added comprehensive tests."})
	def := &protocol.AgentDefinition{
		Name:         "coordinator",
		DisplayName:  "Coordinator",
		SystemPrompt: "You are a coordinator",
		Model: protocol.ModelConfig{
			Provider: "mock",
		},
	}
	coord := agent.NewAgent(def, mockProvider)
	synth := NewSynthesizer(coord)

	messages := []protocol.Message{
		{
			Role:      protocol.MessageRoleUser,
			Content:   "Implement user authentication",
			Timestamp: time.Now(),
		},
		{
			Role:      protocol.MessageRoleAssistant,
			Agent:     "developer",
			Content:   "I've implemented the auth middleware",
			Timestamp: time.Now(),
		},
		{
			Role:      protocol.MessageRoleAssistant,
			Agent:     "tester",
			Content:   "I've added test cases for the auth flow",
			Timestamp: time.Now(),
		},
	}

	ctx := context.Background()
	summary, artifacts, err := synth.Synthesize(ctx, "Implement user authentication", messages)

	require.NoError(t, err)
	assert.NotEmpty(t, summary)
	assert.NotNil(t, artifacts)
}

func TestSynthesizer_SynthesizeExtractsArtifacts(t *testing.T) {
	mockProvider := mock.NewMockProvider("mock", []string{"Successfully implemented authentication with comprehensive tests."})
	def := &protocol.AgentDefinition{
		Name:         "coordinator",
		DisplayName:  "Coordinator",
		SystemPrompt: "You are a coordinator",
		Model: protocol.ModelConfig{
			Provider: "mock",
		},
	}
	coord := agent.NewAgent(def, mockProvider)
	synth := NewSynthesizer(coord)

	// Create message with collaborate tool call containing artifacts
	collaborateInput := protocol.CollaborateInput{
		Action:    protocol.CollaborateBroadcast,
		Message:   "Work complete",
		Artifacts: []string{"src/auth.go", "src/auth_test.go"},
	}
	inputJSON, _ := json.Marshal(collaborateInput)

	messages := []protocol.Message{
		{
			Role:    protocol.MessageRoleUser,
			Content: "Implement auth",
		},
		{
			Role:    protocol.MessageRoleAssistant,
			Agent:   "developer",
			Content: "Done",
			ToolCalls: []protocol.ToolCall{
				{
					ID:        "call_1",
					ToolName:  "collaborate",
					Arguments: inputJSON,
				},
			},
		},
	}

	ctx := context.Background()
	_, artifacts, err := synth.Synthesize(ctx, "Implement auth", messages)

	require.NoError(t, err)
	assert.Contains(t, artifacts, "src/auth.go")
	assert.Contains(t, artifacts, "src/auth_test.go")
}
