package telegram

// Update is the incoming payload sent by Telegram Bot API webhooks.
type Update struct {
	UpdateID int64    `json:"update_id"`
	Message  *Message `json:"message"`
}

// Message contains the Telegram message fields needed during phase one.
type Message struct {
	MessageID int64  `json:"message_id"`
	From      *User  `json:"from"`
	Chat      Chat   `json:"chat"`
	Text      string `json:"text"`
}

// User identifies the Telegram user who sent the message.
type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
}

// Chat identifies the Telegram chat that received the message.
type Chat struct {
	ID int64 `json:"id"`
}

// BotUser contains the Telegram bot identity returned by getMe.
type BotUser struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
}

// WebhookInfo contains Telegram webhook status returned by getWebhookInfo.
type WebhookInfo struct {
	URL                  string   `json:"url"`
	HasCustomCertificate bool     `json:"has_custom_certificate"`
	PendingUpdateCount   int      `json:"pending_update_count"`
	LastErrorDate        int64    `json:"last_error_date"`
	LastErrorMessage     string   `json:"last_error_message"`
	MaxConnections       int      `json:"max_connections"`
	IPAddress            string   `json:"ip_address"`
	AllowedUpdates       []string `json:"allowed_updates"`
}
