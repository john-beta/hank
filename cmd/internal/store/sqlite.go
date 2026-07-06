package store

import (
	"database/sql"

	_ "modernc.org/sqlite" // pure-Go SQLite driver registration
)

// SQLiteStore is a pure-Go SQLite-backed Store. It holds an open connection
// pool but defines no schema yet.
type SQLiteStore struct {
	db *sql.DB
}

// compile-time check that SQLiteStore satisfies the Store interface.
var _ Store = (*SQLiteStore)(nil)

// NewSQLite opens (creating if needed) the database at path and verifies the
// connection.
func NewSQLite(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return &SQLiteStore{db: db}, nil
}

// Close releases the underlying connection pool.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
