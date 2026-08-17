CREATE TABLE IF NOT EXISTS admin_roles (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_by BIGINT UNSIGNED NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_admin_roles_code (code),
    KEY idx_admin_roles_status (status),
    KEY idx_admin_roles_name (name),
    CONSTRAINT fk_admin_roles_created_by
        FOREIGN KEY (created_by) REFERENCES admin_users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO admin_roles (code, name, status)
VALUES ('system_admin', '系統管理員', 'active')
ON DUPLICATE KEY UPDATE
    name = VALUES(name),
    status = VALUES(status);

INSERT INTO admin_roles (code, name, status)
SELECT DISTINCT role, role, 'active'
  FROM admin_users
 WHERE role IS NOT NULL
   AND TRIM(role) <> ''
   AND role <> 'system_admin'
ON DUPLICATE KEY UPDATE
    updated_at = CURRENT_TIMESTAMP;
