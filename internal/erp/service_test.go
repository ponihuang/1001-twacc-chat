package erp

import (
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"
)

type mockRepository struct {
	usersByID           map[int64]User
	usersByExternal     map[string]User
	systemAdmins        map[int64]bool
	adminsByID          map[int64]SystemAdminSummary
	adminsByAccount     map[string]SystemAdminSummary
	settingsByUserID    map[int64]UserSecuritySettings
	whitelistRules      map[int64][]string
	devicesByKey        map[string]Device
	invitationsByEmail  map[string]UserInvitation
	replaceAllowAll     bool
	replaceRules        []string
	replaceCalled       bool
	approveCalled       bool
	approvedTrustedBy   int64
	approvedUserID      int64
	approvedDeviceID    string
	trustedDeviceCount  map[int64]int
	createdAdmin        SystemAdminSummary
	createdInvitation   UserInvitation
	refreshedInvitation UserInvitation
	sentInvitationID    int64
	sentAt              time.Time
}

type mockSessions struct {
	issuedUserID   int64
	issuedDeviceID string
	token          string
	expiresAt      time.Time
}

type mockInvitationMailer struct {
	email             string
	inviteURL         string
	temporaryPassword string
	expiresAt         time.Time
	err               error
}

func (m *mockInvitationMailer) SendUserInvitation(email, inviteURL string, expiresAt time.Time) error {
	m.email = email
	m.inviteURL = inviteURL
	m.expiresAt = expiresAt
	return m.err
}

func (m *mockInvitationMailer) SendTemporaryPassword(email, temporaryPassword string, expiresAt time.Time) error {
	m.email = email
	m.temporaryPassword = temporaryPassword
	m.expiresAt = expiresAt
	return m.err
}

func (m *mockRepository) CreateUser(params RegisterParams) (User, error) {
	key := params.SourceSystem + ":" + params.ExternalUserID
	if _, err := m.FindUserByExternalID(params.ExternalUserID); err == nil {
		return User{}, ErrUserAlreadyExists
	} else if !errors.Is(err, ErrUserNotFound) {
		return User{}, err
	}
	user := User{ID: int64(len(m.usersByID) + 1), SourceSystem: params.SourceSystem, ExternalUserID: params.ExternalUserID, DisplayName: params.DisplayName, PasswordHash: params.PasswordHash, Email: params.Email, Language: params.Language, Status: params.Status}
	if m.usersByID == nil {
		m.usersByID = map[int64]User{}
	}
	if m.usersByExternal == nil {
		m.usersByExternal = map[string]User{}
	}
	m.usersByID[user.ID] = user
	m.usersByExternal[key] = user
	return user, nil
}

func (m *mockRepository) FindUserByID(userID int64) (User, error) {
	user, ok := m.usersByID[userID]
	if !ok {
		return User{}, ErrUserNotFound
	}
	return user, nil
}

func (m *mockRepository) FindUserByExternal(sourceSystem, externalUserID string) (User, error) {
	user, ok := m.usersByExternal[sourceSystem+":"+externalUserID]
	if !ok {
		return User{}, ErrUserNotFound
	}
	return user, nil
}

func (m *mockRepository) FindUserByExternalID(externalUserID string) (User, error) {
	for _, user := range m.usersByExternal {
		if user.ExternalUserID == externalUserID {
			return user, nil
		}
	}
	return User{}, ErrUserNotFound
}

func (m *mockRepository) FindUserByEmail(email string) (User, error) {
	for _, user := range m.usersByID {
		if strings.EqualFold(user.Email, strings.TrimSpace(email)) {
			return user, nil
		}
	}
	return User{}, ErrUserNotFound
}

func (m *mockRepository) FindUserInvitationByEmail(email string) (UserInvitation, error) {
	if m.invitationsByEmail == nil {
		return UserInvitation{}, ErrUserNotFound
	}
	invitation, ok := m.invitationsByEmail[strings.ToLower(strings.TrimSpace(email))]
	if !ok {
		return UserInvitation{}, ErrUserNotFound
	}
	return invitation, nil
}

func (m *mockRepository) FindUserInvitationByTokenHash(tokenHash string) (UserInvitation, error) {
	if m.invitationsByEmail == nil {
		return UserInvitation{}, ErrUserNotFound
	}
	for _, invitation := range m.invitationsByEmail {
		if invitation.TokenHash == strings.TrimSpace(tokenHash) {
			return invitation, nil
		}
	}
	return UserInvitation{}, ErrUserNotFound
}

