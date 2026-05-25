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

	mux.HandleFunc("/", methodHandler("GET", func(w http.ResponseWriter, r *http.Request) { rootHandler(w, r, pages) }))
	mux.HandleFunc("/home", methodHandler("GET", func(w http.ResponseWriter, r *http.Request) { indexHandler(w, r, pages) }))
	mux.HandleFunc("/login", methodHandler("GET", func(w http.ResponseWriter, r *http.Request) { loginHandler(w, r, pages) }))
	mux.HandleFunc("/chat", methodHandler("GET", func(w http.ResponseWriter, r *http.Request) { desktopHandler(w, r, pages) }))
	mux.HandleFunc("/m/chat", methodHandler("GET", func(w http.ResponseWriter, r *http.Request) { mobileHandler(w, r, pages) }))
	mux.HandleFunc("/embed/chat", methodHandler("GET", func(w http.ResponseWriter, r *http.Request) { embedHandler(w, r, pages) }))
	mux.HandleFunc("/healthz", methodHandler("GET", healthHandler))

	if integrationHandler != nil {
		mux.HandleFunc("/api/erp/register", methodHandler("POST", integrationHandler.Register))
		mux.HandleFunc("/api/erp/login", methodHandler("POST", integrationHandler.Login))
		mux.HandleFunc("/api/system-admin/users/{user_id}/devices", methodHandler("GET", integrationHandler.ListDevices))
		mux.HandleFunc("/api/system-admin/users/{user_id}/ip-whitelist", methodHandler("PUT", integrationHandler.UpdateIPWhitelist))
		mux.HandleFunc("/api/users/me/profile", methodHandler("PATCH", integrationHandler.UpdateProfile))
		mux.HandleFunc("/api/users/{user_id}/devices/{device_id}/approve", methodHandler("POST", integrationHandler.ApproveDevice))
	}

	if chatHandler != nil {
		mux.HandleFunc("/ws", methodHandler("GET", chatHandler.ServeWebSocket))
		mux.HandleFunc("/api/conversations", methodHandler("GET", chatHandler.ListConversations))
		mux.HandleFunc("/api/conversations/direct", methodHandler("POST", chatHandler.CreateDirectConversation))
		mux.HandleFunc("/api/conversations/group", methodHandler("POST", chatHandler.CreateGroupConversation))
		mux.HandleFunc("/api/contacts", methodHandler("GET", chatHandler.ListContacts))
		mux.HandleFunc("/api/contacts", methodHandler("POST", chatHandler.AddContact))
		mux.HandleFunc("/api/users/search", methodHandler("GET", chatHandler.SearchUsers))
		mux.HandleFunc("/api/messages/search", methodHandler("GET", chatHandler.SearchMessages))
		mux.HandleFunc("/api/conversations/{conversation_id}", methodHandler("PATCH", chatHandler.UpdateConversation))
		mux.HandleFunc("/api/conversations/{conversation_id}/members", methodHandler("GET", chatHandler.ListConversationMembers))
		mux.HandleFunc("/api/conversations/{conversation_id}/members", methodHandler("POST", chatHandler.AddConversationMembers))
		mux.HandleFunc("/api/conversations/{conversation_id}/members", methodHandler("DELETE", chatHandler.RemoveConversationMember))
		mux.HandleFunc("/api/conversations/{conversation_id}/messages", methodHandler("GET", chatHandler.ListMessages))
		mux.HandleFunc("/api/conversations/{conversation_id}/messages", methodHandler("POST", chatHandler.SendMessage))
		mux.HandleFunc("/api/conversations/{conversation_id}/messages/{message_id}/recall", methodHandler("POST", chatHandler.RecallMessage))
		mux.HandleFunc("/api/conversations/{conversation_id}/messages/{message_id}", methodHandler("DELETE", chatHandler.DeleteMessage))
	}
}
