package erp

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"1001-twacc-chat/internal/adminauth"
	"1001-twacc-chat/internal/auth"
)

// Handler exposes external-system auth endpoints plus admin and device routes.
type Handler struct {
	service                       *Service
	sessions                      SessionAuthenticator
	adminSessions                 AdminSessionAuthenticator
	integrationSharedToken        string
	integrationSignatureSecret    string
	integrationTimestampTolerance time.Duration
}

// NewHandler builds the HTTP handler for external-system access flows.
func NewHandler(service *Service, sessions SessionAuthenticator, adminSessions AdminSessionAuthenticator, integrationSharedToken, integrationSignatureSecret string, integrationTimestampTolerance time.Duration) *Handler {
	return &Handler{
		service:                       service,
		sessions:                      sessions,
		adminSessions:                 adminSessions,
		integrationSharedToken:        strings.TrimSpace(integrationSharedToken),
		integrationSignatureSecret:    strings.TrimSpace(integrationSignatureSecret),
		integrationTimestampTolerance: integrationTimestampTolerance,
	}
}

// Register handles POST /api/erp/register.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if resp, status, ok := h.authorizeIntegrationRequest(r); !ok {
		writeJSON(w, status, resp)
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
	if resp, status, ok := h.authorizeIntegrationRequest(r); !ok {
		writeJSON(w, status, resp)
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

// AdminLogin handles POST /admin/api/login.
func (h *Handler) AdminLogin(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeJSON(w, http.StatusServiceUnavailable, Response{Success: false, Code: "SERVICE_UNAVAILABLE", Message: "服务尚未完成初始化"})
		return
	}

	var req AdminLoginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Success: false, Code: "INVALID_REQUEST", Message: "请求格式错误"})
		return
	}

	resp, status, err := h.service.AdminLogin(req, clientIP(r))
	if err != nil {
		writeJSON(w, status, errorResponse(err))
		return
	}

	writeJSON(w, status, resp)
}

// UpdateProfile handles PATCH /api/users/me/profile.
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if h.service == nil || h.sessions == nil {
		writeJSON(w, http.StatusServiceUnavailable, Response{Success: false, Code: "SERVICE_UNAVAILABLE", Message: "服务尚未完成初始化"})
		return
	}

	principal, ok := h.requireSession(w, r)
	if !ok {
		return
	}

	var req ProfileUpdateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Success: false, Code: "INVALID_REQUEST", Message: "请求格式错误"})
		return
	}

	resp, status, err := h.service.UpdateProfile(principal, req)
	if err != nil {
		writeJSON(w, status, errorResponse(err))
		return
	}

	writeJSON(w, status, resp)
}

// ListUsers handles GET /api/system-admin/users.
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	if h.service == nil || h.adminSessions == nil {
		writeJSON(w, http.StatusServiceUnavailable, Response{Success: false, Code: "SERVICE_UNAVAILABLE", Message: "服务尚未完成初始化"})
		return
	}

	principal, ok := h.requireAdminSession(w, r)
	if !ok {
		return
	}

	filter, err := parseAdminUserFilter(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Success: false, Code: "INVALID_REQUEST", Message: err.Error()})
		return
	}

	resp, status, err := h.service.ListUsers(filter, principal)
	if err != nil {
		writeJSON(w, status, errorResponse(err))
		return
	}

	writeJSON(w, status, resp)
}

// ListUserInvitations handles GET /api/system-admin/user-invitations.
func (h *Handler) ListUserInvitations(w http.ResponseWriter, r *http.Request) {
	if h.service == nil || h.adminSessions == nil {
		writeJSON(w, http.StatusServiceUnavailable, Response{Success: false, Code: "SERVICE_UNAVAILABLE", Message: "服务尚未完成初始化"})
		return
	}

	principal, ok := h.requireAdminSession(w, r)
	if !ok {
		return
	}

	filter, err := parseAdminUserInvitationFilter(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Success: false, Code: "INVALID_REQUEST", Message: err.Error()})
		return
	}

	resp, status, err := h.service.ListUserInvitations(filter, principal)
	if err != nil {
		writeJSON(w, status, errorResponse(err))
		return
	}

	writeJSON(w, status, resp)
}