func (m *mockRepository) CreateUserInvitation(params UserInvitationCreateParams) (UserInvitation, error) {
	if m.invitationsByEmail == nil {
		m.invitationsByEmail = map[string]UserInvitation{}
	}
	key := strings.ToLower(strings.TrimSpace(params.Email))
	if _, ok := m.invitationsByEmail[key]; ok {
		return UserInvitation{}, ErrUserInvitationPending
	}
	now := time.Now().UTC()
	m.createdInvitation = UserInvitation{
		ID:               int64(len(m.invitationsByEmail) + 1),
		Email:            params.Email,
		TokenHash:        params.TokenHash,
		Status:           params.Status,
		InvitedByAdminID: params.InvitedByAdminID,
		ExpiresAt:        params.ExpiresAt,
		SentAt:           params.SentAt,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	m.invitationsByEmail[key] = m.createdInvitation
	return m.createdInvitation, nil
}

func (m *mockRepository) RefreshUserInvitation(params UserInvitationRefreshParams) (UserInvitation, error) {
	for key, invitation := range m.invitationsByEmail {
		if invitation.ID != params.ID {
			continue
		}
		invitation.TokenHash = params.TokenHash
		invitation.Status = "pending"
		invitation.InvitedByAdminID = params.InvitedByAdminID
		invitation.ExpiresAt = params.ExpiresAt
		invitation.SentAt = nil
		invitation.AcceptedAt = nil
		invitation.AcceptedUserID = 0
		invitation.UpdatedAt = time.Now().UTC()
		m.refreshedInvitation = invitation
		m.invitationsByEmail[key] = invitation
		return invitation, nil
	}
	return UserInvitation{}, ErrUserNotFound
}

func (m *mockRepository) MarkUserInvitationSent(invitationID int64, sentAt time.Time) error {
	m.sentInvitationID = invitationID
	m.sentAt = sentAt
	for key, invitation := range m.invitationsByEmail {
		if invitation.ID == invitationID {
			invitation.SentAt = &sentAt
			m.invitationsByEmail[key] = invitation
			return nil
		}
	}
	return ErrUserNotFound
}

func (m *mockRepository) AcceptUserInvitation(params AcceptUserInvitationParams) (User, error) {
	for key, invitation := range m.invitationsByEmail {
		if invitation.TokenHash != params.TokenHash {
			continue
		}
		if invitation.Status != "pending" || !params.AcceptedAt.Before(invitation.ExpiresAt) || invitation.AcceptedAt != nil || invitation.AcceptedUserID > 0 {
			return User{}, ErrUserInvitationNotFound
		}
		user := User{
			ID:             int64(len(m.usersByID) + 1),
			SourceSystem:   params.SourceSystem,
			ExternalUserID: params.ExternalUserID,
			DisplayName:    params.DisplayName,
			PasswordHash:   params.PasswordHash,
			Email:          invitation.Email,
			Language:       "zh-Hans",
			Status:         params.Status,
			CreatedAt:      params.AcceptedAt,
			UpdatedAt:      params.AcceptedAt,
		}
		if m.usersByID == nil {
			m.usersByID = map[int64]User{}
		}
		if m.usersByExternal == nil {
			m.usersByExternal = map[string]User{}
		}
		m.usersByID[user.ID] = user
		m.usersByExternal[user.SourceSystem+":"+user.ExternalUserID] = user
		invitation.Status = "accepted"
		invitation.AcceptedUserID = user.ID
		invitation.AcceptedAt = &params.AcceptedAt
		m.invitationsByEmail[key] = invitation
		return user, nil
	}
	return User{}, ErrUserInvitationNotFound
}

func (m *mockRepository) ListUsers(filter AdminUserFilter) (AdminUserPage, error) {
	users := make([]AdminUserSummary, 0, len(m.usersByID))
	for _, user := range m.usersByID {
		users = append(users, AdminUserSummary{
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
		})
	}
	return AdminUserPage{
		Items:      users,
		Total:      len(users),
		Page:       filter.Page,
		PerPage:    filter.PerPage,
		TotalPages: 1,
	}, nil
}

func (m *mockRepository) ListUserInvitations(filter AdminUserInvitationFilter) (AdminUserInvitationPage, error) {
	invitations := make([]AdminUserInvitationSummary, 0, len(m.invitationsByEmail))
	for _, invitation := range m.invitationsByEmail {
		invitations = append(invitations, AdminUserInvitationSummary{
			ID:               invitation.ID,
			Email:            invitation.Email,
			Status:           invitation.Status,
			InvitedByAdminID: invitation.InvitedByAdminID,
			AcceptedUserID:   invitation.AcceptedUserID,
			ExpiresAt:        invitation.ExpiresAt,
			SentAt:           invitation.SentAt,
			AcceptedAt:       invitation.AcceptedAt,
			CreatedAt:        invitation.CreatedAt,
			UpdatedAt:        invitation.UpdatedAt,
		})
	}
	return AdminUserInvitationPage{
		Items:      invitations,
		Total:      len(invitations),
		Page:       filter.Page,
		PerPage:    filter.PerPage,
		TotalPages: 1,
	}, nil
}

func (m *mockRepository) PrecheckUserInvitationEmails(emails []string, now time.Time) (UserInvitationEmailPrecheck, error) {
	result := UserInvitationEmailPrecheck{
		Registered:        map[string]bool{},
		ActiveInvitations: map[string]bool{},
	}
	for _, email := range emails {
		normalized := strings.ToLower(strings.TrimSpace(email))
		if _, err := m.FindUserByEmail(normalized); err == nil {
			result.Registered[normalized] = true
		}
		if invitation, err := m.FindUserInvitationByEmail(normalized); err == nil &&
			invitation.Status == "pending" &&
			now.Before(invitation.ExpiresAt) &&
			invitation.AcceptedAt == nil &&
			invitation.AcceptedUserID == 0 {
			result.ActiveInvitations[normalized] = true
		}
	}
	return result, nil
}

func (m *mockRepository) ListSystemAdmins(filter SystemAdminFilter) (SystemAdminPage, error) {
	admins := make([]SystemAdminSummary, 0)
	for _, admin := range m.adminsByID {
		admins = append(admins, admin)
	}
	if len(admins) == 0 {
		for userID := range m.systemAdmins {
			admins = append(admins, SystemAdminSummary{
				ID:             userID,
				UserID:         userID,
				AdminUserID:    userID,
				ExternalUserID: "admin",
				DisplayName:    "Admin",
				Role:           "system_admin",
				RoleName:       "系統管理員",
				Status:         "active",
			})
		}
	}
	return SystemAdminPage{
		Items:      admins,
		Total:      len(admins),
		Page:       filter.Page,
		PerPage:    filter.PerPage,
		TotalPages: 1,
	}, nil
}

func (m *mockRepository) ListAdminRoles(filter AdminRoleFilter) (AdminRolePage, error) {
	return AdminRolePage{
		Items:      []AdminRoleSummary{},
		Total:      0,
		Page:       filter.Page,
		PerPage:    filter.PerPage,
		TotalPages: 0,
	}, nil
}

func (m *mockRepository) FindAdminRoleByID(roleID int64) (AdminRoleSummary, error) {
	return AdminRoleSummary{}, ErrUserNotFound
}

func (m *mockRepository) FindAdminRoleByCode(code string) (AdminRoleSummary, error) {
	return AdminRoleSummary{}, ErrUserNotFound
}

func (m *mockRepository) CreateAdminRole(params AdminRoleCreateParams) (AdminRoleSummary, error) {
	return AdminRoleSummary{ID: 1, Code: params.Code, Name: params.Name, Status: params.Status}, nil
}

func (m *mockRepository) UpdateAdminRole(roleID int64, params AdminRoleUpdateParams) (AdminRoleSummary, error) {
	return AdminRoleSummary{ID: roleID, Code: "system_admin", Name: params.Name, Status: params.Status}, nil
}

func (m *mockRepository) ListAdminConversations(filter AdminConversationFilter) (AdminConversationPage, error) {
	return AdminConversationPage{
		Items:      []AdminConversationSummary{},
		Total:      0,
		Page:       filter.Page,
		PerPage:    filter.PerPage,
		TotalPages: 0,
	}, nil
}

func (m *mockRepository) GetAdminConversationDetail(filter AdminConversationMessageFilter) (AdminConversationDetail, error) {
	return AdminConversationDetail{
		Messages: AdminConversationMessagePage{
			Items:      []AdminConversationMessage{},
			Total:      0,
			Page:       filter.Page,
			PerPage:    filter.PerPage,
			TotalPages: 0,
		},
	}, nil
}

func (m *mockRepository) FindSystemAdminByID(adminUserID int64) (SystemAdminSummary, error) {
	if admin, ok := m.adminsByID[adminUserID]; ok {
		return admin, nil
	}
	if m.systemAdmins[adminUserID] {
		return SystemAdminSummary{ID: adminUserID, UserID: adminUserID, AdminUserID: adminUserID, ExternalUserID: "admin", DisplayName: "Admin", PasswordHash: "pbkdf2-sha256:120000:bad:bad", Role: "system_admin", RoleName: "系統管理員", Status: "active"}, nil
	}
	return SystemAdminSummary{}, ErrUserNotFound
}

func (m *mockRepository) FindSystemAdminByAccount(account string) (SystemAdminSummary, error) {
	if admin, ok := m.adminsByAccount[account]; ok {
		return admin, nil
	}
	return SystemAdminSummary{}, ErrUserNotFound
}

func (m *mockRepository) CreateSystemAdmin(params SystemAdminCreateParams) (SystemAdminSummary, error) {
	if m.adminsByID == nil {
		m.adminsByID = map[int64]SystemAdminSummary{}
	}
	if m.adminsByAccount == nil {
		m.adminsByAccount = map[string]SystemAdminSummary{}
	}
	if _, ok := m.adminsByAccount[params.Account]; ok {
		return SystemAdminSummary{}, ErrUserAlreadyExists
	}
	adminID := int64(len(m.adminsByID) + 1)
	m.createdAdmin = SystemAdminSummary{
		ID:             adminID,
		UserID:         adminID,
		AdminUserID:    adminID,
		ExternalUserID: params.Account,
		DisplayName:    params.DisplayName,
		PasswordHash:   params.PasswordHash,
		Role:           params.Role,
		RoleName:       "系統管理員",
		Status:         params.Status,
	}
	m.adminsByID[adminID] = m.createdAdmin
	m.adminsByAccount[params.Account] = m.createdAdmin
	return m.createdAdmin, nil
}

func (m *mockRepository) UpdateSystemAdmin(adminUserID int64, params SystemAdminUpdateParams) (SystemAdminSummary, error) {
	admin, ok := m.adminsByID[adminUserID]
	if !ok {
		return SystemAdminSummary{}, ErrUserNotFound
	}
	admin.DisplayName = params.DisplayName
	admin.Role = params.Role
	admin.RoleName = "系統管理員"
	admin.Status = params.Status
	if params.PasswordHash != "" {
		admin.PasswordHash = params.PasswordHash
	}
	m.adminsByID[adminUserID] = admin
	m.adminsByAccount[admin.ExternalUserID] = admin
	return admin, nil
}

func (m *mockRepository) UpdateSystemAdminLastLogin(adminUserID int64, ip string, at time.Time) error {
	admin, ok := m.adminsByID[adminUserID]
	if !ok {
		return ErrUserNotFound
	}
	admin.LastLoginAt = &at
	admin.LastLoginIP = ip
	m.adminsByID[adminUserID] = admin
	m.adminsByAccount[admin.ExternalUserID] = admin
	return nil
}

func (m *mockRepository) UpdateUser(userID int64, params AdminUpdateUserParams) (User, error) {
	user, ok := m.usersByID[userID]
	if !ok {
		return User{}, ErrUserNotFound
	}
	user.DisplayName = params.DisplayName
	user.Email = params.Email
	user.Status = params.Status
	if params.PasswordHash != "" {
		user.PasswordHash = params.PasswordHash
	}
	m.usersByID[userID] = user
	m.usersByExternal[user.SourceSystem+":"+user.ExternalUserID] = user
	return user, nil
}

func (m *mockRepository) UpdateUserChatMute(userID int64, params AdminUpdateUserChatMuteParams) (User, error) {
	user, ok := m.usersByID[userID]
	if !ok {
		return User{}, ErrUserNotFound
	}
	user.IsChatMuted = params.IsChatMuted
	m.usersByID[userID] = user
	m.usersByExternal[user.SourceSystem+":"+user.ExternalUserID] = user
	return user, nil
}

func (m *mockRepository) UpdateUserTemporaryPassword(userID int64, params TemporaryPasswordParams) (User, error) {
	user, ok := m.usersByID[userID]
	if !ok {
		return User{}, ErrUserNotFound
	}
	user.PasswordHash = params.PasswordHash
	user.MustChangePassword = true
	user.TemporaryPasswordExpiresAt = &params.ExpiresAt
	m.usersByID[userID] = user
	m.usersByExternal[user.SourceSystem+":"+user.ExternalUserID] = user
	return user, nil
}

func (m *mockRepository) UpdateUserPassword(userID int64, passwordHash string) (User, error) {
	user, ok := m.usersByID[userID]
	if !ok {
		return User{}, ErrUserNotFound
	}
	user.PasswordHash = passwordHash
	user.MustChangePassword = false
	user.TemporaryPasswordExpiresAt = nil
	m.usersByID[userID] = user
	m.usersByExternal[user.SourceSystem+":"+user.ExternalUserID] = user
	return user, nil
}

func (m *mockRepository) UpdateUserProfile(userID int64, displayName string, email string) (User, error) {
	user, ok := m.usersByID[userID]
	if !ok {
		return User{}, ErrUserNotFound
	}
	user.DisplayName = displayName
	user.Email = email
	m.usersByID[userID] = user
	m.usersByExternal[user.SourceSystem+":"+user.ExternalUserID] = user
	return user, nil
}

func (m *mockRepository) IsSystemAdmin(userID int64) (bool, error) {
	return m.systemAdmins[userID], nil
}

func (m *mockRepository) GetUserSecuritySettings(userID int64) (UserSecuritySettings, error) {
	if settings, ok := m.settingsByUserID[userID]; ok {
		return settings, nil
	}
	return UserSecuritySettings{UserID: userID, AllowAllIPs: true}, nil
}

func (m *mockRepository) ListActiveIPWhitelistRules(userID int64) ([]string, error) {
	return m.whitelistRules[userID], nil
}

func (m *mockRepository) ReplaceIPWhitelist(userID int64, allowAll bool, rules []string) error {
	m.replaceCalled = true
	m.replaceAllowAll = allowAll
	m.replaceRules = append([]string(nil), rules...)
	return nil
}

func (m *mockRepository) CountTrustedDevices(userID int64) (int, error) {
	return m.trustedDeviceCount[userID], nil
}

func (m *mockRepository) ListDevices(userID int64) ([]Device, error) {
	var devices []Device
	for _, device := range m.devicesByKey {
		if device.UserID == userID {
			devices = append(devices, device)
		}
	}
	return devices, nil
}

func (m *mockRepository) FindDevice(userID int64, deviceID string) (Device, error) {
	device, ok := m.devicesByKey[deviceKey(userID, deviceID)]
	if !ok {
		return Device{}, ErrDeviceNotFound
	}
	return device, nil
}

func (m *mockRepository) UpsertDeviceLogin(userID int64, deviceID, userAgent, ip string, trusted bool, now time.Time) error {
	if m.devicesByKey == nil {
		m.devicesByKey = map[string]Device{}
	}
	key := deviceKey(userID, deviceID)
	device := m.devicesByKey[key]
	if device.DeviceID == "" {
		device = Device{UserID: userID, DeviceID: deviceID, FirstLoginAt: now}
	}
	device.UserAgent = userAgent
	device.IP = ip
	device.LastLoginAt = now
	device.IsTrustedDevice = device.IsTrustedDevice || trusted
	m.devicesByKey[key] = device
	return nil
}

func (m *mockRepository) ApproveDevice(userID int64, deviceID string, trustedBy int64, now time.Time) error {
	device, ok := m.devicesByKey[deviceKey(userID, deviceID)]
	if !ok {
		return ErrDeviceNotFound
	}
	device.IsTrustedDevice = true
	m.devicesByKey[deviceKey(userID, deviceID)] = device
	m.approveCalled = true
	m.approvedTrustedBy = trustedBy
	m.approvedUserID = userID
	m.approvedDeviceID = deviceID
	return nil
}

func (m *mockSessions) Issue(userID int64, deviceID string) (string, time.Time, error) {
	m.issuedUserID = userID
	m.issuedDeviceID = deviceID
	if m.token == "" {
		m.token = "session-token"
	}
	if m.expiresAt.IsZero() {
		m.expiresAt = time.Now().UTC().Add(time.Hour)
	}
	return m.token, m.expiresAt, nil
}

func deviceKey(userID int64, deviceID string) string {
	return strconv.FormatInt(userID, 10) + ":" + deviceID
}

func TestIPAllowed(t *testing.T) {
	tests := []struct {
		name     string
		clientIP string
		rules    []string
		want     bool
	}{
		{name: "single ipv4", clientIP: "192.168.1.10", rules: []string{"192.168.1.10"}, want: true},
		{name: "cidr ipv4", clientIP: "10.0.0.18", rules: []string{"10.0.0.0/24"}, want: true},
		{name: "single ipv6", clientIP: "2001:db8::10", rules: []string{"2001:db8::10"}, want: true},
		{name: "cidr ipv6", clientIP: "2001:db8::12", rules: []string{"2001:db8::/64"}, want: true},
		{name: "not allowed", clientIP: "172.16.1.10", rules: []string{"192.168.1.10", "10.0.0.0/24"}, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ipAllowed(tc.clientIP, tc.rules); got != tc.want {
				t.Fatalf("ipAllowed(%q, %v) = %v, want %v", tc.clientIP, tc.rules, got, tc.want)
			}
		})
	}
}

