package orchestration

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/adevcorn/ensemble/internal/server/agent"
	"github.com/adevcorn/ensemble/internal/server/provider"
	"github.com/adevcorn/ensemble/internal/server/tool"
)

// Engine is the main orchestration engine
type Engine struct {
	pool        *agent.Pool
	registry    *tool.Registry
	coordinator *Coordinator
	moderator   *Moderator
	synthesizer *Synthesizer

	// Callbacks for streaming events
	onMessage  func(protocol.Message) error
	onToolCall func(protocol.ToolCall) (protocol.ToolResult, error)
}

// RunResult contains the final result of orchestration
type RunResult struct {
	Summary   string
	Artifacts []string
	Messages  []protocol.Message
}

// NewEngine creates a new orchestration engine
func NewEngine(
	pool *agent.Pool,
	registry *tool.Registry,
	onMessage func(protocol.Message) error,
	onToolCall func(protocol.ToolCall) (protocol.ToolResult, error),
) (*Engine, error) {
	// Create coordinator
	coord, err := NewCoordinator(pool, registry)
	if err != nil {
		return nil, fmt.Errorf("failed to create coordinator: %w", err)
	}

	// Get coordinator agent for moderator and synthesizer
	coordinatorAgent, err := pool.Get("coordinator")
	if err != nil {
		return nil, fmt.Errorf("coordinator agent not found: %w", err)
	}

	moderator := NewModerator(coordinatorAgent)
	synthesizer := NewSynthesizer(coordinatorAgent)

	return &Engine{
		pool:        pool,
		registry:    registry,
		coordinator: coord,
		moderator:   moderator,
		synthesizer: synthesizer,
		onMessage:   onMessage,
		onToolCall:  onToolCall,
	}, nil
}

// Run executes a task with multi-agent collaboration
func (e *Engine) Run(ctx context.Context, task string, projectInfo *protocol.ProjectInfo) (*RunResult, error) {
	// Step 1: Analyze task and assemble team
	team, err := e.coordinator.AnalyzeTask(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze task: %w", err)
	}

	e.coordinator.SetActiveTeam(team)

	// Initialize conversation with user's task
	messages := []protocol.Message{
		{
			ID:        generateID(),
			Role:      protocol.MessageRoleUser,
			Content:   task,
			Timestamp: time.Now(),
		},
	}

	// Notify about task message
	if e.onMessage != nil {
		_ = e.onMessage(messages[0])
	}

	// Step 2: Collaboration loop
	const maxTurns = 20
	for turn := 0; turn < maxTurns; turn++ {
		// Check if we should continue
		if !e.moderator.ShouldContinue(messages) {
			break
		}

		// Select next agent
		nextAgent, err := e.moderator.SelectNextAgent(ctx, team, messages, task)
		if err != nil {
			return nil, fmt.Errorf("failed to select next agent: %w", err)
		}

		// Check for completion
		if nextAgent == "complete" {
			break
		}

		// Get the selected agent
		selectedAgent, err := e.pool.Get(nextAgent)
		if err != nil {
			return nil, fmt.Errorf("failed to get agent %s: %w", nextAgent, err)
		}

		// Build conversation context for this agent
		// Include system message with agent's role
		agentMessages := []protocol.Message{
			{
				Role:    protocol.MessageRoleSystem,
				Content: selectedAgent.SystemPrompt(),
			},
		}
		agentMessages = append(agentMessages, messages...)

		// Get tools allowed for this agent
		tools := e.registry.GetAllowed(selectedAgent)

		// Create completion request
		req := &provider.CompletionRequest{
			Messages: agentMessages,
			Tools:    tools,
		}

		// Call agent (streaming)
		eventChan, err := selectedAgent.Stream(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("agent %s stream failed: %w", nextAgent, err)
		}

		// Collect response
		var contentBuilder strings.Builder
		var toolCalls []protocol.ToolCall

		for event := range eventChan {
			switch event.Type {
			case provider.StreamEventContent:
				contentBuilder.WriteString(event.Content)
			case provider.StreamEventToolCall:
				if event.ToolCall != nil {
					toolCalls = append(toolCalls, *event.ToolCall)
				}
			case provider.StreamEventError:
				return nil, fmt.Errorf("stream error from agent %s: %w", nextAgent, event.Error)
			}
		}

		// Create message from agent's response
		agentMsg := protocol.Message{
			ID:        generateID(),
			Role:      protocol.MessageRoleAssistant,
			Agent:     nextAgent,
			Content:   contentBuilder.String(),
			ToolCalls: toolCalls,
			Timestamp: time.Now(),
		}

		// Add to conversation
		messages = append(messages, agentMsg)

		// Stream message to client
		if e.onMessage != nil {
			if err := e.onMessage(agentMsg); err != nil {
				return nil, fmt.Errorf("failed to stream message: %w", err)
			}
		}

		// Execute tool calls if any
		if len(toolCalls) > 0 {
			toolResults, err := e.executeToolCalls(ctx, nextAgent, toolCalls)
			if err != nil {
				return nil, fmt.Errorf("tool execution failed: %w", err)
			}

			// Add tool results to conversation
			if len(toolResults) > 0 {
				toolMsg := protocol.Message{
					ID:          generateID(),
					Role:        protocol.MessageRoleTool,
					Agent:       nextAgent,
					ToolResults: toolResults,
					Timestamp:   time.Now(),
				}
				messages = append(messages, toolMsg)
			}
		}
	}

	// Step 3: Synthesize results
	summary, artifacts, err := e.synthesizer.Synthesize(ctx, task, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to synthesize results: %w", err)
	}

	return &RunResult{
		Summary:   summary,
		Artifacts: artifacts,
		Messages:  messages,
	}, nil
}

// executeToolCalls executes tool calls and returns results
func (e *Engine) executeToolCalls(ctx context.Context, agentName string, toolCalls []protocol.ToolCall) ([]protocol.ToolResult, error) {
	var results []protocol.ToolResult

	for _, toolCall := range toolCalls {
		// Check if tool is in registry (server-side tool)
		if e.registry.Has(toolCall.ToolName) {
			// Execute server-side tool
			tool, err := e.registry.Get(toolCall.ToolName)
			if err != nil {
				results = append(results, protocol.ToolResult{
					CallID: toolCall.ID,
					Error:  fmt.Sprintf("tool not found: %v", err),
				})
				continue
			}

			// Add agent name to context
			ctx = context.WithValue(ctx, "agent_name", agentName)

			// Execute tool
			result, err := tool.Execute(ctx, toolCall.Arguments)
			if err != nil {
				results = append(results, protocol.ToolResult{
					CallID: toolCall.ID,
					Error:  err.Error(),
				})
			} else {
				results = append(results, protocol.ToolResult{
					CallID: toolCall.ID,
					Result: result,
				})
			}
		} else {
			// Client-side tool - delegate to onToolCall callback
			if e.onToolCall != nil {
				result, err := e.onToolCall(toolCall)
				if err != nil {
					results = append(results, protocol.ToolResult{
						CallID: toolCall.ID,
						Error:  err.Error(),
					})
				} else {
					results = append(results, result)
				}
			} else {
				results = append(results, protocol.ToolResult{
					CallID: toolCall.ID,
					Error:  "client tool executor not available",
				})
			}
		}
	}

	return results, nil
}

// generateID generates a unique ID for messages
func generateID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}
