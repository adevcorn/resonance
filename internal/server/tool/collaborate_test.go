package tool

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollaborateTool_Name(t *testing.T) {
	tool := NewCollaborateTool(nil)
	assert.Equal(t, "collaborate", tool.Name())
}

func TestCollaborateTool_Description(t *testing.T) {
	tool := NewCollaborateTool(nil)
	assert.NotEmpty(t, tool.Description())
}

func TestCollaborateTool_Parameters(t *testing.T) {
	tool := NewCollaborateTool(nil)
	params := tool.Parameters()
	assert.NotNil(t, params)

	// Parse schema to ensure it's valid JSON
	var schema map[string]interface{}
	err := json.Unmarshal(params, &schema)
	require.NoError(t, err)

	// Check required fields
	props := schema["properties"].(map[string]interface{})
	assert.Contains(t, props, "action")
	assert.Contains(t, props, "message")
	assert.Contains(t, props, "to_agent")
	assert.Contains(t, props, "artifacts")
}

func TestCollaborateTool_ExecuteBroadcast(t *testing.T) {
	var receivedFrom string
	var receivedInput *protocol.CollaborateInput

	onMessage := func(from string, input *protocol.CollaborateInput) error {
		receivedFrom = from
		receivedInput = input
		return nil
	}

	tool := NewCollaborateTool(onMessage)

	input := protocol.CollaborateInput{
		Action:  protocol.CollaborateBroadcast,
		Message: "Hello team!",
	}
	inputJSON, _ := json.Marshal(input)

	ctx := context.WithValue(context.Background(), "agent_name", "developer")
	result, err := tool.Execute(ctx, inputJSON)
	require.NoError(t, err)

	// Check callback was called
	assert.Equal(t, "developer", receivedFrom)
	assert.Equal(t, "Hello team!", receivedInput.Message)
	assert.Equal(t, protocol.CollaborateBroadcast, receivedInput.Action)

	// Check result
	var output protocol.CollaborateOutput
	err = json.Unmarshal(result, &output)
	require.NoError(t, err)
	assert.True(t, output.Delivered)
	assert.Contains(t, output.Recipients[0], "all team members")
}

func TestCollaborateTool_ExecuteDirect(t *testing.T) {
	var receivedInput *protocol.CollaborateInput

	onMessage := func(from string, input *protocol.CollaborateInput) error {
		receivedInput = input
		return nil
	}

	tool := NewCollaborateTool(onMessage)

	input := protocol.CollaborateInput{
		Action:  protocol.CollaborateDirect,
		Message: "Can you help with this?",
		ToAgent: "architect",
	}
	inputJSON, _ := json.Marshal(input)

	ctx := context.WithValue(context.Background(), "agent_name", "developer")
	result, err := tool.Execute(ctx, inputJSON)
	require.NoError(t, err)

	// Check callback was called
	assert.Equal(t, "architect", receivedInput.ToAgent)

	// Check result
	var output protocol.CollaborateOutput
	err = json.Unmarshal(result, &output)
	require.NoError(t, err)
	assert.True(t, output.Delivered)
	assert.Equal(t, []string{"architect"}, output.Recipients)
}

func TestCollaborateTool_ExecuteDirectMissingToAgent(t *testing.T) {
	tool := NewCollaborateTool(nil)

	input := protocol.CollaborateInput{
		Action:  protocol.CollaborateDirect,
		Message: "Can you help with this?",
		// ToAgent is missing
	}
	inputJSON, _ := json.Marshal(input)

	ctx := context.Background()
	_, err := tool.Execute(ctx, inputJSON)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "to_agent is required")
}

func TestCollaborateTool_ExecuteComplete(t *testing.T) {
	var receivedInput *protocol.CollaborateInput

	onMessage := func(from string, input *protocol.CollaborateInput) error {
		receivedInput = input
		return nil
	}

	tool := NewCollaborateTool(onMessage)

	input := protocol.CollaborateInput{
		Action:    protocol.CollaborateComplete,
		Message:   "Task completed successfully!",
		Artifacts: []string{"src/main.go", "src/main_test.go"},
	}
	inputJSON, _ := json.Marshal(input)

	ctx := context.WithValue(context.Background(), "agent_name", "developer")
	result, err := tool.Execute(ctx, inputJSON)
	require.NoError(t, err)

	// Check callback was called
	assert.Equal(t, protocol.CollaborateComplete, receivedInput.Action)
	assert.Equal(t, 2, len(receivedInput.Artifacts))

	// Check result
	var output protocol.CollaborateOutput
	err = json.Unmarshal(result, &output)
	require.NoError(t, err)
	assert.True(t, output.Delivered)
}

func TestCollaborateTool_ExecuteEmptyMessage(t *testing.T) {
	tool := NewCollaborateTool(nil)

	input := protocol.CollaborateInput{
		Action:  protocol.CollaborateBroadcast,
		Message: "",
	}
	inputJSON, _ := json.Marshal(input)

	ctx := context.Background()
	_, err := tool.Execute(ctx, inputJSON)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "message cannot be empty")
}

func TestCollaborateTool_ExecuteInvalidAction(t *testing.T) {
	tool := NewCollaborateTool(nil)

	input := map[string]interface{}{
		"action":  "invalid_action",
		"message": "test",
	}
	inputJSON, _ := json.Marshal(input)

	ctx := context.Background()
	_, err := tool.Execute(ctx, inputJSON)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid action")
}

func TestCollaborateTool_ExecutionLocation(t *testing.T) {
	tool := NewCollaborateTool(nil)
	assert.Equal(t, protocol.ExecutionLocationServer, tool.ExecutionLocation())
}
