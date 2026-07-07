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
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		params.SourceSystem,
		params.ExternalUserID,
		params.DisplayName,
		params.PasswordHash,
		nullableString(params.Email),
		params.Language,
		nullableString(params.WhatsAppAccount),
		nullableString(params.TelegramAccount),
		params.Status,
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

// FindUserByEmail loads a user via email.
func (r *IntegrationRepository) FindUserByEmail(email string) (erp.User, error) {
	var user erp.User
	row := r.db.QueryRow(
		`SELECT id, source_system, external_user_id, display_name, password_hash, COALESCE(email, ''), language,
		        COALESCE(whatsapp_account, ''), COALESCE(telegram_account, ''), status, created_at, updated_at
		   FROM users
		  WHERE LOWER(COALESCE(email, '')) = LOWER(?)
		  LIMIT 1`,
		strings.TrimSpace(email),
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
		return erp.User{}, fmt.Errorf("find user by email: %w", err)
	}

	return user, nil
}

// FindUserInvitationByEmail loads the latest invitation for an email.
func (r *IntegrationRepository) FindUserInvitationByEmail(email string) (erp.UserInvitation, error) {
	var invitation erp.UserInvitation
	var invitedByAdminID sql.NullInt64
	var acceptedUserID sql.NullInt64
	var sentAt sql.NullTime
	var acceptedAt sql.NullTime
	row := r.db.QueryRow(
		`SELECT id, email, token_hash, status, invited_by_admin_id, accepted_user_id,
		        expires_at, sent_at, accepted_at, created_at, updated_at
		   FROM user_invitations
		  WHERE LOWER(email) = LOWER(?)
		  ORDER BY created_at DESC, id DESC
		  LIMIT 1`,
		strings.TrimSpace(email),
	)
	if err := row.Scan(
		&invitation.ID,
		&invitation.Email,
		&invitation.TokenHash,
		&invitation.Status,
		&invitedByAdminID,
		&acceptedUserID,
		&invitation.ExpiresAt,
		&sentAt,
		&acceptedAt,
		&invitation.CreatedAt,
		&invitation.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return erp.UserInvitation{}, erp.ErrUserNotFound
		}
		return erp.UserInvitation{}, fmt.Errorf("find user invitation by email: %w", err)
	}
	if invitedByAdminID.Valid {
		invitation.InvitedByAdminID = invitedByAdminID.Int64
	}
	if acceptedUserID.Valid {
		invitation.AcceptedUserID = acceptedUserID.Int64
	}
	if sentAt.Valid {
		invitation.SentAt = &sentAt.Time
	}
	if acceptedAt.Valid {
		invitation.AcceptedAt = &acceptedAt.Time
	}
	return invitation, nil
}

