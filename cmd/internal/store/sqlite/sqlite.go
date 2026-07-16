package store

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite" // pure-Go SQLite driver registration

	"github.com/john-beta/hank/cmd/internal/store"
)

// SQLiteStore is a pure-Go SQLite-backed Store.
type SQLiteStore struct {
	db *sql.DB
}

// compile-time check that SQLiteStore satisfies the Store interface.
var _ store.Store = (*SQLiteStore)(nil)

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