func TestValidateRegisterRequest(t *testing.T) {
	params, err := validateRegisterRequest(RegisterRequest{SourceSystem: "erp", ExternalUserID: "A12345", Password: "pass123!", DisplayName: "测试用户"})
	if err != nil {
		t.Fatalf("validateRegisterRequest returned error: %v", err)
	}
	if params.Language != "zh-Hans" {
		t.Fatalf("default language = %q, want zh-Hans", params.Language)
	}
}

func TestValidateLoginRequest(t *testing.T) {
	_, _, _, _, err := validateLoginRequest(LoginRequest{SourceSystem: "erp", ExternalUserID: "A12345", Password: "pass123!", DeviceID: "device-1"})
	if err != nil {
		t.Fatalf("validateLoginRequest returned error: %v", err)
	}

	_, _, _, _, err = validateLoginRequest(LoginRequest{SourceSystem: "ERP", ExternalUserID: "A12345", Password: "pass123!", DeviceID: "device-1"})
	if err != ErrInvalidSourceSystem {
		t.Fatalf("unexpected error for invalid source system: %v", err)
	}
}

func TestLoginIssuesSessionToken(t *testing.T) {
	repo := &mockRepository{
		usersByExternal:    map[string]User{"erp:user-1": {ID: 1, SourceSystem: "erp", ExternalUserID: "user-1", DisplayName: "User One", PasswordHash: mustHashPassword(t, "pass123!")}},
		settingsByUserID:   map[int64]UserSecuritySettings{1: {UserID: 1, AllowAllIPs: true}},
		trustedDeviceCount: map[int64]int{1: 0},
	}
	sessions := &mockSessions{token: "session-token", expiresAt: time.Now().UTC().Add(time.Hour)}
	service := NewService(repo, sessions)

	resp, status, err := service.Login(LoginRequest{SourceSystem: "erp", ExternalUserID: "user-1", Password: "pass123!", DeviceID: "device-1"}, "192.168.1.10", "ua")
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}
	if status != 200 {
		t.Fatalf("status = %d, want 200", status)
	}
	if resp.Token != "session-token" {
		t.Fatalf("token = %q, want session-token", resp.Token)
	}
	if sessions.issuedUserID != 1 || sessions.issuedDeviceID != "device-1" {
		t.Fatalf("unexpected session issue payload: %+v", sessions)
	}
	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("login data type = %T, want map", resp.Data)
	}
	if data["display_name"] != "User One" {
		t.Fatalf("display_name = %v, want User One", data["display_name"])
	}
}

