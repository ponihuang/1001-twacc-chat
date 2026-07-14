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

// IsChatMuted checks whether the user is blocked from sending chat messages.
func (r *ChatRepository) IsChatMuted(userID int64) (bool, error) {
	var muted bool
	row := r.db.QueryRow(`SELECT is_chat_muted FROM users WHERE id = ? LIMIT 1`, userID)
	if err := row.Scan(&muted); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, chat.ErrTargetUserNotFound
		}
		return false, fmt.Errorf("check user chat mute: %w", err)
	}

	return muted, nil
}

// MustChangePassword checks whether the user must change password before using chat.
func (r *ChatRepository) MustChangePassword(userID int64) (bool, error) {
	var mustChange bool
	row := r.db.QueryRow(`SELECT must_change_password FROM users WHERE id = ? LIMIT 1`, userID)
	if err := row.Scan(&mustChange); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, chat.ErrTargetUserNotFound
		}
		return false, fmt.Errorf("check user password state: %w", err)
	}

	return mustChange, nil
}

// ListConversations returns conversations for the given user ordered by recent activity.
func (r *ChatRepository) ListConversations(userID int64) ([]chat.ConversationSummary, error) {
	rows, err := r.db.Query(`
		SELECT
			c.id,
			c.type,
			cm.is_notification_muted,
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
			) AS unread_count,
			EXISTS (
				SELECT 1
				  FROM message_mentions mm
				  JOIN messages m_mention ON m_mention.id = mm.message_id
				 WHERE mm.conversation_id = c.id
				   AND mm.mentioned_user_id = ?
				   AND m_mention.is_recalled = FALSE
				   AND m_mention.sender_id <> ?
				   AND mm.message_id > COALESCE((
						SELECT cr.last_read_message_id
						  FROM conversation_reads cr
						 WHERE cr.conversation_id = c.id
						   AND cr.user_id = ?
						 LIMIT 1
				   ), 0)
				 LIMIT 1
			) AS has_unread_mention
		  FROM conversation_members cm
		  JOIN conversations c ON c.id = cm.conversation_id
		 WHERE cm.user_id = ?
		 ORDER BY COALESCE(last_message_at, c.created_at) DESC, c.id DESC`, userID, userID, userID, userID, userID, userID, userID, userID, userID)
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
			&item.NotificationMuted,
			&item.Title,
			&item.DirectSourceSystem,
			&item.DirectExternalID,
			&item.MemberCount,
			&item.LastMessageType,
			&item.LastMessagePreview,
			&lastMessageAt,
			&item.UnreadCount,
			&item.HasUnreadMention,
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

// AddContact saves an active user to the actor's one-way contact list.
func (r *ChatRepository) AddContact(actorUserID int64, sourceSystem, externalUserID, aliasName string) (chat.ContactData, error) {
	var contact chat.ContactData
	row := r.db.QueryRow(`
		SELECT id, source_system, external_user_id, display_name
		  FROM users
		 WHERE source_system = ?
		   AND external_user_id = ?
		   AND status = 'active'
		 LIMIT 1`, sourceSystem, externalUserID)
	if err := row.Scan(&contact.ContactUserID, &contact.SourceSystem, &contact.ExternalUserID, &contact.DisplayName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return chat.ContactData{}, chat.ErrTargetUserNotFound
		}
		return chat.ContactData{}, fmt.Errorf("find contact user: %w", err)
	}
	if contact.ContactUserID == actorUserID {
		return chat.ContactData{}, chat.ErrContactSelfNotAllow
	}

	if _, err := r.db.Exec(`
		INSERT INTO user_contacts (owner_user_id, contact_user_id, alias_name, status)
		VALUES (?, ?, NULLIF(?, ''), 'active')
		ON DUPLICATE KEY UPDATE
			alias_name = VALUES(alias_name),
			status = 'active',
			updated_at = NOW()`, actorUserID, contact.ContactUserID, aliasName); err != nil {
		return chat.ContactData{}, fmt.Errorf("add contact: %w", err)
	}

	contact.AliasName = aliasName
	contact.Status = "active"
	return contact, nil
}

