package chat

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"1001-twacc-chat/internal/auth"
)

// Handler exposes chat HTTP endpoints.
type Handler struct {
	service  *Service
	sessions SessionAuthenticator
}

// NewHandler builds the HTTP handler for chat endpoints.
func NewHandler(service *Service, sessions SessionAuthenticator) *Handler {
	return &Handler{service: service, sessions: sessions}
}

// ListConversations handles GET /api/conversations.
func (h *Handler) ListConversations(w http.ResponseWriter, r *http.Request) {
	if h.service == nil || h.sessions == nil {
		writeJSON(w, http.StatusServiceUnavailable, Response{Success: false, Code: "SERVICE_UNAVAILABLE", Message: "服务尚未完成初始化"})
		return
	}

	actor, ok := h.requireSession(w, r)
	if !ok {
		return
	}

	resp, status, err := h.service.ListConversations(actor)
	if err != nil {
		writeJSON(w, status, errorResponse(err))
		return
	}

	writeJSON(w, status, resp)
}

// ListMessages handles GET /api/conversations/{conversation_id}/messages.
func (h *Handler) ListMessages(w http.ResponseWriter, r *http.Request) {
	if h.service == nil || h.sessions == nil {
		writeJSON(w, http.StatusServiceUnavailable, Response{Success: false, Code: "SERVICE_UNAVAILABLE", Message: "服务尚未完成初始化"})
		return
	}

	actor, ok := h.requireSession(w, r)
	if !ok {
		return
	}

	conversationID, err := strconv.ParseInt(strings.TrimSpace(r.PathValue("conversation_id")), 10, 64)
	if err != nil || conversationID <= 0 {
		writeJSON(w, http.StatusBadRequest, Response{Success: false, Code: "INVALID_REQUEST", Message: "conversation_id 格式错误"})
		return
	}

	resp, status, svcErr := h.service.ListMessages(conversationID, actor)
	if svcErr != nil {
		writeJSON(w, status, errorResponse(svcErr))
		return
	}

	writeJSON(w, status, resp)
}

// CreateDirectConversation handles POST /api/conversations/direct.
func (h *Handler) CreateDirectConversation(w http.ResponseWriter, r *http.Request) {
	if h.service == nil || h.sessions == nil {
		writeJSON(w, http.StatusServiceUnavailable, Response{Success: false, Code: "SERVICE_UNAVAILABLE", Message: "服务尚未完成初始化"})
		return
	}

	actor, ok := h.requireSession(w, r)
	if !ok {
		return
	}

	var req CreateDirectConversationRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Success: false, Code: "INVALID_REQUEST", Message: "请求格式错误"})
		return
	}

	resp, status, svcErr := h.service.CreateDirectConversation(actor, req)
	if svcErr != nil {
		writeJSON(w, status, errorResponse(svcErr))
		return
	}

	writeJSON(w, status, resp)
}

// SendMessage handles POST /api/conversations/{conversation_id}/messages.
func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	if h.service == nil || h.sessions == nil {
		writeJSON(w, http.StatusServiceUnavailable, Response{Success: false, Code: "SERVICE_UNAVAILABLE", Message: "服务尚未完成初始化"})
		return
	}

	actor, ok := h.requireSession(w, r)
	if !ok {
		return
	}

	conversationID, err := strconv.ParseInt(strings.TrimSpace(r.PathValue("conversation_id")), 10, 64)
	if err != nil || conversationID <= 0 {
		writeJSON(w, http.StatusBadRequest, Response{Success: false, Code: "INVALID_REQUEST", Message: "conversation_id 格式错误"})
		return
	}

	var req CreateMessageRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Success: false, Code: "INVALID_REQUEST", Message: "请求格式错误"})
		return
	}

	resp, status, svcErr := h.service.SendMessage(conversationID, actor, req)
	if svcErr != nil {
		writeJSON(w, status, errorResponse(svcErr))
		return
	}

	writeJSON(w, status, resp)
}

func (h *Handler) requireSession(w http.ResponseWriter, r *http.Request) (SessionPrincipal, bool) {
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		writeJSON(w, http.StatusUnauthorized, Response{Success: false, Code: "UNAUTHORIZED", Message: "缺少有效 session token"})
		return SessionPrincipal{}, false
	}

	token := strings.TrimSpace(strings.TrimPrefix(authHeader, prefix))
	session, err := h.sessions.Authenticate(token)
	if err != nil {
		if errors.Is(err, auth.ErrSessionExpired) {
			writeJSON(w, http.StatusUnauthorized, Response{Success: false, Code: "SESSION_EXPIRED", Message: "session 已过期，请重新登录"})
			return SessionPrincipal{}, false
		}
		writeJSON(w, http.StatusUnauthorized, Response{Success: false, Code: "UNAUTHORIZED", Message: "session token 无效"})
		return SessionPrincipal{}, false
	}

	return session, true
}

func decodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return err
	}

	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func errorResponse(err error) Response {
	switch {
	case errors.Is(err, ErrConversationNotFound):
		return Response{Success: false, Code: "CONVERSATION_NOT_FOUND", Message: "找不到指定对话"}
	case errors.Is(err, ErrTargetUserNotFound):
		return Response{Success: false, Code: "USER_NOT_FOUND", Message: "找不到目标使用者"}
	case errors.Is(err, ErrDirectChatSelfNotAllow):
		return Response{Success: false, Code: "INVALID_REQUEST", Message: "不可与自己建立一对一对话"}
	case errors.Is(err, ErrUnsupportedMessageType):
		return Response{Success: false, Code: "INVALID_REQUEST", Message: "目前仅支持 text 讯息"}
	case errors.Is(err, ErrMessageContentRequired):
		return Response{Success: false, Code: "INVALID_REQUEST", Message: "content 不可为空"}
	case errors.Is(err, ErrSystemAdminCannotChat):
		return Response{Success: false, Code: "SYSTEM_ADMIN_CANNOT_CHAT", Message: "system_admin 不可使用聊天功能"}
	case errors.Is(err, ErrInsufficientRole):
		return Response{Success: false, Code: "INSUFFICIENT_ROLE", Message: "目前角色无权执行此操作"}
	default:
		return Response{Success: false, Code: "INTERNAL_ERROR", Message: "系统内部错误"}
	}
}
