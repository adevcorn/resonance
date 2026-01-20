package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/adevcorn/ensemble/internal/server/provider"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// Provider implements the provider.Provider interface for Google Gemini
type Provider struct {
	client *genai.Client
	apiKey string
}

// NewProvider creates a new Gemini provider
func NewProvider(ctx context.Context, apiKey string) (*Provider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("gemini: API key is required")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("gemini: failed to create client: %w", err)
	}

	return &Provider{
		client: client,
		apiKey: apiKey,
	}, nil
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "gemini"
}

// SupportsTools indicates that Gemini supports tool calling
func (p *Provider) SupportsTools() bool {
	return true
}

// Complete performs a non-streaming completion request
func (p *Provider) Complete(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) {
	model := p.client.GenerativeModel(req.Model)

	// Configure model
	model.SetTemperature(float32(req.Temperature))
	model.SetMaxOutputTokens(int32(req.MaxTokens))

	// Convert tools
	if len(req.Tools) > 0 {
		tools := convertTools(req.Tools)
		model.Tools = tools
	}

	// Convert messages to Gemini format
	contents, err := convertMessages(req.Messages)
	if err != nil {
		return nil, fmt.Errorf("gemini: failed to convert messages: %w", err)
	}

	// Start chat session
	session := model.StartChat()
	session.History = contents[:len(contents)-1] // All but the last message

	// Send the last message
	lastMessage := contents[len(contents)-1]
	resp, err := session.SendMessage(ctx, lastMessage.Parts...)
	if err != nil {
		return nil, fmt.Errorf("gemini: completion failed: %w", err)
	}

	return convertResponse(resp)
}

// Stream performs a streaming completion request
func (p *Provider) Stream(ctx context.Context, req *provider.CompletionRequest) (<-chan provider.StreamEvent, error) {
	model := p.client.GenerativeModel(req.Model)

	// Configure model
	model.SetTemperature(float32(req.Temperature))
	model.SetMaxOutputTokens(int32(req.MaxTokens))

	// Convert tools
	if len(req.Tools) > 0 {
		tools := convertTools(req.Tools)
		model.Tools = tools
	}

	// Convert messages to Gemini format
	contents, err := convertMessages(req.Messages)
	if err != nil {
		return nil, fmt.Errorf("gemini: failed to convert messages: %w", err)
	}

	// Start chat session
	session := model.StartChat()
	session.History = contents[:len(contents)-1] // All but the last message

	// Send the last message with streaming
	lastMessage := contents[len(contents)-1]
	iter := session.SendMessageStream(ctx, lastMessage.Parts...)

	eventChan := make(chan provider.StreamEvent, 10)

	go func() {
		defer close(eventChan)

		var fullContent strings.Builder
		var toolCalls []protocol.ToolCall
		var totalInputTokens int
		var totalOutputTokens int

		for {
			resp, err := iter.Next()
			if err != nil {
				if err.Error() != "iterator exhausted" {
					eventChan <- provider.StreamEvent{
						Type:     provider.StreamEventError,
						Error:    err,
						ErrorMsg: err.Error(),
						Done:     true,
					}
				}
				break
			}

			// Process response parts
			for _, candidate := range resp.Candidates {
				if candidate.Content != nil {
					for _, part := range candidate.Content.Parts {
						// Handle text content
						if text, ok := part.(genai.Text); ok {
							content := string(text)
							fullContent.WriteString(content)
							eventChan <- provider.StreamEvent{
								Type:    provider.StreamEventContent,
								Content: content,
								Done:    false,
							}
						}

						// Handle function calls (tool calls)
						if fc, ok := part.(genai.FunctionCall); ok {
							args, err := json.Marshal(fc.Args)
							if err != nil {
								continue
							}

							toolCall := protocol.ToolCall{
								ID:        fc.Name, // Gemini doesn't provide IDs, use name
								ToolName:  fc.Name,
								Arguments: args,
							}
							toolCalls = append(toolCalls, toolCall)

							eventChan <- provider.StreamEvent{
								Type:     provider.StreamEventToolCall,
								ToolCall: &toolCall,
								Done:     false,
							}
						}
					}
				}
			}

			// Track token usage
			if resp.UsageMetadata != nil {
				totalInputTokens = int(resp.UsageMetadata.PromptTokenCount)
				totalOutputTokens = int(resp.UsageMetadata.CandidatesTokenCount)
			}
		}

		// Send final event with usage
		usage := &provider.Usage{
			InputTokens:  totalInputTokens,
			OutputTokens: totalOutputTokens,
			TotalTokens:  totalInputTokens + totalOutputTokens,
		}

		eventChan <- provider.StreamEvent{
			Type:  provider.StreamEventDone,
			Usage: usage,
			Done:  true,
		}
	}()

	return eventChan, nil
}

// Close closes the Gemini client
func (p *Provider) Close() error {
	return p.client.Close()
}

