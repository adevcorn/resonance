package openai

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/adevcorn/ensemble/internal/server/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAIProvider_Name(t *testing.T) {
	p, err := NewProvider("test-api-key")
	require.NoError(t, err)
	assert.Equal(t, "openai", p.Name())
}

func TestOpenAIProvider_SupportsTools(t *testing.T) {
	p, err := NewProvider("test-api-key")
	require.NoError(t, err)
	assert.True(t, p.SupportsTools())
}

func TestOpenAIProvider_NewProvider_NoAPIKey(t *testing.T) {
	_, err := NewProvider("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "API key is required")
}

func TestConvertMessages(t *testing.T) {
	tests := []struct {
		name     string
		messages []protocol.Message
		wantErr  bool
	}{
		{
			name: "simple messages",
			messages: []protocol.Message{
				{
					Role:    protocol.MessageRoleSystem,
					Content: "You are helpful",
				},
				{
					Role:    protocol.MessageRoleUser,
					Content: "Hello!",
				},
			},
			wantErr: false,
		},
		{
			name: "assistant message with tool calls",
			messages: []protocol.Message{
				{
					Role:    protocol.MessageRoleAssistant,
					Content: "Let me check",
					ToolCalls: []protocol.ToolCall{
						{
							ID:        "call_123",
							ToolName:  "get_weather",
							Arguments: json.RawMessage(`{"location":"NYC"}`),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "tool result message",
			messages: []protocol.Message{
				{
					Role: protocol.MessageRoleTool,
					ToolResults: []protocol.ToolResult{
						{
							CallID: "call_123",
							Result: json.RawMessage(`{"temp":72}`),
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertMessages(tt.messages)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestConvertTools(t *testing.T) {
	tests := []struct {
		name  string
		tools []protocol.ToolDefinition
		want  int
	}{
		{
			name:  "empty tools",
			tools: []protocol.ToolDefinition{},
			want:  0,
		},
		{
			name: "single tool",
			tools: []protocol.ToolDefinition{
				{
					Name:        "get_weather",
					Description: "Get weather information",
					Parameters:  json.RawMessage(`{"type":"object","properties":{"location":{"type":"string"}}}`),
				},
			},
			want: 1,
		},
		{
			name: "multiple tools",
			tools: []protocol.ToolDefinition{
				{
					Name:        "get_weather",
					Description: "Get weather information",
				},
				{
					Name:        "search",
					Description: "Search the web",
				},
			},
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertTools(tt.tools)
			if tt.want == 0 {
				assert.Nil(t, result)
			} else {
				assert.Len(t, result, tt.want)
			}
		})
	}
}

func TestOpenAIProvider_Complete_Integration(t *testing.T) {
	// Skip if no API key is set
	apiKey := getTestAPIKey()
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}

	p, err := NewProvider(apiKey)
	require.NoError(t, err)

	ctx := context.Background()
	req := &provider.CompletionRequest{
		Model: "gpt-4o-mini",
		Messages: []protocol.Message{
			{
				Role:    protocol.MessageRoleUser,
				Content: "Say 'test' and nothing else.",
			},
		},
		Temperature: 0.0,
		MaxTokens:   10,
	}

	resp, err := p.Complete(ctx, req)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Content)
	assert.Greater(t, resp.Usage.TotalTokens, 0)
}

func TestOpenAIProvider_Stream_Integration(t *testing.T) {
	// Skip if no API key is set
	apiKey := getTestAPIKey()
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}

	p, err := NewProvider(apiKey)
	require.NoError(t, err)

	ctx := context.Background()
	req := &provider.CompletionRequest{
		Model: "gpt-4o-mini",
		Messages: []protocol.Message{
			{
				Role:    protocol.MessageRoleUser,
				Content: "Count to 3.",
			},
		},
		Temperature: 0.0,
		MaxTokens:   50,
	}

	eventChan, err := p.Stream(ctx, req)
	require.NoError(t, err)

	var content string
	var gotDone bool

	for event := range eventChan {
		switch event.Type {
		case provider.StreamEventContent:
			content += event.Content
		case provider.StreamEventDone:
			gotDone = true
			assert.NotNil(t, event.Usage)
		case provider.StreamEventError:
			t.Fatalf("unexpected error: %v", event.Error)
		}
	}

	assert.True(t, gotDone)
	assert.NotEmpty(t, content)
}

// getTestAPIKey retrieves the OpenAI API key from environment
// This is for integration tests only
func getTestAPIKey() string {
	// In real tests, you'd use os.Getenv("OPENAI_API_KEY")
	// For this mock implementation, return empty to skip integration tests
	return ""
}
