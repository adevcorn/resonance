package agent

import (
	"context"
	"testing"

	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/adevcorn/ensemble/internal/server/provider"
	"github.com/adevcorn/ensemble/internal/server/provider/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAgent(t *testing.T) {
	def := &protocol.AgentDefinition{
		Name:        "test-agent",
		DisplayName: "Test Agent",
		Model: protocol.ModelConfig{
			Provider:    "mock",
			Name:        "test-model",
			Temperature: 0.5,
			MaxTokens:   1000,
		},
	}
	prov := mock.NewMockProvider("mock", []string{"test response"})

	agent := NewAgent(def, prov)

	assert.NotNil(t, agent)
	assert.Equal(t, def, agent.definition)
	assert.Equal(t, prov, agent.provider)
}

func TestAgentGetters(t *testing.T) {
	def := &protocol.AgentDefinition{
		Name:         "developer",
		DisplayName:  "Developer Agent",
		Description:  "Expert developer",
		SystemPrompt: "You are a developer",
		Capabilities: []string{"coding", "debugging"},
		Model: protocol.ModelConfig{
			Provider:    "mock",
			Name:        "test-model",
			Temperature: 0.3,
			MaxTokens:   2000,
		},
		Tools: protocol.ToolsConfig{
			Allowed: []string{"read_file", "write_file"},
			Denied:  []string{"delete_file"},
		},
	}
	prov := mock.NewMockProvider("mock", []string{"test response"})
	agent := NewAgent(def, prov)

	assert.Equal(t, "developer", agent.Name())
	assert.Equal(t, "Developer Agent", agent.DisplayName())
	assert.Equal(t, "Expert developer", agent.Description())
	assert.Equal(t, "You are a developer", agent.SystemPrompt())
	assert.Equal(t, []string{"coding", "debugging"}, agent.Capabilities())
}

func TestAgentHasTool(t *testing.T) {
	def := &protocol.AgentDefinition{
		Name: "test-agent",
		Model: protocol.ModelConfig{
			Provider:    "mock",
			Name:        "test-model",
			Temperature: 0.5,
			MaxTokens:   1000,
		},
		Tools: protocol.ToolsConfig{
			Allowed: []string{"read_file", "write_file", "list_directory"},
		},
	}
	prov := mock.NewMockProvider("mock", []string{"test response"})
	agent := NewAgent(def, prov)

	assert.True(t, agent.HasTool("read_file"))
	assert.True(t, agent.HasTool("write_file"))
	assert.True(t, agent.HasTool("list_directory"))
	assert.False(t, agent.HasTool("execute_command"))
	assert.False(t, agent.HasTool("nonexistent_tool"))
}

func TestAgentIsToolAllowed(t *testing.T) {
	tests := []struct {
		name     string
		allowed  []string
		denied   []string
		tool     string
		expected bool
	}{
		{
			name:     "tool in allowed list",
			allowed:  []string{"read_file", "write_file"},
			denied:   []string{},
			tool:     "read_file",
			expected: true,
		},
		{
			name:     "tool not in allowed list",
			allowed:  []string{"read_file", "write_file"},
			denied:   []string{},
			tool:     "execute_command",
			expected: false,
		},
		{
			name:     "tool in denied list",
			allowed:  []string{"read_file", "write_file", "execute_command"},
			denied:   []string{"execute_command"},
			tool:     "execute_command",
			expected: false,
		},
		{
			name:     "tool in allowed but not denied",
			allowed:  []string{"read_file", "write_file"},
			denied:   []string{"execute_command"},
			tool:     "read_file",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def := &protocol.AgentDefinition{
				Name: "test-agent",
				Model: protocol.ModelConfig{
					Provider:    "mock",
					Name:        "test-model",
					Temperature: 0.5,
					MaxTokens:   1000,
				},
				Tools: protocol.ToolsConfig{
					Allowed: tt.allowed,
					Denied:  tt.denied,
				},
			}
			prov := mock.NewMockProvider("mock", []string{"test response"})
			agent := NewAgent(def, prov)

			result := agent.IsToolAllowed(tt.tool)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAgentDefinitionAndProvider(t *testing.T) {
	def := &protocol.AgentDefinition{
		Name: "test-agent",
		Model: protocol.ModelConfig{
			Provider:    "mock",
			Name:        "test-model",
			Temperature: 0.5,
			MaxTokens:   1000,
		},
	}
	prov := mock.NewMockProvider("mock", []string{"test response"})
	agent := NewAgent(def, prov)

	assert.Equal(t, def, agent.Definition())
	assert.Equal(t, prov, agent.Provider())
}

func TestAgentComplete(t *testing.T) {
	def := &protocol.AgentDefinition{
		Name: "test-agent",
		Model: protocol.ModelConfig{
			Provider:    "mock",
			Name:        "test-model",
			Temperature: 0.7,
			MaxTokens:   2000,
		},
	}
	prov := mock.NewMockProvider("mock", []string{"test response"})
	agent := NewAgent(def, prov)

	req := &provider.CompletionRequest{
		Messages: []protocol.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	// Test that agent uses its configured model settings
	ctx := context.Background()
	resp, err := agent.Complete(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test response", resp.Content)
	// The agent should populate the request with its model settings
	assert.Equal(t, "test-model", req.Model)
	assert.Equal(t, 0.7, req.Temperature)
	assert.Equal(t, 2000, req.MaxTokens)
}
