package mysql

import (
	"database/sql"
	"errors"
	"fmt"

	"1001-twacc-chat/internal/chat"
)

// ChatRepository implements chat persistence with MySQL.
type ChatRepository struct {
	db *sql.DB
}

// NewChatRepository creates a MySQL chat repository.
func NewChatRepository(db *sql.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

// IsSystemAdmin checks whether the user belongs to system_admin_users.
func (r *ChatRepository) IsSystemAdmin(userID int64) (bool, error) {
	var exists int
	row := r.db.QueryRow(`SELECT 1 FROM system_admin_users WHERE user_id = ? LIMIT 1`, userID)
	if err := row.Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("check system admin: %w", err)
	}

	return true, nil
}

// ListConversations returns conversations for the given user ordered by recent activity.
func (r *ChatRepository) ListConversations(userID int64) ([]chat.ConversationSummary, error) {
	rows, err := r.db.Query(`
		SELECT
			c.id,
			c.type,
			CASE
				WHEN c.type = 'direct' THEN COALESCE(
					(
						SELECT u.display_name
						  FROM conversation_members cm2
						  JOIN users u ON u.id = cm2.user_id
						 WHERE cm2.conversation_id = c.id
						   AND cm2.user_id <> ?
						 ORDER BY cm2.id ASC
						 LIMIT 1
					),
					COALESCE(NULLIF(c.name, ''), 'Direct Conversation')
				)
				ELSE COALESCE(NULLIF(c.name, ''), 'Unnamed Group')
			END AS title,
			(
				SELECT COUNT(*)
				  FROM conversation_members cmc
				 WHERE cmc.conversation_id = c.id
			) AS member_count,
			COALESCE((
				SELECT m.message_type
				  FROM messages m
				 WHERE m.conversation_id = c.id
				 ORDER BY m.created_at DESC, m.id DESC
				 LIMIT 1
			), ''),
			COALESCE((
				SELECT m.content
				  FROM messages m
				 WHERE m.conversation_id = c.id
				 ORDER BY m.created_at DESC, m.id DESC
				 LIMIT 1
			), ''),
			(
				SELECT m.created_at
				  FROM messages m
				 WHERE m.conversation_id = c.id
				 ORDER BY m.created_at DESC, m.id DESC
				 LIMIT 1
			) AS last_message_at
		  FROM conversation_members cm
		  JOIN conversations c ON c.id = cm.conversation_id
		 WHERE cm.user_id = ?
		 ORDER BY COALESCE(last_message_at, c.created_at) DESC, c.id DESC`, userID, userID)
	if err != nil {
		return nil, fmt.Errorf("list conversations: %w", err)
	}
	defer rows.Close()

	var items []chat.ConversationSummary
	for rows.Next() {
		var item chat.ConversationSummary
		var lastMessageAt sql.NullTime
		if err := rows.Scan(
			&item.ConversationID,
			&item.Type,
			&item.Title,
			&item.MemberCount,
			&item.LastMessageType,
			&item.LastMessagePreview,
			&lastMessageAt,
		); err != nil {
			return nil, fmt.Errorf("scan conversation summary: %w", err)
		}
		if lastMessageAt.Valid {
			item.LastMessageAt = lastMessageAt.Time
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// CreateOrGetDirectConversation creates a direct conversation with the target user, or returns the existing one.
func (r *ChatRepository) CreateOrGetDirectConversation(actorUserID int64, sourceSystem, externalUserID string) (chat.Conversation, bool, error) {
	var targetUserID int64
	row := r.db.QueryRow(`SELECT id FROM users WHERE source_system = ? AND external_user_id = ? LIMIT 1`, sourceSystem, externalUserID)
	if err := row.Scan(&targetUserID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return chat.Conversation{}, false, chat.ErrTargetUserNotFound
		}
		return chat.Conversation{}, false, fmt.Errorf("find target user: %w", err)
	}
	if targetUserID == actorUserID {
		return chat.Conversation{}, false, chat.ErrDirectChatSelfNotAllow
	}

	tx, err := r.db.Begin()
	if err != nil {
		return chat.Conversation{}, false, fmt.Errorf("begin create direct conversation: %w", err)
	}
	defer tx.Rollback()

	var conversationID int64
	existing := tx.QueryRow(`
		SELECT c.id
		  FROM conversations c
		  JOIN conversation_members cm1 ON cm1.conversation_id = c.id AND cm1.user_id = ?
		  JOIN conversation_members cm2 ON cm2.conversation_id = c.id AND cm2.user_id = ?
		 WHERE c.type = 'direct'
		 LIMIT 1`, actorUserID, targetUserID)
	if err := existing.Scan(&conversationID); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return chat.Conversation{}, false, fmt.Errorf("find direct conversation: %w", err)
	}

	created := false
	if conversationID == 0 {
		result, err := tx.Exec(`INSERT INTO conversations (type, name, created_by) VALUES ('direct', NULL, ?)`, actorUserID)
		if err != nil {
			return chat.Conversation{}, false, fmt.Errorf("insert conversation: %w", err)
		}

		conversationID, err = result.LastInsertId()
		if err != nil {
			return chat.Conversation{}, false, fmt.Errorf("read conversation id: %w", err)
		}

		if _, err := tx.Exec(`INSERT INTO conversation_members (conversation_id, user_id, role) VALUES (?, ?, 'member'), (?, ?, 'member')`,
			conversationID, actorUserID, conversationID, targetUserID,
		); err != nil {
			return chat.Conversation{}, false, fmt.Errorf("insert conversation members: %w", err)
		}
		created = true
	}

	if err := tx.Commit(); err != nil {
		return chat.Conversation{}, false, fmt.Errorf("commit direct conversation: %w", err)
	}

	conversation, err := r.GetConversationForUser(actorUserID, conversationID)
	if err != nil {
		return chat.Conversation{}, false, err
	}

	return conversation, created, nil
}

// ListConversationMemberIDs returns all user IDs in a conversation.
func (r *ChatRepository) ListConversationMemberIDs(conversationID int64) ([]int64, error) {
	rows, err := r.db.Query(`SELECT user_id FROM conversation_members WHERE conversation_id = ? ORDER BY id ASC`, conversationID)
	if err != nil {
		return nil, fmt.Errorf("list conversation members: %w", err)
	}
	defer rows.Close()

	var userIDs []int64
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("scan conversation member: %w", err)
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, rows.Err()
}

// GetConversationForUser returns conversation metadata if the user belongs to the conversation.
func (r *ChatRepository) GetConversationForUser(userID, conversationID int64) (chat.Conversation, error) {
	var conversation chat.Conversation
	row := r.db.QueryRow(`
		SELECT
			c.id,
			c.type,
			CASE
				WHEN c.type = 'direct' THEN COALESCE(
					(
						SELECT u.display_name
						  FROM conversation_members cm2
						  JOIN users u ON u.id = cm2.user_id
						 WHERE cm2.conversation_id = c.id
						   AND cm2.user_id <> ?
						 ORDER BY cm2.id ASC
						 LIMIT 1
					),
					COALESCE(NULLIF(c.name, ''), 'Direct Conversation')
				)
				ELSE COALESCE(NULLIF(c.name, ''), 'Unnamed Group')
			END AS title
		  FROM conversation_members cm
		  JOIN conversations c ON c.id = cm.conversation_id
		 WHERE cm.user_id = ?
		   AND c.id = ?
		 LIMIT 1`, userID, userID, conversationID)
	if err := row.Scan(&conversation.ID, &conversation.Type, &conversation.Title); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return chat.Conversation{}, chat.ErrConversationNotFound
		}
		return chat.Conversation{}, fmt.Errorf("get conversation: %w", err)
	}

	return conversation, nil
}

