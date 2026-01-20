package tool

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/adevcorn/ensemble/internal/server/capability"
	"github.com/adevcorn/ensemble/internal/server/skill"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock capability for testing
type mockCapability struct {
	name     string
	executed bool
}

func (m *mockCapability) Name() string {
	return m.name
}

func (m *mockCapability) ExecutionLocation() protocol.ExecutionLocation {
	return protocol.ExecutionLocationServer
}

func (m *mockCapability) Execute(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	m.executed = true
	result := map[string]interface{}{
		"success": true,
		"params":  string(params),
	}
	return json.Marshal(result)
}

func setupTestActiveTool(t *testing.T) (*ActiveTool, *mockCapability) {
	// Create skill registry with test skills
	skillReg := &skill.Registry{}
	// Since we can't easily initialize the full registry, we'll use reflection/testing tricks
	// For now, let's just test what we can with the public API

	// Create capability registry
	capReg := capability.NewRegistry()
	mockCap := &mockCapability{name: "test_capability"}
	err := capReg.Register(mockCap)
	require.NoError(t, err)

	tool := NewActiveTool(skillReg, capReg)
	return tool, mockCap
}

func TestActiveTool_Name(t *testing.T) {
	tool, _ := setupTestActiveTool(t)
	assert.Equal(t, "active_tool", tool.Name())
}

func TestActiveTool_Description(t *testing.T) {
	tool, _ := setupTestActiveTool(t)
	desc := tool.Description()
	assert.Contains(t, desc, "search_skills")
	assert.Contains(t, desc, "load_skill")
	assert.Contains(t, desc, "execute")
}

func TestActiveTool_Parameters(t *testing.T) {
	tool, _ := setupTestActiveTool(t)
	params := tool.Parameters()

	var schema map[string]interface{}
	err := json.Unmarshal(params, &schema)
	require.NoError(t, err)

	assert.Equal(t, "object", schema["type"])

	properties, ok := schema["properties"].(map[string]interface{})
	require.True(t, ok)

	assert.Contains(t, properties, "action")
	assert.Contains(t, properties, "query")
	assert.Contains(t, properties, "skill_name")
	assert.Contains(t, properties, "capability")
	assert.Contains(t, properties, "parameters")
}

func TestActiveTool_ExecutionLocation(t *testing.T) {
	tool, _ := setupTestActiveTool(t)
	assert.Equal(t, protocol.ExecutionLocationServer, tool.ExecutionLocation())
}

func TestActiveTool_Execute_InvalidJSON(t *testing.T) {
	tool, _ := setupTestActiveTool(t)

	_, err := tool.Execute(context.Background(), json.RawMessage(`invalid json`))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid active_tool input")
}

func TestActiveTool_Execute_UnknownAction(t *testing.T) {
	tool, _ := setupTestActiveTool(t)

	input := ActiveToolInput{
		Action: "unknown_action",
	}
	inputJSON, _ := json.Marshal(input)

	_, err := tool.Execute(context.Background(), inputJSON)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown action")
}

func TestActiveTool_Execute_SearchSkills_MissingQuery(t *testing.T) {
	tool, _ := setupTestActiveTool(t)

	input := ActiveToolInput{
		Action: SearchSkills,
	}
	inputJSON, _ := json.Marshal(input)

	_, err := tool.Execute(context.Background(), inputJSON)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query is required")
}

func TestActiveTool_Execute_LoadSkill_MissingSkillName(t *testing.T) {
	tool, _ := setupTestActiveTool(t)

	input := ActiveToolInput{
		Action: LoadSkill,
	}
	inputJSON, _ := json.Marshal(input)

	_, err := tool.Execute(context.Background(), inputJSON)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "skill_name is required")
}

func TestActiveTool_Execute_Execute_MissingCapability(t *testing.T) {
	tool, _ := setupTestActiveTool(t)

	input := ActiveToolInput{
		Action: Execute,
	}
	inputJSON, _ := json.Marshal(input)

	_, err := tool.Execute(context.Background(), inputJSON)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "capability is required")
}

func TestActiveTool_Execute_Execute_CapabilityNotFound(t *testing.T) {
	tool, _ := setupTestActiveTool(t)

	input := ActiveToolInput{
		Action:     Execute,
		Capability: "nonexistent_capability",
		Parameters: json.RawMessage(`{}`),
	}
	inputJSON, _ := json.Marshal(input)

	_, err := tool.Execute(context.Background(), inputJSON)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestActiveTool_Execute_Execute_Success(t *testing.T) {
	tool, mockCap := setupTestActiveTool(t)

	input := ActiveToolInput{
		Action:     Execute,
		Capability: "test_capability",
		Parameters: json.RawMessage(`{"key": "value"}`),
	}
	inputJSON, _ := json.Marshal(input)

	result, err := tool.Execute(context.Background(), inputJSON)
	require.NoError(t, err)
	assert.True(t, mockCap.executed)

	var resultMap map[string]interface{}
	err = json.Unmarshal(result, &resultMap)
	require.NoError(t, err)
	assert.True(t, resultMap["success"].(bool))
}

func TestActiveToolAction_Constants(t *testing.T) {
	assert.Equal(t, ActiveToolAction("search_skills"), SearchSkills)
	assert.Equal(t, ActiveToolAction("load_skill"), LoadSkill)
	assert.Equal(t, ActiveToolAction("execute"), Execute)
}

func TestActiveToolInput_JSONMarshaling(t *testing.T) {
	input := ActiveToolInput{
		Action:     SearchSkills,
		Query:      "test query",
		MaxResults: 5,
	}

	data, err := json.Marshal(input)
	require.NoError(t, err)

	var decoded ActiveToolInput
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, input.Action, decoded.Action)
	assert.Equal(t, input.Query, decoded.Query)
	assert.Equal(t, input.MaxResults, decoded.MaxResults)
}
