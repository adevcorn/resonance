package agent

import (
	"sync"
	"testing"

	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/adevcorn/ensemble/internal/server/provider"
	"github.com/adevcorn/ensemble/internal/server/provider/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPool(t *testing.T) {
	registry := provider.NewRegistry()
	pool := NewPool(registry)

	assert.NotNil(t, pool)
	assert.Equal(t, 0, pool.Count())
}

func TestPoolLoad(t *testing.T) {
	registry := provider.NewRegistry()
	mockProv := mock.NewMockProvider("mock", []string{"test"})
	registry.Register(mockProv)

	pool := NewPool(registry)

	definitions := []*protocol.AgentDefinition{
		{
			Name:        "agent1",
			DisplayName: "Agent 1",
			Model: protocol.ModelConfig{
				Provider:    "mock",
				Name:        "test-model",
				Temperature: 0.5,
				MaxTokens:   1000,
			},
		},
		{
			Name:        "agent2",
			DisplayName: "Agent 2",
			Model: protocol.ModelConfig{
				Provider:    "mock",
				Name:        "test-model",
				Temperature: 0.7,
				MaxTokens:   2000,
			},
		},
	}

	err := pool.Load(definitions)
	require.NoError(t, err)
	assert.Equal(t, 2, pool.Count())
}

func TestPoolLoadWithMissingProvider(t *testing.T) {
	registry := provider.NewRegistry()
	pool := NewPool(registry)

	definitions := []*protocol.AgentDefinition{
		{
			Name:        "agent1",
			DisplayName: "Agent 1",
			Model: protocol.ModelConfig{
				Provider:    "nonexistent",
				Name:        "test-model",
				Temperature: 0.5,
				MaxTokens:   1000,
			},
		},
	}

	err := pool.Load(definitions)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load some agents")
}

