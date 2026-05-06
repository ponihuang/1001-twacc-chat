package erp

import (
	"time"

	"1001-twacc-chat/internal/auth"
)

// User stores the minimal fields required by external-system registration and login.
type User struct {
	ID              int64
	SourceSystem    string
	ExternalUserID  string
	DisplayName     string
	PasswordHash    string
	Email           string
	Language        string
	WhatsAppAccount string
	TelegramAccount string
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// UserSecuritySettings stores per-user login restriction flags.
type UserSecuritySettings struct {
	UserID      int64
	AllowAllIPs bool
}

// Device stores tracked device metadata.
type Device struct {
	ID              int64
	UserID          int64
	DeviceID        string
	UserAgent       string
	IP              string
	FirstLoginAt    time.Time
	LastLoginAt     time.Time
	IsTrustedDevice bool
}

// DeviceSummary is the API-facing device shape.
type DeviceSummary struct {
	DeviceID        string `json:"device_id"`
	UserAgent       string `json:"user_agent"`
	IP              string `json:"ip"`
	FirstLoginAt    string `json:"first_login_at"`
	LastLoginAt     string `json:"last_login_at"`
	IsTrustedDevice bool   `json:"is_trusted_device"`
}

// IPWhitelistSettings is the API-facing whitelist payload.
type IPWhitelistSettings struct {
	AllowAll bool     `json:"allow_all"`
	Rules    []string `json:"rules"`
}

// SessionPrincipal describes the authenticated actor from a session token.
type SessionPrincipal = auth.Session

// SessionIssuer creates a Bearer session token.
type SessionIssuer interface {
	Issue(userID int64, deviceID string) (string, time.Time, error)
}

// SessionAuthenticator validates a Bearer session token.
type SessionAuthenticator interface {
	Authenticate(token string) (auth.Session, error)
}

// RegisterRequest is the external-system registration payload.
type RegisterRequest struct {
	SourceSystem    string `json:"source_system"`
	ExternalUserID  string `json:"external_user_id"`
	Password        string `json:"password"`
	DisplayName     string `json:"display_name"`
	Email           string `json:"email"`
	Language        string `json:"language"`
	WhatsAppAccount string `json:"whatsapp_account"`
	TelegramAccount string `json:"telegram_account"`
}

// LoginRequest is the external-system login payload.
type LoginRequest struct {
	SourceSystem   string `json:"source_system"`
	ExternalUserID string `json:"external_user_id"`
	Password       string `json:"password"`
	DeviceID       string `json:"device_id"`
}

// IPWhitelistUpdateRequest is the system-admin whitelist payload.
type IPWhitelistUpdateRequest struct {
	AllowAll bool     `json:"allow_all"`
	Rules    []string `json:"rules"`
}

// Response is the API response shape for integration endpoints.
type Response struct {
	Success   bool   `json:"success"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	Token     string `json:"token,omitempty"`
	ExpiresAt string `json:"expires_at,omitempty"`
	Data      any    `json:"data,omitempty"`
}

// RegisterParams carries validated registration fields into storage.
type RegisterParams struct {
	SourceSystem    string
	ExternalUserID  string
	PasswordHash    string
	DisplayName     string
	Email           string
	Language        string
	WhatsAppAccount string
	TelegramAccount string
}

// Repository defines persistence required by external identity, login, and admin flows.
type Repository interface {
	CreateUser(params RegisterParams) (User, error)
	FindUserByID(userID int64) (User, error)
	FindUserByExternal(sourceSystem, externalUserID string) (User, error)
	IsSystemAdmin(userID int64) (bool, error)
	GetUserSecuritySettings(userID int64) (UserSecuritySettings, error)
	ListActiveIPWhitelistRules(userID int64) ([]string, error)
	ReplaceIPWhitelist(userID int64, allowAll bool, rules []string) error
	CountTrustedDevices(userID int64) (int, error)
	ListDevices(userID int64) ([]Device, error)
	FindDevice(userID int64, deviceID string) (Device, error)
	UpsertDeviceLogin(userID int64, deviceID, userAgent, ip string, trusted bool, now time.Time) error
	ApproveDevice(userID int64, deviceID string, trustedBy int64, now time.Time) error
}
