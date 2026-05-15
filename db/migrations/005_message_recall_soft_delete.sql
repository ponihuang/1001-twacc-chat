ALTER TABLE messages
    ADD COLUMN is_recalled BOOLEAN NOT NULL DEFAULT FALSE AFTER content,
    ADD KEY idx_messages_conversation_recalled_created_at (conversation_id, is_recalled, created_at);
