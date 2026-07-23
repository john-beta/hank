package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/john-beta/hank/cmd/internal/store"
)

// SaveCall inserts a call row with result NULL — the INSERT half of the
// one-row-per-call_id lifecycle (UpdateCallResult is the UPDATE half).
func (s *SQLiteStore) SaveCall(ctx context.Context, c store.Call) error {
	const q = `INSERT INTO call (call_id, turn_id, name, args, result, auto_re_feed)
	           VALUES (?, ?, ?, ?, NULL, ?)`
	_, err := s.db.ExecContext(ctx, q, c.CallID, c.TurnID, c.Name, c.Args, c.AutoReFeed)
	if err != nil {
		return fmt.Errorf("store: save call %s: %w", c.CallID, err)
	}
	return nil
}

func (s *SQLiteStore) UpdateCallResult(ctx context.Context, callID string, result string) error {
	const q = `UPDATE call SET result = ? WHERE call_id = ?`
	if _, err := s.db.ExecContext(ctx, q, result, callID); err != nil {
		return fmt.Errorf("store: update call result %s: %w", callID, err)
	}
	return nil
}

// IsPendingCall reports whether callID names an unresolved, non-auto tool
// call — one still awaiting a client-supplied result. It is the desync guard
// for tool-result submission: an unknown, already-resolved, or auto call
// reports false.
func (s *SQLiteStore) IsPendingCall(ctx context.Context, callID string) (bool, error) {
	const q = `SELECT 1 FROM call
	           WHERE call_id = ? AND result IS NULL AND auto_re_feed = 0`
	var one int
	err := s.db.QueryRowContext(ctx, q, callID).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("store: is pending call %s: %w", callID, err)
	}
	return true, nil
}

func (s *SQLiteStore) PendingProposeStructureByCallID(ctx context.Context, callID string) (string, error) {
	const q = `SELECT args FROM call WHERE name = 'ProposeStructure' AND call_id = ?`
	var rawArgs string
	if err := s.db.QueryRowContext(ctx, q, callID).Scan(&rawArgs); err != nil {
		return "", fmt.Errorf("store: get proposal args for call %s: %w", callID, err)
	}

	var args struct {
		ProposedWorkspaceEntries json.RawMessage `json:"proposed_workspace_entries"`
	}
	if err := json.Unmarshal([]byte(rawArgs), &args); err != nil {
		return "", fmt.Errorf("store: unmarshal proposal args for call %s: %w", callID, err)
	}
	return string(args.ProposedWorkspaceEntries), nil
}
