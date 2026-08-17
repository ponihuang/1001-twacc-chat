CREATE TABLE IF NOT EXISTS app_settings (
    setting_key VARCHAR(100) NOT NULL,
    setting_name VARCHAR(100) NOT NULL,
    value_type VARCHAR(20) NOT NULL DEFAULT 'number',
    setting_value VARCHAR(255) NOT NULL,
    updated_by_admin_id BIGINT UNSIGNED NULL,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (setting_key),
    CONSTRAINT fk_app_settings_updated_by_admin
        FOREIGN KEY (updated_by_admin_id) REFERENCES admin_users(id)
        ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO app_settings (setting_key, setting_name, value_type, setting_value)
VALUES
    ('chat_auto_delete_max_days', '前台自動刪除最大天數', 'number', '30'),
    ('admin_chat_history_retention_days', '聊天資料庫保留天數', 'number', '90')
ON DUPLICATE KEY UPDATE
    setting_name = VALUES(setting_name),
    value_type = VALUES(value_type);

ALTER TABLE conversations
    ADD COLUMN auto_delete_days INT UNSIGNED NOT NULL DEFAULT 0 AFTER description,
    ADD COLUMN auto_delete_updated_by BIGINT UNSIGNED NULL AFTER auto_delete_days,
    ADD COLUMN auto_delete_updated_at DATETIME NULL AFTER auto_delete_updated_by,
    ADD KEY idx_conversations_auto_delete (auto_delete_days),
    ADD CONSTRAINT fk_conversations_auto_delete_updated_by
        FOREIGN KEY (auto_delete_updated_by) REFERENCES users(id)
        ON DELETE SET NULL;
