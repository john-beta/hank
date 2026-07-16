package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/john-beta/hank/cmd/internal/store"
)

// SaveCall inserts a call row with result NULL. It is the INSERT half of the
// one-row-per-call_id lifecycle; UpdateCallResult is the UPDATE half.
func (s *SQLiteStore) SaveCall(ctx context.Context, c store.Call) error {
	const q = `INSERT INTO call (call_id, turn_id, name, args, result, auto_re_feed)
	           VALUES (?, ?, ?, ?, NULL, ?)`
	_, err := s.db.ExecContext(ctx, q, c.CallID, c.TurnID, c.Name, c.Args, boolToInt(c.AutoReFeed))
	if err != nil {
		return fmt.Errorf("store: save call %s: %w", c.CallID, err)
	}
	return nil
}

// UpdateCallResult fills in the result of an already-inserted call row.
func (s *SQLiteStore) UpdateCallResult(ctx context.Context, callID string, result string) error {
	const q = `UPDATE call SET result = ? WHERE call_id = ?`
	if _, err := s.db.ExecContext(ctx, q, result, callID); err != nil {
		return fmt.Errorf("store: update call result %s: %w", callID, err)
	}
	return nil
}

// PendingCall returns the single unresolved, non-auto call attached to the
// session's latest agent turn, or nil, nil if none is outstanding.
func (s *SQLiteStore) PendingCall(ctx context.Context, sessionID string) (*store.Call, error) {
	const q = `SELECT call_id, turn_id, name, args, result, auto_re_feed
	           FROM call
	           WHERE turn_id = (
	               SELECT turn_id FROM turn
	               WHERE session_id = ? AND role = 'agent'
	               ORDER BY created_at DESC, rowid DESC
	               LIMIT 1
	           )
	           AND result IS NULL AND auto_re_feed = 0
	           LIMIT 1`
	var c store.Call
	var auto int
	err := s.db.QueryRowContext(ctx, q, sessionID).Scan(
		&c.CallID, &c.TurnID, &c.Name, &c.Args, &c.Result, &auto,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("store: pending call for session %s: %w", sessionID, err)
	}
	c.AutoReFeed = auto != 0
	return &c, nil
}

// boolToInt maps a Go bool to the 0/1 integer SQLite stores for BOOL columns.
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
