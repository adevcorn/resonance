package orchestration

import (
	"context"
	"testing"
	"time"

	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/adevcorn/ensemble/internal/server/agent"
	"github.com/adevcorn/ensemble/internal/server/provider/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createCoordinatorAgent() *agent.Agent {
	mockProvider := mock.NewMockProvider("mock", []string{})
	def := &protocol.AgentDefinition{
		Name:         "coordinator",
		DisplayName:  "Coordinator",
		SystemPrompt: "You are a coordinator",
		Model: protocol.ModelConfig{
			Provider: "mock",
		},
	}
	return agent.NewAgent(def, mockProvider)
}

func TestNewModerator(t *testing.T) {
	coord := createCoordinatorAgent()
	mod := NewModerator(coord)
	assert.NotNil(t, mod)
}

func TestModerator_SelectNextAgentFirstTurn(t *testing.T) {
	coord := createCoordinatorAgent()
	mod := NewModerator(coord)

	ctx := context.Background()
	team := []string{"coordinator", "developer"}
	messages := []protocol.Message{}

	nextAgent, err := mod.SelectNextAgent(ctx, team, messages, "Test task")
	require.NoError(t, err)

	// Should select coordinator for first turn
	assert.Equal(t, "coordinator", nextAgent)
}

func TestModerator_SelectNextAgentWithMessages(t *testing.T) {
	mockProvider := mock.NewMockProvider("mock", []string{"developer"})
	def := &protocol.AgentDefinition{
		Name:         "coordinator",
		DisplayName:  "Coordinator",
		SystemPrompt: "You are a coordinator",
		Model: protocol.ModelConfig{
			Provider: "mock",
		},
	}
	coord := agent.NewAgent(def, mockProvider)
	mod := NewModerator(coord)

	ctx := context.Background()
	team := []string{"coordinator", "developer", "architect"}
	messages := []protocol.Message{
		{
			Role:      protocol.MessageRoleUser,
			Content:   "Build a feature",
			Timestamp: time.Now(),
		},
		{
			Role:      protocol.MessageRoleAssistant,
			Agent:     "coordinator",
			Content:   "Let's start",
			Timestamp: time.Now(),
		},
	}

	nextAgent, err := mod.SelectNextAgent(ctx, team, messages, "Build a feature")
	require.NoError(t, err)

	// Should get a valid team member or "complete"
	assert.NotEmpty(t, nextAgent)
}

func TestModerator_ShouldContinue(t *testing.T) {
	coord := createCoordinatorAgent()
	mod := NewModerator(coord)

	// With few messages, should continue
	messages := []protocol.Message{
		{Role: protocol.MessageRoleUser, Content: "Test"},
		{Role: protocol.MessageRoleAssistant, Content: "Response"},
	}
	assert.True(t, mod.ShouldContinue(messages))

	// With many messages, should stop
	manyMessages := make([]protocol.Message, 60)
	for i := range manyMessages {
		manyMessages[i] = protocol.Message{
			Role:    protocol.MessageRoleAssistant,
			Content: "Message",
		}
	}
	assert.False(t, mod.ShouldContinue(manyMessages))
}

func TestModerator_IsTaskComplete(t *testing.T) {
	coord := createCoordinatorAgent()
	mod := NewModerator(coord)

	// No completion signal
	messages := []protocol.Message{
		{
			Role:    protocol.MessageRoleAssistant,
			Content: "Working on it",
		},
	}
	assert.False(t, mod.isTaskComplete(messages))

	// With completion keyword in content
	messagesComplete := []protocol.Message{
		{
			Role:    protocol.MessageRoleAssistant,
			Content: "Task is complete",
		},
	}
	assert.True(t, mod.isTaskComplete(messagesComplete))
}
