package httpx

import (
	"html/template"
	"net/http"
	"path/filepath"
	"strings"

	"1001-twacc-chat/internal/chat"
	"1001-twacc-chat/internal/erp"
)

type UIConfig struct {
	IntegrationSharedToken string
}

// NewRouter builds the HTTP routes used by the chat MVP scaffold.
func NewRouter(integrationHandler *erp.Handler, chatHandler *chat.Handler, uiConfig UIConfig) http.Handler {
	mux := http.NewServeMux()
	pages := pageData{
		IntegrationSharedToken: strings.TrimSpace(uiConfig.IntegrationSharedToken),
	}

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	mux.Handle("GET /uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(filepath.Join("web", "uploads")))))
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		indexHandler(w, r, pages)
	})
	mux.HandleFunc("GET /chat", func(w http.ResponseWriter, r *http.Request) {
		desktopHandler(w, r, pages)
	})
	mux.HandleFunc("GET /m/chat", func(w http.ResponseWriter, r *http.Request) {
		mobileHandler(w, r, pages)
	})
	mux.HandleFunc("GET /embed/chat", func(w http.ResponseWriter, r *http.Request) {
		embedHandler(w, r, pages)
	})
	mux.HandleFunc("GET /healthz", healthHandler)
	if integrationHandler != nil {
		mux.HandleFunc("POST /api/erp/register", integrationHandler.Register)
		mux.HandleFunc("POST /api/erp/login", integrationHandler.Login)
		mux.HandleFunc("GET /api/system-admin/users/{user_id}/devices", integrationHandler.ListDevices)
		mux.HandleFunc("PUT /api/system-admin/users/{user_id}/ip-whitelist", integrationHandler.UpdateIPWhitelist)
		mux.HandleFunc("POST /api/users/{user_id}/devices/{device_id}/approve", integrationHandler.ApproveDevice)
	}
	if chatHandler != nil {
		mux.HandleFunc("GET /ws", chatHandler.ServeWebSocket)
		mux.HandleFunc("GET /api/conversations", chatHandler.ListConversations)
		mux.HandleFunc("POST /api/conversations/direct", chatHandler.CreateDirectConversation)
		mux.HandleFunc("GET /api/conversations/{conversation_id}/messages", chatHandler.ListMessages)
		mux.HandleFunc("POST /api/conversations/{conversation_id}/messages", chatHandler.SendMessage)
	}

	return mux
}

func indexHandler(w http.ResponseWriter, r *http.Request, data pageData) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data.Title = "1001 TWACC Chat"
	renderTemplate(w, "index.html", data)
}

func desktopHandler(w http.ResponseWriter, _ *http.Request, data pageData) {
	data.Title = "TWACC Chat Desktop"
	renderTemplate(w, "desktop.html", data)
}

func mobileHandler(w http.ResponseWriter, _ *http.Request, data pageData) {
	data.Title = "TWACC Chat Mobile"
	renderTemplate(w, "mobile.html", data)
}

func embedHandler(w http.ResponseWriter, _ *http.Request, data pageData) {
	data.Title = "TWACC Chat Embed"
	renderTemplate(w, "embed.html", data)
}

type pageData struct {
	Title                  string
	IntegrationSharedToken string
}

func renderTemplate(w http.ResponseWriter, name string, data pageData) {
	tmpl, err := template.ParseFiles(filepath.Join("web", "templates", name))
	if err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
		return
	}
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