// CreateUserInvitation inserts a pending invitation.
func (r *IntegrationRepository) CreateUserInvitation(params erp.UserInvitationCreateParams) (erp.UserInvitation, error) {
	result, err := r.db.Exec(
		`INSERT INTO user_invitations (email, token_hash, status, invited_by_admin_id, expires_at, sent_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		params.Email,
		params.TokenHash,
		params.Status,
		nullableInt64(params.InvitedByAdminID),
		params.ExpiresAt,
		nullableTime(params.SentAt),
	)
	if err != nil {
		if isDuplicate(err) {
			return erp.UserInvitation{}, erp.ErrUserInvitationPending
		}
		return erp.UserInvitation{}, fmt.Errorf("create user invitation: %w", err)
	}

	invitationID, err := result.LastInsertId()
	if err != nil {
		return erp.UserInvitation{}, fmt.Errorf("read user invitation id: %w", err)
	}
	return r.findUserInvitationByID(invitationID)
}

// RefreshUserInvitation replaces the token on an existing unsent pending invitation.
func (r *IntegrationRepository) RefreshUserInvitation(params erp.UserInvitationRefreshParams) (erp.UserInvitation, error) {
	result, err := r.db.Exec(
		`UPDATE user_invitations
		    SET token_hash = ?,
		        status = 'pending',
		        invited_by_admin_id = ?,
		        accepted_user_id = NULL,
		        expires_at = ?,
		        sent_at = NULL,
		        accepted_at = NULL,
		        updated_at = CURRENT_TIMESTAMP
		  WHERE id = ? AND status = 'pending'`,
		params.TokenHash,
		nullableInt64(params.InvitedByAdminID),
		params.ExpiresAt,
		params.ID,
	)
	if err != nil {
		if isDuplicate(err) {
			return erp.UserInvitation{}, erp.ErrUserInvitationPending
		}
		return erp.UserInvitation{}, fmt.Errorf("refresh user invitation: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return erp.UserInvitation{}, fmt.Errorf("refresh user invitation rows affected: %w", err)
	}
	if affected == 0 {
		return erp.UserInvitation{}, erp.ErrUserNotFound
	}

	return r.findUserInvitationByID(params.ID)
}

// MarkUserInvitationSent records successful email delivery.
func (r *IntegrationRepository) MarkUserInvitationSent(invitationID int64, sentAt time.Time) error {
	result, err := r.db.Exec(
		`UPDATE user_invitations SET sent_at = ? WHERE id = ?`,
		sentAt,
		invitationID,
	)
	if err != nil {
		return fmt.Errorf("mark user invitation sent: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("mark user invitation sent rows affected: %w", err)
	}
	if affected == 0 {
		return erp.ErrUserNotFound
	}
	return nil
}

func (r *IntegrationRepository) findUserInvitationByID(invitationID int64) (erp.UserInvitation, error) {
	var invitation erp.UserInvitation
	var invitedByAdminID sql.NullInt64
	var acceptedUserID sql.NullInt64
	var sentAt sql.NullTime
	var acceptedAt sql.NullTime
	row := r.db.QueryRow(
		`SELECT id, email, token_hash, status, invited_by_admin_id, accepted_user_id,
		        expires_at, sent_at, accepted_at, created_at, updated_at
		   FROM user_invitations
		  WHERE id = ?
		  LIMIT 1`,
		invitationID,
	)
	if err := row.Scan(
		&invitation.ID,
		&invitation.Email,
		&invitation.TokenHash,
		&invitation.Status,
		&invitedByAdminID,
		&acceptedUserID,
		&invitation.ExpiresAt,
		&sentAt,
		&acceptedAt,
		&invitation.CreatedAt,
		&invitation.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return erp.UserInvitation{}, erp.ErrUserNotFound
		}
		return erp.UserInvitation{}, fmt.Errorf("find user invitation by id: %w", err)
	}
	if invitedByAdminID.Valid {
		invitation.InvitedByAdminID = invitedByAdminID.Int64
	}
	if acceptedUserID.Valid {
		invitation.AcceptedUserID = acceptedUserID.Int64
	}
	if sentAt.Valid {
		invitation.SentAt = &sentAt.Time
	}
	if acceptedAt.Valid {
		invitation.AcceptedAt = &acceptedAt.Time
	}
	return invitation, nil
}

// ListUsers returns one page of users and their most recent authenticated activity.
func (r *IntegrationRepository) ListUsers(filter erp.AdminUserFilter) (erp.AdminUserPage, error) {
	page := filter.Page
	if page < 1 {
		page = 1
	}
	perPage := filter.PerPage
	if perPage < 1 {
		perPage = 10
	}

	var where strings.Builder
	where.WriteString(" WHERE 1 = 1")
	args := make([]any, 0, 7)
	addLikeFilter := func(column, value string) {
		if value == "" {
			return
		}
		where.WriteString(" AND " + column + " LIKE ?")
		args = append(args, "%"+value+"%")
	}
	addLikeFilter("u.external_user_id", strings.TrimSpace(filter.ExternalUserID))
	addLikeFilter("u.display_name", strings.TrimSpace(filter.DisplayName))
	addLikeFilter("COALESCE(u.email, '')", strings.TrimSpace(filter.Email))

	if value := strings.TrimSpace(filter.Status); value != "" {
		where.WriteString(" AND u.status = ?")
		args = append(args, value)
	}
	if value := strings.TrimSpace(filter.SourceSystem); value != "" {
		where.WriteString(" AND u.source_system = ?")
		args = append(args, value)
	}
	if filter.CreatedFrom != nil {
		where.WriteString(" AND u.created_at >= ?")
		args = append(args, *filter.CreatedFrom)
	}
	if filter.CreatedTo != nil {
		where.WriteString(" AND u.created_at <= ?")
		args = append(args, *filter.CreatedTo)
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM users u"+where.String(), args...).Scan(&total); err != nil {
		return erp.AdminUserPage{}, fmt.Errorf("count users: %w", err)
	}

	var query strings.Builder
	query.WriteString(
		`SELECT u.id, u.external_user_id, u.display_name, COALESCE(u.email, ''), u.status,
		        u.source_system, u.created_at, MAX(s.last_used_at) AS last_online_at
		   FROM users u
		   LEFT JOIN auth_sessions s ON s.user_id = u.id`,
	)
	query.WriteString(where.String())

	query.WriteString(" GROUP BY u.id, u.external_user_id, u.display_name, u.email, u.status, u.source_system, u.created_at")
	switch strings.TrimSpace(filter.Sort) {
	case "created_at_asc":
		query.WriteString(" ORDER BY u.created_at ASC, u.id ASC")
	case "last_online_desc":
		query.WriteString(" ORDER BY last_online_at DESC, u.id DESC")
	default:
		query.WriteString(" ORDER BY u.created_at DESC, u.id DESC")
	}
	query.WriteString(" LIMIT ? OFFSET ?")
	queryArgs := append(append([]any{}, args...), perPage, (page-1)*perPage)

	rows, err := r.db.Query(query.String(), queryArgs...)
	if err != nil {
		return erp.AdminUserPage{}, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := make([]erp.AdminUserSummary, 0)
	for rows.Next() {
		var user erp.AdminUserSummary
		var lastOnline sql.NullTime
		if err := rows.Scan(
			&user.ID,
			&user.ExternalUserID,
			&user.DisplayName,
			&user.Email,
			&user.Status,
			&user.SourceSystem,
			&user.CreatedAt,
			&lastOnline,
		); err != nil {
			return erp.AdminUserPage{}, fmt.Errorf("scan user list: %w", err)
		}
		if lastOnline.Valid {
			user.LastOnlineAt = &lastOnline.Time
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return erp.AdminUserPage{}, fmt.Errorf("iterate user list: %w", err)
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + perPage - 1) / perPage
	}
	return erp.AdminUserPage{
		Items:      users,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

// ListSystemAdmins returns one page of backend admin accounts.
func (r *IntegrationRepository) ListSystemAdmins(filter erp.SystemAdminFilter) (erp.SystemAdminPage, error) {
	page := filter.Page
	if page < 1 {
		page = 1
	}
	perPage := filter.PerPage
	if perPage < 1 {
		perPage = 10
	}

	var where strings.Builder
	where.WriteString(" WHERE 1 = 1")
	args := make([]any, 0, 3)
	addLikeFilter := func(column, value string) {
		if value == "" {
			return
		}
		where.WriteString(" AND " + column + " LIKE ?")
		args = append(args, "%"+value+"%")
	}
	addLikeFilter("a.account", strings.TrimSpace(filter.ExternalUserID))
	addLikeFilter("a.display_name", strings.TrimSpace(filter.DisplayName))
	if value := strings.TrimSpace(filter.Status); value != "" {
		where.WriteString(" AND a.status = ?")
		args = append(args, value)
	}

	from := ` FROM admin_users a`

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*)"+from+where.String(), args...).Scan(&total); err != nil {
		return erp.SystemAdminPage{}, fmt.Errorf("count system admins: %w", err)
	}

	query := `SELECT a.id, a.id, a.account, a.display_name, a.password_hash, a.role, a.status,
	                 a.last_login_at, COALESCE(a.last_login_ip, '')
	            ` + from + where.String() + `
	        ORDER BY a.created_at DESC, a.id DESC
	           LIMIT ? OFFSET ?`
	queryArgs := append(append([]any{}, args...), perPage, (page-1)*perPage)

	rows, err := r.db.Query(query, queryArgs...)
	if err != nil {
		return erp.SystemAdminPage{}, fmt.Errorf("list system admins: %w", err)
	}
	defer rows.Close()

	admins := make([]erp.SystemAdminSummary, 0)
	for rows.Next() {
		var admin erp.SystemAdminSummary
		var lastLogin sql.NullTime
		if err := rows.Scan(
			&admin.ID,
			&admin.UserID,
			&admin.ExternalUserID,
			&admin.DisplayName,
			&admin.PasswordHash,
			&admin.Role,
			&admin.Status,
			&lastLogin,
			&admin.LastLoginIP,
		); err != nil {
			return erp.SystemAdminPage{}, fmt.Errorf("scan system admin list: %w", err)
		}
		admin.AdminUserID = admin.UserID
		if lastLogin.Valid {
			admin.LastLoginAt = &lastLogin.Time
		}
		admin.RoleName = systemAdminRoleName(admin.Role)
		admins = append(admins, admin)
	}
	if err := rows.Err(); err != nil {
		return erp.SystemAdminPage{}, fmt.Errorf("iterate system admin list: %w", err)
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + perPage - 1) / perPage
	}
	return erp.SystemAdminPage{
		Items:      admins,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

// FindSystemAdminByID loads a backend admin by numeric id.
func (r *IntegrationRepository) FindSystemAdminByID(adminUserID int64) (erp.SystemAdminSummary, error) {
	return r.findSystemAdmin("a.id = ?", adminUserID)
}

// FindSystemAdminByAccount loads a backend admin by login account.
func (r *IntegrationRepository) FindSystemAdminByAccount(account string) (erp.SystemAdminSummary, error) {
	return r.findSystemAdmin("a.account = ?", strings.TrimSpace(account))
}

func (r *IntegrationRepository) findSystemAdmin(predicate string, arg any) (erp.SystemAdminSummary, error) {
	var admin erp.SystemAdminSummary
	var lastLogin sql.NullTime
	row := r.db.QueryRow(
		`SELECT a.id, a.id, a.account, a.display_name, a.password_hash, a.role, a.status,
		        a.last_login_at, COALESCE(a.last_login_ip, '')
		   FROM admin_users a
		  WHERE `+predicate+`
		  LIMIT 1`,
		arg,
	)
	if err := row.Scan(
		&admin.ID,
		&admin.UserID,
		&admin.ExternalUserID,
		&admin.DisplayName,
		&admin.PasswordHash,
		&admin.Role,
		&admin.Status,
		&lastLogin,
		&admin.LastLoginIP,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return erp.SystemAdminSummary{}, erp.ErrUserNotFound
		}
		return erp.SystemAdminSummary{}, fmt.Errorf("find system admin: %w", err)
	}
	admin.AdminUserID = admin.UserID
	admin.RoleName = systemAdminRoleName(admin.Role)
	if lastLogin.Valid {
		admin.LastLoginAt = &lastLogin.Time
	}
	return admin, nil
}

// CreateSystemAdmin creates a backend admin account.
func (r *IntegrationRepository) CreateSystemAdmin(params erp.SystemAdminCreateParams) (erp.SystemAdminSummary, error) {
	result, err := r.db.Exec(
		`INSERT INTO admin_users (account, display_name, password_hash, role, status, created_by)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		params.Account,
		params.DisplayName,
		params.PasswordHash,
		params.Role,
		params.Status,
		nullableInt64(params.CreatedBy),
	)
	if err != nil {
		if isDuplicate(err) {
			return erp.SystemAdminSummary{}, erp.ErrUserAlreadyExists
		}
		return erp.SystemAdminSummary{}, fmt.Errorf("create system admin: %w", err)
	}

	adminUserID, err := result.LastInsertId()
	if err != nil {
		return erp.SystemAdminSummary{}, fmt.Errorf("read system admin id: %w", err)
	}

	return r.FindSystemAdminByID(adminUserID)
}