// ListSystemAdmins handles GET /api/system-admin/admins.
func (h *Handler) ListSystemAdmins(w http.ResponseWriter, r *http.Request) {
	if h.service == nil || h.adminSessions == nil {
		writeJSON(w, http.StatusServiceUnavailable, Response{Success: false, Code: "SERVICE_UNAVAILABLE", Message: "服务尚未完成初始化"})
		return
	}

	principal, ok := h.requireAdminSession(w, r)
	if !ok {
		return
	}

	filter, err := parseSystemAdminFilter(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Success: false, Code: "INVALID_REQUEST", Message: err.Error()})
		return
	}

	resp, status, err := h.service.ListSystemAdmins(filter, principal)
	if err != nil {
		writeJSON(w, status, errorResponse(err))
		return
	}

	writeJSON(w, status, resp)
}

// CreateUser handles POST /api/system-admin/users.
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if h.service == nil || h.adminSessions == nil {
		writeJSON(w, http.StatusServiceUnavailable, Response{Success: false, Code: "SERVICE_UNAVAILABLE", Message: "服务尚未完成初始化"})
		return
	}

	principal, ok := h.requireAdminSession(w, r)
	if !ok {
		return
	}

	var req AdminCreateUserRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Success: false, Code: "INVALID_REQUEST", Message: "请求格式错误"})
		return
	}

	resp, status, err := h.service.CreateUser(req, principal)
	if err != nil {
		writeJSON(w, status, errorResponse(err))
		return
	}

	writeJSON(w, status, resp)
}

// InviteUser handles POST /api/system-admin/user-invitations.
func (h *Handler) InviteUser(w http.ResponseWriter, r *http.Request) {
	if h.service == nil || h.adminSessions == nil {
		writeJSON(w, http.StatusServiceUnavailable, Response{Success: false, Code: "SERVICE_UNAVAILABLE", Message: "服务尚未完成初始化"})
		return
	}

	principal, ok := h.requireAdminSession(w, r)
	if !ok {
		return
	}

	var req AdminInviteUserRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Success: false, Code: "INVALID_REQUEST", Message: "请求格式错误"})
		return
	}

	resp, status, err := h.service.InviteUser(req, principal)
	if err != nil {
		writeJSON(w, status, errorResponse(err))
		return
	}

	writeJSON(w, status, resp)
}

// ResendUserInvitation handles POST /api/system-admin/user-invitations/resend.
func (h *Handler) ResendUserInvitation(w http.ResponseWriter, r *http.Request) {
	if h.service == nil || h.adminSessions == nil {
		writeJSON(w, http.StatusServiceUnavailable, Response{Success: false, Code: "SERVICE_UNAVAILABLE", Message: "服务尚未完成初始化"})
		return
	}

	principal, ok := h.requireAdminSession(w, r)
	if !ok {
		return
	}

	var req AdminInviteUserRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Success: false, Code: "INVALID_REQUEST", Message: "请求格式错误"})
		return
	}

	resp, status, err := h.service.ResendUserInvitation(req, principal)
	if err != nil {
		writeJSON(w, status, errorResponse(err))
		return
	}

	writeJSON(w, status, resp)
}

// CreateSystemAdmin handles POST /api/system-admin/admins.
func (h *Handler) CreateSystemAdmin(w http.ResponseWriter, r *http.Request) {
	if h.service == nil || h.adminSessions == nil {
		writeJSON(w, http.StatusServiceUnavailable, Response{Success: false, Code: "SERVICE_UNAVAILABLE", Message: "服务尚未完成初始化"})
		return
	}

	principal, ok := h.requireAdminSession(w, r)
	if !ok {
		return
	}

	var req SystemAdminCreateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Success: false, Code: "INVALID_REQUEST", Message: "请求格式错误"})
		return
	}

	resp, status, err := h.service.CreateSystemAdmin(req, principal)
	if err != nil {
		writeJSON(w, status, errorResponse(err))
		return
	}

	writeJSON(w, status, resp)
}

// GetUser handles GET /api/system-admin/users/{user_id}.
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	if h.service == nil || h.adminSessions == nil {
		writeJSON(w, http.StatusServiceUnavailable, Response{Success: false, Code: "SERVICE_UNAVAILABLE", Message: "服务尚未完成初始化"})
		return
	}

	principal, ok := h.requireAdminSession(w, r)
	if !ok {
		return
	}
	targetUserID, ok := parseUserIDPath(w, r)
	if !ok {
		return
	}

	resp, status, err := h.service.GetUser(targetUserID, principal)
	if err != nil {
		writeJSON(w, status, errorResponse(err))
		return
	}
	writeJSON(w, status, resp)
}

