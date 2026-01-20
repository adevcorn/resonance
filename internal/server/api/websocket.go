package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/adevcorn/ensemble/internal/server/orchestration"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for now
		// TODO: Implement proper origin checking in production
		return true
	},
}

// WSClientMessage is a message from client to server
type WSClientMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// WSServerMessage is a message from server to client
type WSServerMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// StartPayload is the payload for the "start" message
type StartPayload struct {
	Task        string                `json:"task"`
	ProjectInfo *protocol.ProjectInfo `json:"project_info,omitempty"`
}

// AgentMessagePayload is the payload for "agent_message" events
type AgentMessagePayload struct {
	Agent     string            `json:"agent"`
	Content   string            `json:"content"`
	Timestamp time.Time         `json:"timestamp"`
	Tokens    *MessageTokenInfo `json:"tokens,omitempty"`
}

// MessageTokenInfo contains token usage information
type MessageTokenInfo struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// ToolCallPayload is the payload for "tool_call" events
type ToolCallPayload struct {
	CallID     string          `json:"call_id"`
	ToolName   string          `json:"tool_name"`
	Arguments  json.RawMessage `json:"arguments"`
	ServerSide bool            `json:"server_side,omitempty"` // true if executed on server
}

// ToolResultPayload is the payload for "tool_result" messages from client
type ToolResultPayload struct {
	CallID string          `json:"call_id"`
	Result json.RawMessage `json:"result"`
	Error  string          `json:"error,omitempty"`
}

// CompletePayload is the payload for "complete" events
type CompletePayload struct {
	Summary   string   `json:"summary"`
	Artifacts []string `json:"artifacts"`
}

// ErrorPayload is the payload for "error" events
type ErrorPayload struct {
	Message string `json:"message"`
}

