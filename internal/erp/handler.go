package erp

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"

	"1001-twacc-chat/internal/auth"
)

// Handler exposes external-system auth endpoints plus admin and device routes.
type Handler struct {
	service                *Service
	sessions               SessionAuthenticator
	integrationSharedToken string
}

// NewHandler builds the HTTP handler for external-system access flows.
func NewHandler(service *Service, sessions SessionAuthenticator, integrationSharedToken string) *Handler {
	return &Handler{service: service, sessions: sessions, integrationSharedToken: strings.TrimSpace(integrationSharedToken)}
}

// Register handles POST /api/erp/register.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if !h.integrationAuthorized(r) {
		writeJSON(w, http.StatusUnauthorized, Response{Success: false, Code: "INVALID_INTEGRATION_TOKEN", Message: "对接凭证错误"})
		return
	}
	if h.service == nil {
		writeJSON(w, http.StatusServiceUnavailable, Response{Success: false, Code: "SERVICE_UNAVAILABLE", Message: "服务尚未完成初始化"})
		return
	}

	var req RegisterRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Success: false, Code: "INVALID_REQUEST", Message: "请求格式错误"})
		return
	}

	resp, status, err := h.service.Register(req)
	if err != nil {
		writeJSON(w, status, errorResponse(err))
		return
	}

	writeJSON(w, status, resp)
}

// Login handles POST /api/erp/login.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if !h.integrationAuthorized(r) {
		writeJSON(w, http.StatusUnauthorized, Response{Success: false, Code: "INVALID_INTEGRATION_TOKEN", Message: "对接凭证错误"})
		return
	}
	if h.service == nil {
		writeJSON(w, http.StatusServiceUnavailable, Response{Success: false, Code: "SERVICE_UNAVAILABLE", Message: "服务尚未完成初始化"})
		return
	}

	var req LoginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Success: false, Code: "INVALID_REQUEST", Message: "请求格式错误"})
		return
	}

	resp, status, err := h.service.Login(req, clientIP(r), r.UserAgent())
	if err != nil {
		writeJSON(w, status, errorResponse(err))
		return
	}

	writeJSON(w, status, resp)
}

// ListDevices handles GET /api/system-admin/users/{user_id}/devices.
func (h *Handler) ListDevices(w http.ResponseWriter, r *http.Request) {
	if h.service == nil || h.sessions == nil {
		writeJSON(w, http.StatusServiceUnavailable, Response{Success: false, Code: "SERVICE_UNAVAILABLE", Message: "服务尚未完成初始化"})
		return
	}

	principal, ok := h.requireSession(w, r)
	if !ok {
		return
	}

	targetUserID, ok := parseUserIDPath(w, r)
	if !ok {
		return
	}

	resp, status, err := h.service.ListDevices(targetUserID, principal)
	if err != nil {
		writeJSON(w, status, errorResponse(err))
		return
	}

	writeJSON(w, status, resp)
}

// UpdateIPWhitelist handles PUT /api/system-admin/users/{user_id}/ip-whitelist.
func (h *Handler) UpdateIPWhitelist(w http.ResponseWriter, r *http.Request) {
	if h.service == nil || h.sessions == nil {
		writeJSON(w, http.StatusServiceUnavailable, Response{Success: false, Code: "SERVICE_UNAVAILABLE", Message: "服务尚未完成初始化"})
		return
	}

	principal, ok := h.requireSession(w, r)
	if !ok {
		return
	}

	targetUserID, ok := parseUserIDPath(w, r)
	if !ok {
		return
	}

	var req IPWhitelistUpdateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Success: false, Code: "INVALID_REQUEST", Message: "请求格式错误"})
		return
	}

	resp, status, err := h.service.UpdateIPWhitelist(targetUserID, principal, req)
	if err != nil {
		writeJSON(w, status, errorResponse(err))
		return
	}

	writeJSON(w, status, resp)
}

