package httpx

import (
	"net/http"
	"path/filepath"

	"1001-twacc-chat/internal/chat"
	"1001-twacc-chat/internal/erp"
)

type methodRouter struct {
	handlers map[string]http.HandlerFunc
}

func (m *methodRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handler, ok := m.handlers[r.Method]; ok {
		handler(w, r)
		return
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func handleFunc(mux *http.ServeMux, registry map[string]*methodRouter, method, pattern string, handler http.HandlerFunc) {
	router, ok := registry[pattern]
	if !ok {
		router = &methodRouter{handlers: make(map[string]http.HandlerFunc)}
		registry[pattern] = router
		mux.Handle(pattern, router)
	}
	router.handlers[method] = handler
}

// registerChatRoutes registers public (frontend) routes and API endpoints for chat.
func registerChatRoutes(mux *http.ServeMux, integrationHandler *erp.Handler, chatHandler *chat.Handler, pages pageData) {
	routeRegistry := make(map[string]*methodRouter)

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(filepath.Join("web", "uploads")))))

	handleFunc(mux, routeRegistry, "GET", "/", func(w http.ResponseWriter, r *http.Request) { rootHandler(w, r, pages) })
	handleFunc(mux, routeRegistry, "GET", "/home", func(w http.ResponseWriter, r *http.Request) { indexHandler(w, r, pages) })
	handleFunc(mux, routeRegistry, "GET", "/login", func(w http.ResponseWriter, r *http.Request) { loginHandler(w, r, pages) })
	handleFunc(mux, routeRegistry, "GET", "/invite", func(w http.ResponseWriter, r *http.Request) { inviteHandler(w, r, pages) })
	handleFunc(mux, routeRegistry, "GET", "/chat", func(w http.ResponseWriter, r *http.Request) { desktopHandler(w, r, pages) })
	handleFunc(mux, routeRegistry, "GET", "/m/chat", func(w http.ResponseWriter, r *http.Request) { mobileHandler(w, r, pages) })
	handleFunc(mux, routeRegistry, "GET", "/embed/chat", func(w http.ResponseWriter, r *http.Request) { embedHandler(w, r, pages) })
	handleFunc(mux, routeRegistry, "GET", "/healthz", healthHandler)

	if integrationHandler != nil {
		handleFunc(mux, routeRegistry, "GET", "/api/user-invitations/{token}", integrationHandler.GetPublicUserInvitation)
		handleFunc(mux, routeRegistry, "POST", "/api/user-invitations/accept", integrationHandler.AcceptUserInvitation)
		handleFunc(mux, routeRegistry, "POST", "/api/erp/register", integrationHandler.Register)
		handleFunc(mux, routeRegistry, "POST", "/api/erp/login", integrationHandler.Login)
		handleFunc(mux, routeRegistry, "GET", "/api/system-admin/admins", integrationHandler.ListSystemAdmins)
		handleFunc(mux, routeRegistry, "POST", "/api/system-admin/admins", integrationHandler.CreateSystemAdmin)
		handleFunc(mux, routeRegistry, "GET", "/api/system-admin/admins/{user_id}", integrationHandler.GetSystemAdmin)
		handleFunc(mux, routeRegistry, "PATCH", "/api/system-admin/admins/{user_id}", integrationHandler.UpdateSystemAdmin)
		handleFunc(mux, routeRegistry, "GET", "/api/system-admin/me/permissions", integrationHandler.GetCurrentAdminPermissions)
		handleFunc(mux, routeRegistry, "GET", "/api/system-admin/roles", integrationHandler.ListAdminRoles)
		handleFunc(mux, routeRegistry, "POST", "/api/system-admin/roles", integrationHandler.CreateAdminRole)
		handleFunc(mux, routeRegistry, "GET", "/api/system-admin/roles/{role_id}", integrationHandler.GetAdminRole)
		handleFunc(mux, routeRegistry, "PATCH", "/api/system-admin/roles/{role_id}", integrationHandler.UpdateAdminRole)
		handleFunc(mux, routeRegistry, "GET", "/api/system-admin/roles/{role_id}/permissions", integrationHandler.GetAdminRolePermissions)
		handleFunc(mux, routeRegistry, "PATCH", "/api/system-admin/roles/{role_id}/permissions", integrationHandler.UpdateAdminRolePermissions)
		handleFunc(mux, routeRegistry, "GET", "/api/system-admin/settings", integrationHandler.GetSystemSettings)
		handleFunc(mux, routeRegistry, "PATCH", "/api/system-admin/settings", integrationHandler.UpdateSystemSettings)
		handleFunc(mux, routeRegistry, "GET", "/api/system-admin/conversations", integrationHandler.ListAdminConversations)
		handleFunc(mux, routeRegistry, "GET", "/api/system-admin/conversations/{conversation_id}", integrationHandler.GetAdminConversationDetail)
		handleFunc(mux, routeRegistry, "GET", "/api/system-admin/users", integrationHandler.ListUsers)
		handleFunc(mux, routeRegistry, "POST", "/api/system-admin/users", integrationHandler.CreateUser)
		handleFunc(mux, routeRegistry, "GET", "/api/system-admin/user-invitations", integrationHandler.ListUserInvitations)
		handleFunc(mux, routeRegistry, "POST", "/api/system-admin/user-invitations", integrationHandler.InviteUser)
		handleFunc(mux, routeRegistry, "POST", "/api/system-admin/user-invitations/bulk-precheck", integrationHandler.PrecheckUserInvitationTXT)
		handleFunc(mux, routeRegistry, "POST", "/api/system-admin/user-invitations/bulk-send", integrationHandler.SendBulkUserInvitations)
		handleFunc(mux, routeRegistry, "POST", "/api/system-admin/user-invitations/resend", integrationHandler.ResendUserInvitation)
		handleFunc(mux, routeRegistry, "GET", "/api/system-admin/users/{user_id}", integrationHandler.GetUser)
		handleFunc(mux, routeRegistry, "PATCH", "/api/system-admin/users/{user_id}", integrationHandler.UpdateUser)
		handleFunc(mux, routeRegistry, "PATCH", "/api/system-admin/users/{user_id}/chat-mute", integrationHandler.UpdateUserChatMute)
		handleFunc(mux, routeRegistry, "POST", "/api/system-admin/users/{user_id}/temporary-password", integrationHandler.SendTemporaryPassword)
		handleFunc(mux, routeRegistry, "GET", "/api/system-admin/users/{user_id}/devices", integrationHandler.ListDevices)
		handleFunc(mux, routeRegistry, "PUT", "/api/system-admin/users/{user_id}/ip-whitelist", integrationHandler.UpdateIPWhitelist)
		handleFunc(mux, routeRegistry, "GET", "/api/users/me/profile", integrationHandler.GetProfile)
		handleFunc(mux, routeRegistry, "PATCH", "/api/users/me/profile", integrationHandler.UpdateProfile)
		handleFunc(mux, routeRegistry, "PATCH", "/api/users/me/password", integrationHandler.UpdatePassword)
		handleFunc(mux, routeRegistry, "POST", "/api/users/{user_id}/devices/{device_id}/approve", integrationHandler.ApproveDevice)
	}

	if chatHandler != nil {
		handleFunc(mux, routeRegistry, "GET", "/ws", chatHandler.ServeWebSocket)
		handleFunc(mux, routeRegistry, "GET", "/api/conversations", chatHandler.ListConversations)
		handleFunc(mux, routeRegistry, "POST", "/api/conversations/direct", chatHandler.CreateDirectConversation)
		handleFunc(mux, routeRegistry, "POST", "/api/conversations/group", chatHandler.CreateGroupConversation)
		handleFunc(mux, routeRegistry, "GET", "/api/contacts", chatHandler.ListContacts)
		handleFunc(mux, routeRegistry, "POST", "/api/contacts", chatHandler.AddContact)
		handleFunc(mux, routeRegistry, "GET", "/api/users/search", chatHandler.SearchUsers)
		handleFunc(mux, routeRegistry, "GET", "/api/messages/search", chatHandler.SearchMessages)
		handleFunc(mux, routeRegistry, "PATCH", "/api/conversations/{conversation_id}", chatHandler.UpdateConversation)
		handleFunc(mux, routeRegistry, "PATCH", "/api/conversations/{conversation_id}/notification-mute", chatHandler.UpdateConversationNotificationMute)
		handleFunc(mux, routeRegistry, "PATCH", "/api/conversations/{conversation_id}/auto-delete", chatHandler.UpdateConversationAutoDelete)
		handleFunc(mux, routeRegistry, "GET", "/api/conversations/{conversation_id}/members", chatHandler.ListConversationMembers)
		handleFunc(mux, routeRegistry, "POST", "/api/conversations/{conversation_id}/members", chatHandler.AddConversationMembers)
		handleFunc(mux, routeRegistry, "DELETE", "/api/conversations/{conversation_id}/members", chatHandler.RemoveConversationMember)
		handleFunc(mux, routeRegistry, "GET", "/api/conversations/{conversation_id}/messages", chatHandler.ListMessages)
		handleFunc(mux, routeRegistry, "POST", "/api/conversations/{conversation_id}/messages", chatHandler.SendMessage)
		handleFunc(mux, routeRegistry, "POST", "/api/conversations/{conversation_id}/read", chatHandler.MarkConversationRead)
		handleFunc(mux, routeRegistry, "POST", "/api/conversations/{conversation_id}/messages/{message_id}/recall", chatHandler.RecallMessage)
		handleFunc(mux, routeRegistry, "DELETE", "/api/conversations/{conversation_id}/messages/{message_id}", chatHandler.DeleteMessage)
	}
}
