package chat

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"1001-twacc-chat/internal/auth"
)

type stubSessionAuthenticator struct {
	session auth.Session
	err     error
}

func (s stubSessionAuthenticator) Authenticate(token string) (auth.Session, error) {
	return s.session, s.err
}

func TestSendMessageHandlerRejectsInvalidJSON(t *testing.T) {
	handler := NewHandler(NewService(&mockRepository{}), stubSessionAuthenticator{session: auth.Session{UserID: 7}})
	request := httptest.NewRequest(http.MethodPost, "/api/conversations/9/messages", bytes.NewBufferString(`{"type":"text"`))
	request.Header.Set("Authorization", "Bearer token")
	request.SetPathValue("conversation_id", "9")

	recorder := httptest.NewRecorder()
	handler.SendMessage(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", recorder.Code)
	}

	var resp Response
	if err := json.NewDecoder(recorder.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Code != "INVALID_REQUEST" {
		t.Fatalf("code = %q, want INVALID_REQUEST", resp.Code)
	}
}

func TestSendMessageHandlerSuccess(t *testing.T) {
	repo := &mockRepository{
		headers: map[string]Conversation{
			conversationKey(7, 9): {ID: 9, Type: "direct", Title: "王小明"},
		},
		createMessageResult: Message{
			ID:             11,
			ConversationID: 9,
			SenderID:       7,
			SenderName:     "你",
			MessageType:    "text",
			Content:        "hello",
			CreatedAt:      time.Date(2026, 4, 7, 3, 0, 0, 0, time.UTC),
		},
	}
	handler := NewHandler(NewService(repo), stubSessionAuthenticator{session: auth.Session{UserID: 7}})
	request := httptest.NewRequest(http.MethodPost, "/api/conversations/9/messages", bytes.NewBufferString(`{"type":"text","content":"hello"}`))
	request.Header.Set("Authorization", "Bearer token")
	request.SetPathValue("conversation_id", "9")

	recorder := httptest.NewRecorder()
	handler.SendMessage(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", recorder.Code)
	}

	var resp struct {
		Success bool            `json:"success"`
		Code    string          `json:"code"`
		Data    SentMessageData `json:"data"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.Success || resp.Code != "MESSAGE_SENT" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if resp.Data.ConversationID != 9 || resp.Data.Message.MessageID != 11 {
		t.Fatalf("unexpected sent message payload: %+v", resp.Data)
	}
}
