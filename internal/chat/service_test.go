package chat

import (
	"errors"
	"testing"
	"time"
)

type mockRepository struct {
	admins               map[int64]bool
	conversations        map[int64][]ConversationSummary
	headers              map[string]Conversation
	messages             map[int64][]Message
	searchUsers          []UserSearchResult
	searchMessages       []Message
	members              []ConversationMember
	firstUnreadMessageID int64
	searchUserID         int64
	searchSourceSystem   string
	searchQuery          string
	searchLimit          int
	memberIDs            map[int64][]int64
	markedReads          []readMarker
	cancelledEmailReads  []readMarker
	createdMentions      []MessageMentionInput
	createdMessages      []CreateMessageInput
	enqueuedEmails       []emailEnqueue
	enqueueEmailErr      error
	createMessageResult  Message
	createMessageErr     error
	deletedMessages      []int64
	deleteMessageErr     error
	directConversation   Conversation
	directCreated        bool
	directErr            error
	contact              ContactData
	contactSourceSystem  string
	contactExternalID    string
	contactAliasName     string
	contactErr           error
	contacts             []ContactData
	listContactsErr      error
	groupConversation    Conversation
	groupName            string
	groupMembers         []GroupMemberRequest
	groupErr             error
	addedMembers         []GroupMemberRequest
	addedMemberIDs       []int64
	addMembersErr        error
	updatedGroup         Conversation
	updatedGroupName     string
	updatedGroupDesc     string
	updateGroupErr       error
	removedMemberID      int64
	removedSourceSystem  string
	removedExternalID    string
	removeMemberErr      error
	listMembersID        int64
	listMembersErr       error
	mutedUsers           map[int64]bool
	notificationMutes    map[string]bool
	recoverStaleCalls    int
	recoverStaleNow      time.Time
	recoverStaleBefore   time.Time
	recoverStaleErr      error
	claimSchedules       []*EmailNotificationSchedule
	claimTokens          []string
	claimErr             error
	delivery             EmailNotificationDelivery
	deliveryErr          error
	sentScheduleID       int64
	cancelledScheduleID  int64
	failedScheduleID     int64
	failedRetryAt        time.Time
	failedMaxAttempts    int
	failedErrMessage     string
}

type readMarker struct {
	userID         int64
	conversationID int64
	messageID      int64
}

type emailEnqueue struct {
	userID         int64
	conversationID int64
	messageID      int64
	dueAt          time.Time
}

func (m *mockRepository) IsSystemAdmin(userID int64) (bool, error) {
	return m.admins[userID], nil
}

func (m *mockRepository) IsChatMuted(userID int64) (bool, error) {
	return m.mutedUsers[userID], nil
}

func (m *mockRepository) MustChangePassword(userID int64) (bool, error) {
	return false, nil
}

func (m *mockRepository) ListConversations(userID int64) ([]ConversationSummary, error) {
	return append([]ConversationSummary(nil), m.conversations[userID]...), nil
}

func (m *mockRepository) SearchUsers(actorUserID int64, sourceSystem, query string, limit int) ([]UserSearchResult, error) {
	m.searchUserID = actorUserID
	m.searchSourceSystem = sourceSystem
	m.searchQuery = query
	m.searchLimit = limit
	return append([]UserSearchResult(nil), m.searchUsers...), nil
}

func (m *mockRepository) AddContact(actorUserID int64, sourceSystem, externalUserID, aliasName string) (ContactData, error) {
	m.searchUserID = actorUserID
	m.contactSourceSystem = sourceSystem
	m.contactExternalID = externalUserID
	m.contactAliasName = aliasName
	if m.contactErr != nil {
		return ContactData{}, m.contactErr
	}
	if m.contact.ContactUserID != 0 {
		return m.contact, nil
	}
	return ContactData{ContactUserID: 8, SourceSystem: sourceSystem, ExternalUserID: externalUserID, DisplayName: externalUserID, AliasName: aliasName, Status: "active"}, nil
}

func (m *mockRepository) ListContacts(actorUserID int64) ([]ContactData, error) {
	m.searchUserID = actorUserID
	if m.listContactsErr != nil {
		return nil, m.listContactsErr
	}
	return append([]ContactData(nil), m.contacts...), nil
}

func (m *mockRepository) GetConversationForUser(userID, conversationID int64) (Conversation, error) {
	key := conversationKey(userID, conversationID)
	conversation, ok := m.headers[key]
	if !ok {
		return Conversation{}, ErrConversationNotFound
	}
	return conversation, nil
}

func (m *mockRepository) UpdateConversationNotificationMute(userID, conversationID int64, muted bool) (Conversation, error) {
	key := conversationKey(userID, conversationID)
	conversation, ok := m.headers[key]
	if !ok {
		return Conversation{}, ErrConversationNotFound
	}
	if m.notificationMutes == nil {
		m.notificationMutes = make(map[string]bool)
	}
	m.notificationMutes[key] = muted
	conversation.NotificationMuted = muted
	m.headers[key] = conversation
	return conversation, nil
}

func (m *mockRepository) ListMessages(conversationID int64, limit int) ([]Message, error) {
	messages := append([]Message(nil), m.messages[conversationID]...)
	if len(messages) > limit {
		messages = messages[len(messages)-limit:]
	}
	return messages, nil
}

func (m *mockRepository) ListConversationMembers(conversationID int64) ([]ConversationMember, error) {
	m.listMembersID = conversationID
	if m.listMembersErr != nil {
		return nil, m.listMembersErr
	}
	return append([]ConversationMember(nil), m.members...), nil
}

func (m *mockRepository) SearchMessages(userID int64, query string, limit int) ([]Message, error) {
	m.searchUserID = userID
	m.searchQuery = query
	m.searchLimit = limit
	return append([]Message(nil), m.searchMessages...), nil
}

func (m *mockRepository) FirstUnreadMessageID(userID, conversationID int64) (int64, error) {
	return m.firstUnreadMessageID, nil
}

func (m *mockRepository) MarkConversationRead(userID, conversationID int64) error {
	m.markedReads = append(m.markedReads, readMarker{userID: userID, conversationID: conversationID})
	return nil
}

func (m *mockRepository) MarkConversationReadUntil(userID, conversationID, messageID int64) error {
	m.markedReads = append(m.markedReads, readMarker{userID: userID, conversationID: conversationID, messageID: messageID})
	return nil
}

