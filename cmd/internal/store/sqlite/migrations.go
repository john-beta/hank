package store

import "database/sql"

// schema defines the session, turn, and call tables. All are created
// idempotently on every startup; no migration framework is used.
//
// output_text semantics on turn:
//   - role='user'  + plain message  -> output_text is the user's message.
//   - role='user'  + tool result    -> output_text is NULL; the payload lives
//     in the matching call.result UPDATE, not here.
//   - role='agent'                  -> output_text is the concatenated
//     text_delta the model streamed for that response.
//
// call lifecycle: exactly one row per call_id. INSERT once (result NULL) when
// the agent emits the call, then UPDATE that same row to fill result when it
// resolves. Never a second INSERT for the same call_id.
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

// RunMigrations creates the session, turn, and call tables if they do not exist.
func RunMigrations(db *sql.DB) error {
	if _, err := db.Exec(schema); err != nil {
		return err
	}
	return nil
}