// ListContacts returns active one-way contacts saved by the actor.
func (r *ChatRepository) ListContacts(actorUserID int64) ([]chat.ContactData, error) {
	rows, err := r.db.Query(`
		SELECT u.id,
		       u.source_system,
		       u.external_user_id,
		       u.display_name,
		       COALESCE(uc.alias_name, ''),
		       uc.status
		  FROM user_contacts uc
		  JOIN users u ON u.id = uc.contact_user_id
		 WHERE uc.owner_user_id = ?
		   AND uc.status = 'active'
		   AND u.status = 'active'
		 ORDER BY COALESCE(NULLIF(uc.alias_name, ''), u.display_name, u.external_user_id),
		          u.external_user_id`, actorUserID)
	if err != nil {
		return nil, fmt.Errorf("list contacts: %w", err)
	}
	defer rows.Close()

	contacts := make([]chat.ContactData, 0)
	for rows.Next() {
		var item chat.ContactData
		if err := rows.Scan(&item.ContactUserID, &item.SourceSystem, &item.ExternalUserID, &item.DisplayName, &item.AliasName, &item.Status); err != nil {
			return nil, fmt.Errorf("scan contact: %w", err)
		}
		contacts = append(contacts, item)
	}
	return contacts, rows.Err()
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

// AddConversationMembers adds active users to an existing conversation.
func (r *ChatRepository) AddConversationMembers(conversationID int64, members []chat.GroupMemberRequest) ([]int64, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin add conversation members: %w", err)
	}
	defer tx.Rollback()

	addedIDs := make([]int64, 0, len(members))
	seenIDs := make(map[int64]struct{})
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
				return nil, chat.ErrTargetUserNotFound
			}
			return nil, fmt.Errorf("find conversation member: %w", err)
		}
		if _, ok := seenIDs[userID]; ok {
			continue
		}
		seenIDs[userID] = struct{}{}
		result, err := tx.Exec(`
			INSERT IGNORE INTO conversation_members (conversation_id, user_id, role)
			VALUES (?, ?, 'member')`, conversationID, userID)
		if err != nil {
			return nil, fmt.Errorf("insert conversation member: %w", err)
		}
		if rowsAffected, _ := result.RowsAffected(); rowsAffected > 0 {
			addedIDs = append(addedIDs, userID)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit add conversation members: %w", err)
	}
	return addedIDs, nil
}

// RemoveConversationMember removes a non-owner member when the actor can manage the group.
func (r *ChatRepository) RemoveConversationMember(conversationID, actorUserID int64, sourceSystem, externalUserID string) (int64, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin remove conversation member: %w", err)
	}
	defer tx.Rollback()

	var actorRole string
	if err := tx.QueryRow(`
		SELECT role
		  FROM conversation_members
		 WHERE conversation_id = ?
		   AND user_id = ?
		 LIMIT 1`, conversationID, actorUserID).Scan(&actorRole); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, chat.ErrConversationNotFound
		}
		return 0, fmt.Errorf("find actor member role: %w", err)
	}
	if actorRole != "owner" && actorRole != "admin" {
		return 0, chat.ErrInsufficientRole
	}

	var targetUserID int64
	var targetRole string
	if err := tx.QueryRow(`
		SELECT u.id, cm.role
		  FROM users u
		  JOIN conversation_members cm ON cm.user_id = u.id
		 WHERE cm.conversation_id = ?
		   AND u.source_system = ?
		   AND u.external_user_id = ?
		   AND u.status = 'active'
		 LIMIT 1`, conversationID, sourceSystem, externalUserID).Scan(&targetUserID, &targetRole); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, chat.ErrTargetUserNotFound
		}
		return 0, fmt.Errorf("find target member: %w", err)
	}
	if targetUserID == actorUserID || targetRole == "owner" {
		return 0, chat.ErrInsufficientRole
	}

	result, err := tx.Exec(`
		DELETE FROM conversation_members
		 WHERE conversation_id = ?
		   AND user_id = ?
		   AND role <> 'owner'`, conversationID, targetUserID)
	if err != nil {
		return 0, fmt.Errorf("remove conversation member: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("remove conversation member rows: %w", err)
	}
	if affected == 0 {
		return 0, chat.ErrTargetUserNotFound
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit remove conversation member: %w", err)
	}
	return targetUserID, nil
}

