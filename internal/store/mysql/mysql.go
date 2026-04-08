package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"1001-twacc-chat/internal/config"
	"1001-twacc-chat/internal/store"
	_ "github.com/go-sql-driver/mysql"
)

const driverName = "mysql"

// Open creates a MySQL-backed store from environment-derived configuration.
func Open(ctx context.Context, cfg config.MySQLConfig) (*store.SQLStore, error) {
	if !cfg.Enabled() {
		return nil, fmt.Errorf("mysql config is incomplete")
	}

	db, err := sql.Open(driverName, cfg.FormattedDSN())
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	return store.NewSQLStore(db), nil
}
