package mysql

import (
	"strings"
	"testing"
)

func TestEnqueueEmailNotificationMergeDoesNotDelayDueAt(t *testing.T) {
	upperSQL := strings.ToUpper(enqueueEmailNotificationSQL)
	if !strings.Contains(upperSQL, "ON DUPLICATE KEY UPDATE") {
		t.Fatalf("enqueue SQL should merge duplicate active notifications")
	}
	if !strings.Contains(upperSQL, "LATEST_MESSAGE_ID = GREATEST") {
		t.Fatalf("enqueue SQL should merge by advancing latest_message_id")
	}

	updateClause := upperSQL[strings.Index(upperSQL, "ON DUPLICATE KEY UPDATE"):]
	if strings.Contains(updateClause, "DUE_AT") {
		t.Fatalf("enqueue SQL must not update due_at when merging pending notifications: %s", updateClause)
	}
}
