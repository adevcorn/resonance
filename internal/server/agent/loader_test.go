package agent

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoaderLoadOne(t *testing.T) {
	// Create a temporary directory for test agent files
	tmpDir, err := os.MkdirTemp("", "agent-loader-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a valid agent definition
	agentYAML := `name: test-agent
display_name: "Test Agent"
description: "A test agent"
system_prompt: "You are a test agent"
capabilities:
  - testing
  - debugging
model:
  provider: anthropic
  name: claude-sonnet-4-20250514
  temperature: 0.5
  max_tokens: 1000
tools:
  allowed:
    - read_file
    - write_file
  denied: []
`

	err = os.WriteFile(filepath.Join(tmpDir, "test-agent.yaml"), []byte(agentYAML), 0644)
	require.NoError(t, err)

	// Create loader and load the agent
	loader := NewLoader(tmpDir)
	def, err := loader.LoadOne("test-agent.yaml")

	require.NoError(t, err)
	assert.NotNil(t, def)
	assert.Equal(t, "test-agent", def.Name)
	assert.Equal(t, "Test Agent", def.DisplayName)
	assert.Equal(t, "A test agent", def.Description)
	assert.Equal(t, "You are a test agent", def.SystemPrompt)
	assert.Equal(t, []string{"testing", "debugging"}, def.Capabilities)
	assert.Equal(t, "anthropic", def.Model.Provider)
	assert.Equal(t, "claude-sonnet-4-20250514", def.Model.Name)
	assert.Equal(t, 0.5, def.Model.Temperature)
	assert.Equal(t, 1000, def.Model.MaxTokens)
	assert.Equal(t, []string{"read_file", "write_file"}, def.Tools.Allowed)
}

func TestLoaderLoadOneMalformedYAML(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agent-loader-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create malformed YAML
	malformedYAML := `name: test-agent
display_name: "Test Agent
description: invalid yaml [[[
`

	err = os.WriteFile(filepath.Join(tmpDir, "malformed.yaml"), []byte(malformedYAML), 0644)
	require.NoError(t, err)

	loader := NewLoader(tmpDir)
	_, err = loader.LoadOne("malformed.yaml")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse YAML")
}

func TestLoaderLoadOneFileNotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agent-loader-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	loader := NewLoader(tmpDir)
	_, err = loader.LoadOne("nonexistent.yaml")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read file")
}

func TestLoaderLoadAll(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agent-loader-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create multiple valid agent files
	agent1 := `name: agent1
display_name: "Agent 1"
description: "First agent"
system_prompt: "You are agent 1"
capabilities: [testing]
model:
  provider: anthropic
  name: claude-sonnet-4-20250514
  temperature: 0.5
  max_tokens: 1000
tools:
  allowed: [read_file]
  denied: []
`

	agent2 := `name: agent2
display_name: "Agent 2"
description: "Second agent"
system_prompt: "You are agent 2"
capabilities: [debugging]
model:
  provider: openai
  name: gpt-4
  temperature: 0.7
  max_tokens: 2000
tools:
  allowed: [write_file]
  denied: []
`

	err = os.WriteFile(filepath.Join(tmpDir, "agent1.yaml"), []byte(agent1), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "agent2.yaml"), []byte(agent2), 0644)
	require.NoError(t, err)

	// Also create a non-YAML file that should be ignored
	err = os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("ignore me"), 0644)
	require.NoError(t, err)

	loader := NewLoader(tmpDir)
	definitions, err := loader.LoadAll()

	require.NoError(t, err)
	assert.Len(t, definitions, 2)

	// Find agents by name (order not guaranteed)
	agentsByName := make(map[string]*protocol.AgentDefinition)
	for _, def := range definitions {
		agentsByName[def.Name] = def
	}

	assert.Contains(t, agentsByName, "agent1")
	assert.Contains(t, agentsByName, "agent2")
	assert.Equal(t, "Agent 1", agentsByName["agent1"].DisplayName)
	assert.Equal(t, "Agent 2", agentsByName["agent2"].DisplayName)
}

func TestLoaderLoadAllWithInvalidFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agent-loader-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create one valid and one invalid agent
	validAgent := `name: valid
display_name: "Valid Agent"
description: "Valid"
system_prompt: "You are valid"
capabilities: [testing]
model:
  provider: anthropic
  name: claude-sonnet-4-20250514
  temperature: 0.5
  max_tokens: 1000
tools:
  allowed: [read_file]
  denied: []
`

	invalidAgent := `name: invalid
# Missing required fields
`

	err = os.WriteFile(filepath.Join(tmpDir, "valid.yaml"), []byte(validAgent), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "invalid.yaml"), []byte(invalidAgent), 0644)
	require.NoError(t, err)

	loader := NewLoader(tmpDir)
	definitions, err := loader.LoadAll()

	// Should return error but also return successfully loaded agents
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load some agents")
	assert.Len(t, definitions, 1)
	assert.Equal(t, "valid", definitions[0].Name)
}

