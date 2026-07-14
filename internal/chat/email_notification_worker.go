package chat

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"strings"
	"time"
)

const (
	emailNotificationPollInterval = 15 * time.Second
	emailNotificationRetryDelay   = 5 * time.Minute
	emailNotificationMaxAttempts  = 3
	emailNotificationStaleAfter   = 10 * time.Minute
)

// EmailNotificationWorker sends persisted unread email notification jobs.
type EmailNotificationWorker struct {
	repo   Repository
	mailer UnreadNotificationMailer
}

// NewEmailNotificationWorker builds a persisted unread notification worker.
func NewEmailNotificationWorker(repo Repository, mailer UnreadNotificationMailer) *EmailNotificationWorker {
	return &EmailNotificationWorker{repo: repo, mailer: mailer}
}

// Start runs the worker loop until the context is cancelled.
func (w *EmailNotificationWorker) Start(ctx context.Context) {
	if w == nil || w.repo == nil || w.mailer == nil {
		return
	}
	ticker := time.NewTicker(emailNotificationPollInterval)
	defer ticker.Stop()

	for {
		w.processDue(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (w *EmailNotificationWorker) processDue(ctx context.Context) {
	now := time.Now().UTC()
	if err := w.repo.RecoverStaleEmailNotifications(now, now.Add(-emailNotificationStaleAfter)); err != nil {
		log.Printf("recover stale chat email notifications: %v", err)
		return
	}
	for {
		if ctx.Err() != nil {
			return
		}
		token := randomLockToken()
		schedule, err := w.repo.ClaimDueEmailNotification(time.Now().UTC(), token)
		if err != nil {
			log.Printf("claim chat email notification: %v", err)
			return
		}
		if schedule == nil {
			return
		}
		w.processSchedule(*schedule)
	}
}

func (w *EmailNotificationWorker) processSchedule(schedule EmailNotificationSchedule) {
	now := time.Now().UTC()
	delivery, err := w.repo.LoadEmailNotificationDelivery(schedule)
	if err != nil {
		_ = w.repo.MarkEmailNotificationFailed(schedule.ID, now.Add(emailNotificationRetryDelay), emailNotificationMaxAttempts, err.Error())
		return
	}
	if strings.TrimSpace(delivery.UserStatus) != "active" ||
		strings.TrimSpace(delivery.Email) == "" ||
		!delivery.IsMember ||
		delivery.UnreadCount <= 0 {
		if err := w.repo.MarkEmailNotificationCancelled(schedule.ID, now); err != nil {
			log.Printf("cancel chat email notification %d: %v", schedule.ID, err)
		}
		return
	}

	if err := w.mailer.SendUnreadNotification(
		delivery.Email,
		delivery.ConversationTitle,
		delivery.UnreadCount,
		delivery.LatestSenderName,
		delivery.LatestMessageText,
	); err != nil {
		if markErr := w.repo.MarkEmailNotificationFailed(schedule.ID, now.Add(emailNotificationRetryDelay), emailNotificationMaxAttempts, err.Error()); markErr != nil {
			log.Printf("mark chat email notification failed %d: %v", schedule.ID, markErr)
		}
		return
	}

	if err := w.repo.MarkEmailNotificationSent(schedule.ID, now); err != nil {
		log.Printf("mark chat email notification sent %d: %v", schedule.ID, err)
	}
}

func randomLockToken() string {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(raw[:])
}