func (m *mockRepository) CancelPendingEmailNotificationIfNoUnread(userID, conversationID int64) error {
	m.cancelledEmailReads = append(m.cancelledEmailReads, readMarker{userID: userID, conversationID: conversationID})
	return nil
}

func (m *mockRepository) CreateMessageMentions(messageID, conversationID int64, mentions []MessageMentionInput) error {
	m.createdMentions = append([]MessageMentionInput(nil), mentions...)
	return nil
}

func (m *mockRepository) EnqueueEmailNotification(userID, conversationID, messageID int64, dueAt time.Time) error {
	if m.enqueueEmailErr != nil {
		return m.enqueueEmailErr
	}
	m.enqueuedEmails = append(m.enqueuedEmails, emailEnqueue{userID: userID, conversationID: conversationID, messageID: messageID, dueAt: dueAt})
	return nil
}

func (m *mockRepository) RecoverStaleEmailNotifications(now time.Time, staleBefore time.Time) error {
	m.recoverStaleCalls++
	m.recoverStaleNow = now
	m.recoverStaleBefore = staleBefore
	return m.recoverStaleErr
}

func (m *mockRepository) ClaimDueEmailNotification(now time.Time, lockToken string) (*EmailNotificationSchedule, error) {
	m.claimTokens = append(m.claimTokens, lockToken)
	if m.claimErr != nil {
		return nil, m.claimErr
	}
	if len(m.claimSchedules) == 0 {
		return nil, nil
	}
	schedule := m.claimSchedules[0]
	m.claimSchedules = m.claimSchedules[1:]
	return schedule, nil
}

func (m *mockRepository) LoadEmailNotificationDelivery(schedule EmailNotificationSchedule) (EmailNotificationDelivery, error) {
	if m.deliveryErr != nil {
		return EmailNotificationDelivery{}, m.deliveryErr
	}
	delivery := m.delivery
	delivery.ScheduleID = schedule.ID
	if delivery.UserID == 0 {
		delivery.UserID = schedule.UserID
	}
	if delivery.ConversationID == 0 {
		delivery.ConversationID = schedule.ConversationID
	}
	return delivery, nil
}

func (m *mockRepository) MarkEmailNotificationSent(scheduleID int64, sentAt time.Time) error {
	m.sentScheduleID = scheduleID
	return nil
}

func (m *mockRepository) MarkEmailNotificationCancelled(scheduleID int64, cancelledAt time.Time) error {
	m.cancelledScheduleID = scheduleID
	return nil
}

func (m *mockRepository) MarkEmailNotificationFailed(scheduleID int64, retryAt time.Time, maxAttempts int, errMessage string) error {
	m.failedScheduleID = scheduleID
	m.failedRetryAt = retryAt
	m.failedMaxAttempts = maxAttempts
	m.failedErrMessage = errMessage
	return nil
}

func (m *mockRepository) CreateMessage(input CreateMessageInput) (Message, error) {
	if m.createMessageErr != nil {
		return Message{}, m.createMessageErr
	}

	m.createdMessages = append(m.createdMessages, input)
	if m.createMessageResult.ID != 0 {
		return m.createMessageResult, nil
	}

	attachment := attachmentFromInput(input.Attachment)
	attachments := attachmentsFromInput(input.Attachments)
	if attachment == nil && len(attachments) > 0 {
		attachment = &attachments[0]
	}
	return Message{
		ID:             int64(len(m.createdMessages)),
		ConversationID: input.ConversationID,
		SenderID:       input.SenderID,
		SenderName:     "你",
		MessageType:    input.MessageType,
		Content:        input.Content,
		CreatedAt:      time.Date(2026, 4, 7, 2, 0, 0, 0, time.UTC),
		Attachment:     attachment,
		Attachments:    attachments,
	}, nil
}

func (m *mockRepository) DeleteMessage(conversationID, messageID, actorUserID int64) error {
	if m.deleteMessageErr != nil {
		return m.deleteMessageErr
	}
	m.deletedMessages = append(m.deletedMessages, messageID)
	return nil
}

func (m *mockRepository) CreateOrGetDirectConversation(actorUserID int64, sourceSystem, externalUserID string) (Conversation, bool, error) {
	if m.directErr != nil {
		return Conversation{}, false, m.directErr
	}
	if m.directConversation.ID != 0 {
		return m.directConversation, m.directCreated, nil
	}
	return Conversation{ID: 15, Type: "direct", Title: externalUserID}, true, nil
}

func (m *mockRepository) CreateGroupConversation(actorUserID int64, name string, members []GroupMemberRequest) (Conversation, error) {
	if m.groupErr != nil {
		return Conversation{}, m.groupErr
	}
	m.groupName = name
	m.groupMembers = append([]GroupMemberRequest(nil), members...)
	if m.groupConversation.ID != 0 {
		return m.groupConversation, nil
	}
	return Conversation{ID: 25, Type: "group", Title: name}, nil
}

func (m *mockRepository) AddConversationMembers(conversationID int64, members []GroupMemberRequest) ([]int64, error) {
	if m.addMembersErr != nil {
		return nil, m.addMembersErr
	}
	m.listMembersID = conversationID
	m.addedMembers = append([]GroupMemberRequest(nil), members...)
	return append([]int64(nil), m.addedMemberIDs...), nil
}

func (m *mockRepository) UpdateGroupConversation(conversationID, actorUserID int64, name, description string) (Conversation, error) {
	if m.updateGroupErr != nil {
		return Conversation{}, m.updateGroupErr
	}
	m.listMembersID = conversationID
	m.searchUserID = actorUserID
	m.updatedGroupName = name
	m.updatedGroupDesc = description
	if m.headers != nil {
		m.headers[conversationKey(actorUserID, conversationID)] = Conversation{ID: conversationID, Type: "group", Title: name, Description: description}
	}
	if m.updatedGroup.ID != 0 {
		return m.updatedGroup, nil
	}
	return Conversation{ID: conversationID, Type: "group", Title: name, Description: description}, nil
}

func (m *mockRepository) RemoveConversationMember(conversationID, actorUserID int64, sourceSystem, externalUserID string) (int64, error) {
	if m.removeMemberErr != nil {
		return 0, m.removeMemberErr
	}
	m.listMembersID = conversationID
	m.searchUserID = actorUserID
	m.removedSourceSystem = sourceSystem
	m.removedExternalID = externalUserID
	if m.removedMemberID != 0 {
		return m.removedMemberID, nil
	}
	return 8, nil
}

