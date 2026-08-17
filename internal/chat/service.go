package chat

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
)

const (
	messageListLimit       = 100
	messageSearchLimit     = 30
	userSearchLimit        = 10
	emailNotificationDelay = 5 * time.Minute
	textMessageType        = "text"
	imageMessageType       = "image"
	fileMessageType        = "file"
)

var conversationAutoDeleteChoices = []int{1, 3, 7, 30}

var (
	ErrConversationNotFound      = errors.New("conversation not found")
	ErrInsufficientRole          = errors.New("insufficient role")
	ErrSystemAdminCannotChat     = errors.New("system admin cannot chat")
	ErrUnsupportedMessageType    = errors.New("unsupported message type")
	ErrMessageContentRequired    = errors.New("message content required")
	ErrTargetUserNotFound        = errors.New("target user not found")
	ErrDirectChatSelfNotAllow    = errors.New("direct conversation with self is not allowed")
	ErrContactSelfNotAllow       = errors.New("contact with self is not allowed")
	ErrGroupNameRequired         = errors.New("group name required")
	ErrGroupMemberRequired       = errors.New("group member required")
	ErrAttachmentRequired        = errors.New("attachment required")
	ErrAttachmentTooLarge        = errors.New("attachment too large")
	ErrUnsupportedAttachmentType = errors.New("unsupported attachment type")
	ErrMessageNotFound           = errors.New("message not found")
	ErrMessageRecallForbidden    = errors.New("message recall forbidden")
	ErrUserChatMuted             = errors.New("user chat muted")
	ErrPasswordChangeRequired    = errors.New("password change required")
)

var mentionPattern = regexp.MustCompile(`(?:^|\s)@([A-Za-z0-9_.-]+|ALL)\b`)

// Service implements chat list and message list rules.
type Service struct {
	repo   Repository
	broker Broker
}

// CreateGroupConversation creates a group conversation owned by the current user.
func (s *Service) CreateGroupConversation(actor SessionPrincipal, req CreateGroupConversationRequest) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("chat service unavailable")
	}
	if err := s.requireChatUser(actor.UserID); err != nil {
		return Response{}, statusCode(err), err
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return Response{}, statusCode(ErrGroupNameRequired), ErrGroupNameRequired
	}
	if len([]rune(name)) > 255 {
		name = string([]rune(name)[:255])
	}
	members := normalizedGroupMembers(req.Members)
	if len(members) == 0 {
		return Response{}, statusCode(ErrGroupMemberRequired), ErrGroupMemberRequired
	}

	conversation, err := s.repo.CreateGroupConversation(actor.UserID, name, members)
	if err != nil {
		return Response{}, statusCode(err), err
	}
	data := DirectConversationData{
		ConversationID: conversation.ID,
		Type:           conversation.Type,
		Title:          conversation.Title,
	}

	if s.broker != nil {
		memberIDs, err := s.repo.ListConversationMemberIDs(conversation.ID)
		if err != nil {
			return Response{}, 500, err
		}
		s.broker.PublishToUsers(memberIDs, RealtimeEvent{
			EventType:      "conversation.ready",
			ConversationID: conversation.ID,
			Conversation:   &data,
		})
	}

	return Response{
		Success: true,
		Code:    "GROUP_CONVERSATION_CREATED",
		Message: "群組對話建立成功",
		Data:    data,
	}, 201, nil
}

// NewService builds a chat service.
func NewService(repo Repository, broker Broker) *Service {
	return &Service{repo: repo, broker: broker}
}

// AddContact saves another active user to the current user's contact list.
func (s *Service) AddContact(actor SessionPrincipal, req AddContactRequest) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("chat service unavailable")
	}
	if err := s.requireChatUser(actor.UserID); err != nil {
		return Response{}, statusCode(err), err
	}

	sourceSystem := strings.TrimSpace(req.SourceSystem)
	externalUserID := strings.TrimSpace(req.ExternalUserID)
	aliasName := strings.TrimSpace(req.AliasName)
	if sourceSystem == "" || externalUserID == "" {
		return Response{}, statusCode(ErrTargetUserNotFound), ErrTargetUserNotFound
	}
	if len([]rune(aliasName)) > 100 {
		aliasName = string([]rune(aliasName)[:100])
	}

	contact, err := s.repo.AddContact(actor.UserID, sourceSystem, externalUserID, aliasName)
	if err != nil {
		return Response{}, statusCode(err), err
	}

	return Response{
		Success: true,
		Code:    "CONTACT_ADDED",
		Message: "联系人已加入",
		Data:    contact,
	}, 201, nil
}