func TestLoginRejectsUnknownUserWithoutCreatingAccount(t *testing.T) {
	repo := &mockRepository{
		usersByExternal: map[string]User{},
	}
	service := NewService(repo, &mockSessions{})

	_, status, err := service.Login(
		LoginRequest{SourceSystem: "erp", ExternalUserID: "missing-user", Password: "pass123!", DeviceID: "device-1"},
		"192.168.1.10",
		"ua",
	)
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
	if status != 404 {
		t.Fatalf("status = %d, want 404", status)
	}
	if len(repo.usersByID) != 0 || len(repo.usersByExternal) != 0 {
		t.Fatal("login created an unknown user")
	}
}

func TestUpdateProfileUpdatesDisplayName(t *testing.T) {
	repo := &mockRepository{
		usersByID:       map[int64]User{1: {ID: 1, SourceSystem: "erp", ExternalUserID: "user-1", DisplayName: "Old Name", Email: "old@example.com"}},
		usersByExternal: map[string]User{"erp:user-1": {ID: 1, SourceSystem: "erp", ExternalUserID: "user-1", DisplayName: "Old Name", Email: "old@example.com"}},
	}
	service := NewService(repo, &mockSessions{})

	resp, status, err := service.UpdateProfile(SessionPrincipal{UserID: 1, DeviceID: "device-1"}, ProfileUpdateRequest{DisplayName: "New Name", Email: "New@Example.com"})
	if err != nil {
		t.Fatalf("UpdateProfile returned error: %v", err)
	}
	if status != 200 {
		t.Fatalf("status = %d, want 200", status)
	}
	if repo.usersByID[1].DisplayName != "New Name" {
		t.Fatalf("stored display name = %q, want New Name", repo.usersByID[1].DisplayName)
	}
	if repo.usersByID[1].Email != "new@example.com" {
		t.Fatalf("stored email = %q, want new@example.com", repo.usersByID[1].Email)
	}
	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("response data type = %T, want map", resp.Data)
	}
	if data["display_name"] != "New Name" {
		t.Fatalf("display_name = %v, want New Name", data["display_name"])
	}
	if data["email"] != "new@example.com" {
		t.Fatalf("email = %v, want new@example.com", data["email"])
	}
}

