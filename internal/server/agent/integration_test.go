package agent

import (
	"testing"

	"github.com/adevcorn/ensemble/internal/server/provider"
	"github.com/adevcorn/ensemble/internal/server/provider/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadAllDefaultAgents is an integration test that verifies all 9 default
// agent definitions can be loaded successfully from the agents/ directory
func TestLoadAllDefaultAgents(t *testing.T) {
	// Use the actual agents directory
	agentsPath := "../../../agents"

	// Create loader
	loader := NewLoader(agentsPath)

	// Load all agents
	definitions, err := loader.LoadAll()
	require.NoError(t, err, "Failed to load agent definitions")

	// Verify we have 9 agents
	assert.Len(t, definitions, 9, "Expected 9 default agents")

	// Create a map of expected agents
	expectedAgents := map[string]bool{
		"coordinator": false,
		"developer":   false,
		"architect":   false,
		"reviewer":    false,
		"researcher":  false,
		"security":    false,
		"writer":      false,
		"tester":      false,
		"devops":      false,
	}

	// Verify each agent definition
	for _, def := range definitions {
		// Mark agent as found
		if _, ok := expectedAgents[def.Name]; ok {
			expectedAgents[def.Name] = true
		}

		// Validate each agent has required fields
		assert.NotEmpty(t, def.Name, "Agent name should not be empty")
		assert.NotEmpty(t, def.DisplayName, "Agent display_name should not be empty")
		assert.NotEmpty(t, def.Description, "Agent description should not be empty")
		assert.NotEmpty(t, def.SystemPrompt, "Agent system_prompt should not be empty")
		assert.NotEmpty(t, def.Capabilities, "Agent should have at least one capability")
		assert.NotEmpty(t, def.Model.Provider, "Agent model provider should not be empty")
		assert.NotEmpty(t, def.Model.Name, "Agent model name should not be empty")
		assert.Greater(t, def.Model.MaxTokens, 0, "Agent max_tokens should be positive")
		assert.GreaterOrEqual(t, def.Model.Temperature, 0.0, "Temperature should be >= 0")
		assert.LessOrEqual(t, def.Model.Temperature, 2.0, "Temperature should be <= 2")

		// Validate system prompt length (should be comprehensive)
		assert.Greater(t, len(def.SystemPrompt), 200,
			"Agent %s: system_prompt should be comprehensive (at least 200 chars)", def.Name)

		// Log agent info
		t.Logf("✓ Agent: %s (%s) - %d capabilities, temp=%.1f, model=%s",
			def.Name, def.DisplayName, len(def.Capabilities),
			def.Model.Temperature, def.Model.Name)
	}

	// Verify all expected agents were found
	for agentName, found := range expectedAgents {
		assert.True(t, found, "Expected agent %s was not found", agentName)
	}
}

// TestLoadAndPoolIntegration tests loading all agents into a pool
func TestLoadAndPoolIntegration(t *testing.T) {
	// Use the actual agents directory
	agentsPath := "../../../agents"

	// Create loader
	loader := NewLoader(agentsPath)

	// Load all agents
	definitions, err := loader.LoadAll()
	require.NoError(t, err)

	// Create provider registry with mock provider
	registry := provider.NewRegistry()

	// Register mock providers for all provider types used by agents
	mockAnthropic := mock.NewMockProvider("anthropic", []string{"test response"})
	mockOpenAI := mock.NewMockProvider("openai", []string{"test response"})
	registry.Register(mockAnthropic)
	registry.Register(mockOpenAI)

	// Create pool and load agents
	pool := NewPool(registry)
	err = pool.Load(definitions)
	require.NoError(t, err, "Failed to load agents into pool")

	// Verify pool has all agents
	assert.Equal(t, 9, pool.Count(), "Pool should contain 9 agents")

	// Test retrieving each agent
	expectedAgents := []string{
		"coordinator", "developer", "architect", "reviewer",
		"researcher", "security", "writer", "tester", "devops",
	}

	for _, name := range expectedAgents {
		agent, err := pool.Get(name)
		require.NoError(t, err, "Failed to get agent %s", name)
		assert.NotNil(t, agent)
		assert.Equal(t, name, agent.Name())
		t.Logf("✓ Successfully loaded agent: %s", agent.DisplayName())
	}

	// Test listing agents
	names := pool.List()
	assert.Len(t, names, 9)
	assert.ElementsMatch(t, expectedAgents, names)
}

// TestAgentToolPermissions verifies that each agent has appropriate tool permissions
func TestAgentToolPermissions(t *testing.T) {
	agentsPath := "../../../agents"
	loader := NewLoader(agentsPath)
	definitions, err := loader.LoadAll()
	require.NoError(t, err)

	for _, def := range definitions {
		t.Run(def.Name, func(t *testing.T) {
			// All agents should have the collaborate tool
			assert.Contains(t, def.Tools.Allowed, "collaborate",
				"Agent %s should have collaborate tool", def.Name)

			// Coordinator should have assemble_team
			if def.Name == "coordinator" {
				assert.Contains(t, def.Tools.Allowed, "assemble_team",
					"Coordinator should have assemble_team tool")
			}

			// Developer and tester should have execute_command
			if def.Name == "developer" || def.Name == "tester" || def.Name == "devops" {
				assert.Contains(t, def.Tools.Allowed, "execute_command",
					"Agent %s should have execute_command tool", def.Name)
			}

			// Researcher should have web_search and fetch_url
			if def.Name == "researcher" {
				assert.Contains(t, def.Tools.Allowed, "web_search",
					"Researcher should have web_search tool")
				assert.Contains(t, def.Tools.Allowed, "fetch_url",
					"Researcher should have fetch_url tool")
			}

			// Agents that don't write code shouldn't have write_file
			readOnlyAgents := []string{"architect", "reviewer", "security"}
			for _, readOnly := range readOnlyAgents {
				if def.Name == readOnly {
					assert.NotContains(t, def.Tools.Allowed, "write_file",
						"Agent %s should not have write_file tool", def.Name)
				}
			}
		})
	}
}

// TestAgentTemperatureSettings verifies appropriate temperature settings
func TestAgentTemperatureSettings(t *testing.T) {
	agentsPath := "../../../agents"
	loader := NewLoader(agentsPath)
	definitions, err := loader.LoadAll()
	require.NoError(t, err)

	tempRanges := map[string]struct{ min, max float64 }{
		"coordinator": {0.4, 0.6}, // Creative coordination
		"developer":   {0.2, 0.4}, // Deterministic code
		"architect":   {0.3, 0.5}, // Balanced design
		"reviewer":    {0.1, 0.3}, // Strict review
		"researcher":  {0.5, 0.7}, // Creative research
		"security":    {0.1, 0.3}, // Thorough analysis
		"writer":      {0.4, 0.6}, // Clear documentation
		"tester":      {0.2, 0.4}, // Precise tests
		"devops":      {0.2, 0.4}, // Reliable infrastructure
	}

	for _, def := range definitions {
		expected, ok := tempRanges[def.Name]
		if !ok {
			continue
		}

		assert.GreaterOrEqual(t, def.Model.Temperature, expected.min,
			"Agent %s temperature too low", def.Name)
		assert.LessOrEqual(t, def.Model.Temperature, expected.max,
			"Agent %s temperature too high", def.Name)

		t.Logf("✓ Agent %s temperature: %.1f (expected %.1f-%.1f)",
			def.Name, def.Model.Temperature, expected.min, expected.max)
	}
}
