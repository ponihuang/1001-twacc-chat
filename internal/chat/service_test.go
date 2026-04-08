package chat

import (
	"errors"
	"testing"
	"time"
)

type mockRepository struct {
	admins              map[int64]bool
	conversations       map[int64][]ConversationSummary
	headers             map[string]Conversation
	messages            map[int64][]Message
	createdMessages     []CreateMessageInput
	createMessageResult Message
	createMessageErr    error
}

func (m *mockRepository) IsSystemAdmin(userID int64) (bool, error) {
	return m.admins[userID], nil
}

func (m *mockRepository) ListConversations(userID int64) ([]ConversationSummary, error) {
	return append([]ConversationSummary(nil), m.conversations[userID]...), nil
}

func (m *mockRepository) GetConversationForUser(userID, conversationID int64) (Conversation, error) {
	key := conversationKey(userID, conversationID)
	conversation, ok := m.headers[key]
	if !ok {
		return Conversation{}, ErrConversationNotFound
	}
	return conversation, nil
}

func (m *mockRepository) ListMessages(conversationID int64, limit int) ([]Message, error) {
	messages := append([]Message(nil), m.messages[conversationID]...)
	if len(messages) > limit {
		messages = messages[len(messages)-limit:]
	}
	return messages, nil
}

func (m *mockRepository) CreateMessage(input CreateMessageInput) (Message, error) {
	if m.createMessageErr != nil {
		return Message{}, m.createMessageErr
	}

	m.createdMessages = append(m.createdMessages, input)
	if m.createMessageResult.ID != 0 {
		return m.createMessageResult, nil
	}

	return Message{
		ID:             int64(len(m.createdMessages)),
		ConversationID: input.ConversationID,
		SenderID:       input.SenderID,
		SenderName:     "你",
		MessageType:    input.MessageType,
		Content:        input.Content,
		CreatedAt:      time.Date(2026, 4, 7, 2, 0, 0, 0, time.UTC),
	}, nil
}

func conversationKey(userID, conversationID int64) string {
	return time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(userID + conversationID)).Format(time.RFC3339Nano)
}

func TestListConversations(t *testing.T) {
	repo := &mockRepository{
		conversations: map[int64][]ConversationSummary{
			7: {{ConversationID: 9, Type: "group", Title: "採購小組", MemberCount: 3, LastMessageType: "text", LastMessagePreview: "hello", LastMessageAt: time.Date(2026, 4, 7, 1, 2, 3, 0, time.UTC)}},
		},
	}
	service := NewService(repo)

	resp, status, err := service.ListConversations(SessionPrincipal{UserID: 7})
	if err != nil {
		t.Fatalf("ListConversations returned error: %v", err)
	}
	if status != 200 {
		t.Fatalf("status = %d, want 200", status)
	}

	items, ok := resp.Data.([]ConversationItem)
	if !ok {
		t.Fatalf("unexpected data type %T", resp.Data)
	}
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
	if items[0].ConversationID != 9 || items[0].Title != "採購小組" {
		t.Fatalf("unexpected conversation item: %+v", items[0])
	}
}

func TestListConversationsRejectsSystemAdmin(t *testing.T) {
	repo := &mockRepository{admins: map[int64]bool{9: true}}
	service := NewService(repo)

	_, _, err := service.ListConversations(SessionPrincipal{UserID: 9})
	if !errors.Is(err, ErrSystemAdminCannotChat) {
		t.Fatalf("expected ErrSystemAdminCannotChat, got %v", err)
	}
}