// convertMessages converts protocol messages to Gemini Content format
func convertMessages(messages []protocol.Message) ([]*genai.Content, error) {
	var contents []*genai.Content

	for _, msg := range messages {
		// Skip system messages - Gemini handles them differently
		if msg.Role == protocol.MessageRoleSystem {
			continue
		}

		// Map roles
		var role string
		if msg.Role == protocol.MessageRoleUser {
			role = "user"
		} else if msg.Role == protocol.MessageRoleAssistant {
			role = "model"
		} else {
			continue
		}

		var parts []genai.Part

		// Handle regular text content
		if msg.Content != "" {
			parts = append(parts, genai.Text(msg.Content))
		}

		// Handle tool calls
		for _, tc := range msg.ToolCalls {
			var args map[string]interface{}
			if err := json.Unmarshal(tc.Arguments, &args); err != nil {
				return nil, fmt.Errorf("invalid tool call arguments: %w", err)
			}

			parts = append(parts, genai.FunctionCall{
				Name: tc.ToolName,
				Args: args,
			})
		}

		// Handle tool results
		for _, tr := range msg.ToolResults {
			var result map[string]interface{}
			if tr.Error != "" {
				result = map[string]interface{}{"error": tr.Error}
			} else {
				// Try to unmarshal as JSON, otherwise use raw string
				if err := json.Unmarshal(tr.Result, &result); err != nil {
					result = map[string]interface{}{"result": string(tr.Result)}
				}
			}

			parts = append(parts, genai.FunctionResponse{
				Name:     tr.CallID, // Use CallID as function name
				Response: result,
			})
		}

		if len(parts) > 0 {
			contents = append(contents, &genai.Content{
				Role:  role,
				Parts: parts,
			})
		}
	}

	return contents, nil
}

// convertTools converts protocol tool definitions to Gemini Tool format
func convertTools(tools []protocol.ToolDefinition) []*genai.Tool {
	if len(tools) == 0 {
		return nil
	}

	declarations := make([]*genai.FunctionDeclaration, 0, len(tools))

	for _, tool := range tools {
		var schema *genai.Schema
		if len(tool.Parameters) > 0 {
			var paramMap map[string]interface{}
			if err := json.Unmarshal(tool.Parameters, &paramMap); err == nil {
				schema = convertSchema(paramMap)
			}
		}

		declarations = append(declarations, &genai.FunctionDeclaration{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  schema,
		})
	}

	return []*genai.Tool{
		{FunctionDeclarations: declarations},
	}
}

// convertSchema converts a JSON schema map to Gemini Schema
func convertSchema(schemaMap map[string]interface{}) *genai.Schema {
	schema := &genai.Schema{
		Type: genai.TypeObject,
	}

	// Extract properties
	if props, ok := schemaMap["properties"].(map[string]interface{}); ok {
		properties := make(map[string]*genai.Schema)

		for name, prop := range props {
			if propMap, ok := prop.(map[string]interface{}); ok {
				propSchema := &genai.Schema{}

				// Convert type
				if typeStr, ok := propMap["type"].(string); ok {
					switch typeStr {
					case "string":
						propSchema.Type = genai.TypeString
					case "number":
						propSchema.Type = genai.TypeNumber
					case "integer":
						propSchema.Type = genai.TypeInteger
					case "boolean":
						propSchema.Type = genai.TypeBoolean
					case "array":
						propSchema.Type = genai.TypeArray
					case "object":
						propSchema.Type = genai.TypeObject
					}
				}

				// Add description
				if desc, ok := propMap["description"].(string); ok {
					propSchema.Description = desc
				}

				properties[name] = propSchema
			}
		}

		schema.Properties = properties
	}

	// Extract required fields
	if required, ok := schemaMap["required"].([]interface{}); ok {
		requiredFields := make([]string, 0, len(required))
		for _, req := range required {
			if reqStr, ok := req.(string); ok {
				requiredFields = append(requiredFields, reqStr)
			}
		}
		schema.Required = requiredFields
	}

	return schema
}

// convertResponse converts a Gemini response to CompletionResponse
func convertResponse(resp *genai.GenerateContentResponse) (*provider.CompletionResponse, error) {
	var content string
	var toolCalls []protocol.ToolCall

	for _, candidate := range resp.Candidates {
		if candidate.Content != nil {
			for _, part := range candidate.Content.Parts {
				// Extract text
				if text, ok := part.(genai.Text); ok {
					content += string(text)
				}

				// Extract function calls
				if fc, ok := part.(genai.FunctionCall); ok {
					args, err := json.Marshal(fc.Args)
					if err != nil {
						return nil, fmt.Errorf("failed to marshal function call args: %w", err)
					}

					toolCalls = append(toolCalls, protocol.ToolCall{
						ID:        fc.Name, // Gemini doesn't provide IDs, use name
						ToolName:  fc.Name,
						Arguments: args,
					})
				}
			}
		}
	}

	var usage provider.Usage
	if resp.UsageMetadata != nil {
		usage = provider.Usage{
			InputTokens:  int(resp.UsageMetadata.PromptTokenCount),
			OutputTokens: int(resp.UsageMetadata.CandidatesTokenCount),
			TotalTokens:  int(resp.UsageMetadata.TotalTokenCount),
		}
	}

	return &provider.CompletionResponse{
		Content:   content,
		ToolCalls: toolCalls,
		Usage:     usage,
	}, nil
}
