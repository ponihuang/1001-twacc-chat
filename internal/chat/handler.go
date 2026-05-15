package chat

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"1001-twacc-chat/internal/auth"
)

// Handler exposes chat HTTP endpoints.
type Handler struct {
	service  *Service
	sessions SessionAuthenticator
	broker   Broker
	files    FileStore
}

// NewHandler builds the HTTP handler for chat endpoints.
func NewHandler(service *Service, sessions SessionAuthenticator, broker Broker, files FileStore) *Handler {
	return &Handler{service: service, sessions: sessions, broker: broker, files: files}
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

// SearchUsers handles GET /api/users/search?q=...
func (h *Handler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	if h.service == nil || h.sessions == nil {
		writeJSON(w, http.StatusServiceUnavailable, Response{Success: false, Code: "SERVICE_UNAVAILABLE", Message: "服务尚未完成初始化"})
		return
	}

	actor, ok := h.requireSession(w, r)
	if !ok {
		return
	}

	resp, status, svcErr := h.service.SearchUsers(actor, r.URL.Query().Get("source_system"), r.URL.Query().Get("q"))
	if svcErr != nil {
		writeJSON(w, status, errorResponse(svcErr))
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

// SearchMessages handles GET /api/messages/search?q=...
func (h *Handler) SearchMessages(w http.ResponseWriter, r *http.Request) {
	if h.service == nil || h.sessions == nil {
		writeJSON(w, http.StatusServiceUnavailable, Response{Success: false, Code: "SERVICE_UNAVAILABLE", Message: "服务尚未完成初始化"})
		return
	}

	actor, ok := h.requireSession(w, r)
	if !ok {
		return
	}

	resp, status, svcErr := h.service.SearchMessages(actor, r.URL.Query().Get("q"))
	if svcErr != nil {
		writeJSON(w, status, errorResponse(svcErr))
		return
	}

	writeJSON(w, status, resp)
}

// RecallMessage handles POST /api/conversations/{conversation_id}/messages/{message_id}/recall.
func (h *Handler) RecallMessage(w http.ResponseWriter, r *http.Request) {
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
	messageID, err := strconv.ParseInt(strings.TrimSpace(r.PathValue("message_id")), 10, 64)
	if err != nil || messageID <= 0 {
		writeJSON(w, http.StatusBadRequest, Response{Success: false, Code: "INVALID_REQUEST", Message: "message_id 格式错误"})
		return
	}

	resp, status, svcErr := h.service.DeleteMessage(conversationID, messageID, actor)
	if svcErr != nil {
		writeJSON(w, status, errorResponse(svcErr))
		return
	}

	writeJSON(w, status, resp)
}

// DeleteMessage handles the legacy DELETE recall route.
func (h *Handler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	h.RecallMessage(w, r)
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

// CreateGroupConversation handles POST /api/conversations/group.
func (h *Handler) CreateGroupConversation(w http.ResponseWriter, r *http.Request) {
	if h.service == nil || h.sessions == nil {
		writeJSON(w, http.StatusServiceUnavailable, Response{Success: false, Code: "SERVICE_UNAVAILABLE", Message: "服务尚未完成初始化"})
		return
	}

	actor, ok := h.requireSession(w, r)
	if !ok {
		return
	}

	var req CreateGroupConversationRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Success: false, Code: "INVALID_REQUEST", Message: "请求格式错误"})
		return
	}

	resp, status, svcErr := h.service.CreateGroupConversation(actor, req)
	if svcErr != nil {
		writeJSON(w, status, errorResponse(svcErr))
		return
	}

	writeJSON(w, status, resp)
}

// ServeWebSocket upgrades the request and streams chat events for the current user.
func (h *Handler) ServeWebSocket(w http.ResponseWriter, r *http.Request) {
	if h.service == nil || h.sessions == nil || h.broker == nil {
		writeJSON(w, http.StatusServiceUnavailable, Response{Success: false, Code: "SERVICE_UNAVAILABLE", Message: "服务尚未完成初始化"})
		return
	}

	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		writeJSON(w, http.StatusUnauthorized, Response{Success: false, Code: "UNAUTHORIZED", Message: "缺少有效 session token"})
		return
	}
	if !websocketUpgradeRequested(r) {
		writeJSON(w, http.StatusBadRequest, Response{Success: false, Code: "INVALID_REQUEST", Message: "缺少 websocket upgrade header"})
		return
	}
	if !ensureWebSocketHandshake(w, r) {
		return
	}

	session, err := h.sessions.Authenticate(token)
	if err != nil {
		if errors.Is(err, auth.ErrSessionExpired) {
			writeJSON(w, http.StatusUnauthorized, Response{Success: false, Code: "SESSION_EXPIRED", Message: "session 已过期，请重新登录"})
			return
		}
		writeJSON(w, http.StatusUnauthorized, Response{Success: false, Code: "UNAUTHORIZED", Message: "session token 无效"})
		return
	}

	if err := h.service.requireChatUser(session.UserID); err != nil {
		writeJSON(w, statusCode(err), errorResponse(err))
		return
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "websocket unsupported", http.StatusInternalServerError)
		return
	}

	conn, rw, err := hijacker.Hijack()
	if err != nil {
		return
	}
	if err := writeHandshakeResponse(rw, r.Header.Get("Sec-WebSocket-Key")); err != nil {
		_ = conn.Close()
		return
	}

	events, unsubscribe := h.broker.Subscribe(session.UserID)
	defer unsubscribe()
	defer conn.Close()

	go drainIncomingFrames(conn)

	pingTicker := time.NewTicker(25 * time.Second)
	defer pingTicker.Stop()

	for {
		select {
		case event, ok := <-events:
			if !ok {
				return
			}
			if err := writeWebSocketJSON(conn, event); err != nil {
				return
			}
		case <-pingTicker.C:
			if err := writeWebSocketControl(conn, websocketOpcodePing, nil); err != nil {
				return
			}
		case <-r.Context().Done():
			return
		}
	}
}