// UpdateUser handles PATCH /api/system-admin/users/{user_id}.
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	if h.service == nil || h.adminSessions == nil {
		writeJSON(w, http.StatusServiceUnavailable, Response{Success: false, Code: "SERVICE_UNAVAILABLE", Message: "服务尚未完成初始化"})
		return
	}

	principal, ok := h.requireAdminSession(w, r)
	if !ok {
		return
	}
	targetUserID, ok := parseUserIDPath(w, r)
	if !ok {
		return
	}

	var req AdminUpdateUserRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Success: false, Code: "INVALID_REQUEST", Message: "请求格式错误"})
		return
	}

	resp, status, err := h.service.UpdateUser(targetUserID, req, principal)
	if err != nil {
		writeJSON(w, status, errorResponse(err))
		return
	}
	writeJSON(w, status, resp)
}

// GetSystemAdmin handles GET /api/system-admin/admins/{user_id}.
func (h *Handler) GetSystemAdmin(w http.ResponseWriter, r *http.Request) {
	if h.service == nil || h.adminSessions == nil {
		writeJSON(w, http.StatusServiceUnavailable, Response{Success: false, Code: "SERVICE_UNAVAILABLE", Message: "服务尚未完成初始化"})
		return
	}

	principal, ok := h.requireAdminSession(w, r)
	if !ok {
		return
	}
	targetAdminID, ok := parseUserIDPath(w, r)
	if !ok {
		return
	}

	resp, status, err := h.service.GetSystemAdmin(targetAdminID, principal)
	if err != nil {
		writeJSON(w, status, errorResponse(err))
		return
	}
	writeJSON(w, status, resp)
}

// UpdateSystemAdmin handles PATCH /api/system-admin/admins/{user_id}.
func (h *Handler) UpdateSystemAdmin(w http.ResponseWriter, r *http.Request) {
	if h.service == nil || h.adminSessions == nil {
		writeJSON(w, http.StatusServiceUnavailable, Response{Success: false, Code: "SERVICE_UNAVAILABLE", Message: "服务尚未完成初始化"})
		return
	}

	principal, ok := h.requireAdminSession(w, r)
	if !ok {
		return
	}
	targetAdminID, ok := parseUserIDPath(w, r)
	if !ok {
		return
	}

	var req SystemAdminCreateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Success: false, Code: "INVALID_REQUEST", Message: "请求格式错误"})
		return
	}

	resp, status, err := h.service.UpdateSystemAdmin(targetAdminID, req, principal)
	if err != nil {
		writeJSON(w, status, errorResponse(err))
		return
	}
	writeJSON(w, status, resp)
}

