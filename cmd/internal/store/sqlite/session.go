package store

import (
	"context"
	"database/sql"
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

func (s *SQLiteStore) IsSessionImplemented(ctx context.Context, sessionID string) (bool, error){
	const q=`SELECT 1 FROM turn 
		INNER JOIN call ON turn.turn_id = call.turn_id 
		WHERE turn.role = "assistant" AND call.name = "RunImplementation" AND turn.session_id = ?`;

	var one int

	err := s.db.QueryRowContext(ctx, q, sessionID).Scan(&one)

	if err == sql.ErrNoRows{
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("store: is session implemented %s: %w", sessionID, err)
	}

	return true, nil;
}