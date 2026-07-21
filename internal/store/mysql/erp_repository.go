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
	if _, err := r.FindUserByExternalID(params.ExternalUserID); err == nil {
		return erp.User{}, erp.ErrUserAlreadyExists
	} else if !errors.Is(err, erp.ErrUserNotFound) {
		return erp.User{}, err
	}

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
		if isDuplicateKey(err, "uk_users_email") {
			return erp.User{}, erp.ErrEmailAlreadyExists
		}
		if isDuplicateKey(err, "uk_users_external_user_id") {
			return erp.User{}, erp.ErrUserAlreadyExists
		}
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
		        COALESCE(whatsapp_account, ''), COALESCE(telegram_account, ''), status, is_chat_muted,
		        must_change_password, temporary_password_expires_at, created_at, updated_at
		   FROM users
		  WHERE id = ?
		  LIMIT 1`,
		userID,
	)
	var temporaryPasswordExpiresAt sql.NullTime
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
		&user.IsChatMuted,
		&user.MustChangePassword,
		&temporaryPasswordExpiresAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return erp.User{}, erp.ErrUserNotFound
		}
		return erp.User{}, fmt.Errorf("find user by id: %w", err)
	}
	if temporaryPasswordExpiresAt.Valid {
		user.TemporaryPasswordExpiresAt = &temporaryPasswordExpiresAt.Time
	}

	return user, nil
}

// FindUserByExternal loads a user via source_system + external_user_id.
func (r *IntegrationRepository) FindUserByExternal(sourceSystem, externalUserID string) (erp.User, error) {
	var user erp.User
	row := r.db.QueryRow(
		`SELECT id, source_system, external_user_id, display_name, password_hash, COALESCE(email, ''), language,
		        COALESCE(whatsapp_account, ''), COALESCE(telegram_account, ''), status, is_chat_muted,
		        must_change_password, temporary_password_expires_at, created_at, updated_at
		   FROM users
		  WHERE source_system = ? AND external_user_id = ?
		  LIMIT 1`,
		sourceSystem,
		externalUserID,
	)
	var temporaryPasswordExpiresAt sql.NullTime
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
		&user.IsChatMuted,
		&user.MustChangePassword,
		&temporaryPasswordExpiresAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return erp.User{}, erp.ErrUserNotFound
		}
		return erp.User{}, fmt.Errorf("find user: %w", err)
	}
	if temporaryPasswordExpiresAt.Valid {
		user.TemporaryPasswordExpiresAt = &temporaryPasswordExpiresAt.Time
	}

	return user, nil
}

// FindUserByExternalID loads a user by account across all source systems.
func (r *IntegrationRepository) FindUserByExternalID(externalUserID string) (erp.User, error) {
	var user erp.User
	row := r.db.QueryRow(
		`SELECT id, source_system, external_user_id, display_name, password_hash, COALESCE(email, ''), language,
		        COALESCE(whatsapp_account, ''), COALESCE(telegram_account, ''), status, is_chat_muted,
		        must_change_password, temporary_password_expires_at, created_at, updated_at
		   FROM users
		  WHERE external_user_id = ?
		  ORDER BY id
		  LIMIT 1`,
		externalUserID,
	)
	var temporaryPasswordExpiresAt sql.NullTime
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
		&user.IsChatMuted,
		&user.MustChangePassword,
		&temporaryPasswordExpiresAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return erp.User{}, erp.ErrUserNotFound
		}
		return erp.User{}, fmt.Errorf("find user by account: %w", err)
	}
	if temporaryPasswordExpiresAt.Valid {
		user.TemporaryPasswordExpiresAt = &temporaryPasswordExpiresAt.Time
	}

	return user, nil
}

// FindUserByEmail loads a user via email.
func (r *IntegrationRepository) FindUserByEmail(email string) (erp.User, error) {
	var user erp.User
	row := r.db.QueryRow(
		`SELECT id, source_system, external_user_id, display_name, password_hash, COALESCE(email, ''), language,
		        COALESCE(whatsapp_account, ''), COALESCE(telegram_account, ''), status, is_chat_muted,
		        must_change_password, temporary_password_expires_at, created_at, updated_at
		   FROM users
		  WHERE LOWER(COALESCE(email, '')) = LOWER(?)
		  LIMIT 1`,
		strings.TrimSpace(email),
	)
	var temporaryPasswordExpiresAt sql.NullTime
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
		&user.IsChatMuted,
		&user.MustChangePassword,
		&temporaryPasswordExpiresAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return erp.User{}, erp.ErrUserNotFound
		}
		return erp.User{}, fmt.Errorf("find user by email: %w", err)
	}
	if temporaryPasswordExpiresAt.Valid {
		user.TemporaryPasswordExpiresAt = &temporaryPasswordExpiresAt.Time
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

// FindUserInvitationByTokenHash loads an invitation by its hashed token.
func (r *IntegrationRepository) FindUserInvitationByTokenHash(tokenHash string) (erp.UserInvitation, error) {
	var invitation erp.UserInvitation
	var invitedByAdminID sql.NullInt64
	var acceptedUserID sql.NullInt64
	var sentAt sql.NullTime
	var acceptedAt sql.NullTime
	row := r.db.QueryRow(
		`SELECT id, email, token_hash, status, invited_by_admin_id, accepted_user_id,
		        expires_at, sent_at, accepted_at, created_at, updated_at
		   FROM user_invitations
		  WHERE token_hash = ?
		  LIMIT 1`,
		strings.TrimSpace(tokenHash),
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
		return erp.UserInvitation{}, fmt.Errorf("find user invitation by token hash: %w", err)
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

// RefreshUserInvitation replaces the token on an unfinished invitation.
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
		  WHERE id = ? AND status <> 'accepted'`,
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

