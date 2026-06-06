package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/even-app/even-app/libs/http/middleware"
	"github.com/even-app/even-app/libs/media"
	libs3 "github.com/even-app/even-app/libs/s3"
	"github.com/even-app/even-app/services/lexicon/internal/store"
	"github.com/google/uuid"
)

type PlatformMediaHandler struct {
	Store          *store.MediaStore
	S3             *libs3.Client
	Bucket         string
	UserQuotaBytes int64
}

func (h *PlatformMediaHandler) Register(mux *http.ServeMux, jwt func(http.Handler) http.Handler) {
	mux.Handle("POST /api/v1/platform/media/presign", jwt(http.HandlerFunc(h.presign)))
	mux.Handle("POST /api/v1/platform/media/confirm", jwt(http.HandlerFunc(h.confirm)))
	mux.Handle("GET /api/v1/platform/languages/{code}/media", jwt(http.HandlerFunc(h.list)))
	mux.Handle("GET /api/v1/platform/media/{id}", jwt(http.HandlerFunc(h.get)))
	mux.Handle("PATCH /api/v1/platform/media/{id}", jwt(http.HandlerFunc(h.patch)))
	mux.Handle("DELETE /api/v1/platform/media/{id}", jwt(http.HandlerFunc(h.delete)))
}

type presignBody struct {
	Filename   string `json:"filename"`
	MimeType   string `json:"mime_type"`
	SizeBytes  int64  `json:"size_bytes"`
	LanguageID string `json:"language_id"`
}

func (h *PlatformMediaHandler) presign(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	var body presignBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if body.SizeBytes <= 0 {
		writeErr(w, http.StatusBadRequest, "size_bytes required")
		return
	}
	langID, err := uuid.Parse(body.LanguageID)
	if err != nil {
		langID, err = h.Store.LanguageIDByCode(r.Context(), "evn")
		if err != nil {
			writeErr(w, http.StatusBadRequest, "language_id required")
			return
		}
	}
	claims, _ := middleware.ClaimsFromContext(r.Context())
	if err := h.checkQuota(r, claims.UserID, body.SizeBytes, claims.IsAdmin); err != nil {
		writeErr(w, http.StatusForbidden, err.Error())
		return
	}
	assetID := uuid.New()
	ext := path.Ext(body.Filename)
	if ext == "" {
		ext = ".bin"
	}
	objectKey := "media/" + assetID.String() + ext
	if err := h.Store.InsertPending(r.Context(), assetID, langID, claims.UserID, objectKey, h.Bucket, body.SizeBytes); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	uploadURL, err := h.S3.PresignPut(r.Context(), objectKey, body.MimeType)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"upload_url": uploadURL, "object_key": objectKey, "media_asset_id": assetID.String(),
	})
}

type confirmBody struct {
	ObjectKey      string  `json:"object_key"`
	MimeType       string  `json:"mime_type"`
	SizeBytes      int64   `json:"size_bytes"`
	LanguageID     string  `json:"language_id"`
	DisplayName    string  `json:"display_name"`
	LinkedLexemeID *string `json:"linked_lexeme_id"`
	TTLSeconds     *int64  `json:"ttl_seconds"`
	ExpiresAt      *string `json:"expires_at"`
	Width          *int    `json:"width"`
	Height         *int    `json:"height"`
	DurationMs     *int    `json:"duration_ms"`
}

func (h *PlatformMediaHandler) confirm(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	var body confirmBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := media.ValidateDisplayName(body.DisplayName); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if body.SizeBytes <= 0 {
		writeErr(w, http.StatusBadRequest, "size_bytes required")
		return
	}
	expiresAt, err := media.ResolveExpires(body.TTLSeconds, body.ExpiresAt)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	kind, err := media.KindFromMIME(body.MimeType)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	langID, err := uuid.Parse(body.LanguageID)
	if err != nil {
		langID, _ = h.Store.LanguageIDByCode(r.Context(), "evn")
	}
	var linked *uuid.UUID
	if body.LinkedLexemeID != nil && *body.LinkedLexemeID != "" {
		id, err := uuid.Parse(*body.LinkedLexemeID)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "invalid linked_lexeme_id")
			return
		}
		linked = &id
	}
	claims, _ := middleware.ClaimsFromContext(r.Context())
	pendingSize, _ := h.Store.SizeBytesByObjectKey(r.Context(), body.ObjectKey)
	delta := body.SizeBytes - pendingSize
	if delta > 0 {
		if err := h.checkQuota(r, claims.UserID, delta, claims.IsAdmin); err != nil {
			writeErr(w, http.StatusForbidden, err.Error())
			return
		}
	}
	a, err := h.Store.Confirm(r.Context(), store.ConfirmInput{
		ObjectKey: body.ObjectKey, LanguageID: langID, MimeType: body.MimeType, MediaKind: kind,
		SizeBytes: body.SizeBytes, Width: body.Width, Height: body.Height, DurationMs: body.DurationMs,
		DisplayName: strings.TrimSpace(body.DisplayName), LinkedLexemeID: linked,
		UploadedBy: claims.UserID, Bucket: h.Bucket, ExpiresAt: expiresAt,
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	dto, err := h.toDTO(r, a)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, dto)
}

