package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"github.com/john-beta/hank/cmd/internal/store"
)

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

// LastAssistantTurn returns the most recent assistant turn, or nil, nil if none yet.
// created_at has one-second resolution and a loop can persist several assistant
// turns within one second, so rowid (monotonic per INSERT) breaks the tie.
func (s *SQLiteStore) LastAssistantTurn(ctx context.Context, sessionID string) (*store.Turn, error) {
	const q = `SELECT turn_id, session_id, role, output_text, response_id, created_at
	           FROM turn
	           WHERE session_id = ? AND role = 'assistant'
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
		return nil, fmt.Errorf("store: last assistant turn for session %s: %w", sessionID, err)
	}
	return &t, nil
}