func TestUpdateProfileRejectsDuplicateEmail(t *testing.T) {
	repo := &mockRepository{
		usersByID: map[int64]User{
			1: {ID: 1, SourceSystem: "erp", ExternalUserID: "user-1", DisplayName: "User One", Email: "user1@example.com"},
			2: {ID: 2, SourceSystem: "erp", ExternalUserID: "user-2", DisplayName: "User Two", Email: "taken@example.com"},
		},
		usersByExternal: map[string]User{
			"erp:user-1": {ID: 1, SourceSystem: "erp", ExternalUserID: "user-1", DisplayName: "User One", Email: "user1@example.com"},
			"erp:user-2": {ID: 2, SourceSystem: "erp", ExternalUserID: "user-2", DisplayName: "User Two", Email: "taken@example.com"},
		},
	}
	service := NewService(repo, &mockSessions{})

	_, status, err := service.UpdateProfile(SessionPrincipal{UserID: 1, DeviceID: "device-1"}, ProfileUpdateRequest{DisplayName: "User One", Email: "taken@example.com"})
	if !errors.Is(err, ErrEmailAlreadyExists) {
		t.Fatalf("expected ErrEmailAlreadyExists, got %v", err)
	}
	if status != 409 {
		t.Fatalf("status = %d, want 409", status)
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	repo := &mockRepository{
		usersByExternal: map[string]User{"erp:user-1": {ID: 1, SourceSystem: "erp", ExternalUserID: "user-1", PasswordHash: mustHashPassword(t, "pass123!")}},
		settingsByUserID: map[int64]UserSecuritySettings{
			1: {UserID: 1, AllowAllIPs: true},
		},
		trustedDeviceCount: map[int64]int{1: 0},
	}
	service := NewService(repo, &mockSessions{})

	_, status, err := service.Login(LoginRequest{SourceSystem: "erp", ExternalUserID: "user-1", Password: "wrong123!", DeviceID: "device-1"}, "192.168.1.10", "ua")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
	if status != 401 {
		t.Fatalf("status = %d, want 401", status)
	}
}

func TestLoginAllowsNewDeviceWhenTrustedDeviceRequirementDisabled(t *testing.T) {
	repo := &mockRepository{
		usersByExternal: map[string]User{"erp:user-1": {ID: 1, SourceSystem: "erp", ExternalUserID: "user-1", PasswordHash: mustHashPassword(t, "pass123!")}},
		settingsByUserID: map[int64]UserSecuritySettings{
			1: {UserID: 1, AllowAllIPs: true},
		},
		trustedDeviceCount: map[int64]int{1: 1},
		devicesByKey: map[string]Device{
			deviceKey(1, "trusted-device"): {UserID: 1, DeviceID: "trusted-device", IsTrustedDevice: true},
		},
	}
	service := NewService(repo, &mockSessions{})

	_, status, err := service.Login(LoginRequest{SourceSystem: "erp", ExternalUserID: "user-1", Password: "pass123!", DeviceID: "new-device"}, "192.168.1.10", "ua")
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}
	if status != 200 {
		t.Fatalf("status = %d, want 200", status)
	}
	if device := repo.devicesByKey[deviceKey(1, "new-device")]; !device.IsTrustedDevice {
		t.Fatalf("new device should be trusted when requirement is disabled: %+v", device)
	}
}

func TestLoginRejectsNewDeviceWhenTrustedDeviceRequirementEnabled(t *testing.T) {
	repo := &mockRepository{
		usersByExternal: map[string]User{"erp:user-1": {ID: 1, SourceSystem: "erp", ExternalUserID: "user-1", PasswordHash: mustHashPassword(t, "pass123!")}},
		settingsByUserID: map[int64]UserSecuritySettings{
			1: {UserID: 1, AllowAllIPs: true},
		},
		trustedDeviceCount: map[int64]int{1: 1},
		devicesByKey: map[string]Device{
			deviceKey(1, "trusted-device"): {UserID: 1, DeviceID: "trusted-device", IsTrustedDevice: true},
		},
	}
	service := NewService(repo, &mockSessions{})
	service.SetRequireTrustedDevice(true)

	_, status, err := service.Login(LoginRequest{SourceSystem: "erp", ExternalUserID: "user-1", Password: "pass123!", DeviceID: "new-device"}, "192.168.1.10", "ua")
	if !errors.Is(err, ErrNewDeviceApprovalRequired) {
		t.Fatalf("expected ErrNewDeviceApprovalRequired, got %v", err)
	}
	if status != 403 {
		t.Fatalf("status = %d, want 403", status)
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{name: "min length", password: "a1!b", wantErr: false},
		{name: "max length", password: "Abcdef1234567890!@#$", wantErr: false},
		{name: "too short", password: "a1!", wantErr: true},
		{name: "too long", password: "Abcdef1234567890!@#$x", wantErr: true},
		{name: "space rejected", password: "abc 123", wantErr: true},
		{name: "unicode rejected", password: "密碼1234", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validatePassword(tc.password)
			if tc.wantErr && !errors.Is(err, ErrInvalidPassword) {
				t.Fatalf("expected ErrInvalidPassword, got %v", err)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("validatePassword returned error: %v", err)
			}
		})
	}
}

func mustHashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := hashPassword(password)
	if err != nil {
		t.Fatalf("hashPassword returned error: %v", err)
	}
	return hash
}

func TestUpdateIPWhitelistRejectsInvalidRule(t *testing.T) {
	repo := &mockRepository{
		usersByID:    map[int64]User{1: {ID: 1}, 9: {ID: 9}},
		systemAdmins: map[int64]bool{9: true},
	}
	service := NewService(repo, &mockSessions{})

	_, _, err := service.UpdateIPWhitelist(1, AdminSessionPrincipal{AdminUserID: 9}, IPWhitelistUpdateRequest{AllowAll: false, Rules: []string{"not-an-ip"}})
	if !errors.Is(err, ErrInvalidIPWhitelistRule) {
		t.Fatalf("expected ErrInvalidIPWhitelistRule, got %v", err)
	}
}

func TestListUsersRequiresSystemAdmin(t *testing.T) {
	repo := &mockRepository{
		usersByID:    map[int64]User{1: {ID: 1}},
		systemAdmins: map[int64]bool{},
	}
	service := NewService(repo, &mockSessions{})

	_, _, err := service.ListUsers(AdminUserFilter{}, AdminSessionPrincipal{AdminUserID: 1})
	if !errors.Is(err, ErrInsufficientRole) {
		t.Fatalf("expected ErrInsufficientRole, got %v", err)
	}
}

func TestListUsersReturnsRepositoryUsers(t *testing.T) {
	repo := &mockRepository{
		usersByID: map[int64]User{
			1: {ID: 1, ExternalUserID: "user-1", DisplayName: "User One", Status: "active", SourceSystem: "erp"},
			9: {ID: 9, ExternalUserID: "admin", DisplayName: "Admin", Status: "active", SourceSystem: "erp"},
		},
		systemAdmins: map[int64]bool{9: true},
	}
	service := NewService(repo, &mockSessions{})

	resp, status, err := service.ListUsers(AdminUserFilter{}, AdminSessionPrincipal{AdminUserID: 9})
	if err != nil {
		t.Fatalf("ListUsers returned error: %v", err)
	}
	if status != 200 {
		t.Fatalf("status = %d, want 200", status)
	}
	users, ok := resp.Data.(AdminUserPage)
	if !ok || len(users.Items) != 2 || users.Total != 2 {
		t.Fatalf("users = %#v, want 2 users", resp.Data)
	}
}

func TestCreateUserRequiresSystemAdmin(t *testing.T) {
	repo := &mockRepository{
		usersByID:       map[int64]User{1: {ID: 1}},
		usersByExternal: map[string]User{},
		systemAdmins:    map[int64]bool{},
	}
	service := NewService(repo, &mockSessions{})

	_, _, err := service.CreateUser(AdminCreateUserRequest{
		SourceSystem:   "erp",
		ExternalUserID: "new-user",
		DisplayName:    "New User",
		Password:       "pass123!",
	}, AdminSessionPrincipal{AdminUserID: 1})
	if !errors.Is(err, ErrInsufficientRole) {
		t.Fatalf("expected ErrInsufficientRole, got %v", err)
	}
}