func (m *mockRepository) ListConversationMemberIDs(conversationID int64) ([]int64, error) {
	return append([]int64(nil), m.memberIDs[conversationID]...), nil
}

type mockBroker struct {
	userIDs []int64
	event   RealtimeEvent
	calls   int
}

func (m *mockBroker) PublishToUsers(userIDs []int64, event RealtimeEvent) {
	m.userIDs = append([]int64(nil), userIDs...)
	m.event = event
	m.calls++
}

func (m *mockBroker) Subscribe(userID int64) (<-chan RealtimeEvent, func()) {
	ch := make(chan RealtimeEvent)
	close(ch)
	return ch, func() {}
}

func attachmentFromInput(value *AttachmentInput) *Attachment {
	if value == nil {
		return nil
	}

	return &Attachment{
		OriginalName: value.OriginalName,
		StoragePath:  value.StoragePath,
		MIMEType:     value.MIMEType,
		SizeBytes:    value.SizeBytes,
	}
}

func attachmentsFromInput(values []AttachmentInput) []Attachment {
	if len(values) == 0 {
		return nil
	}

	attachments := make([]Attachment, 0, len(values))
	for _, value := range values {
		attachments = append(attachments, Attachment{
			OriginalName: value.OriginalName,
			StoragePath:  value.StoragePath,
			MIMEType:     value.MIMEType,
			SizeBytes:    value.SizeBytes,
		})
	}
	return attachments
}

func conversationKey(userID, conversationID int64) string {
	return time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(userID + conversationID)).Format(time.RFC3339Nano)
}

func TestListConversations(t *testing.T) {
	repo := &mockRepository{
		conversations: map[int64][]ConversationSummary{
			7: {{ConversationID: 9, Type: "direct", Title: "採購小組", DirectSourceSystem: "erp", DirectExternalID: "user_b", MemberCount: 3, LastMessageType: "text", LastMessagePreview: "hello", LastMessageAt: time.Date(2026, 4, 7, 1, 2, 3, 0, time.UTC), UnreadCount: 2, HasUnreadMention: true}},
		},
		firstUnreadMessageID: 2,
	}
	service := NewService(repo, nil)

	resp, status, err := service.ListConversations(SessionPrincipal{UserID: 7})
	if err != nil {
		t.Fatalf("ListConversations returned error: %v", err)
	}
	if status != 200 {
		t.Fatalf("status = %d, want 200", status)
	}

	items, ok := resp.Data.([]ConversationItem)
	if !ok {
		t.Fatalf("unexpected data type %T", resp.Data)
	}
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
	if items[0].ConversationID != 9 || items[0].Title != "採購小組" {
		t.Fatalf("unexpected conversation item: %+v", items[0])
	}
	if items[0].UnreadCount != 2 {
		t.Fatalf("unread_count = %d, want 2", items[0].UnreadCount)
	}
	if !items[0].HasUnreadMention {
		t.Fatalf("has_unread_mention = false, want true")
	}
	if items[0].DirectSourceSystem != "erp" || items[0].DirectExternalID != "user_b" {
		t.Fatalf("unexpected direct peer fields: %+v", items[0])
	}
}

func TestListConversationsRejectsSystemAdmin(t *testing.T) {
	repo := &mockRepository{admins: map[int64]bool{9: true}}
	service := NewService(repo, nil)

	_, _, err := service.ListConversations(SessionPrincipal{UserID: 9})
	if !errors.Is(err, ErrSystemAdminCannotChat) {
		t.Fatalf("expected ErrSystemAdminCannotChat, got %v", err)
	}
}

func TestSearchUsers(t *testing.T) {
	repo := &mockRepository{
		admins: map[int64]bool{},
		searchUsers: []UserSearchResult{
			{SourceSystem: "erp", ExternalUserID: "test01", DisplayName: "Test One"},
		},
	}
	service := NewService(repo, nil)

	resp, status, err := service.SearchUsers(SessionPrincipal{UserID: 7}, "erp", "@test")
	if err != nil {
		t.Fatalf("SearchUsers returned error: %v", err)
	}
	if status != 200 || !resp.Success {
		t.Fatalf("unexpected response status=%d resp=%+v", status, resp)
	}
	if repo.searchUserID != 7 || repo.searchSourceSystem != "erp" || repo.searchQuery != "test" || repo.searchLimit != userSearchLimit {
		t.Fatalf("unexpected search call: user=%d source=%q query=%q limit=%d", repo.searchUserID, repo.searchSourceSystem, repo.searchQuery, repo.searchLimit)
	}
	data, ok := resp.Data.(UserSearchData)
	if !ok {
		t.Fatalf("response data type = %T, want UserSearchData", resp.Data)
	}
	if len(data.Users) != 1 || data.Users[0].ExternalUserID != "test01" || data.Users[0].DisplayName != "Test One" {
		t.Fatalf("unexpected search data: %+v", data)
	}
}

func TestAddContact(t *testing.T) {
	repo := &mockRepository{
		contact: ContactData{
			ContactUserID:  8,
			SourceSystem:   "erp",
			ExternalUserID: "test01",
			DisplayName:    "Test One",
			AliasName:      "測試",
			Status:         "active",
		},
	}
	service := NewService(repo, nil)

	resp, status, err := service.AddContact(SessionPrincipal{UserID: 7}, AddContactRequest{
		SourceSystem:   "erp",
		ExternalUserID: "test01",
		AliasName:      " 測試 ",
	})
	if err != nil {
		t.Fatalf("AddContact returned error: %v", err)
	}
	if status != 201 {
		t.Fatalf("status = %d, want 201", status)
	}
	if repo.searchUserID != 7 || repo.contactSourceSystem != "erp" || repo.contactExternalID != "test01" || repo.contactAliasName != "測試" {
		t.Fatalf("unexpected contact call: user=%d source=%q external=%q alias=%q", repo.searchUserID, repo.contactSourceSystem, repo.contactExternalID, repo.contactAliasName)
	}
	data, ok := resp.Data.(ContactData)
	if !ok {
		t.Fatalf("response data type = %T, want ContactData", resp.Data)
	}
	if data.ContactUserID != 8 || data.AliasName != "測試" {
		t.Fatalf("unexpected contact data: %+v", data)
	}
}

