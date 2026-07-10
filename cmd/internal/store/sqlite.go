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

// CreateSession inserts a new session in PhaseOne (phase 0) and returns it.
func (s *SQLiteStore) CreateSession(ctx context.Context) (Session, error) {
	sess := Session{SessionID: uuid.NewString(), Phase: 1}
	const q = `INSERT INTO session (session_id, phase) VALUES (?, ?)`
	if _, err := s.db.ExecContext(ctx, q, sess.SessionID, sess.Phase); err != nil {
		return Session{}, fmt.Errorf("store: create session: %w", err)
	}
	return s.GetSession(ctx, sess.SessionID)
}

// GetSession loads a session by ID.
func (s *SQLiteStore) GetSession(ctx context.Context, sessionID string) (Session, error) {
	const q = `SELECT session_id, phase, created_at FROM session WHERE session_id = ?`
	var sess Session
	err := s.db.QueryRowContext(ctx, q, sessionID).Scan(&sess.SessionID, &sess.Phase, &sess.CreatedAt)
	if err != nil {
		return Session{}, fmt.Errorf("store: get session %s: %w", sessionID, err)
	}
	return sess, nil
}

// SaveTurn inserts a turn, generating its turn_id.
func (s *SQLiteStore) SaveTurn(ctx context.Context, t Turn) error {
	const q = `INSERT INTO turn (turn_id, session_id, response_id, phase, role, content)
	           VALUES (?, ?, ?, ?, ?, ?)`
	turnID := uuid.NewString()
	_, err := s.db.ExecContext(ctx, q, turnID, t.SessionID, t.ResponseID, t.Phase, t.Role, t.Content)
	if err != nil {
		return fmt.Errorf("store: save turn for session %s: %w", t.SessionID, err)
	}
	return nil
}

// LastAgentTurn returns the most recent agent turn for a session, or nil, nil
// if the session has no agent turn yet. created_at has only one-second
// resolution, and a single loop can persist several agent turns within the
// same second (intermediate tool-call turns plus the final text turn), so
// rowid (monotonic per INSERT) breaks the tie in favor of the last-saved turn.
func (s *SQLiteStore) LastAgentTurn(ctx context.Context, sessionID string) (*Turn, error) {
	const q = `SELECT turn_id, session_id, response_id, phase, role, content, created_at
	           FROM turn
	           WHERE session_id = ? AND role = 'agent'
	           ORDER BY created_at DESC, rowid DESC
	           LIMIT 1`
	var t Turn
	err := s.db.QueryRowContext(ctx, q, sessionID).Scan(
		&t.TurnID, &t.SessionID, &t.ResponseID, &t.Phase, &t.Role, &t.Content, &t.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("store: last agent turn for session %s: %w", sessionID, err)
	}
	return &t, nil
}