// ListContacts returns the current user's active contact list.
func (s *Service) ListContacts(actor SessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("chat service unavailable")
	}
	if err := s.requireChatUser(actor.UserID); err != nil {
		return Response{}, statusCode(err), err
	}

	contacts, err := s.repo.ListContacts(actor.UserID)
	if err != nil {
		return Response{}, statusCode(err), err
	}

	return Response{
		Success: true,
		Code:    "CONTACTS_OK",
		Message: "联系人列表已载入",
		Data:    ContactListData{Contacts: contacts},
	}, 200, nil
}

// ListConversations returns conversations visible to the current user.
func (s *Service) ListConversations(actor SessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("chat service unavailable")
	}
	if err := s.requireChatUser(actor.UserID); err != nil {
		return Response{}, statusCode(err), err
	}

	conversations, err := s.repo.ListConversations(actor.UserID)
	if err != nil {
		return Response{}, 500, err
	}

	items := make([]ConversationItem, 0, len(conversations))
	for _, conversation := range conversations {
		item := ConversationItem{
			ConversationID:     conversation.ConversationID,
			Type:               conversation.Type,
			Title:              conversation.Title,
			DirectSourceSystem: conversation.DirectSourceSystem,
			DirectExternalID:   conversation.DirectExternalID,
			MemberCount:        conversation.MemberCount,
			LastMessageType:    conversation.LastMessageType,
			LastMessagePreview: conversation.LastMessagePreview,
			UnreadCount:        conversation.UnreadCount,
			HasUnreadMention:   conversation.HasUnreadMention,
			NotificationMuted:  conversation.NotificationMuted,
			AutoDeleteDays:     conversation.AutoDeleteDays,
		}
		if !conversation.LastMessageAt.IsZero() {
			item.LastMessageAt = conversation.LastMessageAt.Format(time.RFC3339)
		}
		items = append(items, item)
	}

	return Response{Success: true, Code: "CONVERSATIONS_OK", Message: "对话列表读取成功", Data: items}, 200, nil
}

// ListConversationMembers returns member data for a conversation visible to the actor.
func (s *Service) ListConversationMembers(actor SessionPrincipal, conversationID int64) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("chat service unavailable")
	}
	if err := s.requireChatUser(actor.UserID); err != nil {
		return Response{}, statusCode(err), err
	}
	if conversationID <= 0 {
		return Response{}, statusCode(ErrConversationNotFound), ErrConversationNotFound
	}

	conversation, err := s.repo.GetConversationForUser(actor.UserID, conversationID)
	if err != nil {
		return Response{}, statusCode(err), err
	}
	members, err := s.repo.ListConversationMembers(conversationID)
	if err != nil {
		return Response{}, statusCode(err), err
	}

	items := make([]ConversationMemberItem, 0, len(members))
	for _, member := range members {
		items = append(items, ConversationMemberItem{
			UserID:         member.UserID,
			SourceSystem:   member.SourceSystem,
			ExternalUserID: member.ExternalUserID,
			DisplayName:    member.DisplayName,
			Role:           member.Role,
		})
	}

	return Response{
		Success: true,
		Code:    "CONVERSATION_MEMBERS_OK",
		Message: "成员列表读取成功",
		Data: ConversationMembersData{
			ConversationID: conversation.ID,
			Type:           conversation.Type,
			Title:          conversation.Title,
			Description:    conversation.Description,
			MemberCount:    len(items),
			Members:        items,
		},
	}, 200, nil
}

