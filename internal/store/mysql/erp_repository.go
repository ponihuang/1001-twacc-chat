package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"1001-twacc-chat/internal/erp"
	mysqlDriver "github.com/go-sql-driver/mysql"
)

// IntegrationRepository implements integration-facing persistence with MySQL.
type IntegrationRepository struct {
	db *sql.DB
}

// NewIntegrationRepository creates a MySQL repository for external-system login flows.
func NewIntegrationRepository(db *sql.DB) *IntegrationRepository {
	return &IntegrationRepository{db: db}
}

// NewERPRepository preserves the old constructor name while the package is still under internal/erp.
func NewERPRepository(db *sql.DB) *IntegrationRepository {
	return NewIntegrationRepository(db)
}

// CreateUser inserts a new externally-sourced user and default security settings.
func (r *IntegrationRepository) CreateUser(params erp.RegisterParams) (erp.User, error) {
	result, err := r.db.Exec(
		`INSERT INTO users (source_system, external_user_id, display_name, password_hash, email, language, whatsapp_account, telegram_account, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'active')`,
		params.SourceSystem,
		params.ExternalUserID,
		params.DisplayName,
		params.PasswordHash,
		nullableString(params.Email),
		params.Language,
		nullableString(params.WhatsAppAccount),
		nullableString(params.TelegramAccount),
	)
	if err != nil {
		if isDuplicate(err) {
			return erp.User{}, erp.ErrUserAlreadyExists
		}
		return erp.User{}, fmt.Errorf("create user: %w", err)
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return erp.User{}, fmt.Errorf("read user id: %w", err)
	}

	if _, err := r.db.Exec(`INSERT INTO user_security_settings (user_id, allow_all_ips) VALUES (?, 1)`, userID); err != nil {
		return erp.User{}, fmt.Errorf("create user security settings: %w", err)
	}

	return r.FindUserByExternal(params.SourceSystem, params.ExternalUserID)
}

// FindUserByID loads a user via numeric id.
func (r *IntegrationRepository) FindUserByID(userID int64) (erp.User, error) {
	var user erp.User
	row := r.db.QueryRow(
		`SELECT id, source_system, external_user_id, display_name, password_hash, COALESCE(email, ''), language,
		        COALESCE(whatsapp_account, ''), COALESCE(telegram_account, ''), status, created_at, updated_at
		   FROM users
		  WHERE id = ?
		  LIMIT 1`,
		userID,
	)
	if err := row.Scan(
		&user.ID,
		&user.SourceSystem,
		&user.ExternalUserID,
		&user.DisplayName,
		&user.PasswordHash,
		&user.Email,
		&user.Language,
		&user.WhatsAppAccount,
		&user.TelegramAccount,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return erp.User{}, erp.ErrUserNotFound
		}
		return erp.User{}, fmt.Errorf("find user by id: %w", err)
	}

	return user, nil
}

// FindUserByExternal loads a user via source_system + external_user_id.
func (r *IntegrationRepository) FindUserByExternal(sourceSystem, externalUserID string) (erp.User, error) {
	var user erp.User
	row := r.db.QueryRow(
		`SELECT id, source_system, external_user_id, display_name, password_hash, COALESCE(email, ''), language,
		        COALESCE(whatsapp_account, ''), COALESCE(telegram_account, ''), status, created_at, updated_at
		   FROM users
		  WHERE source_system = ? AND external_user_id = ?
		  LIMIT 1`,
		sourceSystem,
		externalUserID,
	)
	if err := row.Scan(
		&user.ID,
		&user.SourceSystem,
		&user.ExternalUserID,
		&user.DisplayName,
		&user.PasswordHash,
		&user.Email,
		&user.Language,
		&user.WhatsAppAccount,
		&user.TelegramAccount,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return erp.User{}, erp.ErrUserNotFound
		}
		return erp.User{}, fmt.Errorf("find user: %w", err)
	}

	return user, nil
}

// UpdateUserProfile updates mutable profile fields for a user.
func (r *IntegrationRepository) UpdateUserProfile(userID int64, displayName string) (erp.User, error) {
	result, err := r.db.Exec(`UPDATE users SET display_name = ? WHERE id = ?`, strings.TrimSpace(displayName), userID)
	if err != nil {
		return erp.User{}, fmt.Errorf("update user profile: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return erp.User{}, fmt.Errorf("update user profile rows affected: %w", err)
	}
	if affected == 0 {
		return erp.User{}, erp.ErrUserNotFound
	}

	return r.FindUserByID(userID)
}

// IsSystemAdmin checks whether the user belongs to system_admin_users.
func (r *IntegrationRepository) IsSystemAdmin(userID int64) (bool, error) {
	var exists int
	row := r.db.QueryRow(`SELECT 1 FROM system_admin_users WHERE user_id = ? LIMIT 1`, userID)
	if err := row.Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("check system admin: %w", err)
	}

	return true, nil
}

// GetUserSecuritySettings returns per-user IP restriction settings.
func (r *IntegrationRepository) GetUserSecuritySettings(userID int64) (erp.UserSecuritySettings, error) {
	var settings erp.UserSecuritySettings
	var allowAll bool
	row := r.db.QueryRow(`SELECT user_id, allow_all_ips FROM user_security_settings WHERE user_id = ? LIMIT 1`, userID)
	if err := row.Scan(&settings.UserID, &allowAll); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return erp.UserSecuritySettings{UserID: userID, AllowAllIPs: true}, nil
		}
		return erp.UserSecuritySettings{}, fmt.Errorf("get user security settings: %w", err)
	}

	settings.AllowAllIPs = allowAll
	return settings, nil
}

