package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/gorilla/websocket"
)

// Client connects to the Ensemble server
type Client struct {
	serverURL  string
	httpClient *http.Client
}

// AgentSummary represents a brief agent summary
type AgentSummary struct {
	Name         string   `json:"name"`
	DisplayName  string   `json:"display_name"`
	Description  string   `json:"description"`
	Capabilities []string `json:"capabilities"`
}

// AgentDetail represents full agent details
type AgentDetail struct {
	AgentSummary
	SystemPrompt string                 `json:"system_prompt"`
	Model        map[string]interface{} `json:"model"`
	Tools        map[string]interface{} `json:"tools"`
}

// NewClient creates a new client instance
func NewClient(serverURL string) *Client {
	return &Client{
		serverURL:  serverURL,
		httpClient: &http.Client{},
	}
}

// CreateSession creates a new session on the server
func (c *Client) CreateSession(ctx context.Context, projectPath string) (*protocol.Session, error) {
	reqBody := map[string]string{
		"project_path": projectPath,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.serverURL+"/api/sessions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errResp map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("server error (%d): %s", resp.StatusCode, errResp["error"])
	}

	var session protocol.Session
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &session, nil
}

// GetSession retrieves a session by ID
func (c *Client) GetSession(ctx context.Context, sessionID string) (*protocol.Session, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.serverURL+"/api/sessions/"+sessionID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("server error (%d): %s", resp.StatusCode, errResp["error"])
	}

	var session protocol.Session
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &session, nil
}

// DeleteSession deletes a session
func (c *Client) DeleteSession(ctx context.Context, sessionID string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", c.serverURL+"/api/sessions/"+sessionID, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		var errResp map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("server error (%d): %s", resp.StatusCode, errResp["error"])
	}

	return nil
}

// ListSessions lists all sessions
func (c *Client) ListSessions(ctx context.Context) ([]*protocol.Session, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.serverURL+"/api/sessions", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("server error (%d): %s", resp.StatusCode, errResp["error"])
	}

	var result struct {
		Sessions []*protocol.Session `json:"sessions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Sessions, nil
}

// ListAgents lists available agents
func (c *Client) ListAgents(ctx context.Context) ([]AgentSummary, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.serverURL+"/api/agents", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("server error (%d): %s", resp.StatusCode, errResp["error"])
	}

	var result struct {
		Agents []AgentSummary `json:"agents"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Agents, nil
}

// GetAgent retrieves agent details
func (c *Client) GetAgent(ctx context.Context, name string) (*AgentDetail, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.serverURL+"/api/agents/"+name, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("server error (%d): %s", resp.StatusCode, errResp["error"])
	}

	var agent AgentDetail
	if err := json.NewDecoder(resp.Body).Decode(&agent); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &agent, nil
}

// ConnectWebSocket connects to the session WebSocket
func (c *Client) ConnectWebSocket(ctx context.Context, sessionID string) (*WebSocketConn, error) {
	// Parse server URL
	u, err := url.Parse(c.serverURL)
	if err != nil {
		return nil, fmt.Errorf("invalid server URL: %w", err)
	}

	// Build WebSocket URL
	wsURL := fmt.Sprintf("ws://%s/api/sessions/%s/run", u.Host, sessionID)
	if u.Scheme == "https" {
		wsURL = fmt.Sprintf("wss://%s/api/sessions/%s/run", u.Host, sessionID)
	}

	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	return NewWebSocketConn(conn), nil
}
