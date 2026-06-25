package erp

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

type stubIntegrationRepository struct{}

func (s *stubIntegrationRepository) CreateUser(params RegisterParams) (User, error) {
	return User{ID: 1, SourceSystem: params.SourceSystem, ExternalUserID: params.ExternalUserID, DisplayName: params.DisplayName, Status: params.Status}, nil
}

func (s *stubIntegrationRepository) FindUserByID(userID int64) (User, error) {
	return User{}, ErrUserNotFound
}

func (s *stubIntegrationRepository) FindUserByExternal(sourceSystem, externalUserID string) (User, error) {
	return User{}, ErrUserNotFound
}

func (s *stubIntegrationRepository) ListUsers(filter AdminUserFilter) ([]AdminUserSummary, error) {
	return nil, nil
}

func (s *stubIntegrationRepository) UpdateUser(userID int64, params AdminUpdateUserParams) (User, error) {
	return User{ID: userID, SourceSystem: "office", ExternalUserID: "user_a", DisplayName: params.DisplayName, Email: params.Email, Status: params.Status}, nil
}

func (s *stubIntegrationRepository) UpdateUserProfile(userID int64, displayName string) (User, error) {
	return User{ID: userID, SourceSystem: "erp", ExternalUserID: "user_a", DisplayName: displayName}, nil
}

func (s *stubIntegrationRepository) IsSystemAdmin(userID int64) (bool, error) {
	return false, nil
}

func (s *stubIntegrationRepository) GetUserSecuritySettings(userID int64) (UserSecuritySettings, error) {
	return UserSecuritySettings{}, nil
}

func (s *stubIntegrationRepository) ListActiveIPWhitelistRules(userID int64) ([]string, error) {
	return nil, nil
}

func (s *stubIntegrationRepository) ReplaceIPWhitelist(userID int64, allowAll bool, rules []string) error {
	return nil
}

func (s *stubIntegrationRepository) CountTrustedDevices(userID int64) (int, error) {
	return 0, nil
}

func (s *stubIntegrationRepository) ListDevices(userID int64) ([]Device, error) {
	return nil, nil
}

func (s *stubIntegrationRepository) FindDevice(userID int64, deviceID string) (Device, error) {
	return Device{}, ErrDeviceNotFound
}

func (s *stubIntegrationRepository) UpsertDeviceLogin(userID int64, deviceID, userAgent, ip string, trusted bool, now time.Time) error {
	return nil
}

func (s *stubIntegrationRepository) ApproveDevice(userID int64, deviceID string, trustedBy int64, now time.Time) error {
	return nil
}

func TestRegisterRejectsMissingSignatureWhenSecretConfigured(t *testing.T) {
	handler := NewHandler(
		NewService(&stubIntegrationRepository{}, nil),
		nil,
		"shared-token",
		"signature-secret",
		5*time.Minute,
	)

	request := httptest.NewRequest(http.MethodPost, "/api/erp/register", bytes.NewBufferString(`{"source_system":"erp","external_user_id":"user_a","password":"pass123!","display_name":"User A"}`))
	request.Header.Set("Authorization", "Bearer shared-token")
	request.Header.Set("X-TWACC-Timestamp", strconvFormatInt(time.Now().UTC().Unix()))

	recorder := httptest.NewRecorder()
	handler.Register(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", recorder.Code)
	}

	var resp Response
	if err := json.NewDecoder(recorder.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Code != "INTEGRATION_SIGNATURE_REQUIRED" {
		t.Fatalf("code = %q, want INTEGRATION_SIGNATURE_REQUIRED", resp.Code)
	}
}

func TestRegisterRejectsExpiredTimestamp(t *testing.T) {
	handler := NewHandler(
		NewService(&stubIntegrationRepository{}, nil),
		nil,
		"shared-token",
		"signature-secret",
		5*time.Minute,
	)

	body := []byte(`{"source_system":"erp","external_user_id":"user_a","password":"pass123!","display_name":"User A"}`)
	timestamp := time.Now().UTC().Add(-10 * time.Minute).Unix()
	request := httptest.NewRequest(http.MethodPost, "/api/erp/register", bytes.NewReader(body))
	request.Header.Set("Authorization", "Bearer shared-token")
	request.Header.Set("X-TWACC-Timestamp", strconvFormatInt(timestamp))
	request.Header.Set("X-TWACC-Signature", signIntegrationRequest("signature-secret", http.MethodPost, "/api/erp/register", strconvFormatInt(timestamp), body))

	recorder := httptest.NewRecorder()
	handler.Register(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", recorder.Code)
	}

	var resp Response
	if err := json.NewDecoder(recorder.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Code != "INTEGRATION_TIMESTAMP_EXPIRED" {
		t.Fatalf("code = %q, want INTEGRATION_TIMESTAMP_EXPIRED", resp.Code)
	}
}

func TestRegisterAcceptsValidSignedRequest(t *testing.T) {
	handler := NewHandler(
		NewService(&stubIntegrationRepository{}, nil),
		nil,
		"shared-token",
		"signature-secret",
		5*time.Minute,
	)

	body := []byte(`{"source_system":"erp","external_user_id":"user_a","password":"pass123!","display_name":"User A"}`)
	timestamp := strconvFormatInt(time.Now().UTC().Unix())
	request := httptest.NewRequest(http.MethodPost, "/api/erp/register", bytes.NewReader(body))
	request.Header.Set("Authorization", "Bearer shared-token")
	request.Header.Set("X-TWACC-Timestamp", timestamp)
	request.Header.Set("X-TWACC-Signature", signIntegrationRequest("signature-secret", http.MethodPost, "/api/erp/register", timestamp, body))

	recorder := httptest.NewRecorder()
	handler.Register(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", recorder.Code, recorder.Body.String())
	}

	var resp Response
	if err := json.NewDecoder(recorder.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.Success || resp.Code != "REGISTERED" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func signIntegrationRequest(secret, method, path, timestamp string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(method))
	mac.Write([]byte("\n"))
	mac.Write([]byte(path))
	mac.Write([]byte("\n"))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("\n"))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func strconvFormatInt(value int64) string {
	return strconv.FormatInt(value, 10)
}
