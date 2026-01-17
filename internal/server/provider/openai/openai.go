package openai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/adevcorn/ensemble/internal/server/provider"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// Provider implements the provider.Provider interface for OpenAI
type Provider struct {
	client *openai.Client
	apiKey string
}

// NewProvider creates a new OpenAI provider
func NewProvider(apiKey string) (*Provider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("openai: API key is required")
	}

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)

	return &Provider{
		client: client,
		apiKey: apiKey,
	}, nil
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "openai"
}

// SupportsTools indicates that OpenAI supports tool calling
func (p *Provider) SupportsTools() bool {
	return true
}

// Complete performs a non-streaming completion request
func (p *Provider) Complete(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) {
	messages, err := convertMessages(req.Messages)
	if err != nil {
		return nil, fmt.Errorf("openai: failed to convert messages: %w", err)
	}

	tools := convertTools(req.Tools)

	params := openai.ChatCompletionNewParams{
		Model:       openai.F(req.Model),
		Messages:    openai.F(messages),
		MaxTokens:   openai.Int(int64(req.MaxTokens)),
		Temperature: openai.Float(req.Temperature),
	}

	if len(tools) > 0 {
		params.Tools = openai.F(tools)
	}

	completion, err := p.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("openai: completion failed: %w", err)
	}

	return convertResponse(completion)
}

// Stream performs a streaming completion request
func (p *Provider) Stream(ctx context.Context, req *provider.CompletionRequest) (<-chan provider.StreamEvent, error) {
	messages, err := convertMessages(req.Messages)
	if err != nil {
		return nil, fmt.Errorf("openai: failed to convert messages: %w", err)
	}

	tools := convertTools(req.Tools)

	params := openai.ChatCompletionNewParams{
		Model:       openai.F(req.Model),
		Messages:    openai.F(messages),
		MaxTokens:   openai.Int(int64(req.MaxTokens)),
		Temperature: openai.Float(req.Temperature),
	}

	if len(tools) > 0 {
		params.Tools = openai.F(tools)
	}

	stream := p.client.Chat.Completions.NewStreaming(ctx, params)

	eventChan := make(chan provider.StreamEvent, 10)

	go func() {
		defer close(eventChan)

		var currentContent string
		var currentToolCalls map[int]*protocol.ToolCall = make(map[int]*protocol.ToolCall)
		var usage provider.Usage

		for stream.Next() {
			chunk := stream.Current()

			if len(chunk.Choices) == 0 {
				continue
			}

			choice := chunk.Choices[0]
			delta := choice.Delta

			// Handle content
			if delta.Content != "" {
				currentContent += delta.Content
				eventChan <- provider.StreamEvent{
					Type:    provider.StreamEventContent,
					Content: delta.Content,
					Done:    false,
				}
			}

			// Handle tool calls
			if len(delta.ToolCalls) > 0 {
				for _, tc := range delta.ToolCalls {
					idx := int(tc.Index)

					if _, exists := currentToolCalls[idx]; !exists {
						currentToolCalls[idx] = &protocol.ToolCall{
							ID:       tc.ID,
							ToolName: tc.Function.Name,
						}
					}

					// Accumulate function arguments
					if tc.Function.Arguments != "" {
						currentToolCalls[idx].Arguments = append(currentToolCalls[idx].Arguments, []byte(tc.Function.Arguments)...)
					}
				}
			}

			// Handle finish reason
			if choice.FinishReason != "" {
				// Send any accumulated tool calls
				for _, tc := range currentToolCalls {
					eventChan <- provider.StreamEvent{
						Type:     provider.StreamEventToolCall,
						ToolCall: tc,
						Done:     false,
					}
				}
			}

			// Handle usage (if present in streaming)
			if chunk.Usage.PromptTokens > 0 || chunk.Usage.CompletionTokens > 0 {
				usage.InputTokens = int(chunk.Usage.PromptTokens)
				usage.OutputTokens = int(chunk.Usage.CompletionTokens)
				usage.TotalTokens = int(chunk.Usage.TotalTokens)
			}
		}

		if err := stream.Err(); err != nil {
			eventChan <- provider.StreamEvent{
				Type:     provider.StreamEventError,
				Error:    err,
				ErrorMsg: err.Error(),
				Done:     true,
			}
			return
		}

		// Send done event
		eventChan <- provider.StreamEvent{
			Type:  provider.StreamEventDone,
			Usage: &usage,
			Done:  true,
		}
	}()

	return eventChan, nil
}

