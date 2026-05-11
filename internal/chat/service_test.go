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
	searchUsers         []UserSearchResult
	searchMessages      []Message
	searchUserID        int64
	searchSourceSystem  string
	searchQuery         string
	searchLimit         int
	memberIDs           map[int64][]int64
	markedReads         []readMarker
	createdMessages     []CreateMessageInput
	createMessageResult Message
	createMessageErr    error
	directConversation  Conversation
	directCreated       bool
	directErr           error
}

type readMarker struct {
	userID         int64
	conversationID int64
}

func (m *mockRepository) IsSystemAdmin(userID int64) (bool, error) {
	return m.admins[userID], nil
}

func (m *mockRepository) ListConversations(userID int64) ([]ConversationSummary, error) {
	return append([]ConversationSummary(nil), m.conversations[userID]...), nil
}

func (m *mockRepository) SearchUsers(actorUserID int64, sourceSystem, query string, limit int) ([]UserSearchResult, error) {
	m.searchUserID = actorUserID
	m.searchSourceSystem = sourceSystem
	m.searchQuery = query
	m.searchLimit = limit
	return append([]UserSearchResult(nil), m.searchUsers...), nil
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

func (m *mockRepository) SearchMessages(userID int64, query string, limit int) ([]Message, error) {
	m.searchUserID = userID
	m.searchQuery = query
	m.searchLimit = limit
	return append([]Message(nil), m.searchMessages...), nil
}

func (m *mockRepository) MarkConversationRead(userID, conversationID int64) error {
	m.markedReads = append(m.markedReads, readMarker{userID: userID, conversationID: conversationID})
	return nil
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
		Attachment:     attachmentFromInput(input.Attachment),
	}, nil
}

func (m *mockRepository) CreateOrGetDirectConversation(actorUserID int64, sourceSystem, externalUserID string) (Conversation, bool, error) {
	if m.directErr != nil {
		return Conversation{}, false, m.directErr
	}
	if m.directConversation.ID != 0 {
		return m.directConversation, m.directCreated, nil
	}
	return Conversation{ID: 15, Type: "direct", Title: externalUserID}, true, nil
}

func (m *mockRepository) ListConversationMemberIDs(conversationID int64) ([]int64, error) {
	return append([]int64(nil), m.memberIDs[conversationID]...), nil
}

type mockBroker struct {
	userIDs []int64
	event   RealtimeEvent
	calls   int
}

func (m *mockBroker) PublishToUsers(userIDs []int64, event RealtimeEvent) {
	m.userIDs = append([]int64(nil), userIDs...)
	m.event = event
	m.calls++
}

func (m *mockBroker) Subscribe(userID int64) (<-chan RealtimeEvent, func()) {
	ch := make(chan RealtimeEvent)
	close(ch)
	return ch, func() {}
}

func attachmentFromInput(value *AttachmentInput) *Attachment {
	if value == nil {
		return nil
	}

	return &Attachment{
		OriginalName: value.OriginalName,
		StoragePath:  value.StoragePath,
		MIMEType:     value.MIMEType,
		SizeBytes:    value.SizeBytes,
	}
}

func conversationKey(userID, conversationID int64) string {
	return time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(userID + conversationID)).Format(time.RFC3339Nano)
}

func TestListConversations(t *testing.T) {
	repo := &mockRepository{
		conversations: map[int64][]ConversationSummary{
			7: {{ConversationID: 9, Type: "direct", Title: "採購小組", DirectSourceSystem: "erp", DirectExternalID: "user_b", MemberCount: 3, LastMessageType: "text", LastMessagePreview: "hello", LastMessageAt: time.Date(2026, 4, 7, 1, 2, 3, 0, time.UTC), UnreadCount: 2}},
		},
	}
	service := NewService(repo, nil)

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
	if items[0].UnreadCount != 2 {
		t.Fatalf("unread_count = %d, want 2", items[0].UnreadCount)
	}
	if items[0].DirectSourceSystem != "erp" || items[0].DirectExternalID != "user_b" {
		t.Fatalf("unexpected direct peer fields: %+v", items[0])
	}
}

