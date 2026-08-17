package telegram

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// Handler exposes Telegram webhook endpoints.
type Handler struct {
	client *Client
	logger *log.Logger
}

// NewHandler builds the Telegram HTTP handler.
func NewHandler(client *Client) *Handler {
	return &Handler{
		client: client,
		logger: log.Default(),
	}
}

// SetLogger replaces the default logger. It is primarily used by tests.
func (h *Handler) SetLogger(logger *log.Logger) {
	if logger != nil {
		h.logger = logger
	}
}

// Webhook handles POST /api/telegram/webhook.
func (h *Handler) Webhook(w http.ResponseWriter, r *http.Request) {
	if h.client == nil || !h.client.Configured() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"ok":      false,
			"message": "telegram bot token is not configured",
		})
		return
	}

	var update Update
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&update); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"ok":      false,
			"message": "invalid telegram update",
		})
		return
	}

	h.logTextMessage(update)

	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true,
	})
}

func (h *Handler) logTextMessage(update Update) {
	if update.Message == nil || update.Message.Text == "" {
		return
	}

	message := update.Message
	var userID int64
	var username string
	var firstName string
	if message.From != nil {
		userID = message.From.ID
		username = strings.TrimSpace(message.From.Username)
		firstName = strings.TrimSpace(message.From.FirstName)
	}

	h.logger.Printf(
		"telegram text message received chat_id=%d user_id=%d username=%q first_name=%q message_id=%d text=%q",
		message.Chat.ID,
		userID,
		username,
		firstName,
		message.MessageID,
		message.Text,
	)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if payload == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(payload)
}
