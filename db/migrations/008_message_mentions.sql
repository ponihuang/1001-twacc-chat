CREATE TABLE IF NOT EXISTS message_mentions (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    message_id BIGINT UNSIGNED NOT NULL,
    conversation_id BIGINT UNSIGNED NOT NULL,
    mentioned_user_id BIGINT UNSIGNED NOT NULL,
    mention_type VARCHAR(16) NOT NULL DEFAULT 'user',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_message_mentions_user (message_id, mentioned_user_id),
    KEY idx_message_mentions_unread (mentioned_user_id, conversation_id, message_id),
    CONSTRAINT fk_message_mentions_message
        FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
    CONSTRAINT fk_message_mentions_conversation
        FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
    CONSTRAINT fk_message_mentions_user
        FOREIGN KEY (mentioned_user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