func TestListContacts(t *testing.T) {
	repo := &mockRepository{
		contacts: []ContactData{
			{
				ContactUserID:  8,
				SourceSystem:   "erp",
				ExternalUserID: "test01",
				DisplayName:    "Test One",
				AliasName:      "測試",
				Status:         "active",
			},
		},
	}
	service := NewService(repo, nil)

	resp, status, err := service.ListContacts(SessionPrincipal{UserID: 7})
	if err != nil {
		t.Fatalf("ListContacts returned error: %v", err)
	}
	if status != 200 {
		t.Fatalf("status = %d, want 200", status)
	}
	if repo.searchUserID != 7 {
		t.Fatalf("unexpected actor user id: %d", repo.searchUserID)
	}
	data, ok := resp.Data.(ContactListData)
	if !ok {
		t.Fatalf("response data type = %T, want ContactListData", resp.Data)
	}
	if len(data.Contacts) != 1 || data.Contacts[0].ExternalUserID != "test01" {
		t.Fatalf("unexpected contacts data: %+v", data)
	}
}

func TestCreateGroupConversationCreatesGroupAndPublishesEvent(t *testing.T) {
	repo := &mockRepository{
		groupConversation: Conversation{ID: 31, Type: "group", Title: "測試群組"},
		memberIDs: map[int64][]int64{
			31: {7, 8, 9},
		},
	}
	broker := &mockBroker{}
	service := NewService(repo, broker)

	resp, status, err := service.CreateGroupConversation(SessionPrincipal{UserID: 7}, CreateGroupConversationRequest{
		Name: " 測試群組 ",
		Members: []GroupMemberRequest{
			{SourceSystem: "erp", ExternalUserID: "u8"},
			{SourceSystem: "erp", ExternalUserID: "u8"},
			{SourceSystem: "erp", ExternalUserID: "u9"},
		},
	})
	if err != nil {
		t.Fatalf("CreateGroupConversation returned error: %v", err)
	}
	if status != 201 {
		t.Fatalf("status = %d, want 201", status)
	}
	if repo.groupName != "測試群組" {
		t.Fatalf("group name = %q, want 測試群組", repo.groupName)
	}
	if len(repo.groupMembers) != 2 {
		t.Fatalf("group members = %+v, want 2 unique members", repo.groupMembers)
	}
	data, ok := resp.Data.(DirectConversationData)
	if !ok {
		t.Fatalf("response data type = %T, want DirectConversationData", resp.Data)
	}
	if data.ConversationID != 31 || data.Type != "group" || data.Title != "測試群組" {
		t.Fatalf("unexpected group data: %+v", data)
	}
	if broker.calls != 1 || broker.event.EventType != "conversation.ready" || broker.event.ConversationID != 31 {
		t.Fatalf("unexpected broker event: %+v", broker.event)
	}
}

func TestCreateGroupConversationRejectsMissingNameOrMembers(t *testing.T) {
	service := NewService(&mockRepository{}, nil)

	_, status, err := service.CreateGroupConversation(SessionPrincipal{UserID: 7}, CreateGroupConversationRequest{
		Members: []GroupMemberRequest{{SourceSystem: "erp", ExternalUserID: "u8"}},
	})
	if !errors.Is(err, ErrGroupNameRequired) || status != 400 {
		t.Fatalf("missing name err=%v status=%d, want ErrGroupNameRequired 400", err, status)
	}

	_, status, err = service.CreateGroupConversation(SessionPrincipal{UserID: 7}, CreateGroupConversationRequest{Name: "測試群組"})
	if !errors.Is(err, ErrGroupMemberRequired) || status != 400 {
		t.Fatalf("missing members err=%v status=%d, want ErrGroupMemberRequired 400", err, status)
	}
}

func TestListMessages(t *testing.T) {
	repo := &mockRepository{
		headers: map[string]Conversation{
			conversationKey(7, 9): {ID: 9, Type: "direct", Title: "王小明"},
		},
		messages: map[int64][]Message{
			9: {
				{ID: 1, ConversationID: 9, SenderID: 7, SenderName: "你", MessageType: "text", Content: "hello", CreatedAt: time.Date(2026, 4, 7, 1, 0, 0, 0, time.UTC)},
				{ID: 2, ConversationID: 9, SenderID: 8, SenderName: "王小明", MessageType: "text", Content: "hi", CreatedAt: time.Date(2026, 4, 7, 1, 1, 0, 0, time.UTC)},
			},
		},
		firstUnreadMessageID: 2,
	}
	service := NewService(repo, nil)

	resp, status, err := service.ListMessages(9, SessionPrincipal{UserID: 7})
	if err != nil {
		t.Fatalf("ListMessages returned error: %v", err)
	}
	if status != 200 {
		t.Fatalf("status = %d, want 200", status)
	}

	data, ok := resp.Data.(MessageListData)
	if !ok {
		t.Fatalf("unexpected data type %T", resp.Data)
	}
	if data.ConversationID != 9 || data.Title != "王小明" {
		t.Fatalf("unexpected message list data: %+v", data)
	}
	if len(data.Messages) != 2 {
		t.Fatalf("len(messages) = %d, want 2", len(data.Messages))
	}
	if data.FirstUnreadMessageID != 2 {
		t.Fatalf("first unread message id = %d, want 2", data.FirstUnreadMessageID)
	}
	if len(repo.markedReads) != 0 {
		t.Fatalf("unexpected read markers: %+v", repo.markedReads)
	}
}

func TestMarkConversationRead(t *testing.T) {
	repo := &mockRepository{
		headers: map[string]Conversation{
			conversationKey(7, 9): {ID: 9, Type: "direct", Title: "王小明"},
		},
	}
	service := NewService(repo, nil)

	resp, status, err := service.MarkConversationRead(9, SessionPrincipal{UserID: 7}, MarkConversationReadRequest{LastReadMessageID: 22})
	if err != nil {
		t.Fatalf("MarkConversationRead returned error: %v", err)
	}
	if status != 200 || !resp.Success {
		t.Fatalf("unexpected response status=%d resp=%+v", status, resp)
	}
	if len(repo.markedReads) != 1 || repo.markedReads[0] != (readMarker{userID: 7, conversationID: 9, messageID: 22}) {
		t.Fatalf("unexpected read markers: %+v", repo.markedReads)
	}
	if len(repo.cancelledEmailReads) != 1 || repo.cancelledEmailReads[0] != (readMarker{userID: 7, conversationID: 9}) {
		t.Fatalf("unexpected email notification cancel calls: %+v", repo.cancelledEmailReads)
	}
}

