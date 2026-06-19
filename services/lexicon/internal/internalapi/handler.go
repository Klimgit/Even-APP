package internalapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/even-app/even-app/libs/clients/dto"
	"github.com/even-app/even-app/services/lexicon/internal/gen/query"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Handler struct {
	q *query.Queries
}

func New(q *query.Queries) *Handler {
	return &Handler{q: q}
}

func (h *Handler) Mount(mux *http.ServeMux) {
	mux.HandleFunc("GET /languages/{languageId}", h.getLanguage)
	mux.HandleFunc("POST /languages/by-ids", h.languagesByIDs)
	mux.HandleFunc("POST /lexemes/by-ids", h.lexemesByIDs)
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

func (h *Handler) getLanguage(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("languageId"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	row, err := h.q.GetLanguageByID(r.Context(), query.GetLanguageByIDParams{ID: id})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeErr(w, http.StatusNotFound, "not found")
			return
		}
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, dto.LanguageView{
		ID: row.ID, Code: row.Code, Name: row.Name, NativeName: row.NativeName,
		Direction: row.Direction, IsActive: row.IsActive,
	})
}

func (h *Handler) languagesByIDs(w http.ResponseWriter, r *http.Request) {
	var req dto.IDsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	rows, err := h.q.ListLanguagesByIDs(r.Context(), query.ListLanguagesByIDsParams{Column1: req.IDs})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]dto.LanguageView, 0, len(rows))
	for _, row := range rows {
		out = append(out, dto.LanguageView{
			ID: row.ID, Code: row.Code, Name: row.Name, NativeName: row.NativeName,
			Direction: row.Direction, IsActive: row.IsActive,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) lexemesByIDs(w http.ResponseWriter, r *http.Request) {
	var req dto.IDsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	rows, err := h.q.ListLexemesByIDs(r.Context(), query.ListLexemesByIDsParams{Column1: req.IDs})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]dto.LexemeView, 0, len(rows))
	for _, row := range rows {
		out = append(out, dto.LexemeView{ID: row.ID, Lemma: row.Lemma, PartOfSpeech: row.PartOfSpeech})
	}
	writeJSON(w, http.StatusOK, out)
}