// convertMessages converts protocol messages to OpenAI format
func convertMessages(messages []protocol.Message) ([]openai.ChatCompletionMessageParamUnion, error) {
	var result []openai.ChatCompletionMessageParamUnion

	for _, msg := range messages {
		switch msg.Role {
		case protocol.MessageRoleSystem:
			result = append(result, openai.SystemMessage(msg.Content))

		case protocol.MessageRoleUser:
			result = append(result, openai.UserMessage(msg.Content))

		case protocol.MessageRoleAssistant:
			if len(msg.ToolCalls) > 0 {
				// Assistant message with tool calls
				toolCalls := make([]openai.ChatCompletionMessageToolCallParam, 0, len(msg.ToolCalls))
				for _, tc := range msg.ToolCalls {
					toolCalls = append(toolCalls, openai.ChatCompletionMessageToolCallParam{
						ID:   openai.F(tc.ID),
						Type: openai.F(openai.ChatCompletionMessageToolCallTypeFunction),
						Function: openai.F(openai.ChatCompletionMessageToolCallFunctionParam{
							Name:      openai.F(tc.ToolName),
							Arguments: openai.F(string(tc.Arguments)),
						}),
					})
				}

				// Create assistant message with both content and tool calls
				assistantMsg := openai.ChatCompletionAssistantMessageParam{
					Role:      openai.F(openai.ChatCompletionAssistantMessageParamRoleAssistant),
					ToolCalls: openai.F(toolCalls),
				}
				if msg.Content != "" {
					assistantMsg.Content = openai.F([]openai.ChatCompletionAssistantMessageParamContentUnion{
						openai.ChatCompletionAssistantMessageParamContent{
							Type: openai.F(openai.ChatCompletionAssistantMessageParamContentTypeText),
							Text: openai.F(msg.Content),
						},
					})
				}
				result = append(result, assistantMsg)
			} else {
				result = append(result, openai.AssistantMessage(msg.Content))
			}

		case protocol.MessageRoleTool:
			// Tool result messages
			for _, tr := range msg.ToolResults {
				content := string(tr.Result)
				if tr.Error != "" {
					content = fmt.Sprintf("Error: %s", tr.Error)
				}

				result = append(result, openai.ToolMessage(tr.CallID, content))
			}
		}
	}

	return result, nil
}

// convertTools converts protocol tool definitions to OpenAI format
func convertTools(tools []protocol.ToolDefinition) []openai.ChatCompletionToolParam {
	if len(tools) == 0 {
		return nil
	}

	result := make([]openai.ChatCompletionToolParam, 0, len(tools))
	for _, tool := range tools {
		// Convert json.RawMessage to map for parameters
		var params map[string]interface{}
		if len(tool.Parameters) > 0 {
			_ = json.Unmarshal(tool.Parameters, &params)
		}

		result = append(result, openai.ChatCompletionToolParam{
			Type: openai.F(openai.ChatCompletionToolTypeFunction),
			Function: openai.F(openai.FunctionDefinitionParam{
				Name:        openai.F(tool.Name),
				Description: openai.F(tool.Description),
				Parameters:  openai.F(openai.FunctionParameters(params)),
			}),
		})
	}

	return result
}

// convertResponse converts an OpenAI completion to a CompletionResponse
func convertResponse(completion *openai.ChatCompletion) (*provider.CompletionResponse, error) {
	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("openai: no choices in completion response")
	}

	choice := completion.Choices[0]
	message := choice.Message

	var toolCalls []protocol.ToolCall
	for _, tc := range message.ToolCalls {
		toolCalls = append(toolCalls, protocol.ToolCall{
			ID:        tc.ID,
			ToolName:  tc.Function.Name,
			Arguments: json.RawMessage(tc.Function.Arguments),
		})
	}

	usage := provider.Usage{
		InputTokens:  int(completion.Usage.PromptTokens),
		OutputTokens: int(completion.Usage.CompletionTokens),
		TotalTokens:  int(completion.Usage.TotalTokens),
	}

	return &provider.CompletionResponse{
		Content:    message.Content,
		ToolCalls:  toolCalls,
		Usage:      usage,
		StopReason: string(choice.FinishReason),
	}, nil
}
