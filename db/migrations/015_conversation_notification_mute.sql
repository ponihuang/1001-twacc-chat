ALTER TABLE conversation_members
    ADD COLUMN is_notification_muted TINYINT(1) NOT NULL DEFAULT 0 AFTER role;
