package httpx

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"html/template"
	"io"
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

	// register routes split between chat (frontend/API) and admin
	registerChatRoutes(mux, integrationHandler, chatHandler, pages)
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

func adminLoginProxyHandler(handler *erp.Handler, sharedToken string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Ensure we preserve and possibly augment the JSON request body
		// so that downstream integration.Login receives a device_id.
		if r.Body != nil {
			body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
			_ = r.Body.Close()
			if err == nil && len(body) > 0 {
				var payload map[string]any
				if json.Unmarshal(body, &payload) == nil {
					if _, ok := payload["device_id"]; !ok {
						// generate a short random device id
						b := make([]byte, 8)
						if _, err := rand.Read(b); err == nil {
							payload["device_id"] = "admin-" + hex.EncodeToString(b)
						} else {
							payload["device_id"] = "admin-unknown"
						}
						if nb, err := json.Marshal(payload); err == nil {
							r.Body = io.NopCloser(bytes.NewReader(nb))
							r.ContentLength = int64(len(nb))
						} else {
							r.Body = io.NopCloser(bytes.NewReader(body))
						}
					} else {
						r.Body = io.NopCloser(bytes.NewReader(body))
					}
				} else {
					// not JSON or unmarshal failed, restore original body
					r.Body = io.NopCloser(bytes.NewReader(body))
				}
			} else {
				// empty body: create minimal JSON with a device_id
				payload := map[string]any{"device_id": "admin-unknown"}
				if nb, err := json.Marshal(payload); err == nil {
					r.Body = io.NopCloser(bytes.NewReader(nb))
					r.ContentLength = int64(len(nb))
				}
			}
		}

		if handler == nil {
			writeJSONResponse(w, http.StatusServiceUnavailable, map[string]any{"success": false, "code": "SERVICE_UNAVAILABLE", "message": "服务尚未完成初始化"})
			return
		}

		if strings.TrimSpace(sharedToken) == "" {
			writeJSONResponse(w, http.StatusInternalServerError, map[string]any{"success": false, "code": "INTEGRATION_TOKEN_MISSING", "message": "服务尚未配置对接凭证"})
			return
		}

		r.Header.Set("Authorization", "Bearer "+strings.TrimSpace(sharedToken))
		handler.Login(w, r)
	}
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
