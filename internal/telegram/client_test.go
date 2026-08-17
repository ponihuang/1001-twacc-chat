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