// UpdateConversation updates group metadata when the actor can manage it.
func (s *Service) UpdateConversation(actor SessionPrincipal, conversationID int64, req UpdateConversationRequest) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("chat service unavailable")
	}
	if err := s.requireChatUser(actor.UserID); err != nil {
		return Response{}, statusCode(err), err
	}
	if conversationID <= 0 {
		return Response{}, statusCode(ErrConversationNotFound), ErrConversationNotFound
	}
	conversation, err := s.repo.GetConversationForUser(actor.UserID, conversationID)
	if err != nil {
		return Response{}, statusCode(err), err
	}
	if conversation.Type != "group" {
		return Response{}, statusCode(ErrConversationNotFound), ErrConversationNotFound
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return Response{}, statusCode(ErrGroupNameRequired), ErrGroupNameRequired
	}
	if len([]rune(name)) > 255 {
		name = string([]rune(name)[:255])
	}
	description := strings.TrimSpace(req.Description)
	if len([]rune(description)) > 1000 {
		description = string([]rune(description)[:1000])
	}

	if _, err := s.repo.UpdateGroupConversation(conversationID, actor.UserID, name, description); err != nil {
		return Response{}, statusCode(err), err
	}
	if s.broker != nil {
		memberIDs, err := s.repo.ListConversationMemberIDs(conversationID)
		if err != nil {
			return Response{}, 500, err
		}
		s.broker.PublishToUsers(memberIDs, RealtimeEvent{
			EventType:      "conversation.members.updated",
			ConversationID: conversationID,
		})
	}
	return s.ListConversationMembers(actor, conversationID)
}

// UpdateConversationNotificationMute updates the current user's email notification mute state for a conversation.
func (s *Service) UpdateConversationNotificationMute(actor SessionPrincipal, conversationID int64, req UpdateConversationNotificationMuteRequest) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("chat service unavailable")
	}
	if err := s.requireChatUser(actor.UserID); err != nil {
		return Response{}, statusCode(err), err
	}
	if conversationID <= 0 {
		return Response{}, statusCode(ErrConversationNotFound), ErrConversationNotFound
	}

	conversation, err := s.repo.UpdateConversationNotificationMute(actor.UserID, conversationID, req.NotificationMuted)
	if err != nil {
		return Response{}, statusCode(err), err
	}

	return Response{
		Success: true,
		Code:    "CONVERSATION_NOTIFICATION_MUTE_UPDATED",
		Message: "通知靜音設定已更新",
		Data: ConversationNotificationMuteData{
			ConversationID:    conversation.ID,
			NotificationMuted: conversation.NotificationMuted,
		},
	}, 200, nil
}

// UpdateConversationAutoDelete updates auto-delete days for a conversation. Any member may change it.
func (s *Service) UpdateConversationAutoDelete(actor SessionPrincipal, conversationID int64, req UpdateConversationAutoDeleteRequest) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("chat service unavailable")
	}
	if err := s.requireChatUser(actor.UserID); err != nil {
		return Response{}, statusCode(err), err
	}
	if conversationID <= 0 {
		return Response{}, statusCode(ErrConversationNotFound), ErrConversationNotFound
	}

	maxDays, err := s.repo.MaxConversationAutoDeleteDays()
	if err != nil {
		return Response{}, 500, err
	}
	options := allowedAutoDeleteOptions(maxDays)
	if !validAutoDeleteDays(req.Days, options) {
		return Response{}, 400, fmt.Errorf("auto_delete_days must be 0 or one of available options")
	}

	conversation, err := s.repo.UpdateConversationAutoDelete(actor.UserID, conversationID, req.Days)
	if err != nil {
		return Response{}, statusCode(err), err
	}
	if _, err := s.repo.PurgeExpiredConversationMessages(conversationID, time.Now().UTC()); err != nil {
		return Response{}, 500, err
	}

	if s.broker != nil {
		if memberIDs, err := s.repo.ListConversationMemberIDs(conversationID); err == nil {
			s.broker.PublishToUsers(memberIDs, RealtimeEvent{
				EventType:      "conversation.auto_delete.updated",
				ConversationID: conversationID,
				AutoDelete: &ConversationAutoDeleteData{
					ConversationID:    conversationID,
					AutoDeleteDays:    conversation.AutoDeleteDays,
					AutoDeleteOptions: options,
					MaxAutoDeleteDays: maxDays,
				},
			})
		}
	}

	return Response{
		Success: true,
		Code:    "CONVERSATION_AUTO_DELETE_UPDATED",
		Message: "自動刪除設定已更新",
		Data: ConversationAutoDeleteData{
			ConversationID:    conversation.ID,
			AutoDeleteDays:    conversation.AutoDeleteDays,
			AutoDeleteOptions: options,
			MaxAutoDeleteDays: maxDays,
		},
	}, 200, nil
}

