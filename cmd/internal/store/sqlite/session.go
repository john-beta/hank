package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/john-beta/hank/cmd/internal/store"
)

// CreateSession inserts a new session for the given workspace path and
// returns it.
func (s *SQLiteStore) CreateSession(ctx context.Context, rootDir string) (store.Session, error) {
	sessionID := uuid.NewString()
	const q = `INSERT INTO session (session_id, root_dir) VALUES (?, ?)`
	if _, err := s.db.ExecContext(ctx, q, sessionID, rootDir); err != nil {
		return store.Session{}, fmt.Errorf("store: create session: %w", err)
	}
	return s.GetSession(ctx, sessionID)
}

// GetSession loads a session by ID.
func (s *SQLiteStore) GetSession(ctx context.Context, sessionID string) (store.Session, error) {
	const q = `SELECT session_id, root_dir, created_at FROM session WHERE session_id = ?`
	var sess store.Session
	err := s.db.QueryRowContext(ctx, q, sessionID).Scan(&sess.SessionID, &sess.RootDir, &sess.CreatedAt)
	if err != nil {
		return store.Session{}, fmt.Errorf("store: get session %s: %w", sessionID, err)
	}
	return sess, nil
}