// ListActiveIPWhitelistRules returns active whitelist entries for a user.
func (r *IntegrationRepository) ListActiveIPWhitelistRules(userID int64) ([]string, error) {
	rows, err := r.db.Query(`SELECT ip_or_cidr FROM ip_whitelists WHERE user_id = ? AND enabled = 1 ORDER BY id ASC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list ip whitelist rules: %w", err)
	}
	defer rows.Close()

	var rules []string
	for rows.Next() {
		var rule string
		if err := rows.Scan(&rule); err != nil {
			return nil, fmt.Errorf("scan ip whitelist rule: %w", err)
		}
		rules = append(rules, rule)
	}

	return rules, rows.Err()
}

// ReplaceIPWhitelist updates allow_all and replaces active rules.
func (r *IntegrationRepository) ReplaceIPWhitelist(userID int64, allowAll bool, rules []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin replace ip whitelist: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(
		`INSERT INTO user_security_settings (user_id, allow_all_ips)
		 VALUES (?, ?)
		 ON DUPLICATE KEY UPDATE allow_all_ips = VALUES(allow_all_ips), updated_at = CURRENT_TIMESTAMP`,
		userID,
		allowAll,
	); err != nil {
		return fmt.Errorf("upsert user security settings: %w", err)
	}

	if _, err := tx.Exec(`DELETE FROM ip_whitelists WHERE user_id = ?`, userID); err != nil {
		return fmt.Errorf("delete ip whitelist rules: %w", err)
	}

	for _, rule := range rules {
		if _, err := tx.Exec(
			`INSERT INTO ip_whitelists (user_id, rule_type, ip_or_cidr, enabled) VALUES (?, ?, ?, 1)`,
			userID,
			ruleType(rule),
			rule,
		); err != nil {
			return fmt.Errorf("insert ip whitelist rule: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit replace ip whitelist: %w", err)
	}

	return nil
}

// CountTrustedDevices counts trusted devices for a user.
func (r *IntegrationRepository) CountTrustedDevices(userID int64) (int, error) {
	var count int
	row := r.db.QueryRow(`SELECT COUNT(*) FROM devices WHERE user_id = ? AND is_trusted_device = 1`, userID)
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("count trusted devices: %w", err)
	}

	return count, nil
}

// ListDevices returns tracked devices ordered by recent login.
func (r *IntegrationRepository) ListDevices(userID int64) ([]erp.Device, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, device_id, user_agent, ip, first_login_at, last_login_at, is_trusted_device
		   FROM devices
		  WHERE user_id = ?
		  ORDER BY last_login_at DESC, id DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list devices: %w", err)
	}
	defer rows.Close()

	var devices []erp.Device
	for rows.Next() {
		var device erp.Device
		if err := rows.Scan(
			&device.ID,
			&device.UserID,
			&device.DeviceID,
			&device.UserAgent,
			&device.IP,
			&device.FirstLoginAt,
			&device.LastLoginAt,
			&device.IsTrustedDevice,
		); err != nil {
			return nil, fmt.Errorf("scan device: %w", err)
		}
		devices = append(devices, device)
	}

	return devices, rows.Err()
}

// FindDevice returns a tracked device record.
func (r *IntegrationRepository) FindDevice(userID int64, deviceID string) (erp.Device, error) {
	var device erp.Device
	row := r.db.QueryRow(
		`SELECT id, user_id, device_id, user_agent, ip, first_login_at, last_login_at, is_trusted_device
		   FROM devices
		  WHERE user_id = ? AND device_id = ?
		  LIMIT 1`,
		userID,
		deviceID,
	)
	if err := row.Scan(
		&device.ID,
		&device.UserID,
		&device.DeviceID,
		&device.UserAgent,
		&device.IP,
		&device.FirstLoginAt,
		&device.LastLoginAt,
		&device.IsTrustedDevice,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return erp.Device{}, erp.ErrDeviceNotFound
		}
		return erp.Device{}, fmt.Errorf("find device: %w", err)
	}

	return device, nil
}

// UpsertDeviceLogin inserts a new device or updates last login metadata.
func (r *IntegrationRepository) UpsertDeviceLogin(userID int64, deviceID, userAgent, ip string, trusted bool, now time.Time) error {
	_, err := r.db.Exec(
		`INSERT INTO devices (user_id, device_id, user_agent, ip, first_login_at, last_login_at, is_trusted_device)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		   user_agent = VALUES(user_agent),
		   ip = VALUES(ip),
		   last_login_at = VALUES(last_login_at),
		   is_trusted_device = IF(is_trusted_device = 1, 1, VALUES(is_trusted_device))`,
		userID,
		deviceID,
		userAgent,
		ip,
		now,
		now,
		trusted,
	)
	if err != nil {
		return fmt.Errorf("upsert device login: %w", err)
	}

	return nil
}

// ApproveDevice marks an existing device as trusted.
func (r *IntegrationRepository) ApproveDevice(userID int64, deviceID string, trustedBy int64, now time.Time) error {
	result, err := r.db.Exec(
		`UPDATE devices
		    SET is_trusted_device = 1,
		        trusted_by = ?,
		        updated_at = ?
		  WHERE user_id = ? AND device_id = ?`,
		trustedBy,
		now,
		userID,
		deviceID,
	)
	if err != nil {
		return fmt.Errorf("approve device: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("approve device rows affected: %w", err)
	}
	if affected == 0 {
		return erp.ErrDeviceNotFound
	}

	return nil
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}

	return value
}

func ruleType(rule string) string {
	rule = strings.TrimSpace(rule)
	if strings.Contains(rule, "/") {
		return "cidr"
	}
	if ip := net.ParseIP(rule); ip != nil {
		if ip.To4() != nil {
			return "ipv4"
		}
		return "ipv6"
	}

	return "cidr"
}

func isDuplicate(err error) bool {
	var mysqlErr *mysqlDriver.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062
	}

	return false
}
