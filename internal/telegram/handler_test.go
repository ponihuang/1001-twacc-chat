package telegram

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWebhookLogsTextMessage(t *testing.T) {
	handler := NewHandler(NewClient("test-token"))
	var logs bytes.Buffer
	handler.SetLogger(log.New(&logs, "", 0))

	req := httptest.NewRequest(http.MethodPost, "/api/telegram/webhook", strings.NewReader(`{
		"update_id": 1000,
		"message": {
			"message_id": 42,
			"from": {
				"id": 123,
				"username": "alice",
				"first_name": "Alice"
			},
			"chat": {
				"id": -456
			},
			"text": "hello"
		}
	}`))
	rec := httptest.NewRecorder()

	handler.Webhook(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	logged := logs.String()
	for _, want := range []string{
		"chat_id=-456",
		"user_id=123",
		`username="alice"`,
		`first_name="Alice"`,
		"message_id=42",
		`text="hello"`,
	} {
		if !strings.Contains(logged, want) {
			t.Fatalf("expected log to contain %q, got %q", want, logged)
		}
	}
}

func TestWebhookRequiresConfiguredBotToken(t *testing.T) {
	handler := NewHandler(NewClient(""))
	req := httptest.NewRequest(http.MethodPost, "/api/telegram/webhook", strings.NewReader(`{"update_id":1000}`))
	rec := httptest.NewRecorder()

	handler.Webhook(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, rec.Code)
	}
}
