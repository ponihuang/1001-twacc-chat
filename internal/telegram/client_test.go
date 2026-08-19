package telegram

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetMeReturnsBotIdentity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bottest-token/getMe" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"ok": true,
			"result": {
				"id": 987,
				"username": "twacc_bot",
				"first_name": "TWACC"
			}
		}`))
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.baseURL = server.URL

	bot, err := client.GetMe(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if bot.ID != 987 || bot.Username != "twacc_bot" || bot.FirstName != "TWACC" {
		t.Fatalf("unexpected bot identity: %+v", bot)
	}
}

func TestGetMeReturnsTelegramErrorDescription(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"ok": false,
			"error_code": 401,
			"description": "Unauthorized"
		}`))
	}))
	defer server.Close()

	client := NewClient("bad-token")
	client.baseURL = server.URL

	_, err := client.GetMe(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "status=401") || !strings.Contains(err.Error(), "Unauthorized") {
		t.Fatalf("expected clear Telegram error, got %v", err)
	}
}

func TestGetMeRequiresConfiguredToken(t *testing.T) {
	client := NewClient("")

	_, err := client.GetMe(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "TELEGRAM_BOT_TOKEN is not configured") {
		t.Fatalf("expected missing token error, got %v", err)
	}
}

func TestSetWebhookPostsWebhookURL(t *testing.T) {
	const webhookURL = "https://example.com/api/telegram/webhook"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bottest-token/setWebhook" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("expected method %s, got %s", http.MethodPost, r.Method)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if got := r.Form.Get("url"); got != webhookURL {
			t.Fatalf("expected webhook URL %q, got %q", webhookURL, got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok": true, "result": true, "description": "Webhook was set"}`))
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.baseURL = server.URL

	if err := client.SetWebhook(context.Background(), webhookURL); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestSetWebhookRequiresURL(t *testing.T) {
	client := NewClient("test-token")

	err := client.SetWebhook(context.Background(), " ")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "telegram webhook URL is required") {
		t.Fatalf("expected webhook URL error, got %v", err)
	}
}

func TestDeleteWebhookCallsTelegramAPI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bottest-token/deleteWebhook" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("expected method %s, got %s", http.MethodPost, r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok": true, "result": true}`))
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.baseURL = server.URL

	if err := client.DeleteWebhook(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestGetWebhookInfoReturnsStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bottest-token/getWebhookInfo" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Fatalf("expected method %s, got %s", http.MethodGet, r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"ok": true,
			"result": {
				"url": "https://example.com/api/telegram/webhook",
				"has_custom_certificate": false,
				"pending_update_count": 2,
				"last_error_date": 1787020000,
				"last_error_message": "connection refused",
				"max_connections": 40,
				"ip_address": "203.0.113.10",
				"allowed_updates": ["message"]
			}
		}`))
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.baseURL = server.URL

	info, err := client.GetWebhookInfo(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if info.URL != "https://example.com/api/telegram/webhook" {
		t.Fatalf("unexpected webhook URL %q", info.URL)
	}
	if info.PendingUpdateCount != 2 {
		t.Fatalf("expected pending update count 2, got %d", info.PendingUpdateCount)
	}
	if info.LastErrorMessage != "connection refused" {
		t.Fatalf("unexpected last error message %q", info.LastErrorMessage)
	}
	if len(info.AllowedUpdates) != 1 || info.AllowedUpdates[0] != "message" {
		t.Fatalf("unexpected allowed updates: %+v", info.AllowedUpdates)
	}
}
