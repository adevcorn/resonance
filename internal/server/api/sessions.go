package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// CreateSessionRequest is the request body for creating a session
type CreateSessionRequest struct {
	ProjectPath string `json:"project_path"`
}

// CreateSessionResponse is the response for creating a session
type CreateSessionResponse struct {
	ID          string `json:"id"`
	ProjectPath string `json:"project_path"`
	CreatedAt   string `json:"created_at"`
	State       string `json:"state"`
}

// handleCreateSession creates a new session
func (s *Server) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	var req CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate project path
	if req.ProjectPath == "" {
		respondError(w, http.StatusBadRequest, "project_path is required")
		return
	}

	// Create session
	session, err := s.sessionManager.Create(r.Context(), req.ProjectPath)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	// Build response
	resp := CreateSessionResponse{
		ID:          session.ID,
		ProjectPath: session.ProjectPath,
		CreatedAt:   session.CreatedAt.Format(rfc3339Milli),
		State:       string(session.State),
	}

	respondJSON(w, http.StatusCreated, resp)
}

// handleGetSession retrieves a session by ID
func (s *Server) handleGetSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	session, err := s.sessionManager.Get(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Session not found")
		return
	}

	respondJSON(w, http.StatusOK, session)
}

// handleDeleteSession deletes a session
func (s *Server) handleDeleteSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := s.sessionManager.Delete(r.Context(), id); err != nil {
		respondError(w, http.StatusNotFound, "Session not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleListSessions lists all sessions
func (s *Server) handleListSessions(w http.ResponseWriter, r *http.Request) {
	// Check for project_path query parameter
	projectPath := r.URL.Query().Get("project_path")

	var sessions interface{}
	var err error

	if projectPath != "" {
		sessions, err = s.sessionManager.ListByProject(r.Context(), projectPath)
	} else {
		sessions, err = s.sessionManager.List(r.Context())
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list sessions")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"sessions": sessions,
	})
}

// rfc3339Milli is the time format for API responses
const rfc3339Milli = "2006-01-02T15:04:05.000Z07:00"