// AddConversationMembers adds members to a group visible to the actor.
func (s *Service) AddConversationMembers(actor SessionPrincipal, conversationID int64, req AddConversationMembersRequest) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("chat service unavailable")
	}
	if err := s.requireChatUser(actor.UserID); err != nil {
		return Response{}, statusCode(err), err
	}
	if conversationID <= 0 {
		return Response{}, statusCode(ErrConversationNotFound), ErrConversationNotFound
	}
	conversation, err := s.repo.GetConversationForUser(actor.UserID, conversationID)
	if err != nil {
		return Response{}, statusCode(err), err
	}
	if conversation.Type != "group" {
		return Response{}, statusCode(ErrConversationNotFound), ErrConversationNotFound
	}
	members := normalizedGroupMembers(req.Members)
	if len(members) == 0 {
		return Response{}, statusCode(ErrGroupMemberRequired), ErrGroupMemberRequired
	}
	addedIDs, err := s.repo.AddConversationMembers(conversationID, members)
	if err != nil {
		return Response{}, statusCode(err), err
	}
	if s.broker != nil && len(addedIDs) > 0 {
		s.broker.PublishToUsers(addedIDs, RealtimeEvent{
			EventType:      "conversation.ready",
			ConversationID: conversationID,
		})
	}
	return s.ListConversationMembers(actor, conversationID)
}

// RemoveConversationMember removes a member from a group when the actor can manage it.
func (s *Service) RemoveConversationMember(actor SessionPrincipal, conversationID int64, req RemoveConversationMemberRequest) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("chat service unavailable")
	}
	if err := s.requireChatUser(actor.UserID); err != nil {
		return Response{}, statusCode(err), err
	}
	if conversationID <= 0 {
		return Response{}, statusCode(ErrConversationNotFound), ErrConversationNotFound
	}
	conversation, err := s.repo.GetConversationForUser(actor.UserID, conversationID)
	if err != nil {
		return Response{}, statusCode(err), err
	}
	if conversation.Type != "group" {
		return Response{}, statusCode(ErrConversationNotFound), ErrConversationNotFound
	}
	sourceSystem := strings.TrimSpace(req.SourceSystem)
	externalUserID := strings.TrimSpace(req.ExternalUserID)
	if sourceSystem == "" || externalUserID == "" {
		return Response{}, statusCode(ErrTargetUserNotFound), ErrTargetUserNotFound
	}

	removedID, err := s.repo.RemoveConversationMember(conversationID, actor.UserID, sourceSystem, externalUserID)
	if err != nil {
		return Response{}, statusCode(err), err
	}
	if s.broker != nil {
		memberIDs, err := s.repo.ListConversationMemberIDs(conversationID)
		if err != nil {
			return Response{}, 500, err
		}
		memberIDs = append(memberIDs, removedID)
		s.broker.PublishToUsers(memberIDs, RealtimeEvent{
			EventType:      "conversation.members.updated",
			ConversationID: conversationID,
		})
	}
	return s.ListConversationMembers(actor, conversationID)
}

// SearchUsers returns active users matching the query for direct conversation creation.
func (s *Service) SearchUsers(actor SessionPrincipal, sourceSystem, query string) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("chat service unavailable")
	}
	if err := s.requireChatUser(actor.UserID); err != nil {
		return Response{}, statusCode(err), err
	}

	source := strings.TrimSpace(sourceSystem)
	trimmed := strings.TrimSpace(strings.TrimPrefix(query, "@"))
	if source == "" || trimmed == "" {
		return Response{
			Success: true,
			Code:    "USER_SEARCH_OK",
			Message: "使用者搜寻完成",
			Data:    UserSearchData{Query: trimmed, Users: []UserSearchItem{}},
		}, 200, nil
	}

	users, err := s.repo.SearchUsers(actor.UserID, source, trimmed, userSearchLimit)
	if err != nil {
		return Response{}, 500, err
	}

	items := make([]UserSearchItem, 0, len(users))
	for _, user := range users {
		items = append(items, UserSearchItem{
			SourceSystem:   user.SourceSystem,
			ExternalUserID: user.ExternalUserID,
			DisplayName:    user.DisplayName,
		})
	}

	return Response{
		Success: true,
		Code:    "USER_SEARCH_OK",
		Message: "使用者搜寻完成",
		Data: UserSearchData{
			Query: trimmed,
			Users: items,
		},
	}, 200, nil
}

