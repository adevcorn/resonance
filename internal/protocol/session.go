package protocol

import "time"

// SessionState represents the current state of a session
type SessionState string

const (
	SessionStateActive    SessionState = "active"
	SessionStatePaused    SessionState = "paused"
	SessionStateCompleted SessionState = "completed"
	SessionStateError     SessionState = "error"
)

// Session represents a conversation session tied to a project
type Session struct {
	ID          string         `json:"id"`
	ProjectPath string         `json:"project_path"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	State       SessionState   `json:"state"`
	Messages    []Message      `json:"messages"`
	ActiveTeam  []string       `json:"active_team"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}