// ApproveDevice handles POST /api/users/{user_id}/devices/{device_id}/approve.
func (h *Handler) ApproveDevice(w http.ResponseWriter, r *http.Request) {
	if h.service == nil || h.sessions == nil {
		writeJSON(w, http.StatusServiceUnavailable, Response{Success: false, Code: "SERVICE_UNAVAILABLE", Message: "服务尚未完成初始化"})
		return
	}

	principal, ok := h.requireSession(w, r)
	if !ok {
		return
	}

	targetUserID, ok := parseUserIDPath(w, r)
	if !ok {
		return
	}

	targetDeviceID := strings.TrimSpace(r.PathValue("device_id"))
	if targetDeviceID == "" {
		writeJSON(w, http.StatusBadRequest, Response{Success: false, Code: "DEVICE_ID_REQUIRED", Message: "缺少 device_id"})
		return
	}

	resp, status, err := h.service.ApproveDevice(targetUserID, targetDeviceID, principal)
	if err != nil {
		writeJSON(w, status, errorResponse(err))
		return
	}

	writeJSON(w, status, resp)
}

func (h *Handler) integrationAuthorized(r *http.Request) bool {
	if h.integrationSharedToken == "" {
		return true
	}

	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		return false
	}

	return strings.TrimSpace(strings.TrimPrefix(authHeader, prefix)) == h.integrationSharedToken
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

func parseUserIDPath(w http.ResponseWriter, r *http.Request) (int64, bool) {
	targetUserID, err := strconv.ParseInt(strings.TrimSpace(r.PathValue("user_id")), 10, 64)
	if err != nil || targetUserID <= 0 {
		writeJSON(w, http.StatusBadRequest, Response{Success: false, Code: "INVALID_REQUEST", Message: "user_id 格式错误"})
		return 0, false
	}

	return targetUserID, true
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
	case errors.Is(err, ErrInvalidSourceSystem):
		return Response{Success: false, Code: "INVALID_SOURCE_SYSTEM", Message: "source_system 格式错误或未开放"}
	case errors.Is(err, ErrInvalidExternalUserID):
		return Response{Success: false, Code: "INVALID_EXTERNAL_USER_ID", Message: "external_user_id 不可为空"}
	case errors.Is(err, ErrInvalidDisplayName):
		return Response{Success: false, Code: "INVALID_REQUEST", Message: "display_name 不可为空"}
	case errors.Is(err, ErrInvalidUserID):
		return Response{Success: false, Code: "INVALID_REQUEST", Message: "user_id 格式错误"}
	case errors.Is(err, ErrDeviceIDRequired):
		return Response{Success: false, Code: "DEVICE_ID_REQUIRED", Message: "缺少 device_id"}
	case errors.Is(err, ErrInvalidIPWhitelistRule):
		return Response{Success: false, Code: "INVALID_REQUEST", Message: "白名单规则必须是单一 IP 或 CIDR"}
	case errors.Is(err, ErrWhitelistRulesRequired):
		return Response{Success: false, Code: "INVALID_REQUEST", Message: "关闭一律放行时至少要提供一条白名单规则"}
	case errors.Is(err, ErrUserNotFound):
		return Response{Success: false, Code: "USER_NOT_FOUND", Message: "查无对应用户"}
	case errors.Is(err, ErrDeviceNotFound):
		return Response{Success: false, Code: "NOT_FOUND", Message: "查无对应装置"}
	case errors.Is(err, ErrUserAlreadyExists):
		return Response{Success: false, Code: "USER_ALREADY_EXISTS", Message: "该外部用户已存在"}
	case errors.Is(err, ErrSystemAdminCannotChat):
		return Response{Success: false, Code: "SYSTEM_ADMIN_CANNOT_CHAT", Message: "system_admin 不可使用聊天功能"}
	case errors.Is(err, ErrInsufficientRole):
		return Response{Success: false, Code: "INSUFFICIENT_ROLE", Message: "目前角色无权执行此操作"}
	case errors.Is(err, ErrDeviceNotTrusted):
		return Response{Success: false, Code: "DEVICE_NOT_TRUSTED", Message: "需要由既有信任设备或 system_admin 核准"}
	case errors.Is(err, ErrIPNotAllowed):
		return Response{Success: false, Code: "IP_NOT_ALLOWED", Message: "目前登录 IP 不在白名单中，请联系管理员或确认白名单设置"}
	case errors.Is(err, ErrNewDeviceApprovalRequired):
		return Response{Success: false, Code: "NEW_DEVICE_APPROVAL_REQUIRED", Message: "新设备需要既有信任设备或 system_admin 核准"}
	default:
		return Response{Success: false, Code: "INTERNAL_ERROR", Message: "系统内部错误"}
	}
}

func clientIP(r *http.Request) string {
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}

	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil {
		return host
	}

	return strings.TrimSpace(r.RemoteAddr)
}