// UpdateSystemAdmin applies mutable backend admin fields.
func (r *IntegrationRepository) UpdateSystemAdmin(adminUserID int64, params erp.SystemAdminUpdateParams) (erp.SystemAdminSummary, error) {
	var err error
	if params.PasswordHash == "" {
		_, err = r.db.Exec(
			`UPDATE admin_users
			    SET display_name = ?, role = ?, status = ?
			  WHERE id = ?`,
			params.DisplayName,
			params.Role,
			params.Status,
			adminUserID,
		)
	} else {
		_, err = r.db.Exec(
			`UPDATE admin_users
			    SET display_name = ?, role = ?, status = ?, password_hash = ?
			  WHERE id = ?`,
			params.DisplayName,
			params.Role,
			params.Status,
			params.PasswordHash,
			adminUserID,
		)
	}
	if err != nil {
		if isDuplicate(err) {
			return erp.SystemAdminSummary{}, erp.ErrUserAlreadyExists
		}
		return erp.SystemAdminSummary{}, fmt.Errorf("update system admin: %w", err)
	}

	return r.FindSystemAdminByID(adminUserID)
}

// UpdateSystemAdminLastLogin records the last successful backend admin login.
func (r *IntegrationRepository) UpdateSystemAdminLastLogin(adminUserID int64, ip string, at time.Time) error {
	_, err := r.db.Exec(
		`UPDATE admin_users
		    SET last_login_at = ?, last_login_ip = ?
		  WHERE id = ?`,
		at,
		nullableString(strings.TrimSpace(ip)),
		adminUserID,
	)
	if err != nil {
		return fmt.Errorf("update system admin last login: %w", err)
	}
	return nil
}