// UpdateGroupConversation updates group metadata when the actor can manage the group.
func (r *ChatRepository) UpdateGroupConversation(conversationID, actorUserID int64, name, description string) (chat.Conversation, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return chat.Conversation{}, fmt.Errorf("begin update group conversation: %w", err)
	}
	defer tx.Rollback()

	var actorRole string
	if err := tx.QueryRow(`
		SELECT cm.role
		  FROM conversation_members cm
		  JOIN conversations c ON c.id = cm.conversation_id
		 WHERE cm.conversation_id = ?
		   AND cm.user_id = ?
		   AND c.type = 'group'
		 LIMIT 1`, conversationID, actorUserID).Scan(&actorRole); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return chat.Conversation{}, chat.ErrConversationNotFound
		}
		return chat.Conversation{}, fmt.Errorf("find actor group role: %w", err)
	}
	if actorRole != "owner" && actorRole != "admin" {
		return chat.Conversation{}, chat.ErrInsufficientRole
	}

	if _, err := tx.Exec(`
		UPDATE conversations
		   SET name = ?, description = ?
		 WHERE id = ?
		   AND type = 'group'`, name, description, conversationID); err != nil {
		return chat.Conversation{}, fmt.Errorf("update group conversation: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return chat.Conversation{}, fmt.Errorf("commit update group conversation: %w", err)
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

// ListConversationMembers returns active user details for all members in a conversation.
func (r *ChatRepository) ListConversationMembers(conversationID int64) ([]chat.ConversationMember, error) {
	rows, err := r.db.Query(`
		SELECT u.id,
		       u.source_system,
		       u.external_user_id,
		       u.display_name,
		       cm.role
		  FROM conversation_members cm
		  JOIN users u ON u.id = cm.user_id
		 WHERE cm.conversation_id = ?
		   AND u.status = 'active'
		 ORDER BY
		   CASE cm.role
		     WHEN 'owner' THEN 0
		     WHEN 'admin' THEN 1
		     ELSE 2
		   END,
		   cm.id ASC`, conversationID)
	if err != nil {
		return nil, fmt.Errorf("list conversation member details: %w", err)
	}
	defer rows.Close()

	members := make([]chat.ConversationMember, 0)
	for rows.Next() {
		var item chat.ConversationMember
		if err := rows.Scan(&item.UserID, &item.SourceSystem, &item.ExternalUserID, &item.DisplayName, &item.Role); err != nil {
			return nil, fmt.Errorf("scan conversation member details: %w", err)
		}
		members = append(members, item)
	}

	return members, rows.Err()
}

// GetConversationForUser returns conversation metadata if the user belongs to the conversation.
func (r *ChatRepository) GetConversationForUser(userID, conversationID int64) (chat.Conversation, error) {
	var conversation chat.Conversation
	row := r.db.QueryRow(`
		SELECT
			c.id,
			c.type,
			cm.is_notification_muted,
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
			COALESCE(c.description, '') AS description
		  FROM conversation_members cm
		  JOIN conversations c ON c.id = cm.conversation_id
		 WHERE cm.user_id = ?
		   AND c.id = ?
		 LIMIT 1`, userID, userID, conversationID)
	if err := row.Scan(&conversation.ID, &conversation.Type, &conversation.NotificationMuted, &conversation.Title, &conversation.Description); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return chat.Conversation{}, chat.ErrConversationNotFound
		}
		return chat.Conversation{}, fmt.Errorf("get conversation: %w", err)
	}

	return conversation, nil
}