// AcceptUserInvitation creates the invited user and marks the invitation accepted atomically.
func (r *IntegrationRepository) AcceptUserInvitation(params erp.AcceptUserInvitationParams) (erp.User, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return erp.User{}, fmt.Errorf("begin accept user invitation: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var invitation erp.UserInvitation
	var invitedByAdminID sql.NullInt64
	var acceptedUserID sql.NullInt64
	var sentAt sql.NullTime
	var acceptedAt sql.NullTime
	row := tx.QueryRow(
		`SELECT id, email, token_hash, status, invited_by_admin_id, accepted_user_id,
		        expires_at, sent_at, accepted_at, created_at, updated_at
		   FROM user_invitations
		  WHERE token_hash = ?
		  LIMIT 1
		    FOR UPDATE`,
		strings.TrimSpace(params.TokenHash),
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
			return erp.User{}, erp.ErrUserInvitationNotFound
		}
		return erp.User{}, fmt.Errorf("lock user invitation: %w", err)
	}
	if acceptedUserID.Valid {
		invitation.AcceptedUserID = acceptedUserID.Int64
	}
	if acceptedAt.Valid {
		invitation.AcceptedAt = &acceptedAt.Time
	}
	if invitation.Status != "pending" ||
		!params.AcceptedAt.Before(invitation.ExpiresAt) ||
		invitation.AcceptedAt != nil ||
		invitation.AcceptedUserID > 0 {
		return erp.User{}, erp.ErrUserInvitationNotFound
	}

	var existingEmailUserID int64
	emailRow := tx.QueryRow(
		`SELECT id
		   FROM users
		  WHERE LOWER(COALESCE(email, '')) = LOWER(?)
		  LIMIT 1
		    FOR UPDATE`,
		strings.TrimSpace(invitation.Email),
	)
	if err := emailRow.Scan(&existingEmailUserID); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return erp.User{}, fmt.Errorf("lock user by email: %w", err)
		}
	} else {
		return erp.User{}, erp.ErrEmailAlreadyExists
	}

	var existingAccountUserID int64
	accountRow := tx.QueryRow(
		`SELECT id
		   FROM users
		  WHERE external_user_id = ?
		  LIMIT 1
		    FOR UPDATE`,
		strings.TrimSpace(params.ExternalUserID),
	)
	if err := accountRow.Scan(&existingAccountUserID); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return erp.User{}, fmt.Errorf("lock user by account: %w", err)
		}
	} else {
		return erp.User{}, erp.ErrUserAlreadyExists
	}

	result, err := tx.Exec(
		`INSERT INTO users (source_system, external_user_id, display_name, password_hash, email, language, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		params.SourceSystem,
		params.ExternalUserID,
		params.DisplayName,
		params.PasswordHash,
		nullableString(invitation.Email),
		"zh-Hans",
		params.Status,
	)
	if err != nil {
		if isDuplicateKey(err, "uk_users_email") {
			return erp.User{}, erp.ErrEmailAlreadyExists
		}
		if isDuplicateKey(err, "uk_users_external_user_id") {
			return erp.User{}, erp.ErrUserAlreadyExists
		}
		if isDuplicate(err) {
			return erp.User{}, erp.ErrUserAlreadyExists
		}
		return erp.User{}, fmt.Errorf("create invited user: %w", err)
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return erp.User{}, fmt.Errorf("read invited user id: %w", err)
	}

	if _, err := tx.Exec(`INSERT INTO user_security_settings (user_id, allow_all_ips) VALUES (?, 1)`, userID); err != nil {
		return erp.User{}, fmt.Errorf("create invited user security settings: %w", err)
	}

	updateResult, err := tx.Exec(
		`UPDATE user_invitations
		    SET status = 'accepted',
		        accepted_user_id = ?,
		        accepted_at = ?,
		        updated_at = CURRENT_TIMESTAMP
		  WHERE id = ?
		    AND status = 'pending'
		    AND accepted_user_id IS NULL
		    AND accepted_at IS NULL`,
		userID,
		params.AcceptedAt,
		invitation.ID,
	)
	if err != nil {
		return erp.User{}, fmt.Errorf("mark user invitation accepted: %w", err)
	}
	affected, err := updateResult.RowsAffected()
	if err != nil {
		return erp.User{}, fmt.Errorf("mark user invitation accepted rows affected: %w", err)
	}
	if affected == 0 {
		return erp.User{}, erp.ErrUserInvitationNotFound
	}

	user, err := findUserByIDTx(tx, userID)
	if err != nil {
		return erp.User{}, err
	}

	if err := tx.Commit(); err != nil {
		return erp.User{}, fmt.Errorf("commit accept user invitation: %w", err)
	}
	return user, nil
}

// ListUserInvitations returns one page of registration invitations.
func (r *IntegrationRepository) ListUserInvitations(filter erp.AdminUserInvitationFilter) (erp.AdminUserInvitationPage, error) {
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
	args := make([]any, 0, 1)
	if email := strings.TrimSpace(filter.Email); email != "" {
		where.WriteString(" AND ui.email LIKE ?")
		args = append(args, "%"+email+"%")
	}
	if status := strings.TrimSpace(filter.Status); status != "" && validUserInvitationStatus(status) {
		where.WriteString(" AND ui.status = ?")
		args = append(args, status)
	}

	from := ` FROM user_invitations ui
	          LEFT JOIN admin_users a ON a.id = ui.invited_by_admin_id`

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*)"+from+where.String(), args...).Scan(&total); err != nil {
		return erp.AdminUserInvitationPage{}, fmt.Errorf("count user invitations: %w", err)
	}

	query := `SELECT ui.id, ui.email, ui.status, ui.invited_by_admin_id,
	                 COALESCE(a.account, ''), COALESCE(a.display_name, ''),
	                 ui.accepted_user_id, ui.expires_at, ui.sent_at, ui.accepted_at,
	                 ui.created_at, ui.updated_at
	            ` + from + where.String() + `
	        ` + userInvitationOrderBy(filter.Sorting) + `
	           LIMIT ? OFFSET ?`
	queryArgs := append(append([]any{}, args...), perPage, (page-1)*perPage)

	rows, err := r.db.Query(query, queryArgs...)
	if err != nil {
		return erp.AdminUserInvitationPage{}, fmt.Errorf("list user invitations: %w", err)
	}
	defer rows.Close()

	invitations := make([]erp.AdminUserInvitationSummary, 0)
	for rows.Next() {
		var invitation erp.AdminUserInvitationSummary
		var invitedByAdminID sql.NullInt64
		var acceptedUserID sql.NullInt64
		var sentAt sql.NullTime
		var acceptedAt sql.NullTime
		if err := rows.Scan(
			&invitation.ID,
			&invitation.Email,
			&invitation.Status,
			&invitedByAdminID,
			&invitation.InvitedByAdminAccount,
			&invitation.InvitedByAdminName,
			&acceptedUserID,
			&invitation.ExpiresAt,
			&sentAt,
			&acceptedAt,
			&invitation.CreatedAt,
			&invitation.UpdatedAt,
		); err != nil {
			return erp.AdminUserInvitationPage{}, fmt.Errorf("scan user invitation list: %w", err)
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
		invitations = append(invitations, invitation)
	}
	if err := rows.Err(); err != nil {
		return erp.AdminUserInvitationPage{}, fmt.Errorf("iterate user invitation list: %w", err)
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + perPage - 1) / perPage
	}
	return erp.AdminUserInvitationPage{
		Items:      invitations,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

// PrecheckUserInvitationEmails checks existing users and active invitations for a batch of emails.
func (r *IntegrationRepository) PrecheckUserInvitationEmails(emails []string, now time.Time) (erp.UserInvitationEmailPrecheck, error) {
	result := erp.UserInvitationEmailPrecheck{
		Registered:        map[string]bool{},
		ActiveInvitations: map[string]bool{},
	}
	if len(emails) == 0 {
		return result, nil
	}

	placeholders := queryPlaceholders(len(emails))
	userArgs := make([]any, 0, len(emails))
	for _, email := range emails {
		userArgs = append(userArgs, strings.ToLower(strings.TrimSpace(email)))
	}
	userRows, err := r.db.Query(
		`SELECT DISTINCT LOWER(email)
		   FROM users
		  WHERE email IS NOT NULL
		    AND LOWER(email) IN (`+placeholders+`)`,
		userArgs...,
	)
	if err != nil {
		return erp.UserInvitationEmailPrecheck{}, fmt.Errorf("precheck invitation registered emails: %w", err)
	}
	defer userRows.Close()
	for userRows.Next() {
		var email string
		if err := userRows.Scan(&email); err != nil {
			return erp.UserInvitationEmailPrecheck{}, fmt.Errorf("scan precheck registered email: %w", err)
		}
		result.Registered[email] = true
	}
	if err := userRows.Err(); err != nil {
		return erp.UserInvitationEmailPrecheck{}, fmt.Errorf("iterate precheck registered emails: %w", err)
	}

	invitationArgs := make([]any, 0, len(emails)+1)
	invitationArgs = append(invitationArgs, now)
	for _, email := range emails {
		invitationArgs = append(invitationArgs, strings.ToLower(strings.TrimSpace(email)))
	}
	invitationRows, err := r.db.Query(
		`SELECT DISTINCT LOWER(email)
		   FROM user_invitations
		  WHERE status = 'pending'
		    AND accepted_user_id IS NULL
		    AND accepted_at IS NULL
		    AND expires_at > ?
		    AND LOWER(email) IN (`+placeholders+`)`,
		invitationArgs...,
	)
	if err != nil {
		return erp.UserInvitationEmailPrecheck{}, fmt.Errorf("precheck active invitations: %w", err)
	}
	defer invitationRows.Close()
	for invitationRows.Next() {
		var email string
		if err := invitationRows.Scan(&email); err != nil {
			return erp.UserInvitationEmailPrecheck{}, fmt.Errorf("scan precheck active invitation: %w", err)
		}
		result.ActiveInvitations[email] = true
	}
	if err := invitationRows.Err(); err != nil {
		return erp.UserInvitationEmailPrecheck{}, fmt.Errorf("iterate precheck active invitations: %w", err)
	}

	return result, nil
}

func validUserInvitationStatus(status string) bool {
	switch status {
	case "pending", "accepted", "expired":
		return true
	default:
		return false
	}
}

func queryPlaceholders(count int) string {
	if count <= 0 {
		return ""
	}
	return strings.TrimRight(strings.Repeat("?,", count), ",")
}

func userInvitationOrderBy(sorting string) string {
	switch strings.TrimSpace(sorting) {
	case "expires_at":
		return "ORDER BY ui.expires_at DESC, ui.id DESC"
	case "accepted_at":
		return "ORDER BY ui.accepted_at IS NULL, ui.accepted_at DESC, ui.id DESC"
	default:
		return "ORDER BY ui.sent_at IS NULL, ui.sent_at DESC, ui.created_at DESC, ui.id DESC"
	}
}

func findUserByIDTx(tx *sql.Tx, userID int64) (erp.User, error) {
	var user erp.User
	row := tx.QueryRow(
		`SELECT id, source_system, external_user_id, display_name, password_hash, COALESCE(email, ''), language,
		        COALESCE(whatsapp_account, ''), COALESCE(telegram_account, ''), status, is_chat_muted,
		        must_change_password, temporary_password_expires_at, created_at, updated_at
		   FROM users
		  WHERE id = ?
		  LIMIT 1`,
		userID,
	)
	var temporaryPasswordExpiresAt sql.NullTime
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
		&user.IsChatMuted,
		&user.MustChangePassword,
		&temporaryPasswordExpiresAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return erp.User{}, erp.ErrUserNotFound
		}
		return erp.User{}, fmt.Errorf("find user by id: %w", err)
	}
	if temporaryPasswordExpiresAt.Valid {
		user.TemporaryPasswordExpiresAt = &temporaryPasswordExpiresAt.Time
	}

	return user, nil
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
		`SELECT u.id, u.external_user_id, u.display_name, COALESCE(u.email, ''), u.status, u.is_chat_muted,
		        u.must_change_password, u.temporary_password_expires_at, u.source_system, u.created_at, MAX(s.last_used_at) AS last_online_at
		   FROM users u
		   LEFT JOIN auth_sessions s ON s.user_id = u.id`,
	)
	query.WriteString(where.String())

	query.WriteString(" GROUP BY u.id, u.external_user_id, u.display_name, u.email, u.status, u.is_chat_muted, u.must_change_password, u.temporary_password_expires_at, u.source_system, u.created_at")
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
		var temporaryPasswordExpiresAt sql.NullTime
		if err := rows.Scan(
			&user.ID,
			&user.ExternalUserID,
			&user.DisplayName,
			&user.Email,
			&user.Status,
			&user.IsChatMuted,
			&user.MustChangePassword,
			&temporaryPasswordExpiresAt,
			&user.SourceSystem,
			&user.CreatedAt,
			&lastOnline,
		); err != nil {
			return erp.AdminUserPage{}, fmt.Errorf("scan user list: %w", err)
		}
		if lastOnline.Valid {
			user.LastOnlineAt = &lastOnline.Time
		}
		if temporaryPasswordExpiresAt.Valid {
			user.TemporaryPasswordExpiresAt = &temporaryPasswordExpiresAt.Time
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

// ListAdminConversations returns one read-only page of chat conversations for backend admins.
func (r *IntegrationRepository) ListAdminConversations(filter erp.AdminConversationFilter) (erp.AdminConversationPage, error) {
	page := filter.Page
	if page < 1 {
		page = 1
	}
	perPage := filter.PerPage
	if perPage < 1 {
		perPage = 10
	}

	from := adminConversationsFromSQL()
	activityExpr := "COALESCE(lm.created_at, c.created_at)"
	var where strings.Builder
	where.WriteString(" WHERE 1 = 1")
	args := make([]any, 0, 6)
	if value := strings.TrimSpace(filter.Type); value != "" {
		where.WriteString(" AND c.type = ?")
		args = append(args, value)
	}
	if value := strings.TrimSpace(filter.Keyword); value != "" {
		like := "%" + value + "%"
		where.WriteString(` AND (
			COALESCE(c.name, '') LIKE ?
			OR EXISTS (
				SELECT 1
				  FROM conversation_members search_cm
				  JOIN users search_u ON search_u.id = search_cm.user_id
				 WHERE search_cm.conversation_id = c.id
				   AND (search_u.external_user_id LIKE ? OR search_u.display_name LIKE ?)
			)
		)`)
		args = append(args, like, like, like)
	}
	if filter.LastActivityFrom != nil {
		where.WriteString(" AND " + activityExpr + " >= ?")
		args = append(args, *filter.LastActivityFrom)
	}
	if filter.LastActivityTo != nil {
		where.WriteString(" AND " + activityExpr + " <= ?")
		args = append(args, *filter.LastActivityTo)
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) "+from+where.String(), args...).Scan(&total); err != nil {
		return erp.AdminConversationPage{}, fmt.Errorf("count admin conversations: %w", err)
	}

	query := `SELECT c.id, c.type,
	                 CASE
	                   WHEN c.type = 'direct' THEN COALESCE(NULLIF(ms.member_names, ''), CONCAT('聊天室#', c.id))
	                   ELSE COALESCE(NULLIF(c.name, ''), CONCAT('群組#', c.id))
	                 END AS display_name,
	                 CASE WHEN c.type = 'direct' THEN 2 ELSE COALESCE(ms.member_count, 0) END AS member_count,
	                 lm.id AS last_message_id,
	                 COALESCE(lm.message_type, '') AS last_message_type,
	                 COALESCE(lm.content, '') AS last_message_content,
	                 lm.created_at AS last_message_at,
	                 last_sender.id AS last_sender_id,
	                 COALESCE(last_sender.external_user_id, '') AS last_sender_account,
	                 COALESCE(last_sender.display_name, '') AS last_sender_name,
	                 ` + activityExpr + ` AS last_activity_at,
	                 c.created_at
	            ` + from + where.String() + `
	        ORDER BY last_activity_at DESC, c.id DESC
	           LIMIT ? OFFSET ?`
	queryArgs := append(append([]any{}, args...), perPage, (page-1)*perPage)

	rows, err := r.db.Query(query, queryArgs...)
	if err != nil {
		return erp.AdminConversationPage{}, fmt.Errorf("list admin conversations: %w", err)
	}
	defer rows.Close()

	conversations := make([]erp.AdminConversationSummary, 0)
	for rows.Next() {
		var item erp.AdminConversationSummary
		var lastMessageID sql.NullInt64
		var lastMessageAt sql.NullTime
		var lastSenderID sql.NullInt64
		var lastMessageContent string
		if err := rows.Scan(
			&item.ID,
			&item.Type,
			&item.Name,
			&item.MemberCount,
			&lastMessageID,
			&item.LastMessageType,
			&lastMessageContent,
			&lastMessageAt,
			&lastSenderID,
			&item.LastSenderAccount,
			&item.LastSenderName,
			&item.LastActivityAt,
			&item.CreatedAt,
		); err != nil {
			return erp.AdminConversationPage{}, fmt.Errorf("scan admin conversation list: %w", err)
		}
		if lastMessageID.Valid {
			item.LatestMessageAvailable = true
			item.LastMessage = adminLastMessagePreview(item.LastMessageType, lastMessageContent)
		}
		if lastMessageAt.Valid {
			item.LastMessageAt = &lastMessageAt.Time
		}
		if lastSenderID.Valid {
			item.LastSenderID = lastSenderID.Int64
		}
		conversations = append(conversations, item)
	}
	if err := rows.Err(); err != nil {
		return erp.AdminConversationPage{}, fmt.Errorf("iterate admin conversation list: %w", err)
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + perPage - 1) / perPage
	}
	return erp.AdminConversationPage{
		Items:      conversations,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

// GetAdminConversationDetail returns one read-only conversation record with one page of messages.
func (r *IntegrationRepository) GetAdminConversationDetail(filter erp.AdminConversationMessageFilter) (erp.AdminConversationDetail, error) {
	page := filter.Page
	if page < 1 {
		page = 1
	}
	perPage := filter.PerPage
	if perPage <= 0 {
		perPage = 20
	}

	conversation, err := r.getAdminConversationSummary(filter.ConversationID)
	if err != nil {
		return erp.AdminConversationDetail{}, err
	}

	participants, err := r.listAdminConversationParticipants(filter.ConversationID)
	if err != nil {
		return erp.AdminConversationDetail{}, err
	}

	messageWhere, messageArgs := adminConversationMessageWhere(filter)
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM messages m LEFT JOIN users sender ON sender.id = m.sender_id "+messageWhere, messageArgs...).Scan(&total); err != nil {
		return erp.AdminConversationDetail{}, fmt.Errorf("count admin conversation messages: %w", err)
	}

	query := `SELECT m.id,
	                 m.created_at,
	                 m.sender_id,
	                 COALESCE(sender.external_user_id, '') AS sender_account,
	                 COALESCE(sender.display_name, '') AS sender_name,
	                 m.message_type,
	                 CASE WHEN m.is_recalled THEN '' ELSE m.content END AS content,
	                 m.is_recalled
	            FROM messages m
	       LEFT JOIN users sender ON sender.id = m.sender_id
	       ` + messageWhere + `
	        ORDER BY m.created_at DESC, m.id DESC
	           LIMIT ? OFFSET ?`
	queryArgs := append(append([]any{}, messageArgs...), perPage, (page-1)*perPage)

	rows, err := r.db.Query(query, queryArgs...)
	if err != nil {
		return erp.AdminConversationDetail{}, fmt.Errorf("list admin conversation messages: %w", err)
	}
	defer rows.Close()

	messages := make([]erp.AdminConversationMessage, 0)
	messageIDs := make([]int64, 0)
	visibleAttachmentIDs := make([]int64, 0)
	for rows.Next() {
		var item erp.AdminConversationMessage
		if err := rows.Scan(
			&item.ID,
			&item.SentAt,
			&item.SenderID,
			&item.SenderAccount,
			&item.SenderName,
			&item.MessageType,
			&item.Content,
			&item.IsRecalled,
		); err != nil {
			return erp.AdminConversationDetail{}, fmt.Errorf("scan admin conversation message: %w", err)
		}
		if item.IsRecalled {
			item.Status = "recalled"
		} else {
			item.Status = "normal"
			visibleAttachmentIDs = append(visibleAttachmentIDs, item.ID)
		}
		messages = append(messages, item)
		messageIDs = append(messageIDs, item.ID)
	}
	if err := rows.Err(); err != nil {
		return erp.AdminConversationDetail{}, fmt.Errorf("iterate admin conversation messages: %w", err)
	}

	if len(visibleAttachmentIDs) > 0 {
		attachments, err := r.listAdminConversationAttachments(visibleAttachmentIDs)
		if err != nil {
			return erp.AdminConversationDetail{}, err
		}
		for index := range messages {
			if values := attachments[messages[index].ID]; len(values) > 0 {
				messages[index].Attachments = values
			}
		}
	}
	if len(messageIDs) > 0 {
		mentions, err := r.listAdminConversationMentions(messageIDs)
		if err != nil {
			return erp.AdminConversationDetail{}, err
		}
		for index := range messages {
			if values := mentions[messages[index].ID]; len(values) > 0 {
				messages[index].Mentions = values
			}
		}
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + perPage - 1) / perPage
	}
	return erp.AdminConversationDetail{
		Conversation: conversation,
		Participants: participants,
		Messages: erp.AdminConversationMessagePage{
			Items:      messages,
			Total:      total,
			Page:       page,
			PerPage:    perPage,
			TotalPages: totalPages,
		},
	}, nil
}

func (r *IntegrationRepository) getAdminConversationSummary(conversationID int64) (erp.AdminConversationSummary, error) {
	activityExpr := "COALESCE(lm.created_at, c.created_at)"
	query := `SELECT c.id, c.type,
	                 CASE
	                   WHEN c.type = 'direct' THEN COALESCE(NULLIF(ms.member_names, ''), CONCAT('聊天室#', c.id))
	                   ELSE COALESCE(NULLIF(c.name, ''), CONCAT('群組#', c.id))
	                 END AS display_name,
	                 CASE WHEN c.type = 'direct' THEN 2 ELSE COALESCE(ms.member_count, 0) END AS member_count,
	                 lm.id AS last_message_id,
	                 COALESCE(lm.message_type, '') AS last_message_type,
	                 COALESCE(lm.content, '') AS last_message_content,
	                 lm.created_at AS last_message_at,
	                 last_sender.id AS last_sender_id,
	                 COALESCE(last_sender.external_user_id, '') AS last_sender_account,
	                 COALESCE(last_sender.display_name, '') AS last_sender_name,
	                 ` + activityExpr + ` AS last_activity_at,
	                 c.created_at,
	                 c.created_by,
	                 COALESCE(creator.external_user_id, '') AS created_by_account,
	                 COALESCE(creator.display_name, '') AS created_by_name
	            ` + adminConversationsFromSQL() + `
	       LEFT JOIN users creator ON creator.id = c.created_by
	           WHERE c.id = ?`

	var item erp.AdminConversationSummary
	var lastMessageID sql.NullInt64
	var lastMessageAt sql.NullTime
	var lastSenderID sql.NullInt64
	var lastMessageContent string
	err := r.db.QueryRow(query, conversationID).Scan(
		&item.ID,
		&item.Type,
		&item.Name,
		&item.MemberCount,
		&lastMessageID,
		&item.LastMessageType,
		&lastMessageContent,
		&lastMessageAt,
		&lastSenderID,
		&item.LastSenderAccount,
		&item.LastSenderName,
		&item.LastActivityAt,
		&item.CreatedAt,
		&item.CreatedByID,
		&item.CreatedByAccount,
		&item.CreatedByName,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return erp.AdminConversationSummary{}, erp.ErrConversationNotFound
		}
		return erp.AdminConversationSummary{}, fmt.Errorf("get admin conversation summary: %w", err)
	}
	if lastMessageID.Valid {
		item.LatestMessageAvailable = true
		item.LastMessage = adminLastMessagePreview(item.LastMessageType, lastMessageContent)
	}
	if lastMessageAt.Valid {
		item.LastMessageAt = &lastMessageAt.Time
	}
	if lastSenderID.Valid {
		item.LastSenderID = lastSenderID.Int64
	}
	return item, nil
}

func (r *IntegrationRepository) listAdminConversationParticipants(conversationID int64) ([]erp.AdminConversationParticipant, error) {
	rows, err := r.db.Query(`
		SELECT cm.user_id,
		       COALESCE(u.external_user_id, '') AS account,
		       COALESCE(u.display_name, '') AS display_name
		  FROM conversation_members cm
		  LEFT JOIN users u ON u.id = cm.user_id
		 WHERE cm.conversation_id = ?
		 ORDER BY cm.user_id ASC`, conversationID)
	if err != nil {
		return nil, fmt.Errorf("list admin conversation participants: %w", err)
	}
	defer rows.Close()

	participants := make([]erp.AdminConversationParticipant, 0)
	for rows.Next() {
		var item erp.AdminConversationParticipant
		if err := rows.Scan(&item.UserID, &item.Account, &item.DisplayName); err != nil {
			return nil, fmt.Errorf("scan admin conversation participant: %w", err)
		}
		participants = append(participants, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate admin conversation participants: %w", err)
	}
	return participants, nil
}

func adminConversationMessageWhere(filter erp.AdminConversationMessageFilter) (string, []any) {
	var where strings.Builder
	args := []any{filter.ConversationID}
	where.WriteString("WHERE m.conversation_id = ?")

	if filter.Keyword != "" {
		where.WriteString(" AND m.is_recalled = FALSE AND m.content LIKE ?")
		args = append(args, "%"+filter.Keyword+"%")
	}
	if filter.SenderKeyword != "" {
		like := "%" + filter.SenderKeyword + "%"
		where.WriteString(" AND (sender.external_user_id LIKE ? OR sender.display_name LIKE ?)")
		args = append(args, like, like)
	}
	if filter.MessageType != "" {
		where.WriteString(" AND m.message_type = ?")
		args = append(args, filter.MessageType)
	}
	if filter.SentFrom != nil {
		where.WriteString(" AND m.created_at >= ?")
		args = append(args, *filter.SentFrom)
	}
	if filter.SentTo != nil {
		where.WriteString(" AND m.created_at <= ?")
		args = append(args, *filter.SentTo)
	}
	switch filter.MentionType {
	case "user":
		where.WriteString(" AND EXISTS (SELECT 1 FROM message_mentions mm WHERE mm.message_id = m.id AND mm.mention_type = 'user')")
	case "all":
		where.WriteString(" AND EXISTS (SELECT 1 FROM message_mentions mm WHERE mm.message_id = m.id AND mm.mention_type = 'all')")
	case "none":
		where.WriteString(" AND NOT EXISTS (SELECT 1 FROM message_mentions mm WHERE mm.message_id = m.id)")
	}

	return where.String(), args
}

func (r *IntegrationRepository) listAdminConversationAttachments(messageIDs []int64) (map[int64][]erp.AdminConversationAttachment, error) {
	placeholders := queryPlaceholders(len(messageIDs))
	args := make([]any, 0, len(messageIDs))
	for _, id := range messageIDs {
		args = append(args, id)
	}
	rows, err := r.db.Query(`
		SELECT message_id, id, original_name, storage_path, mime_type, size_bytes
		  FROM attachments
		 WHERE message_id IN (`+placeholders+`)
		 ORDER BY message_id DESC, id ASC`, args...)
	if err != nil {
		return nil, fmt.Errorf("list admin conversation attachments: %w", err)
	}
	defer rows.Close()

	result := make(map[int64][]erp.AdminConversationAttachment)
	for rows.Next() {
		var messageID int64
		var item erp.AdminConversationAttachment
		if err := rows.Scan(&messageID, &item.ID, &item.OriginalName, &item.StoragePath, &item.MIMEType, &item.SizeBytes); err != nil {
			return nil, fmt.Errorf("scan admin conversation attachment: %w", err)
		}
		if item.StoragePath != "" {
			if strings.HasPrefix(item.StoragePath, "/uploads/") {
				item.URL = item.StoragePath
			} else {
				item.URL = "/uploads/" + strings.TrimLeft(item.StoragePath, "/")
			}
		}
		result[messageID] = append(result[messageID], item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate admin conversation attachments: %w", err)
	}
	return result, nil
}

func (r *IntegrationRepository) listAdminConversationMentions(messageIDs []int64) (map[int64][]erp.AdminConversationMention, error) {
	placeholders := queryPlaceholders(len(messageIDs))
	args := make([]any, 0, len(messageIDs))
	for _, id := range messageIDs {
		args = append(args, id)
	}
	rows, err := r.db.Query(`
		SELECT mm.message_id,
		       mm.mentioned_user_id,
		       COALESCE(u.external_user_id, '') AS account,
		       COALESCE(u.display_name, '') AS display_name,
		       mm.mention_type
		  FROM message_mentions mm
		  LEFT JOIN users u ON u.id = mm.mentioned_user_id
		 WHERE mm.message_id IN (`+placeholders+`)
		 ORDER BY mm.message_id DESC, mm.id ASC`, args...)
	if err != nil {
		return nil, fmt.Errorf("list admin conversation mentions: %w", err)
	}
	defer rows.Close()

	result := make(map[int64][]erp.AdminConversationMention)
	for rows.Next() {
		var messageID int64
		var item erp.AdminConversationMention
		if err := rows.Scan(&messageID, &item.UserID, &item.Account, &item.DisplayName, &item.MentionType); err != nil {
			return nil, fmt.Errorf("scan admin conversation mention: %w", err)
		}
		result[messageID] = append(result[messageID], item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate admin conversation mentions: %w", err)
	}
	return result, nil
}

func adminConversationsFromSQL() string {
	return `FROM conversations c
	  LEFT JOIN (
		SELECT cm.conversation_id,
		       COUNT(*) AS member_count,
		       GROUP_CONCAT(
		         COALESCE(NULLIF(u.display_name, ''), NULLIF(u.external_user_id, ''), CONCAT('使用者#', cm.user_id))
		         ORDER BY cm.user_id SEPARATOR ' ↔ '
		       ) AS member_names
		  FROM conversation_members cm
		  LEFT JOIN users u ON u.id = cm.user_id
		 GROUP BY cm.conversation_id
	  ) ms ON ms.conversation_id = c.id
	  LEFT JOIN (
		SELECT m.id, m.conversation_id, m.sender_id, m.message_type, m.content, m.created_at
		  FROM messages m
		  LEFT JOIN messages newer
		    ON newer.conversation_id = m.conversation_id
		   AND newer.is_recalled = FALSE
		   AND (
		     newer.created_at > m.created_at
		     OR (newer.created_at = m.created_at AND newer.id > m.id)
		   )
		 WHERE m.is_recalled = FALSE
		   AND newer.id IS NULL
	  ) lm ON lm.conversation_id = c.id
	  LEFT JOIN users last_sender ON last_sender.id = lm.sender_id`
}

func adminLastMessagePreview(messageType, content string) string {
	switch strings.ToLower(strings.TrimSpace(messageType)) {
	case "":
		return ""
	case "image":
		return "[圖片]"
	case "file":
		return "[檔案]"
	case "text":
		return truncateRunes(strings.TrimSpace(content), 80)
	default:
		text := strings.TrimSpace(content)
		if text == "" {
			return "[" + messageType + "]"
		}
		return truncateRunes(text, 80)
	}
}

func truncateRunes(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit]) + "..."
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

	from := ` FROM admin_users a
	          LEFT JOIN admin_roles ar ON ar.code = a.role`

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*)"+from+where.String(), args...).Scan(&total); err != nil {
		return erp.SystemAdminPage{}, fmt.Errorf("count system admins: %w", err)
	}

	query := `SELECT a.id, a.id, a.account, a.display_name, a.password_hash, a.role, COALESCE(ar.name, ''), a.status,
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
			&admin.RoleName,
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
		if admin.RoleName == "" {
			admin.RoleName = systemAdminRoleName(admin.Role)
		}
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

// ListAdminRoles returns one page of backend admin roles.
func (r *IntegrationRepository) ListAdminRoles(filter erp.AdminRoleFilter) (erp.AdminRolePage, error) {
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
	args := make([]any, 0, 1)
	if value := strings.TrimSpace(filter.Role); value != "" {
		where.WriteString(" AND (r.code LIKE ? OR r.name LIKE ?)")
		like := "%" + value + "%"
		args = append(args, like, like)
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM admin_roles r"+where.String(), args...).Scan(&total); err != nil {
		return erp.AdminRolePage{}, fmt.Errorf("count admin roles: %w", err)
	}

	query := `SELECT r.id, r.code, r.name, COUNT(a.id) AS admin_count, r.status, r.created_at, r.updated_at
	            FROM admin_roles r
	       LEFT JOIN admin_users a ON a.role = r.code
	         ` + where.String() + `
	        GROUP BY r.id, r.code, r.name, r.status, r.created_at, r.updated_at
	        ORDER BY r.created_at DESC, r.id DESC
	           LIMIT ? OFFSET ?`
	queryArgs := append(append([]any{}, args...), perPage, (page-1)*perPage)

	rows, err := r.db.Query(query, queryArgs...)
	if err != nil {
		return erp.AdminRolePage{}, fmt.Errorf("list admin roles: %w", err)
	}
	defer rows.Close()

	roles := make([]erp.AdminRoleSummary, 0)
	for rows.Next() {
		role, err := scanAdminRole(rows)
		if err != nil {
			return erp.AdminRolePage{}, fmt.Errorf("scan admin role list: %w", err)
		}
		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return erp.AdminRolePage{}, fmt.Errorf("iterate admin role list: %w", err)
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + perPage - 1) / perPage
	}
	return erp.AdminRolePage{
		Items:      roles,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

// FindAdminRoleByID loads a backend admin role by numeric id.
func (r *IntegrationRepository) FindAdminRoleByID(roleID int64) (erp.AdminRoleSummary, error) {
	return r.findAdminRole("r.id = ?", roleID)
}

// FindAdminRoleByCode loads a backend admin role by code.
func (r *IntegrationRepository) FindAdminRoleByCode(code string) (erp.AdminRoleSummary, error) {
	return r.findAdminRole("r.code = ?", strings.TrimSpace(code))
}

func (r *IntegrationRepository) findAdminRole(predicate string, arg any) (erp.AdminRoleSummary, error) {
	row := r.db.QueryRow(
		`SELECT r.id, r.code, r.name, COUNT(a.id) AS admin_count, r.status, r.created_at, r.updated_at
		   FROM admin_roles r
	  LEFT JOIN admin_users a ON a.role = r.code
		  WHERE `+predicate+`
	   GROUP BY r.id, r.code, r.name, r.status, r.created_at, r.updated_at
		  LIMIT 1`,
		arg,
	)
	role, err := scanAdminRole(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return erp.AdminRoleSummary{}, erp.ErrUserNotFound
		}
		return erp.AdminRoleSummary{}, fmt.Errorf("find admin role: %w", err)
	}
	return role, nil
}

// CreateAdminRole creates a backend admin role.
func (r *IntegrationRepository) CreateAdminRole(params erp.AdminRoleCreateParams) (erp.AdminRoleSummary, error) {
	result, err := r.db.Exec(
		`INSERT INTO admin_roles (code, name, status, created_by)
		 VALUES (?, ?, ?, ?)`,
		params.Code,
		params.Name,
		params.Status,
		nullableInt64(params.CreatedBy),
	)
	if err != nil {
		if isDuplicate(err) {
			return erp.AdminRoleSummary{}, erp.ErrRoleAlreadyExists
		}
		return erp.AdminRoleSummary{}, fmt.Errorf("create admin role: %w", err)
	}

	roleID, err := result.LastInsertId()
	if err != nil {
		return erp.AdminRoleSummary{}, fmt.Errorf("read admin role id: %w", err)
	}

	return r.FindAdminRoleByID(roleID)
}

// UpdateAdminRole applies mutable backend admin role fields.
func (r *IntegrationRepository) UpdateAdminRole(roleID int64, params erp.AdminRoleUpdateParams) (erp.AdminRoleSummary, error) {
	result, err := r.db.Exec(
		`UPDATE admin_roles
		    SET name = ?, status = ?
		  WHERE id = ?`,
		params.Name,
		params.Status,
		roleID,
	)
	if err != nil {
		return erp.AdminRoleSummary{}, fmt.Errorf("update admin role: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return erp.AdminRoleSummary{}, fmt.Errorf("read admin role update result: %w", err)
	}
	if affected == 0 {
		return erp.AdminRoleSummary{}, erp.ErrUserNotFound
	}

	return r.FindAdminRoleByID(roleID)
}

// ListAdminRolePermissions loads all stored permission switches for a backend role.
func (r *IntegrationRepository) ListAdminRolePermissions(roleID int64) (map[string]bool, error) {
	rows, err := r.db.Query(
		`SELECT permission_key, enabled
		   FROM admin_role_permissions
		  WHERE role_id = ?`,
		roleID,
	)
	if err != nil {
		return nil, fmt.Errorf("list admin role permissions: %w", err)
	}
	defer rows.Close()

	permissions := make(map[string]bool)
	for rows.Next() {
		var key string
		var enabled bool
		if err := rows.Scan(&key, &enabled); err != nil {
			return nil, fmt.Errorf("scan admin role permission: %w", err)
		}
		permissions[key] = enabled
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate admin role permissions: %w", err)
	}
	return permissions, nil
}

// ReplaceAdminRolePermissions replaces all stored permission switches for a backend role.
func (r *IntegrationRepository) ReplaceAdminRolePermissions(roleID int64, permissions map[string]bool) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin admin role permission transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM admin_role_permissions WHERE role_id = ?`, roleID); err != nil {
		return fmt.Errorf("delete admin role permissions: %w", err)
	}

	stmt, err := tx.Prepare(
		`INSERT INTO admin_role_permissions (role_id, permission_key, enabled)
		 VALUES (?, ?, ?)`,
	)
	if err != nil {
		return fmt.Errorf("prepare admin role permissions insert: %w", err)
	}
	defer stmt.Close()

	for key, enabled := range permissions {
		if _, err := stmt.Exec(roleID, key, enabled); err != nil {
			return fmt.Errorf("insert admin role permission: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit admin role permissions: %w", err)
	}
	return nil
}

type adminRoleScanner interface {
	Scan(dest ...any) error
}

func scanAdminRole(scanner adminRoleScanner) (erp.AdminRoleSummary, error) {
	var role erp.AdminRoleSummary
	var createdAt sql.NullTime
	var updatedAt sql.NullTime
	if err := scanner.Scan(
		&role.ID,
		&role.Code,
		&role.Name,
		&role.AdminCount,
		&role.Status,
		&createdAt,
		&updatedAt,
	); err != nil {
		return erp.AdminRoleSummary{}, err
	}
	if createdAt.Valid {
		role.CreatedAt = &createdAt.Time
	}
	if updatedAt.Valid {
		role.UpdatedAt = &updatedAt.Time
	}
	return role, nil
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
		`SELECT a.id, a.id, a.account, a.display_name, a.password_hash, a.role, COALESCE(ar.name, ''), a.status,
		        a.last_login_at, COALESCE(a.last_login_ip, '')
		   FROM admin_users a
		   LEFT JOIN admin_roles ar ON ar.code = a.role
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
		&admin.RoleName,
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
	if admin.RoleName == "" {
		admin.RoleName = systemAdminRoleName(admin.Role)
	}
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
		if isDuplicateKey(err, "uk_users_email") {
			return erp.User{}, erp.ErrEmailAlreadyExists
		}
		if isDuplicate(err) {
			return erp.User{}, erp.ErrUserAlreadyExists
		}
		return erp.User{}, fmt.Errorf("update user: %w", err)
	}

	return r.FindUserByID(userID)
}

