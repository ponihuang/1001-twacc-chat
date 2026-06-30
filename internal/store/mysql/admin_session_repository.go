package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"1001-twacc-chat/internal/adminauth"
)

// AdminSessionRepository implements admin session storage in MySQL.
type AdminSessionRepository struct {
	db *sql.DB
}

// NewAdminSessionRepository creates a MySQL-backed admin session repository.
func NewAdminSessionRepository(db *sql.DB) *AdminSessionRepository {
	return &AdminSessionRepository{db: db}
}

// CreateSession persists a newly issued admin session token hash.
func (r *AdminSessionRepository) CreateSession(tokenHash string, adminUserID int64, deviceID string, expiresAt, now time.Time) error {
	_, err := r.db.Exec(
		`INSERT INTO admin_auth_sessions (token_hash, admin_user_id, device_id, expires_at, created_at, last_used_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		tokenHash,
		adminUserID,
		deviceID,
		expiresAt,
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("create admin session: %w", err)
	}

	return nil
}

// FindSessionByHash loads an admin session by token hash.
func (r *AdminSessionRepository) FindSessionByHash(tokenHash string) (adminauth.Session, error) {
	var session adminauth.Session
	row := r.db.QueryRow(
		`SELECT id, admin_user_id, device_id, expires_at
		   FROM admin_auth_sessions
		  WHERE token_hash = ? AND revoked_at IS NULL
		  LIMIT 1`,
		tokenHash,
	)
	if err := row.Scan(&session.ID, &session.AdminUserID, &session.DeviceID, &session.ExpiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return adminauth.Session{}, adminauth.ErrInvalidSession
		}
		return adminauth.Session{}, fmt.Errorf("find admin session: %w", err)
	}

	return session, nil
}

// TouchSession updates last_used_at for an authenticated admin session.
func (r *AdminSessionRepository) TouchSession(sessionID int64, usedAt time.Time) error {
	_, err := r.db.Exec(`UPDATE admin_auth_sessions SET last_used_at = ? WHERE id = ?`, usedAt, sessionID)
	if err != nil {
		return fmt.Errorf("touch admin session: %w", err)
	}

	return nil
}
