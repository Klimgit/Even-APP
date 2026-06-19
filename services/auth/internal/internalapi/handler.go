package internalapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/even-app/even-app/libs/clients/dto"
	"github.com/even-app/even-app/services/auth/internal/gen/query"
	"github.com/jackc/pgx/v5"
)

type Handler struct {
	q *query.Queries
}

func New(q *query.Queries) *Handler {
	return &Handler{q: q}
}

func (h *Handler) Mount(mux *http.ServeMux) {
	mux.HandleFunc("GET /users/by-email", h.getUserByEmail)
	mux.HandleFunc("POST /users/by-ids", h.listUsersByIDs)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	inner := http.NewServeMux()
	h.Mount(inner)
	inner.ServeHTTP(w, r)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"message": msg, "error": msg})
}

func (h *Handler) getUserByEmail(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		writeErr(w, http.StatusBadRequest, "email required")
		return
	}
	row, err := h.q.GetUserByEmail(r.Context(), query.GetUserByEmailParams{Email: email})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeErr(w, http.StatusNotFound, "not found")
			return
		}
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, dto.UserView{
		ID: row.ID, Email: row.Email, DisplayName: row.DisplayName,
		Role: row.Role, IsAdmin: row.IsAdmin,
	})
}

func (h *Handler) listUsersByIDs(w http.ResponseWriter, r *http.Request) {
	var req dto.IDsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	rows, err := h.q.ListUsersByIDs(r.Context(), query.ListUsersByIDsParams{Column1: req.IDs})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]dto.UserView, 0, len(rows))
	for _, row := range rows {
		out = append(out, dto.UserView{
			ID: row.ID, Email: row.Email, DisplayName: row.DisplayName,
			Role: row.Role, IsAdmin: row.IsAdmin,
		})
	}
	writeJSON(w, http.StatusOK, out)
}