// UpdateUserChatMute updates whether a user can send chat messages.
func (r *IntegrationRepository) UpdateUserChatMute(userID int64, params erp.AdminUpdateUserChatMuteParams) (erp.User, error) {
	result, err := r.db.Exec(
		`UPDATE users
		    SET is_chat_muted = ?
		  WHERE id = ?`,
		params.IsChatMuted,
		userID,
	)
	if err != nil {
		return erp.User{}, fmt.Errorf("update user chat mute: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return erp.User{}, fmt.Errorf("update user chat mute rows affected: %w", err)
	}
	if affected == 0 {
		return erp.User{}, erp.ErrUserNotFound
	}

	return r.FindUserByID(userID)
}

// UpdateUserTemporaryPassword replaces a user's password with a temporary password hash.
func (r *IntegrationRepository) UpdateUserTemporaryPassword(userID int64, params erp.TemporaryPasswordParams) (erp.User, error) {
	result, err := r.db.Exec(
		`UPDATE users
		    SET password_hash = ?,
		        must_change_password = TRUE,
		        temporary_password_expires_at = ?
		  WHERE id = ?`,
		params.PasswordHash,
		params.ExpiresAt,
		userID,
	)
	if err != nil {
		return erp.User{}, fmt.Errorf("update user temporary password: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return erp.User{}, fmt.Errorf("update user temporary password rows affected: %w", err)
	}
	if affected == 0 {
		return erp.User{}, erp.ErrUserNotFound
	}

	return r.FindUserByID(userID)
}

// UpdateUserPassword replaces a user's password and clears temporary password flags.
func (r *IntegrationRepository) UpdateUserPassword(userID int64, passwordHash string) (erp.User, error) {
	result, err := r.db.Exec(
		`UPDATE users
		    SET password_hash = ?,
		        must_change_password = FALSE,
		        temporary_password_expires_at = NULL
		  WHERE id = ?`,
		passwordHash,
		userID,
	)
	if err != nil {
		return erp.User{}, fmt.Errorf("update user password: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return erp.User{}, fmt.Errorf("update user password rows affected: %w", err)
	}
	if affected == 0 {
		return erp.User{}, erp.ErrUserNotFound
	}

	return r.FindUserByID(userID)
}

// UpdateUserProfile updates mutable profile fields for a user.
func (r *IntegrationRepository) UpdateUserProfile(userID int64, displayName string, email string) (erp.User, error) {
	result, err := r.db.Exec(
		`UPDATE users SET display_name = ?, email = NULLIF(?, '') WHERE id = ?`,
		strings.TrimSpace(displayName),
		strings.TrimSpace(email),
		userID,
	)
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

func isDuplicateKey(err error, keyName string) bool {
	var mysqlErr *mysqlDriver.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062 && strings.Contains(mysqlErr.Message, keyName)
	}

	return false
}
