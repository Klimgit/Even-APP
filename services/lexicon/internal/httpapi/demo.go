package httpapi

import (
	"net/http"

	"github.com/even-app/even-app/libs/http/middleware"
	libjwt "github.com/even-app/even-app/libs/jwt"
	"github.com/even-app/even-app/services/lexicon/internal/store"
	"github.com/google/uuid"
)

type DemoHandler struct {
	Store *store.DemoStore
}

func (h *DemoHandler) Register(mux *http.ServeMux, jwt func(http.Handler) http.Handler) {
	mux.HandleFunc("GET /api/v1/platform/demo/public", h.public)
	mux.Handle("GET /api/v1/platform/demo/auth", jwt(http.HandlerFunc(h.auth)))
	mux.Handle("GET /api/v1/platform/demo/ping", jwt(http.HandlerFunc(h.auth)))
	mux.Handle("GET /api/v1/platform/demo/admin", jwt(http.HandlerFunc(h.admin)))
	mux.Handle("GET /api/v1/platform/demo/teacher", jwt(http.HandlerFunc(h.teacher)))
	mux.Handle("GET /api/v1/platform/demo/owner", jwt(http.HandlerFunc(h.owner)))
}

func (h *DemoHandler) public(w http.ResponseWriter, r *http.Request) {
	h.writeDemo(w, r, "public", nil)
}

func (h *DemoHandler) auth(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	h.writeDemo(w, r, "jwt", &claims)
}

func (h *DemoHandler) admin(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	claims, _ := middleware.ClaimsFromContext(r.Context())
	h.writeDemo(w, r, "platform-admin", &claims)
}

func (h *DemoHandler) teacher(w http.ResponseWriter, r *http.Request) {
	if !requireTeacher(w, r) {
		return
	}
	claims, _ := middleware.ClaimsFromContext(r.Context())
	h.writeDemo(w, r, "teacher", &claims)
}

func (h *DemoHandler) owner(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	ownerID := r.URL.Query().Get("user_id")
	if ownerID == "" {
		writeErr(w, http.StatusBadRequest, "user_id query required")
		return
	}
	if _, err := uuid.Parse(ownerID); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid user_id")
		return
	}
	if !claims.IsAdmin && claims.UserID.String() != ownerID {
		writeErr(w, http.StatusForbidden, "not resource owner")
		return
	}
	h.writeDemo(w, r, "owner", &claims)
}

func (h *DemoHandler) writeDemo(w http.ResponseWriter, r *http.Request, auth string, claims *libjwt.Claims) {
	n, err := h.Store.Ping(r.Context())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	body := map[string]any{
		"message": "pong",
		"service": "lexicon",
		"auth":    auth,
		"db":      n,
	}
	if claims != nil {
		body["user_id"] = claims.UserID.String()
		body["role"] = claims.Role
		body["is_admin"] = claims.IsAdmin
	}
	writeJSON(w, http.StatusOK, body)
}

func requireTeacher(w http.ResponseWriter, r *http.Request) bool {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok || (claims.Role != "teacher" && !claims.IsAdmin) {
		writeErr(w, http.StatusForbidden, "teacher required")
		return false
	}
	return true
}
