package erp

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"
)

var sourceSystemPattern = regexp.MustCompile(`^[a-z0-9_-]{2,50}$`)

var (
	ErrInvalidSourceSystem       = errors.New("invalid source system")
	ErrInvalidExternalUserID     = errors.New("invalid external user id")
	ErrInvalidDisplayName        = errors.New("invalid display name")
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
	repo     Repository
	sessions SessionIssuer
}

// NewService builds an integration-backed identity service.
func NewService(repo Repository, sessions SessionIssuer) *Service {
	return &Service{repo: repo, sessions: sessions}
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

	_, err = s.repo.FindUserByExternal(params.SourceSystem, params.ExternalUserID)
	if err == nil {
		return Response{}, statusCode(ErrUserAlreadyExists), ErrUserAlreadyExists
	}
	if !errors.Is(err, ErrUserNotFound) {
		return Response{}, 500, err
	}

	if _, err := s.repo.CreateUser(params); err != nil {
		if errors.Is(err, ErrUserAlreadyExists) {
			return Response{}, statusCode(err), err
		}
		return Response{}, 500, err
	}

	return Response{Success: true, Code: "REGISTERED", Message: "用户注册成功"}, 200, nil
}

// Login validates whitelist and device state, then issues a persisted session token.
func (s *Service) Login(req LoginRequest, clientIP, userAgent string) (Response, int, error) {
	if s == nil || s.repo == nil || s.sessions == nil {
		return Response{}, 503, fmt.Errorf("integration service unavailable")
	}

	sourceSystem, externalUserID, deviceID, err := validateLoginRequest(req)
	if err != nil {
		return Response{}, statusCode(err), err
	}

	user, err := s.repo.FindUserByExternal(sourceSystem, externalUserID)
	if err != nil {
		return Response{}, statusCode(err), err
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

	if isAdmin, err := s.repo.IsSystemAdmin(user.ID); err == nil {
		if isAdmin {
			response.Data = map[string]any{"role": "system_admin"}
		} else {
			response.Data = map[string]any{"role": "user"}
		}
	}

	return response, 200, nil
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
	}

	if !sourceSystemPattern.MatchString(params.SourceSystem) {
		return RegisterParams{}, ErrInvalidSourceSystem
	}
	if params.ExternalUserID == "" {
		return RegisterParams{}, ErrInvalidExternalUserID
	}
	if params.DisplayName == "" {
		return RegisterParams{}, ErrInvalidDisplayName
	}
	if params.Language == "" {
		params.Language = "zh-Hans"
	}

	return params, nil
}

func validateLoginRequest(req LoginRequest) (string, string, string, error) {
	sourceSystem := strings.TrimSpace(req.SourceSystem)
	externalUserID := strings.TrimSpace(req.ExternalUserID)
	deviceID := strings.TrimSpace(req.DeviceID)

	if !sourceSystemPattern.MatchString(sourceSystem) {
		return "", "", "", ErrInvalidSourceSystem
	}
	if externalUserID == "" {
		return "", "", "", ErrInvalidExternalUserID
	}
	if deviceID == "" {
		return "", "", "", ErrDeviceIDRequired
	}

	return sourceSystem, externalUserID, deviceID, nil
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
		errors.Is(err, ErrInvalidUserID),
		errors.Is(err, ErrDeviceIDRequired),
		errors.Is(err, ErrInvalidIPWhitelistRule),
		errors.Is(err, ErrWhitelistRulesRequired):
		return 400
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
