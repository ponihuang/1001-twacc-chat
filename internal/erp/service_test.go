package erp

import (
	"errors"
	"strconv"
	"testing"
	"time"
)

type mockRepository struct {
	usersByID          map[int64]User
	usersByExternal    map[string]User
	systemAdmins       map[int64]bool
	settingsByUserID   map[int64]UserSecuritySettings
	whitelistRules     map[int64][]string
	devicesByKey       map[string]Device
	replaceAllowAll    bool
	replaceRules       []string
	replaceCalled      bool
	approveCalled      bool
	approvedTrustedBy  int64
	approvedUserID     int64
	approvedDeviceID   string
	trustedDeviceCount map[int64]int
}

type mockSessions struct {
	issuedUserID   int64
	issuedDeviceID string
	token          string
	expiresAt      time.Time
}

func (m *mockRepository) CreateUser(params RegisterParams) (User, error) {
	key := params.SourceSystem + ":" + params.ExternalUserID
	if _, ok := m.usersByExternal[key]; ok {
		return User{}, ErrUserAlreadyExists
	}
	user := User{ID: int64(len(m.usersByID) + 1), SourceSystem: params.SourceSystem, ExternalUserID: params.ExternalUserID, DisplayName: params.DisplayName, PasswordHash: params.PasswordHash, Language: params.Language}
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
		usersByExternal:    map[string]User{"erp:user-1": {ID: 1, SourceSystem: "erp", ExternalUserID: "user-1", PasswordHash: mustHashPassword(t, "pass123!")}},
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

	_, _, err := service.UpdateIPWhitelist(1, SessionPrincipal{UserID: 9}, IPWhitelistUpdateRequest{AllowAll: false, Rules: []string{"not-an-ip"}})
	if !errors.Is(err, ErrInvalidIPWhitelistRule) {
		t.Fatalf("expected ErrInvalidIPWhitelistRule, got %v", err)
	}
}

func TestUpdateIPWhitelistRequiresSystemAdmin(t *testing.T) {
	repo := &mockRepository{
		usersByID:    map[int64]User{1: {ID: 1}, 2: {ID: 2}},
		systemAdmins: map[int64]bool{},
	}
	service := NewService(repo, &mockSessions{})

	_, _, err := service.UpdateIPWhitelist(1, SessionPrincipal{UserID: 2}, IPWhitelistUpdateRequest{AllowAll: true})
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
