package chat

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	messageListLimit   = 100
	messageSearchLimit = 30
	userSearchLimit    = 10
	textMessageType    = "text"
	imageMessageType   = "image"
	fileMessageType    = "file"
)

var (
	ErrConversationNotFound      = errors.New("conversation not found")
	ErrInsufficientRole          = errors.New("insufficient role")
	ErrSystemAdminCannotChat     = errors.New("system admin cannot chat")
	ErrUnsupportedMessageType    = errors.New("unsupported message type")
	ErrMessageContentRequired    = errors.New("message content required")
	ErrTargetUserNotFound        = errors.New("target user not found")
	ErrDirectChatSelfNotAllow    = errors.New("direct conversation with self is not allowed")
	ErrAttachmentRequired        = errors.New("attachment required")
	ErrAttachmentTooLarge        = errors.New("attachment too large")
	ErrUnsupportedAttachmentType = errors.New("unsupported attachment type")
)

// Service implements chat list and message list rules.
type Service struct {
	repo   Repository
	broker Broker
}

// NewService builds a chat service.
func NewService(repo Repository, broker Broker) *Service {
	return &Service{repo: repo, broker: broker}
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
		}
		if !conversation.LastMessageAt.IsZero() {
			item.LastMessageAt = conversation.LastMessageAt.Format(time.RFC3339)
		}
		items = append(items, item)
	}

	return Response{Success: true, Code: "CONVERSATIONS_OK", Message: "对话列表读取成功", Data: items}, 200, nil
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

	messages, err := s.repo.ListMessages(conversationID, messageListLimit)
	if err != nil {
		return Response{}, 500, err
	}
	if err := s.repo.MarkConversationRead(actor.UserID, conversationID); err != nil {
		return Response{}, 500, err
	}

	items := make([]MessageItem, 0, len(messages))
	for _, message := range messages {
		items = append(items, MessageItem{
			MessageID:   message.ID,
			SenderID:    message.SenderID,
			SenderName:  message.SenderName,
			MessageType: message.MessageType,
			Content:     message.Content,
			CreatedAt:   message.CreatedAt.Format(time.RFC3339),
			Attachment:  attachmentItem(message.Attachment),
		})
	}

	return Response{
		Success: true,
		Code:    "MESSAGES_OK",
		Message: "讯息列表读取成功",
		Data: MessageListData{
			ConversationID: conversation.ID,
			Type:           conversation.Type,
			Title:          conversation.Title,
			Messages:       items,
		},
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

	messageType := strings.ToLower(strings.TrimSpace(req.Type))
	if messageType != textMessageType && messageType != imageMessageType && messageType != fileMessageType {
		return Response{}, statusCode(ErrUnsupportedMessageType), ErrUnsupportedMessageType
	}

	content := strings.TrimSpace(req.Content)
	if messageType == textMessageType && content == "" {
		return Response{}, statusCode(ErrMessageContentRequired), ErrMessageContentRequired
	}
	if messageType != textMessageType && req.Attachment == nil {
		return Response{}, statusCode(ErrAttachmentRequired), ErrAttachmentRequired
	}

	if _, err := s.repo.GetConversationForUser(actor.UserID, conversationID); err != nil {
		return Response{}, statusCode(err), err
	}

	created, err := s.repo.CreateMessage(CreateMessageInput{
		ConversationID: conversationID,
		SenderID:       actor.UserID,
		MessageType:    messageType,
		Content:        content,
		Attachment:     req.Attachment,
	})
	if err != nil {
		return Response{}, statusCode(err), err
	}

	item := MessageItem{
		MessageID:   created.ID,
		SenderID:    created.SenderID,
		SenderName:  created.SenderName,
		MessageType: created.MessageType,
		Content:     created.Content,
		CreatedAt:   created.CreatedAt.Format(time.RFC3339),
		Attachment:  attachmentItem(created.Attachment),
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

	return nil
}

func statusCode(err error) int {
	switch {
	case errors.Is(err, ErrUnsupportedMessageType), errors.Is(err, ErrMessageContentRequired), errors.Is(err, ErrAttachmentRequired), errors.Is(err, ErrAttachmentTooLarge), errors.Is(err, ErrUnsupportedAttachmentType):
		return 400
	case errors.Is(err, ErrDirectChatSelfNotAllow):
		return 409
	case errors.Is(err, ErrSystemAdminCannotChat), errors.Is(err, ErrInsufficientRole):
		return 403
	case errors.Is(err, ErrConversationNotFound):
		return 404
	case errors.Is(err, ErrTargetUserNotFound):
		return 404
	default:
		return 500
	}
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
