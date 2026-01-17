package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/google/uuid"
)

// SessionManager manages session lifecycle
type SessionManager struct {
	storage Storage
}

// NewSessionManager creates a new session manager
func NewSessionManager(storage Storage) *SessionManager {
	return &SessionManager{
		storage: storage,
	}
}

// Create creates a new session
func (sm *SessionManager) Create(ctx context.Context, projectPath string) (*protocol.Session, error) {
	now := time.Now()

	session := &protocol.Session{
		ID:          generateSessionID(),
		ProjectPath: projectPath,
		CreatedAt:   now,
		UpdatedAt:   now,
		State:       protocol.SessionStateActive,
		Messages:    []protocol.Message{},
		ActiveTeam:  []string{},
		Metadata:    make(map[string]any),
	}

	if err := sm.storage.CreateSession(session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// Get retrieves a session by ID
func (sm *SessionManager) Get(ctx context.Context, id string) (*protocol.Session, error) {
	session, err := sm.storage.GetSession(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return session, nil
}

// Update updates an existing session
func (sm *SessionManager) Update(ctx context.Context, session *protocol.Session) error {
	session.UpdatedAt = time.Now()

	if err := sm.storage.UpdateSession(session); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// Delete deletes a session
func (sm *SessionManager) Delete(ctx context.Context, id string) error {
	if err := sm.storage.DeleteSession(id); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// List lists all sessions
func (sm *SessionManager) List(ctx context.Context) ([]*protocol.Session, error) {
	sessions, err := sm.storage.ListSessions()
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	return sessions, nil
}

// ListByProject lists sessions for a specific project
func (sm *SessionManager) ListByProject(ctx context.Context, projectPath string) ([]*protocol.Session, error) {
	sessions, err := sm.storage.ListSessionsByProject(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list project sessions: %w", err)
	}

	return sessions, nil
}

// AddMessage adds a message to a session
func (sm *SessionManager) AddMessage(ctx context.Context, sessionID string, message protocol.Message) error {
	session, err := sm.Get(ctx, sessionID)
	if err != nil {
		return err
	}

	// Set session ID on message if not already set
	if message.SessionID == "" {
		message.SessionID = sessionID
	}

	session.Messages = append(session.Messages, message)

	return sm.Update(ctx, session)
}

// SetActiveTeam sets the active team for a session
func (sm *SessionManager) SetActiveTeam(ctx context.Context, sessionID string, team []string) error {
	session, err := sm.Get(ctx, sessionID)
	if err != nil {
		return err
	}

	session.ActiveTeam = team

	return sm.Update(ctx, session)
}

// SetState sets the session state
func (sm *SessionManager) SetState(ctx context.Context, sessionID string, state protocol.SessionState) error {
	session, err := sm.Get(ctx, sessionID)
	if err != nil {
		return err
	}

	session.State = state

	return sm.Update(ctx, session)
}

// generateSessionID generates a unique session ID
func generateSessionID() string {
	return fmt.Sprintf("session_%s", uuid.New().String())
}