func TestListConversationsRejectsSystemAdmin(t *testing.T) {
	repo := &mockRepository{admins: map[int64]bool{9: true}}
	service := NewService(repo, nil)

	_, _, err := service.ListConversations(SessionPrincipal{UserID: 9})
	if !errors.Is(err, ErrSystemAdminCannotChat) {
		t.Fatalf("expected ErrSystemAdminCannotChat, got %v", err)
	}
}

func TestSearchUsers(t *testing.T) {
	repo := &mockRepository{
		admins: map[int64]bool{},
		searchUsers: []UserSearchResult{
			{SourceSystem: "erp", ExternalUserID: "test01", DisplayName: "Test One"},
		},
	}
	service := NewService(repo, nil)

	resp, status, err := service.SearchUsers(SessionPrincipal{UserID: 7}, "erp", "@test")
	if err != nil {
		t.Fatalf("SearchUsers returned error: %v", err)
	}
	if status != 200 || !resp.Success {
		t.Fatalf("unexpected response status=%d resp=%+v", status, resp)
	}
	if repo.searchUserID != 7 || repo.searchSourceSystem != "erp" || repo.searchQuery != "test" || repo.searchLimit != userSearchLimit {
		t.Fatalf("unexpected search call: user=%d source=%q query=%q limit=%d", repo.searchUserID, repo.searchSourceSystem, repo.searchQuery, repo.searchLimit)
	}
	data, ok := resp.Data.(UserSearchData)
	if !ok {
		t.Fatalf("response data type = %T, want UserSearchData", resp.Data)
	}
	if len(data.Users) != 1 || data.Users[0].ExternalUserID != "test01" || data.Users[0].DisplayName != "Test One" {
		t.Fatalf("unexpected search data: %+v", data)
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
	service := NewService(repo, nil)

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
	if len(repo.markedReads) != 1 || repo.markedReads[0] != (readMarker{userID: 7, conversationID: 9}) {
		t.Fatalf("unexpected read markers: %+v", repo.markedReads)
	}
}

func TestListMessagesRejectsMissingConversation(t *testing.T) {
	service := NewService(&mockRepository{}, nil)

	_, _, err := service.ListMessages(9, SessionPrincipal{UserID: 7})
	if !errors.Is(err, ErrConversationNotFound) {
		t.Fatalf("expected ErrConversationNotFound, got %v", err)
	}
}

func TestSearchMessages(t *testing.T) {
	repo := &mockRepository{
		admins: map[int64]bool{},
		searchMessages: []Message{
			{
				ID:                3,
				ConversationID:    9,
				ConversationTitle: "開發",
				SenderID:          8,
				SenderName:        "王小明",
				MessageType:       "text",
				Content:           "hello search",
				CreatedAt:         time.Date(2026, 4, 7, 1, 2, 0, 0, time.UTC),
			},
		},
	}
	service := NewService(repo, nil)

	resp, status, err := service.SearchMessages(SessionPrincipal{UserID: 7}, " search ")
	if err != nil {
		t.Fatalf("SearchMessages returned error: %v", err)
	}
	if status != 200 || !resp.Success {
		t.Fatalf("unexpected response status=%d resp=%+v", status, resp)
	}
	if repo.searchUserID != 7 || repo.searchQuery != "search" || repo.searchLimit != messageSearchLimit {
		t.Fatalf("unexpected search call: user=%d query=%q limit=%d", repo.searchUserID, repo.searchQuery, repo.searchLimit)
	}
	data, ok := resp.Data.(MessageSearchData)
	if !ok {
		t.Fatalf("response data type = %T, want MessageSearchData", resp.Data)
	}
	if len(data.Messages) != 1 || data.Messages[0].ConversationTitle != "開發" || data.Messages[0].Content != "hello search" {
		t.Fatalf("unexpected search data: %+v", data)
	}
}

func TestCreateDirectConversation(t *testing.T) {
	repo := &mockRepository{
		directConversation: Conversation{ID: 13, Type: "direct", Title: "User B"},
		directCreated:      true,
		memberIDs: map[int64][]int64{
			13: {7, 8},
		},
	}
	broker := &mockBroker{}
	service := NewService(repo, broker)

	resp, status, err := service.CreateDirectConversation(SessionPrincipal{UserID: 7}, CreateDirectConversationRequest{
		SourceSystem:   "erp",
		ExternalUserID: "user_b",
	})
	if err != nil {
		t.Fatalf("CreateDirectConversation returned error: %v", err)
	}
	if status != 201 {
		t.Fatalf("status = %d, want 201", status)
	}

	data, ok := resp.Data.(DirectConversationData)
	if !ok {
		t.Fatalf("unexpected data type %T", resp.Data)
	}
	if data.ConversationID != 13 || data.Title != "User B" {
		t.Fatalf("unexpected direct conversation data: %+v", data)
	}
	if broker.calls != 1 {
		t.Fatalf("broker calls = %d, want 1", broker.calls)
	}
	if broker.event.EventType != "conversation.ready" || broker.event.ConversationID != 13 {
		t.Fatalf("unexpected realtime event: %+v", broker.event)
	}
}

func TestSendMessage(t *testing.T) {
	repo := &mockRepository{
		headers: map[string]Conversation{
			conversationKey(7, 9): {ID: 9, Type: "direct", Title: "王小明"},
		},
		memberIDs: map[int64][]int64{
			9: {7, 8},
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
	broker := &mockBroker{}
	service := NewService(repo, broker)

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
	if broker.calls != 1 {
		t.Fatalf("broker calls = %d, want 1", broker.calls)
	}
	if broker.event.EventType != "message.created" || broker.event.ConversationID != 9 {
		t.Fatalf("unexpected realtime event: %+v", broker.event)
	}
}

func TestSendMessageRejectsUnsupportedType(t *testing.T) {
	service := NewService(&mockRepository{}, nil)

	_, status, err := service.SendMessage(9, SessionPrincipal{UserID: 7}, CreateMessageRequest{Type: "voice", Content: "hello"})
	if !errors.Is(err, ErrUnsupportedMessageType) {
		t.Fatalf("expected ErrUnsupportedMessageType, got %v", err)
	}
	if status != 400 {
		t.Fatalf("status = %d, want 400", status)
	}
}

func TestSendImageMessage(t *testing.T) {
	repo := &mockRepository{
		headers: map[string]Conversation{
			conversationKey(7, 9): {ID: 9, Type: "direct", Title: "王小明"},
		},
		memberIDs: map[int64][]int64{
			9: {7, 8},
		},
	}
	service := NewService(repo, nil)

	resp, status, err := service.SendMessage(9, SessionPrincipal{UserID: 7}, CreateMessageRequest{
		Type:    "image",
		Content: "",
		Attachment: &AttachmentInput{
			OriginalName: "photo.png",
			StoragePath:  "/uploads/photo.png",
			MIMEType:     "image/png",
			SizeBytes:    1234,
		},
	})
	if err != nil {
		t.Fatalf("SendMessage returned error: %v", err)
	}
	if status != 201 {
		t.Fatalf("status = %d, want 201", status)
	}

	data := resp.Data.(SentMessageData)
	if data.Message.Attachment == nil || data.Message.Attachment.URL != "/uploads/photo.png" {
		t.Fatalf("unexpected attachment payload: %+v", data.Message.Attachment)
	}
}

func TestSendMessageRejectsEmptyContent(t *testing.T) {
	service := NewService(&mockRepository{}, nil)

	_, status, err := service.SendMessage(9, SessionPrincipal{UserID: 7}, CreateMessageRequest{Type: "text", Content: "   "})
	if !errors.Is(err, ErrMessageContentRequired) {
		t.Fatalf("expected ErrMessageContentRequired, got %v", err)
	}
	if status != 400 {
		t.Fatalf("status = %d, want 400", status)
	}
}

func TestSendMessageRejectsMissingConversation(t *testing.T) {
	service := NewService(&mockRepository{}, nil)

	_, status, err := service.SendMessage(9, SessionPrincipal{UserID: 7}, CreateMessageRequest{Type: "text", Content: "hello"})
	if !errors.Is(err, ErrConversationNotFound) {
		t.Fatalf("expected ErrConversationNotFound, got %v", err)
	}
	if status != 404 {
		t.Fatalf("status = %d, want 404", status)
	}
}
