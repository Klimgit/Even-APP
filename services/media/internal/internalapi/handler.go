package internalapi

import (
	"encoding/json"
	"net/http"

	"github.com/even-app/even-app/libs/clients/dto"
	mediasvc "github.com/even-app/even-app/services/media/internal/service"
)

type Handler struct {
	svc *mediasvc.MediaService
}

func New(svc *mediasvc.MediaService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Mount(mux *http.ServeMux) {
	mux.HandleFunc("POST /media/by-ids", h.mediaByIDs)
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

func (h *Handler) mediaByIDs(w http.ResponseWriter, r *http.Request) {
	var req dto.IDsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	views, err := h.svc.MediaViewsByIDs(r.Context(), req.IDs)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, views)
}