// UpdateConversationNotificationMute stores the current user's email notification mute state for a conversation.
func (r *ChatRepository) UpdateConversationNotificationMute(userID, conversationID int64, muted bool) (chat.Conversation, error) {
	if _, err := r.db.Exec(`
		UPDATE conversation_members
		   SET is_notification_muted = ?
		 WHERE conversation_id = ?
		   AND user_id = ?`, muted, conversationID, userID); err != nil {
		return chat.Conversation{}, fmt.Errorf("update conversation notification mute: %w", err)
	}

	return r.GetConversationForUser(userID, conversationID)
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

// FirstUnreadMessageID returns the oldest unread incoming message for a user in a conversation.
func (r *ChatRepository) FirstUnreadMessageID(userID, conversationID int64) (int64, error) {
	var messageID sql.NullInt64
	row := r.db.QueryRow(`
		SELECT MIN(m.id)
		  FROM conversation_members cm
		  JOIN messages m ON m.conversation_id = cm.conversation_id
		 WHERE cm.user_id = ?
		   AND cm.conversation_id = ?
		   AND m.is_recalled = FALSE
		   AND m.sender_id <> ?
		   AND m.id > COALESCE((
				SELECT cr.last_read_message_id
				  FROM conversation_reads cr
				 WHERE cr.conversation_id = cm.conversation_id
				   AND cr.user_id = cm.user_id
				 LIMIT 1
		   ), 0)`, userID, conversationID, userID)
	if err := row.Scan(&messageID); err != nil {
		return 0, fmt.Errorf("load first unread message: %w", err)
	}
	if !messageID.Valid {
		return 0, nil
	}
	return messageID.Int64, nil
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

// MarkConversationReadUntil stores the latest message the user has actually reached.
func (r *ChatRepository) MarkConversationReadUntil(userID, conversationID, messageID int64) error {
	var latestMessageID sql.NullInt64
	row := r.db.QueryRow(`
		SELECT MAX(m.id)
		  FROM conversation_members cm
		  JOIN messages m ON m.conversation_id = cm.conversation_id
		 WHERE cm.user_id = ?
		   AND cm.conversation_id = ?
		   AND m.id <= ?
		   AND m.is_recalled = FALSE`, userID, conversationID, messageID)
	if err := row.Scan(&latestMessageID); err != nil {
		return fmt.Errorf("load visible read marker: %w", err)
	}
	if !latestMessageID.Valid {
		return chat.ErrMessageNotFound
	}

	if _, err := r.db.Exec(`
		INSERT INTO conversation_reads (conversation_id, user_id, last_read_message_id, last_read_at)
		VALUES (?, ?, ?, NOW())
		ON DUPLICATE KEY UPDATE
			last_read_message_id = GREATEST(COALESCE(last_read_message_id, 0), VALUES(last_read_message_id)),
			last_read_at = CASE
				WHEN COALESCE(last_read_message_id, 0) < VALUES(last_read_message_id) THEN VALUES(last_read_at)
				ELSE last_read_at
			END`, conversationID, userID, latestMessageID.Int64); err != nil {
		return fmt.Errorf("mark conversation read until message: %w", err)
	}

	return nil
}

// CreateMessageMentions stores mention targets for a message.
func (r *ChatRepository) CreateMessageMentions(messageID, conversationID int64, mentions []chat.MessageMentionInput) error {
	if len(mentions) == 0 {
		return nil
	}
	stmt, err := r.db.Prepare(`
		INSERT IGNORE INTO message_mentions (message_id, conversation_id, mentioned_user_id, mention_type)
		VALUES (?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare message mentions: %w", err)
	}
	defer stmt.Close()

	for _, mention := range mentions {
		if mention.UserID <= 0 {
			continue
		}
		mentionType := mention.MentionType
		if mentionType == "" {
			mentionType = "user"
		}
		if _, err := stmt.Exec(messageID, conversationID, mention.UserID, mentionType); err != nil {
			return fmt.Errorf("create message mention: %w", err)
		}
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
