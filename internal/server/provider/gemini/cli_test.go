package gemini

import (
	"context"
	"testing"

	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/adevcorn/ensemble/internal/server/provider"
	"github.com/stretchr/testify/assert"
)

func TestCLIProvider_Name(t *testing.T) {
	// This test doesn't require the bridge to be running
	p := &CLIProvider{
		bridgeURL: "http://localhost:3001",
	}
	assert.Equal(t, "gemini-cli", p.Name())
}

func TestCLIProvider_SupportsTools(t *testing.T) {
	p := &CLIProvider{
		bridgeURL: "http://localhost:3001",
	}
	assert.True(t, p.SupportsTools())
}

func TestNewCLIProvider_BridgeNotRunning(t *testing.T) {
	// Should fail gracefully when bridge isn't running
	_, err := NewCLIProvider("http://localhost:9999")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bridge not available")
}

// Integration test - requires bridge to be running
func TestCLIProvider_Complete_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test - requires Gemini CLI bridge to be running")
	}

	p, err := NewCLIProvider("http://localhost:3001")
	if err != nil {
		t.Skip("Skipping integration test - Gemini CLI bridge not available:", err)
	}

	ctx := context.Background()
	req := &provider.CompletionRequest{
		Model: "gemini-2.0-flash-exp",
		Messages: []protocol.Message{
			{
				Role:    protocol.MessageRoleUser,
				Content: "Say 'Hello from Gemini CLI' and nothing else.",
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

// Integration test - requires bridge to be running
func TestCLIProvider_Stream_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test - requires Gemini CLI bridge to be running")
	}

	p, err := NewCLIProvider("http://localhost:3001")
	if err != nil {
		t.Skip("Skipping integration test - Gemini CLI bridge not available:", err)
	}

	ctx := context.Background()
	req := &provider.CompletionRequest{
		Model: "gemini-2.0-flash-exp",
		Messages: []protocol.Message{
			{
				Role:    protocol.MessageRoleUser,
				Content: "Count from 1 to 3.",
			},
		},
		Temperature: 0.0,
		MaxTokens:   100,
	}

	eventChan, err := p.Stream(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, eventChan)

	var contentChunks []string
	var gotDone bool

	for event := range eventChan {
		switch event.Type {
		case provider.StreamEventContent:
			contentChunks = append(contentChunks, event.Content)
		case provider.StreamEventDone:
			gotDone = true
			assert.NotNil(t, event.Usage)
		case provider.StreamEventError:
			t.Fatalf("Unexpected error event: %v", event.Error)
		}
	}

	assert.True(t, gotDone, "Should receive done event")
	assert.NotEmpty(t, contentChunks, "Should receive content chunks")
}
