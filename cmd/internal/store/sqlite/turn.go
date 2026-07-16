package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"github.com/john-beta/hank/cmd/internal/store"
)

// SaveTurn inserts a turn, generating and returning its turn_id.
func (s *SQLiteStore) SaveTurn(ctx context.Context, t store.Turn) (string, error) {
	const q = `INSERT INTO turn (turn_id, session_id, role, output_text, response_id)
	           VALUES (?, ?, ?, ?, ?)`
	turnID := uuid.NewString()
	_, err := s.db.ExecContext(ctx, q, turnID, t.SessionID, t.Role, t.OutputText, t.ResponseID)
	if err != nil {
		return "", fmt.Errorf("store: save turn for session %s: %w", t.SessionID, err)
	}
	return turnID, nil
}

// LastAgentTurn returns the most recent agent turn for a session, or nil, nil
// if the session has no agent turn yet. created_at has only one-second
// resolution, and a single loop can persist several agent turns within the
// same second (intermediate tool-call turns plus the final text turn), so
// rowid (monotonic per INSERT) breaks the tie in favor of the last-saved turn.
func (s *SQLiteStore) LastAgentTurn(ctx context.Context, sessionID string) (*store.Turn, error) {
	const q = `SELECT turn_id, session_id, role, output_text, response_id, created_at
	           FROM turn
	           WHERE session_id = ? AND role = 'agent'
	           ORDER BY created_at DESC, rowid DESC
	           LIMIT 1`
	var t store.Turn
	err := s.db.QueryRowContext(ctx, q, sessionID).Scan(
		&t.TurnID, &t.SessionID, &t.Role, &t.OutputText, &t.ResponseID, &t.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("store: last agent turn for session %s: %w", sessionID, err)
	}
	return &t, nil
}
