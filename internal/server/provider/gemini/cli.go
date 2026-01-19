package gemini

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/adevcorn/ensemble/internal/server/provider"
)

// CLIProvider implements the provider.Provider interface using Gemini CLI bridge
type CLIProvider struct {
	bridgeURL string
	client    *http.Client
}

// NewCLIProvider creates a new Gemini CLI provider that communicates with Node.js bridge
func NewCLIProvider(bridgeURL string) (*CLIProvider, error) {
	if bridgeURL == "" {
		bridgeURL = "http://localhost:3001"
	}

	// Test connection to bridge
	resp, err := http.Get(bridgeURL + "/health")
	if err != nil {
		return nil, fmt.Errorf("gemini-cli: bridge not available at %s - is the Node.js bridge running? Error: %w", bridgeURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("gemini-cli: bridge unhealthy (status %d)", resp.StatusCode)
	}

	return &CLIProvider{
		bridgeURL: bridgeURL,
		client:    &http.Client{},
	}, nil
}

// Name returns the provider name
func (p *CLIProvider) Name() string {
	return "gemini"
}

// SupportsTools indicates that Gemini CLI supports tool calling
func (p *CLIProvider) SupportsTools() bool {
	return true
}

// bridgeRequest represents a request to the Node.js bridge
type bridgeRequest struct {
	Model       string                    `json:"model"`
	Messages    []protocol.Message        `json:"messages"`
	Tools       []protocol.ToolDefinition `json:"tools,omitempty"`
	Temperature float64                   `json:"temperature"`
	MaxTokens   int                       `json:"maxTokens"`
}

// bridgeEvent represents a streaming event from the bridge
type bridgeEvent struct {
	Type     string             `json:"type"`
	Content  string             `json:"content,omitempty"`
	ToolCall *protocol.ToolCall `json:"toolCall,omitempty"`
	Usage    *provider.Usage    `json:"usage,omitempty"`
	Error    string             `json:"error,omitempty"`
}

// Complete performs a non-streaming completion request
func (p *CLIProvider) Complete(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) {
	// For non-streaming, we collect all events from the stream
	eventChan, err := p.Stream(ctx, req)
	if err != nil {
		return nil, err
	}

	var content strings.Builder
	var toolCalls []protocol.ToolCall
	var usage provider.Usage

	for event := range eventChan {
		switch event.Type {
		case provider.StreamEventContent:
			content.WriteString(event.Content)
		case provider.StreamEventToolCall:
			if event.ToolCall != nil {
				toolCalls = append(toolCalls, *event.ToolCall)
			}
		case provider.StreamEventDone:
			if event.Usage != nil {
				usage = *event.Usage
			}
		case provider.StreamEventError:
			return nil, event.Error
		}
	}

	return &provider.CompletionResponse{
		Content:   content.String(),
		ToolCalls: toolCalls,
		Usage:     usage,
	}, nil
}

// Stream performs a streaming completion request via the Node.js bridge
func (p *CLIProvider) Stream(ctx context.Context, req *provider.CompletionRequest) (<-chan provider.StreamEvent, error) {
	bridgeReq := bridgeRequest{
		Model:       req.Model,
		Messages:    req.Messages,
		Tools:       req.Tools,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
	}

	reqBody, err := json.Marshal(bridgeReq)
	if err != nil {
		return nil, fmt.Errorf("gemini-cli: failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.bridgeURL+"/v1/completions", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("gemini-cli: failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("gemini-cli: request failed: %w", err)
	}

	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gemini-cli: request failed with status %d: %s", resp.StatusCode, string(body))
	}

	eventChan := make(chan provider.StreamEvent, 10)

	go func() {
		defer close(eventChan)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			// SSE format: "data: {json}"
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			var event bridgeEvent
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				eventChan <- provider.StreamEvent{
					Type:     provider.StreamEventError,
					Error:    fmt.Errorf("failed to parse event: %w", err),
					ErrorMsg: err.Error(),
					Done:     true,
				}
				return
			}

			switch event.Type {
			case "content":
				eventChan <- provider.StreamEvent{
					Type:    provider.StreamEventContent,
					Content: event.Content,
					Done:    false,
				}

			case "tool_call":
				eventChan <- provider.StreamEvent{
					Type:     provider.StreamEventToolCall,
					ToolCall: event.ToolCall,
					Done:     false,
				}

			case "done":
				eventChan <- provider.StreamEvent{
					Type:  provider.StreamEventDone,
					Usage: event.Usage,
					Done:  true,
				}

			case "error":
				eventChan <- provider.StreamEvent{
					Type:     provider.StreamEventError,
					Error:    fmt.Errorf("bridge error: %s", event.Error),
					ErrorMsg: event.Error,
					Done:     true,
				}
			}
		}

		if err := scanner.Err(); err != nil {
			eventChan <- provider.StreamEvent{
				Type:     provider.StreamEventError,
				Error:    err,
				ErrorMsg: err.Error(),
				Done:     true,
			}
		}
	}()

	return eventChan, nil
}
