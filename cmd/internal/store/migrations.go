package store

import "database/sql"

// schema defines the session and turn tables. Both are created idempotently on
// every startup; no migration framework is used.
const schema = `
CREATE TABLE IF NOT EXISTS session (
    session_id TEXT PRIMARY KEY,
    phase      INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS turn (
    turn_id     TEXT PRIMARY KEY,
    session_id  TEXT NOT NULL REFERENCES session(session_id),
    response_id TEXT,
    phase       INTEGER NOT NULL,
    role        TEXT NOT NULL CHECK(role IN ('user', 'agent')),
    content     TEXT,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`

// RunMigrations creates the session and turn tables if they do not exist.
func RunMigrations(db *sql.DB) error {
	if _, err := db.Exec(schema); err != nil {
		return err
	}
	return nil
}
