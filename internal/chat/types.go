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
	DirectSourceSystem string
	DirectExternalID   string
	MemberCount        int
	LastMessageType    string
	LastMessagePreview string
	LastMessageAt      time.Time
	UnreadCount        int
}

// Conversation stores the minimal metadata needed by the message list response.
type Conversation struct {
	ID    int64
	Type  string
	Title string
}

// Message stores the chat message payload returned from persistence.
type Message struct {
	ID                int64
	ConversationID    int64
	ConversationTitle string
	SenderID          int64
	SenderName        string
	MessageType       string
	Content           string
	CreatedAt         time.Time
	Attachment        *Attachment
	Attachments       []Attachment
}

// CreateMessageInput is the normalized payload stored for a new message.
type CreateMessageInput struct {
	ConversationID int64
	SenderID       int64
	MessageType    string
	Content        string
	Attachment     *AttachmentInput
	Attachments    []AttachmentInput
}

// Attachment stores metadata for a single uploaded file.
type Attachment struct {
	ID           int64
	OriginalName string
	StoragePath  string
	MIMEType     string
	SizeBytes    int64
}

// AttachmentInput is the storage payload for a new uploaded file.
type AttachmentInput struct {
	OriginalName string
	StoragePath  string
	MIMEType     string
	SizeBytes    int64
}

// ConversationItem is the API-facing chat list item.
type ConversationItem struct {
	ConversationID     int64  `json:"conversation_id"`
	Type               string `json:"type"`
	Title              string `json:"title"`
	DirectSourceSystem string `json:"direct_source_system,omitempty"`
	DirectExternalID   string `json:"direct_external_user_id,omitempty"`
	MemberCount        int    `json:"member_count"`
	LastMessageType    string `json:"last_message_type,omitempty"`
	LastMessagePreview string `json:"last_message_preview,omitempty"`
	LastMessageAt      string `json:"last_message_at,omitempty"`
	UnreadCount        int    `json:"unread_count"`
}

// DirectConversationData is the API-facing payload for a direct conversation create/load request.
type DirectConversationData struct {
	ConversationID int64  `json:"conversation_id"`
	Type           string `json:"type"`
	Title          string `json:"title"`
}

// UserSearchResult stores a searchable chat user.
type UserSearchResult struct {
	SourceSystem   string
	ExternalUserID string
	DisplayName    string
}

// UserSearchItem is the API-facing search result for creating direct conversations.
type UserSearchItem struct {
	SourceSystem   string `json:"source_system"`
	ExternalUserID string `json:"external_user_id"`
	DisplayName    string `json:"display_name"`
}

// UserSearchData is the API-facing user search payload.
type UserSearchData struct {
	Query string           `json:"query"`
	Users []UserSearchItem `json:"users"`
}

// RealtimeEvent is the WebSocket payload pushed to chat clients.
type RealtimeEvent struct {
	EventType      string                  `json:"event_type"`
	ConversationID int64                   `json:"conversation_id,omitempty"`
	Conversation   *DirectConversationData `json:"conversation,omitempty"`
	Message        MessageItem             `json:"message,omitempty"`
	MessageID      int64                   `json:"message_id,omitempty"`
}

// MessageItem is the API-facing message item.
type MessageItem struct {
	MessageID   int64            `json:"message_id"`
	SenderID    int64            `json:"sender_id"`
	SenderName  string           `json:"sender_name"`
	MessageType string           `json:"message_type"`
	Content     string           `json:"content"`
	CreatedAt   string           `json:"created_at"`
	Attachment  *AttachmentItem  `json:"attachment,omitempty"`
	Attachments []AttachmentItem `json:"attachments,omitempty"`
}

// AttachmentItem is the API-facing uploaded file metadata.
type AttachmentItem struct {
	OriginalName string `json:"original_name"`
	URL          string `json:"url"`
	MIMEType     string `json:"mime_type"`
	SizeBytes    int64  `json:"size_bytes"`
}

// MessageListData is the API-facing message list payload.
type MessageListData struct {
	ConversationID int64         `json:"conversation_id"`
	Type           string        `json:"type"`
	Title          string        `json:"title"`
	Messages       []MessageItem `json:"messages"`
}

// MessageSearchItem is the API-facing result for global message search.
type MessageSearchItem struct {
	ConversationID    int64  `json:"conversation_id"`
	ConversationTitle string `json:"conversation_title"`
	MessageID         int64  `json:"message_id"`
	SenderID          int64  `json:"sender_id"`
	SenderName        string `json:"sender_name"`
	MessageType       string `json:"message_type"`
	Content           string `json:"content"`
	CreatedAt         string `json:"created_at"`
}

// MessageSearchData is the API-facing global message search payload.
type MessageSearchData struct {
	Query    string              `json:"query"`
	Messages []MessageSearchItem `json:"messages"`
}

// CreateMessageRequest is the HTTP payload for sending a chat message.
type CreateMessageRequest struct {
	Type        string            `json:"type"`
	Content     string            `json:"content"`
	Attachment  *AttachmentInput  `json:"-"`
	Attachments []AttachmentInput `json:"-"`
}

// CreateDirectConversationRequest is the HTTP payload for creating/loading a direct conversation.
type CreateDirectConversationRequest struct {
	SourceSystem   string `json:"source_system"`
	ExternalUserID string `json:"external_user_id"`
}

// SentMessageData is the API-facing payload for a newly created message.
type SentMessageData struct {
	ConversationID int64       `json:"conversation_id"`
	Message        MessageItem `json:"message"`
}

// RecalledMessageData is the API-facing payload for a recalled message.
type RecalledMessageData struct {
	ConversationID int64 `json:"conversation_id"`
	MessageID      int64 `json:"message_id"`
}

// Response is the API response shape for chat endpoints.
type Response struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Broker publishes realtime chat events to connected clients.
type Broker interface {
	PublishToUsers(userIDs []int64, event RealtimeEvent)
	Subscribe(userID int64) (<-chan RealtimeEvent, func())
}

// Repository defines persistence required by chat list and message list endpoints.
type Repository interface {
	IsSystemAdmin(userID int64) (bool, error)
	ListConversations(userID int64) ([]ConversationSummary, error)
	SearchUsers(actorUserID int64, sourceSystem, query string, limit int) ([]UserSearchResult, error)
	CreateOrGetDirectConversation(actorUserID int64, sourceSystem, externalUserID string) (Conversation, bool, error)
	ListConversationMemberIDs(conversationID int64) ([]int64, error)
	GetConversationForUser(userID, conversationID int64) (Conversation, error)
	ListMessages(conversationID int64, limit int) ([]Message, error)
	SearchMessages(userID int64, query string, limit int) ([]Message, error)
	MarkConversationRead(userID, conversationID int64) error
	CreateMessage(input CreateMessageInput) (Message, error)
	DeleteMessage(conversationID, messageID, actorUserID int64) error
}
