package auth

import (
	"errors"
	"testing"
	"time"
)

type mockRepository struct {
	session Session
	create  struct {
		tokenHash string
		userID    int64
		deviceID  string
	}
}

func (m *mockRepository) CreateSession(tokenHash string, userID int64, deviceID string, expiresAt, now time.Time) error {
	m.create.tokenHash = tokenHash
	m.create.userID = userID
	m.create.deviceID = deviceID
	m.session = Session{ID: 1, UserID: userID, DeviceID: deviceID, ExpiresAt: expiresAt}
	return nil
}

func (m *mockRepository) FindSessionByHash(tokenHash string) (Session, error) {
	if tokenHash == "missing" {
		return Session{}, ErrInvalidSession
	}
	return m.session, nil
}

func (m *mockRepository) TouchSession(sessionID int64, usedAt time.Time) error {
	return nil
}

func TestIssueCreatesSession(t *testing.T) {
	repo := &mockRepository{}
	service := NewService(repo, time.Hour)

	token, expiresAt, err := service.Issue(7, "device-1")
	if err != nil {
		t.Fatalf("Issue returned error: %v", err)
	}
	if token == "" {
		t.Fatal("expected token to be set")
	}
	if expiresAt.IsZero() {
		t.Fatal("expected expiresAt to be set")
	}
	if repo.create.userID != 7 || repo.create.deviceID != "device-1" {
		t.Fatalf("unexpected session create payload: %+v", repo.create)
	}
}

func TestAuthenticateRejectsExpiredSession(t *testing.T) {
	repo := &mockRepository{session: Session{ID: 1, UserID: 7, DeviceID: "device-1", ExpiresAt: time.Now().UTC().Add(-time.Minute)}}
	service := NewService(repo, time.Hour)

	_, err := service.Authenticate("token")
	if !errors.Is(err, ErrSessionExpired) {
		t.Fatalf("expected ErrSessionExpired, got %v", err)
	}
}
