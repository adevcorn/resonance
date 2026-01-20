package provider

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testProvider is a simple test implementation of Provider
type testProvider struct {
	name          string
	supportsTools bool
}

func (p *testProvider) Name() string {
	return p.name
}

func (p *testProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	return &CompletionResponse{
		Content: "test response",
		Usage: Usage{
			InputTokens:  10,
			OutputTokens: 20,
			TotalTokens:  30,
		},
	}, nil
}

func (p *testProvider) Stream(ctx context.Context, req *CompletionRequest) (<-chan StreamEvent, error) {
	ch := make(chan StreamEvent, 1)
	go func() {
		defer close(ch)
		ch <- StreamEvent{
			Type:    StreamEventContent,
			Content: "test",
			Done:    true,
		}
	}()
	return ch, nil
}

func (p *testProvider) SupportsTools() bool {
	return p.supportsTools
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()

	provider := &testProvider{name: "test", supportsTools: true}
	registry.Register(provider)

	assert.True(t, registry.Has("test"))

	retrieved, err := registry.Get("test")
	require.NoError(t, err)
	assert.Equal(t, "test", retrieved.Name())
}

func TestRegistry_Get_NotFound(t *testing.T) {
	registry := NewRegistry()

	_, err := registry.Get("nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRegistry_List(t *testing.T) {
	registry := NewRegistry()

	provider1 := &testProvider{name: "provider1", supportsTools: true}
	provider2 := &testProvider{name: "provider2", supportsTools: true}

	registry.Register(provider1)
	registry.Register(provider2)

	names := registry.List()
	assert.Len(t, names, 2)
	assert.Contains(t, names, "provider1")
	assert.Contains(t, names, "provider2")
}

func TestRegistry_Has(t *testing.T) {
	registry := NewRegistry()

	provider := &testProvider{name: "test", supportsTools: true}
	registry.Register(provider)

	assert.True(t, registry.Has("test"))
	assert.False(t, registry.Has("nonexistent"))
}

func TestRegistry_MultipleProviders(t *testing.T) {
	registry := NewRegistry()

	// Register multiple providers
	for i := 0; i < 5; i++ {
		name := "provider" + string(rune('0'+i))
		provider := &testProvider{name: name, supportsTools: true}
		registry.Register(provider)
	}

	names := registry.List()
	assert.Len(t, names, 5)

	// Verify all can be retrieved
	for i := 0; i < 5; i++ {
		name := "provider" + string(rune('0'+i))
		p, err := registry.Get(name)
		require.NoError(t, err)
		assert.Equal(t, name, p.Name())
	}
}

func TestRegistry_Overwrite(t *testing.T) {
	registry := NewRegistry()

	// Register a provider
	provider1 := &testProvider{name: "test", supportsTools: true}
	registry.Register(provider1)

	// Register another provider with the same name
	provider2 := &testProvider{name: "test", supportsTools: false}
	registry.Register(provider2)

	// Should retrieve the second one
	retrieved, err := registry.Get("test")
	require.NoError(t, err)
	assert.Equal(t, provider2, retrieved)
	assert.False(t, retrieved.SupportsTools())
}
