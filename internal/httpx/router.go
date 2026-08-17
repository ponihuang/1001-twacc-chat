package httpx

import (
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"

	"1001-twacc-chat/internal/chat"
	"1001-twacc-chat/internal/erp"
	"1001-twacc-chat/internal/telegram"
)

type UIConfig struct {
	IntegrationSharedToken string
}

// NewRouter builds the HTTP routes used by the chat MVP scaffold.
func NewRouter(integrationHandler *erp.Handler, chatHandler *chat.Handler, telegramHandler *telegram.Handler, uiConfig UIConfig) http.Handler {
	mux := http.NewServeMux()
	pages := pageData{
		IntegrationSharedToken: strings.TrimSpace(uiConfig.IntegrationSharedToken),
	}

	// register routes split between chat (frontend/API) and admin
	registerChatRoutes(mux, integrationHandler, chatHandler, telegramHandler, pages)
	registerAdminRoutes(mux, integrationHandler, pages)

	return mux
}

func methodHandler(method string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.Header().Set("Allow", method)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler(w, r)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request, data pageData) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data.Title = "1001 TWACC Chat"
	renderTemplate(w, "root.html", data)
}

func indexHandler(w http.ResponseWriter, r *http.Request, data pageData) {
	if r.URL.Path != "/home" {
		http.NotFound(w, r)
		return
	}

	data.Title = "1001 TWACC Chat"
	renderTemplate(w, "index.html", data)
}

func loginHandler(w http.ResponseWriter, _ *http.Request, data pageData) {
	data.Title = "TWACC Chat Login"
	renderTemplate(w, "login.html", data)
}

func inviteHandler(w http.ResponseWriter, _ *http.Request, data pageData) {
	data.Title = "TWACC Chat Invite"
	renderTemplate(w, "invite.html", data)
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

func writeJSONResponse(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if payload == nil {
		return
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	_, _ = w.Write(data)
}

func adminLoginHandler(w http.ResponseWriter, _ *http.Request) {
	data := pageData{Title: "系統管理後台 - 登入"}
	renderAdminTemplate(w, "admin-login.html", data)
}

func adminDashboardHandler(w http.ResponseWriter, _ *http.Request) {
	data := pageData{Title: "管理儀表板"}
	renderAdminTemplate(w, "admin-dashboard.html", data)
}

func renderAdminTemplate(w http.ResponseWriter, name string, data pageData) {
	tmpl, err := template.ParseFiles(filepath.Join("chat_admin", "templates", name))
	if err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
		return
	}
}
