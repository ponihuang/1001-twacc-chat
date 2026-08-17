package erp

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var sourceSystemPattern = regexp.MustCompile(`^[a-z0-9_-]{2,50}$`)
var passwordPattern = regexp.MustCompile(`^[A-Za-z0-9[:punct:]]{4,20}$`)
var adminRoleCodePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{2,50}$`)

const passwordHashIterations = 120000
const bulkInvitationTXTMaxBytes = 2 << 20
const bulkInvitationTXTMaxEntries = 1000

var adminPermissionCatalog = []AdminPermissionGroup{
	{
		Key:  "users",
		Name: "用戶",
		Items: []AdminPermissionItem{
			{Key: "office.user.index", Name: "用戶管理", Category: "users"},
			{Key: "office.user.create", Name: "用戶新增", Category: "users"},
			{Key: "office.user.edit", Name: "用戶編輯", Category: "users"},
			{Key: "office.user.chat-mute", Name: "禁止/解除發言", Category: "users"},
			{Key: "office.user.temporary-password", Name: "臨時密碼", Category: "users"},
			{Key: "office.user.invitation.index", Name: "註冊邀請", Category: "users"},
			{Key: "office.user.invitation.create", Name: "邀請新增", Category: "users"},
			{Key: "office.user.invitation.bulk", Name: "批量新增", Category: "users"},
			{Key: "office.user.invitation.resend", Name: "重送邀請", Category: "users"},
		},
	},
	{
		Key:  "chat",
		Name: "聊天",
		Items: []AdminPermissionItem{
			{Key: "office.conversations.index", Name: "聊天紀錄", Category: "chat"},
			{Key: "office.conversations.detail", Name: "查看紀錄", Category: "chat"},
		},
	},
	{
		Key:  "system",
		Name: "系統",
		Items: []AdminPermissionItem{
			{Key: "office.admin.index", Name: "管理員帳號", Category: "system"},
			{Key: "office.admin.create", Name: "管理員新增", Category: "system"},
			{Key: "office.admin.edit", Name: "管理員編輯", Category: "system"},
			{Key: "office.permission.index", Name: "權限管理", Category: "system"},
			{Key: "office.permission.create", Name: "權限新增", Category: "system"},
			{Key: "office.permission.edit", Name: "權限編輯", Category: "system"},
			{Key: "office.permission.detail", Name: "權限詳細", Category: "system"},
			{Key: "office.system-settings.index", Name: "系統設定", Category: "system"},
			{Key: "office.system-settings.edit", Name: "系統設定編輯", Category: "system"},
		},
	},
}

var (
	ErrInvalidSourceSystem         = errors.New("invalid source system")
	ErrInvalidExternalUserID       = errors.New("invalid external user id")
	ErrInvalidDisplayName          = errors.New("invalid display name")
	ErrInvalidEmail                = errors.New("invalid email")
	ErrEmailAlreadyExists          = errors.New("email already exists")
	ErrBulkInvitationTooMany       = errors.New("bulk invitation too many")
	ErrUserInvitationPending       = errors.New("user invitation pending")
	ErrUserInvitationCompleted     = errors.New("user invitation completed")
	ErrUserInvitationNotFound      = errors.New("user invitation not found")
	ErrInvitationMailerUnavailable = errors.New("invitation mailer unavailable")
	ErrTemporaryPasswordMailFailed = errors.New("temporary password mail failed")
	ErrPasswordConfirmation        = errors.New("password confirmation mismatch")
	ErrInvalidStatus               = errors.New("invalid status")
	ErrInvalidRole                 = errors.New("invalid role")
	ErrRoleAlreadyExists           = errors.New("role already exists")
	ErrInvalidPassword             = errors.New("invalid password")
	ErrInvalidCredentials          = errors.New("invalid credentials")
	ErrInvalidUserID               = errors.New("invalid user id")
	ErrDeviceIDRequired            = errors.New("device id required")
	ErrInvalidIPWhitelistRule      = errors.New("invalid ip whitelist rule")
	ErrWhitelistRulesRequired      = errors.New("whitelist rules required")
	ErrUserNotFound                = errors.New("user not found")
	ErrDeviceNotFound              = errors.New("device not found")
	ErrConversationNotFound        = errors.New("conversation not found")
	ErrUserAlreadyExists           = errors.New("user already exists")
	ErrSystemAdminCannotChat       = errors.New("system admin cannot chat")
	ErrInsufficientRole            = errors.New("insufficient role")
	ErrDeviceNotTrusted            = errors.New("device not trusted")
	ErrIPNotAllowed                = errors.New("ip not allowed")
	ErrNewDeviceApprovalRequired   = errors.New("new device approval required")
	ErrTemporaryPasswordExpired    = errors.New("temporary password expired")
)

// Service implements external-system identity and login rules for chat access.
type Service struct {
	repo                 Repository
	sessions             SessionIssuer
	adminSessions        AdminSessionIssuer
	requireTrustedDevice bool
	invitationTTL        time.Duration
	invitationMailer     InvitationMailer
	invitationBaseURL    string
}

// NewService builds an integration-backed identity service.
func NewService(repo Repository, sessions SessionIssuer) *Service {
	return &Service{repo: repo, sessions: sessions, invitationTTL: 7 * 24 * time.Hour}
}

// SetAdminSessions sets the session issuer used by backend admin login.
func (s *Service) SetAdminSessions(adminSessions AdminSessionIssuer) {
	if s == nil {
		return
	}
	s.adminSessions = adminSessions
}

// SetRequireTrustedDevice controls whether non-trusted devices must be approved before login.
func (s *Service) SetRequireTrustedDevice(require bool) {
	if s == nil {
		return
	}
	s.requireTrustedDevice = require
}

// SetInvitationTTL overrides the user invitation lifetime.
func (s *Service) SetInvitationTTL(ttl time.Duration) {
	if s == nil || ttl <= 0 {
		return
	}
	s.invitationTTL = ttl
}

// SetInvitationMailer sets the sender used for user invitations.
func (s *Service) SetInvitationMailer(mailer InvitationMailer) {
	if s == nil {
		return
	}
	s.invitationMailer = mailer
}

// SetInvitationBaseURL sets the public URL prefix for invitation links.
func (s *Service) SetInvitationBaseURL(baseURL string) {
	if s == nil {
		return
	}
	s.invitationBaseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
}

// Register creates a user if it does not exist yet.
func (s *Service) Register(req RegisterRequest) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}

	params, err := validateRegisterRequest(req)
	if err != nil {
		return Response{}, statusCode(err), err
	}
	if _, err := s.createUser(params, req.Password); err != nil {
		return Response{}, statusCode(err), err
	}

	return Response{Success: true, Code: "REGISTERED", Message: "用户注册成功"}, 200, nil
}

// Login validates whitelist and device state, then issues a persisted session token.
func (s *Service) Login(req LoginRequest, clientIP, userAgent string) (Response, int, error) {
	if s == nil || s.repo == nil || s.sessions == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}

	sourceSystem, externalUserID, password, deviceID, err := validateLoginRequest(req)
	if err != nil {
		return Response{}, statusCode(err), err
	}

	user, err := s.repo.FindUserByExternal(sourceSystem, externalUserID)
	if err != nil {
		return Response{}, statusCode(err), err
	}
	if !passwordMatches(password, user.PasswordHash) {
		return Response{}, statusCode(ErrInvalidCredentials), ErrInvalidCredentials
	}
	if user.TemporaryPasswordExpiresAt != nil && user.MustChangePassword && time.Now().UTC().After(*user.TemporaryPasswordExpiresAt) {
		return Response{}, statusCode(ErrTemporaryPasswordExpired), ErrTemporaryPasswordExpired
	}

	settings, err := s.repo.GetUserSecuritySettings(user.ID)
	if err != nil {
		return Response{}, 500, err
	}
	if !settings.AllowAllIPs {
		rules, err := s.repo.ListActiveIPWhitelistRules(user.ID)
		if err != nil {
			return Response{}, 500, err
		}
		if !ipAllowed(clientIP, rules) {
			return Response{}, statusCode(ErrIPNotAllowed), ErrIPNotAllowed
		}
	}

	trustedDevices, err := s.repo.CountTrustedDevices(user.ID)
	if err != nil {
		return Response{}, 500, err
	}

	device, err := s.repo.FindDevice(user.ID, deviceID)
	if err != nil && !errors.Is(err, ErrDeviceNotFound) {
		return Response{}, 500, err
	}

	now := time.Now().UTC()
	trusted := trustedDevices == 0 || device.IsTrustedDevice
	if !s.requireTrustedDevice {
		trusted = true
	}
	if err := s.repo.UpsertDeviceLogin(user.ID, deviceID, userAgent, clientIP, trusted, now); err != nil {
		return Response{}, 500, err
	}
	if !trusted {
		return Response{}, statusCode(ErrNewDeviceApprovalRequired), ErrNewDeviceApprovalRequired
	}

	token, expiresAt, err := s.sessions.Issue(user.ID, deviceID)
	if err != nil {
		return Response{}, 500, err
	}

	response := Response{
		Success:   true,
		Code:      "LOGIN_OK",
		Message:   "登录成功",
		Token:     token,
		ExpiresAt: expiresAt.Format(time.RFC3339),
	}

	if data, err := s.profileResponseData(user); err == nil {
		response.Data = data
	}

	return response, 200, nil
}

// AdminLogin validates a backend admin account and issues an admin session token.
func (s *Service) AdminLogin(req AdminLoginRequest, clientIP string) (Response, int, error) {
	if s == nil || s.repo == nil || s.adminSessions == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}

	account, password, deviceID, err := validateAdminLoginRequest(req)
	if err != nil {
		return Response{}, statusCode(err), err
	}

	admin, err := s.repo.FindSystemAdminByAccount(account)
	if err != nil {
		return Response{}, statusCode(err), err
	}
	if admin.Status != "active" || !passwordMatches(password, adminPasswordHash(admin)) {
		return Response{}, statusCode(ErrInvalidCredentials), ErrInvalidCredentials
	}

	token, expiresAt, err := s.adminSessions.Issue(admin.UserID, deviceID)
	if err != nil {
		return Response{}, 500, err
	}
	if err := s.repo.UpdateSystemAdminLastLogin(admin.UserID, clientIP, time.Now().UTC()); err != nil {
		return Response{}, 500, err
	}

	return Response{
		Success:   true,
		Code:      "ADMIN_LOGIN_OK",
		Message:   "登入成功",
		Token:     token,
		ExpiresAt: expiresAt.Format(time.RFC3339),
		Data: map[string]any{
			"role":             admin.Role,
			"is_system_admin":  true,
			"user_id":          admin.UserID,
			"admin_user_id":    admin.UserID,
			"external_user_id": admin.ExternalUserID,
			"display_name":     admin.DisplayName,
		},
	}, 200, nil
}

// UpdateProfile updates the authenticated user's profile fields.
func (s *Service) UpdateProfile(actor SessionPrincipal, req ProfileUpdateRequest) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}
	if actor.UserID <= 0 {
		return Response{}, statusCode(ErrInsufficientRole), ErrInsufficientRole
	}

	displayName := strings.TrimSpace(req.DisplayName)
	if displayName == "" || len([]rune(displayName)) > 100 {
		return Response{}, statusCode(ErrInvalidDisplayName), ErrInvalidDisplayName
	}

	email := normalizeEmail(req.Email)
	if email != "" && !isValidEmail(email) {
		return Response{}, statusCode(ErrInvalidEmail), ErrInvalidEmail
	}
	if email != "" {
		existing, err := s.repo.FindUserByEmail(email)
		if err == nil && existing.ID != actor.UserID {
			return Response{}, statusCode(ErrEmailAlreadyExists), ErrEmailAlreadyExists
		}
		if err != nil && !errors.Is(err, ErrUserNotFound) {
			return Response{}, statusCode(err), err
		}
	}

	user, err := s.repo.UpdateUserProfile(actor.UserID, displayName, email)
	if err != nil {
		return Response{}, statusCode(err), err
	}

	data, err := s.profileResponseData(user)
	if err != nil {
		return Response{}, 500, err
	}

	return Response{Success: true, Code: "PROFILE_UPDATED", Message: "个人资料更新成功", Data: data}, 200, nil
}

// GetProfile returns the authenticated user's current profile data.
func (s *Service) GetProfile(actor SessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}
	if actor.UserID <= 0 {
		return Response{}, statusCode(ErrInsufficientRole), ErrInsufficientRole
	}

	user, err := s.repo.FindUserByID(actor.UserID)
	if err != nil {
		return Response{}, statusCode(err), err
	}
	data, err := s.profileResponseData(user)
	if err != nil {
		return Response{}, 500, err
	}

	return Response{Success: true, Code: "PROFILE_LOADED", Message: "个人资料读取成功", Data: data}, 200, nil
}

// UpdatePassword changes the authenticated user's password and clears temporary password flags.
func (s *Service) UpdatePassword(actor SessionPrincipal, req PasswordUpdateRequest) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}
	if actor.UserID <= 0 {
		return Response{}, statusCode(ErrInsufficientRole), ErrInsufficientRole
	}

	user, err := s.repo.FindUserByID(actor.UserID)
	if err != nil {
		return Response{}, statusCode(err), err
	}
	if !passwordMatches(req.CurrentPassword, user.PasswordHash) {
		return Response{}, statusCode(ErrInvalidCredentials), ErrInvalidCredentials
	}
	if strings.TrimSpace(req.Password) != strings.TrimSpace(req.PasswordConfirmation) {
		return Response{}, statusCode(ErrPasswordConfirmation), ErrPasswordConfirmation
	}
	passwordHash, err := hashPassword(req.Password)
	if err != nil {
		return Response{}, statusCode(err), err
	}

	updated, err := s.repo.UpdateUserPassword(actor.UserID, passwordHash)
	if err != nil {
		return Response{}, statusCode(err), err
	}
	data, err := s.profileResponseData(updated)
	if err != nil {
		return Response{}, 500, err
	}

	return Response{Success: true, Code: "PASSWORD_UPDATED", Message: "密碼已更新", Data: data}, 200, nil
}
func (s *Service) adminPrincipalHasPermission(actor AdminSessionPrincipal, key string) bool {
	if s == nil || s.repo == nil || actor.AdminUserID <= 0 {
		return false
	}

	admin, err := s.repo.FindSystemAdminByID(actor.AdminUserID)
	if err != nil {
		return false
	}

	role, err := s.repo.FindAdminRoleByCode(admin.Role)
	if err != nil {
		return false
	}
	if role.Code == "system_admin" {
		return true
	}

	permissions, err := s.repo.ListAdminRolePermissions(role.ID)
	if err != nil {
		return false
	}

	if permissions[key] {
		return true
	}

	// 有任何 user 分類權限就允許列表顯示
	if strings.HasPrefix(key, "office.user.") {
		for perm, enabled := range permissions {
			if enabled && strings.HasPrefix(perm, "office.user.") {
				return true
			}
		}
	}

	return false
}

// ListUsers returns users for the admin console. Only system_admin is allowed.
func (s *Service) ListUsers(filter AdminUserFilter, actor AdminSessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}

	if !(s.adminPrincipalHasPermission(actor, "office.user.index") ||
		s.adminPrincipalHasPermission(actor, "office.user.*") ||
		s.adminPrincipalHasPermission(actor, "office.*")) {
		if err := s.requireSystemAdmin(actor.AdminUserID); err != nil {
			return Response{}, statusCode(err), err
		}
	}

	users, err := s.repo.ListUsers(filter)
	if err != nil {
		return Response{}, 500, err
	}

	return Response{
		Success: true,
		Code:    "USERS_OK",
		Message: "用户列表读取成功",
		Data:    users,
	}, 200, nil
}

// ListUserInvitations returns registration invitations for the admin console.
func (s *Service) ListUserInvitations(filter AdminUserInvitationFilter, actor AdminSessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}
	if err := s.requireSystemAdmin(actor.AdminUserID); err != nil {
		return Response{}, statusCode(err), err
	}

	invitations, err := s.repo.ListUserInvitations(filter)
	if err != nil {
		return Response{}, 500, err
	}

	return Response{
		Success: true,
		Code:    "USER_INVITATIONS_OK",
		Message: "註冊邀請列表讀取成功",
		Data:    invitations,
	}, 200, nil
}

// PrecheckUserInvitationTXT classifies an uploaded TXT before creating any invitation.
func (s *Service) PrecheckUserInvitationTXT(content []byte, actor AdminSessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}
	if err := s.requireSystemAdmin(actor.AdminUserID); err != nil {
		return Response{}, statusCode(err), err
	}

	result, err := s.classifyUserInvitationTXT(content)
	if err != nil {
		return Response{}, statusCode(err), err
	}

	return Response{
		Success: true,
		Code:    "USER_INVITATION_BULK_PRECHECK_OK",
		Message: "註冊邀請 TXT 預檢完成",
		Data:    result,
	}, 200, nil
}

// SendBulkUserInvitations creates and sends invitations for currently sendable TXT emails.
func (s *Service) SendBulkUserInvitations(content []byte, actor AdminSessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}
	if err := s.requireSystemAdmin(actor.AdminUserID); err != nil {
		return Response{}, statusCode(err), err
	}

	precheck, err := s.classifyUserInvitationTXT(content)
	if err != nil {
		return Response{}, statusCode(err), err
	}

	result := BulkUserInvitationSendResult{
		Succeeded: []BulkUserInvitationPrecheckItem{},
		Failed:    []BulkUserInvitationSendFailure{},
		Ignored: BulkUserInvitationIgnoredResult{
			DuplicateInFile:   precheck.DuplicateInFile,
			AlreadyRegistered: precheck.AlreadyRegistered,
			ActiveInvitation:  precheck.ActiveInvitation,
			InvalidEmail:      precheck.InvalidEmail,
		},
		Precheck: precheck,
	}

	for _, item := range precheck.Sendable {
		_, _, err := s.InviteUser(AdminInviteUserRequest{Email: item.Email}, actor)
		if err != nil {
			if errors.Is(err, ErrEmailAlreadyExists) {
				result.Ignored.AlreadyRegistered = append(result.Ignored.AlreadyRegistered, item)
				continue
			}
			if errors.Is(err, ErrUserInvitationPending) {
				result.Ignored.ActiveInvitation = append(result.Ignored.ActiveInvitation, item)
				continue
			}
			errResp := errorResponse(err)
			result.Failed = append(result.Failed, BulkUserInvitationSendFailure{
				Line:    item.Line,
				Email:   item.Email,
				Code:    errResp.Code,
				Message: errResp.Message,
			})
			continue
		}
		result.Succeeded = append(result.Succeeded, item)
	}
	result.SuccessCount = len(result.Succeeded)
	result.FailureCount = len(result.Failed)
	result.IgnoredCount = len(result.Ignored.DuplicateInFile) +
		len(result.Ignored.AlreadyRegistered) +
		len(result.Ignored.ActiveInvitation) +
		len(result.Ignored.InvalidEmail)

	return Response{
		Success: true,
		Code:    "USER_INVITATION_BULK_SEND_DONE",
		Message: "批量註冊邀請發送完成",
		Data:    result,
	}, 200, nil
}

func (s *Service) classifyUserInvitationTXT(content []byte) (BulkUserInvitationPrecheckResult, error) {
	result := BulkUserInvitationPrecheckResult{
		Sendable:          []BulkUserInvitationPrecheckItem{},
		DuplicateInFile:   []BulkUserInvitationPrecheckItem{},
		AlreadyRegistered: []BulkUserInvitationPrecheckItem{},
		ActiveInvitation:  []BulkUserInvitationPrecheckItem{},
		InvalidEmail:      []BulkUserInvitationInvalidItem{},
	}
	seen := map[string]bool{}
	bomRemoved := false
	unique := make([]BulkUserInvitationPrecheckItem, 0)
	for i, raw := range strings.Split(string(content), "\n") {
		line := i + 1
		value := strings.TrimSpace(raw)
		if !bomRemoved && value != "" {
			value = strings.TrimPrefix(value, "\uFEFF")
			value = strings.TrimSpace(value)
			bomRemoved = true
		}
		value = strings.TrimSuffix(value, "\r")
		if value == "" {
			result.IgnoredBlankLines++
			continue
		}
		result.TotalLines++
		email := normalizeEmail(value)
		if !isValidEmail(email) {
			result.InvalidEmail = append(result.InvalidEmail, BulkUserInvitationInvalidItem{Line: line, Value: value})
			continue
		}
		item := BulkUserInvitationPrecheckItem{Line: line, Email: email}
		if seen[email] {
			result.DuplicateInFile = append(result.DuplicateInFile, item)
			continue
		}
		seen[email] = true
		unique = append(unique, item)
		if len(unique) > bulkInvitationTXTMaxEntries {
			return BulkUserInvitationPrecheckResult{}, ErrBulkInvitationTooMany
		}
	}
	result.UniqueEmails = len(unique)

	emails := make([]string, 0, len(unique))
	for _, item := range unique {
		emails = append(emails, item.Email)
	}
	precheck, err := s.repo.PrecheckUserInvitationEmails(emails, time.Now().UTC())
	if err != nil {
		return BulkUserInvitationPrecheckResult{}, err
	}

	for _, item := range unique {
		switch {
		case precheck.Registered[item.Email]:
			result.AlreadyRegistered = append(result.AlreadyRegistered, item)
		case precheck.ActiveInvitations[item.Email]:
			result.ActiveInvitation = append(result.ActiveInvitation, item)
		default:
			result.Sendable = append(result.Sendable, item)
		}
	}
	result.Counts = BulkUserInvitationPrecheckCounts{
		Sendable:          len(result.Sendable),
		DuplicateInFile:   len(result.DuplicateInFile),
		AlreadyRegistered: len(result.AlreadyRegistered),
		ActiveInvitation:  len(result.ActiveInvitation),
		InvalidEmail:      len(result.InvalidEmail),
	}

	return result, nil
}

// ListSystemAdmins returns system-admin accounts for the admin console.
func (s *Service) ListSystemAdmins(filter SystemAdminFilter, actor AdminSessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}
	if err := s.requireSystemAdmin(actor.AdminUserID); err != nil {
		return Response{}, statusCode(err), err
	}

	admins, err := s.repo.ListSystemAdmins(filter)
	if err != nil {
		return Response{}, 500, err
	}

	return Response{
		Success: true,
		Code:    "SYSTEM_ADMINS_OK",
		Message: "管理員列表讀取成功",
		Data:    admins,
	}, 200, nil
}

// ListAdminRoles returns backend admin roles for the admin console.
func (s *Service) ListAdminRoles(filter AdminRoleFilter, actor AdminSessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}
	if err := s.requireSystemAdmin(actor.AdminUserID); err != nil {
		return Response{}, statusCode(err), err
	}

	roles, err := s.repo.ListAdminRoles(filter)
	if err != nil {
		return Response{}, 500, err
	}

	return Response{
		Success: true,
		Code:    "ADMIN_ROLES_OK",
		Message: "角色列表讀取成功",
		Data:    roles,
	}, 200, nil
}

// ListAdminConversations returns read-only chat conversations for the admin console.
func (s *Service) ListAdminConversations(filter AdminConversationFilter, actor AdminSessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}
	if err := s.requireSystemAdmin(actor.AdminUserID); err != nil {
		return Response{}, statusCode(err), err
	}
	if err := s.applyAdminChatHistoryRetention(&filter); err != nil {
		return Response{}, 500, err
	}

	conversations, err := s.repo.ListAdminConversations(filter)
	if err != nil {
		return Response{}, 500, err
	}

	return Response{
		Success: true,
		Code:    "ADMIN_CONVERSATIONS_OK",
		Message: "聊天紀錄列表讀取成功",
		Data:    conversations,
	}, 200, nil
}

// GetAdminConversationDetail returns one read-only conversation record page for the admin console.
func (s *Service) GetAdminConversationDetail(filter AdminConversationMessageFilter, actor AdminSessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}
	if err := s.requireSystemAdmin(actor.AdminUserID); err != nil {
		return Response{}, statusCode(err), err
	}
	if err := s.applyAdminChatMessageRetention(&filter); err != nil {
		return Response{}, 500, err
	}

	detail, err := s.repo.GetAdminConversationDetail(filter)
	if err != nil {
		if errors.Is(err, ErrConversationNotFound) {
			return Response{}, 404, err
		}
		return Response{}, 500, err
	}

	return Response{
		Success: true,
		Code:    "ADMIN_CONVERSATION_DETAIL_OK",
		Message: "聊天紀錄明細讀取成功",
		Data:    detail,
	}, 200, nil
}

// CreateUser creates an active user from the admin console. Only system_admin is allowed.
func (s *Service) CreateUser(req AdminCreateUserRequest, actor AdminSessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}
	if err := s.requireSystemAdmin(actor.AdminUserID); err != nil {
		return Response{}, statusCode(err), err
	}

	params, err := validateRegisterRequest(RegisterRequest{
		SourceSystem:   req.SourceSystem,
		ExternalUserID: req.ExternalUserID,
		Password:       req.Password,
		DisplayName:    req.DisplayName,
		Email:          req.Email,
		Language:       req.Language,
	})
	if err != nil {
		return Response{}, statusCode(err), err
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "active"
	}
	if status != "active" && status != "inactive" {
		return Response{}, statusCode(ErrInvalidStatus), ErrInvalidStatus
	}
	params.Status = status

	user, err := s.createUser(params, req.Password)
	if err != nil {
		return Response{}, statusCode(err), err
	}

	return Response{
		Success: true,
		Code:    "USER_CREATED",
		Message: "用户建立成功",
		Data: AdminUserSummary{
			ID:             user.ID,
			ExternalUserID: user.ExternalUserID,
			DisplayName:    user.DisplayName,
			Email:          user.Email,
			Status:         user.Status,
			SourceSystem:   user.SourceSystem,
			CreatedAt:      user.CreatedAt,
		},
	}, 201, nil
}

// InviteUser creates a pending user invitation. The invited user completes account setup separately.
func (s *Service) InviteUser(req AdminInviteUserRequest, actor AdminSessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}
	if err := s.requireSystemAdmin(actor.AdminUserID); err != nil {
		return Response{}, statusCode(err), err
	}

	email := normalizeEmail(req.Email)
	if !isValidEmail(email) {
		return Response{}, statusCode(ErrInvalidEmail), ErrInvalidEmail
	}

	if _, err := s.repo.FindUserByEmail(email); err == nil {
		return Response{}, statusCode(ErrEmailAlreadyExists), ErrEmailAlreadyExists
	} else if !errors.Is(err, ErrUserNotFound) {
		return Response{}, statusCode(err), err
	}

	if s.invitationMailer == nil {
		return Response{}, statusCode(ErrInvitationMailerUnavailable), ErrInvitationMailerUnavailable
	}

	now := time.Now().UTC()
	token, tokenHash, err := newInvitationToken()
	if err != nil {
		return Response{}, 500, err
	}
	expiresAt := now.Add(s.invitationTTL)
	var invitation UserInvitation
	existingInvitation, err := s.repo.FindUserInvitationByEmail(email)
	if err == nil {
		if existingInvitation.Status == "pending" && now.Before(existingInvitation.ExpiresAt) {
			if existingInvitation.SentAt != nil {
				return Response{}, statusCode(ErrUserInvitationPending), ErrUserInvitationPending
			}
			refreshed, err := s.repo.RefreshUserInvitation(UserInvitationRefreshParams{
				ID:               existingInvitation.ID,
				TokenHash:        tokenHash,
				InvitedByAdminID: actor.AdminUserID,
				ExpiresAt:        expiresAt,
			})
			if err != nil {
				return Response{}, statusCode(err), err
			}
			invitation = refreshed
		} else {
			return Response{}, statusCode(ErrEmailAlreadyExists), ErrEmailAlreadyExists
		}
	} else if !errors.Is(err, ErrUserNotFound) {
		return Response{}, statusCode(err), err
	} else {
		invitation, err = s.repo.CreateUserInvitation(UserInvitationCreateParams{
			Email:            email,
			TokenHash:        tokenHash,
			Status:           "pending",
			InvitedByAdminID: actor.AdminUserID,
			ExpiresAt:        expiresAt,
		})
		if err != nil {
			return Response{}, statusCode(err), err
		}
	}

	inviteURL := buildInvitationURL(s.invitationBaseURL, token)
	if err := s.invitationMailer.SendUserInvitation(invitation.Email, inviteURL, invitation.ExpiresAt); err != nil {
		return Response{}, 502, err
	}
	sentAt := time.Now().UTC()
	if err := s.repo.MarkUserInvitationSent(invitation.ID, sentAt); err != nil {
		return Response{}, 500, err
	}
	invitation.SentAt = &sentAt

	return Response{
		Success: true,
		Code:    "USER_INVITATION_CREATED",
		Message: "邀請已建立",
		Data: map[string]any{
			"id":         invitation.ID,
			"email":      invitation.Email,
			"token":      token,
			"invite_url": inviteURL,
			"expires_at": invitation.ExpiresAt.Format(time.RFC3339),
		},
	}, 201, nil
}

// ResendUserInvitation refreshes and resends an invitation that has not been completed.
func (s *Service) ResendUserInvitation(req AdminInviteUserRequest, actor AdminSessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}
	if err := s.requireSystemAdmin(actor.AdminUserID); err != nil {
		return Response{}, statusCode(err), err
	}

	email := normalizeEmail(req.Email)
	if !isValidEmail(email) {
		return Response{}, statusCode(ErrInvalidEmail), ErrInvalidEmail
	}

	if _, err := s.repo.FindUserByEmail(email); err == nil {
		return Response{}, statusCode(ErrEmailAlreadyExists), ErrEmailAlreadyExists
	} else if !errors.Is(err, ErrUserNotFound) {
		return Response{}, statusCode(err), err
	}

	if s.invitationMailer == nil {
		return Response{}, statusCode(ErrInvitationMailerUnavailable), ErrInvitationMailerUnavailable
	}

	existingInvitation, err := s.repo.FindUserInvitationByEmail(email)
	if err != nil {
		return Response{}, statusCode(err), err
	}
	if existingInvitation.Status == "accepted" || existingInvitation.AcceptedAt != nil || existingInvitation.AcceptedUserID > 0 {
		return Response{}, statusCode(ErrUserInvitationCompleted), ErrUserInvitationCompleted
	}

	token, tokenHash, err := newInvitationToken()
	if err != nil {
		return Response{}, 500, err
	}
	expiresAt := time.Now().UTC().Add(s.invitationTTL)
	invitation, err := s.repo.RefreshUserInvitation(UserInvitationRefreshParams{
		ID:               existingInvitation.ID,
		TokenHash:        tokenHash,
		InvitedByAdminID: actor.AdminUserID,
		ExpiresAt:        expiresAt,
	})
	if err != nil {
		return Response{}, statusCode(err), err
	}

	inviteURL := buildInvitationURL(s.invitationBaseURL, token)
	if err := s.invitationMailer.SendUserInvitation(invitation.Email, inviteURL, invitation.ExpiresAt); err != nil {
		return Response{}, 502, err
	}
	sentAt := time.Now().UTC()
	if err := s.repo.MarkUserInvitationSent(invitation.ID, sentAt); err != nil {
		return Response{}, 500, err
	}
	invitation.SentAt = &sentAt

	return Response{
		Success: true,
		Code:    "USER_INVITATION_RESENT",
		Message: "註冊邀請已重新發送",
		Data: map[string]any{
			"id":         invitation.ID,
			"email":      invitation.Email,
			"token":      token,
			"invite_url": inviteURL,
			"expires_at": invitation.ExpiresAt.Format(time.RFC3339),
		},
	}, 200, nil
}

// GetPublicUserInvitation returns the invited email for a valid pending invitation token.
func (s *Service) GetPublicUserInvitation(token string) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return Response{}, statusCode(ErrUserInvitationNotFound), ErrUserInvitationNotFound
	}

	tokenHash := invitationTokenHash(token)
	invitation, err := s.repo.FindUserInvitationByTokenHash(tokenHash)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return Response{}, statusCode(ErrUserInvitationNotFound), ErrUserInvitationNotFound
		}
		return Response{}, statusCode(err), err
	}

	now := time.Now().UTC()
	if invitation.Status != "pending" ||
		!now.Before(invitation.ExpiresAt) ||
		invitation.AcceptedAt != nil ||
		invitation.AcceptedUserID > 0 {
		return Response{}, statusCode(ErrUserInvitationNotFound), ErrUserInvitationNotFound
	}

	return Response{
		Success: true,
		Code:    "USER_INVITATION_FOUND",
		Message: "邀請有效",
		Data: PublicUserInvitation{
			Email: invitation.Email,
		},
	}, 200, nil
}

// AcceptUserInvitation completes registration for a valid invitation token.
func (s *Service) AcceptUserInvitation(req AcceptUserInvitationRequest, clientIP, userAgent string) (Response, int, error) {
	if s == nil || s.repo == nil || s.sessions == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}

	token := strings.TrimSpace(req.Token)
	if token == "" {
		return Response{}, statusCode(ErrUserInvitationNotFound), ErrUserInvitationNotFound
	}

	account := strings.TrimSpace(req.Account)
	if account == "" || len([]rune(account)) > 100 {
		return Response{}, statusCode(ErrInvalidExternalUserID), ErrInvalidExternalUserID
	}

	nickname := strings.TrimSpace(req.Nickname)
	if nickname == "" || len([]rune(nickname)) > 100 {
		return Response{}, statusCode(ErrInvalidDisplayName), ErrInvalidDisplayName
	}

	password := strings.TrimSpace(req.Password)
	if password != strings.TrimSpace(req.PasswordConfirmation) {
		return Response{}, statusCode(ErrPasswordConfirmation), ErrPasswordConfirmation
	}

	passwordHash, err := hashPassword(password)
	if err != nil {
		return Response{}, statusCode(err), err
	}

	deviceID := strings.TrimSpace(req.DeviceID)
	if deviceID == "" {
		return Response{}, statusCode(ErrDeviceIDRequired), ErrDeviceIDRequired
	}

	tokenHash := invitationTokenHash(token)
	invitation, err := s.repo.FindUserInvitationByTokenHash(tokenHash)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return Response{}, statusCode(ErrUserInvitationNotFound), ErrUserInvitationNotFound
		}
		return Response{}, statusCode(err), err
	}
	if invitation.Status != "pending" ||
		!time.Now().UTC().Before(invitation.ExpiresAt) ||
		invitation.AcceptedAt != nil ||
		invitation.AcceptedUserID > 0 {
		return Response{}, statusCode(ErrUserInvitationNotFound), ErrUserInvitationNotFound
	}

	if _, err := s.repo.FindUserByExternalID(account); err == nil {
		return Response{}, statusCode(ErrUserAlreadyExists), ErrUserAlreadyExists
	} else if !errors.Is(err, ErrUserNotFound) {
		return Response{}, statusCode(err), err
	}

	if _, err := s.repo.FindUserByEmail(invitation.Email); err == nil {
		return Response{}, statusCode(ErrEmailAlreadyExists), ErrEmailAlreadyExists
	} else if !errors.Is(err, ErrUserNotFound) {
		return Response{}, statusCode(err), err
	}

	now := time.Now().UTC()
	user, err := s.repo.AcceptUserInvitation(AcceptUserInvitationParams{
		TokenHash:      tokenHash,
		SourceSystem:   "office",
		ExternalUserID: account,
		PasswordHash:   passwordHash,
		DisplayName:    nickname,
		Status:         "active",
		AcceptedAt:     now,
	})
	if err != nil {
		return Response{}, statusCode(err), err
	}

	if err := s.repo.UpsertDeviceLogin(user.ID, deviceID, userAgent, clientIP, true, now); err != nil {
		return Response{}, 500, err
	}

	sessionToken, expiresAt, err := s.sessions.Issue(user.ID, deviceID)
	if err != nil {
		return Response{}, 500, err
	}

	resp := Response{
		Success: true,
		Code:    "USER_INVITATION_ACCEPTED",
		Message: "註冊完成",
		Token:   sessionToken,
		Data: map[string]any{
			"role":             "user",
			"user_id":          user.ID,
			"source_system":    user.SourceSystem,
			"external_user_id": user.ExternalUserID,
			"display_name":     user.DisplayName,
			"email":            user.Email,
			"status":           user.Status,
			"created_at":       user.CreatedAt,
		},
		ExpiresAt: expiresAt.Format(time.RFC3339),
	}
	if data, err := s.profileResponseData(user); err == nil {
		data["email"] = user.Email
		data["status"] = user.Status
		data["created_at"] = user.CreatedAt
		resp.Data = data
	} else {
		resp.Data = AdminUserSummary{
			ID:             user.ID,
			ExternalUserID: user.ExternalUserID,
			DisplayName:    user.DisplayName,
			Email:          user.Email,
			Status:         user.Status,
			SourceSystem:   user.SourceSystem,
			CreatedAt:      user.CreatedAt,
		}
	}

	return resp, 201, nil
}

// CreateSystemAdmin creates a backend admin account.
func (s *Service) CreateSystemAdmin(req SystemAdminCreateRequest, actor AdminSessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}
	if err := s.requireSystemAdmin(actor.AdminUserID); err != nil {
		return Response{}, statusCode(err), err
	}

	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "active"
	}
	if status != "active" && status != "inactive" {
		return Response{}, statusCode(ErrInvalidStatus), ErrInvalidStatus
	}

	params, err := validateSystemAdminCreateRequest(req, actor.AdminUserID)
	if err != nil {
		return Response{}, statusCode(err), err
	}
	params.Status = status

	admin, err := s.repo.CreateSystemAdmin(params)
	if err != nil {
		return Response{}, statusCode(err), err
	}

	return Response{
		Success: true,
		Code:    "SYSTEM_ADMIN_CREATED",
		Message: "管理員已建立",
		Data:    admin,
	}, 201, nil
}

// BootstrapSystemAdmin creates a backend system admin.
// It is intended for first-time deployment and is safe to run more than once.
func (s *Service) BootstrapSystemAdmin(req SystemAdminCreateRequest) (SystemAdminBootstrapResult, error) {
	if s == nil || s.repo == nil {
		return SystemAdminBootstrapResult{}, fmt.Errorf("integration service unavailable")
	}

	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "active"
	}
	if status != "active" && status != "inactive" {
		return SystemAdminBootstrapResult{}, ErrInvalidStatus
	}

	params, err := validateSystemAdminCreateRequest(req, 0)
	if err != nil {
		return SystemAdminBootstrapResult{}, err
	}
	params.Status = status

	admin, err := s.repo.CreateSystemAdmin(params)
	adminCreated := true
	if errors.Is(err, ErrUserAlreadyExists) {
		adminCreated = false
		admin, err = s.repo.FindSystemAdminByAccount(params.Account)
		if err != nil {
			return SystemAdminBootstrapResult{}, err
		}
	} else if err != nil {
		return SystemAdminBootstrapResult{}, err
	}

	return SystemAdminBootstrapResult{
		Admin:        admin,
		AdminCreated: adminCreated,
	}, nil
}

// GetUser returns one user for editing. Only system_admin is allowed.
func (s *Service) GetUser(targetUserID int64, actor AdminSessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}
	if targetUserID <= 0 {
		return Response{}, statusCode(ErrInvalidUserID), ErrInvalidUserID
	}
	if err := s.requireSystemAdmin(actor.AdminUserID); err != nil {
		return Response{}, statusCode(err), err
	}

	user, err := s.repo.FindUserByID(targetUserID)
	if err != nil {
		return Response{}, statusCode(err), err
	}

	return Response{
		Success: true,
		Code:    "USER_OK",
		Message: "用户读取成功",
		Data:    adminUserSummary(user),
	}, 200, nil
}

// UpdateUser updates mutable user fields. A blank password keeps the current password.
func (s *Service) UpdateUser(targetUserID int64, req AdminUpdateUserRequest, actor AdminSessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}
	if targetUserID <= 0 {
		return Response{}, statusCode(ErrInvalidUserID), ErrInvalidUserID
	}
	if err := s.requireSystemAdmin(actor.AdminUserID); err != nil {
		return Response{}, statusCode(err), err
	}

	displayName := strings.TrimSpace(req.DisplayName)
	email := strings.TrimSpace(req.Email)
	status := strings.TrimSpace(req.Status)
	if displayName == "" || len([]rune(displayName)) > 100 {
		return Response{}, statusCode(ErrInvalidDisplayName), ErrInvalidDisplayName
	}
	if len(email) > 255 || (email != "" && !strings.Contains(email, "@")) {
		return Response{}, statusCode(ErrInvalidEmail), ErrInvalidEmail
	}
	if status != "active" && status != "inactive" {
		return Response{}, statusCode(ErrInvalidStatus), ErrInvalidStatus
	}

	params := AdminUpdateUserParams{
		DisplayName: displayName,
		Email:       email,
		Status:      status,
	}
	if strings.TrimSpace(req.Password) != "" {
		passwordHash, err := hashPassword(req.Password)
		if err != nil {
			return Response{}, statusCode(err), err
		}
		params.PasswordHash = passwordHash
	}

	user, err := s.repo.UpdateUser(targetUserID, params)
	if err != nil {
		return Response{}, statusCode(err), err
	}

	return Response{
		Success: true,
		Code:    "USER_UPDATED",
		Message: "用户更新成功",
		Data:    adminUserSummary(user),
	}, 200, nil
}

// UpdateUserChatMute toggles whether a user can send chat messages.
func (s *Service) UpdateUserChatMute(targetUserID int64, req AdminUpdateUserChatMuteRequest, actor AdminSessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}
	if targetUserID <= 0 {
		return Response{}, statusCode(ErrInvalidUserID), ErrInvalidUserID
	}
	if err := s.requireSystemAdmin(actor.AdminUserID); err != nil {
		return Response{}, statusCode(err), err
	}

	user, err := s.repo.UpdateUserChatMute(targetUserID, AdminUpdateUserChatMuteParams{IsChatMuted: req.IsChatMuted})
	if err != nil {
		return Response{}, statusCode(err), err
	}

	code := "USER_CHAT_UNMUTED"
	message := "用户已解除禁言"
	if user.IsChatMuted {
		code = "USER_CHAT_MUTED"
		message = "用户已禁止发言"
	}

	return Response{
		Success: true,
		Code:    code,
		Message: message,
		Data:    adminUserSummary(user),
	}, 200, nil
}

// SendTemporaryPassword sends a generated temporary password to a user's email.
func (s *Service) SendTemporaryPassword(targetUserID int64, actor AdminSessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}
	if targetUserID <= 0 {
		return Response{}, statusCode(ErrInvalidUserID), ErrInvalidUserID
	}
	if err := s.requireSystemAdmin(actor.AdminUserID); err != nil {
		return Response{}, statusCode(err), err
	}
	if s.invitationMailer == nil {
		return Response{}, statusCode(ErrInvitationMailerUnavailable), ErrInvitationMailerUnavailable
	}

	user, err := s.repo.FindUserByID(targetUserID)
	if err != nil {
		return Response{}, statusCode(err), err
	}
	email := strings.TrimSpace(user.Email)
	if email == "" || !strings.Contains(email, "@") {
		return Response{}, statusCode(ErrInvalidEmail), ErrInvalidEmail
	}

	temporaryPassword, err := newTemporaryPassword()
	if err != nil {
		return Response{}, 500, err
	}
	passwordHash, err := hashPassword(temporaryPassword)
	if err != nil {
		return Response{}, statusCode(err), err
	}
	expiresAt := time.Now().UTC().Add(30 * time.Minute)

	if err := s.invitationMailer.SendTemporaryPassword(email, temporaryPassword, expiresAt); err != nil {
		return Response{}, statusCode(ErrTemporaryPasswordMailFailed), ErrTemporaryPasswordMailFailed
	}
	updated, err := s.repo.UpdateUserTemporaryPassword(targetUserID, TemporaryPasswordParams{
		PasswordHash: passwordHash,
		ExpiresAt:    expiresAt,
	})
	if err != nil {
		return Response{}, statusCode(err), err
	}

	return Response{
		Success: true,
		Code:    "TEMPORARY_PASSWORD_SENT",
		Message: "臨時密碼已寄出",
		Data:    adminUserSummary(updated),
	}, 200, nil
}

// GetSystemAdmin returns one backend admin for editing. Only system_admin is allowed.
func (s *Service) GetSystemAdmin(targetAdminID int64, actor AdminSessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}
	if targetAdminID <= 0 {
		return Response{}, statusCode(ErrInvalidUserID), ErrInvalidUserID
	}
	if err := s.requireSystemAdmin(actor.AdminUserID); err != nil {
		return Response{}, statusCode(err), err
	}

	admin, err := s.repo.FindSystemAdminByID(targetAdminID)
	if err != nil {
		return Response{}, statusCode(err), err
	}

	return Response{
		Success: true,
		Code:    "SYSTEM_ADMIN_OK",
		Message: "管理員讀取成功",
		Data:    admin,
	}, 200, nil
}

// UpdateSystemAdmin updates mutable backend admin fields. A blank password keeps the current password.
func (s *Service) UpdateSystemAdmin(targetAdminID int64, req SystemAdminCreateRequest, actor AdminSessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}
	if targetAdminID <= 0 {
		return Response{}, statusCode(ErrInvalidUserID), ErrInvalidUserID
	}
	if err := s.requireSystemAdmin(actor.AdminUserID); err != nil {
		return Response{}, statusCode(err), err
	}

	displayName := strings.TrimSpace(req.DisplayName)
	role := normalizeSystemAdminRole(req.Role)
	status := strings.TrimSpace(req.Status)
	if displayName == "" || len([]rune(displayName)) > 100 {
		return Response{}, statusCode(ErrInvalidDisplayName), ErrInvalidDisplayName
	}
	if status != "active" && status != "inactive" {
		return Response{}, statusCode(ErrInvalidStatus), ErrInvalidStatus
	}

	params := SystemAdminUpdateParams{
		DisplayName: displayName,
		Role:        role,
		Status:      status,
	}
	if strings.TrimSpace(req.Password) != "" {
		passwordHash, err := hashPassword(req.Password)
		if err != nil {
			return Response{}, statusCode(err), err
		}
		params.PasswordHash = passwordHash
	}

	admin, err := s.repo.UpdateSystemAdmin(targetAdminID, params)
	if err != nil {
		return Response{}, statusCode(err), err
	}

	return Response{
		Success: true,
		Code:    "SYSTEM_ADMIN_UPDATED",
		Message: "管理員資料已更新",
		Data:    admin,
	}, 200, nil
}

// CreateAdminRole creates a backend admin role.
func (s *Service) CreateAdminRole(req AdminRoleCreateRequest, actor AdminSessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}
	if err := s.requireSystemAdmin(actor.AdminUserID); err != nil {
		return Response{}, statusCode(err), err
	}

	params, err := validateAdminRoleCreateRequest(req, actor.AdminUserID)
	if err != nil {
		return Response{}, statusCode(err), err
	}

	role, err := s.repo.CreateAdminRole(params)
	if err != nil {
		return Response{}, statusCode(err), err
	}

	return Response{
		Success: true,
		Code:    "ADMIN_ROLE_CREATED",
		Message: "角色已新增",
		Data:    role,
	}, 201, nil
}

// GetAdminRole returns one backend admin role for editing.
func (s *Service) GetAdminRole(roleID int64, actor AdminSessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}
	if roleID <= 0 {
		return Response{}, statusCode(ErrInvalidRole), ErrInvalidRole
	}
	if err := s.requireSystemAdmin(actor.AdminUserID); err != nil {
		return Response{}, statusCode(err), err
	}

	role, err := s.repo.FindAdminRoleByID(roleID)
	if err != nil {
		return Response{}, statusCode(err), err
	}

	return Response{
		Success: true,
		Code:    "ADMIN_ROLE_OK",
		Message: "角色讀取成功",
		Data:    role,
	}, 200, nil
}

// UpdateAdminRole updates mutable backend admin role fields.
func (s *Service) UpdateAdminRole(roleID int64, req AdminRoleCreateRequest, actor AdminSessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}
	if roleID <= 0 {
		return Response{}, statusCode(ErrInvalidRole), ErrInvalidRole
	}
	if err := s.requireSystemAdmin(actor.AdminUserID); err != nil {
		return Response{}, statusCode(err), err
	}

	params, err := validateAdminRoleUpdateRequest(req)
	if err != nil {
		return Response{}, statusCode(err), err
	}

	role, err := s.repo.UpdateAdminRole(roleID, params)
	if err != nil {
		return Response{}, statusCode(err), err
	}

	return Response{
		Success: true,
		Code:    "ADMIN_ROLE_UPDATED",
		Message: "角色已更新",
		Data:    role,
	}, 200, nil
}

// GetCurrentAdminPermissions returns enabled route permission keys for the current admin.
func (s *Service) GetCurrentAdminPermissions(actor AdminSessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}

	admin, err := s.repo.FindSystemAdminByID(actor.AdminUserID)
	if err != nil {
		return Response{}, statusCode(err), err
	}
	role, err := s.repo.FindAdminRoleByCode(admin.Role)
	if err != nil {
		return Response{}, statusCode(err), err
	}
	stored, err := s.repo.ListAdminRolePermissions(role.ID)
	if err != nil {
		return Response{}, 500, err
	}

	state := buildAdminPermissionState(role, stored)
	permissions := make([]string, 0, len(state))
	for key, enabled := range state {
		if enabled {
			permissions = append(permissions, key)
		}
	}
	sort.Strings(permissions)

	return Response{
		Success: true,
		Code:    "CURRENT_ADMIN_PERMISSIONS_OK",
		Message: "目前管理員權限讀取成功",
		Data: CurrentAdminPermissionResponse{
			Permissions: permissions,
		},
	}, 200, nil
}

// GetAdminRolePermissions returns the editable backend permission switches for a role.
func (s *Service) GetAdminRolePermissions(roleID int64, keyword string, actor AdminSessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}
	if roleID <= 0 {
		return Response{}, statusCode(ErrInvalidRole), ErrInvalidRole
	}
	if err := s.requireSystemAdmin(actor.AdminUserID); err != nil {
		return Response{}, statusCode(err), err
	}

	role, err := s.repo.FindAdminRoleByID(roleID)
	if err != nil {
		return Response{}, statusCode(err), err
	}
	stored, err := s.repo.ListAdminRolePermissions(roleID)
	if err != nil {
		return Response{}, 500, err
	}

	detail := AdminRolePermissionDetail{
		Role:   role,
		Groups: buildAdminPermissionGroups(role, stored, keyword),
	}
	return Response{
		Success: true,
		Code:    "ADMIN_ROLE_PERMISSIONS_OK",
		Message: "角色權限讀取成功",
		Data:    detail,
	}, 200, nil
}

// UpdateAdminRolePermissions replaces the backend permission switches for a role.
func (s *Service) UpdateAdminRolePermissions(roleID int64, req AdminRolePermissionUpdateRequest, actor AdminSessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}
	if roleID <= 0 {
		return Response{}, statusCode(ErrInvalidRole), ErrInvalidRole
	}
	if err := s.requireSystemAdmin(actor.AdminUserID); err != nil {
		return Response{}, statusCode(err), err
	}

	role, err := s.repo.FindAdminRoleByID(roleID)
	if err != nil {
		return Response{}, statusCode(err), err
	}
	updates, err := validateAdminRolePermissionUpdate(req)
	if err != nil {
		return Response{}, statusCode(err), err
	}
	stored, err := s.repo.ListAdminRolePermissions(roleID)
	if err != nil {
		return Response{}, 500, err
	}
	permissions := buildAdminPermissionState(role, stored)
	for key, enabled := range updates {
		permissions[key] = enabled
	}
	if err := s.repo.ReplaceAdminRolePermissions(roleID, permissions); err != nil {
		return Response{}, 500, err
	}

	detail := AdminRolePermissionDetail{
		Role:   role,
		Groups: buildAdminPermissionGroups(role, permissions, ""),
	}
	return Response{
		Success: true,
		Code:    "ADMIN_ROLE_PERMISSIONS_UPDATED",
		Message: "角色權限已更新",
		Data:    detail,
	}, 200, nil
}

// GetSystemSettings returns editable backend system settings.
func (s *Service) GetSystemSettings(actor AdminSessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}
	if !s.adminPrincipalHasPermission(actor, "office.system-settings.index") {
		if err := s.requireSystemAdmin(actor.AdminUserID); err != nil {
			return Response{}, statusCode(err), err
		}
	}

	settings, err := s.repo.GetSystemSettings()
	if err != nil {
		return Response{}, 500, err
	}
	normalized := normalizeSystemSettings(settings)
	return Response{
		Success: true,
		Code:    "SYSTEM_SETTINGS_OK",
		Message: "系統設定讀取成功",
		Data:    normalized,
	}, 200, nil
}

// UpdateSystemSettings updates editable backend system settings.
func (s *Service) UpdateSystemSettings(req SystemSettingsUpdateRequest, actor AdminSessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}
	if !s.adminPrincipalHasPermission(actor, "office.system-settings.edit") {
		if err := s.requireSystemAdmin(actor.AdminUserID); err != nil {
			return Response{}, statusCode(err), err
		}
	}

	settings, err := validateSystemSettingsUpdate(req)
	if err != nil {
		return Response{}, statusCode(err), err
	}
	updated, err := s.repo.UpdateSystemSettings(settings, actor.AdminUserID)
	if err != nil {
		return Response{}, 500, err
	}
	return Response{
		Success: true,
		Code:    "SYSTEM_SETTINGS_UPDATED",
		Message: "系統設定已更新",
		Data:    normalizeSystemSettings(updated),
	}, 200, nil
}

// ListDevices returns the recent devices for a target user. Only system_admin is allowed.
func (s *Service) ListDevices(targetUserID int64, actor AdminSessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}
	if targetUserID <= 0 {
		return Response{}, statusCode(ErrInvalidUserID), ErrInvalidUserID
	}
	if err := s.requireSystemAdmin(actor.AdminUserID); err != nil {
		return Response{}, statusCode(err), err
	}

	if _, err := s.repo.FindUserByID(targetUserID); err != nil {
		return Response{}, statusCode(err), err
	}

	devices, err := s.repo.ListDevices(targetUserID)
	if err != nil {
		return Response{}, 500, err
	}

	summaries := make([]DeviceSummary, 0, len(devices))
	for _, device := range devices {
		summaries = append(summaries, DeviceSummary{
			DeviceID:        device.DeviceID,
			UserAgent:       device.UserAgent,
			IP:              device.IP,
			FirstLoginAt:    device.FirstLoginAt.Format(time.RFC3339),
			LastLoginAt:     device.LastLoginAt.Format(time.RFC3339),
			IsTrustedDevice: device.IsTrustedDevice,
		})
	}

	return Response{Success: true, Code: "DEVICES_OK", Message: "装置列表读取成功", Data: summaries}, 200, nil
}

// UpdateIPWhitelist replaces a user's whitelist settings. Only system_admin is allowed.
func (s *Service) UpdateIPWhitelist(targetUserID int64, actor AdminSessionPrincipal, req IPWhitelistUpdateRequest) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}
	if targetUserID <= 0 {
		return Response{}, statusCode(ErrInvalidUserID), ErrInvalidUserID
	}
	if err := s.requireSystemAdmin(actor.AdminUserID); err != nil {
		return Response{}, statusCode(err), err
	}

	if _, err := s.repo.FindUserByID(targetUserID); err != nil {
		return Response{}, statusCode(err), err
	}

	rules, err := validateWhitelistRules(req.AllowAll, req.Rules)
	if err != nil {
		return Response{}, statusCode(err), err
	}

	if err := s.repo.ReplaceIPWhitelist(targetUserID, req.AllowAll, rules); err != nil {
		return Response{}, 500, err
	}

	return Response{
		Success: true,
		Code:    "IP_WHITELIST_UPDATED",
		Message: "IP 白名单更新成功",
		Data:    IPWhitelistSettings{AllowAll: req.AllowAll, Rules: rules},
	}, 200, nil
}

// ApproveDevice trusts a pending device. It can be approved by system_admin or the same user with another trusted device session.
func (s *Service) ApproveDevice(targetUserID int64, targetDeviceID string, actor SessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}
	if targetUserID <= 0 {
		return Response{}, statusCode(ErrInvalidUserID), ErrInvalidUserID
	}
	if strings.TrimSpace(targetDeviceID) == "" {
		return Response{}, statusCode(ErrDeviceIDRequired), ErrDeviceIDRequired
	}

	if _, err := s.repo.FindUserByID(targetUserID); err != nil {
		return Response{}, statusCode(err), err
	}

	targetDevice, err := s.repo.FindDevice(targetUserID, strings.TrimSpace(targetDeviceID))
	if err != nil {
		return Response{}, statusCode(err), err
	}

	isAdmin, err := s.repo.IsSystemAdmin(actor.UserID)
	if err != nil {
		return Response{}, 500, err
	}
	if !isAdmin {
		if actor.UserID != targetUserID {
			return Response{}, statusCode(ErrInsufficientRole), ErrInsufficientRole
		}
		if strings.TrimSpace(actor.DeviceID) == "" || actor.DeviceID == targetDevice.DeviceID {
			return Response{}, statusCode(ErrDeviceNotTrusted), ErrDeviceNotTrusted
		}

		approverDevice, err := s.repo.FindDevice(targetUserID, strings.TrimSpace(actor.DeviceID))
		if err != nil {
			return Response{}, statusCode(err), err
		}
		if !approverDevice.IsTrustedDevice {
			return Response{}, statusCode(ErrDeviceNotTrusted), ErrDeviceNotTrusted
		}
	}

	now := time.Now().UTC()
	if err := s.repo.ApproveDevice(targetUserID, targetDevice.DeviceID, actor.UserID, now); err != nil {
		return Response{}, statusCode(err), err
	}

	return Response{Success: true, Code: "DEVICE_APPROVED", Message: "设备已设为信任设备"}, 200, nil
}

func validateRegisterRequest(req RegisterRequest) (RegisterParams, error) {
	params := RegisterParams{
		SourceSystem:    strings.TrimSpace(req.SourceSystem),
		ExternalUserID:  strings.TrimSpace(req.ExternalUserID),
		DisplayName:     strings.TrimSpace(req.DisplayName),
		Email:           strings.TrimSpace(req.Email),
		Language:        strings.TrimSpace(req.Language),
		WhatsAppAccount: strings.TrimSpace(req.WhatsAppAccount),
		TelegramAccount: strings.TrimSpace(req.TelegramAccount),
		Status:          "active",
	}

	if !sourceSystemPattern.MatchString(params.SourceSystem) {
		return RegisterParams{}, ErrInvalidSourceSystem
	}
	if params.ExternalUserID == "" || len([]rune(params.ExternalUserID)) > 100 {
		return RegisterParams{}, ErrInvalidExternalUserID
	}
	if params.DisplayName == "" || len([]rune(params.DisplayName)) > 100 {
		return RegisterParams{}, ErrInvalidDisplayName
	}
	if len(params.Email) > 255 || (params.Email != "" && !strings.Contains(params.Email, "@")) {
		return RegisterParams{}, ErrInvalidEmail
	}
	if err := validatePassword(req.Password); err != nil {
		return RegisterParams{}, err
	}
	if params.Language == "" {
		params.Language = "zh-Hans"
	}

	return params, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func isValidEmail(email string) bool {
	if email == "" || len(email) > 255 {
		return false
	}
	at := strings.Index(email, "@")
	return at > 0 && at < len(email)-1 && strings.Contains(email[at+1:], ".")
}

func newInvitationToken() (string, string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", fmt.Errorf("generate invitation token: %w", err)
	}

	token := hex.EncodeToString(raw)
	return token, invitationTokenHash(token), nil
}

func invitationTokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func buildInvitationURL(baseURL, token string) string {
	if strings.TrimSpace(baseURL) == "" {
		return "/invite?token=" + token
	}
	return strings.TrimRight(strings.TrimSpace(baseURL), "/") + "/invite?token=" + token
}

func (s *Service) createUser(params RegisterParams, password string) (User, error) {
	passwordHash, err := hashPassword(password)
	if err != nil {
		return User{}, err
	}
	params.PasswordHash = passwordHash

	_, err = s.repo.FindUserByExternalID(params.ExternalUserID)
	if err == nil {
		return User{}, ErrUserAlreadyExists
	}
	if !errors.Is(err, ErrUserNotFound) {
		return User{}, err
	}

	user, err := s.repo.CreateUser(params)
	if err != nil {
		if errors.Is(err, ErrUserAlreadyExists) {
			return User{}, ErrUserAlreadyExists
		}
		return User{}, err
	}
	return user, nil
}

func validateLoginRequest(req LoginRequest) (string, string, string, string, error) {
	sourceSystem := strings.TrimSpace(req.SourceSystem)
	externalUserID := strings.TrimSpace(req.ExternalUserID)
	password := strings.TrimSpace(req.Password)
	deviceID := strings.TrimSpace(req.DeviceID)

	if !sourceSystemPattern.MatchString(sourceSystem) {
		return "", "", "", "", ErrInvalidSourceSystem
	}
	if externalUserID == "" {
		return "", "", "", "", ErrInvalidExternalUserID
	}
	if err := validatePassword(password); err != nil {
		return "", "", "", "", err
	}
	if deviceID == "" {
		return "", "", "", "", ErrDeviceIDRequired
	}

	return sourceSystem, externalUserID, password, deviceID, nil
}

func validateAdminLoginRequest(req AdminLoginRequest) (string, string, string, error) {
	account := strings.TrimSpace(req.Account)
	password := strings.TrimSpace(req.Password)
	deviceID := strings.TrimSpace(req.DeviceID)

	if account == "" || len([]rune(account)) > 100 {
		return "", "", "", ErrInvalidExternalUserID
	}
	if err := validatePassword(password); err != nil {
		return "", "", "", err
	}
	if deviceID == "" {
		return "", "", "", ErrDeviceIDRequired
	}

	return account, password, deviceID, nil
}

func validateSystemAdminCreateRequest(req SystemAdminCreateRequest, createdBy int64) (SystemAdminCreateParams, error) {
	account := strings.TrimSpace(req.ExternalUserID)
	displayName := strings.TrimSpace(req.DisplayName)
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "active"
	}
	if account == "" || len([]rune(account)) > 100 {
		return SystemAdminCreateParams{}, ErrInvalidExternalUserID
	}
	if displayName == "" || len([]rune(displayName)) > 100 {
		return SystemAdminCreateParams{}, ErrInvalidDisplayName
	}
	if status != "active" && status != "inactive" {
		return SystemAdminCreateParams{}, ErrInvalidStatus
	}

	passwordHash, err := hashPassword(req.Password)
	if err != nil {
		return SystemAdminCreateParams{}, err
	}

	return SystemAdminCreateParams{
		Account:      account,
		PasswordHash: passwordHash,
		DisplayName:  displayName,
		Role:         normalizeSystemAdminRole(req.Role),
		Status:       status,
		CreatedBy:    createdBy,
	}, nil
}

func normalizeSystemAdminRole(role string) string {
	role = strings.TrimSpace(role)
	if role == "" {
		return "system_admin"
	}
	return role
}

func validateAdminRoleCreateRequest(req AdminRoleCreateRequest, createdBy int64) (AdminRoleCreateParams, error) {
	code := strings.TrimSpace(req.Code)
	name := strings.TrimSpace(req.Name)
	status, err := normalizeAdminRoleStatus(req.Status)
	if err != nil {
		return AdminRoleCreateParams{}, err
	}
	if !adminRoleCodePattern.MatchString(code) {
		return AdminRoleCreateParams{}, ErrInvalidRole
	}
	if name == "" || len([]rune(name)) > 100 {
		return AdminRoleCreateParams{}, ErrInvalidRole
	}
	return AdminRoleCreateParams{
		Code:      code,
		Name:      name,
		Status:    status,
		CreatedBy: createdBy,
	}, nil
}

func validateAdminRoleUpdateRequest(req AdminRoleCreateRequest) (AdminRoleUpdateParams, error) {
	name := strings.TrimSpace(req.Name)
	status, err := normalizeAdminRoleStatus(req.Status)
	if err != nil {
		return AdminRoleUpdateParams{}, err
	}
	if name == "" || len([]rune(name)) > 100 {
		return AdminRoleUpdateParams{}, ErrInvalidRole
	}
	return AdminRoleUpdateParams{Name: name, Status: status}, nil
}

func normalizeAdminRoleStatus(status string) (string, error) {
	status = strings.TrimSpace(status)
	if status == "" {
		status = "active"
	}
	if status != "active" && status != "inactive" {
		return "", ErrInvalidStatus
	}
	return status, nil
}

func validateSystemSettingsUpdate(req SystemSettingsUpdateRequest) (SystemSettings, error) {
	settings := SystemSettings{
		ChatAutoDeleteMaxDays:         req.ChatAutoDeleteMaxDays,
		AdminChatHistoryRetentionDays: req.AdminChatHistoryRetentionDays,
	}
	if settings.ChatAutoDeleteMaxDays < 1 || settings.ChatAutoDeleteMaxDays > 30 {
		return SystemSettings{}, ErrInvalidRole
	}
	if settings.AdminChatHistoryRetentionDays < 1 {
		return SystemSettings{}, ErrInvalidRole
	}
	return settings, nil
}

func normalizeSystemSettings(settings SystemSettings) SystemSettings {
	if settings.ChatAutoDeleteMaxDays < 1 {
		settings.ChatAutoDeleteMaxDays = 30
	}
	if settings.ChatAutoDeleteMaxDays > 30 {
		settings.ChatAutoDeleteMaxDays = 30
	}
	if settings.AdminChatHistoryRetentionDays < 1 {
		settings.AdminChatHistoryRetentionDays = 90
	}
	return settings
}

func (s *Service) applyAdminChatHistoryRetention(filter *AdminConversationFilter) error {
	cutoff, err := s.adminChatHistoryCutoff()
	if err != nil {
		return err
	}
	if filter.LastActivityFrom == nil || filter.LastActivityFrom.Before(cutoff) {
		filter.LastActivityFrom = &cutoff
	}
	return nil
}

func (s *Service) applyAdminChatMessageRetention(filter *AdminConversationMessageFilter) error {
	cutoff, err := s.adminChatHistoryCutoff()
	if err != nil {
		return err
	}
	if filter.SentFrom == nil || filter.SentFrom.Before(cutoff) {
		filter.SentFrom = &cutoff
	}
	return nil
}

func (s *Service) adminChatHistoryCutoff() (time.Time, error) {
	settings, err := s.repo.GetSystemSettings()
	if err != nil {
		return time.Time{}, err
	}
	settings = normalizeSystemSettings(settings)
	return time.Now().AddDate(0, 0, -settings.AdminChatHistoryRetentionDays), nil
}

func buildAdminPermissionGroups(role AdminRoleSummary, stored map[string]bool, keyword string) []AdminPermissionGroup {
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	groups := make([]AdminPermissionGroup, 0, len(adminPermissionCatalog))
	for _, group := range adminPermissionCatalog {
		items := make([]AdminPermissionItem, 0, len(group.Items))
		for _, item := range group.Items {
			if keyword != "" && !adminPermissionMatchesKeyword(group, item, keyword) {
				continue
			}
			enabled, ok := stored[item.Key]
			if !ok && role.Code == "system_admin" {
				enabled = true
			}
			items = append(items, AdminPermissionItem{
				Key:      item.Key,
				Name:     item.Name,
				Category: item.Category,
				Enabled:  enabled,
			})
		}
		if len(items) == 0 {
			continue
		}
		groups = append(groups, AdminPermissionGroup{
			Key:   group.Key,
			Name:  group.Name,
			Items: items,
		})
	}
	return groups
}

func adminPermissionMatchesKeyword(group AdminPermissionGroup, item AdminPermissionItem, keyword string) bool {
	return strings.Contains(strings.ToLower(group.Key), keyword) ||
		strings.Contains(strings.ToLower(group.Name), keyword) ||
		strings.Contains(strings.ToLower(item.Key), keyword) ||
		strings.Contains(strings.ToLower(item.Name), keyword)
}

func buildAdminPermissionState(role AdminRoleSummary, stored map[string]bool) map[string]bool {
	allowed := adminPermissionKeySet()
	permissions := make(map[string]bool, len(allowed))
	for key := range allowed {
		if role.Code == "system_admin" {
			permissions[key] = true
		}
	}
	for key, enabled := range stored {
		if allowed[key] {
			permissions[key] = enabled
		}
	}
	return permissions
}

func validateAdminRolePermissionUpdate(req AdminRolePermissionUpdateRequest) (map[string]bool, error) {
	allowed := adminPermissionKeySet()
	permissions := make(map[string]bool, len(req.Permissions))
	for _, setting := range req.Permissions {
		key := strings.TrimSpace(setting.Key)
		if !allowed[key] {
			return nil, ErrInvalidRole
		}
		permissions[key] = setting.Enabled
	}
	return permissions, nil
}

func adminPermissionKeySet() map[string]bool {
	allowed := make(map[string]bool)
	for _, group := range adminPermissionCatalog {
		for _, item := range group.Items {
			allowed[item.Key] = true
		}
	}
	return allowed
}

func validatePassword(password string) error {
	if !passwordPattern.MatchString(strings.TrimSpace(password)) {
		return ErrInvalidPassword
	}
	return nil
}

func newTemporaryPassword() (string, error) {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func hashPassword(password string) (string, error) {
	password = strings.TrimSpace(password)
	if err := validatePassword(password); err != nil {
		return "", err
	}

	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	derived, err := pbkdf2.Key(sha256.New, password, salt, passwordHashIterations, 32)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("pbkdf2-sha256:%d:%s:%s", passwordHashIterations, base64.RawURLEncoding.EncodeToString(salt), hex.EncodeToString(derived)), nil
}

func passwordMatches(password, stored string) bool {
	password = strings.TrimSpace(password)
	parts := strings.Split(strings.TrimSpace(stored), ":")
	if len(parts) != 4 || parts[0] != "pbkdf2-sha256" {
		return false
	}

	iterations, err := strconv.Atoi(parts[1])
	if err != nil || iterations <= 0 {
		return false
	}

	salt, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return false
	}

	derived, err := pbkdf2.Key(sha256.New, password, salt, iterations, 32)
	if err != nil {
		return false
	}
	return hex.EncodeToString(derived) == parts[3]
}

func validateWhitelistRules(allowAll bool, rules []string) ([]string, error) {
	normalized := make([]string, 0, len(rules))
	seen := make(map[string]struct{}, len(rules))
	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}
		if _, ok := seen[rule]; ok {
			continue
		}
		if !isValidIPRule(rule) {
			return nil, ErrInvalidIPWhitelistRule
		}
		seen[rule] = struct{}{}
		normalized = append(normalized, rule)
	}

	if !allowAll && len(normalized) == 0 {
		return nil, ErrWhitelistRulesRequired
	}

	return normalized, nil
}

func (s *Service) requireSystemAdmin(userID int64) error {
	if userID <= 0 {
		return ErrInsufficientRole
	}
	admin, err := s.repo.FindSystemAdminByID(userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return ErrInsufficientRole
		}
		return err
	}
	if admin.Status != "active" || admin.Role != "system_admin" {
		return ErrInsufficientRole
	}

	return nil
}

func adminPasswordHash(admin SystemAdminSummary) string {
	return admin.PasswordHash
}

func (s *Service) profileResponseData(user User) (map[string]any, error) {
	role := "user"
	isAdmin, err := s.repo.IsSystemAdmin(user.ID)
	if err != nil {
		return nil, err
	}
	if isAdmin {
		role = "system_admin"
	}

	return map[string]any{
		"role":                          role,
		"user_id":                       user.ID,
		"source_system":                 user.SourceSystem,
		"external_user_id":              user.ExternalUserID,
		"display_name":                  user.DisplayName,
		"email":                         user.Email,
		"is_chat_muted":                 user.IsChatMuted,
		"must_change_password":          user.MustChangePassword,
		"temporary_password_expires_at": user.TemporaryPasswordExpiresAt,
	}, nil
}

func adminUserSummary(user User) AdminUserSummary {
	return AdminUserSummary{
		ID:                         user.ID,
		ExternalUserID:             user.ExternalUserID,
		DisplayName:                user.DisplayName,
		Email:                      user.Email,
		Status:                     user.Status,
		IsChatMuted:                user.IsChatMuted,
		MustChangePassword:         user.MustChangePassword,
		TemporaryPasswordExpiresAt: user.TemporaryPasswordExpiresAt,
		SourceSystem:               user.SourceSystem,
		CreatedAt:                  user.CreatedAt,
	}
}

func ipAllowed(clientIP string, rules []string) bool {
	addr := net.ParseIP(strings.TrimSpace(clientIP))
	if addr == nil {
		return false
	}

	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}
		if strings.Contains(rule, "/") {
			_, network, err := net.ParseCIDR(rule)
			if err == nil && network.Contains(addr) {
				return true
			}
			continue
		}
		if parsed := net.ParseIP(rule); parsed != nil && parsed.Equal(addr) {
			return true
		}
	}

	return false
}

func isValidIPRule(rule string) bool {
	rule = strings.TrimSpace(rule)
	if rule == "" {
		return false
	}
	if strings.Contains(rule, "/") {
		_, _, err := net.ParseCIDR(rule)
		return err == nil
	}

	return net.ParseIP(rule) != nil
}

func generateToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func statusCode(err error) int {
	switch {
	case errors.Is(err, ErrInvalidSourceSystem),
		errors.Is(err, ErrInvalidExternalUserID),
		errors.Is(err, ErrInvalidDisplayName),
		errors.Is(err, ErrInvalidEmail),
		errors.Is(err, ErrInvalidStatus),
		errors.Is(err, ErrInvalidRole),
		errors.Is(err, ErrInvalidPassword),
		errors.Is(err, ErrPasswordConfirmation),
		errors.Is(err, ErrInvalidUserID),
		errors.Is(err, ErrBulkInvitationTooMany),
		errors.Is(err, ErrDeviceIDRequired),
		errors.Is(err, ErrInvalidIPWhitelistRule),
		errors.Is(err, ErrWhitelistRulesRequired):
		return 400
	case errors.Is(err, ErrInvalidCredentials):
		return 401
	case errors.Is(err, ErrTemporaryPasswordExpired):
		return 403
	case errors.Is(err, ErrInvitationMailerUnavailable):
		return 503
	case errors.Is(err, ErrTemporaryPasswordMailFailed):
		return 502
	case errors.Is(err, ErrSystemAdminCannotChat),
		errors.Is(err, ErrInsufficientRole),
		errors.Is(err, ErrDeviceNotTrusted),
		errors.Is(err, ErrIPNotAllowed),
		errors.Is(err, ErrNewDeviceApprovalRequired):
		return 403
	case errors.Is(err, ErrUserNotFound),
		errors.Is(err, ErrDeviceNotFound),
		errors.Is(err, ErrUserInvitationNotFound),
		errors.Is(err, ErrConversationNotFound):
		return 404
	case errors.Is(err, ErrUserAlreadyExists),
		errors.Is(err, ErrRoleAlreadyExists):
		return 409
	case errors.Is(err, ErrEmailAlreadyExists),
		errors.Is(err, ErrUserInvitationPending),
		errors.Is(err, ErrUserInvitationCompleted):
		return 409
	default:
		return 500
	}
}
