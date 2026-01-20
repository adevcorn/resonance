package storage

import (
	"github.com/adevcorn/ensemble/internal/protocol"
)

// Storage interface for persistence
type Storage interface {
	// Sessions
	CreateSession(session *protocol.Session) error
	GetSession(id string) (*protocol.Session, error)
	UpdateSession(session *protocol.Session) error
	DeleteSession(id string) error
	ListSessions() ([]*protocol.Session, error)
	ListSessionsByProject(projectPath string) ([]*protocol.Session, error)
}