func TestPoolGet(t *testing.T) {
	registry := provider.NewRegistry()
	mockProv := mock.NewMockProvider("mock", []string{"test"})
	registry.Register(mockProv)

	pool := NewPool(registry)

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

	err := pool.Load([]*protocol.AgentDefinition{def})
	require.NoError(t, err)

	// Get existing agent
	agent, err := pool.Get("test-agent")
	require.NoError(t, err)
	assert.NotNil(t, agent)
	assert.Equal(t, "test-agent", agent.Name())
	assert.Equal(t, "Test Agent", agent.DisplayName())

	// Get non-existent agent
	_, err = pool.Get("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestPoolList(t *testing.T) {
	registry := provider.NewRegistry()
	mockProv := mock.NewMockProvider("mock", []string{"test"})
	registry.Register(mockProv)

	pool := NewPool(registry)

	definitions := []*protocol.AgentDefinition{
		{
			Name:        "developer",
			DisplayName: "Developer",
			Model: protocol.ModelConfig{
				Provider:    "mock",
				Name:        "test-model",
				Temperature: 0.5,
				MaxTokens:   1000,
			},
		},
		{
			Name:        "architect",
			DisplayName: "Architect",
			Model: protocol.ModelConfig{
				Provider:    "mock",
				Name:        "test-model",
				Temperature: 0.5,
				MaxTokens:   1000,
			},
		},
		{
			Name:        "reviewer",
			DisplayName: "Reviewer",
			Model: protocol.ModelConfig{
				Provider:    "mock",
				Name:        "test-model",
				Temperature: 0.5,
				MaxTokens:   1000,
			},
		},
	}

	err := pool.Load(definitions)
	require.NoError(t, err)

	names := pool.List()
	assert.Len(t, names, 3)
	// List should be sorted alphabetically
	assert.Equal(t, []string{"architect", "developer", "reviewer"}, names)
}

func TestPoolHas(t *testing.T) {
	registry := provider.NewRegistry()
	mockProv := mock.NewMockProvider("mock", []string{"test"})
	registry.Register(mockProv)

	pool := NewPool(registry)

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

	err := pool.Load([]*protocol.AgentDefinition{def})
	require.NoError(t, err)

	assert.True(t, pool.Has("test-agent"))
	assert.False(t, pool.Has("nonexistent"))
}

func TestPoolReload(t *testing.T) {
	registry := provider.NewRegistry()
	mockProv := mock.NewMockProvider("mock", []string{"test"})
	registry.Register(mockProv)

	pool := NewPool(registry)

	// Load initial agent
	def := &protocol.AgentDefinition{
		Name:        "test-agent",
		DisplayName: "Test Agent",
		Description: "Original description",
		Model: protocol.ModelConfig{
			Provider:    "mock",
			Name:        "test-model",
			Temperature: 0.5,
			MaxTokens:   1000,
		},
	}

	err := pool.Load([]*protocol.AgentDefinition{def})
	require.NoError(t, err)

	agent, err := pool.Get("test-agent")
	require.NoError(t, err)
	assert.Equal(t, "Original description", agent.Description())

	// Reload with updated definition
	updatedDef := &protocol.AgentDefinition{
		Name:        "test-agent",
		DisplayName: "Test Agent Updated",
		Description: "Updated description",
		Model: protocol.ModelConfig{
			Provider:    "mock",
			Name:        "test-model",
			Temperature: 0.7,
			MaxTokens:   2000,
		},
	}

	err = pool.Reload(updatedDef)
	require.NoError(t, err)

	// Get the reloaded agent
	agent, err = pool.Get("test-agent")
	require.NoError(t, err)
	assert.Equal(t, "Test Agent Updated", agent.DisplayName())
	assert.Equal(t, "Updated description", agent.Description())
}

func TestPoolReloadWithMissingProvider(t *testing.T) {
	registry := provider.NewRegistry()
	pool := NewPool(registry)

	def := &protocol.AgentDefinition{
		Name:        "test-agent",
		DisplayName: "Test Agent",
		Model: protocol.ModelConfig{
			Provider:    "nonexistent",
			Name:        "test-model",
			Temperature: 0.5,
			MaxTokens:   1000,
		},
	}

	err := pool.Reload(def)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get provider")
}

func TestPoolRemove(t *testing.T) {
	registry := provider.NewRegistry()
	mockProv := mock.NewMockProvider("mock", []string{"test"})
	registry.Register(mockProv)

	pool := NewPool(registry)

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

	err := pool.Load([]*protocol.AgentDefinition{def})
	require.NoError(t, err)
	assert.Equal(t, 1, pool.Count())

	// Remove existing agent
	err = pool.Remove("test-agent")
	require.NoError(t, err)
	assert.Equal(t, 0, pool.Count())
	assert.False(t, pool.Has("test-agent"))

	// Try to remove non-existent agent
	err = pool.Remove("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestPoolCount(t *testing.T) {
	registry := provider.NewRegistry()
	mockProv := mock.NewMockProvider("mock", []string{"test"})
	registry.Register(mockProv)

	pool := NewPool(registry)
	assert.Equal(t, 0, pool.Count())

	definitions := []*protocol.AgentDefinition{
		{
			Name:        "agent1",
			DisplayName: "Agent 1",
			Model: protocol.ModelConfig{
				Provider:    "mock",
				Name:        "test-model",
				Temperature: 0.5,
				MaxTokens:   1000,
			},
		},
		{
			Name:        "agent2",
			DisplayName: "Agent 2",
			Model: protocol.ModelConfig{
				Provider:    "mock",
				Name:        "test-model",
				Temperature: 0.5,
				MaxTokens:   1000,
			},
		},
	}

	err := pool.Load(definitions)
	require.NoError(t, err)
	assert.Equal(t, 2, pool.Count())

	err = pool.Remove("agent1")
	require.NoError(t, err)
	assert.Equal(t, 1, pool.Count())
}

func TestPoolGetAll(t *testing.T) {
	registry := provider.NewRegistry()
	mockProv := mock.NewMockProvider("mock", []string{"test"})
	registry.Register(mockProv)

	pool := NewPool(registry)

	definitions := []*protocol.AgentDefinition{
		{
			Name:        "agent1",
			DisplayName: "Agent 1",
			Model: protocol.ModelConfig{
				Provider:    "mock",
				Name:        "test-model",
				Temperature: 0.5,
				MaxTokens:   1000,
			},
		},
		{
			Name:        "agent2",
			DisplayName: "Agent 2",
			Model: protocol.ModelConfig{
				Provider:    "mock",
				Name:        "test-model",
				Temperature: 0.5,
				MaxTokens:   1000,
			},
		},
	}

	err := pool.Load(definitions)
	require.NoError(t, err)

	agents := pool.GetAll()
	assert.Len(t, agents, 2)

	// Verify all agents are present
	agentNames := make(map[string]bool)
	for _, agent := range agents {
		agentNames[agent.Name()] = true
	}
	assert.True(t, agentNames["agent1"])
	assert.True(t, agentNames["agent2"])
}

// TestPoolThreadSafety tests concurrent access to the pool
func TestPoolThreadSafety(t *testing.T) {
	registry := provider.NewRegistry()
	mockProv := mock.NewMockProvider("mock", []string{"test"})
	registry.Register(mockProv)

	pool := NewPool(registry)

	// Load initial agents
	definitions := []*protocol.AgentDefinition{
		{
			Name:        "agent1",
			DisplayName: "Agent 1",
			Model: protocol.ModelConfig{
				Provider:    "mock",
				Name:        "test-model",
				Temperature: 0.5,
				MaxTokens:   1000,
			},
		},
		{
			Name:        "agent2",
			DisplayName: "Agent 2",
			Model: protocol.ModelConfig{
				Provider:    "mock",
				Name:        "test-model",
				Temperature: 0.5,
				MaxTokens:   1000,
			},
		},
	}

	err := pool.Load(definitions)
	require.NoError(t, err)

	// Perform concurrent operations
	var wg sync.WaitGroup
	iterations := 100

	// Concurrent reads
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pool.Get("agent1")
			pool.List()
			pool.Has("agent2")
			pool.Count()
		}()
	}

	// Concurrent reloads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			def := &protocol.AgentDefinition{
				Name:        "agent1",
				DisplayName: "Agent 1",
				Description: "concurrent update",
				Model: protocol.ModelConfig{
					Provider:    "mock",
					Name:        "test-model",
					Temperature: 0.5,
					MaxTokens:   1000,
				},
			}
			pool.Reload(def)
		}(i)
	}

	wg.Wait()

	// Verify pool is still consistent
	assert.Equal(t, 2, pool.Count())
	agent, err := pool.Get("agent1")
	require.NoError(t, err)
	assert.NotNil(t, agent)
}