// CreateDirectConversation creates or loads a direct conversation with another externally-identified user.
func (s *Service) CreateDirectConversation(actor SessionPrincipal, req CreateDirectConversationRequest) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("chat service unavailable")
	}
	if err := s.requireChatUser(actor.UserID); err != nil {
		return Response{}, statusCode(err), err
	}

	sourceSystem := strings.TrimSpace(req.SourceSystem)
	externalUserID := strings.TrimSpace(req.ExternalUserID)
	if sourceSystem == "" || externalUserID == "" {
		return Response{}, 400, ErrTargetUserNotFound
	}

	conversation, created, err := s.repo.CreateOrGetDirectConversation(actor.UserID, sourceSystem, externalUserID)
	if err != nil {
		return Response{}, statusCode(err), err
	}

	code := "CONVERSATION_EXISTS"
	message := "一对一对话已存在"
	status := 200
	if created {
		code = "CONVERSATION_CREATED"
		message = "一对一对话创建成功"
		status = 201
	}

	data := DirectConversationData{
		ConversationID: conversation.ID,
		Type:           conversation.Type,
		Title:          conversation.Title,
		AutoDeleteDays: conversation.AutoDeleteDays,
	}

	if s.broker != nil {
		memberIDs, err := s.repo.ListConversationMemberIDs(conversation.ID)
		if err != nil {
			return Response{}, 500, err
		}
		s.broker.PublishToUsers(memberIDs, RealtimeEvent{
			EventType:      "conversation.ready",
			ConversationID: conversation.ID,
			Conversation:   &data,
		})
	}

	return Response{
		Success: true,
		Code:    code,
		Message: message,
		Data:    data,
	}, status, nil
}

// ListMessages returns recent messages for a conversation visible to the current user.
func (s *Service) ListMessages(conversationID int64, actor SessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("chat service unavailable")
	}
	if conversationID <= 0 {
		return Response{}, statusCode(ErrConversationNotFound), ErrConversationNotFound
	}
	if err := s.requireChatUser(actor.UserID); err != nil {
		return Response{}, statusCode(err), err
	}

	conversation, err := s.repo.GetConversationForUser(actor.UserID, conversationID)
	if err != nil {
		return Response{}, statusCode(err), err
	}
	if _, err := s.repo.PurgeExpiredConversationMessages(conversationID, time.Now().UTC()); err != nil {
		return Response{}, 500, err
	}

	messages, err := s.repo.ListMessages(conversationID, messageListLimit)
	if err != nil {
		return Response{}, 500, err
	}
	firstUnreadMessageID, err := s.repo.FirstUnreadMessageID(actor.UserID, conversationID)
	if err != nil {
		return Response{}, 500, err
	}
	maxAutoDeleteDays, err := s.repo.MaxConversationAutoDeleteDays()
	if err != nil {
		return Response{}, 500, err
	}
	autoDeleteOptions := allowedAutoDeleteOptions(maxAutoDeleteDays)

	items := make([]MessageItem, 0, len(messages))
	for _, message := range messages {
		items = append(items, MessageItem{
			MessageID:   message.ID,
			SenderID:    message.SenderID,
			SenderName:  message.SenderName,
			MessageType: message.MessageType,
			Content:     message.Content,
			CreatedAt:   message.CreatedAt.Format(time.RFC3339),
			Attachment:  firstAttachmentItem(message),
			Attachments: attachmentItems(message.Attachments),
		})
	}

	return Response{
		Success: true,
		Code:    "MESSAGES_OK",
		Message: "讯息列表读取成功",
		Data: MessageListData{
			ConversationID:       conversation.ID,
			Type:                 conversation.Type,
			Title:                conversation.Title,
			NotificationMuted:    conversation.NotificationMuted,
			AutoDeleteDays:       conversation.AutoDeleteDays,
			AutoDeleteOptions:    autoDeleteOptions,
			MaxAutoDeleteDays:    maxAutoDeleteDays,
			FirstUnreadMessageID: firstUnreadMessageID,
			Messages:             items,
		},
	}, 200, nil
}

