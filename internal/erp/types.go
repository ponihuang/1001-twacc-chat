package erp

import (
	"time"

	"1001-twacc-chat/internal/adminauth"
	"1001-twacc-chat/internal/auth"
)

// User stores the minimal fields required by external-system registration and login.
type User struct {
	ID                         int64
	SourceSystem               string
	ExternalUserID             string
	DisplayName                string
	PasswordHash               string
	Email                      string
	Language                   string
	WhatsAppAccount            string
	TelegramAccount            string
	Status                     string
	IsChatMuted                bool
	MustChangePassword         bool
	TemporaryPasswordExpiresAt *time.Time
	CreatedAt                  time.Time
	UpdatedAt                  time.Time
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

// AdminUserFilter contains supported filters for the admin user list.
type AdminUserFilter struct {
	ExternalUserID string
	DisplayName    string
	Email          string
	Status         string
	SourceSystem   string
	CreatedFrom    *time.Time
	CreatedTo      *time.Time
	Sort           string
	Page           int
	PerPage        int
}

// AdminUserSummary is the API-facing user shape for the admin console.
type AdminUserSummary struct {
	ID                         int64      `json:"id"`
	ExternalUserID             string     `json:"external_user_id"`
	DisplayName                string     `json:"display_name"`
	Email                      string     `json:"email"`
	Status                     string     `json:"status"`
	IsChatMuted                bool       `json:"is_chat_muted"`
	MustChangePassword         bool       `json:"must_change_password"`
	TemporaryPasswordExpiresAt *time.Time `json:"temporary_password_expires_at,omitempty"`
	SourceSystem               string     `json:"source_system"`
	CreatedAt                  time.Time  `json:"created_at"`
	LastOnlineAt               *time.Time `json:"last_online_at,omitempty"`
}

// AdminUserPage is the paginated response for the admin user list.
type AdminUserPage struct {
	Items      []AdminUserSummary `json:"items"`
	Total      int                `json:"total"`
	Page       int                `json:"page"`
	PerPage    int                `json:"per_page"`
	TotalPages int                `json:"total_pages"`
}

// AdminUserInvitationFilter contains supported filters for the invitation list.
type AdminUserInvitationFilter struct {
	Email   string
	Status  string
	Sorting string
	Page    int
	PerPage int
}

// UserInvitation stores a pending invitation for a user to complete registration.
type UserInvitation struct {
	ID               int64
	Email            string
	TokenHash        string
	Status           string
	InvitedByAdminID int64
	AcceptedUserID   int64
	ExpiresAt        time.Time
	SentAt           *time.Time
	AcceptedAt       *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// PublicUserInvitation is the API-facing shape for public invitation lookup.
type PublicUserInvitation struct {
	Email string `json:"email"`
}

// AdminUserInvitationSummary is the API-facing invitation shape for the admin console.
type AdminUserInvitationSummary struct {
	ID                    int64      `json:"id"`
	Email                 string     `json:"email"`
	Status                string     `json:"status"`
	InvitedByAdminID      int64      `json:"invited_by_admin_id,omitempty"`
	InvitedByAdminAccount string     `json:"invited_by_admin_account,omitempty"`
	InvitedByAdminName    string     `json:"invited_by_admin_name,omitempty"`
	AcceptedUserID        int64      `json:"accepted_user_id,omitempty"`
	ExpiresAt             time.Time  `json:"expires_at"`
	SentAt                *time.Time `json:"sent_at,omitempty"`
	AcceptedAt            *time.Time `json:"accepted_at,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

// AdminUserInvitationPage is the paginated response for the invitation list.
type AdminUserInvitationPage struct {
	Items      []AdminUserInvitationSummary `json:"items"`
	Total      int                          `json:"total"`
	Page       int                          `json:"page"`
	PerPage    int                          `json:"per_page"`
	TotalPages int                          `json:"total_pages"`
}

// BulkUserInvitationPrecheckItem describes one classified email from an uploaded TXT.
type BulkUserInvitationPrecheckItem struct {
	Line  int    `json:"line"`
	Email string `json:"email"`
}

// BulkUserInvitationInvalidItem describes one invalid TXT line.
type BulkUserInvitationInvalidItem struct {
	Line  int    `json:"line"`
	Value string `json:"value"`
}

// BulkUserInvitationPrecheckCounts contains category totals for the uploaded TXT.
type BulkUserInvitationPrecheckCounts struct {
	Sendable          int `json:"sendable"`
	DuplicateInFile   int `json:"duplicate_in_file"`
	AlreadyRegistered int `json:"already_registered"`
	ActiveInvitation  int `json:"active_invitation"`
	InvalidEmail      int `json:"invalid_email"`
}

// BulkUserInvitationPrecheckResult is the API-facing result for invitation TXT precheck.
type BulkUserInvitationPrecheckResult struct {
	TotalLines        int                              `json:"total_lines"`
	IgnoredBlankLines int                              `json:"ignored_blank_lines"`
	UniqueEmails      int                              `json:"unique_emails"`
	Counts            BulkUserInvitationPrecheckCounts `json:"counts"`
	Sendable          []BulkUserInvitationPrecheckItem `json:"sendable"`
	DuplicateInFile   []BulkUserInvitationPrecheckItem `json:"duplicate_in_file"`
	AlreadyRegistered []BulkUserInvitationPrecheckItem `json:"already_registered"`
	ActiveInvitation  []BulkUserInvitationPrecheckItem `json:"active_invitation"`
	InvalidEmail      []BulkUserInvitationInvalidItem  `json:"invalid_email"`
}

// BulkUserInvitationSendResult is the API-facing result for invitation TXT bulk send.
type BulkUserInvitationSendResult struct {
	SuccessCount int                              `json:"success_count"`
	FailureCount int                              `json:"failure_count"`
	IgnoredCount int                              `json:"ignored_count"`
	Succeeded    []BulkUserInvitationPrecheckItem `json:"succeeded"`
	Failed       []BulkUserInvitationSendFailure  `json:"failed"`
	Ignored      BulkUserInvitationIgnoredResult  `json:"ignored"`
	Precheck     BulkUserInvitationPrecheckResult `json:"precheck"`
}

// BulkUserInvitationSendFailure describes one email that failed during send.
type BulkUserInvitationSendFailure struct {
	Line    int    `json:"line"`
	Email   string `json:"email"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// BulkUserInvitationIgnoredResult groups emails skipped by bulk send.
type BulkUserInvitationIgnoredResult struct {
	DuplicateInFile   []BulkUserInvitationPrecheckItem `json:"duplicate_in_file"`
	AlreadyRegistered []BulkUserInvitationPrecheckItem `json:"already_registered"`
	ActiveInvitation  []BulkUserInvitationPrecheckItem `json:"active_invitation"`
	InvalidEmail      []BulkUserInvitationInvalidItem  `json:"invalid_email"`
}

// UserInvitationEmailPrecheck contains repository lookup results for a batch of emails.
type UserInvitationEmailPrecheck struct {
	Registered        map[string]bool
	ActiveInvitations map[string]bool
}

// SystemAdminFilter contains supported filters for the system-admin list.
type SystemAdminFilter struct {
	ExternalUserID string
	DisplayName    string
	Status         string
	Page           int
	PerPage        int
}

// SystemAdminSummary is the API-facing system-admin shape.
type SystemAdminSummary struct {
	ID             int64      `json:"id"`
	UserID         int64      `json:"user_id"`
	AdminUserID    int64      `json:"admin_user_id"`
	ExternalUserID string     `json:"external_user_id"`
	DisplayName    string     `json:"display_name"`
	PasswordHash   string     `json:"-"`
	Role           string     `json:"role"`
	RoleName       string     `json:"role_name"`
	Status         string     `json:"status"`
	LastLoginAt    *time.Time `json:"last_login_at,omitempty"`
	LastLoginIP    string     `json:"last_login_ip"`
}

// SystemAdminPage is the paginated response for the system-admin list.
type SystemAdminPage struct {
	Items      []SystemAdminSummary `json:"items"`
	Total      int                  `json:"total"`
	Page       int                  `json:"page"`
	PerPage    int                  `json:"per_page"`
	TotalPages int                  `json:"total_pages"`
}

// AdminRoleFilter contains supported filters for backend admin roles.
type AdminRoleFilter struct {
	Role    string
	Page    int
	PerPage int
}

// AdminRoleSummary is the API-facing backend admin role shape.
type AdminRoleSummary struct {
	ID         int64      `json:"id"`
	Code       string     `json:"code"`
	Name       string     `json:"name"`
	AdminCount int        `json:"admin_count"`
	Status     string     `json:"status"`
	CreatedAt  *time.Time `json:"created_at,omitempty"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty"`
}

// AdminRolePage is the paginated response for backend admin roles.
type AdminRolePage struct {
	Items      []AdminRoleSummary `json:"items"`
	Total      int                `json:"total"`
	Page       int                `json:"page"`
	PerPage    int                `json:"per_page"`
	TotalPages int                `json:"total_pages"`
}

// AdminPermissionItem is one editable permission switch for a backend role.
type AdminPermissionItem struct {
	Key      string `json:"key"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Enabled  bool   `json:"enabled"`
}

// AdminPermissionGroup groups permissions by the admin sidebar category.
type AdminPermissionGroup struct {
	Key   string                `json:"key"`
	Name  string                `json:"name"`
	Items []AdminPermissionItem `json:"items"`
}

// AdminRolePermissionDetail is the API-facing permission detail shape.
type AdminRolePermissionDetail struct {
	Role   AdminRoleSummary       `json:"role"`
	Groups []AdminPermissionGroup `json:"groups"`
}

// CurrentAdminPermissionResponse contains the current admin's enabled route permission keys.
type CurrentAdminPermissionResponse struct {
	Permissions []string `json:"permissions"`
}

// SystemSettings is the admin-facing chat system settings payload.
type SystemSettings struct {
	ChatAutoDeleteMaxDays         int        `json:"chat_auto_delete_max_days"`
	AdminChatHistoryRetentionDays int        `json:"admin_chat_history_retention_days"`
	UpdatedAt                     *time.Time `json:"updated_at,omitempty"`
}

// SystemSettingsUpdateRequest is the payload for updating admin system settings.
type SystemSettingsUpdateRequest struct {
	ChatAutoDeleteMaxDays         int `json:"chat_auto_delete_max_days"`
	AdminChatHistoryRetentionDays int `json:"admin_chat_history_retention_days"`
}

// AdminConversationFilter contains supported filters for the admin conversation list.
type AdminConversationFilter struct {
	Type             string
	Keyword          string
	LastActivityFrom *time.Time
	LastActivityTo   *time.Time
	Page             int
	PerPage          int
}

// AdminConversationSummary is the API-facing conversation shape for the admin console.
type AdminConversationSummary struct {
	ID                     int64      `json:"id"`
	Type                   string     `json:"type"`
	Name                   string     `json:"name"`
	MemberCount            int        `json:"member_count"`
	LastMessage            string     `json:"last_message"`
	LastMessageType        string     `json:"last_message_type"`
	LastMessageAt          *time.Time `json:"last_message_at,omitempty"`
	LastSenderID           int64      `json:"last_sender_id,omitempty"`
	LastSenderAccount      string     `json:"last_sender_account,omitempty"`
	LastSenderName         string     `json:"last_sender_name,omitempty"`
	LastActivityAt         time.Time  `json:"last_activity_at"`
	CreatedAt              time.Time  `json:"created_at"`
	CreatedByID            int64      `json:"created_by_id,omitempty"`
	CreatedByAccount       string     `json:"created_by_account,omitempty"`
	CreatedByName          string     `json:"created_by_name,omitempty"`
	LatestMessageAvailable bool       `json:"latest_message_available"`
}

// AdminConversationPage is the paginated response for the admin conversation list.
type AdminConversationPage struct {
	Items      []AdminConversationSummary `json:"items"`
	Total      int                        `json:"total"`
	Page       int                        `json:"page"`
	PerPage    int                        `json:"per_page"`
	TotalPages int                        `json:"total_pages"`
}

// AdminConversationMessageFilter contains supported filters for one admin read-only message page.
type AdminConversationMessageFilter struct {
	ConversationID int64
	Keyword        string
	SenderKeyword  string
	MessageType    string
	SentFrom       *time.Time
	SentTo         *time.Time
	MentionType    string
	Page           int
	PerPage        int
}

// AdminConversationParticipant is a read-only participant summary for admin conversation records.
type AdminConversationParticipant struct {
	UserID      int64  `json:"user_id"`
	Account     string `json:"account"`
	DisplayName string `json:"display_name"`
}

// AdminConversationAttachment is a read-only attachment summary for admin conversation records.
type AdminConversationAttachment struct {
	ID           int64  `json:"id"`
	OriginalName string `json:"original_name"`
	StoragePath  string `json:"storage_path"`
	MIMEType     string `json:"mime_type"`
	SizeBytes    int64  `json:"size_bytes"`
	URL          string `json:"url"`
}

// AdminConversationMention is a read-only mention summary for admin conversation records.
type AdminConversationMention struct {
	UserID      int64  `json:"user_id"`
	Account     string `json:"account"`
	DisplayName string `json:"display_name"`
	MentionType string `json:"mention_type"`
}

// AdminConversationMessage is one read-only message row for admin conversation records.
type AdminConversationMessage struct {
	ID            int64                         `json:"id"`
	SentAt        time.Time                     `json:"sent_at"`
	SenderID      int64                         `json:"sender_id"`
	SenderAccount string                        `json:"sender_account"`
	SenderName    string                        `json:"sender_name"`
	MessageType   string                        `json:"message_type"`
	Content       string                        `json:"content"`
	Status        string                        `json:"status"`
	IsRecalled    bool                          `json:"is_recalled"`
	Attachments   []AdminConversationAttachment `json:"attachments"`
	Mentions      []AdminConversationMention    `json:"mentions"`
}

// AdminConversationMessagePage is the paginated message result for admin conversation records.
type AdminConversationMessagePage struct {
	Items      []AdminConversationMessage `json:"items"`
	Total      int                        `json:"total"`
	Page       int                        `json:"page"`
	PerPage    int                        `json:"per_page"`
	TotalPages int                        `json:"total_pages"`
}

// AdminConversationDetail is the read-only admin detail payload for one conversation.
type AdminConversationDetail struct {
	Conversation AdminConversationSummary       `json:"conversation"`
	Participants []AdminConversationParticipant `json:"participants"`
	Messages     AdminConversationMessagePage   `json:"messages"`
}

// SystemAdminBootstrapResult describes an idempotent first-admin bootstrap run.
type SystemAdminBootstrapResult struct {
	Admin        SystemAdminSummary
	AdminCreated bool
}

// IPWhitelistSettings is the API-facing whitelist payload.
type IPWhitelistSettings struct {
	AllowAll bool     `json:"allow_all"`
	Rules    []string `json:"rules"`
}

// SessionPrincipal describes the authenticated actor from a session token.
type SessionPrincipal = auth.Session

// AdminSessionPrincipal describes the authenticated admin actor from an admin session token.
type AdminSessionPrincipal = adminauth.Session

// SessionIssuer creates a Bearer session token.
type SessionIssuer interface {
	Issue(userID int64, deviceID string) (string, time.Time, error)
}

// AdminSessionIssuer creates a Bearer session token for backend admins.
type AdminSessionIssuer interface {
	Issue(adminUserID int64, deviceID string) (string, time.Time, error)
}

// InvitationMailer sends user registration invitations.
type InvitationMailer interface {
	SendUserInvitation(email, inviteURL string, expiresAt time.Time) error
	SendTemporaryPassword(email, temporaryPassword string, expiresAt time.Time) error
}

// SessionAuthenticator validates a Bearer session token.
type SessionAuthenticator interface {
	Authenticate(token string) (auth.Session, error)
}

// AdminSessionAuthenticator validates a Bearer admin session token.
type AdminSessionAuthenticator interface {
	Authenticate(token string) (adminauth.Session, error)
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

// AdminCreateUserRequest is the system-admin payload for creating a user.
type AdminCreateUserRequest struct {
	SourceSystem   string `json:"source_system"`
	ExternalUserID string `json:"external_user_id"`
	Password       string `json:"password"`
	DisplayName    string `json:"display_name"`
	Email          string `json:"email"`
	Language       string `json:"language"`
	Status         string `json:"status"`
}

// AdminInviteUserRequest is the system-admin payload for inviting a user by email.
type AdminInviteUserRequest struct {
	Email string `json:"email"`
}

// AcceptUserInvitationRequest is the public payload for completing invited registration.
type AcceptUserInvitationRequest struct {
	Token                string `json:"token"`
	Account              string `json:"account"`
	Nickname             string `json:"nickname"`
	Password             string `json:"password"`
	PasswordConfirmation string `json:"password_confirmation"`
	DeviceID             string `json:"device_id"`
}

// SystemAdminCreateRequest is the payload for creating an office system admin.
type SystemAdminCreateRequest struct {
	ExternalUserID string `json:"external_user_id"`
	Password       string `json:"password"`
	DisplayName    string `json:"display_name"`
	Role           string `json:"role"`
	Status         string `json:"status"`
}

// AdminRoleCreateRequest is the payload for creating or updating a backend admin role.
type AdminRoleCreateRequest struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// AdminRolePermissionSetting is one role permission setting submitted by the admin UI.
type AdminRolePermissionSetting struct {
	Key     string `json:"key"`
	Enabled bool   `json:"enabled"`
}

// AdminRolePermissionUpdateRequest is the payload for replacing role permissions.
type AdminRolePermissionUpdateRequest struct {
	Permissions []AdminRolePermissionSetting `json:"permissions"`
}

// AdminUpdateUserRequest is the system-admin payload for updating a user.
type AdminUpdateUserRequest struct {
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Status      string `json:"status"`
}

// AdminUpdateUserChatMuteRequest is the system-admin payload for muting a user in chat.
type AdminUpdateUserChatMuteRequest struct {
	IsChatMuted bool `json:"is_chat_muted"`
}

// PasswordUpdateRequest is the authenticated user payload for changing password.
type PasswordUpdateRequest struct {
	CurrentPassword      string `json:"current_password"`
	Password             string `json:"password"`
	PasswordConfirmation string `json:"password_confirmation"`
}

// LoginRequest is the external-system login payload.
type LoginRequest struct {
	SourceSystem   string `json:"source_system"`
	ExternalUserID string `json:"external_user_id"`
	Password       string `json:"password"`
	DeviceID       string `json:"device_id"`
}

// AdminLoginRequest is the backend admin login payload.
type AdminLoginRequest struct {
	Account  string `json:"account"`
	Password string `json:"password"`
	DeviceID string `json:"device_id"`
}

// ProfileUpdateRequest is the current-user profile update payload.
type ProfileUpdateRequest struct {
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
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
	Status          string
}

// SystemAdminCreateParams carries validated backend admin fields into storage.
type SystemAdminCreateParams struct {
	Account      string
	PasswordHash string
	DisplayName  string
	Role         string
	Status       string
	CreatedBy    int64
}

// SystemAdminUpdateParams carries mutable backend admin edits into storage.
type SystemAdminUpdateParams struct {
	PasswordHash string
	DisplayName  string
	Role         string
	Status       string
}

// AdminRoleCreateParams carries validated backend admin role fields into storage.
type AdminRoleCreateParams struct {
	Code      string
	Name      string
	Status    string
	CreatedBy int64
}

// AdminRoleUpdateParams carries mutable backend admin role fields into storage.
type AdminRoleUpdateParams struct {
	Name   string
	Status string
}

// AdminUpdateUserParams carries validated admin edits into storage.
type AdminUpdateUserParams struct {
	PasswordHash string
	DisplayName  string
	Email        string
	Status       string
}

// AdminUpdateUserChatMuteParams carries the admin chat mute flag into storage.
type AdminUpdateUserChatMuteParams struct {
	IsChatMuted bool
}

// TemporaryPasswordParams carries a generated temporary password into storage.
type TemporaryPasswordParams struct {
	PasswordHash string
	ExpiresAt    time.Time
}

// UserInvitationCreateParams carries validated invitation fields into storage.
type UserInvitationCreateParams struct {
	Email            string
	TokenHash        string
	Status           string
	InvitedByAdminID int64
	ExpiresAt        time.Time
	SentAt           *time.Time
}

// UserInvitationRefreshParams carries a new token for an existing unsent invitation.
type UserInvitationRefreshParams struct {
	ID               int64
	TokenHash        string
	InvitedByAdminID int64
	ExpiresAt        time.Time
}

// AcceptUserInvitationParams carries validated invited registration fields into storage.
type AcceptUserInvitationParams struct {
	TokenHash      string
	SourceSystem   string
	ExternalUserID string
	PasswordHash   string
	DisplayName    string
	Status         string
	AcceptedAt     time.Time
}

// Repository defines persistence required by external identity, login, and admin flows.
type Repository interface {
	CreateUser(params RegisterParams) (User, error)
	FindUserByID(userID int64) (User, error)
	FindUserByExternal(sourceSystem, externalUserID string) (User, error)
	FindUserByExternalID(externalUserID string) (User, error)
	FindUserByEmail(email string) (User, error)
	ListUsers(filter AdminUserFilter) (AdminUserPage, error)
	ListUserInvitations(filter AdminUserInvitationFilter) (AdminUserInvitationPage, error)
	PrecheckUserInvitationEmails(emails []string, now time.Time) (UserInvitationEmailPrecheck, error)
	FindUserInvitationByEmail(email string) (UserInvitation, error)
	FindUserInvitationByTokenHash(tokenHash string) (UserInvitation, error)
	CreateUserInvitation(params UserInvitationCreateParams) (UserInvitation, error)
	RefreshUserInvitation(params UserInvitationRefreshParams) (UserInvitation, error)
	MarkUserInvitationSent(invitationID int64, sentAt time.Time) error
	AcceptUserInvitation(params AcceptUserInvitationParams) (User, error)
	ListSystemAdmins(filter SystemAdminFilter) (SystemAdminPage, error)
	ListAdminRoles(filter AdminRoleFilter) (AdminRolePage, error)
	FindAdminRoleByID(roleID int64) (AdminRoleSummary, error)
	FindAdminRoleByCode(code string) (AdminRoleSummary, error)
	CreateAdminRole(params AdminRoleCreateParams) (AdminRoleSummary, error)
	UpdateAdminRole(roleID int64, params AdminRoleUpdateParams) (AdminRoleSummary, error)
	ListAdminRolePermissions(roleID int64) (map[string]bool, error)
	ReplaceAdminRolePermissions(roleID int64, permissions map[string]bool) error
	GetSystemSettings() (SystemSettings, error)
	UpdateSystemSettings(settings SystemSettings, adminUserID int64) (SystemSettings, error)
	PurgeExpiredAdminChatMessages(now time.Time) (int64, error)
	ListAdminConversations(filter AdminConversationFilter) (AdminConversationPage, error)
	GetAdminConversationDetail(filter AdminConversationMessageFilter) (AdminConversationDetail, error)
	FindSystemAdminByID(adminUserID int64) (SystemAdminSummary, error)
	FindSystemAdminByAccount(account string) (SystemAdminSummary, error)
	CreateSystemAdmin(params SystemAdminCreateParams) (SystemAdminSummary, error)
	UpdateSystemAdmin(adminUserID int64, params SystemAdminUpdateParams) (SystemAdminSummary, error)
	UpdateSystemAdminLastLogin(adminUserID int64, ip string, at time.Time) error
	UpdateUser(userID int64, params AdminUpdateUserParams) (User, error)
	UpdateUserChatMute(userID int64, params AdminUpdateUserChatMuteParams) (User, error)
	UpdateUserTemporaryPassword(userID int64, params TemporaryPasswordParams) (User, error)
	UpdateUserPassword(userID int64, passwordHash string) (User, error)
	UpdateUserProfile(userID int64, displayName string, email string) (User, error)
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
