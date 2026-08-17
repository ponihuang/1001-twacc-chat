package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"1001-twacc-chat/internal/chat"
)

// ChatRepository implements chat persistence with MySQL.
type ChatRepository struct {
	db *sql.DB
}

const enqueueEmailNotificationSQL = `
		INSERT INTO chat_email_notifications
			(user_id, conversation_id, first_message_id, latest_message_id, due_at, status, active_key, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, 'pending', ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE
			latest_message_id = GREATEST(latest_message_id, VALUES(latest_message_id)),
			updated_at = NOW()`

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
			c.auto_delete_days,
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
			&item.AutoDeleteDays,
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
		       cm.role,
		       cm.is_notification_muted
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
		if err := rows.Scan(&item.UserID, &item.SourceSystem, &item.ExternalUserID, &item.DisplayName, &item.Role, &item.NotificationMuted); err != nil {
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
			c.auto_delete_days,
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
	if err := row.Scan(&conversation.ID, &conversation.Type, &conversation.NotificationMuted, &conversation.AutoDeleteDays, &conversation.Title, &conversation.Description); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return chat.Conversation{}, chat.ErrConversationNotFound
		}
		return chat.Conversation{}, fmt.Errorf("get conversation: %w", err)
	}

	return conversation, nil
}

func (r *ChatRepository) conversationTitleForUser(userID, conversationID int64) (string, error) {
	var title string
	row := r.db.QueryRow(`
		SELECT
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
	if err := row.Scan(&title); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", chat.ErrConversationNotFound
		}
		return "", fmt.Errorf("load conversation title: %w", err)
	}
	return title, nil
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

// MaxConversationAutoDeleteDays returns the system-configured frontend maximum auto-delete days.
func (r *ChatRepository) MaxConversationAutoDeleteDays() (int, error) {
	settings, err := r.chatRetentionSettings()
	if err != nil {
		return 0, err
	}
	maxDays := settings.frontendMaxDays
	if settings.databaseRetentionDays > 0 && settings.databaseRetentionDays < maxDays {
		maxDays = settings.databaseRetentionDays
	}
	if maxDays <= 0 {
		return 30, nil
	}
	if maxDays > 30 {
		return 30, nil
	}
	return maxDays, nil
}

type chatRetentionSettings struct {
	frontendMaxDays       int
	databaseRetentionDays int
}

func (r *ChatRepository) chatRetentionSettings() (chatRetentionSettings, error) {
	rows, err := r.db.Query(`
		SELECT setting_key, setting_value
		  FROM app_settings
		 WHERE setting_key IN ('chat_auto_delete_max_days', 'admin_chat_history_retention_days')`)
	if err != nil {
		return chatRetentionSettings{}, fmt.Errorf("get chat retention settings: %w", err)
	}
	defer rows.Close()

	settings := chatRetentionSettings{
		frontendMaxDays:       30,
		databaseRetentionDays: 90,
	}
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return chatRetentionSettings{}, fmt.Errorf("scan chat retention settings: %w", err)
		}
		days, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil || days <= 0 {
			continue
		}
		switch key {
		case "chat_auto_delete_max_days":
			settings.frontendMaxDays = days
		case "admin_chat_history_retention_days":
			settings.databaseRetentionDays = days
		}
	}
	if err := rows.Err(); err != nil {
		return chatRetentionSettings{}, fmt.Errorf("iterate chat retention settings: %w", err)
	}
	return settings, nil
}

func (r *ChatRepository) databaseChatRetentionDays() (int, error) {
	var value string
	row := r.db.QueryRow(`SELECT setting_value FROM app_settings WHERE setting_key = 'admin_chat_history_retention_days' LIMIT 1`)
	if err := row.Scan(&value); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 90, nil
		}
		return 0, fmt.Errorf("get database chat retention days: %w", err)
	}
	days, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || days <= 0 {
		return 90, nil
	}
	return days, nil
}

// UpdateConversationAutoDelete stores the conversation auto-delete retention days.
func (r *ChatRepository) UpdateConversationAutoDelete(userID, conversationID int64, days int) (chat.Conversation, error) {
	result, err := r.db.Exec(`
		UPDATE conversations c
		   JOIN conversation_members cm ON cm.conversation_id = c.id AND cm.user_id = ?
		   SET c.auto_delete_days = ?, c.auto_delete_updated_by = ?, c.auto_delete_updated_at = NOW()
		 WHERE c.id = ?`, userID, days, userID, conversationID)
	if err != nil {
		return chat.Conversation{}, fmt.Errorf("update conversation auto-delete: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return chat.Conversation{}, fmt.Errorf("update conversation auto-delete rows: %w", err)
	}
	if affected == 0 {
		return chat.Conversation{}, chat.ErrConversationNotFound
	}
	return r.GetConversationForUser(userID, conversationID)
}

// PurgeExpiredConversationMessages deletes messages older than the conversation auto-delete setting.
func (r *ChatRepository) PurgeExpiredConversationMessages(conversationID int64, now time.Time) (int64, error) {
	var days int
	if err := r.db.QueryRow(`SELECT auto_delete_days FROM conversations WHERE id = ? LIMIT 1`, conversationID).Scan(&days); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, chat.ErrConversationNotFound
		}
		return 0, fmt.Errorf("load conversation auto-delete days: %w", err)
	}
	if days <= 0 {
		defaultDays, err := r.MaxConversationAutoDeleteDays()
		if err != nil {
			return 0, err
		}
		days = defaultDays
	}
	if days <= 0 {
		return 0, nil
	}
	return r.purgeMessagesWhere(`conversation_id = ? AND created_at < ?`, conversationID, now.AddDate(0, 0, -days))
}

// PurgeExpiredMessages deletes expired messages for every auto-delete-enabled conversation.
func (r *ChatRepository) PurgeExpiredMessages(now time.Time) (int64, error) {
	total, err := r.PurgeExpiredDatabaseMessages(now)
	if err != nil {
		return total, err
	}
	defaultDays, err := r.MaxConversationAutoDeleteDays()
	if err != nil {
		return total, err
	}
	rows, err := r.db.Query(`SELECT id, auto_delete_days FROM conversations`)
	if err != nil {
		return total, fmt.Errorf("list auto-delete conversations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var conversationID int64
		var days int
		if err := rows.Scan(&conversationID, &days); err != nil {
			return total, fmt.Errorf("scan auto-delete conversation: %w", err)
		}
		if days <= 0 {
			days = defaultDays
		}
		if days <= 0 {
			continue
		}
		deleted, err := r.purgeMessagesWhere(`conversation_id = ? AND created_at < ?`, conversationID, now.AddDate(0, 0, -days))
		if err != nil {
			return total, err
		}
		total += deleted
	}
	return total, rows.Err()
}

// PurgeExpiredDatabaseMessages deletes messages older than the system database retention setting.
func (r *ChatRepository) PurgeExpiredDatabaseMessages(now time.Time) (int64, error) {
	days, err := r.databaseChatRetentionDays()
	if err != nil {
		return 0, err
	}
	if days <= 0 {
		return 0, nil
	}
	return r.purgeMessagesWhere(`created_at < ?`, now.AddDate(0, 0, -days))
}

func (r *ChatRepository) purgeMessagesWhere(where string, args ...any) (int64, error) {
	query := `SELECT id FROM messages WHERE ` + where
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return 0, fmt.Errorf("select expired messages: %w", err)
	}
	defer rows.Close()

	messageIDs := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return 0, fmt.Errorf("scan expired message: %w", err)
		}
		messageIDs = append(messageIDs, id)
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("iterate expired messages: %w", err)
	}
	if len(messageIDs) == 0 {
		return 0, nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin purge messages: %w", err)
	}
	defer tx.Rollback()

	placeholders := int64Placeholders(len(messageIDs))
	idArgs := make([]any, 0, len(messageIDs))
	for _, id := range messageIDs {
		idArgs = append(idArgs, id)
	}

	if _, err := tx.Exec(`DELETE FROM chat_email_notifications WHERE first_message_id IN (`+placeholders+`) OR latest_message_id IN (`+placeholders+`)`, append(idArgs, idArgs...)...); err != nil {
		return 0, fmt.Errorf("delete expired message notifications: %w", err)
	}
	if _, err := tx.Exec(`UPDATE conversation_reads SET last_read_message_id = NULL WHERE last_read_message_id IN (`+placeholders+`)`, idArgs...); err != nil {
		return 0, fmt.Errorf("clear expired message read pointers: %w", err)
	}
	if _, err := tx.Exec(`DELETE FROM attachments WHERE message_id IN (`+placeholders+`)`, idArgs...); err != nil {
		return 0, fmt.Errorf("delete expired message attachments: %w", err)
	}
	if _, err := tx.Exec(`DELETE FROM message_mentions WHERE message_id IN (`+placeholders+`)`, idArgs...); err != nil {
		return 0, fmt.Errorf("delete expired message mentions: %w", err)
	}
	result, err := tx.Exec(`DELETE FROM messages WHERE id IN (`+placeholders+`)`, idArgs...)
	if err != nil {
		return 0, fmt.Errorf("delete expired messages: %w", err)
	}
	deleted, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("read expired message delete rows: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit purge messages: %w", err)
	}
	return deleted, nil
}

func int64Placeholders(count int) string {
	if count <= 0 {
		return ""
	}
	values := make([]string, count)
	for i := range values {
		values[i] = "?"
	}
	return strings.Join(values, ",")
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

// CancelPendingEmailNotificationIfNoUnread cancels pending email notifications when the user has no unread incoming messages.
func (r *ChatRepository) CancelPendingEmailNotificationIfNoUnread(userID, conversationID int64) error {
	if _, err := r.db.Exec(`
		UPDATE chat_email_notifications n
		   SET n.status = 'cancelled',
		       n.cancelled_at = NOW(),
		       n.active_key = NULL,
		       n.updated_at = NOW()
		 WHERE n.user_id = ?
		   AND n.conversation_id = ?
		   AND n.status = 'pending'
		   AND NOT EXISTS (
				SELECT 1
				  FROM messages m
				 WHERE m.conversation_id = n.conversation_id
				   AND m.is_recalled = FALSE
				   AND m.sender_id <> n.user_id
				   AND m.id > COALESCE((
						SELECT cr.last_read_message_id
						  FROM conversation_reads cr
						 WHERE cr.conversation_id = n.conversation_id
						   AND cr.user_id = n.user_id
						 LIMIT 1
				   ), 0)
				 LIMIT 1
		   )`, userID, conversationID); err != nil {
		return fmt.Errorf("cancel pending chat email notification if read: %w", err)
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

// EnqueueEmailNotification creates or merges a pending unread email notification.
func (r *ChatRepository) EnqueueEmailNotification(userID, conversationID, messageID int64, dueAt time.Time) error {
	activeKey := fmt.Sprintf("%d:%d", userID, conversationID)
	if _, err := r.db.Exec(enqueueEmailNotificationSQL,
		userID,
		conversationID,
		messageID,
		messageID,
		dueAt,
		activeKey,
	); err != nil {
		return fmt.Errorf("enqueue chat email notification: %w", err)
	}
	return nil
}

// RecoverStaleEmailNotifications returns abandoned processing jobs to pending for retry.
func (r *ChatRepository) RecoverStaleEmailNotifications(now time.Time, staleBefore time.Time) error {
	if _, err := r.db.Exec(`
		UPDATE chat_email_notifications
		   SET status = 'pending',
		       due_at = ?,
		       processing_started_at = NULL,
		       lock_token = NULL,
		       active_key = NULL,
		       updated_at = NOW()
		 WHERE status = 'processing'
		   AND processing_started_at IS NOT NULL
		   AND processing_started_at < ?`, now, staleBefore); err != nil {
		return fmt.Errorf("recover stale chat email notifications: %w", err)
	}
	return nil
}

// ClaimDueEmailNotification atomically claims one due pending email notification.
func (r *ChatRepository) ClaimDueEmailNotification(now time.Time, lockToken string) (*chat.EmailNotificationSchedule, error) {
	result, err := r.db.Exec(`
		UPDATE chat_email_notifications
		   SET status = 'processing',
		       processing_started_at = ?,
		       lock_token = ?,
		       active_key = NULL,
		       updated_at = NOW()
		 WHERE status = 'pending'
		   AND due_at <= ?
		 ORDER BY due_at ASC, id ASC
		 LIMIT 1`, now, lockToken, now)
	if err != nil {
		return nil, fmt.Errorf("claim chat email notification: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("claim chat email notification rows affected: %w", err)
	}
	if affected == 0 {
		return nil, nil
	}

	var schedule chat.EmailNotificationSchedule
	row := r.db.QueryRow(`
		SELECT id, user_id, conversation_id, first_message_id, latest_message_id, attempt_count
		  FROM chat_email_notifications
		 WHERE lock_token = ?
		   AND status = 'processing'
		 LIMIT 1`, lockToken)
	if err := row.Scan(
		&schedule.ID,
		&schedule.UserID,
		&schedule.ConversationID,
		&schedule.FirstMessageID,
		&schedule.LatestMessageID,
		&schedule.AttemptCount,
	); err != nil {
		return nil, fmt.Errorf("load claimed chat email notification: %w", err)
	}
	return &schedule, nil
}

// LoadEmailNotificationDelivery returns data needed to decide whether a claimed notification should be sent.
func (r *ChatRepository) LoadEmailNotificationDelivery(schedule chat.EmailNotificationSchedule) (chat.EmailNotificationDelivery, error) {
	delivery := chat.EmailNotificationDelivery{
		ScheduleID:     schedule.ID,
		UserID:         schedule.UserID,
		ConversationID: schedule.ConversationID,
	}

	row := r.db.QueryRow(`
		SELECT COALESCE(email, ''), status
		  FROM users
		 WHERE id = ?
		 LIMIT 1`, schedule.UserID)
	if err := row.Scan(&delivery.Email, &delivery.UserStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return delivery, nil
		}
		return chat.EmailNotificationDelivery{}, fmt.Errorf("load chat email notification user: %w", err)
	}

	var memberExists int
	row = r.db.QueryRow(`
		SELECT 1
		  FROM conversation_members
		 WHERE conversation_id = ?
		   AND user_id = ?
		 LIMIT 1`, schedule.ConversationID, schedule.UserID)
	if err := row.Scan(&memberExists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			delivery.IsMember = false
			return delivery, nil
		}
		return chat.EmailNotificationDelivery{}, fmt.Errorf("load chat email notification member: %w", err)
	}
	delivery.IsMember = true

	if title, err := r.conversationTitleForUser(schedule.UserID, schedule.ConversationID); err != nil {
		return chat.EmailNotificationDelivery{}, err
	} else {
		delivery.ConversationTitle = title
	}

	var latestUnreadID sql.NullInt64
	row = r.db.QueryRow(`
		SELECT COUNT(*), MAX(m.id)
		  FROM messages m
		 WHERE m.conversation_id = ?
		   AND m.is_recalled = FALSE
		   AND m.sender_id <> ?
		   AND m.id BETWEEN ? AND ?
		   AND m.id > COALESCE((
				SELECT cr.last_read_message_id
				  FROM conversation_reads cr
				 WHERE cr.conversation_id = ?
				   AND cr.user_id = ?
				 LIMIT 1
		   ), 0)`,
		schedule.ConversationID,
		schedule.UserID,
		schedule.FirstMessageID,
		schedule.LatestMessageID,
		schedule.ConversationID,
		schedule.UserID,
	)
	if err := row.Scan(&delivery.UnreadCount, &latestUnreadID); err != nil {
		return chat.EmailNotificationDelivery{}, fmt.Errorf("load chat email notification unread count: %w", err)
	}
	if delivery.UnreadCount <= 0 || !latestUnreadID.Valid {
		return delivery, nil
	}

	row = r.db.QueryRow(`
		SELECT m.message_type, m.content, u.display_name
		  FROM messages m
		  JOIN users u ON u.id = m.sender_id
		 WHERE m.id = ?
		 LIMIT 1`, latestUnreadID.Int64)
	if err := row.Scan(&delivery.LatestMessageType, &delivery.LatestMessageText, &delivery.LatestSenderName); err != nil {
		return chat.EmailNotificationDelivery{}, fmt.Errorf("load chat email notification latest message: %w", err)
	}
	delivery.LatestMessageText = notificationPreview(delivery.LatestMessageType, delivery.LatestMessageText)
	return delivery, nil
}

// MarkEmailNotificationSent marks a claimed email notification as sent.
func (r *ChatRepository) MarkEmailNotificationSent(scheduleID int64, sentAt time.Time) error {
	if _, err := r.db.Exec(`
		UPDATE chat_email_notifications
		   SET status = 'sent',
		       sent_at = ?,
		       lock_token = NULL,
		       updated_at = NOW()
		 WHERE id = ?`, sentAt, scheduleID); err != nil {
		return fmt.Errorf("mark chat email notification sent: %w", err)
	}
	return nil
}

// MarkEmailNotificationCancelled marks a claimed email notification as cancelled.
func (r *ChatRepository) MarkEmailNotificationCancelled(scheduleID int64, cancelledAt time.Time) error {
	if _, err := r.db.Exec(`
		UPDATE chat_email_notifications
		   SET status = 'cancelled',
		       cancelled_at = ?,
		       lock_token = NULL,
		       active_key = NULL,
		       updated_at = NOW()
		 WHERE id = ?`, cancelledAt, scheduleID); err != nil {
		return fmt.Errorf("mark chat email notification cancelled: %w", err)
	}
	return nil
}

// MarkEmailNotificationFailed records an email notification delivery failure.
func (r *ChatRepository) MarkEmailNotificationFailed(scheduleID int64, retryAt time.Time, maxAttempts int, errMessage string) error {
	message := strings.TrimSpace(errMessage)
	if len(message) > 1000 {
		message = message[:1000]
	}
	if _, err := r.db.Exec(`
		UPDATE chat_email_notifications
		   SET attempt_count = attempt_count + 1,
		       last_error = ?,
		       status = CASE WHEN attempt_count + 1 >= ? THEN 'failed' ELSE 'pending' END,
		       due_at = CASE WHEN attempt_count + 1 >= ? THEN due_at ELSE ? END,
		       lock_token = NULL,
		       active_key = NULL,
		       updated_at = NOW()
		 WHERE id = ?`,
		message,
		maxAttempts,
		maxAttempts,
		retryAt,
		scheduleID,
	); err != nil {
		return fmt.Errorf("mark chat email notification failed: %w", err)
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

func notificationPreview(messageType, content string) string {
	switch strings.ToLower(strings.TrimSpace(messageType)) {
	case "image":
		return "傳送了圖片"
	case "file":
		return "傳送了附件"
	default:
		return strings.TrimSpace(content)
	}
}
