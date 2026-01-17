package mock

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/adevcorn/ensemble/internal/server/provider"
)

// MockProvider is a mock provider for testing
type MockProvider struct {
	mu            sync.Mutex
	name          string
	responses     []string
	currentIndex  int
	shouldError   bool
	errorMsg      string
	supportsTools bool
	toolCalls     []protocol.ToolCall
}

// NewMockProvider creates a new mock provider with pre-defined responses
func NewMockProvider(name string, responses []string) *MockProvider {
	return &MockProvider{
		name:          name,
		responses:     responses,
		supportsTools: true,
	}
}

// WithError configures the mock to return an error
func (m *MockProvider) WithError(errorMsg string) *MockProvider {
	m.shouldError = true
	m.errorMsg = errorMsg
	return m
}

// WithToolCalls configures the mock to return tool calls
func (m *MockProvider) WithToolCalls(toolCalls []protocol.ToolCall) *MockProvider {
	m.toolCalls = toolCalls
	return m
}

// WithToolSupport configures whether the mock supports tools
func (m *MockProvider) WithToolSupport(supports bool) *MockProvider {
	m.supportsTools = supports
	return m
}

// Name returns the provider name
func (m *MockProvider) Name() string {
	return m.name
}

// SupportsTools indicates whether this mock supports tool calling
func (m *MockProvider) SupportsTools() bool {
	return m.supportsTools
}

// Complete performs a non-streaming completion request
func (m *MockProvider) Complete(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldError {
		return nil, fmt.Errorf("mock error: %s", m.errorMsg)
	}

	if m.currentIndex >= len(m.responses) {
		return nil, fmt.Errorf("mock provider: no more responses available")
	}

	content := m.responses[m.currentIndex]
	m.currentIndex++

	return &provider.CompletionResponse{
		Content:   content,
		ToolCalls: m.toolCalls,
		Usage: provider.Usage{
			InputTokens:  10,
			OutputTokens: 20,
			TotalTokens:  30,
		},
		StopReason: "end_turn",
	}, nil
}

// Stream performs a streaming completion request
func (m *MockProvider) Stream(ctx context.Context, req *provider.CompletionRequest) (<-chan provider.StreamEvent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldError {
		return nil, fmt.Errorf("mock error: %s", m.errorMsg)
	}

	if m.currentIndex >= len(m.responses) {
		return nil, fmt.Errorf("mock provider: no more responses available")
	}

	content := m.responses[m.currentIndex]
	m.currentIndex++

	eventChan := make(chan provider.StreamEvent, 10)

	go func() {
		defer close(eventChan)

		// Stream content word by word
		words := splitIntoWords(content)
		for _, word := range words {
			select {
			case <-ctx.Done():
				eventChan <- provider.StreamEvent{
					Type:     provider.StreamEventError,
					Error:    ctx.Err(),
					ErrorMsg: ctx.Err().Error(),
					Done:     true,
				}
				return
			case eventChan <- provider.StreamEvent{
				Type:    provider.StreamEventContent,
				Content: word,
				Done:    false,
			}:
			}
		}

		// Send tool calls if configured
		for _, tc := range m.toolCalls {
			tcCopy := tc
			eventChan <- provider.StreamEvent{
				Type:     provider.StreamEventToolCall,
				ToolCall: &tcCopy,
				Done:     false,
			}
		}

		// Send done event
		eventChan <- provider.StreamEvent{
			Type: provider.StreamEventDone,
			Usage: &provider.Usage{
				InputTokens:  10,
				OutputTokens: 20,
				TotalTokens:  30,
			},
			Done: true,
		}
	}()

	return eventChan, nil
}

// Reset resets the mock provider's state
func (m *MockProvider) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentIndex = 0
	m.shouldError = false
	m.errorMsg = ""
	m.toolCalls = nil
}

// SetResponses updates the mock provider's responses
func (m *MockProvider) SetResponses(responses []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses = responses
	m.currentIndex = 0
}

// splitIntoWords splits content into words for simulated streaming
func splitIntoWords(content string) []string {
	if content == "" {
		return nil
	}

	// Simple word splitting - could be enhanced
	var words []string
	currentWord := ""

	for _, char := range content {
		if char == ' ' || char == '\n' || char == '\t' {
			if currentWord != "" {
				words = append(words, currentWord+string(char))
				currentWord = ""
			}
		} else {
			currentWord += string(char)
		}
	}

	if currentWord != "" {
		words = append(words, currentWord)
	}

	return words
}

// CallCount returns the number of times the provider has been called
func (m *MockProvider) CallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.currentIndex
}

// CreateMockToolCall is a helper to create a mock tool call for testing
func CreateMockToolCall(id, toolName string, args map[string]interface{}) protocol.ToolCall {
	argsJSON, _ := json.Marshal(args)
	return protocol.ToolCall{
		ID:        id,
		ToolName:  toolName,
		Arguments: argsJSON,
	}
}
