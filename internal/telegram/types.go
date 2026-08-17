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
