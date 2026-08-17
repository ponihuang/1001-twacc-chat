CREATE TABLE IF NOT EXISTS admin_role_permissions (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    role_id BIGINT UNSIGNED NOT NULL,
    permission_key VARCHAR(120) NOT NULL,
    enabled TINYINT(1) NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_admin_role_permissions_role_key (role_id, permission_key),
    KEY idx_admin_role_permissions_role (role_id),
    KEY idx_admin_role_permissions_key (permission_key),
    CONSTRAINT fk_admin_role_permissions_role
        FOREIGN KEY (role_id) REFERENCES admin_roles(id)
        ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
