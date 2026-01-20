package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/adevcorn/ensemble/internal/server/agent"
	"github.com/adevcorn/ensemble/internal/server/orchestration"
	"github.com/adevcorn/ensemble/internal/server/provider"
	"github.com/adevcorn/ensemble/internal/server/provider/mock"
	"github.com/adevcorn/ensemble/internal/server/storage"
	"github.com/adevcorn/ensemble/internal/server/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestServer(t *testing.T) *Server {
	// Create mock provider
	mockProvider := mock.NewMockProvider("mock", []string{"test response"})
	registry := provider.NewRegistry()
	registry.Register(mockProvider)

	// Create agent pool
	agentPool := agent.NewPool(registry)

	// Load coordinator agent (required by engine)
	coordinatorAgent := &protocol.AgentDefinition{
		Name:         "coordinator",
		DisplayName:  "Coordinator",
		Description:  "Coordinator agent",
		SystemPrompt: "You are a coordinator",
		Capabilities: []string{"coordination"},
		Model: protocol.ModelConfig{
			Provider:    "mock",
			Name:        "mock-model",
			Temperature: 0.5,
			MaxTokens:   1000,
		},
		Tools: protocol.ToolsConfig{
			Allowed: []string{"collaborate", "assemble_team"},
			Denied:  []string{},
		},
	}

	// Load test agent
	testAgent := &protocol.AgentDefinition{
		Name:         "test-agent",
		DisplayName:  "Test Agent",
		Description:  "Test agent for testing",
		SystemPrompt: "You are a test agent",
		Capabilities: []string{"testing"},
		Model: protocol.ModelConfig{
			Provider:    "mock",
			Name:        "mock-model",
			Temperature: 0.5,
			MaxTokens:   1000,
		},
		Tools: protocol.ToolsConfig{
			Allowed: []string{"collaborate"},
			Denied:  []string{},
		},
	}

	err := agentPool.Load([]*protocol.AgentDefinition{coordinatorAgent, testAgent})
	require.NoError(t, err)

	// Create storage
	tmpDir := t.TempDir()
	jsonStorage, err := storage.NewJSONStorage(tmpDir)
	require.NoError(t, err)
	sessionManager := storage.NewSessionManager(jsonStorage)

	// Create tool registry
	toolRegistry := tool.NewRegistry()

	// Create engine
	engine, err := orchestration.NewEngine(agentPool, toolRegistry, nil, nil)
	require.NoError(t, err)

	// Create server
	return NewServer(sessionManager, agentPool, engine, toolRegistry)
}

func TestSessionEndpoints(t *testing.T) {
	server := setupTestServer(t)

	t.Run("POST /api/sessions creates session", func(t *testing.T) {
		reqBody := CreateSessionRequest{
			ProjectPath: "/test/project",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/sessions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.Handler().ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp CreateSessionResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.ID)
		assert.Equal(t, "/test/project", resp.ProjectPath)
		assert.Equal(t, "active", resp.State)
	})

	t.Run("POST /api/sessions with empty project_path returns error", func(t *testing.T) {
		reqBody := CreateSessionRequest{
			ProjectPath: "",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/sessions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.Handler().ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("GET /api/sessions/:id returns session", func(t *testing.T) {
		// Create session first
		session, err := server.sessionManager.Create(context.Background(), "/test/project")
		require.NoError(t, err)

		req := httptest.NewRequest("GET", "/api/sessions/"+session.ID, nil)
		w := httptest.NewRecorder()

		server.Handler().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp protocol.Session
		err = json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, session.ID, resp.ID)
	})

	t.Run("GET /api/sessions/:id with invalid ID returns 404", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/sessions/invalid_id", nil)
		w := httptest.NewRecorder()

		server.Handler().ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("DELETE /api/sessions/:id deletes session", func(t *testing.T) {
		// Create session first
		session, err := server.sessionManager.Create(context.Background(), "/test/project")
		require.NoError(t, err)

		req := httptest.NewRequest("DELETE", "/api/sessions/"+session.ID, nil)
		w := httptest.NewRecorder()

		server.Handler().ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify deletion
		_, err = server.sessionManager.Get(context.Background(), session.ID)
		assert.Error(t, err)
	})

	t.Run("GET /api/sessions lists sessions", func(t *testing.T) {
		// Create a few sessions
		for i := 0; i < 3; i++ {
			_, err := server.sessionManager.Create(context.Background(), "/test/project")
			require.NoError(t, err)
		}

		req := httptest.NewRequest("GET", "/api/sessions", nil)
		w := httptest.NewRecorder()

		server.Handler().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Contains(t, resp, "sessions")
	})
}

func TestAgentEndpoints(t *testing.T) {
	server := setupTestServer(t)

	t.Run("GET /api/agents lists agents", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/agents", nil)
		w := httptest.NewRecorder()

		server.Handler().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string][]AgentSummary
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Contains(t, resp, "agents")
		assert.GreaterOrEqual(t, len(resp["agents"]), 1)
	})

	t.Run("GET /api/agents/:name returns agent details", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/agents/test-agent", nil)
		w := httptest.NewRecorder()

		server.Handler().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp AgentDetail
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "test-agent", resp.Name)
		assert.Equal(t, "Test Agent", resp.DisplayName)
	})

	t.Run("GET /api/agents/:name with invalid name returns 404", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/agents/nonexistent", nil)
		w := httptest.NewRecorder()

		server.Handler().ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestHealthEndpoint(t *testing.T) {
	server := setupTestServer(t)

	t.Run("GET /api/health returns OK", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/health", nil)
		w := httptest.NewRecorder()

		server.Handler().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]string
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "ok", resp["status"])
	})
}
