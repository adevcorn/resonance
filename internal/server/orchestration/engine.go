package orchestration

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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
	onMessage         func(protocol.Message) error
	onToolCall        func(protocol.ToolCall) (protocol.ToolResult, error)
	onServerToolStart func(agentName string, toolCall protocol.ToolCall) error
	onServerToolEnd   func(agentName string, toolCall protocol.ToolCall, result protocol.ToolResult) error

	// Collaboration tracking
	pendingCollaboration map[string][]string // agent -> list of agents waiting for their response
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
		pool:                 pool,
		registry:             registry,
		coordinator:          coord,
		moderator:            moderator,
		synthesizer:          synthesizer,
		onMessage:            onMessage,
		onToolCall:           onToolCall,
		pendingCollaboration: make(map[string][]string),
	}, nil
}

// SetServerToolCallbacks sets callbacks for server-side tool execution
func (e *Engine) SetServerToolCallbacks(
	onStart func(agentName string, toolCall protocol.ToolCall) error,
	onEnd func(agentName string, toolCall protocol.ToolCall, result protocol.ToolResult) error,
) {
	e.onServerToolStart = onStart
	e.onServerToolEnd = onEnd
}

// TrackCollaboration records when an agent sends a collaboration message to another
func (e *Engine) TrackCollaboration(from string, input *protocol.CollaborateInput) {
	if input.Action == protocol.CollaborateDirect || input.Action == protocol.CollaborateHelp {
		if input.ToAgent != "" {
			e.pendingCollaboration[input.ToAgent] = append(e.pendingCollaboration[input.ToAgent], from)
			fmt.Fprintf(os.Stderr, "[DEBUG] Tracked collaboration: %s -> %s\n", from, input.ToAgent)
		}
	}
}

// GetPendingCollaborators returns agents that have been addressed but haven't responded
func (e *Engine) GetPendingCollaborators() map[string][]string {
	return e.pendingCollaboration
}

// ClearPendingCollaboration clears pending collaboration for an agent after they respond
func (e *Engine) ClearPendingCollaboration(agent string) {
	delete(e.pendingCollaboration, agent)
}

