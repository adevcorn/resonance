package capability

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock capability for testing
type mockCapability struct {
	name     string
	location protocol.ExecutionLocation
	executed bool
}

func (m *mockCapability) Name() string {
	return m.name
}

func (m *mockCapability) ExecutionLocation() protocol.ExecutionLocation {
	return m.location
}

func (m *mockCapability) Execute(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	m.executed = true
	result := map[string]interface{}{
		"success": true,
		"params":  string(params),
	}
	return json.Marshal(result)
}

func TestNewRegistry(t *testing.T) {
	reg := NewRegistry()
	assert.NotNil(t, reg)
	assert.NotNil(t, reg.capabilities)
}

func TestRegister(t *testing.T) {
	reg := NewRegistry()
	mock := &mockCapability{name: "test_capability", location: protocol.ExecutionLocationServer}

	err := reg.Register(mock)
	require.NoError(t, err)

	// Test duplicate registration
	err = reg.Register(mock)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestHas(t *testing.T) {
	reg := NewRegistry()
	mock := &mockCapability{name: "test_capability", location: protocol.ExecutionLocationServer}

	assert.False(t, reg.Has("test_capability"))

	err := reg.Register(mock)
	require.NoError(t, err)

	assert.True(t, reg.Has("test_capability"))
	assert.False(t, reg.Has("nonexistent"))
}

func TestGet(t *testing.T) {
	reg := NewRegistry()
	mock := &mockCapability{name: "test_capability", location: protocol.ExecutionLocationServer}

	// Test getting non-existent capability
	cap, err := reg.Get("test_capability")
	assert.Error(t, err)
	assert.Nil(t, cap)

	// Register and get
	err = reg.Register(mock)
	require.NoError(t, err)

	cap, err = reg.Get("test_capability")
	assert.NoError(t, err)
	assert.NotNil(t, cap)
	assert.Equal(t, "test_capability", cap.Name())
}

func TestExecute(t *testing.T) {
	reg := NewRegistry()
	mock := &mockCapability{name: "test_capability", location: protocol.ExecutionLocationServer}

	err := reg.Register(mock)
	require.NoError(t, err)

	params := json.RawMessage(`{"key": "value"}`)
	result, err := reg.Execute(context.Background(), "test_capability", params)
	require.NoError(t, err)
	assert.True(t, mock.executed)

	var resultMap map[string]interface{}
	err = json.Unmarshal(result, &resultMap)
	require.NoError(t, err)
	assert.True(t, resultMap["success"].(bool))

	// Test executing non-existent capability
	_, err = reg.Execute(context.Background(), "nonexistent", params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestList(t *testing.T) {
	reg := NewRegistry()

	// Empty registry
	list := reg.List()
	assert.Empty(t, list)

	// Add capabilities
	cap1 := &mockCapability{name: "cap1", location: protocol.ExecutionLocationServer}
	cap2 := &mockCapability{name: "cap2", location: protocol.ExecutionLocationClient}

	err := reg.Register(cap1)
	require.NoError(t, err)
	err = reg.Register(cap2)
	require.NoError(t, err)

	list = reg.List()
	assert.Len(t, list, 2)
	assert.Contains(t, list, "cap1")
	assert.Contains(t, list, "cap2")
}

func TestCount(t *testing.T) {
	reg := NewRegistry()
	assert.Equal(t, 0, reg.Count())

	cap1 := &mockCapability{name: "cap1", location: protocol.ExecutionLocationServer}
	cap2 := &mockCapability{name: "cap2", location: protocol.ExecutionLocationClient}

	err := reg.Register(cap1)
	require.NoError(t, err)
	assert.Equal(t, 1, reg.Count())

	err = reg.Register(cap2)
	require.NoError(t, err)
	assert.Equal(t, 2, reg.Count())
}
