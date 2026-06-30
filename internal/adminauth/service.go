package adminauth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidSession = errors.New("invalid admin session")
	ErrSessionExpired = errors.New("admin session expired")
)

// Session is the persisted admin session state used for Bearer authentication.
type Session struct {
	ID          int64
	AdminUserID int64
	DeviceID    string
	ExpiresAt   time.Time
}

// Repository defines persistence required by admin session auth.
type Repository interface {
	CreateSession(tokenHash string, adminUserID int64, deviceID string, expiresAt, now time.Time) error
	FindSessionByHash(tokenHash string) (Session, error)
	TouchSession(sessionID int64, usedAt time.Time) error
}

// Service issues and validates admin session tokens.
type Service struct {
	repo Repository
	ttl  time.Duration
}

// NewService creates an admin session service.
func NewService(repo Repository, ttl time.Duration) *Service {
	return &Service{repo: repo, ttl: ttl}
}

// Issue creates a new session token for the given admin account and device.
func (s *Service) Issue(adminUserID int64, deviceID string) (string, time.Time, error) {
	if s == nil || s.repo == nil {
		return "", time.Time{}, fmt.Errorf("admin session service unavailable")
	}
	if adminUserID <= 0 || strings.TrimSpace(deviceID) == "" {
		return "", time.Time{}, ErrInvalidSession
	}

	token, err := generateToken()
	if err != nil {
		return "", time.Time{}, err
	}

	now := time.Now().UTC()
	expiresAt := now.Add(s.ttl)
	if err := s.repo.CreateSession(hashToken(token), adminUserID, strings.TrimSpace(deviceID), expiresAt, now); err != nil {
		return "", time.Time{}, err
	}

	return token, expiresAt, nil
}

// Authenticate validates a Bearer token and returns the admin session principal.
func (s *Service) Authenticate(token string) (Session, error) {
	if s == nil || s.repo == nil {
		return Session{}, fmt.Errorf("admin session service unavailable")
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return Session{}, ErrInvalidSession
	}

	session, err := s.repo.FindSessionByHash(hashToken(token))
	if err != nil {
		return Session{}, err
	}

	now := time.Now().UTC()
	if now.After(session.ExpiresAt) {
		return Session{}, ErrSessionExpired
	}

	if err := s.repo.TouchSession(session.ID, now); err != nil {
		return Session{}, err
	}

	return session, nil
}

func generateToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