func TestCreateUserCreatesUserWithHashedPassword(t *testing.T) {
	repo := &mockRepository{
		usersByID:       map[int64]User{9: {ID: 9}},
		usersByExternal: map[string]User{},
		systemAdmins:    map[int64]bool{9: true},
	}
	service := NewService(repo, &mockSessions{})

	resp, status, err := service.CreateUser(AdminCreateUserRequest{
		SourceSystem:   "office",
		ExternalUserID: "new-user",
		DisplayName:    "New User",
		Email:          "new@example.com",
		Password:       "pass123!",
		Language:       "zh-Hant",
		Status:         "inactive",
	}, AdminSessionPrincipal{AdminUserID: 9})
	if err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}
	if status != 201 {
		t.Fatalf("status = %d, want 201", status)
	}
	if resp.Code != "USER_CREATED" {
		t.Fatalf("code = %q, want USER_CREATED", resp.Code)
	}

	created, ok := repo.usersByExternal["office:new-user"]
	if !ok {
		t.Fatal("created user not stored")
	}
	if created.PasswordHash == "" || created.PasswordHash == "pass123!" {
		t.Fatal("password was not hashed")
	}
	if created.Status != "inactive" {
		t.Fatalf("status = %q, want inactive", created.Status)
	}
}

func TestInviteUserCreatesPendingInvitation(t *testing.T) {
	repo := &mockRepository{
		usersByID:          map[int64]User{9: {ID: 9}},
		usersByExternal:    map[string]User{},
		invitationsByEmail: map[string]UserInvitation{},
		systemAdmins:       map[int64]bool{9: true},
		adminsByID:         map[int64]SystemAdminSummary{9: {ID: 9, UserID: 9, AdminUserID: 9, Role: "system_admin", Status: "active"}},
	}
	service := NewService(repo, &mockSessions{})
	service.SetInvitationTTL(time.Hour)
	service.SetInvitationBaseURL("https://chat.example.com")
	mailer := &mockInvitationMailer{}
	service.SetInvitationMailer(mailer)

	resp, status, err := service.InviteUser(AdminInviteUserRequest{Email: "New@Example.COM"}, AdminSessionPrincipal{AdminUserID: 9})
	if err != nil {
		t.Fatalf("InviteUser returned error: %v", err)
	}
	if status != 201 {
		t.Fatalf("status = %d, want 201", status)
	}
	if resp.Code != "USER_INVITATION_CREATED" {
		t.Fatalf("code = %q, want USER_INVITATION_CREATED", resp.Code)
	}
	if repo.createdInvitation.Email != "new@example.com" {
		t.Fatalf("email = %q, want normalized new@example.com", repo.createdInvitation.Email)
	}
	if repo.createdInvitation.TokenHash == "" {
		t.Fatal("token hash was not stored")
	}
	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("data = %#v, want map", resp.Data)
	}
	token, _ := data["token"].(string)
	inviteURL, _ := data["invite_url"].(string)
	if token == "" || !strings.Contains(inviteURL, token) {
		t.Fatalf("token/invite_url not returned correctly: %#v", data)
	}
	if repo.createdInvitation.Status != "pending" {
		t.Fatalf("status = %q, want pending", repo.createdInvitation.Status)
	}
	if mailer.email != "new@example.com" || mailer.inviteURL == "" {
		t.Fatalf("mailer was not called correctly: %#v", mailer)
	}
	if repo.sentInvitationID != repo.createdInvitation.ID {
		t.Fatalf("sent invitation id = %d, want %d", repo.sentInvitationID, repo.createdInvitation.ID)
	}
}

func TestInviteUserRequiresMailer(t *testing.T) {
	repo := &mockRepository{
		usersByID:          map[int64]User{9: {ID: 9}},
		usersByExternal:    map[string]User{},
		invitationsByEmail: map[string]UserInvitation{},
		systemAdmins:       map[int64]bool{9: true},
		adminsByID:         map[int64]SystemAdminSummary{9: {ID: 9, UserID: 9, AdminUserID: 9, Role: "system_admin", Status: "active"}},
	}
	service := NewService(repo, &mockSessions{})

	_, status, err := service.InviteUser(AdminInviteUserRequest{Email: "new@example.com"}, AdminSessionPrincipal{AdminUserID: 9})
	if !errors.Is(err, ErrInvitationMailerUnavailable) {
		t.Fatalf("expected ErrInvitationMailerUnavailable, got %v", err)
	}
	if status != 503 {
		t.Fatalf("status = %d, want 503", status)
	}
	if len(repo.invitationsByEmail) != 0 {
		t.Fatal("invitation should not be created when mailer is unavailable")
	}
}

func TestInviteUserRejectsExistingUserEmail(t *testing.T) {
	repo := &mockRepository{
		usersByID: map[int64]User{
			1: {ID: 1, Email: "new@example.com"},
			9: {ID: 9},
		},
		usersByExternal: map[string]User{},
		systemAdmins:    map[int64]bool{9: true},
		adminsByID:      map[int64]SystemAdminSummary{9: {ID: 9, UserID: 9, AdminUserID: 9, Role: "system_admin", Status: "active"}},
	}
	service := NewService(repo, &mockSessions{})

	_, status, err := service.InviteUser(AdminInviteUserRequest{Email: "NEW@example.com"}, AdminSessionPrincipal{AdminUserID: 9})
	if !errors.Is(err, ErrEmailAlreadyExists) {
		t.Fatalf("expected ErrEmailAlreadyExists, got %v", err)
	}
	if status != 409 {
		t.Fatalf("status = %d, want 409", status)
	}
}

func TestInviteUserRejectsPendingInvitation(t *testing.T) {
	sentAt := time.Now().UTC().Add(-time.Minute)
	repo := &mockRepository{
		usersByID:       map[int64]User{9: {ID: 9}},
		usersByExternal: map[string]User{},
		invitationsByEmail: map[string]UserInvitation{
			"new@example.com": {
				ID:        1,
				Email:     "new@example.com",
				Status:    "pending",
				ExpiresAt: time.Now().UTC().Add(time.Hour),
				SentAt:    &sentAt,
			},
		},
		systemAdmins: map[int64]bool{9: true},
		adminsByID:   map[int64]SystemAdminSummary{9: {ID: 9, UserID: 9, AdminUserID: 9, Role: "system_admin", Status: "active"}},
	}
	service := NewService(repo, &mockSessions{})
	service.SetInvitationMailer(&mockInvitationMailer{})

	_, status, err := service.InviteUser(AdminInviteUserRequest{Email: "new@example.com"}, AdminSessionPrincipal{AdminUserID: 9})
	if !errors.Is(err, ErrUserInvitationPending) {
		t.Fatalf("expected ErrUserInvitationPending, got %v", err)
	}
	if status != 409 {
		t.Fatalf("status = %d, want 409", status)
	}
}

