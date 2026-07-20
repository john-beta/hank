package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/john-beta/hank/cmd/internal/store"
)

func (s *SQLiteStore) CreateSession(ctx context.Context, rootDir string) (store.Session, error) {
	sessionID := uuid.NewString()
	const q = `INSERT INTO session (session_id, root_dir) VALUES (?, ?) RETURNING session_id, root_dir, created_at`
	var sess store.Session
	if err := s.db.QueryRowContext(ctx, q, sessionID, rootDir).Scan(
		&sess.SessionID, &sess.RootDir, &sess.CreatedAt,
	); err != nil {
		return store.Session{}, fmt.Errorf("store: create session: %w", err)
	}
	return sess, nil
}

func (s *SQLiteStore) GetSession(ctx context.Context, sessionID string) (store.Session, error) {
	const q = `SELECT session_id, root_dir, created_at FROM session WHERE session_id = ?`
	var sess store.Session
	err := s.db.QueryRowContext(ctx, q, sessionID).Scan(&sess.SessionID, &sess.RootDir, &sess.CreatedAt)
	if err != nil {
		return store.Session{}, fmt.Errorf("store: get session %s: %w", sessionID, err)
	}
	return sess, nil
}
