package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"1001-twacc-chat/internal/auth"
)

// SessionRepository implements session storage in MySQL.
type SessionRepository struct {
	db *sql.DB
}

// NewSessionRepository creates a MySQL-backed session repository.
func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// CreateSession persists a newly issued session token hash.
func (r *SessionRepository) CreateSession(tokenHash string, userID int64, deviceID string, expiresAt, now time.Time) error {
	_, err := r.db.Exec(
		`INSERT INTO auth_sessions (token_hash, user_id, device_id, expires_at, created_at, last_used_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		tokenHash,
		userID,
		deviceID,
		expiresAt,
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}

	return nil
}

// FindSessionByHash loads a session by token hash.
func (r *SessionRepository) FindSessionByHash(tokenHash string) (auth.Session, error) {
	var session auth.Session
	row := r.db.QueryRow(
		`SELECT id, user_id, device_id, expires_at
		   FROM auth_sessions
		  WHERE token_hash = ? AND revoked_at IS NULL
		  LIMIT 1`,
		tokenHash,
	)
	if err := row.Scan(&session.ID, &session.UserID, &session.DeviceID, &session.ExpiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return auth.Session{}, auth.ErrInvalidSession
		}
		return auth.Session{}, fmt.Errorf("find session: %w", err)
	}

	return session, nil
}

// TouchSession updates last_used_at for an authenticated session.
func (r *SessionRepository) TouchSession(sessionID int64, usedAt time.Time) error {
	_, err := r.db.Exec(`UPDATE auth_sessions SET last_used_at = ? WHERE id = ?`, usedAt, sessionID)
	if err != nil {
		return fmt.Errorf("touch session: %w", err)
	}

	return nil
}
