package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	_ "modernc.org/sqlite" // pure-Go SQLite driver registration
)

// SQLiteStore is a pure-Go SQLite-backed Store.
type SQLiteStore struct {
	db *sql.DB
}

// compile-time check that SQLiteStore satisfies the Store interface.
var _ Store = (*SQLiteStore)(nil)

// NewSQLite opens (creating if needed) the database at path, verifies the
// connection, and runs migrations.
func NewSQLite(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	if err := RunMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("store: run migrations: %w", err)
	}
	return &SQLiteStore{db: db}, nil
}

// Close releases the underlying connection pool.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// CreateSession inserts a new session for the given workspace path with
// approved_proposal false (Planning mode) and returns it.
func (s *SQLiteStore) CreateSession(ctx context.Context, rootDir string) (Session, error) {
	sessionID := uuid.NewString()
	const q = `INSERT INTO session (session_id, root_dir, approved_proposal) VALUES (?, ?, 0)`
	if _, err := s.db.ExecContext(ctx, q, sessionID, rootDir); err != nil {
		return Session{}, fmt.Errorf("store: create session: %w", err)
	}
	return s.GetSession(ctx, sessionID)
}

// GetSession loads a session by ID.
func (s *SQLiteStore) GetSession(ctx context.Context, sessionID string) (Session, error) {
	const q = `SELECT session_id, root_dir, approved_proposal, created_at FROM session WHERE session_id = ?`
	var sess Session
	var approved int
	err := s.db.QueryRowContext(ctx, q, sessionID).Scan(&sess.SessionID, &sess.RootDir, &approved, &sess.CreatedAt)
	if err != nil {
		return Session{}, fmt.Errorf("store: get session %s: %w", sessionID, err)
	}
	sess.ApprovedProposal = approved != 0
	return sess, nil
}

// SetApprovedProposal flips a session's approval boolean. This is the one and
// only persisted signal that moves the agent from Planning to Executing.
func (s *SQLiteStore) SetApprovedProposal(ctx context.Context, sessionID string, v bool) error {
	const q = `UPDATE session SET approved_proposal = ? WHERE session_id = ?`
	if _, err := s.db.ExecContext(ctx, q, boolToInt(v), sessionID); err != nil {
		return fmt.Errorf("store: set approved_proposal for session %s: %w", sessionID, err)
	}
	return nil
}

// SaveTurn inserts a turn, generating and returning its turn_id.
func (s *SQLiteStore) SaveTurn(ctx context.Context, t Turn) (string, error) {
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
func (s *SQLiteStore) LastAgentTurn(ctx context.Context, sessionID string) (*Turn, error) {
	const q = `SELECT turn_id, session_id, role, output_text, response_id, created_at
	           FROM turn
	           WHERE session_id = ? AND role = 'agent'
	           ORDER BY created_at DESC, rowid DESC
	           LIMIT 1`
	var t Turn
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

// SaveCall inserts a call row with result NULL. It is the INSERT half of the
// one-row-per-call_id lifecycle; UpdateCallResult is the UPDATE half.
func (s *SQLiteStore) SaveCall(ctx context.Context, c Call) error {
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
func (s *SQLiteStore) PendingCall(ctx context.Context, sessionID string) (*Call, error) {
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
	var c Call
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
