package tool

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/adevcorn/ensemble/internal/server/agent"
	"github.com/adevcorn/ensemble/internal/server/provider/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	assert.NotNil(t, r)
	assert.Equal(t, 0, r.Count())
}

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()

	// Create a simple tool
	tool := NewFunc("test", "Test tool", []byte(`{}`), protocol.ExecutionLocationServer,
		func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
			return []byte(`{}`), nil
		})

	// Register tool
	err := r.Register(tool)
	require.NoError(t, err)
	assert.Equal(t, 1, r.Count())

	// Attempt to register same tool again
	err = r.Register(tool)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestRegistry_RegisterNil(t *testing.T) {
	r := NewRegistry()
	err := r.Register(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil tool")
}

func TestRegistry_Get(t *testing.T) {
	r := NewRegistry()

	tool := NewFunc("test", "Test tool", []byte(`{}`), protocol.ExecutionLocationServer,
		func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
			return []byte(`{}`), nil
		})

	_ = r.Register(tool)

	// Get existing tool
	retrieved, err := r.Get("test")
	require.NoError(t, err)
	assert.Equal(t, "test", retrieved.Name())

	// Get non-existent tool
	_, err = r.Get("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRegistry_Has(t *testing.T) {
	r := NewRegistry()

	tool := NewFunc("test", "Test tool", []byte(`{}`), protocol.ExecutionLocationServer,
		func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
			return []byte(`{}`), nil
		})

	_ = r.Register(tool)

	assert.True(t, r.Has("test"))
	assert.False(t, r.Has("nonexistent"))
}

func TestRegistry_List(t *testing.T) {
	r := NewRegistry()

	// Register multiple tools
	tools := []string{"alice", "bob", "charlie"}
	for _, name := range tools {
		tool := NewFunc(name, "Tool "+name, []byte(`{}`), protocol.ExecutionLocationServer,
			func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				return []byte(`{}`), nil
			})
		_ = r.Register(tool)
	}

	list := r.List()
	assert.Equal(t, 3, len(list))
	// List should be sorted alphabetically
	assert.Equal(t, []string{"alice", "bob", "charlie"}, list)
}

func TestRegistry_GetAllowed(t *testing.T) {
	r := NewRegistry()

	// Register tools
	tool1 := NewFunc("read_file", "Read a file", []byte(`{}`), protocol.ExecutionLocationClient, nil)
	tool2 := NewFunc("write_file", "Write a file", []byte(`{}`), protocol.ExecutionLocationClient, nil)
	tool3 := NewFunc("web_search", "Search the web", []byte(`{}`), protocol.ExecutionLocationServer, nil)

	_ = r.Register(tool1)
	_ = r.Register(tool2)
	_ = r.Register(tool3)

	// Create agent with specific tools allowed
	mockProvider := mock.NewMockProvider("mock", []string{})
	agentDef := &protocol.AgentDefinition{
		Name: "test_agent",
		Tools: protocol.ToolsConfig{
			Allowed: []string{"read_file", "web_search"},
			Denied:  []string{},
		},
	}
	ag := agent.NewAgent(agentDef, mockProvider)

	// Get allowed tools for agent
	allowed := r.GetAllowed(ag)
	assert.Equal(t, 2, len(allowed))

	toolNames := make(map[string]bool)
	for _, tool := range allowed {
		toolNames[tool.Name] = true
	}

	assert.True(t, toolNames["read_file"])
	assert.True(t, toolNames["web_search"])
	assert.False(t, toolNames["write_file"])
}

func TestRegistry_GetAllowedWithDenied(t *testing.T) {
	r := NewRegistry()

	// Register tools
	tool1 := NewFunc("read_file", "Read a file", []byte(`{}`), protocol.ExecutionLocationClient, nil)
	tool2 := NewFunc("write_file", "Write a file", []byte(`{}`), protocol.ExecutionLocationClient, nil)

	_ = r.Register(tool1)
	_ = r.Register(tool2)

	// Create agent with write_file denied
	mockProvider := mock.NewMockProvider("mock", []string{})
	agentDef := &protocol.AgentDefinition{
		Name: "test_agent",
		Tools: protocol.ToolsConfig{
			Allowed: []string{"read_file", "write_file"},
			Denied:  []string{"write_file"},
		},
	}
	ag := agent.NewAgent(agentDef, mockProvider)

	// Get allowed tools for agent
	allowed := r.GetAllowed(ag)
	assert.Equal(t, 1, len(allowed))
	assert.Equal(t, "read_file", allowed[0].Name)
}
