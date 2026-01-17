package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/adevcorn/ensemble/internal/protocol"
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
	Agent     string    `json:"agent"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// ToolCallPayload is the payload for "tool_call" events
type ToolCallPayload struct {
	CallID    string          `json:"call_id"`
	ToolName  string          `json:"tool_name"`
	Arguments json.RawMessage `json:"arguments"`
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

			// Run orchestration
			// Note: This is a simplified version. In production, we'd need to:
			// 1. Stream agent messages to client via "agent_message" events
			// 2. Send tool calls to client via "tool_call" events
			// 3. Wait for "tool_result" messages from client
			// 4. Handle cancellation

			result, err := s.engine.Run(ctx, start.Task, start.ProjectInfo)
			if err != nil {
				sendError(conn, "Orchestration failed: "+err.Error())
				session.State = protocol.SessionStateError
				_ = s.sessionManager.Update(ctx, session)
				continue
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

		case "tool_result":
			// TODO: Handle tool results
			log.Warn().Msg("Tool result handling not yet implemented")

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