func TestListMessagesRejectsMissingConversation(t *testing.T) {
	service := NewService(&mockRepository{}, nil)

	_, _, err := service.ListMessages(9, SessionPrincipal{UserID: 7})
	if !errors.Is(err, ErrConversationNotFound) {
		t.Fatalf("expected ErrConversationNotFound, got %v", err)
	}
}

func TestListConversationMembers(t *testing.T) {
	repo := &mockRepository{
		headers: map[string]Conversation{
			conversationKey(7, 31): {ID: 31, Type: "group", Title: "測試群組"},
		},
		members: []ConversationMember{
			{UserID: 7, SourceSystem: "erp", ExternalUserID: "u7", DisplayName: "王小明", Role: "owner"},
			{UserID: 8, SourceSystem: "erp", ExternalUserID: "u8", DisplayName: "陳小華", Role: "member"},
		},
	}
	service := NewService(repo, nil)

	resp, status, err := service.ListConversationMembers(SessionPrincipal{UserID: 7}, 31)
	if err != nil {
		t.Fatalf("ListConversationMembers returned error: %v", err)
	}
	if status != 200 || !resp.Success {
		t.Fatalf("unexpected response status=%d resp=%+v", status, resp)
	}
	data, ok := resp.Data.(ConversationMembersData)
	if !ok {
		t.Fatalf("response data type = %T, want ConversationMembersData", resp.Data)
	}
	if repo.listMembersID != 31 {
		t.Fatalf("listMembersID = %d, want 31", repo.listMembersID)
	}
	if data.ConversationID != 31 || data.Title != "測試群組" || data.MemberCount != 2 {
		t.Fatalf("unexpected members data: %+v", data)
	}
	if len(data.Members) != 2 || data.Members[0].Role != "owner" || data.Members[1].ExternalUserID != "u8" {
		t.Fatalf("unexpected members: %+v", data.Members)
	}
}

func TestAddConversationMembers(t *testing.T) {
	repo := &mockRepository{
		headers: map[string]Conversation{
			conversationKey(7, 31): {ID: 31, Type: "group", Title: "測試群組"},
		},
		addedMemberIDs: []int64{8},
		members: []ConversationMember{
			{UserID: 7, SourceSystem: "erp", ExternalUserID: "u7", DisplayName: "王小明", Role: "owner"},
			{UserID: 8, SourceSystem: "erp", ExternalUserID: "u8", DisplayName: "陳小華", Role: "member"},
		},
	}
	broker := &mockBroker{}
	service := NewService(repo, broker)

	resp, status, err := service.AddConversationMembers(SessionPrincipal{UserID: 7}, 31, AddConversationMembersRequest{
		Members: []GroupMemberRequest{{SourceSystem: "erp", ExternalUserID: "u8"}},
	})
	if err != nil {
		t.Fatalf("AddConversationMembers returned error: %v", err)
	}
	if status != 200 || !resp.Success {
		t.Fatalf("unexpected response status=%d resp=%+v", status, resp)
	}
	if len(repo.addedMembers) != 1 || repo.addedMembers[0].ExternalUserID != "u8" {
		t.Fatalf("unexpected added members: %+v", repo.addedMembers)
	}
	if broker.calls != 1 || len(broker.userIDs) != 1 || broker.userIDs[0] != 8 || broker.event.EventType != "conversation.ready" {
		t.Fatalf("unexpected broker event: users=%+v event=%+v", broker.userIDs, broker.event)
	}
	data, ok := resp.Data.(ConversationMembersData)
	if !ok || data.MemberCount != 2 {
		t.Fatalf("unexpected response data: %#v", resp.Data)
	}
}

func TestUpdateConversationUpdatesGroupMetadata(t *testing.T) {
	repo := &mockRepository{
		headers: map[string]Conversation{
			conversationKey(7, 31): {ID: 31, Type: "group", Title: "舊群組", Description: "舊描述"},
		},
		memberIDs: map[int64][]int64{31: {7, 8}},
		members: []ConversationMember{
			{UserID: 7, SourceSystem: "erp", ExternalUserID: "u7", DisplayName: "王小明", Role: "owner"},
			{UserID: 8, SourceSystem: "erp", ExternalUserID: "u8", DisplayName: "陳小華", Role: "member"},
		},
	}
	broker := &mockBroker{}
	service := NewService(repo, broker)

	resp, status, err := service.UpdateConversation(SessionPrincipal{UserID: 7}, 31, UpdateConversationRequest{
		Name:        "新群組",
		Description: "新描述",
	})
	if err != nil {
		t.Fatalf("UpdateConversation returned error: %v", err)
	}
	if status != 200 || !resp.Success {
		t.Fatalf("unexpected response status=%d resp=%+v", status, resp)
	}
	if repo.updatedGroupName != "新群組" || repo.updatedGroupDesc != "新描述" {
		t.Fatalf("unexpected updated metadata: name=%q desc=%q", repo.updatedGroupName, repo.updatedGroupDesc)
	}
	if broker.calls != 1 || broker.event.EventType != "conversation.members.updated" || len(broker.userIDs) != 2 {
		t.Fatalf("unexpected broker event: users=%+v event=%+v", broker.userIDs, broker.event)
	}
	data, ok := resp.Data.(ConversationMembersData)
	if !ok || data.Title != "新群組" || data.Description != "新描述" {
		t.Fatalf("unexpected response data: %#v", resp.Data)
	}
}

