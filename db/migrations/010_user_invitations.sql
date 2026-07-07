CREATE TABLE IF NOT EXISTS user_invitations (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    email VARCHAR(255) NOT NULL,
    token_hash CHAR(64) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    invited_by_admin_id BIGINT UNSIGNED NULL,
    accepted_user_id BIGINT UNSIGNED NULL,
    expires_at DATETIME NOT NULL,
    sent_at DATETIME NULL,
    accepted_at DATETIME NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_user_invitations_email (email),
    UNIQUE KEY uk_user_invitations_token_hash (token_hash),
    KEY idx_user_invitations_status (status),
    KEY idx_user_invitations_expires_at (expires_at),
    CONSTRAINT fk_user_invitations_invited_by_admin
        FOREIGN KEY (invited_by_admin_id) REFERENCES admin_users(id)
        ON DELETE SET NULL,
    CONSTRAINT fk_user_invitations_accepted_user
        FOREIGN KEY (accepted_user_id) REFERENCES users(id)
        ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
