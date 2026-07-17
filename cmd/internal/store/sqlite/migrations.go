package store

import "database/sql"

// schema defines the session, turn, and call tables, created idempotently on
// startup (no migration framework).
//
// turn.output_text: the user's message or the agent's streamed text; NULL for a
// user turn carrying a tool result (payload lives in call.result instead).
//
// call: one row per call_id — INSERT with result NULL when emitted, UPDATE in
// place when it resolves. Never a second INSERT for the same call_id.
const schema = `
CREATE TABLE IF NOT EXISTS session (
    session_id        TEXT PRIMARY KEY,
    root_dir          TEXT NOT NULL,
    created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS turn (
    turn_id     TEXT PRIMARY KEY,
    session_id  TEXT NOT NULL REFERENCES session(session_id),
    role        TEXT NOT NULL CHECK(role IN ('user', 'agent')),
    output_text TEXT,
    response_id TEXT,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS call (
    call_id      TEXT PRIMARY KEY,
    turn_id      TEXT NOT NULL REFERENCES turn(turn_id),
    name         TEXT NOT NULL,
    args         TEXT,
    result       TEXT,
    auto_re_feed INTEGER NOT NULL
);
`

func RunMigrations(db *sql.DB) error {
	if _, err := db.Exec(schema); err != nil {
		return err
	}
	return nil
}
