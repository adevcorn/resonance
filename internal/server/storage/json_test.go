package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONStorage(t *testing.T) {
	// Create temporary directory for testing
	tmpDir := t.TempDir()

	t.Run("NewJSONStorage creates directories", func(t *testing.T) {
		storage, err := NewJSONStorage(tmpDir)
		require.NoError(t, err)
		require.NotNil(t, storage)

		// Verify sessions directory was created
		sessionsDir := filepath.Join(tmpDir, "sessions")
		info, err := os.Stat(sessionsDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("CreateSession and GetSession", func(t *testing.T) {
		storage, err := NewJSONStorage(tmpDir)
		require.NoError(t, err)

		session := &protocol.Session{
			ID:          "test_session_1",
			ProjectPath: "/test/project",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			State:       protocol.SessionStateActive,
			Messages:    []protocol.Message{},
			ActiveTeam:  []string{"coordinator"},
		}

		// Create session
		err = storage.CreateSession(session)
		require.NoError(t, err)

		// Get session
		retrieved, err := storage.GetSession("test_session_1")
		require.NoError(t, err)
		assert.Equal(t, session.ID, retrieved.ID)
		assert.Equal(t, session.ProjectPath, retrieved.ProjectPath)
		assert.Equal(t, session.State, retrieved.State)
	})

	t.Run("CreateSession duplicate returns error", func(t *testing.T) {
		storage, err := NewJSONStorage(tmpDir)
		require.NoError(t, err)

		session := &protocol.Session{
			ID:          "test_session_2",
			ProjectPath: "/test/project",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			State:       protocol.SessionStateActive,
		}

		// Create once
		err = storage.CreateSession(session)
		require.NoError(t, err)

		// Try to create again
		err = storage.CreateSession(session)
		assert.Error(t, err)
	})

	t.Run("UpdateSession", func(t *testing.T) {
		storage, err := NewJSONStorage(tmpDir)
		require.NoError(t, err)

		session := &protocol.Session{
			ID:          "test_session_3",
			ProjectPath: "/test/project",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			State:       protocol.SessionStateActive,
		}

		// Create session
		err = storage.CreateSession(session)
		require.NoError(t, err)

		// Update session
		session.State = protocol.SessionStateCompleted
		session.ActiveTeam = []string{"coordinator", "developer"}
		err = storage.UpdateSession(session)
		require.NoError(t, err)

		// Verify update
		retrieved, err := storage.GetSession("test_session_3")
		require.NoError(t, err)
		assert.Equal(t, protocol.SessionStateCompleted, retrieved.State)
		assert.Equal(t, 2, len(retrieved.ActiveTeam))
	})

	t.Run("DeleteSession", func(t *testing.T) {
		storage, err := NewJSONStorage(tmpDir)
		require.NoError(t, err)

		session := &protocol.Session{
			ID:          "test_session_4",
			ProjectPath: "/test/project",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			State:       protocol.SessionStateActive,
		}

		// Create session
		err = storage.CreateSession(session)
		require.NoError(t, err)

		// Delete session
		err = storage.DeleteSession("test_session_4")
		require.NoError(t, err)

		// Verify deletion
		_, err = storage.GetSession("test_session_4")
		assert.Error(t, err)
	})

	t.Run("ListSessions", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewJSONStorage(tmpDir)
		require.NoError(t, err)

		// Create multiple sessions
		for i := 0; i < 3; i++ {
			session := &protocol.Session{
				ID:          "list_test_" + string(rune('a'+i)),
				ProjectPath: "/test/project",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
				State:       protocol.SessionStateActive,
			}
			err = storage.CreateSession(session)
			require.NoError(t, err)
		}

		// List sessions
		sessions, err := storage.ListSessions()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(sessions), 3)
	})

	t.Run("ListSessionsByProject", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewJSONStorage(tmpDir)
		require.NoError(t, err)

		// Create sessions for different projects
		project1Sessions := []string{"proj1_a", "proj1_b"}
		for _, id := range project1Sessions {
			session := &protocol.Session{
				ID:          id,
				ProjectPath: "/test/project1",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
				State:       protocol.SessionStateActive,
			}
			err = storage.CreateSession(session)
			require.NoError(t, err)
		}

		session := &protocol.Session{
			ID:          "proj2_a",
			ProjectPath: "/test/project2",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			State:       protocol.SessionStateActive,
		}
		err = storage.CreateSession(session)
		require.NoError(t, err)

		// List sessions for project1
		sessions, err := storage.ListSessionsByProject("/test/project1")
		require.NoError(t, err)
		assert.Equal(t, 2, len(sessions))

		// List sessions for project2
		sessions, err = storage.ListSessionsByProject("/test/project2")
		require.NoError(t, err)
		assert.Equal(t, 1, len(sessions))
	})
}
