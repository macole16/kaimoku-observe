package postgres

import (
	"database/sql"
	"embed"

	observe "github.com/macole16/kaimoku-observe"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Store implements observe.Store backed by PostgreSQL.
type Store struct {
	db *sql.DB
}

// New creates a new PostgreSQL-backed Store.
func New(db *sql.DB) *Store {
	return &Store{db: db}
}

// DB returns the underlying database connection.
func (s *Store) DB() *sql.DB { return s.db }

// MigrationsFS returns the embedded migrations filesystem.
func MigrationsFS() embed.FS { return migrationsFS }

// Compile-time interface check.
var _ observe.Store = (*Store)(nil)
