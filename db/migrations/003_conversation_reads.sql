CREATE TABLE IF NOT EXISTS conversation_reads (
    conversation_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    last_read_message_id BIGINT UNSIGNED NULL,
    last_read_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (conversation_id, user_id),
    KEY idx_conversation_reads_user (user_id),
    CONSTRAINT fk_conversation_reads_conversation
        FOREIGN KEY (conversation_id) REFERENCES conversations(id),
    CONSTRAINT fk_conversation_reads_user
        FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_conversation_reads_message
        FOREIGN KEY (last_read_message_id) REFERENCES messages(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
