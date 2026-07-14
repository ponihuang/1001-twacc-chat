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

func TestNotificationPreviewForMediaMessages(t *testing.T) {
	cases := []struct {
		name        string
		messageType string
		content     string
		want        string
	}{
		{name: "image", messageType: "image", content: "", want: "傳送了圖片"},
		{name: "file", messageType: "file", content: "", want: "傳送了附件"},
		{name: "text", messageType: "text", content: "  hello  ", want: "hello"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := notificationPreview(tc.messageType, tc.content); got != tc.want {
				t.Fatalf("notificationPreview(%q, %q) = %q, want %q", tc.messageType, tc.content, got, tc.want)
			}
		})
	}
}
