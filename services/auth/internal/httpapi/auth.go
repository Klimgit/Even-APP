package httpapi

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	libjwt "github.com/even-app/even-app/libs/jwt"
	"github.com/even-app/even-app/libs/http/middleware"
	"github.com/even-app/even-app/services/auth/internal/store"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	Users      *store.UserStore
	JWT        *libjwt.Manager
	RefreshTTL time.Duration
}

func (h *AuthHandler) Register(mux *http.ServeMux, jwtMiddleware func(http.Handler) http.Handler) {
	mux.HandleFunc("POST /api/v1/auth/register", h.register)
	mux.HandleFunc("POST /api/v1/auth/login", h.login)
	mux.HandleFunc("POST /api/v1/auth/refresh", h.refresh)
	mux.Handle("GET /api/v1/auth/me", jwtMiddleware(http.HandlerFunc(h.me)))
}

type registerReq struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

func (h *AuthHandler) register(w http.ResponseWriter, r *http.Request) {
	var body registerReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	email := strings.TrimSpace(strings.ToLower(body.Email))
	if email == "" || len(body.Password) < 8 {
		writeErr(w, http.StatusBadRequest, "email and password (min 8) required")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "hash failed")
		return
	}
	var dn *string
	if body.DisplayName != "" {
		dn = &body.DisplayName
	}
	u, err := h.Users.Create(r.Context(), email, string(hash), body.Role, dn)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			writeErr(w, http.StatusConflict, "email already registered")
			return
		}
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeAuthResponse(w, r, u, http.StatusCreated)
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) login(w http.ResponseWriter, r *http.Request) {
	var body loginReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	u, err := h.Users.ByEmail(r.Context(), strings.TrimSpace(strings.ToLower(body.Email)))
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(body.Password)); err != nil {
		writeErr(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	h.writeAuthResponse(w, r, u, http.StatusOK)
}

type refreshReq struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) refresh(w http.ResponseWriter, r *http.Request) {
	var body refreshReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	hash := hashToken(body.RefreshToken)
	userID, err := h.Users.UserIDByRefreshHash(r.Context(), hash)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}
	_ = h.Users.DeleteRefreshToken(r.Context(), hash)
	u, err := h.Users.ByID(r.Context(), userID)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}
	access, refresh, err := h.issueTokens(r, u)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"access_token":  access,
		"refresh_token": refresh,
	})
}

func (h *AuthHandler) me(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	u, err := h.Users.ByID(r.Context(), claims.UserID)
	if err != nil {
		writeErr(w, http.StatusNotFound, "user not found")
		return
	}
	writeJSON(w, http.StatusOK, userDTO(u))
}

func (h *AuthHandler) writeAuthResponse(w http.ResponseWriter, r *http.Request, u *store.User, status int) {
	access, refresh, err := h.issueTokens(r, u)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, status, map[string]any{
		"access_token":  access,
		"refresh_token": refresh,
		"user":          userDTO(u),
	})
}

func (h *AuthHandler) issueTokens(r *http.Request, u *store.User) (access, refresh string, err error) {
	access, err = h.JWT.IssueAccess(u.ID, u.Role, u.IsAdmin)
	if err != nil {
		return "", "", err
	}
	refresh, err = newRefreshToken()
	if err != nil {
		return "", "", err
	}
	if err := h.Users.SaveRefreshToken(r.Context(), u.ID, hashToken(refresh), time.Now().Add(h.RefreshTTL)); err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

func userDTO(u *store.User) map[string]any {
	m := map[string]any{
		"id": u.ID.String(), "email": u.Email, "role": u.Role,
		"is_admin": u.IsAdmin, "created_at": u.CreatedAt.UTC().Format(time.RFC3339),
	}
	if u.DisplayName != nil {
		m["display_name"] = *u.DisplayName
	}
	return m
}

func newRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg, "message": msg})
}
