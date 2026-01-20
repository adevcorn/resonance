package api

import (
	"net/http"

	"github.com/adevcorn/ensemble/internal/server/agent"
	"github.com/adevcorn/ensemble/internal/server/orchestration"
	"github.com/adevcorn/ensemble/internal/server/storage"
	"github.com/adevcorn/ensemble/internal/server/tool"
	"github.com/gorilla/mux"
)

// Server is the HTTP API server
type Server struct {
	router         *mux.Router
	sessionManager *storage.SessionManager
	agentPool      *agent.Pool
	engine         *orchestration.Engine
	registry       *tool.Registry
}

// NewServer creates a new API server
func NewServer(
	sessionManager *storage.SessionManager,
	agentPool *agent.Pool,
	engine *orchestration.Engine,
	registry *tool.Registry,
) *Server {
	s := &Server{
		router:         mux.NewRouter(),
		sessionManager: sessionManager,
		agentPool:      agentPool,
		engine:         engine,
		registry:       registry,
	}

	s.SetupRoutes()

	return s
}

// SetupRoutes configures all API routes
func (s *Server) SetupRoutes() {
	// Apply middleware
	s.router.Use(loggingMiddleware)
	s.router.Use(recoveryMiddleware)
	s.router.Use(corsMiddleware)
	s.router.Use(requestIDMiddleware)

	// API routes
	api := s.router.PathPrefix("/api").Subrouter()

	// Session endpoints
	api.HandleFunc("/sessions", s.handleCreateSession).Methods("POST")
	api.HandleFunc("/sessions", s.handleListSessions).Methods("GET")
	api.HandleFunc("/sessions/{id}", s.handleGetSession).Methods("GET")
	api.HandleFunc("/sessions/{id}", s.handleDeleteSession).Methods("DELETE")

	// Agent endpoints
	api.HandleFunc("/agents", s.handleListAgents).Methods("GET")
	api.HandleFunc("/agents/{name}", s.handleGetAgent).Methods("GET")

	// WebSocket endpoint for running tasks
	api.HandleFunc("/sessions/{id}/ws", s.handleWebSocket).Methods("GET")

	// Health check
	api.HandleFunc("/health", s.handleHealth).Methods("GET")
}

// Handler returns the HTTP handler
func (s *Server) Handler() http.Handler {
	return s.router
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}