// handleWebSocket handles WebSocket connections for streaming collaboration
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	// Verify session exists
	session, err := s.sessionManager.Get(r.Context(), sessionID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Session not found")
		return
	}

	// Upgrade connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upgrade WebSocket connection")
		return
	}
	defer conn.Close()

	log.Info().Str("session_id", sessionID).Msg("WebSocket connection established")

	// Handle WebSocket communication
	ctx := r.Context()

	// Track pending tool calls
	pendingToolCalls := make(map[string]chan protocol.ToolResult)
	var pendingMu sync.Mutex

	for {
		var msg WSClientMessage
		if err := conn.ReadJSON(&msg); err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Error().Err(err).Msg("WebSocket read error")
			}
			return
		}

		switch msg.Type {
		case "start":
			var start StartPayload
			if err := json.Unmarshal(msg.Payload, &start); err != nil {
				sendError(conn, "Invalid start payload")
				continue
			}

			if start.Task == "" {
				sendError(conn, "Task is required")
				continue
			}

			log.Info().Str("session_id", sessionID).Str("task", start.Task).Msg("Starting task")

			// Create a per-session engine with streaming callbacks
			sessionEngine, err := orchestration.NewEngine(
				s.agentPool,
				s.registry,
				func(msg protocol.Message) error {
					// Stream agent messages to client
					tokenInfo := &MessageTokenInfo{}
					if msg.Metadata != nil {
						if tokens, ok := msg.Metadata["tokens"].(map[string]interface{}); ok {
							tokenInfo = &MessageTokenInfo{
								InputTokens:  int(tokens["input_tokens"].(float64)),
								OutputTokens: int(tokens["output_tokens"].(float64)),
								TotalTokens:  int(tokens["total_tokens"].(float64)),
							}
						}
					}

					return conn.WriteJSON(WSServerMessage{
						Type: "agent_message",
						Payload: AgentMessagePayload{
							Agent:     msg.Agent,
							Content:   msg.Content,
							Timestamp: msg.Timestamp,
							Tokens:    tokenInfo,
						},
					})
				},
				func(call protocol.ToolCall) (protocol.ToolResult, error) {
					// Send tool call to client
					log.Debug().
						Str("call_id", call.ID).
						Str("tool_name", call.ToolName).
						Msg("Sending client-side tool call")

					if err := conn.WriteJSON(WSServerMessage{
						Type: "tool_call",
						Payload: ToolCallPayload{
							CallID:    call.ID,
							ToolName:  call.ToolName,
							Arguments: call.Arguments,
						},
					}); err != nil {
						return protocol.ToolResult{}, err
					}

					// Create channel for this tool call result
					resultChan := make(chan protocol.ToolResult, 1)

					// Register pending tool call
					pendingMu.Lock()
					pendingToolCalls[call.ID] = resultChan
					log.Debug().
						Str("call_id", call.ID).
						Int("pending_count", len(pendingToolCalls)).
						Msg("Registered pending tool call")
					pendingMu.Unlock()

					// Wait for client to send tool_result message
					select {
					case result := <-resultChan:
						// Clean up
						pendingMu.Lock()
						delete(pendingToolCalls, call.ID)
						pendingMu.Unlock()

						// Check if result has error
						if result.Error != "" {
							return protocol.ToolResult{}, fmt.Errorf("tool execution failed: %s", result.Error)
						}
						return result, nil

					case <-ctx.Done():
						// Clean up on context cancellation
						pendingMu.Lock()
						delete(pendingToolCalls, call.ID)
						pendingMu.Unlock()
						return protocol.ToolResult{}, ctx.Err()

					case <-time.After(120 * time.Second):
						// Clean up on timeout (increased to 120s for complex multi-tool operations)
						pendingMu.Lock()
						delete(pendingToolCalls, call.ID)
						pendingMu.Unlock()
						log.Warn().
							Str("call_id", call.ID).
							Str("tool_name", call.ToolName).
							Msg("Tool execution timeout - removed from pending calls")
						return protocol.ToolResult{}, fmt.Errorf("tool execution timeout")
					}
				},
			)

			// Check for engine creation error before using it
			if err != nil {
				sendError(conn, "Failed to create engine: "+err.Error())
				session.State = protocol.SessionStateError
				_ = s.sessionManager.Update(ctx, session)
				continue
			}

			sessionEngine.SetServerToolCallbacks(
				// onServerToolStart - notify client that server tool is starting
				func(agentName string, toolCall protocol.ToolCall) error {
					return conn.WriteJSON(WSServerMessage{
						Type: "tool_call",
						Payload: ToolCallPayload{
							CallID:     toolCall.ID,
							ToolName:   toolCall.ToolName,
							Arguments:  toolCall.Arguments,
							ServerSide: true,
						},
					})
				},
				// onServerToolEnd - notify client of server tool result
				func(agentName string, toolCall protocol.ToolCall, result protocol.ToolResult) error {
					return conn.WriteJSON(WSServerMessage{
						Type: "tool_result",
						Payload: ToolResultPayload{
							CallID: result.CallID,
							Result: result.Result,
							Error:  result.Error,
						},
					})
				},
			)

			// Run orchestration in goroutine to avoid blocking message loop
			go func() {
				result, err := sessionEngine.Run(ctx, start.Task, start.ProjectInfo)
				if err != nil {
					sendError(conn, "Orchestration failed: "+err.Error())
					session.State = protocol.SessionStateError
					_ = s.sessionManager.Update(ctx, session)
					return
				}

				// Send completion
				if err := conn.WriteJSON(WSServerMessage{
					Type: "complete",
					Payload: CompletePayload{
						Summary:   result.Summary,
						Artifacts: result.Artifacts,
					},
				}); err != nil {
					log.Error().Err(err).Msg("Failed to send complete message")
					return
				}

				// Update session
				session.State = protocol.SessionStateCompleted
				session.Messages = result.Messages
				_ = s.sessionManager.Update(ctx, session)
			}()

		case "tool_result":
			var result ToolResultPayload
			if err := json.Unmarshal(msg.Payload, &result); err != nil {
				sendError(conn, "Invalid tool result payload")
				continue
			}

			log.Debug().
				Str("call_id", result.CallID).
				Str("error", result.Error).
				Msg("Received tool result from client")

			// Find the pending tool call and send result
			pendingMu.Lock()
			resultChan, exists := pendingToolCalls[result.CallID]
			if !exists {
				// Log all pending call IDs to help debug
				pendingIDs := make([]string, 0, len(pendingToolCalls))
				for id := range pendingToolCalls {
					pendingIDs = append(pendingIDs, id)
				}
				pendingMu.Unlock()
				log.Warn().
					Str("call_id", result.CallID).
					Strs("pending_calls", pendingIDs).
					Msg("Received result for unknown tool call")
				continue
			}
			pendingMu.Unlock()

			// Send result to waiting goroutine
			select {
			case resultChan <- protocol.ToolResult{
				CallID: result.CallID,
				Result: result.Result,
				Error:  result.Error,
			}:
				// Successfully delivered
			default:
				// Channel was already closed or full (shouldn't happen with buffered channel)
				log.Warn().Str("call_id", result.CallID).Msg("Failed to deliver tool result")
			}

		case "cancel":
			log.Info().Str("session_id", sessionID).Msg("Task cancelled by client")
			session.State = protocol.SessionStateCompleted
			_ = s.sessionManager.Update(ctx, session)
			return

		case "ping":
			if err := conn.WriteJSON(WSServerMessage{
				Type:    "pong",
				Payload: nil,
			}); err != nil {
				log.Error().Err(err).Msg("Failed to send pong")
				return
			}

		default:
			log.Warn().Str("type", msg.Type).Msg("Unknown message type")
		}
	}
}

// sendError sends an error message to the WebSocket client
func sendError(conn *websocket.Conn, message string) {
	err := conn.WriteJSON(WSServerMessage{
		Type: "error",
		Payload: ErrorPayload{
			Message: message,
		},
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to send error message")
	}
}
