package httpx

import (
	"net/http"
	"path/filepath"

	"1001-twacc-chat/internal/chat"
	"1001-twacc-chat/internal/erp"
)

// registerChatRoutes registers public (frontend) routes and API endpoints for chat.
func registerChatRoutes(mux *http.ServeMux, integrationHandler *erp.Handler, chatHandler *chat.Handler, pages pageData) {
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(filepath.Join("web", "uploads")))))

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) { rootHandler(w, r, pages) })
	mux.HandleFunc("GET /home", func(w http.ResponseWriter, r *http.Request) { indexHandler(w, r, pages) })
	mux.HandleFunc("GET /login", func(w http.ResponseWriter, r *http.Request) { loginHandler(w, r, pages) })
	mux.HandleFunc("GET /chat", func(w http.ResponseWriter, r *http.Request) { desktopHandler(w, r, pages) })
	mux.HandleFunc("GET /m/chat", func(w http.ResponseWriter, r *http.Request) { mobileHandler(w, r, pages) })
	mux.HandleFunc("GET /embed/chat", func(w http.ResponseWriter, r *http.Request) { embedHandler(w, r, pages) })
	mux.HandleFunc("GET /healthz", healthHandler)

	if integrationHandler != nil {
		mux.HandleFunc("POST /api/erp/register", integrationHandler.Register)
		mux.HandleFunc("POST /api/erp/login", integrationHandler.Login)
		mux.HandleFunc("GET /api/system-admin/users/{user_id}/devices", integrationHandler.ListDevices)
		mux.HandleFunc("PUT /api/system-admin/users/{user_id}/ip-whitelist", integrationHandler.UpdateIPWhitelist)
		mux.HandleFunc("PATCH /api/users/me/profile", integrationHandler.UpdateProfile)
		mux.HandleFunc("POST /api/users/{user_id}/devices/{device_id}/approve", integrationHandler.ApproveDevice)
	}

	if chatHandler != nil {
		mux.HandleFunc("GET /ws", chatHandler.ServeWebSocket)
		mux.HandleFunc("GET /api/conversations", chatHandler.ListConversations)
		mux.HandleFunc("POST /api/conversations/direct", chatHandler.CreateDirectConversation)
		mux.HandleFunc("POST /api/conversations/group", chatHandler.CreateGroupConversation)
		mux.HandleFunc("GET /api/contacts", chatHandler.ListContacts)
		mux.HandleFunc("POST /api/contacts", chatHandler.AddContact)
		mux.HandleFunc("GET /api/users/search", chatHandler.SearchUsers)
		mux.HandleFunc("GET /api/messages/search", chatHandler.SearchMessages)
		mux.HandleFunc("PATCH /api/conversations/{conversation_id}", chatHandler.UpdateConversation)
		mux.HandleFunc("GET /api/conversations/{conversation_id}/members", chatHandler.ListConversationMembers)
		mux.HandleFunc("POST /api/conversations/{conversation_id}/members", chatHandler.AddConversationMembers)
		mux.HandleFunc("DELETE /api/conversations/{conversation_id}/members", chatHandler.RemoveConversationMember)
		mux.HandleFunc("GET /api/conversations/{conversation_id}/messages", chatHandler.ListMessages)
		mux.HandleFunc("POST /api/conversations/{conversation_id}/messages", chatHandler.SendMessage)
		mux.HandleFunc("POST /api/conversations/{conversation_id}/messages/{message_id}/recall", chatHandler.RecallMessage)
		mux.HandleFunc("DELETE /api/conversations/{conversation_id}/messages/{message_id}", chatHandler.DeleteMessage)
	}
}
