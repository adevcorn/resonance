package provider

import (
	"context"

	"github.com/adevcorn/ensemble/internal/protocol"
)

// Provider interface for LLM providers
type Provider interface {
	// Name returns the provider's name (e.g., "anthropic", "openai")
	Name() string

	// Complete performs a non-streaming completion request
	Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)

	// Stream performs a streaming completion request
	Stream(ctx context.Context, req *CompletionRequest) (<-chan StreamEvent, error)

	// SupportsTools indicates whether this provider supports tool calling
	SupportsTools() bool
}

// CompletionRequest represents a request to the LLM
type CompletionRequest struct {
	Model       string                    `json:"model"`
	Messages    []protocol.Message        `json:"messages"`
	Tools       []protocol.ToolDefinition `json:"tools,omitempty"`
	Temperature float64                   `json:"temperature"`
	MaxTokens   int                       `json:"max_tokens"`
}

// CompletionResponse represents the LLM response
type CompletionResponse struct {
	Content    string              `json:"content"`
	ToolCalls  []protocol.ToolCall `json:"tool_calls,omitempty"`
	Usage      Usage               `json:"usage"`
	StopReason string              `json:"stop_reason"`
}

// StreamEventType represents the type of streaming event
type StreamEventType string

const (
	StreamEventContent  StreamEventType = "content"
	StreamEventToolCall StreamEventType = "tool_call"
	StreamEventDone     StreamEventType = "done"
	StreamEventError    StreamEventType = "error"
)

// StreamEvent represents a streaming event
type StreamEvent struct {
	Type     StreamEventType    `json:"type"`
	Content  string             `json:"content,omitempty"`
	ToolCall *protocol.ToolCall `json:"tool_call,omitempty"`
	Usage    *Usage             `json:"usage,omitempty"`
	Error    error              `json:"-"`
	ErrorMsg string             `json:"error,omitempty"`
	Done     bool               `json:"done"`
}

// Usage tracks token usage
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}