// UpdateUser applies mutable admin-managed fields. An empty password hash keeps the existing password.
func (r *IntegrationRepository) UpdateUser(userID int64, params erp.AdminUpdateUserParams) (erp.User, error) {
	var err error
	if params.PasswordHash == "" {
		_, err = r.db.Exec(
			`UPDATE users
			    SET display_name = ?, email = ?, status = ?
			  WHERE id = ?`,
			params.DisplayName,
			nullableString(params.Email),
			params.Status,
			userID,
		)
	} else {
		_, err = r.db.Exec(
			`UPDATE users
			    SET display_name = ?, email = ?, status = ?, password_hash = ?
			  WHERE id = ?`,
			params.DisplayName,
			nullableString(params.Email),
			params.Status,
			params.PasswordHash,
			userID,
		)
	}
	if err != nil {
		if isDuplicate(err) {
			return erp.User{}, erp.ErrUserAlreadyExists
		}
		return erp.User{}, fmt.Errorf("update user: %w", err)
	}

	return r.FindUserByID(userID)
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

func nullableInt64(value int64) any {
	if value <= 0 {
		return nil
	}

	return value
}

func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}

	return *value
}

func systemAdminRoleName(role string) string {
	switch strings.TrimSpace(role) {
	case "system_admin":
		return "系統管理員"
	default:
		return strings.TrimSpace(role)
	}
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
