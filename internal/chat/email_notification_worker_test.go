package chat

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type fakeUnreadMailer struct {
	calls int
	err   error
}

func (m *fakeUnreadMailer) SendUnreadNotification(email string, conversationTitle string, unreadCount int, latestSenderName string, latestMessageText string) error {
	m.calls++
	return m.err
}

func TestEmailNotificationWorkerCancelsWhenNoUnreadBeforeSending(t *testing.T) {
	repo := &mockRepository{
		claimSchedules: []*EmailNotificationSchedule{{
			ID:              11,
			UserID:          7,
			ConversationID:  9,
			FirstMessageID:  3,
			LatestMessageID: 5,
		}},
		delivery: EmailNotificationDelivery{
			UserStatus:  "active",
			Email:       "user@example.com",
			IsMember:    true,
			UnreadCount: 0,
		},
	}
	mailer := &fakeUnreadMailer{}
	worker := NewEmailNotificationWorker(repo, mailer)

	worker.processDue(testContext(t))

	if repo.recoverStaleCalls != 1 {
		t.Fatalf("recover stale calls = %d, want 1", repo.recoverStaleCalls)
	}
	if repo.cancelledScheduleID != 11 {
		t.Fatalf("cancelledScheduleID = %d, want 11", repo.cancelledScheduleID)
	}
	if mailer.calls != 0 {
		t.Fatalf("mailer calls = %d, want 0", mailer.calls)
	}
}

func TestEmailNotificationWorkerSendsWhenStillUnread(t *testing.T) {
	repo := &mockRepository{
		claimSchedules: []*EmailNotificationSchedule{{
			ID:              12,
			UserID:          7,
			ConversationID:  9,
			FirstMessageID:  3,
			LatestMessageID: 5,
		}},
		delivery: EmailNotificationDelivery{
			UserStatus:        "active",
			Email:             "user@example.com",
			IsMember:          true,
			UnreadCount:       2,
			ConversationTitle: "測試聊天室",
			LatestSenderName:  "Elva",
			LatestMessageText: "hello",
		},
	}
	mailer := &fakeUnreadMailer{}
	worker := NewEmailNotificationWorker(repo, mailer)

	worker.processDue(testContext(t))

	if mailer.calls != 1 {
		t.Fatalf("mailer calls = %d, want 1", mailer.calls)
	}
	if repo.sentScheduleID != 12 {
		t.Fatalf("sentScheduleID = %d, want 12", repo.sentScheduleID)
	}
}

func TestEmailNotificationWorkerRetriesWhenSMTPFails(t *testing.T) {
	repo := &mockRepository{
		claimSchedules: []*EmailNotificationSchedule{{
			ID:              13,
			UserID:          7,
			ConversationID:  9,
			FirstMessageID:  3,
			LatestMessageID: 5,
		}},
		delivery: EmailNotificationDelivery{
			UserStatus:  "active",
			Email:       "user@example.com",
			IsMember:    true,
			UnreadCount: 1,
		},
	}
	mailer := &fakeUnreadMailer{err: errors.New("smtp unavailable")}
	worker := NewEmailNotificationWorker(repo, mailer)

	worker.processDue(testContext(t))

	if repo.failedScheduleID != 13 {
		t.Fatalf("failedScheduleID = %d, want 13", repo.failedScheduleID)
	}
	if repo.failedMaxAttempts != emailNotificationMaxAttempts {
		t.Fatalf("failedMaxAttempts = %d, want %d", repo.failedMaxAttempts, emailNotificationMaxAttempts)
	}
	if !strings.Contains(repo.failedErrMessage, "smtp unavailable") {
		t.Fatalf("failedErrMessage = %q, want SMTP error", repo.failedErrMessage)
	}
	if time.Until(repo.failedRetryAt) <= 0 {
		t.Fatalf("failedRetryAt should be in the future: %v", repo.failedRetryAt)
	}
}

func TestEmailNotificationWorkerRecoversStaleProcessingJobs(t *testing.T) {
	repo := &mockRepository{}
	worker := NewEmailNotificationWorker(repo, &fakeUnreadMailer{})

	worker.processDue(testContext(t))

	if repo.recoverStaleCalls != 1 {
		t.Fatalf("recover stale calls = %d, want 1", repo.recoverStaleCalls)
	}
	if repo.recoverStaleNow.IsZero() || repo.recoverStaleBefore.IsZero() {
		t.Fatalf("recover stale timestamps should be set")
	}
	if repo.recoverStaleNow.Sub(repo.recoverStaleBefore) < emailNotificationStaleAfter {
		t.Fatalf("stale window = %s, want at least %s", repo.recoverStaleNow.Sub(repo.recoverStaleBefore), emailNotificationStaleAfter)
	}
}

func testContext(t *testing.T) context.Context {
	t.Helper()
	return context.Background()
}
