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
			CASE
				WHEN c.type = 'direct' THEN COALESCE(
					(
						SELECT u.source_system
						  FROM conversation_members cm2
						  JOIN users u ON u.id = cm2.user_id
						 WHERE cm2.conversation_id = c.id
						   AND cm2.user_id <> ?
						 ORDER BY cm2.id ASC
						 LIMIT 1
					),
					''
				)
				ELSE ''
			END AS direct_source_system,
			CASE
				WHEN c.type = 'direct' THEN COALESCE(
					(
						SELECT u.external_user_id
						  FROM conversation_members cm2
						  JOIN users u ON u.id = cm2.user_id
						 WHERE cm2.conversation_id = c.id
						   AND cm2.user_id <> ?
						 ORDER BY cm2.id ASC
						 LIMIT 1
					),
					''
				)
				ELSE ''
			END AS direct_external_user_id,
			(
				SELECT COUNT(*)
				  FROM conversation_members cmc
				 WHERE cmc.conversation_id = c.id
			) AS member_count,
			COALESCE((
					SELECT m.message_type
					  FROM messages m
					 WHERE m.conversation_id = c.id
					   AND m.is_recalled = FALSE
					 ORDER BY m.created_at DESC, m.id DESC
					 LIMIT 1
				), ''),
				COALESCE((
					SELECT m.content
					  FROM messages m
					 WHERE m.conversation_id = c.id
					   AND m.is_recalled = FALSE
					 ORDER BY m.created_at DESC, m.id DESC
					 LIMIT 1
				), ''),
				(
					SELECT m.created_at
					  FROM messages m
					 WHERE m.conversation_id = c.id
					   AND m.is_recalled = FALSE
					 ORDER BY m.created_at DESC, m.id DESC
					 LIMIT 1
				) AS last_message_at,
			(
				SELECT COUNT(*)
					  FROM messages m_unread
					 WHERE m_unread.conversation_id = c.id
					   AND m_unread.is_recalled = FALSE
					   AND m_unread.sender_id <> ?
					   AND m_unread.id > COALESCE((
						SELECT cr.last_read_message_id
						  FROM conversation_reads cr
						 WHERE cr.conversation_id = c.id
						   AND cr.user_id = ?
						 LIMIT 1
				   ), 0)
			) AS unread_count
		  FROM conversation_members cm
		  JOIN conversations c ON c.id = cm.conversation_id
		 WHERE cm.user_id = ?
		 ORDER BY COALESCE(last_message_at, c.created_at) DESC, c.id DESC`, userID, userID, userID, userID, userID, userID)
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
			&item.DirectSourceSystem,
			&item.DirectExternalID,
			&item.MemberCount,
			&item.LastMessageType,
			&item.LastMessagePreview,
			&lastMessageAt,
			&item.UnreadCount,
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

// SearchUsers returns active users matching display name or external user id.
func (r *ChatRepository) SearchUsers(actorUserID int64, sourceSystem, query string, limit int) ([]chat.UserSearchResult, error) {
	rows, err := r.db.Query(`
		SELECT source_system, external_user_id, display_name
		  FROM users
		 WHERE id <> ?
		   AND source_system = ?
		   AND status = 'active'
		   AND (external_user_id LIKE ? OR display_name LIKE ?)
		 ORDER BY
		   CASE
		     WHEN external_user_id = ? THEN 0
		     WHEN external_user_id LIKE ? THEN 1
		     WHEN display_name LIKE ? THEN 2
		     ELSE 3
		   END,
		   display_name ASC,
		   external_user_id ASC
		 LIMIT ?`,
		actorUserID,
		sourceSystem,
		"%"+query+"%",
		"%"+query+"%",
		query,
		query+"%",
		query+"%",
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("search users: %w", err)
	}
	defer rows.Close()

	var items []chat.UserSearchResult
	for rows.Next() {
		var item chat.UserSearchResult
		if err := rows.Scan(&item.SourceSystem, &item.ExternalUserID, &item.DisplayName); err != nil {
			return nil, fmt.Errorf("scan user search result: %w", err)
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

// CreateGroupConversation creates a group conversation with the actor as owner.
func (r *ChatRepository) CreateGroupConversation(actorUserID int64, name string, members []chat.GroupMemberRequest) (chat.Conversation, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return chat.Conversation{}, fmt.Errorf("begin create group conversation: %w", err)
	}
	defer tx.Rollback()

	memberIDs := make([]int64, 0, len(members)+1)
	seenIDs := map[int64]struct{}{actorUserID: {}}
	for _, member := range members {
		var userID int64
		err := tx.QueryRow(`
			SELECT id
			  FROM users
			 WHERE source_system = ?
			   AND external_user_id = ?
			   AND status = 'active'
			 LIMIT 1`, member.SourceSystem, member.ExternalUserID).Scan(&userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return chat.Conversation{}, chat.ErrTargetUserNotFound
			}
			return chat.Conversation{}, fmt.Errorf("find group member: %w", err)
		}
		if _, ok := seenIDs[userID]; ok {
			continue
		}
		seenIDs[userID] = struct{}{}
		memberIDs = append(memberIDs, userID)
	}
	if len(memberIDs) == 0 {
		return chat.Conversation{}, chat.ErrGroupMemberRequired
	}

	result, err := tx.Exec(`INSERT INTO conversations (type, name, created_by) VALUES ('group', ?, ?)`, name, actorUserID)
	if err != nil {
		return chat.Conversation{}, fmt.Errorf("insert group conversation: %w", err)
	}
	conversationID, err := result.LastInsertId()
	if err != nil {
		return chat.Conversation{}, fmt.Errorf("read group conversation id: %w", err)
	}

	if _, err := tx.Exec(`INSERT INTO conversation_members (conversation_id, user_id, role) VALUES (?, ?, 'owner')`, conversationID, actorUserID); err != nil {
		return chat.Conversation{}, fmt.Errorf("insert group owner: %w", err)
	}
	for _, memberID := range memberIDs {
		if _, err := tx.Exec(`INSERT INTO conversation_members (conversation_id, user_id, role) VALUES (?, ?, 'member')`, conversationID, memberID); err != nil {
			return chat.Conversation{}, fmt.Errorf("insert group member: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return chat.Conversation{}, fmt.Errorf("commit group conversation: %w", err)
	}
	return r.GetConversationForUser(actorUserID, conversationID)
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
			recent.id,
			recent.conversation_id,
			recent.sender_id,
			recent.sender_name,
			recent.message_type,
			recent.content,
			recent.created_at,
			a.id AS attachment_id,
			a.original_name AS attachment_original_name,
			a.storage_path AS attachment_storage_path,
			a.mime_type AS attachment_mime_type,
			a.size_bytes AS attachment_size_bytes
		  FROM (
				SELECT
					m.id,
					m.conversation_id,
					m.sender_id,
					u.display_name AS sender_name,
					m.message_type,
					m.content,
					m.created_at
					  FROM messages m
					  JOIN users u ON u.id = m.sender_id
					 WHERE m.conversation_id = ?
					   AND m.is_recalled = FALSE
					 ORDER BY m.created_at DESC, m.id DESC
					 LIMIT ?
		  ) recent
		  LEFT JOIN attachments a ON a.message_id = recent.id
		 ORDER BY recent.created_at ASC, recent.id ASC, a.id ASC`, conversationID, limit)
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
			attachment := chat.Attachment{
				ID:           attachmentID.Int64,
				OriginalName: attachmentOriginalName.String,
				StoragePath:  attachmentStoragePath.String,
				MIMEType:     attachmentMIMEType.String,
				SizeBytes:    attachmentSizeBytes.Int64,
			}
			item.Attachments = append(item.Attachments, attachment)
			if item.Attachment == nil {
				item.Attachment = &item.Attachments[0]
			}
		}
		if len(items) > 0 && items[len(items)-1].ID == item.ID {
			items[len(items)-1].Attachments = append(items[len(items)-1].Attachments, item.Attachments...)
			if items[len(items)-1].Attachment == nil && len(items[len(items)-1].Attachments) > 0 {
				items[len(items)-1].Attachment = &items[len(items)-1].Attachments[0]
			}
			continue
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// SearchMessages returns text messages in conversations visible to the user.
func (r *ChatRepository) SearchMessages(userID int64, query string, limit int) ([]chat.Message, error) {
	rows, err := r.db.Query(`
		SELECT
			m.id,
			m.conversation_id,
			CASE
				WHEN c.type = 'direct' THEN COALESCE(
					(
						SELECT u2.display_name
						  FROM conversation_members cm2
						  JOIN users u2 ON u2.id = cm2.user_id
						 WHERE cm2.conversation_id = c.id
						   AND cm2.user_id <> ?
						 ORDER BY cm2.id ASC
						 LIMIT 1
					),
					COALESCE(NULLIF(c.name, ''), 'Direct Conversation')
				)
				ELSE COALESCE(NULLIF(c.name, ''), 'Unnamed Group')
			END AS conversation_title,
			m.sender_id,
			u.display_name AS sender_name,
			m.message_type,
			m.content,
			m.created_at
		  FROM messages m
		  JOIN conversations c ON c.id = m.conversation_id
		  JOIN conversation_members cm ON cm.conversation_id = c.id AND cm.user_id = ?
		  JOIN users u ON u.id = m.sender_id
			 WHERE m.message_type = 'text'
			   AND m.is_recalled = FALSE
			   AND m.content LIKE ?
		 ORDER BY m.created_at DESC, m.id DESC
		 LIMIT ?`, userID, userID, "%"+query+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("search messages: %w", err)
	}
	defer rows.Close()

	var items []chat.Message
	for rows.Next() {
		var item chat.Message
		if err := rows.Scan(
			&item.ID,
			&item.ConversationID,
			&item.ConversationTitle,
			&item.SenderID,
			&item.SenderName,
			&item.MessageType,
			&item.Content,
			&item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan message search result: %w", err)
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// MarkConversationRead stores the latest message the user has read in a conversation.
func (r *ChatRepository) MarkConversationRead(userID, conversationID int64) error {
	var latestMessageID sql.NullInt64
	row := r.db.QueryRow(`
		SELECT (
			SELECT MAX(m.id)
				  FROM messages m
				 WHERE m.conversation_id = cm.conversation_id
				   AND m.is_recalled = FALSE
			)
		  FROM conversation_members cm
		 WHERE cm.user_id = ?
		   AND cm.conversation_id = ?
		 LIMIT 1`, userID, conversationID)
	if err := row.Scan(&latestMessageID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return chat.ErrConversationNotFound
		}
		return fmt.Errorf("load latest conversation message: %w", err)
	}

	if _, err := r.db.Exec(`
		INSERT INTO conversation_reads (conversation_id, user_id, last_read_message_id, last_read_at)
		VALUES (?, ?, ?, NOW())
		ON DUPLICATE KEY UPDATE
			last_read_message_id = VALUES(last_read_message_id),
			last_read_at = VALUES(last_read_at)`, conversationID, userID, latestMessageID); err != nil {
		return fmt.Errorf("mark conversation read: %w", err)
	}

	return nil
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

	attachments := input.Attachments
	if len(attachments) == 0 && input.Attachment != nil {
		attachments = []chat.AttachmentInput{*input.Attachment}
	}
	for _, attachment := range attachments {
		if _, err := tx.Exec(
			`INSERT INTO attachments (message_id, original_name, storage_path, mime_type, size_bytes) VALUES (?, ?, ?, ?, ?)`,
			messageID,
			attachment.OriginalName,
			attachment.StoragePath,
			attachment.MIMEType,
			attachment.SizeBytes,
		); err != nil {
			return chat.Message{}, fmt.Errorf("create attachment: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return chat.Message{}, fmt.Errorf("commit create message: %w", err)
	}

	rows, err := r.db.Query(`
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
		 WHERE m.id = ?
		   AND m.is_recalled = FALSE
		 ORDER BY a.id ASC`, messageID)
	if err != nil {
		return chat.Message{}, fmt.Errorf("load created message: %w", err)
	}
	defer rows.Close()

	var message chat.Message
	for rows.Next() {
		var attachmentID sql.NullInt64
		var attachmentOriginalName sql.NullString
		var attachmentStoragePath sql.NullString
		var attachmentMIMEType sql.NullString
		var attachmentSizeBytes sql.NullInt64
		if err := rows.Scan(
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
			return chat.Message{}, fmt.Errorf("scan created message: %w", err)
		}
		if attachmentID.Valid {
			message.Attachments = append(message.Attachments, chat.Attachment{
				ID:           attachmentID.Int64,
				OriginalName: attachmentOriginalName.String,
				StoragePath:  attachmentStoragePath.String,
				MIMEType:     attachmentMIMEType.String,
				SizeBytes:    attachmentSizeBytes.Int64,
			})
		}
	}
	if err := rows.Err(); err != nil {
		return chat.Message{}, fmt.Errorf("load created message rows: %w", err)
	}
	if message.ID == 0 {
		return chat.Message{}, chat.ErrConversationNotFound
	}
	if len(message.Attachments) > 0 {
		message.Attachment = &message.Attachments[0]
	}

	return message, nil
}

// DeleteMessage recalls a message owned by the actor.
func (r *ChatRepository) DeleteMessage(conversationID, messageID, actorUserID int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin recall message: %w", err)
	}
	defer tx.Rollback()

	var senderID int64
	if err := tx.QueryRow(`
		SELECT m.sender_id
		  FROM messages m
		  JOIN conversation_members cm ON cm.conversation_id = m.conversation_id AND cm.user_id = ?
		 WHERE m.conversation_id = ?
		   AND m.id = ?
		   AND m.is_recalled = FALSE
		 LIMIT 1`, actorUserID, conversationID, messageID).Scan(&senderID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return chat.ErrMessageNotFound
		}
		return fmt.Errorf("find message for recall: %w", err)
	}
	if senderID != actorUserID {
		return chat.ErrMessageRecallForbidden
	}

	if _, err := tx.Exec(`
		UPDATE conversation_reads
		   SET last_read_message_id = NULL
		 WHERE last_read_message_id = ?`, messageID); err != nil {
		return fmt.Errorf("clear read marker for recalled message: %w", err)
	}
	result, err := tx.Exec(`
		UPDATE messages
		   SET is_recalled = TRUE
		 WHERE conversation_id = ?
		   AND id = ?
		   AND sender_id = ?
		   AND is_recalled = FALSE`, conversationID, messageID, actorUserID)
	if err != nil {
		return fmt.Errorf("recall message: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("recall message rows affected: %w", err)
	}
	if affected == 0 {
		return chat.ErrMessageNotFound
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit recall message: %w", err)
	}
	return nil
}
