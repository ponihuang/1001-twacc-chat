package chat

import (
	"time"

	"1001-twacc-chat/internal/auth"
)

// SessionPrincipal describes the authenticated actor from a session token.
type SessionPrincipal = auth.Session

// SessionAuthenticator validates a Bearer session token.
type SessionAuthenticator interface {
	Authenticate(token string) (auth.Session, error)
}

// ConversationSummary is the storage-facing chat list shape.
type ConversationSummary struct {
	ConversationID     int64
	Type               string
	Title              string
	MemberCount        int
	LastMessageType    string
	LastMessagePreview string
	LastMessageAt      time.Time
}

// Conversation stores the minimal metadata needed by the message list response.
type Conversation struct {
	ID    int64
	Type  string
	Title string
}

// Message stores the chat message payload returned from persistence.
type Message struct {
	ID             int64
	ConversationID int64
	SenderID       int64
	SenderName     string
	MessageType    string
	Content        string
	CreatedAt      time.Time
}

// CreateMessageInput is the normalized payload stored for a new message.
type CreateMessageInput struct {
	ConversationID int64
	SenderID       int64
	MessageType    string
	Content        string
}

// ConversationItem is the API-facing chat list item.
type ConversationItem struct {
	ConversationID     int64  `json:"conversation_id"`
	Type               string `json:"type"`
	Title              string `json:"title"`
	MemberCount        int    `json:"member_count"`
	LastMessageType    string `json:"last_message_type,omitempty"`
	LastMessagePreview string `json:"last_message_preview,omitempty"`
	LastMessageAt      string `json:"last_message_at,omitempty"`
}

// MessageItem is the API-facing message item.
type MessageItem struct {
	MessageID   int64  `json:"message_id"`
	SenderID    int64  `json:"sender_id"`
	SenderName  string `json:"sender_name"`
	MessageType string `json:"message_type"`
	Content     string `json:"content"`
	CreatedAt   string `json:"created_at"`
}

// MessageListData is the API-facing message list payload.
type MessageListData struct {
	ConversationID int64         `json:"conversation_id"`
	Type           string        `json:"type"`
	Title          string        `json:"title"`
	Messages       []MessageItem `json:"messages"`
}

// CreateMessageRequest is the HTTP payload for sending a chat message.
type CreateMessageRequest struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

// SentMessageData is the API-facing payload for a newly created message.
type SentMessageData struct {
	ConversationID int64       `json:"conversation_id"`
	Message        MessageItem `json:"message"`
}

// Response is the API response shape for chat endpoints.
type Response struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Repository defines persistence required by chat list and message list endpoints.
type Repository interface {
	IsSystemAdmin(userID int64) (bool, error)
	ListConversations(userID int64) ([]ConversationSummary, error)
	GetConversationForUser(userID, conversationID int64) (Conversation, error)
	ListMessages(conversationID int64, limit int) ([]Message, error)
	CreateMessage(input CreateMessageInput) (Message, error)
}