// ListDevices handles GET /api/system-admin/users/{user_id}/devices.
func (h *Handler) ListDevices(w http.ResponseWriter, r *http.Request) {
	if h.service == nil || h.adminSessions == nil {
		writeJSON(w, http.StatusServiceUnavailable, Response{Success: false, Code: "SERVICE_UNAVAILABLE", Message: "服务尚未完成初始化"})
		return
	}

	principal, ok := h.requireAdminSession(w, r)
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
	if h.service == nil || h.adminSessions == nil {
		writeJSON(w, http.StatusServiceUnavailable, Response{Success: false, Code: "SERVICE_UNAVAILABLE", Message: "服务尚未完成初始化"})
		return
	}

	principal, ok := h.requireAdminSession(w, r)
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

func (h *Handler) authorizeIntegrationRequest(r *http.Request) (Response, int, bool) {
	if h.integrationSharedToken == "" && h.integrationSignatureSecret == "" {
		return Response{}, 0, true
	}

	if !h.integrationTokenAuthorized(r) {
		return Response{Success: false, Code: "INVALID_INTEGRATION_TOKEN", Message: "对接凭证错误"}, http.StatusUnauthorized, false
	}

	if h.integrationSignatureSecret == "" {
		return Response{}, 0, true
	}

	timestampHeader := strings.TrimSpace(r.Header.Get("X-TWACC-Timestamp"))
	if timestampHeader == "" {
		return Response{Success: false, Code: "INTEGRATION_TIMESTAMP_REQUIRED", Message: "缺少对接请求时间戳"}, http.StatusUnauthorized, false
	}

	signatureHeader := strings.TrimSpace(r.Header.Get("X-TWACC-Signature"))
	if signatureHeader == "" {
		return Response{Success: false, Code: "INTEGRATION_SIGNATURE_REQUIRED", Message: "缺少对接请求签章"}, http.StatusUnauthorized, false
	}

	requestTimestamp, err := strconv.ParseInt(timestampHeader, 10, 64)
	if err != nil {
		return Response{Success: false, Code: "INVALID_INTEGRATION_TIMESTAMP", Message: "对接请求时间戳格式错误"}, http.StatusUnauthorized, false
	}

	now := time.Now().UTC()
	requestTime := time.Unix(requestTimestamp, 0).UTC()
	tolerance := h.integrationTimestampTolerance
	if tolerance <= 0 {
		tolerance = 5 * time.Minute
	}
	if requestTime.Before(now.Add(-tolerance)) || requestTime.After(now.Add(tolerance)) {
		return Response{Success: false, Code: "INTEGRATION_TIMESTAMP_EXPIRED", Message: "对接请求时间戳已过期"}, http.StatusUnauthorized, false
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return Response{Success: false, Code: "INVALID_REQUEST", Message: "请求读取失败"}, http.StatusBadRequest, false
	}
	_ = r.Body.Close()
	r.Body = io.NopCloser(bytes.NewReader(body))

	if !h.integrationSignatureValid(r.Method, r.URL.Path, timestampHeader, body, signatureHeader) {
		return Response{Success: false, Code: "INVALID_INTEGRATION_SIGNATURE", Message: "对接请求签章错误"}, http.StatusUnauthorized, false
	}

	return Response{}, 0, true
}

func (h *Handler) integrationTokenAuthorized(r *http.Request) bool {
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

func (h *Handler) integrationSignatureValid(method, path, timestamp string, body []byte, provided string) bool {
	if h.integrationSignatureSecret == "" {
		return true
	}

	mac := hmac.New(sha256.New, []byte(h.integrationSignatureSecret))
	mac.Write([]byte(method))
	mac.Write([]byte("\n"))
	mac.Write([]byte(path))
	mac.Write([]byte("\n"))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("\n"))
	mac.Write(body)

	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(strings.ToLower(strings.TrimSpace(provided))), []byte(expected))
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

func (h *Handler) requireAdminSession(w http.ResponseWriter, r *http.Request) (AdminSessionPrincipal, bool) {
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		writeJSON(w, http.StatusUnauthorized, Response{Success: false, Code: "UNAUTHORIZED", Message: "缺少有效 admin session token"})
		return AdminSessionPrincipal{}, false
	}

	token := strings.TrimSpace(strings.TrimPrefix(authHeader, prefix))
	session, err := h.adminSessions.Authenticate(token)
	if err != nil {
		if errors.Is(err, adminauth.ErrSessionExpired) {
			writeJSON(w, http.StatusUnauthorized, Response{Success: false, Code: "SESSION_EXPIRED", Message: "admin session 已过期，请重新登录"})
			return AdminSessionPrincipal{}, false
		}
		writeJSON(w, http.StatusUnauthorized, Response{Success: false, Code: "UNAUTHORIZED", Message: "admin session token 无效"})
		return AdminSessionPrincipal{}, false
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

func parseAdminUserFilter(r *http.Request) (AdminUserFilter, error) {
	query := r.URL.Query()
	filter := AdminUserFilter{
		ExternalUserID: strings.TrimSpace(query.Get("external_user_id")),
		DisplayName:    strings.TrimSpace(query.Get("display_name")),
		Email:          strings.TrimSpace(query.Get("email")),
		Status:         strings.TrimSpace(query.Get("status")),
		SourceSystem:   strings.TrimSpace(query.Get("source_system")),
		Sort:           strings.TrimSpace(query.Get("sort")),
		Page:           1,
		PerPage:        10,
	}

	var err error
	if value := strings.TrimSpace(query.Get("page")); value != "" {
		filter.Page, err = strconv.Atoi(value)
		if err != nil || filter.Page < 1 {
			return AdminUserFilter{}, fmt.Errorf("頁碼格式錯誤")
		}
	}
	if value := strings.TrimSpace(query.Get("per_page")); value != "" {
		filter.PerPage, err = strconv.Atoi(value)
		if err != nil || !validAdminPerPage(filter.PerPage) {
			return AdminUserFilter{}, fmt.Errorf("每頁筆數格式錯誤")
		}
	}
	if value := strings.TrimSpace(query.Get("created_from")); value != "" {
		filter.CreatedFrom, err = parseAdminDate(value, false)
		if err != nil {
			return AdminUserFilter{}, fmt.Errorf("開始時間格式錯誤")
		}
	}
	if value := strings.TrimSpace(query.Get("created_to")); value != "" {
		filter.CreatedTo, err = parseAdminDate(value, true)
		if err != nil {
			return AdminUserFilter{}, fmt.Errorf("結束時間格式錯誤")
		}
	}

	return filter, nil
}

func parseAdminUserInvitationFilter(r *http.Request) (AdminUserInvitationFilter, error) {
	query := r.URL.Query()
	filter := AdminUserInvitationFilter{
		Email:   strings.TrimSpace(query.Get("email")),
		Status:  strings.TrimSpace(query.Get("status")),
		Sorting: strings.TrimSpace(query.Get("sorting")),
		Page:    1,
		PerPage: 10,
	}

	var err error
	if value := strings.TrimSpace(query.Get("page")); value != "" {
		filter.Page, err = strconv.Atoi(value)
		if err != nil || filter.Page < 1 {
			return AdminUserInvitationFilter{}, fmt.Errorf("頁碼格式錯誤")
		}
	}
	if value := strings.TrimSpace(query.Get("per_page")); value != "" {
		filter.PerPage, err = strconv.Atoi(value)
		if err != nil || !validAdminPerPage(filter.PerPage) {
			return AdminUserInvitationFilter{}, fmt.Errorf("每頁筆數格式錯誤")
		}
	}

	return filter, nil
}

func parseSystemAdminFilter(r *http.Request) (SystemAdminFilter, error) {
	query := r.URL.Query()
	filter := SystemAdminFilter{
		ExternalUserID: strings.TrimSpace(query.Get("external_user_id")),
		DisplayName:    strings.TrimSpace(query.Get("display_name")),
		Status:         strings.TrimSpace(query.Get("status")),
		Page:           1,
		PerPage:        10,
	}

	var err error
	if value := strings.TrimSpace(query.Get("page")); value != "" {
		filter.Page, err = strconv.Atoi(value)
		if err != nil || filter.Page < 1 {
			return SystemAdminFilter{}, fmt.Errorf("頁碼格式錯誤")
		}
	}
	if value := strings.TrimSpace(query.Get("per_page")); value != "" {
		filter.PerPage, err = strconv.Atoi(value)
		if err != nil || !validAdminPerPage(filter.PerPage) {
			return SystemAdminFilter{}, fmt.Errorf("每頁筆數格式錯誤")
		}
	}

	return filter, nil
}

func validAdminPerPage(value int) bool {
	return value == 10 || value == 20 || value == 50 || value == 100
}

func parseAdminDate(value string, endOfDay bool) (*time.Time, error) {
	parsed, err := time.ParseInLocation("2006-01-02", value, time.Local)
	if err != nil {
		return nil, err
	}
	if endOfDay {
		parsed = parsed.Add(24*time.Hour - time.Nanosecond)
	}
	return &parsed, nil
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
	case errors.Is(err, ErrInvalidEmail):
		return Response{Success: false, Code: "INVALID_EMAIL", Message: "Email 格式错误"}
	case errors.Is(err, ErrInvalidStatus):
		return Response{Success: false, Code: "INVALID_STATUS", Message: "状态仅支持 active 或 inactive"}
	case errors.Is(err, ErrInvalidPassword):
		return Response{Success: false, Code: "INVALID_PASSWORD", Message: "密码需为 4 到 20 码，且只能包含英文、数字或特殊符号"}
	case errors.Is(err, ErrInvalidCredentials):
		return Response{Success: false, Code: "INVALID_CREDENTIALS", Message: "帐号或密码错误"}
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
	case errors.Is(err, ErrEmailAlreadyExists):
		return Response{Success: false, Code: "EMAIL_ALREADY_EXISTS", Message: "Email 已存在"}
	case errors.Is(err, ErrUserInvitationPending):
		return Response{Success: false, Code: "USER_INVITATION_PENDING", Message: "該 Email 已有待註冊邀請"}
	case errors.Is(err, ErrUserInvitationCompleted):
		return Response{Success: false, Code: "USER_INVITATION_COMPLETED", Message: "該註冊邀請已完成"}
	case errors.Is(err, ErrInvitationMailerUnavailable):
		return Response{Success: false, Code: "MAILER_UNAVAILABLE", Message: "寄信服務尚未設定"}
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
