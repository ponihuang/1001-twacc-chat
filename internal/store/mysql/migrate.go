package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ApplyMigrations runs unapplied SQL files from the given directory.
func ApplyMigrations(ctx context.Context, db *sql.DB, dir string) (int, error) {
	if strings.TrimSpace(dir) == "" {
		return 0, fmt.Errorf("migrations dir is required")
	}

	if err := ensureSchemaMigrationsTable(ctx, db); err != nil {
		return 0, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, fmt.Errorf("read migrations dir: %w", err)
	}

	applied, err := listAppliedMigrations(ctx, db)
	if err != nil {
		return 0, err
	}

	filenames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		filenames = append(filenames, entry.Name())
	}
	sort.Strings(filenames)

	appliedCount := 0
	for _, filename := range filenames {
		if _, ok := applied[filename]; ok {
			continue
		}

		statements, err := readMigrationStatements(filepath.Join(dir, filename))
		if err != nil {
			return appliedCount, fmt.Errorf("read migration %s: %w", filename, err)
		}

		for _, stmt := range statements {
			if _, err := db.ExecContext(ctx, stmt); err != nil {
				return appliedCount, fmt.Errorf("apply migration %s: %w", filename, err)
			}
		}

		if _, err := db.ExecContext(
			ctx,
			`INSERT INTO schema_migrations (filename, applied_at) VALUES (?, ?)`,
			filename,
			time.Now().UTC(),
		); err != nil {
			return appliedCount, fmt.Errorf("record migration %s: %w", filename, err)
		}

		appliedCount++
	}

	return appliedCount, nil
}

func ensureSchemaMigrationsTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			filename VARCHAR(255) NOT NULL,
			applied_at DATETIME NOT NULL,
			PRIMARY KEY (id),
			UNIQUE KEY uk_schema_migrations_filename (filename)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	if err != nil {
		return fmt.Errorf("ensure schema_migrations: %w", err)
	}

	return nil
}

func listAppliedMigrations(ctx context.Context, db *sql.DB) (map[string]struct{}, error) {
	rows, err := db.QueryContext(ctx, `SELECT filename FROM schema_migrations ORDER BY id ASC`)
	if err != nil {
		return nil, fmt.Errorf("list applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]struct{})
	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return nil, fmt.Errorf("scan applied migration: %w", err)
		}
		applied[filename] = struct{}{}
	}

	return applied, rows.Err()
}

func readMigrationStatements(path string) ([]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return splitSQLStatements(string(content))
}

func splitSQLStatements(raw string) ([]string, error) {
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	lines := strings.Split(raw, "\n")

	var statements []string
	var builder strings.Builder
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}

		builder.WriteString(line)
		builder.WriteString("\n")

		if strings.HasSuffix(trimmed, ";") {
			statement := strings.TrimSpace(builder.String())
			statement = strings.TrimSuffix(statement, ";")
			statement = strings.TrimSpace(statement)
			if statement != "" {
				statements = append(statements, statement)
			}
			builder.Reset()
		}
	}

	if remainder := strings.TrimSpace(builder.String()); remainder != "" {
		return nil, fmt.Errorf("unterminated SQL statement")
	}

	return statements, nil
}
