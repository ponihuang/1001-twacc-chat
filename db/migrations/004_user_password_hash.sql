ALTER TABLE users
    ADD COLUMN password_hash VARCHAR(120) NOT NULL DEFAULT '' AFTER display_name;
