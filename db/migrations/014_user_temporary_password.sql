ALTER TABLE users
    ADD COLUMN must_change_password TINYINT(1) NOT NULL DEFAULT 0 AFTER is_chat_muted,
    ADD COLUMN temporary_password_expires_at DATETIME NULL AFTER must_change_password;