// MarkConversationRead records the latest visible message the actor has actually reached.
func (s *Service) MarkConversationRead(conversationID int64, actor SessionPrincipal, req MarkConversationReadRequest) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("chat service unavailable")
	}
	if conversationID <= 0 || req.LastReadMessageID <= 0 {
		return Response{}, statusCode(ErrConversationNotFound), ErrConversationNotFound
	}
	if err := s.requireChatUser(actor.UserID); err != nil {
		return Response{}, statusCode(err), err
	}
	if _, err := s.repo.GetConversationForUser(actor.UserID, conversationID); err != nil {
		return Response{}, statusCode(err), err
	}
	if err := s.repo.MarkConversationReadUntil(actor.UserID, conversationID, req.LastReadMessageID); err != nil {
		return Response{}, statusCode(err), err
	}
	if err := s.repo.CancelPendingEmailNotificationIfNoUnread(actor.UserID, conversationID); err != nil {
		return Response{}, 500, err
	}

	return Response{
		Success: true,
		Code:    "CONVERSATION_READ_OK",
		Message: "已讀狀態已更新",
	}, 200, nil
}

// SearchMessages returns text messages visible to the current user that match the query.
func (s *Service) SearchMessages(actor SessionPrincipal, query string) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("chat service unavailable")
	}
	if err := s.requireChatUser(actor.UserID); err != nil {
		return Response{}, statusCode(err), err
	}

	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return Response{
			Success: true,
			Code:    "MESSAGE_SEARCH_OK",
			Message: "讯息搜寻完成",
			Data:    MessageSearchData{Query: trimmed, Messages: []MessageSearchItem{}},
		}, 200, nil
	}

	messages, err := s.repo.SearchMessages(actor.UserID, trimmed, messageSearchLimit)
	if err != nil {
		return Response{}, 500, err
	}

	items := make([]MessageSearchItem, 0, len(messages))
	for _, message := range messages {
		items = append(items, MessageSearchItem{
			ConversationID:    message.ConversationID,
			ConversationTitle: message.ConversationTitle,
			MessageID:         message.ID,
			SenderID:          message.SenderID,
			SenderName:        message.SenderName,
			MessageType:       message.MessageType,
			Content:           message.Content,
			CreatedAt:         message.CreatedAt.Format(time.RFC3339),
		})
	}

	return Response{
		Success: true,
		Code:    "MESSAGE_SEARCH_OK",
		Message: "讯息搜寻完成",
		Data: MessageSearchData{
			Query:    trimmed,
			Messages: items,
		},
	}, 200, nil
}

// SendMessage validates and stores a new message for the current user.
func (s *Service) SendMessage(conversationID int64, actor SessionPrincipal, req CreateMessageRequest) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("chat service unavailable")
	}
	if conversationID <= 0 {
		return Response{}, statusCode(ErrConversationNotFound), ErrConversationNotFound
	}
	if err := s.requireChatUser(actor.UserID); err != nil {
		return Response{}, statusCode(err), err
	}
	if muted, err := s.repo.IsChatMuted(actor.UserID); err != nil {
		return Response{}, statusCode(err), err
	} else if muted {
		return Response{}, statusCode(ErrUserChatMuted), ErrUserChatMuted
	}

	messageType := strings.ToLower(strings.TrimSpace(req.Type))
	if messageType != textMessageType && messageType != imageMessageType && messageType != fileMessageType {
		return Response{}, statusCode(ErrUnsupportedMessageType), ErrUnsupportedMessageType
	}

	content := strings.TrimSpace(req.Content)
	if messageType == textMessageType && content == "" {
		return Response{}, statusCode(ErrMessageContentRequired), ErrMessageContentRequired
	}
	attachments := normalizedAttachmentInputs(req)
	if messageType != textMessageType && len(attachments) == 0 {
		return Response{}, statusCode(ErrAttachmentRequired), ErrAttachmentRequired
	}

	conversation, err := s.repo.GetConversationForUser(actor.UserID, conversationID)
	if err != nil {
		return Response{}, statusCode(err), err
	}

	created, err := s.repo.CreateMessage(CreateMessageInput{
		ConversationID: conversationID,
		SenderID:       actor.UserID,
		MessageType:    messageType,
		Content:        content,
		Attachments:    attachments,
	})
	if err != nil {
		return Response{}, statusCode(err), err
	}
	var mentions []MessageMentionInput
	if mentions, err = s.messageMentions(conversationID, actor.UserID, content); err != nil {
		return Response{}, 500, err
	} else if len(mentions) > 0 {
		if err := s.repo.CreateMessageMentions(created.ID, conversationID, mentions); err != nil {
			return Response{}, 500, err
		}
	}
	if err := s.enqueueUnreadEmailNotifications(conversation, actor.UserID, created.ID, mentions); err != nil {
		log.Printf("enqueue unread email notifications conversation=%d message=%d: %v", conversation.ID, created.ID, err)
	}

	item := MessageItem{
		MessageID:   created.ID,
		SenderID:    created.SenderID,
		SenderName:  created.SenderName,
		MessageType: created.MessageType,
		Content:     created.Content,
		CreatedAt:   created.CreatedAt.Format(time.RFC3339),
		Attachment:  firstAttachmentItem(created),
		Attachments: attachmentItems(created.Attachments),
	}

	if s.broker != nil {
		memberIDs, err := s.repo.ListConversationMemberIDs(conversationID)
		if err != nil {
			return Response{}, 500, err
		}
		s.broker.PublishToUsers(memberIDs, RealtimeEvent{
			EventType:      "message.created",
			ConversationID: conversationID,
			Message:        item,
		})
	}

	return Response{
		Success: true,
		Code:    "MESSAGE_SENT",
		Message: "讯息发送成功",
		Data: SentMessageData{
			ConversationID: conversationID,
			Message:        item,
		},
	}, 201, nil
}