func (h *PlatformMediaHandler) checkQuota(r *http.Request, userID uuid.UUID, addBytes int64, isAdmin bool) error {
	if isAdmin || h.UserQuotaBytes <= 0 {
		return nil
	}
	used, err := h.Store.SumActiveSizeByUser(r.Context(), userID)
	if err != nil {
		return fmt.Errorf("quota check failed")
	}
	if used+addBytes > h.UserQuotaBytes {
		return fmt.Errorf("media storage quota exceeded")
	}
	return nil
}

func (h *PlatformMediaHandler) list(w http.ResponseWriter, r *http.Request) {
	// Read-only: any authenticated user (teacher picker, admin UI).
	code := r.PathValue("code")
	langID, err := h.Store.LanguageIDByCode(r.Context(), code)
	if err != nil {
		writeErr(w, http.StatusNotFound, "language not found")
		return
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, total, err := h.Store.List(r.Context(), store.ListFilter{
		LanguageID: langID, Query: r.URL.Query().Get("q"), Kind: r.URL.Query().Get("kind"),
		Page: page, Limit: limit,
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	dtos := make([]map[string]any, 0, len(items))
	for i := range items {
		d, err := h.toDTO(r, &items[i])
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		dtos = append(dtos, d)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items": dtos, "total": total, "page": max1(page), "limit": maxLimit(limit),
	})
}

func (h *PlatformMediaHandler) get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	a, err := h.Store.GetByID(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	if a.IsExpired() {
		writeErr(w, http.StatusGone, "media expired")
		return
	}
	dto, err := h.toDTO(r, a)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, dto)
}

func (h *PlatformMediaHandler) patch(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body struct {
		DisplayName    *string `json:"display_name"`
		LinkedLexemeID *string `json:"linked_lexeme_id"`
		TTLSeconds     *int64  `json:"ttl_seconds"`
		ExpiresAt      *string `json:"expires_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	var linked *uuid.UUID
	clearLinked := body.LinkedLexemeID != nil && *body.LinkedLexemeID == ""
	if !clearLinked && body.LinkedLexemeID != nil && *body.LinkedLexemeID != "" {
		lid, err := uuid.Parse(*body.LinkedLexemeID)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "invalid linked_lexeme_id")
			return
		}
		linked = &lid
	}
	var expiresAt *time.Time
	clearExpires := body.TTLSeconds != nil && *body.TTLSeconds == 0 &&
		(body.ExpiresAt == nil || strings.TrimSpace(*body.ExpiresAt) == "")
	if !clearExpires {
		var err error
		expiresAt, err = media.ResolveExpires(body.TTLSeconds, body.ExpiresAt)
		if err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	if body.DisplayName != nil {
		if err := media.ValidateDisplayName(*body.DisplayName); err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	a, err := h.Store.Patch(r.Context(), id, body.DisplayName, linked, clearLinked, expiresAt, clearExpires)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	dto, err := h.toDTO(r, a)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, dto)
}

func (h *PlatformMediaHandler) delete(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.Store.Delete(r.Context(), id); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *PlatformMediaHandler) toDTO(r *http.Request, a *store.MediaAsset) (map[string]any, error) {
	url, err := h.S3.PresignGet(r.Context(), a.ObjectKey)
	if err != nil {
		return nil, err
	}
	m := map[string]any{
		"id": a.ID.String(), "scope": a.Scope, "language_id": a.LanguageID.String(),
		"display_name": a.DisplayName,
		"mime_type": a.MimeType, "media_kind": a.MediaKind,
		"url": url, "size_bytes": a.SizeBytes,
		"created_at": a.CreatedAt.UTC().Format(time.RFC3339),
	}
	if a.LinkedLexemeID != nil {
		m["linked_lexeme_id"] = a.LinkedLexemeID.String()
	}
	if a.Width != nil {
		m["width"] = *a.Width
	}
	if a.Height != nil {
		m["height"] = *a.Height
	}
	if a.DurationMs != nil {
		m["duration_ms"] = *a.DurationMs
	}
	if a.ExpiresAt != nil {
		m["expires_at"] = a.ExpiresAt.UTC().Format(time.RFC3339)
	}
	return m, nil
}

func requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok || !claims.IsAdmin {
		writeErr(w, http.StatusForbidden, "platform admin required")
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg, "message": msg})
}

func max1(p int) int {
	if p < 1 {
		return 1
	}
	return p
}

func maxLimit(l int) int {
	if l < 1 {
		return 20
	}
	return l
}