func TestRemoveConversationMember(t *testing.T) {
	repo := &mockRepository{
		headers: map[string]Conversation{
			conversationKey(7, 31): {ID: 31, Type: "group", Title: "測試群組"},
		},
		removedMemberID: 8,
		memberIDs:       map[int64][]int64{31: {7, 9}},
		members: []ConversationMember{
			{UserID: 7, SourceSystem: "erp", ExternalUserID: "u7", DisplayName: "王小明", Role: "owner"},
			{UserID: 9, SourceSystem: "erp", ExternalUserID: "u9", DisplayName: "林小美", Role: "member"},
		},
	}
	broker := &mockBroker{}
	service := NewService(repo, broker)

	resp, status, err := service.RemoveConversationMember(SessionPrincipal{UserID: 7}, 31, RemoveConversationMemberRequest{
		SourceSystem:   "erp",
		ExternalUserID: "u8",
	})
	if err != nil {
		t.Fatalf("RemoveConversationMember returned error: %v", err)
	}
	if status != 200 || !resp.Success {
		t.Fatalf("unexpected response status=%d resp=%+v", status, resp)
	}
	if repo.removedSourceSystem != "erp" || repo.removedExternalID != "u8" {
		t.Fatalf("unexpected removed member: source=%q external=%q", repo.removedSourceSystem, repo.removedExternalID)
	}
	if broker.calls != 1 || broker.event.EventType != "conversation.members.updated" {
		t.Fatalf("unexpected broker event: users=%+v event=%+v", broker.userIDs, broker.event)
	}
	if len(broker.userIDs) != 3 || broker.userIDs[2] != 8 {
		t.Fatalf("removed user should be included in broker recipients: %+v", broker.userIDs)
	}
	data, ok := resp.Data.(ConversationMembersData)
	if !ok || data.MemberCount != 2 {
		t.Fatalf("unexpected response data: %#v", resp.Data)
	}
}

func TestSearchMessages(t *testing.T) {
	repo := &mockRepository{
		admins: map[int64]bool{},
		searchMessages: []Message{
			{
				ID:                3,
				ConversationID:    9,
				ConversationTitle: "開發",
				SenderID:          8,
				SenderName:        "王小明",
				MessageType:       "text",
				Content:           "hello search",
				CreatedAt:         time.Date(2026, 4, 7, 1, 2, 0, 0, time.UTC),
			},
		},
	}
	service := NewService(repo, nil)

	resp, status, err := service.SearchMessages(SessionPrincipal{UserID: 7}, " search ")
	if err != nil {
		t.Fatalf("SearchMessages returned error: %v", err)
	}
	if status != 200 || !resp.Success {
		t.Fatalf("unexpected response status=%d resp=%+v", status, resp)
	}
	if repo.searchUserID != 7 || repo.searchQuery != "search" || repo.searchLimit != messageSearchLimit {
		t.Fatalf("unexpected search call: user=%d query=%q limit=%d", repo.searchUserID, repo.searchQuery, repo.searchLimit)
	}
	data, ok := resp.Data.(MessageSearchData)
	if !ok {
		t.Fatalf("response data type = %T, want MessageSearchData", resp.Data)
	}
	if len(data.Messages) != 1 || data.Messages[0].ConversationTitle != "開發" || data.Messages[0].Content != "hello search" {
		t.Fatalf("unexpected search data: %+v", data)
	}
}

func TestCreateDirectConversation(t *testing.T) {
	repo := &mockRepository{
		directConversation: Conversation{ID: 13, Type: "direct", Title: "User B"},
		directCreated:      true,
		memberIDs: map[int64][]int64{
			13: {7, 8},
		},
	}
	broker := &mockBroker{}
	service := NewService(repo, broker)

	resp, status, err := service.CreateDirectConversation(SessionPrincipal{UserID: 7}, CreateDirectConversationRequest{
		SourceSystem:   "erp",
		ExternalUserID: "user_b",
	})
	if err != nil {
		t.Fatalf("CreateDirectConversation returned error: %v", err)
	}
	if status != 201 {
		t.Fatalf("status = %d, want 201", status)
	}

	data, ok := resp.Data.(DirectConversationData)
	if !ok {
		t.Fatalf("unexpected data type %T", resp.Data)
	}
	if data.ConversationID != 13 || data.Title != "User B" {
		t.Fatalf("unexpected direct conversation data: %+v", data)
	}
	if broker.calls != 1 {
		t.Fatalf("broker calls = %d, want 1", broker.calls)
	}
	if broker.event.EventType != "conversation.ready" || broker.event.ConversationID != 13 {
		t.Fatalf("unexpected realtime event: %+v", broker.event)
	}
}

func TestSendMessage(t *testing.T) {
	repo := &mockRepository{
		headers: map[string]Conversation{
			conversationKey(7, 9): {ID: 9, Type: "direct", Title: "王小明"},
		},
		memberIDs: map[int64][]int64{
			9: {7, 8},
		},
		members: []ConversationMember{
			{UserID: 7, SourceSystem: "erp", ExternalUserID: "sender", DisplayName: "Sender"},
			{UserID: 8, SourceSystem: "erp", ExternalUserID: "receiver", DisplayName: "Receiver"},
		},
		createMessageResult: Message{
			ID:             3,
			ConversationID: 9,
			SenderID:       7,
			SenderName:     "你",
			MessageType:    "text",
			Content:        "hello",
			CreatedAt:      time.Date(2026, 4, 7, 1, 2, 0, 0, time.UTC),
		},
	}
	broker := &mockBroker{}
	service := NewService(repo, broker)

	resp, status, err := service.SendMessage(9, SessionPrincipal{UserID: 7}, CreateMessageRequest{Type: "TEXT", Content: "  hello  "})
	if err != nil {
		t.Fatalf("SendMessage returned error: %v", err)
	}
	if status != 201 {
		t.Fatalf("status = %d, want 201", status)
	}
	if len(repo.createdMessages) != 1 {
		t.Fatalf("len(createdMessages) = %d, want 1", len(repo.createdMessages))
	}
	if repo.createdMessages[0].MessageType != "text" || repo.createdMessages[0].Content != "hello" {
		t.Fatalf("unexpected created message input: %+v", repo.createdMessages[0])
	}

	data, ok := resp.Data.(SentMessageData)
	if !ok {
		t.Fatalf("unexpected data type %T", resp.Data)
	}
	if data.ConversationID != 9 || data.Message.MessageID != 3 {
		t.Fatalf("unexpected sent message data: %+v", data)
	}
	if broker.calls != 1 {
		t.Fatalf("broker calls = %d, want 1", broker.calls)
	}
	if broker.event.EventType != "message.created" || broker.event.ConversationID != 9 {
		t.Fatalf("unexpected realtime event: %+v", broker.event)
	}
	if len(repo.enqueuedEmails) != 1 {
		t.Fatalf("len(enqueuedEmails) = %d, want 1", len(repo.enqueuedEmails))
	}
	if repo.enqueuedEmails[0].userID != 8 || repo.enqueuedEmails[0].userID == 7 {
		t.Fatalf("unexpected email enqueue recipients: %+v", repo.enqueuedEmails)
	}
}