func TestListMessages(t *testing.T) {
	repo := &mockRepository{
		headers: map[string]Conversation{
			conversationKey(7, 9): {ID: 9, Type: "direct", Title: "王小明"},
		},
		messages: map[int64][]Message{
			9: {
				{ID: 1, ConversationID: 9, SenderID: 7, SenderName: "你", MessageType: "text", Content: "hello", CreatedAt: time.Date(2026, 4, 7, 1, 0, 0, 0, time.UTC)},
				{ID: 2, ConversationID: 9, SenderID: 8, SenderName: "王小明", MessageType: "text", Content: "hi", CreatedAt: time.Date(2026, 4, 7, 1, 1, 0, 0, time.UTC)},
			},
		},
	}
	service := NewService(repo)

	resp, status, err := service.ListMessages(9, SessionPrincipal{UserID: 7})
	if err != nil {
		t.Fatalf("ListMessages returned error: %v", err)
	}
	if status != 200 {
		t.Fatalf("status = %d, want 200", status)
	}

	data, ok := resp.Data.(MessageListData)
	if !ok {
		t.Fatalf("unexpected data type %T", resp.Data)
	}
	if data.ConversationID != 9 || data.Title != "王小明" {
		t.Fatalf("unexpected message list data: %+v", data)
	}
	if len(data.Messages) != 2 {
		t.Fatalf("len(messages) = %d, want 2", len(data.Messages))
	}
}

func TestListMessagesRejectsMissingConversation(t *testing.T) {
	service := NewService(&mockRepository{})

	_, _, err := service.ListMessages(9, SessionPrincipal{UserID: 7})
	if !errors.Is(err, ErrConversationNotFound) {
		t.Fatalf("expected ErrConversationNotFound, got %v", err)
	}
}

func TestSendMessage(t *testing.T) {
	repo := &mockRepository{
		headers: map[string]Conversation{
			conversationKey(7, 9): {ID: 9, Type: "direct", Title: "王小明"},
		},
		createMessageResult: Message{
			ID:             3,
			ConversationID: 9,
			SenderID:       7,
			SenderName:     "你",
			MessageType:    "text",
			Content:        "hello",
			CreatedAt:      time.Date(2026, 4, 7, 1, 2, 0, 0, time.UTC),
		},
	}
	service := NewService(repo)

	resp, status, err := service.SendMessage(9, SessionPrincipal{UserID: 7}, CreateMessageRequest{Type: "TEXT", Content: "  hello  "})
	if err != nil {
		t.Fatalf("SendMessage returned error: %v", err)
	}
	if status != 201 {
		t.Fatalf("status = %d, want 201", status)
	}
	if len(repo.createdMessages) != 1 {
		t.Fatalf("len(createdMessages) = %d, want 1", len(repo.createdMessages))
	}
	if repo.createdMessages[0].MessageType != "text" || repo.createdMessages[0].Content != "hello" {
		t.Fatalf("unexpected created message input: %+v", repo.createdMessages[0])
	}

	data, ok := resp.Data.(SentMessageData)
	if !ok {
		t.Fatalf("unexpected data type %T", resp.Data)
	}
	if data.ConversationID != 9 || data.Message.MessageID != 3 {
		t.Fatalf("unexpected sent message data: %+v", data)
	}
}

func TestSendMessageRejectsUnsupportedType(t *testing.T) {
	service := NewService(&mockRepository{})

	_, status, err := service.SendMessage(9, SessionPrincipal{UserID: 7}, CreateMessageRequest{Type: "image", Content: "hello"})
	if !errors.Is(err, ErrUnsupportedMessageType) {
		t.Fatalf("expected ErrUnsupportedMessageType, got %v", err)
	}
	if status != 400 {
		t.Fatalf("status = %d, want 400", status)
	}
}

func TestSendMessageRejectsEmptyContent(t *testing.T) {
	service := NewService(&mockRepository{})

	_, status, err := service.SendMessage(9, SessionPrincipal{UserID: 7}, CreateMessageRequest{Type: "text", Content: "   "})
	if !errors.Is(err, ErrMessageContentRequired) {
		t.Fatalf("expected ErrMessageContentRequired, got %v", err)
	}
	if status != 400 {
		t.Fatalf("status = %d, want 400", status)
	}
}

func TestSendMessageRejectsMissingConversation(t *testing.T) {
	service := NewService(&mockRepository{})

	_, status, err := service.SendMessage(9, SessionPrincipal{UserID: 7}, CreateMessageRequest{Type: "text", Content: "hello"})
	if !errors.Is(err, ErrConversationNotFound) {
		t.Fatalf("expected ErrConversationNotFound, got %v", err)
	}
	if status != 404 {
		t.Fatalf("status = %d, want 404", status)
	}
}
