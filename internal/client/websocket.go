package client

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/gorilla/websocket"
)

// WebSocketConn wraps a WebSocket connection to the server
type WebSocketConn struct {
	conn               *websocket.Conn
	mu                 sync.Mutex
	onMessage          func(protocol.Message) error
	onToolCall         func(protocol.ToolCall) (protocol.ToolResult, error)
	onServerToolStart  func(protocol.ToolCall) error
	onServerToolEnd    func(protocol.ToolCall, protocol.ToolResult) error
	onComplete         func(summary string, artifacts []string) error
	onError            func(error) error
	pendingServerTools map[string]protocol.ToolCall // Track server-side tool calls by ID
}

// WSClientMessage is a message from client to server
type WSClientMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// WSServerMessage is a message from server to client
type WSServerMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// StartPayload is the payload for starting a task
type StartPayload struct {
	Task        string                `json:"task"`
	ProjectInfo *protocol.ProjectInfo `json:"project_info,omitempty"`
}

// AgentMessagePayload is the payload for agent messages
type AgentMessagePayload struct {
	Agent     string `json:"agent"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

// ToolCallPayload is the payload for tool calls
type ToolCallPayload struct {
	CallID     string          `json:"call_id"`
	ToolName   string          `json:"tool_name"`
	Arguments  json.RawMessage `json:"arguments"`
	ServerSide bool            `json:"server_side,omitempty"` // true if executed on server
}

// ToolResultPayload is the payload for tool results
type ToolResultPayload struct {
	CallID string          `json:"call_id"`
	Result json.RawMessage `json:"result"`
	Error  string          `json:"error,omitempty"`
}

// CompletePayload is the payload for completion
type CompletePayload struct {
	Summary   string   `json:"summary"`
	Artifacts []string `json:"artifacts"`
}

// ErrorPayload is the payload for errors
type ErrorPayload struct {
	Message string `json:"message"`
}

// NewWebSocketConn creates a new WebSocket connection wrapper
func NewWebSocketConn(conn *websocket.Conn) *WebSocketConn {
	return &WebSocketConn{
		conn:               conn,
		pendingServerTools: make(map[string]protocol.ToolCall),
	}
}

// SetCallbacks sets the event callbacks
func (ws *WebSocketConn) SetCallbacks(
	onMessage func(protocol.Message) error,
	onToolCall func(protocol.ToolCall) (protocol.ToolResult, error),
	onComplete func(summary string, artifacts []string) error,
	onError func(error) error,
) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	ws.onMessage = onMessage
	ws.onToolCall = onToolCall
	ws.onComplete = onComplete
	ws.onError = onError
}

// SetServerToolCallbacks sets callbacks for server-side tool execution notifications
func (ws *WebSocketConn) SetServerToolCallbacks(
	onStart func(protocol.ToolCall) error,
	onEnd func(protocol.ToolCall, protocol.ToolResult) error,
) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	ws.onServerToolStart = onStart
	ws.onServerToolEnd = onEnd
}

// Start starts a task
func (ws *WebSocketConn) Start(task string, projectInfo *protocol.ProjectInfo) error {
	payload := StartPayload{
		Task:        task,
		ProjectInfo: projectInfo,
	}

	return ws.send("start", payload)
}

// Listen starts listening for server messages
func (ws *WebSocketConn) Listen(ctx context.Context) error {
	done := make(chan error, 1)

	go func() {
		for {
			var msg WSServerMessage
			if err := ws.conn.ReadJSON(&msg); err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					done <- nil
					return
				}
				done <- fmt.Errorf("WebSocket read error: %w", err)
				return
			}

			if err := ws.handleMessage(msg); err != nil {
				done <- err
				return
			}
		}
	}()

	select {
	case <-ctx.Done():
		ws.Cancel()
		return ctx.Err()
	case err := <-done:
		return err
	}
}

// handleMessage processes a server message
func (ws *WebSocketConn) handleMessage(msg WSServerMessage) error {
	switch msg.Type {
	case "agent_message":
		var payload AgentMessagePayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return fmt.Errorf("failed to unmarshal agent message: %w", err)
		}

		if ws.onMessage != nil {
			protoMsg := protocol.Message{
				Agent:   payload.Agent,
				Content: payload.Content,
			}
			if err := ws.onMessage(protoMsg); err != nil {
				return err
			}
		}

	case "tool_call":
		var payload ToolCallPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return fmt.Errorf("failed to unmarshal tool call: %w", err)
		}

		toolCall := protocol.ToolCall{
			ID:        payload.CallID,
			ToolName:  payload.ToolName,
			Arguments: payload.Arguments,
		}

		if payload.ServerSide {
			// Server-side tool - just notify for display, don't execute
			if ws.onServerToolStart != nil {
				if err := ws.onServerToolStart(toolCall); err != nil {
					return err
				}
			}
			// Store pending tool call to match with result later
			ws.mu.Lock()
			ws.pendingServerTools[payload.CallID] = toolCall
			ws.mu.Unlock()
		} else {
			// Client-side tool - execute and send result back
			if ws.onToolCall != nil {
				result, err := ws.onToolCall(toolCall)
				if err != nil {
					return ws.SendToolResult(payload.CallID, nil, err)
				}
				return ws.SendToolResult(payload.CallID, result.Result, nil)
			}
		}

	case "tool_result":
		// Server-side tool result notification
		var payload ToolResultPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return fmt.Errorf("failed to unmarshal tool result: %w", err)
		}

		// Get the corresponding tool call
		ws.mu.Lock()
		toolCall, found := ws.pendingServerTools[payload.CallID]
		delete(ws.pendingServerTools, payload.CallID)
		ws.mu.Unlock()

		if found && ws.onServerToolEnd != nil {
			result := protocol.ToolResult{
				CallID: payload.CallID,
				Result: payload.Result,
				Error:  payload.Error,
			}
			if err := ws.onServerToolEnd(toolCall, result); err != nil {
				return err
			}
		}

	case "complete":
		var payload CompletePayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return fmt.Errorf("failed to unmarshal complete: %w", err)
		}

		if ws.onComplete != nil {
			return ws.onComplete(payload.Summary, payload.Artifacts)
		}

	case "error":
		var payload ErrorPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return fmt.Errorf("failed to unmarshal error: %w", err)
		}

		if ws.onError != nil {
			return ws.onError(fmt.Errorf("server error: %s", payload.Message))
		}

	case "pong":
		// Ignore pong messages

	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}

	return nil
}

// SendToolResult sends a tool result back to the server
func (ws *WebSocketConn) SendToolResult(callID string, result json.RawMessage, err error) error {
	payload := ToolResultPayload{
		CallID: callID,
		Result: result,
	}

	if err != nil {
		payload.Error = err.Error()
	}

	return ws.send("tool_result", payload)
}

// Cancel cancels the current task
func (ws *WebSocketConn) Cancel() error {
	return ws.send("cancel", nil)
}

// Close closes the WebSocket connection
func (ws *WebSocketConn) Close() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.conn == nil {
		return nil
	}

	// Send close message
	msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
	_ = ws.conn.WriteMessage(websocket.CloseMessage, msg)

	return ws.conn.Close()
}

// send sends a message to the server
func (ws *WebSocketConn) send(msgType string, payload interface{}) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	var payloadBytes json.RawMessage
	if payload != nil {
		var err error
		payloadBytes, err = json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
	}

	msg := WSClientMessage{
		Type:    msgType,
		Payload: payloadBytes,
	}

	if err := ws.conn.WriteJSON(msg); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}
