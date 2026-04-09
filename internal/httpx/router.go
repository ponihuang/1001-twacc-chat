package httpx

import (
	"html/template"
	"net/http"
	"path/filepath"

	"1001-twacc-chat/internal/chat"
	"1001-twacc-chat/internal/erp"
)

// NewRouter builds the HTTP routes used by the chat MVP scaffold.
func NewRouter(integrationHandler *erp.Handler, chatHandler *chat.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	mux.HandleFunc("GET /", indexHandler)
	mux.HandleFunc("GET /chat", desktopHandler)
	mux.HandleFunc("GET /m/chat", mobileHandler)
	mux.HandleFunc("GET /embed/chat", embedHandler)
	mux.HandleFunc("GET /healthz", healthHandler)
	if integrationHandler != nil {
		mux.HandleFunc("POST /api/erp/register", integrationHandler.Register)
		mux.HandleFunc("POST /api/erp/login", integrationHandler.Login)
		mux.HandleFunc("GET /api/system-admin/users/{user_id}/devices", integrationHandler.ListDevices)
		mux.HandleFunc("PUT /api/system-admin/users/{user_id}/ip-whitelist", integrationHandler.UpdateIPWhitelist)
		mux.HandleFunc("POST /api/users/{user_id}/devices/{device_id}/approve", integrationHandler.ApproveDevice)
	}
	if chatHandler != nil {
		mux.HandleFunc("GET /api/conversations", chatHandler.ListConversations)
		mux.HandleFunc("GET /api/conversations/{conversation_id}/messages", chatHandler.ListMessages)
		mux.HandleFunc("POST /api/conversations/{conversation_id}/messages", chatHandler.SendMessage)
	}

	return mux
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	renderTemplate(w, "index.html", pageData{Title: "1001 TWACC Chat"})
}

func desktopHandler(w http.ResponseWriter, _ *http.Request) {
	renderTemplate(w, "desktop.html", pageData{Title: "TWACC Chat Desktop"})
}

func mobileHandler(w http.ResponseWriter, _ *http.Request) {
	renderTemplate(w, "mobile.html", pageData{Title: "TWACC Chat Mobile"})
}

func embedHandler(w http.ResponseWriter, _ *http.Request) {
	renderTemplate(w, "embed.html", pageData{Title: "TWACC Chat Embed"})
}

type pageData struct {
	Title string
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