func TestValidateSuccess(t *testing.T) {
	loader := NewLoader("/tmp")

	def := &protocol.AgentDefinition{
		Name:         "test",
		DisplayName:  "Test Agent",
		SystemPrompt: "You are a test",
		Capabilities: []string{"testing"},
		Model: protocol.ModelConfig{
			Provider:    "anthropic",
			Name:        "claude-sonnet-4-20250514",
			Temperature: 0.5,
			MaxTokens:   1000,
		},
	}

	err := loader.Validate(def)
	assert.NoError(t, err)
}

func TestValidateFailures(t *testing.T) {
	tests := []struct {
		name        string
		def         *protocol.AgentDefinition
		expectedErr string
	}{
		{
			name: "missing name",
			def: &protocol.AgentDefinition{
				DisplayName:  "Test",
				SystemPrompt: "Test",
				Capabilities: []string{"test"},
				Model: protocol.ModelConfig{
					Provider:    "anthropic",
					Name:        "claude",
					Temperature: 0.5,
					MaxTokens:   1000,
				},
			},
			expectedErr: "name is required",
		},
		{
			name: "missing display_name",
			def: &protocol.AgentDefinition{
				Name:         "test",
				SystemPrompt: "Test",
				Capabilities: []string{"test"},
				Model: protocol.ModelConfig{
					Provider:    "anthropic",
					Name:        "claude",
					Temperature: 0.5,
					MaxTokens:   1000,
				},
			},
			expectedErr: "display_name is required",
		},
		{
			name: "missing system_prompt",
			def: &protocol.AgentDefinition{
				Name:         "test",
				DisplayName:  "Test",
				Capabilities: []string{"test"},
				Model: protocol.ModelConfig{
					Provider:    "anthropic",
					Name:        "claude",
					Temperature: 0.5,
					MaxTokens:   1000,
				},
			},
			expectedErr: "system_prompt is required",
		},
		{
			name: "missing model provider",
			def: &protocol.AgentDefinition{
				Name:         "test",
				DisplayName:  "Test",
				SystemPrompt: "Test",
				Capabilities: []string{"test"},
				Model: protocol.ModelConfig{
					Name:        "claude",
					Temperature: 0.5,
					MaxTokens:   1000,
				},
			},
			expectedErr: "model.provider is required",
		},
		{
			name: "missing model name",
			def: &protocol.AgentDefinition{
				Name:         "test",
				DisplayName:  "Test",
				SystemPrompt: "Test",
				Capabilities: []string{"test"},
				Model: protocol.ModelConfig{
					Provider:    "anthropic",
					Temperature: 0.5,
					MaxTokens:   1000,
				},
			},
			expectedErr: "model.name is required",
		},
		{
			name: "temperature too low",
			def: &protocol.AgentDefinition{
				Name:         "test",
				DisplayName:  "Test",
				SystemPrompt: "Test",
				Capabilities: []string{"test"},
				Model: protocol.ModelConfig{
					Provider:    "anthropic",
					Name:        "claude",
					Temperature: -0.1,
					MaxTokens:   1000,
				},
			},
			expectedErr: "temperature must be between 0 and 2",
		},
		{
			name: "temperature too high",
			def: &protocol.AgentDefinition{
				Name:         "test",
				DisplayName:  "Test",
				SystemPrompt: "Test",
				Capabilities: []string{"test"},
				Model: protocol.ModelConfig{
					Provider:    "anthropic",
					Name:        "claude",
					Temperature: 2.1,
					MaxTokens:   1000,
				},
			},
			expectedErr: "temperature must be between 0 and 2",
		},
		{
			name: "max_tokens too low",
			def: &protocol.AgentDefinition{
				Name:         "test",
				DisplayName:  "Test",
				SystemPrompt: "Test",
				Capabilities: []string{"test"},
				Model: protocol.ModelConfig{
					Provider:    "anthropic",
					Name:        "claude",
					Temperature: 0.5,
					MaxTokens:   0,
				},
			},
			expectedErr: "max_tokens must be greater than 0",
		},
		{
			name: "no capabilities",
			def: &protocol.AgentDefinition{
				Name:         "test",
				DisplayName:  "Test",
				SystemPrompt: "Test",
				Capabilities: []string{},
				Model: protocol.ModelConfig{
					Provider:    "anthropic",
					Name:        "claude",
					Temperature: 0.5,
					MaxTokens:   1000,
				},
			},
			expectedErr: "at least one capability is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewLoader("/tmp")
			err := loader.Validate(tt.def)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}
