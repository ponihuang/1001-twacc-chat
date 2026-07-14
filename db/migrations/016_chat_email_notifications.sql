CREATE TABLE IF NOT EXISTS chat_email_notifications (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NOT NULL,
    conversation_id BIGINT UNSIGNED NOT NULL,
    first_message_id BIGINT UNSIGNED NOT NULL,
    latest_message_id BIGINT UNSIGNED NOT NULL,
    due_at DATETIME NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    attempt_count INT UNSIGNED NOT NULL DEFAULT 0,
    last_error TEXT NULL,
    sent_at DATETIME NULL,
    cancelled_at DATETIME NULL,
    processing_started_at DATETIME NULL,
    lock_token VARCHAR(64) NULL,
    active_key VARCHAR(128) NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_chat_email_notifications_active (active_key),
    KEY idx_chat_email_notifications_due (status, due_at, id),
    KEY idx_chat_email_notifications_user_conversation (user_id, conversation_id, status),
    KEY idx_chat_email_notifications_lock (lock_token),
    CONSTRAINT fk_chat_email_notifications_user
        FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_chat_email_notifications_conversation
        FOREIGN KEY (conversation_id) REFERENCES conversations(id),
    CONSTRAINT fk_chat_email_notifications_first_message
        FOREIGN KEY (first_message_id) REFERENCES messages(id),
    CONSTRAINT fk_chat_email_notifications_latest_message
        FOREIGN KEY (latest_message_id) REFERENCES messages(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