// ListMessages returns recent messages for a conversation ordered from old to new.
func (r *ChatRepository) ListMessages(conversationID int64, limit int) ([]chat.Message, error) {
	rows, err := r.db.Query(`
		SELECT
			id,
			conversation_id,
			sender_id,
			sender_name,
			message_type,
			content,
			created_at,
			attachment_id,
			attachment_original_name,
			attachment_storage_path,
			attachment_mime_type,
			attachment_size_bytes
		  FROM (
				SELECT
					m.id,
					m.conversation_id,
					m.sender_id,
					u.display_name AS sender_name,
					m.message_type,
					m.content,
					m.created_at,
					a.id AS attachment_id,
					a.original_name AS attachment_original_name,
					a.storage_path AS attachment_storage_path,
					a.mime_type AS attachment_mime_type,
					a.size_bytes AS attachment_size_bytes
				  FROM messages m
				  JOIN users u ON u.id = m.sender_id
				  LEFT JOIN attachments a ON a.message_id = m.id
				 WHERE m.conversation_id = ?
				 ORDER BY m.created_at DESC, m.id DESC
				 LIMIT ?
		  ) recent
		 ORDER BY created_at ASC, id ASC`, conversationID, limit)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()

	var items []chat.Message
	for rows.Next() {
		var item chat.Message
		var attachmentID sql.NullInt64
		var attachmentOriginalName sql.NullString
		var attachmentStoragePath sql.NullString
		var attachmentMIMEType sql.NullString
		var attachmentSizeBytes sql.NullInt64
		if err := rows.Scan(
			&item.ID,
			&item.ConversationID,
			&item.SenderID,
			&item.SenderName,
			&item.MessageType,
			&item.Content,
			&item.CreatedAt,
			&attachmentID,
			&attachmentOriginalName,
			&attachmentStoragePath,
			&attachmentMIMEType,
			&attachmentSizeBytes,
		); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		if attachmentID.Valid {
			item.Attachment = &chat.Attachment{
				ID:           attachmentID.Int64,
				OriginalName: attachmentOriginalName.String,
				StoragePath:  attachmentStoragePath.String,
				MIMEType:     attachmentMIMEType.String,
				SizeBytes:    attachmentSizeBytes.Int64,
			}
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// CreateMessage inserts a new message for a conversation member and returns the stored row.
func (r *ChatRepository) CreateMessage(input chat.CreateMessageInput) (chat.Message, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return chat.Message{}, fmt.Errorf("begin create message: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		INSERT INTO messages (conversation_id, sender_id, message_type, content)
		SELECT ?, ?, ?, ?
		  FROM conversation_members cm
		 WHERE cm.conversation_id = ?
		   AND cm.user_id = ?
		 LIMIT 1`, input.ConversationID, input.SenderID, input.MessageType, input.Content, input.ConversationID, input.SenderID)
	if err != nil {
		return chat.Message{}, fmt.Errorf("create message: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return chat.Message{}, fmt.Errorf("create message rows affected: %w", err)
	}
	if affected == 0 {
		return chat.Message{}, chat.ErrConversationNotFound
	}

	messageID, err := result.LastInsertId()
	if err != nil {
		return chat.Message{}, fmt.Errorf("create message last insert id: %w", err)
	}

	if input.Attachment != nil {
		if _, err := tx.Exec(
			`INSERT INTO attachments (message_id, original_name, storage_path, mime_type, size_bytes) VALUES (?, ?, ?, ?, ?)`,
			messageID,
			input.Attachment.OriginalName,
			input.Attachment.StoragePath,
			input.Attachment.MIMEType,
			input.Attachment.SizeBytes,
		); err != nil {
			return chat.Message{}, fmt.Errorf("create attachment: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return chat.Message{}, fmt.Errorf("commit create message: %w", err)
	}

	var message chat.Message
	row := r.db.QueryRow(`
		SELECT
			m.id,
			m.conversation_id,
			m.sender_id,
			u.display_name AS sender_name,
			m.message_type,
			m.content,
			m.created_at,
			a.id,
			a.original_name,
			a.storage_path,
			a.mime_type,
			a.size_bytes
		  FROM messages m
		  JOIN users u ON u.id = m.sender_id
		  LEFT JOIN attachments a ON a.message_id = m.id
		 WHERE m.id = ?`, messageID)
	var attachmentID sql.NullInt64
	var attachmentOriginalName sql.NullString
	var attachmentStoragePath sql.NullString
	var attachmentMIMEType sql.NullString
	var attachmentSizeBytes sql.NullInt64
	if err := row.Scan(
		&message.ID,
		&message.ConversationID,
		&message.SenderID,
		&message.SenderName,
		&message.MessageType,
		&message.Content,
		&message.CreatedAt,
		&attachmentID,
		&attachmentOriginalName,
		&attachmentStoragePath,
		&attachmentMIMEType,
		&attachmentSizeBytes,
	); err != nil {
		return chat.Message{}, fmt.Errorf("load created message: %w", err)
	}
	if attachmentID.Valid {
		message.Attachment = &chat.Attachment{
			ID:           attachmentID.Int64,
			OriginalName: attachmentOriginalName.String,
			StoragePath:  attachmentStoragePath.String,
			MIMEType:     attachmentMIMEType.String,
			SizeBytes:    attachmentSizeBytes.Int64,
		}
	}

	return message, nil
}
