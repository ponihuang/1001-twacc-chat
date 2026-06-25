package httpx

import (
	"net/http"

	"1001-twacc-chat/internal/erp"
)

// registerAdminRoutes registers admin UI routes and the admin login proxy.
func registerAdminRoutes(mux *http.ServeMux, integrationHandler *erp.Handler, pages pageData) {
	mux.HandleFunc("/admin/login", methodHandler("GET", adminLoginHandler))
	mux.HandleFunc("/admin/dashboard", methodHandler("GET", redirectAdminDashboardHandler))
	mux.HandleFunc("/office", methodHandler("GET", adminDashboardHandler))
	mux.HandleFunc("/office/user", methodHandler("GET", adminDashboardHandler))
	mux.HandleFunc("/office/user/create", methodHandler("GET", adminDashboardHandler))
	mux.HandleFunc("GET /office/user/{user_id}/edit", adminDashboardHandler)
	mux.HandleFunc("/office/conversations", methodHandler("GET", adminDashboardHandler))
	mux.HandleFunc("/office/admins", methodHandler("GET", adminDashboardHandler))
	mux.HandleFunc("/admin/api/login", methodHandler("POST", adminLoginProxyHandler(integrationHandler, pages.IntegrationSharedToken)))
	mux.Handle("/admin/static/", http.StripPrefix("/admin/static/", http.FileServer(http.Dir("chat_admin/static"))))
}

func redirectAdminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/office", http.StatusFound)
}