func TestInviteUserRefreshesUnsentPendingInvitation(t *testing.T) {
	repo := &mockRepository{
		usersByID:       map[int64]User{9: {ID: 9}},
		usersByExternal: map[string]User{},
		invitationsByEmail: map[string]UserInvitation{
			"new@example.com": {
				ID:        7,
				Email:     "new@example.com",
				TokenHash: "old-token-hash",
				Status:    "pending",
				ExpiresAt: time.Now().UTC().Add(time.Hour),
			},
		},
		systemAdmins: map[int64]bool{9: true},
		adminsByID:   map[int64]SystemAdminSummary{9: {ID: 9, UserID: 9, AdminUserID: 9, Role: "system_admin", Status: "active"}},
	}
	mailer := &mockInvitationMailer{}
	service := NewService(repo, &mockSessions{})
	service.SetInvitationMailer(mailer)
	service.SetInvitationBaseURL("https://chat.example.com")

	resp, status, err := service.InviteUser(AdminInviteUserRequest{Email: "new@example.com"}, AdminSessionPrincipal{AdminUserID: 9})
	if err != nil {
		t.Fatalf("InviteUser returned error: %v", err)
	}
	if status != 201 {
		t.Fatalf("status = %d, want 201", status)
	}
	if repo.refreshedInvitation.ID != 7 {
		t.Fatalf("refreshed invitation id = %d, want 7", repo.refreshedInvitation.ID)
	}
	if repo.refreshedInvitation.TokenHash == "" || repo.refreshedInvitation.TokenHash == "old-token-hash" {
		t.Fatalf("token hash was not refreshed: %q", repo.refreshedInvitation.TokenHash)
	}
	if mailer.email != "new@example.com" {
		t.Fatalf("mailer email = %q, want new@example.com", mailer.email)
	}
	if repo.sentInvitationID != 7 {
		t.Fatalf("sent invitation id = %d, want 7", repo.sentInvitationID)
	}
	data, ok := resp.Data.(map[string]any)
	if !ok || data["token"] == "" {
		t.Fatalf("response did not include new token: %#v", resp.Data)
	}
}

func TestResendUserInvitationRefreshesPendingInvitation(t *testing.T) {
	sentAt := time.Now().UTC().Add(-time.Hour)
	repo := &mockRepository{
		usersByID:       map[int64]User{9: {ID: 9}},
		usersByExternal: map[string]User{},
		invitationsByEmail: map[string]UserInvitation{
			"new@example.com": {
				ID:        12,
				Email:     "new@example.com",
				TokenHash: "old-token-hash",
				Status:    "pending",
				ExpiresAt: time.Now().UTC().Add(time.Hour),
				SentAt:    &sentAt,
			},
		},
		systemAdmins: map[int64]bool{9: true},
		adminsByID:   map[int64]SystemAdminSummary{9: {ID: 9, UserID: 9, AdminUserID: 9, Role: "system_admin", Status: "active"}},
	}
	mailer := &mockInvitationMailer{}
	service := NewService(repo, &mockSessions{})
	service.SetInvitationMailer(mailer)
	service.SetInvitationBaseURL("https://chat.example.com")

	resp, status, err := service.ResendUserInvitation(AdminInviteUserRequest{Email: "new@example.com"}, AdminSessionPrincipal{AdminUserID: 9})
	if err != nil {
		t.Fatalf("ResendUserInvitation returned error: %v", err)
	}
	if status != 200 {
		t.Fatalf("status = %d, want 200", status)
	}
	if resp.Code != "USER_INVITATION_RESENT" {
		t.Fatalf("code = %q, want USER_INVITATION_RESENT", resp.Code)
	}
	if repo.refreshedInvitation.ID != 12 {
		t.Fatalf("refreshed invitation id = %d, want 12", repo.refreshedInvitation.ID)
	}
	if repo.refreshedInvitation.TokenHash == "" || repo.refreshedInvitation.TokenHash == "old-token-hash" {
		t.Fatalf("token hash was not refreshed: %q", repo.refreshedInvitation.TokenHash)
	}
	if mailer.email != "new@example.com" {
		t.Fatalf("mailer email = %q, want new@example.com", mailer.email)
	}
	if repo.sentInvitationID != 12 {
		t.Fatalf("sent invitation id = %d, want 12", repo.sentInvitationID)
	}
}

func TestResendUserInvitationRejectsCompletedInvitation(t *testing.T) {
	acceptedAt := time.Now().UTC().Add(-time.Hour)
	repo := &mockRepository{
		usersByID:       map[int64]User{9: {ID: 9}},
		usersByExternal: map[string]User{},
		invitationsByEmail: map[string]UserInvitation{
			"new@example.com": {
				ID:             12,
				Email:          "new@example.com",
				Status:         "accepted",
				AcceptedUserID: 88,
				AcceptedAt:     &acceptedAt,
				ExpiresAt:      time.Now().UTC().Add(time.Hour),
			},
		},
		systemAdmins: map[int64]bool{9: true},
		adminsByID:   map[int64]SystemAdminSummary{9: {ID: 9, UserID: 9, AdminUserID: 9, Role: "system_admin", Status: "active"}},
	}
	service := NewService(repo, &mockSessions{})
	service.SetInvitationMailer(&mockInvitationMailer{})

	_, status, err := service.ResendUserInvitation(AdminInviteUserRequest{Email: "new@example.com"}, AdminSessionPrincipal{AdminUserID: 9})
	if !errors.Is(err, ErrUserInvitationCompleted) {
		t.Fatalf("expected ErrUserInvitationCompleted, got %v", err)
	}
	if status != 409 {
		t.Fatalf("status = %d, want 409", status)
	}
	if repo.refreshedInvitation.ID != 0 {
		t.Fatal("completed invitation should not be refreshed")
	}
}

func TestCreateSystemAdminCreatesBackendAdmin(t *testing.T) {
	repo := &mockRepository{
		usersByID:       map[int64]User{9: {ID: 9}},
		usersByExternal: map[string]User{},
		systemAdmins:    map[int64]bool{9: true},
		adminsByID:      map[int64]SystemAdminSummary{},
		adminsByAccount: map[string]SystemAdminSummary{},
	}
	service := NewService(repo, &mockSessions{})

	resp, status, err := service.CreateSystemAdmin(SystemAdminCreateRequest{
		ExternalUserID: "manager01",
		DisplayName:    "管理員一",
		Password:       "pass123!",
		Status:         "active",
	}, AdminSessionPrincipal{AdminUserID: 9})
	if err != nil {
		t.Fatalf("CreateSystemAdmin returned error: %v", err)
	}
	if status != 201 {
		t.Fatalf("status = %d, want 201", status)
	}
	if resp.Code != "SYSTEM_ADMIN_CREATED" {
		t.Fatalf("code = %q, want SYSTEM_ADMIN_CREATED", resp.Code)
	}

	created, ok := repo.adminsByAccount["manager01"]
	if !ok {
		t.Fatal("created system admin not stored")
	}
	if created.PasswordHash == "" || created.PasswordHash == "pass123!" {
		t.Fatal("password was not hashed")
	}
	if _, ok := repo.usersByExternal["office:manager01"]; ok {
		t.Fatal("system admin should not be stored as a chat user")
	}
}

