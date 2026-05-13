package chat

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
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
	handler := NewHandler(NewService(&mockRepository{}, nil), stubSessionAuthenticator{session: auth.Session{UserID: 7}}, nil, nil)
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
	handler := NewHandler(NewService(repo, nil), stubSessionAuthenticator{session: auth.Session{UserID: 7}}, nil, nil)
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

func TestCreateDirectConversationHandlerSuccess(t *testing.T) {
	handler := NewHandler(NewService(&mockRepository{
		directConversation: Conversation{ID: 13, Type: "direct", Title: "User B"},
		directCreated:      true,
	}, nil), stubSessionAuthenticator{session: auth.Session{UserID: 7}}, nil, nil)

	request := httptest.NewRequest(http.MethodPost, "/api/conversations/direct", bytes.NewBufferString(`{"source_system":"erp","external_user_id":"user_b"}`))
	request.Header.Set("Authorization", "Bearer token")

	recorder := httptest.NewRecorder()
	handler.CreateDirectConversation(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", recorder.Code)
	}

	var resp struct {
		Success bool                   `json:"success"`
		Code    string                 `json:"code"`
		Data    DirectConversationData `json:"data"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.Success || resp.Data.ConversationID != 13 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestSendMessageHandlerMultipartSuccess(t *testing.T) {
	repo := &mockRepository{
		headers: map[string]Conversation{
			conversationKey(7, 9): {ID: 9, Type: "direct", Title: "王小明"},
		},
		createMessageResult: Message{
			ID:             12,
			ConversationID: 9,
			SenderID:       7,
			SenderName:     "你",
			MessageType:    "image",
			Content:        "sample.png",
			CreatedAt:      time.Date(2026, 4, 7, 3, 5, 0, 0, time.UTC),
			Attachment: &Attachment{
				OriginalName: "sample.png",
				StoragePath:  "/uploads/sample.png",
				MIMEType:     "image/png",
				SizeBytes:    4,
			},
		},
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("content", "")
	part, err := writer.CreateFormFile("file", "sample.png")
	if err != nil {
		t.Fatalf("CreateFormFile error: %v", err)
	}
	if _, err := io.WriteString(part, "png!"); err != nil {
		t.Fatalf("WriteString error: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close error: %v", err)
	}

	handler := NewHandler(
		NewService(repo, nil),
		stubSessionAuthenticator{session: auth.Session{UserID: 7}},
		nil,
		NewLocalFileStore(t.TempDir()),
	)
	request := httptest.NewRequest(http.MethodPost, "/api/conversations/9/messages", &body)
	request.Header.Set("Authorization", "Bearer token")
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.SetPathValue("conversation_id", "9")

	recorder := httptest.NewRecorder()
	handler.SendMessage(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201, body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestSendMessageHandlerMultipartMultipleFilesCreatesOneMessage(t *testing.T) {
	repo := &mockRepository{
		headers: map[string]Conversation{
			conversationKey(7, 9): {ID: 9, Type: "direct", Title: "王小明"},
		},
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("content", "caption")
	for _, name := range []string{"sample.png", "sample.docx"} {
		part, err := writer.CreateFormFile("file", name)
		if err != nil {
			t.Fatalf("CreateFormFile error: %v", err)
		}
		if _, err := io.WriteString(part, "file"); err != nil {
			t.Fatalf("WriteString error: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close error: %v", err)
	}

	handler := NewHandler(
		NewService(repo, nil),
		stubSessionAuthenticator{session: auth.Session{UserID: 7}},
		nil,
		NewLocalFileStore(t.TempDir()),
	)
	request := httptest.NewRequest(http.MethodPost, "/api/conversations/9/messages", &body)
	request.Header.Set("Authorization", "Bearer token")
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.SetPathValue("conversation_id", "9")

	recorder := httptest.NewRecorder()
	handler.SendMessage(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201, body=%s", recorder.Code, recorder.Body.String())
	}
	if len(repo.createdMessages) != 1 {
		t.Fatalf("created message count = %d, want 1", len(repo.createdMessages))
	}
	if got := len(repo.createdMessages[0].Attachments); got != 2 {
		t.Fatalf("attachment count = %d, want 2", got)
	}
	if repo.createdMessages[0].Content != "caption" {
		t.Fatalf("content = %q, want caption", repo.createdMessages[0].Content)
	}
}