func TestSendMessageDoesNotFailWhenEmailNotificationEnqueueFails(t *testing.T) {
	repo := &mockRepository{
		headers: map[string]Conversation{
			conversationKey(7, 9): {ID: 9, Type: "direct", Title: "王小明"},
		},
		memberIDs: map[int64][]int64{9: {7, 8}},
		members: []ConversationMember{
			{UserID: 7, SourceSystem: "erp", ExternalUserID: "sender", DisplayName: "Sender"},
			{UserID: 8, SourceSystem: "erp", ExternalUserID: "receiver", DisplayName: "Receiver"},
		},
		createMessageResult: Message{
			ID:             3,
			ConversationID: 9,
			SenderID:       7,
			SenderName:     "你",
			MessageType:    "text",
			Content:        "hello",
			CreatedAt:      time.Date(2026, 4, 7, 1, 2, 0, 0, time.UTC),
		},
		enqueueEmailErr: errors.New("email queue unavailable"),
	}
	service := NewService(repo, nil)

	resp, status, err := service.SendMessage(9, SessionPrincipal{UserID: 7}, CreateMessageRequest{Type: "text", Content: "hello"})
	if err != nil {
		t.Fatalf("SendMessage returned error: %v", err)
	}
	if status != 201 || !resp.Success {
		t.Fatalf("unexpected response status=%d resp=%+v", status, resp)
	}
	if len(repo.createdMessages) != 1 {
		t.Fatalf("message should still be created when email enqueue fails")
	}
}

func TestSendMessageSkipsMutedConversationNormalMessage(t *testing.T) {
	repo := &mockRepository{
		headers: map[string]Conversation{
			conversationKey(7, 9): {ID: 9, Type: "direct", Title: "王小明"},
		},
		members: []ConversationMember{
			{UserID: 7, SourceSystem: "erp", ExternalUserID: "sender", DisplayName: "Sender"},
			{UserID: 8, SourceSystem: "erp", ExternalUserID: "receiver", DisplayName: "Receiver", NotificationMuted: true},
		},
		createMessageResult: Message{
			ID:             4,
			ConversationID: 9,
			SenderID:       7,
			SenderName:     "你",
			MessageType:    "text",
			Content:        "hello",
			CreatedAt:      time.Date(2026, 4, 7, 1, 2, 0, 0, time.UTC),
		},
	}
	service := NewService(repo, nil)

	_, status, err := service.SendMessage(9, SessionPrincipal{UserID: 7}, CreateMessageRequest{Type: "text", Content: "hello"})
	if err != nil || status != 201 {
		t.Fatalf("SendMessage err=%v status=%d, want nil 201", err, status)
	}
	if len(repo.enqueuedEmails) != 0 {
		t.Fatalf("muted normal message should not enqueue email: %+v", repo.enqueuedEmails)
	}
}

func TestSendMessageEnqueuesMutedConversationMentionedUser(t *testing.T) {
	repo := &mockRepository{
		headers: map[string]Conversation{
			conversationKey(7, 9): {ID: 9, Type: "group", Title: "測試群組"},
		},
		members: []ConversationMember{
			{UserID: 7, SourceSystem: "erp", ExternalUserID: "sender", DisplayName: "Sender"},
			{UserID: 8, SourceSystem: "erp", ExternalUserID: "etest01", DisplayName: "E Test", NotificationMuted: true},
			{UserID: 9, SourceSystem: "erp", ExternalUserID: "other", DisplayName: "Other", NotificationMuted: true},
		},
		createMessageResult: Message{
			ID:             5,
			ConversationID: 9,
			SenderID:       7,
			SenderName:     "Sender",
			MessageType:    "text",
			Content:        "@etest01 hello",
			CreatedAt:      time.Date(2026, 4, 7, 1, 2, 0, 0, time.UTC),
		},
	}
	service := NewService(repo, nil)

	_, status, err := service.SendMessage(9, SessionPrincipal{UserID: 7}, CreateMessageRequest{Type: "text", Content: "@etest01 hello"})
	if err != nil || status != 201 {
		t.Fatalf("SendMessage err=%v status=%d, want nil 201", err, status)
	}
	if len(repo.enqueuedEmails) != 1 || repo.enqueuedEmails[0].userID != 8 {
		t.Fatalf("expected only mentioned muted user to be enqueued: %+v", repo.enqueuedEmails)
	}
}

func TestSendMessageEnqueuesMutedConversationAllMention(t *testing.T) {
	repo := &mockRepository{
		headers: map[string]Conversation{
			conversationKey(7, 9): {ID: 9, Type: "group", Title: "測試群組"},
		},
		members: []ConversationMember{
			{UserID: 7, SourceSystem: "erp", ExternalUserID: "sender", DisplayName: "Sender"},
			{UserID: 8, SourceSystem: "erp", ExternalUserID: "one", DisplayName: "One", NotificationMuted: true},
			{UserID: 9, SourceSystem: "erp", ExternalUserID: "two", DisplayName: "Two", NotificationMuted: true},
		},
		createMessageResult: Message{
			ID:             6,
			ConversationID: 9,
			SenderID:       7,
			SenderName:     "Sender",
			MessageType:    "text",
			Content:        "@ALL hello",
			CreatedAt:      time.Date(2026, 4, 7, 1, 2, 0, 0, time.UTC),
		},
	}
	service := NewService(repo, nil)

	_, status, err := service.SendMessage(9, SessionPrincipal{UserID: 7}, CreateMessageRequest{Type: "text", Content: "@ALL hello"})
	if err != nil || status != 201 {
		t.Fatalf("SendMessage err=%v status=%d, want nil 201", err, status)
	}
	if len(repo.enqueuedEmails) != 2 {
		t.Fatalf("expected two muted @ALL recipients: %+v", repo.enqueuedEmails)
	}
	for _, item := range repo.enqueuedEmails {
		if item.userID == 7 {
			t.Fatalf("sender should not be enqueued: %+v", repo.enqueuedEmails)
		}
	}
}