const (
	websocketOpcodeText = 0x1
	websocketOpcodePing = 0x9
)

func websocketUpgradeRequested(r *http.Request) bool {
	if !strings.EqualFold(strings.TrimSpace(r.Header.Get("Upgrade")), "websocket") {
		return false
	}

	for _, token := range strings.Split(r.Header.Get("Connection"), ",") {
		if strings.EqualFold(strings.TrimSpace(token), "Upgrade") {
			return true
		}
	}

	return false
}

func ensureWebSocketHandshake(w http.ResponseWriter, r *http.Request) bool {
	if !strings.EqualFold(strings.TrimSpace(r.Header.Get("Sec-WebSocket-Version")), "13") {
		writeJSON(w, http.StatusBadRequest, Response{Success: false, Code: "INVALID_REQUEST", Message: "只支持 websocket version 13"})
		return false
	}
	if strings.TrimSpace(r.Header.Get("Sec-WebSocket-Key")) == "" {
		writeJSON(w, http.StatusBadRequest, Response{Success: false, Code: "INVALID_REQUEST", Message: "缺少 websocket key"})
		return false
	}
	return true
}

func writeHandshakeResponse(rw *bufio.ReadWriter, key string) error {
	sum := sha1.Sum([]byte(strings.TrimSpace(key) + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	accept := base64.StdEncoding.EncodeToString(sum[:])

	lines := []string{
		"HTTP/1.1 101 Switching Protocols\r\n",
		"Upgrade: websocket\r\n",
		"Connection: Upgrade\r\n",
		"Sec-WebSocket-Accept: " + accept + "\r\n",
		"\r\n",
	}

	for _, line := range lines {
		if _, err := rw.WriteString(line); err != nil {
			return err
		}
	}

	return rw.Flush()
}

func writeWebSocketJSON(conn net.Conn, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return writeWebSocketFrame(conn, websocketOpcodeText, data)
}

func writeWebSocketControl(conn net.Conn, opcode byte, payload []byte) error {
	return writeWebSocketFrame(conn, opcode, payload)
}

func writeWebSocketFrame(conn net.Conn, opcode byte, payload []byte) error {
	frame := []byte{0x80 | opcode}

	switch length := len(payload); {
	case length < 126:
		frame = append(frame, byte(length))
	case length <= 65535:
		frame = append(frame, 126, byte(length>>8), byte(length))
	default:
		frame = append(frame,
			127,
			byte(uint64(length)>>56),
			byte(uint64(length)>>48),
			byte(uint64(length)>>40),
			byte(uint64(length)>>32),
			byte(uint64(length)>>24),
			byte(uint64(length)>>16),
			byte(uint64(length)>>8),
			byte(uint64(length)),
		)
	}

	if _, err := conn.Write(frame); err != nil {
		return err
	}
	if len(payload) == 0 {
		return nil
	}
	_, err := conn.Write(payload)
	return err
}

func drainIncomingFrames(conn net.Conn) {
	header := make([]byte, 2)

	for {
		if _, err := io.ReadFull(conn, header); err != nil {
			return
		}

		payloadLen := int(header[1] & 0x7f)
		switch payloadLen {
		case 126:
			extended := make([]byte, 2)
			if _, err := io.ReadFull(conn, extended); err != nil {
				return
			}
			payloadLen = int(extended[0])<<8 | int(extended[1])
		case 127:
			extended := make([]byte, 8)
			if _, err := io.ReadFull(conn, extended); err != nil {
				return
			}
			payloadLen = int(uint64(extended[0])<<56 |
				uint64(extended[1])<<48 |
				uint64(extended[2])<<40 |
				uint64(extended[3])<<32 |
				uint64(extended[4])<<24 |
				uint64(extended[5])<<16 |
				uint64(extended[6])<<8 |
				uint64(extended[7]))
		}

		if header[1]&0x80 != 0 {
			mask := make([]byte, 4)
			if _, err := io.ReadFull(conn, mask); err != nil {
				return
			}
		}

		if payloadLen > 0 {
			payload := make([]byte, payloadLen)
			if _, err := io.ReadFull(conn, payload); err != nil {
				return
			}
		}

		if header[0]&0x0f == 0x8 {
			return
		}
	}
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

	req, err := h.decodeMessageRequest(r)
	if err != nil {
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

func (h *Handler) decodeMessageRequest(r *http.Request) (CreateMessageRequest, error) {
	contentType := strings.TrimSpace(r.Header.Get("Content-Type"))
	mediaType, _, _ := mime.ParseMediaType(contentType)
	if strings.EqualFold(mediaType, "multipart/form-data") {
		return h.decodeMultipartMessage(r)
	}

	var req CreateMessageRequest
	if err := decodeJSON(r, &req); err != nil {
		return CreateMessageRequest{}, err
	}
	return req, nil
}

func (h *Handler) decodeMultipartMessage(r *http.Request) (CreateMessageRequest, error) {
	if h.files == nil {
		return CreateMessageRequest{}, ErrAttachmentRequired
	}

	if err := r.ParseMultipartForm(maxAttachmentSize + (1 << 20)); err != nil {
		return CreateMessageRequest{}, err
	}

	req := CreateMessageRequest{
		Type:    strings.TrimSpace(r.FormValue("type")),
		Content: strings.TrimSpace(r.FormValue("content")),
	}

	if r.MultipartForm == nil || len(r.MultipartForm.File["file"]) == 0 {
		return req, nil
	}

	hasFileAttachment := false
	for _, header := range r.MultipartForm.File["file"] {
		file, err := header.Open()
		if err != nil {
			return CreateMessageRequest{}, err
		}
		attachment, detectedType, err := h.files.Save(file, header)
		closeErr := file.Close()
		if err != nil {
			return CreateMessageRequest{}, err
		}
		if closeErr != nil {
			return CreateMessageRequest{}, closeErr
		}

		if req.Type == "" {
			req.Type = detectedType
		}
		if detectedType == fileMessageType {
			hasFileAttachment = true
		}
		req.Attachments = append(req.Attachments, *attachment)
	}

	if hasFileAttachment {
		req.Type = fileMessageType
	}
	if len(req.Attachments) == 1 {
		req.Attachment = &req.Attachments[0]
	}
	return req, nil
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
		return Response{Success: false, Code: "INVALID_REQUEST", Message: "目前仅支持 text / image / file 讯息"}
	case errors.Is(err, ErrMessageContentRequired):
		return Response{Success: false, Code: "INVALID_REQUEST", Message: "content 不可为空"}
	case errors.Is(err, ErrAttachmentRequired):
		return Response{Success: false, Code: "INVALID_REQUEST", Message: "附件不可为空"}
	case errors.Is(err, ErrAttachmentTooLarge):
		return Response{Success: false, Code: "INVALID_REQUEST", Message: "附件大小超过 40MB 限制"}
	case errors.Is(err, ErrUnsupportedAttachmentType):
		return Response{Success: false, Code: "INVALID_REQUEST", Message: "附件格式不支持"}
	case errors.Is(err, ErrSystemAdminCannotChat):
		return Response{Success: false, Code: "SYSTEM_ADMIN_CANNOT_CHAT", Message: "system_admin 不可使用聊天功能"}
	case errors.Is(err, ErrInsufficientRole):
		return Response{Success: false, Code: "INSUFFICIENT_ROLE", Message: "目前角色无权执行此操作"}
	default:
		return Response{Success: false, Code: "INTERNAL_ERROR", Message: "系统内部错误"}
	}
}
