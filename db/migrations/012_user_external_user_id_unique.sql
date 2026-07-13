ALTER TABLE users
    ADD UNIQUE KEY uk_users_external_user_id (external_user_id);
