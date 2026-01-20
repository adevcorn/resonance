package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/adevcorn/ensemble/internal/server/provider"
	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// Provider implements the provider.Provider interface for Anthropic Claude
type Provider struct {
	client *anthropic.Client
	apiKey string
}

// NewProvider creates a new Anthropic provider
func NewProvider(apiKey string) (*Provider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("anthropic: API key is required")
	}

	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)

	return &Provider{
		client: client,
		apiKey: apiKey,
	}, nil
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "anthropic"
}

// SupportsTools indicates that Anthropic supports tool calling
func (p *Provider) SupportsTools() bool {
	return true
}

// Complete performs a non-streaming completion request
func (p *Provider) Complete(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) {
	messages, err := convertMessages(req.Messages)
	if err != nil {
		return nil, fmt.Errorf("anthropic: failed to convert messages: %w", err)
	}

	tools := convertTools(req.Tools)

	params := anthropic.MessageNewParams{
		Model:       anthropic.F(req.Model),
		Messages:    anthropic.F(messages),
		MaxTokens:   anthropic.Int(int64(req.MaxTokens)),
		Temperature: anthropic.Float(req.Temperature),
	}

	if len(tools) > 0 {
		params.Tools = anthropic.F(tools)
	}

	message, err := p.client.Messages.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("anthropic: completion failed: %w", err)
	}

	return convertResponse(message)
}

