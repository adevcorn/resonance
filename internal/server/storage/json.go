package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/adevcorn/ensemble/internal/protocol"
)

// JSONStorage implements file-based JSON storage
type JSONStorage struct {
	basePath string
	mu       sync.RWMutex // Mutex for thread-safe file operations
}

// NewJSONStorage creates a new JSON storage instance
func NewJSONStorage(basePath string) (*JSONStorage, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	// Create sessions subdirectory
	sessionsDir := filepath.Join(basePath, "sessions")
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create sessions directory: %w", err)
	}

	return &JSONStorage{
		basePath: basePath,
	}, nil
}

// CreateSession creates a new session
func (s *JSONStorage) CreateSession(session *protocol.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sessionPath := s.sessionPath(session.ID)

	// Check if session already exists
	if _, err := os.Stat(sessionPath); err == nil {
		return fmt.Errorf("session %s already exists", session.ID)
	}

	return s.writeSessionAtomic(sessionPath, session)
}

// GetSession retrieves a session by ID
func (s *JSONStorage) GetSession(id string) (*protocol.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessionPath := s.sessionPath(id)

	data, err := os.ReadFile(sessionPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("session %s not found", id)
		}
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	var session protocol.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// UpdateSession updates an existing session
func (s *JSONStorage) UpdateSession(session *protocol.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sessionPath := s.sessionPath(session.ID)

	// Check if session exists
	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		return fmt.Errorf("session %s not found", session.ID)
	}

	return s.writeSessionAtomic(sessionPath, session)
}

// DeleteSession deletes a session
func (s *JSONStorage) DeleteSession(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sessionPath := s.sessionPath(id)

	if err := os.Remove(sessionPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("session %s not found", id)
		}
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// ListSessions lists all sessions
func (s *JSONStorage) ListSessions() ([]*protocol.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessionsDir := filepath.Join(s.basePath, "sessions")

	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read sessions directory: %w", err)
	}

	var sessions []*protocol.Session
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		sessionPath := filepath.Join(sessionsDir, entry.Name())
		data, err := os.ReadFile(sessionPath)
		if err != nil {
			// Skip files that can't be read
			continue
		}

		var session protocol.Session
		if err := json.Unmarshal(data, &session); err != nil {
			// Skip files that can't be parsed
			continue
		}

		sessions = append(sessions, &session)
	}

	return sessions, nil
}

// ListSessionsByProject lists sessions for a specific project
func (s *JSONStorage) ListSessionsByProject(projectPath string) ([]*protocol.Session, error) {
	allSessions, err := s.ListSessions()
	if err != nil {
		return nil, err
	}

	var projectSessions []*protocol.Session
	for _, session := range allSessions {
		if session.ProjectPath == projectPath {
			projectSessions = append(projectSessions, session)
		}
	}

	return projectSessions, nil
}

// sessionPath returns the file path for a session
func (s *JSONStorage) sessionPath(id string) string {
	return filepath.Join(s.basePath, "sessions", fmt.Sprintf("%s.json", id))
}

// writeSessionAtomic writes a session to disk atomically
// This prevents corruption if the process crashes during write
func (s *JSONStorage) writeSessionAtomic(path string, session *protocol.Session) error {
	// Marshal session to JSON with indentation for readability
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	// Write to temporary file first
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Atomically rename temp file to final path
	if err := os.Rename(tmpPath, path); err != nil {
		// Clean up temp file on error
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}
