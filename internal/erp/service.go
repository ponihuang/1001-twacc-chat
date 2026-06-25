package erp

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var sourceSystemPattern = regexp.MustCompile(`^[a-z0-9_-]{2,50}$`)
var passwordPattern = regexp.MustCompile(`^[A-Za-z0-9[:punct:]]{4,20}$`)

const passwordHashIterations = 120000

var (
	ErrInvalidSourceSystem       = errors.New("invalid source system")
	ErrInvalidExternalUserID     = errors.New("invalid external user id")
	ErrInvalidDisplayName        = errors.New("invalid display name")
	ErrInvalidEmail              = errors.New("invalid email")
	ErrInvalidStatus             = errors.New("invalid status")
	ErrInvalidPassword           = errors.New("invalid password")
	ErrInvalidCredentials        = errors.New("invalid credentials")
	ErrInvalidUserID             = errors.New("invalid user id")
	ErrDeviceIDRequired          = errors.New("device id required")
	ErrInvalidIPWhitelistRule    = errors.New("invalid ip whitelist rule")
	ErrWhitelistRulesRequired    = errors.New("whitelist rules required")
	ErrUserNotFound              = errors.New("user not found")
	ErrDeviceNotFound            = errors.New("device not found")
	ErrUserAlreadyExists         = errors.New("user already exists")
	ErrSystemAdminCannotChat     = errors.New("system admin cannot chat")
	ErrInsufficientRole          = errors.New("insufficient role")
	ErrDeviceNotTrusted          = errors.New("device not trusted")
	ErrIPNotAllowed              = errors.New("ip not allowed")
	ErrNewDeviceApprovalRequired = errors.New("new device approval required")
)

// Service implements external-system identity and login rules for chat access.
type Service struct {
	repo                 Repository
	sessions             SessionIssuer
	requireTrustedDevice bool
}

// NewService builds an integration-backed identity service.
func NewService(repo Repository, sessions SessionIssuer) *Service {
	return &Service{repo: repo, sessions: sessions}
}

// SetRequireTrustedDevice controls whether non-trusted devices must be approved before login.
func (s *Service) SetRequireTrustedDevice(require bool) {
	if s == nil {
		return
	}
	s.requireTrustedDevice = require
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

	user, err := s.repo.UpdateUserProfile(actor.UserID, displayName)
	if err != nil {
		return Response{}, statusCode(err), err
	}

	data, err := s.profileResponseData(user)
	if err != nil {
		return Response{}, 500, err
	}

	return Response{Success: true, Code: "PROFILE_UPDATED", Message: "个人资料更新成功", Data: data}, 200, nil
}

// ListUsers returns users for the admin console. Only system_admin is allowed.
func (s *Service) ListUsers(filter AdminUserFilter, actor SessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}
	if err := s.requireSystemAdmin(actor.UserID); err != nil {
		return Response{}, statusCode(err), err
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

// CreateUser creates an active user from the admin console. Only system_admin is allowed.
func (s *Service) CreateUser(req AdminCreateUserRequest, actor SessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}
	if err := s.requireSystemAdmin(actor.UserID); err != nil {
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

// GetUser returns one user for editing. Only system_admin is allowed.
func (s *Service) GetUser(targetUserID int64, actor SessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}
	if targetUserID <= 0 {
		return Response{}, statusCode(ErrInvalidUserID), ErrInvalidUserID
	}
	if err := s.requireSystemAdmin(actor.UserID); err != nil {
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
func (s *Service) UpdateUser(targetUserID int64, req AdminUpdateUserRequest, actor SessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}
	if targetUserID <= 0 {
		return Response{}, statusCode(ErrInvalidUserID), ErrInvalidUserID
	}
	if err := s.requireSystemAdmin(actor.UserID); err != nil {
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

// ListDevices returns the recent devices for a target user. Only system_admin is allowed.
func (s *Service) ListDevices(targetUserID int64, actor SessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}
	if targetUserID <= 0 {
		return Response{}, statusCode(ErrInvalidUserID), ErrInvalidUserID
	}
	if err := s.requireSystemAdmin(actor.UserID); err != nil {
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
func (s *Service) UpdateIPWhitelist(targetUserID int64, actor SessionPrincipal, req IPWhitelistUpdateRequest) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}
	if targetUserID <= 0 {
		return Response{}, statusCode(ErrInvalidUserID), ErrInvalidUserID
	}
	if err := s.requireSystemAdmin(actor.UserID); err != nil {
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

func (s *Service) createUser(params RegisterParams, password string) (User, error) {
	passwordHash, err := hashPassword(password)
	if err != nil {
		return User{}, err
	}
	params.PasswordHash = passwordHash

	_, err = s.repo.FindUserByExternal(params.SourceSystem, params.ExternalUserID)
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

func validatePassword(password string) error {
	if !passwordPattern.MatchString(strings.TrimSpace(password)) {
		return ErrInvalidPassword
	}
	return nil
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
	isAdmin, err := s.repo.IsSystemAdmin(userID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return ErrInsufficientRole
	}

	return nil
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
		"role":             role,
		"user_id":          user.ID,
		"source_system":    user.SourceSystem,
		"external_user_id": user.ExternalUserID,
		"display_name":     user.DisplayName,
	}, nil
}

func adminUserSummary(user User) AdminUserSummary {
	return AdminUserSummary{
		ID:             user.ID,
		ExternalUserID: user.ExternalUserID,
		DisplayName:    user.DisplayName,
		Email:          user.Email,
		Status:         user.Status,
		SourceSystem:   user.SourceSystem,
		CreatedAt:      user.CreatedAt,
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
		errors.Is(err, ErrInvalidPassword),
		errors.Is(err, ErrInvalidUserID),
		errors.Is(err, ErrDeviceIDRequired),
		errors.Is(err, ErrInvalidIPWhitelistRule),
		errors.Is(err, ErrWhitelistRulesRequired):
		return 400
	case errors.Is(err, ErrInvalidCredentials):
		return 401
	case errors.Is(err, ErrSystemAdminCannotChat),
		errors.Is(err, ErrInsufficientRole),
		errors.Is(err, ErrDeviceNotTrusted),
		errors.Is(err, ErrIPNotAllowed),
		errors.Is(err, ErrNewDeviceApprovalRequired):
		return 403
	case errors.Is(err, ErrUserNotFound),
		errors.Is(err, ErrDeviceNotFound):
		return 404
	case errors.Is(err, ErrUserAlreadyExists):
		return 409
	default:
		return 500
	}
}