// Stream performs a streaming completion request
func (p *Provider) Stream(ctx context.Context, req *provider.CompletionRequest) (<-chan provider.StreamEvent, error) {
	messages, err := convertMessages(req.Messages)
	if err != nil {
		return nil, fmt.Errorf("anthropic: failed to convert messages: %w", err)
	}

	tools := convertTools(req.Tools)

	params := anthropic.MessageNewParams{
		Model:       anthropic.F(req.Model),
		Messages:    anthropic.F(messages),
		MaxTokens:   anthropic.Int(int64(req.MaxTokens)),
		Temperature: anthropic.Float(req.Temperature),
	}

	if len(tools) > 0 {
		params.Tools = anthropic.F(tools)
	}

	stream := p.client.Messages.NewStreaming(ctx, params)

	eventChan := make(chan provider.StreamEvent, 10)

	go func() {
		defer close(eventChan)
		defer stream.Close()

		var currentContent string
		var currentToolCalls []protocol.ToolCall
		var usage provider.Usage

		// Map content block index to tool call index
		// Anthropic uses content block index (includes text blocks), we need tool call index
		contentBlockIndexToToolIndex := make(map[int]int)

		for stream.Next() {
			event := stream.Current()

			switch event := event.AsUnion().(type) {
			case anthropic.ContentBlockDeltaEvent:
				if delta := event.Delta.AsUnion(); delta != nil {
					if textDelta, ok := delta.(anthropic.TextDelta); ok {
						currentContent += textDelta.Text
						eventChan <- provider.StreamEvent{
							Type:    provider.StreamEventContent,
							Content: textDelta.Text,
							Done:    false,
						}
					} else if inputJSONDelta, ok := delta.(anthropic.InputJSONDelta); ok {
						// Accumulate tool call arguments
						contentBlockIdx := int(event.Index)

						// Map content block index to tool call index
						toolIdx, exists := contentBlockIndexToToolIndex[contentBlockIdx]
						if !exists {
							fmt.Fprintf(os.Stderr, "[DEBUG] WARNING: no tool mapping for content block index %d\n", contentBlockIdx)
							continue
						}

						fmt.Fprintf(os.Stderr, "[DEBUG] InputJSONDelta for content block %d (tool idx %d): %q\n",
							contentBlockIdx, toolIdx, inputJSONDelta.PartialJSON)

						if toolIdx >= 0 && toolIdx < len(currentToolCalls) {
							currentToolCalls[toolIdx].Arguments = append(
								currentToolCalls[toolIdx].Arguments,
								[]byte(inputJSONDelta.PartialJSON)...,
							)
						} else {
							fmt.Fprintf(os.Stderr, "[DEBUG] WARNING: tool index %d out of range (len: %d)\n", toolIdx, len(currentToolCalls))
						}
					}
				}

			case anthropic.ContentBlockStartEvent:
				if block := event.ContentBlock.AsUnion(); block != nil {
					if toolUse, ok := block.(anthropic.ToolUseBlock); ok {
						// Tool use started - record the mapping from content block index to tool index
						contentBlockIndex := int(event.Index)
						toolIndex := len(currentToolCalls)
						contentBlockIndexToToolIndex[contentBlockIndex] = toolIndex

						fmt.Fprintf(os.Stderr, "[DEBUG] Tool use started: %s (ID: %s, content block idx: %d, tool idx: %d)\n",
							toolUse.Name, toolUse.ID, contentBlockIndex, toolIndex)

						currentToolCalls = append(currentToolCalls, protocol.ToolCall{
							ID:       toolUse.ID,
							ToolName: toolUse.Name,
						})
					}
				}

			case anthropic.ContentBlockStopEvent:
				// Content block completed
				if len(currentToolCalls) > 0 {
					// Send tool call event
					for _, tc := range currentToolCalls {
						fmt.Fprintf(os.Stderr, "[DEBUG] Sending tool call: %s (ID: %s, args len: %d)\n", tc.ToolName, tc.ID, len(tc.Arguments))
						eventChan <- provider.StreamEvent{
							Type:     provider.StreamEventToolCall,
							ToolCall: &tc,
							Done:     false,
						}
					}
				}

			case anthropic.MessageDeltaEvent:
				// Track usage in event for later attachment to message
				if event.Usage.OutputTokens > 0 {
					usage.OutputTokens = int(event.Usage.OutputTokens)
				}

			case anthropic.MessageStartEvent:
				// Track usage in event for later attachment to message
				if event.Message.Usage.InputTokens > 0 {
					usage.InputTokens = int(event.Message.Usage.InputTokens)
				}

			case anthropic.MessageStopEvent:
				// Stream completed
				usage.TotalTokens = usage.InputTokens + usage.OutputTokens
				eventChan <- provider.StreamEvent{
					Type:  provider.StreamEventDone,
					Usage: &usage,
					Done:  true,
				}
			}
		}

		if err := stream.Err(); err != nil {
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

// convertMessages converts protocol messages to Anthropic format
func convertMessages(messages []protocol.Message) ([]anthropic.MessageParam, error) {
	var result []anthropic.MessageParam

	// Track skipped tool call IDs to skip corresponding results
	skippedToolCalls := make(map[string]bool)

	for _, msg := range messages {
		// Skip system messages - they should be handled separately
		if msg.Role == protocol.MessageRoleSystem {
			continue
		}

		role := anthropic.MessageParamRole(msg.Role)

		// Handle different content types
		if len(msg.ToolCalls) > 0 {
			// Message with tool calls - these should be assistant messages
			var contentBlocks []anthropic.ContentBlockParamUnion

			// Trim whitespace from content to avoid Anthropic API errors
			trimmedContent := strings.TrimSpace(msg.Content)
			if trimmedContent != "" {
				contentBlocks = append(contentBlocks, anthropic.NewTextBlock(trimmedContent))
			}

			for _, tc := range msg.ToolCalls {
				// Skip tool calls with empty arguments
				if len(tc.Arguments) == 0 {
					fmt.Fprintf(os.Stderr, "[DEBUG] Skipping tool call %s (tool: %s): empty arguments\n", tc.ID, tc.ToolName)
					skippedToolCalls[tc.ID] = true
					continue
				}

				// Try to unmarshal arguments - skip if invalid
				var input map[string]interface{}
				if err := json.Unmarshal(tc.Arguments, &input); err != nil {
					// Skip tool calls with malformed JSON instead of failing the entire request
					// This can happen when agents create invalid tool calls
					fmt.Fprintf(os.Stderr, "[DEBUG] Skipping tool call %s (tool: %s): invalid JSON: %v (args: %q)\n", tc.ID, tc.ToolName, err, string(tc.Arguments))
					skippedToolCalls[tc.ID] = true
					continue
				}

				contentBlocks = append(contentBlocks, anthropic.ToolUseBlockParam{
					Type:  anthropic.F(anthropic.ToolUseBlockParamTypeToolUse),
					ID:    anthropic.F(tc.ID),
					Name:  anthropic.F(tc.ToolName),
					Input: anthropic.F[any](input),
				})
			}

			result = append(result, anthropic.MessageParam{
				Role:    anthropic.F(anthropic.MessageParamRoleAssistant),
				Content: anthropic.F(contentBlocks),
			})
		} else if len(msg.ToolResults) > 0 {
			// Message with tool results
			var contentBlocks []anthropic.ContentBlockParamUnion

			for _, tr := range msg.ToolResults {
				// Skip tool results for skipped tool calls
				if skippedToolCalls[tr.CallID] {
					fmt.Fprintf(os.Stderr, "[DEBUG] Skipping tool result for skipped call %s\n", tr.CallID)
					continue
				}

				var content string
				if tr.Error != "" {
					content = fmt.Sprintf("Error: %s", tr.Error)
				} else {
					content = string(tr.Result)
				}

				contentBlocks = append(contentBlocks, anthropic.NewToolResultBlock(tr.CallID, content, false))
			}

			// Only add the message if there are content blocks
			if len(contentBlocks) > 0 {
				result = append(result, anthropic.NewUserMessage(contentBlocks...))
			}
		} else {
			// Regular text message
			// Trim whitespace to avoid Anthropic API errors about trailing whitespace
			trimmedContent := strings.TrimSpace(msg.Content)

			// Skip empty messages (after trimming)
			if trimmedContent == "" {
				continue
			}

			result = append(result, anthropic.MessageParam{
				Role:    anthropic.F(role),
				Content: anthropic.F([]anthropic.ContentBlockParamUnion{anthropic.NewTextBlock(trimmedContent)}),
			})
		}
	}

	return result, nil
}

// convertTools converts protocol tool definitions to Anthropic format
func convertTools(tools []protocol.ToolDefinition) []anthropic.ToolParam {
	if len(tools) == 0 {
		return nil
	}

	result := make([]anthropic.ToolParam, 0, len(tools))
	for _, tool := range tools {
		var inputSchema interface{}
		if len(tool.Parameters) > 0 {
			_ = json.Unmarshal(tool.Parameters, &inputSchema)
		}

		result = append(result, anthropic.ToolParam{
			Name:        anthropic.F(tool.Name),
			Description: anthropic.F(tool.Description),
			InputSchema: anthropic.F(inputSchema),
		})
	}

	return result
}

// convertResponse converts an Anthropic message to a CompletionResponse
func convertResponse(message *anthropic.Message) (*provider.CompletionResponse, error) {
	var content string
	var toolCalls []protocol.ToolCall

	for _, block := range message.Content {
		switch block := block.AsUnion().(type) {
		case anthropic.TextBlock:
			content += block.Text
		case anthropic.ToolUseBlock:
			args, err := json.Marshal(block.Input)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal tool input: %w", err)
			}
			toolCalls = append(toolCalls, protocol.ToolCall{
				ID:        block.ID,
				ToolName:  block.Name,
				Arguments: args,
			})
		}
	}

	usage := provider.Usage{
		InputTokens:  int(message.Usage.InputTokens),
		OutputTokens: int(message.Usage.OutputTokens),
		TotalTokens:  int(message.Usage.InputTokens + message.Usage.OutputTokens),
	}

	return &provider.CompletionResponse{
		Content:    content,
		ToolCalls:  toolCalls,
		Usage:      usage,
		StopReason: string(message.StopReason),
	}, nil
}