func (s *Service) enqueueUnreadEmailNotifications(conversation Conversation, senderUserID, messageID int64, mentions []MessageMentionInput) error {
	if messageID <= 0 || conversation.ID <= 0 {
		return nil
	}
	members, err := s.repo.ListConversationMembers(conversation.ID)
	if err != nil {
		return err
	}

	mentioned := make(map[int64]bool, len(mentions))
	for _, mention := range mentions {
		if mention.UserID <= 0 {
			continue
		}
		if mention.MentionType == "all" && conversation.Type != "group" {
			continue
		}
		mentioned[mention.UserID] = true
	}

	dueAt := time.Now().UTC().Add(emailNotificationDelay)
	for _, member := range members {
		if member.UserID <= 0 || member.UserID == senderUserID {
			continue
		}
		if member.NotificationMuted && !mentioned[member.UserID] {
			continue
		}
		if err := s.repo.EnqueueEmailNotification(member.UserID, conversation.ID, messageID, dueAt); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) messageMentions(conversationID, actorUserID int64, content string) ([]MessageMentionInput, error) {
	tokens := mentionTokens(content)
	if len(tokens) == 0 {
		return nil, nil
	}
	members, err := s.repo.ListConversationMembers(conversationID)
	if err != nil {
		return nil, err
	}

	hasAll := false
	wanted := make(map[string]bool)
	for _, token := range tokens {
		if strings.EqualFold(token, "ALL") {
			hasAll = true
			continue
		}
		wanted[strings.ToLower(token)] = true
	}

	seen := make(map[int64]bool)
	mentions := make([]MessageMentionInput, 0)
	for _, member := range members {
		if member.UserID == actorUserID || member.UserID == 0 || seen[member.UserID] {
			continue
		}
		mentionType := ""
		if hasAll {
			mentionType = "all"
		} else {
			externalID := strings.ToLower(strings.TrimSpace(member.ExternalUserID))
			displayName := strings.ToLower(strings.TrimSpace(member.DisplayName))
			if wanted[externalID] || (displayName != "" && wanted[displayName]) {
				mentionType = "user"
			}
		}
		if mentionType == "" {
			continue
		}
		seen[member.UserID] = true
		mentions = append(mentions, MessageMentionInput{UserID: member.UserID, MentionType: mentionType})
	}
	return mentions, nil
}

func mentionTokens(content string) []string {
	matches := mentionPattern.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}
	tokens := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			token := strings.TrimSpace(match[1])
			if token != "" {
				tokens = append(tokens, token)
			}
		}
	}
	return tokens
}

// DeleteMessage recalls a message sent by the current user.
func (s *Service) DeleteMessage(conversationID, messageID int64, actor SessionPrincipal) (Response, int, error) {
	if s == nil || s.repo == nil {
		return Response{}, 503, fmt.Errorf("chat service unavailable")
	}
	if conversationID <= 0 || messageID <= 0 {
		return Response{}, statusCode(ErrMessageNotFound), ErrMessageNotFound
	}
	if err := s.requireChatUser(actor.UserID); err != nil {
		return Response{}, statusCode(err), err
	}
	if _, err := s.repo.GetConversationForUser(actor.UserID, conversationID); err != nil {
		return Response{}, statusCode(err), err
	}
	if err := s.repo.DeleteMessage(conversationID, messageID, actor.UserID); err != nil {
		return Response{}, statusCode(err), err
	}

	if s.broker != nil {
		memberIDs, err := s.repo.ListConversationMemberIDs(conversationID)
		if err != nil {
			return Response{}, 500, err
		}
		s.broker.PublishToUsers(memberIDs, RealtimeEvent{
			EventType:      "message.recalled",
			ConversationID: conversationID,
			MessageID:      messageID,
		})
	}

	return Response{
		Success: true,
		Code:    "MESSAGE_RECALLED",
		Message: "訊息已收回",
		Data: RecalledMessageData{
			ConversationID: conversationID,
			MessageID:      messageID,
		},
	}, 200, nil
}

