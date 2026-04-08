package store

import (
	"context"
	"database/sql"
)

// Store defines the minimal database lifecycle used by the app bootstrap.
type Store interface {
	Ping(context.Context) error
	Close() error
	DB() *sql.DB
}

// SQLStore wraps a sql.DB so higher layers can depend on a narrow interface.
type SQLStore struct {
	db *sql.DB
}

// NewSQLStore builds a store from an initialized sql.DB.
func NewSQLStore(db *sql.DB) *SQLStore {
	return &SQLStore{db: db}
}

// Ping verifies database connectivity.
func (s *SQLStore) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// Close releases database resources.
func (s *SQLStore) Close() error {
	return s.db.Close()
}

// DB exposes the underlying sql.DB for repository implementations.
func (s *SQLStore) DB() *sql.DB {
	return s.db
}