func TestSendMessageRejectsMutedUser(t *testing.T) {
	repo := &mockRepository{
		headers: map[string]Conversation{
			conversationKey(7, 9): {ID: 9, Type: "direct", Title: "王小明"},
		},
		mutedUsers: map[int64]bool{7: true},
	}
	service := NewService(repo, nil)

	_, status, err := service.SendMessage(9, SessionPrincipal{UserID: 7}, CreateMessageRequest{Type: "text", Content: "hello"})
	if !errors.Is(err, ErrUserChatMuted) {
		t.Fatalf("expected ErrUserChatMuted, got %v", err)
	}
	if status != 403 {
		t.Fatalf("status = %d, want 403", status)
	}
	if len(repo.createdMessages) != 0 {
		t.Fatalf("message should not be created when user is muted")
	}
}

func TestSendMessageCreatesMentions(t *testing.T) {
	repo := &mockRepository{
		headers: map[string]Conversation{
			conversationKey(7, 9): {ID: 9, Type: "group", Title: "測試群組"},
		},
		memberIDs: map[int64][]int64{
			9: {7, 8, 9},
		},
		members: []ConversationMember{
			{UserID: 7, SourceSystem: "erp", ExternalUserID: "sender", DisplayName: "Sender"},
			{UserID: 8, SourceSystem: "erp", ExternalUserID: "etest01", DisplayName: "E Test"},
			{UserID: 9, SourceSystem: "erp", ExternalUserID: "elva", DisplayName: "Elva"},
		},
		createMessageResult: Message{
			ID:             33,
			ConversationID: 9,
			SenderID:       7,
			SenderName:     "Sender",
			MessageType:    "text",
			Content:        "@etest01 hello",
			CreatedAt:      time.Date(2026, 4, 7, 1, 2, 0, 0, time.UTC),
		},
	}
	service := NewService(repo, nil)

	_, status, err := service.SendMessage(9, SessionPrincipal{UserID: 7}, CreateMessageRequest{Type: "text", Content: "@etest01 hello"})
	if err != nil {
		t.Fatalf("SendMessage returned error: %v", err)
	}
	if status != 201 {
		t.Fatalf("status = %d, want 201", status)
	}
	if len(repo.createdMentions) != 1 || repo.createdMentions[0] != (MessageMentionInput{UserID: 8, MentionType: "user"}) {
		t.Fatalf("unexpected mentions: %+v", repo.createdMentions)
	}
}

func TestSendMessageRejectsUnsupportedType(t *testing.T) {
	service := NewService(&mockRepository{}, nil)

	_, status, err := service.SendMessage(9, SessionPrincipal{UserID: 7}, CreateMessageRequest{Type: "voice", Content: "hello"})
	if !errors.Is(err, ErrUnsupportedMessageType) {
		t.Fatalf("expected ErrUnsupportedMessageType, got %v", err)
	}
	if status != 400 {
		t.Fatalf("status = %d, want 400", status)
	}
}

func TestSendImageMessage(t *testing.T) {
	repo := &mockRepository{
		headers: map[string]Conversation{
			conversationKey(7, 9): {ID: 9, Type: "direct", Title: "王小明"},
		},
		memberIDs: map[int64][]int64{
			9: {7, 8},
		},
	}
	service := NewService(repo, nil)

	resp, status, err := service.SendMessage(9, SessionPrincipal{UserID: 7}, CreateMessageRequest{
		Type:    "image",
		Content: "",
		Attachment: &AttachmentInput{
			OriginalName: "photo.png",
			StoragePath:  "/uploads/photo.png",
			MIMEType:     "image/png",
			SizeBytes:    1234,
		},
	})
	if err != nil {
		t.Fatalf("SendMessage returned error: %v", err)
	}
	if status != 201 {
		t.Fatalf("status = %d, want 201", status)
	}

	data := resp.Data.(SentMessageData)
	if data.Message.Attachment == nil || data.Message.Attachment.URL != "/uploads/photo.png" {
		t.Fatalf("unexpected attachment payload: %+v", data.Message.Attachment)
	}
}

func TestSendMessageRejectsEmptyContent(t *testing.T) {
	service := NewService(&mockRepository{}, nil)

	_, status, err := service.SendMessage(9, SessionPrincipal{UserID: 7}, CreateMessageRequest{Type: "text", Content: "   "})
	if !errors.Is(err, ErrMessageContentRequired) {
		t.Fatalf("expected ErrMessageContentRequired, got %v", err)
	}
	if status != 400 {
		t.Fatalf("status = %d, want 400", status)
	}
}

func TestSendMessageRejectsMissingConversation(t *testing.T) {
	service := NewService(&mockRepository{}, nil)

	_, status, err := service.SendMessage(9, SessionPrincipal{UserID: 7}, CreateMessageRequest{Type: "text", Content: "hello"})
	if !errors.Is(err, ErrConversationNotFound) {
		t.Fatalf("expected ErrConversationNotFound, got %v", err)
	}
	if status != 404 {
		t.Fatalf("status = %d, want 404", status)
	}
}

func TestDeleteMessageRecallsOwnMessageAndPublishesEvent(t *testing.T) {
	repo := &mockRepository{
		headers: map[string]Conversation{
			conversationKey(7, 9): {ID: 9, Type: "direct", Title: "王小明"},
		},
		memberIDs: map[int64][]int64{
			9: {7, 8},
		},
	}
	broker := &mockBroker{}
	service := NewService(repo, broker)

	resp, status, err := service.DeleteMessage(9, 22, SessionPrincipal{UserID: 7})
	if err != nil {
		t.Fatalf("DeleteMessage returned error: %v", err)
	}
	if status != 200 {
		t.Fatalf("status = %d, want 200", status)
	}
	if len(repo.deletedMessages) != 1 || repo.deletedMessages[0] != 22 {
		t.Fatalf("deleted messages = %+v, want [22]", repo.deletedMessages)
	}
	if broker.calls != 1 || broker.event.EventType != "message.recalled" || broker.event.MessageID != 22 {
		t.Fatalf("unexpected broker event: %+v", broker.event)
	}
	data := resp.Data.(RecalledMessageData)
	if data.ConversationID != 9 || data.MessageID != 22 {
		t.Fatalf("unexpected deleted message data: %+v", data)
	}
}