func (s *Service) requireChatUser(userID int64) error {
	if userID <= 0 {
		return ErrInsufficientRole
	}

	isAdmin, err := s.repo.IsSystemAdmin(userID)
	if err != nil {
		return err
	}
	if isAdmin {
		return ErrSystemAdminCannotChat
	}
	mustChange, err := s.repo.MustChangePassword(userID)
	if err != nil {
		return err
	}
	if mustChange {
		return ErrPasswordChangeRequired
	}

	return nil
}

func normalizedGroupMembers(members []GroupMemberRequest) []GroupMemberRequest {
	seen := make(map[string]struct{})
	normalized := make([]GroupMemberRequest, 0, len(members))
	for _, member := range members {
		source := strings.TrimSpace(member.SourceSystem)
		externalID := strings.TrimSpace(member.ExternalUserID)
		if source == "" || externalID == "" {
			continue
		}
		key := source + "\x00" + strings.ToLower(externalID)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, GroupMemberRequest{
			SourceSystem:   source,
			ExternalUserID: externalID,
		})
	}
	return normalized
}

func statusCode(err error) int {
	switch {
	case errors.Is(err, ErrUnsupportedMessageType), errors.Is(err, ErrMessageContentRequired), errors.Is(err, ErrAttachmentRequired), errors.Is(err, ErrAttachmentTooLarge), errors.Is(err, ErrUnsupportedAttachmentType), errors.Is(err, ErrGroupNameRequired), errors.Is(err, ErrGroupMemberRequired):
		return 400
	case errors.Is(err, ErrDirectChatSelfNotAllow), errors.Is(err, ErrContactSelfNotAllow):
		return 409
	case errors.Is(err, ErrSystemAdminCannotChat), errors.Is(err, ErrInsufficientRole), errors.Is(err, ErrUserChatMuted), errors.Is(err, ErrPasswordChangeRequired):
		return 403
	case errors.Is(err, ErrMessageRecallForbidden):
		return 403
	case errors.Is(err, ErrConversationNotFound), errors.Is(err, ErrMessageNotFound):
		return 404
	case errors.Is(err, ErrTargetUserNotFound):
		return 404
	default:
		return 500
	}
}

func allowedAutoDeleteOptions(maxDays int) []int {
	if maxDays <= 0 {
		return nil
	}
	if maxDays > 30 {
		maxDays = 30
	}
	options := make([]int, 0, len(conversationAutoDeleteChoices))
	for _, days := range conversationAutoDeleteChoices {
		if days <= maxDays {
			options = append(options, days)
		}
	}
	return options
}

func validAutoDeleteDays(days int, options []int) bool {
	if days == 0 {
		return true
	}
	for _, option := range options {
		if days == option {
			return true
		}
	}
	return false
}

func normalizedAttachmentInputs(req CreateMessageRequest) []AttachmentInput {
	if len(req.Attachments) > 0 {
		return req.Attachments
	}
	if req.Attachment != nil {
		return []AttachmentInput{*req.Attachment}
	}
	return nil
}

func firstAttachmentItem(message Message) *AttachmentItem {
	if len(message.Attachments) > 0 {
		return attachmentItem(&message.Attachments[0])
	}
	return attachmentItem(message.Attachment)
}

func attachmentItems(values []Attachment) []AttachmentItem {
	if len(values) == 0 {
		return nil
	}

	items := make([]AttachmentItem, 0, len(values))
	for index := range values {
		items = append(items, *attachmentItem(&values[index]))
	}
	return items
}

func attachmentItem(value *Attachment) *AttachmentItem {
	if value == nil {
		return nil
	}

	return &AttachmentItem{
		OriginalName: value.OriginalName,
		URL:          value.StoragePath,
		MIMEType:     value.MIMEType,
		SizeBytes:    value.SizeBytes,
	}
}