func TestBootstrapSystemAdminCreatesFirstAdmin(t *testing.T) {
	repo := &mockRepository{
		usersByID:       map[int64]User{},
		usersByExternal: map[string]User{},
		systemAdmins:    map[int64]bool{},
		adminsByID:      map[int64]SystemAdminSummary{},
		adminsByAccount: map[string]SystemAdminSummary{},
	}
	service := NewService(repo, nil)

	result, err := service.BootstrapSystemAdmin(SystemAdminCreateRequest{
		ExternalUserID: "admin01",
		DisplayName:    "系統管理員",
		Password:       "pass123!",
	})
	if err != nil {
		t.Fatalf("BootstrapSystemAdmin returned error: %v", err)
	}
	if !result.AdminCreated {
		t.Fatal("AdminCreated = false, want true")
	}

	created, ok := repo.adminsByAccount["admin01"]
	if !ok {
		t.Fatal("bootstrap admin not stored")
	}
	if created.PasswordHash == "" || created.PasswordHash == "pass123!" {
		t.Fatal("password was not hashed")
	}
}

func TestBootstrapSystemAdminIsIdempotent(t *testing.T) {
	hash := mustHashPassword(t, "oldpass!")
	admin := SystemAdminSummary{ID: 1, UserID: 1, AdminUserID: 1, ExternalUserID: "admin01", DisplayName: "Existing Admin", PasswordHash: hash, Role: "system_admin", RoleName: "系統管理員", Status: "active"}
	repo := &mockRepository{
		usersByID:       map[int64]User{},
		usersByExternal: map[string]User{},
		systemAdmins:    map[int64]bool{},
		adminsByID:      map[int64]SystemAdminSummary{1: admin},
		adminsByAccount: map[string]SystemAdminSummary{"admin01": admin},
	}
	service := NewService(repo, nil)

	result, err := service.BootstrapSystemAdmin(SystemAdminCreateRequest{
		ExternalUserID: "admin01",
		DisplayName:    "系統管理員",
		Password:       "newpass!",
	})
	if err != nil {
		t.Fatalf("BootstrapSystemAdmin returned error: %v", err)
	}
	if result.AdminCreated {
		t.Fatal("AdminCreated = true, want false")
	}
	if repo.adminsByID[1].PasswordHash != hash {
		t.Fatal("bootstrap should not replace an existing password")
	}
}

func TestUpdateUserKeepsPasswordWhenBlank(t *testing.T) {
	originalHash := mustHashPassword(t, "oldpass!")
	repo := &mockRepository{
		usersByID: map[int64]User{
			1: {ID: 1, SourceSystem: "office", ExternalUserID: "old-user", DisplayName: "Old", PasswordHash: originalHash, Status: "active"},
			9: {ID: 9},
		},
		usersByExternal: map[string]User{},
		systemAdmins:    map[int64]bool{9: true},
	}
	service := NewService(repo, &mockSessions{})

	_, status, err := service.UpdateUser(1, AdminUpdateUserRequest{
		DisplayName: "Updated",
		Email:       "updated@example.com",
		Status:      "inactive",
		Password:    "",
	}, AdminSessionPrincipal{AdminUserID: 9})
	if err != nil {
		t.Fatalf("UpdateUser returned error: %v", err)
	}
	if status != 200 {
		t.Fatalf("status = %d, want 200", status)
	}
	if repo.usersByID[1].PasswordHash != originalHash {
		t.Fatal("blank password changed password hash")
	}
	if repo.usersByID[1].ExternalUserID != "old-user" {
		t.Fatal("update changed immutable external user id")
	}
}

func TestUpdateUserChangesPasswordWhenProvided(t *testing.T) {
	originalHash := mustHashPassword(t, "oldpass!")
	repo := &mockRepository{
		usersByID: map[int64]User{
			1: {ID: 1, SourceSystem: "office", ExternalUserID: "user-1", DisplayName: "User", PasswordHash: originalHash, Status: "active"},
			9: {ID: 9},
		},
		usersByExternal: map[string]User{},
		systemAdmins:    map[int64]bool{9: true},
	}
	service := NewService(repo, &mockSessions{})

	_, _, err := service.UpdateUser(1, AdminUpdateUserRequest{
		DisplayName: "User",
		Status:      "active",
		Password:    "newpass!",
	}, AdminSessionPrincipal{AdminUserID: 9})
	if err != nil {
		t.Fatalf("UpdateUser returned error: %v", err)
	}
	if repo.usersByID[1].PasswordHash == originalHash {
		t.Fatal("provided password did not change password hash")
	}
	if !passwordMatches("newpass!", repo.usersByID[1].PasswordHash) {
		t.Fatal("updated password hash does not match new password")
	}
}

func TestUpdateIPWhitelistRequiresSystemAdmin(t *testing.T) {
	repo := &mockRepository{
		usersByID:    map[int64]User{1: {ID: 1}, 2: {ID: 2}},
		systemAdmins: map[int64]bool{},
	}
	service := NewService(repo, &mockSessions{})

	_, _, err := service.UpdateIPWhitelist(1, AdminSessionPrincipal{AdminUserID: 2}, IPWhitelistUpdateRequest{AllowAll: true})
	if !errors.Is(err, ErrInsufficientRole) {
		t.Fatalf("expected ErrInsufficientRole, got %v", err)
	}
}

func TestApproveDeviceByTrustedSessionDevice(t *testing.T) {
	repo := &mockRepository{
		usersByID:    map[int64]User{1: {ID: 1}},
		systemAdmins: map[int64]bool{},
		devicesByKey: map[string]Device{
			deviceKey(1, "trusted-device"): {UserID: 1, DeviceID: "trusted-device", IsTrustedDevice: true},
			deviceKey(1, "pending-device"): {UserID: 1, DeviceID: "pending-device", IsTrustedDevice: false},
		},
	}
	service := NewService(repo, &mockSessions{})

	_, _, err := service.ApproveDevice(1, "pending-device", SessionPrincipal{UserID: 1, DeviceID: "trusted-device"})
	if err != nil {
		t.Fatalf("ApproveDevice returned error: %v", err)
	}
	if !repo.approveCalled {
		t.Fatal("expected approve device to be called")
	}
	if repo.approvedTrustedBy != 1 {
		t.Fatalf("approvedTrustedBy = %d, want 1", repo.approvedTrustedBy)
	}
}

func TestApproveDeviceRejectsUntrustedSessionDevice(t *testing.T) {
	repo := &mockRepository{
		usersByID:    map[int64]User{1: {ID: 1}},
		systemAdmins: map[int64]bool{},
		devicesByKey: map[string]Device{
			deviceKey(1, "untrusted-device"): {UserID: 1, DeviceID: "untrusted-device", IsTrustedDevice: false},
			deviceKey(1, "pending-device"):   {UserID: 1, DeviceID: "pending-device", IsTrustedDevice: false},
		},
	}
	service := NewService(repo, &mockSessions{})

	_, _, err := service.ApproveDevice(1, "pending-device", SessionPrincipal{UserID: 1, DeviceID: "untrusted-device"})
	if !errors.Is(err, ErrDeviceNotTrusted) {
		t.Fatalf("expected ErrDeviceNotTrusted, got %v", err)
	}
}
