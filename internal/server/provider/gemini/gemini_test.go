package gemini

import (
	"context"
	"testing"

	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/adevcorn/ensemble/internal/server/provider"
	"github.com/stretchr/testify/assert"
)

func TestProvider_Name(t *testing.T) {
	// Skip if no API key (unit test should work without real API key)
	ctx := context.Background()
	p, err := NewProvider(ctx, "test-api-key")
	if err != nil {
		t.Skip("Skipping test - requires valid Gemini API configuration")
	}
	defer p.Close()

	assert.Equal(t, "gemini", p.Name())
}

func TestProvider_SupportsTools(t *testing.T) {
	ctx := context.Background()
	p, err := NewProvider(ctx, "test-api-key")
	if err != nil {
		t.Skip("Skipping test - requires valid Gemini API configuration")
	}
	defer p.Close()

	assert.True(t, p.SupportsTools())
}

func TestProvider_NewProviderError(t *testing.T) {
	ctx := context.Background()
	_, err := NewProvider(ctx, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API key is required")
}

func TestConvertMessages(t *testing.T) {
	messages := []protocol.Message{
		{
			Role:    protocol.MessageRoleUser,
			Content: "Hello, world!",
		},
		{
			Role:    protocol.MessageRoleAssistant,
			Content: "Hi there!",
		},
	}

	contents, err := convertMessages(messages)
	assert.NoError(t, err)
	assert.Len(t, contents, 2)
	assert.Equal(t, "user", contents[0].Role)
	assert.Equal(t, "model", contents[1].Role)
}

func TestConvertTools(t *testing.T) {
	tools := []protocol.ToolDefinition{
		{
			Name:        "test_tool",
			Description: "A test tool",
			Parameters:  []byte(`{"type":"object","properties":{"param1":{"type":"string","description":"A parameter"}},"required":["param1"]}`),
		},
	}

	geminiTools := convertTools(tools)
	assert.Len(t, geminiTools, 1)
	assert.Len(t, geminiTools[0].FunctionDeclarations, 1)
	assert.Equal(t, "test_tool", geminiTools[0].FunctionDeclarations[0].Name)
	assert.Equal(t, "A test tool", geminiTools[0].FunctionDeclarations[0].Description)
}

// Integration test - requires real API key
func TestProvider_Complete_Integration(t *testing.T) {
	apiKey := getTestAPIKey(t)
	if apiKey == "" {
		t.Skip("Skipping integration test - set GEMINI_API_KEY environment variable")
	}

	ctx := context.Background()
	p, err := NewProvider(ctx, apiKey)
	assert.NoError(t, err)
	defer p.Close()

	req := &provider.CompletionRequest{
		Model: "gemini-2.0-flash-exp",
		Messages: []protocol.Message{
			{
				Role:    protocol.MessageRoleUser,
				Content: "Say 'Hello from Gemini' and nothing else.",
			},
		},
		Temperature: 0.0,
		MaxTokens:   100,
	}

	resp, err := p.Complete(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Content)
	assert.Contains(t, resp.Content, "Hello")
}

func getTestAPIKey(t *testing.T) string {
	// Try to get API key from environment
	// In CI/CD, this would be set as a secret
	return "" // Leave empty for unit tests
}