// HandleCollaboration processes collaboration actions and streams them to client
func (e *Engine) HandleCollaboration(from string, input *protocol.CollaborateInput) error {
	// Format message based on action type
	var content string
	switch input.Action {
	case protocol.CollaborateDirect:
		content = fmt.Sprintf("[%s → %s]: %s", from, input.ToAgent, input.Message)
	case protocol.CollaborateHelp:
		content = fmt.Sprintf("[%s requests help from %s]: %s", from, input.ToAgent, input.Message)
	case protocol.CollaborateBroadcast:
		content = fmt.Sprintf("[%s → TEAM]: %s", from, input.Message)
	case protocol.CollaborateComplete:
		content = fmt.Sprintf("[%s signals completion]: %s", from, input.Message)
	default:
		return fmt.Errorf("unknown collaboration action: %s", input.Action)
	}

	// Create collaboration message
	collabMsg := protocol.Message{
		ID:        generateID(),
		Role:      protocol.MessageRoleSystem,
		Agent:     from,
		Content:   content,
		Timestamp: time.Now(),
		Metadata: map[string]any{
			"type":       "collaboration",
			"action":     string(input.Action),
			"from_agent": from,
		},
	}

	// Add to_agent for direct/help actions
	if input.ToAgent != "" {
		collabMsg.Metadata["to_agent"] = input.ToAgent
	}

	// Add artifacts if present
	if len(input.Artifacts) > 0 {
		collabMsg.Metadata["artifacts"] = input.Artifacts
	}

	// Stream to client via onMessage callback (NOT added to conversation history)
	if e.onMessage != nil {
		if err := e.onMessage(collabMsg); err != nil {
			return fmt.Errorf("failed to stream collaboration message: %w", err)
		}
	}

	// Track for moderator prioritization
	e.TrackCollaboration(from, input)

	return nil
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

		// Get current team (may be updated during collaboration)
		currentTeam := e.coordinator.GetActiveTeam()

		// Select next agent (considering pending collaborations)
		nextAgent, err := e.moderator.SelectNextAgent(ctx, currentTeam, messages, task, e.pendingCollaboration)
		if err != nil {
			return nil, fmt.Errorf("failed to select next agent: %w", err)
		}

		// Check for completion
		if nextAgent == "complete" {
			break
		}

		// Clear pending collaboration for this agent (they're responding now)
		e.ClearPendingCollaboration(nextAgent)

		// Get the selected agent
		selectedAgent, err := e.pool.Get(nextAgent)
		if err != nil {
			// Send helpful error message to client before failing
			if e.onMessage != nil {
				_ = e.onMessage(protocol.Message{
					Role:    protocol.MessageRoleAssistant,
					Agent:   "system",
					Content: fmt.Sprintf("❌ Error: %v\n\nThe orchestration tried to use an agent that doesn't exist. This is usually because the coordinator or an agent made a mistake in selecting team members.", err),
				})
			}
			return nil, fmt.Errorf("invalid agent selection: %w", err)
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
			fmt.Fprintf(os.Stderr, "[DEBUG] Executing %d tool calls for agent %s\n", len(toolCalls), nextAgent)
			toolResults, err := e.executeToolCalls(ctx, nextAgent, toolCalls)
			if err != nil {
				return nil, fmt.Errorf("tool execution failed: %w", err)
			}

			fmt.Fprintf(os.Stderr, "[DEBUG] Got %d tool results\n", len(toolResults))

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
				fmt.Fprintf(os.Stderr, "[DEBUG] Added tool results message to conversation\n")
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
		// Check if tool is in registry
		if e.registry.Has(toolCall.ToolName) {
			// Get the tool to check its execution location
			tool, err := e.registry.Get(toolCall.ToolName)
			if err != nil {
				toolResult := protocol.ToolResult{
					CallID: toolCall.ID,
					Error:  fmt.Sprintf("tool not found: %v", err),
				}
				results = append(results, toolResult)
				continue
			}

			// Check if this is a server-side or client-side tool
			if tool.ExecutionLocation() == protocol.ExecutionLocationServer {
				// SERVER-SIDE TOOL EXECUTION
				// Notify client that server tool is starting
				if e.onServerToolStart != nil {
					if err := e.onServerToolStart(agentName, toolCall); err != nil {
						fmt.Fprintf(os.Stderr, "[WARN] onServerToolStart callback failed: %v\n", err)
					}
				}

				// Add agent name to context
				ctx = context.WithValue(ctx, "agent_name", agentName)

				// Execute tool
				result, err := tool.Execute(ctx, toolCall.Arguments)
				var toolResult protocol.ToolResult
				if err != nil {
					fmt.Fprintf(os.Stderr, "[DEBUG] Tool %s (ID: %s) failed: %v\n", toolCall.ToolName, toolCall.ID, err)
					toolResult = protocol.ToolResult{
						CallID: toolCall.ID,
						Error:  err.Error(),
					}
				} else {
					fmt.Fprintf(os.Stderr, "[DEBUG] Tool %s (ID: %s) succeeded, result len: %d\n", toolCall.ToolName, toolCall.ID, len(result))
					toolResult = protocol.ToolResult{
						CallID: toolCall.ID,
						Result: result,
					}

					// Handle collaboration tool specially - inject visible message
					if toolCall.ToolName == "collaborate" {
						var colInput protocol.CollaborateInput
						if err := json.Unmarshal(toolCall.Arguments, &colInput); err == nil {
							// Handle collaboration (streams message to client + tracks for moderator)
							if handleErr := e.HandleCollaboration(agentName, &colInput); handleErr != nil {
								fmt.Fprintf(os.Stderr, "[WARN] Failed to handle collaboration: %v\n", handleErr)
							}
						}
					}

					// Track team assembly to keep moderator's team list updated
					if toolCall.ToolName == "assemble_team" {
						var teamInput struct {
							Agents []string `json:"agents"`
							Reason string   `json:"reason"`
						}
						if err := json.Unmarshal(toolCall.Arguments, &teamInput); err == nil {
							fmt.Fprintf(os.Stderr, "[DEBUG] Updating active team with: %v\n", teamInput.Agents)
							e.coordinator.SetActiveTeam(teamInput.Agents)
						}
					}
				}
				results = append(results, toolResult)

				// Notify client of tool completion
				if e.onServerToolEnd != nil {
					if err := e.onServerToolEnd(agentName, toolCall, toolResult); err != nil {
						fmt.Fprintf(os.Stderr, "[WARN] onServerToolEnd callback failed: %v\n", err)
					}
				}
			} else {
				// CLIENT-SIDE TOOL - delegate to onToolCall callback
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
		} else {
			// Tool not in registry at all - delegate to client
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
