CREATE TABLE IF NOT EXISTS user_contacts (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    owner_user_id BIGINT UNSIGNED NOT NULL,
    contact_user_id BIGINT UNSIGNED NOT NULL,
    alias_name VARCHAR(100) NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_user_contacts_owner_contact (owner_user_id, contact_user_id),
    KEY idx_user_contacts_contact (contact_user_id),
    CONSTRAINT fk_user_contacts_owner
        FOREIGN KEY (owner_user_id) REFERENCES users(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_user_contacts_contact
        FOREIGN KEY (contact_user_id) REFERENCES users(id)
        ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
