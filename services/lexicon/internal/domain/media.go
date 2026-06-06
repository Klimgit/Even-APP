package domain

import (
	"time"

	"github.com/google/uuid"
)

type MediaAsset struct {
	ID             uuid.UUID
	Scope          string
	LanguageID     uuid.UUID
	OwnerID        *uuid.UUID
	ObjectKey      string
	Bucket         string
	MimeType       string
	MediaKind      string
	SizeBytes      int64
	Width          *int
	Height         *int
	DurationMs     *int
	DisplayName    string
	LinkedLexemeID *uuid.UUID
	UploadedBy     uuid.UUID
	ExpiresAt      *time.Time
	CreatedAt      time.Time
}

func (a *MediaAsset) IsExpired() bool {
	return a.ExpiresAt != nil && !a.ExpiresAt.After(time.Now().UTC())
}

type ListFilter struct {
	LanguageID uuid.UUID
	Query      string
	Kind       string
	Page       int
	Limit      int
}

type PresignInput struct {
	Filename   string
	MimeType   string
	SizeBytes  int64
	LanguageID string
	UserID     uuid.UUID
	IsAdmin    bool
}

type PresignResult struct {
	UploadURL    string
	ObjectKey    string
	MediaAssetID string
}

type PresignResponse struct {
	UploadURL    string `json:"upload_url"`
	ObjectKey    string `json:"object_key"`
	MediaAssetID string `json:"media_asset_id"`
}

// MediaConfirmInput — вход service-слоя (строковые поля из HTTP до нормализации).
type MediaConfirmInput struct {
	ObjectKey      string
	MimeType       string
	SizeBytes      int64
	LanguageID     string
	DisplayName    string
	LinkedLexemeID *string
	TTLSeconds     *int64
	ExpiresAt      *string
	Width          *int
	Height         *int
	DurationMs     *int
	UserID         uuid.UUID
	IsAdmin        bool
}

type MediaPatchInput struct {
	DisplayName    *string
	LinkedLexemeID *string
	TTLSeconds     *int64
	ExpiresAt      *string
}

type PresignRequest struct {
	Filename   string `json:"filename"`
	MimeType   string `json:"mime_type"`
	SizeBytes  int64  `json:"size_bytes"`
	LanguageID string `json:"language_id"`
}

type ConfirmRequest struct {
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

type PatchMediaRequest struct {
	DisplayName    *string `json:"display_name"`
	LinkedLexemeID *string `json:"linked_lexeme_id"`
	TTLSeconds     *int64  `json:"ttl_seconds"`
	ExpiresAt      *string `json:"expires_at"`
}

type MediaAssetDTO struct {
	ID             string  `json:"id"`
	Scope          string  `json:"scope"`
	LanguageID     string  `json:"language_id"`
	DisplayName    string  `json:"display_name"`
	MimeType       string  `json:"mime_type"`
	MediaKind      string  `json:"media_kind"`
	URL            string  `json:"url"`
	SizeBytes      int64   `json:"size_bytes"`
	CreatedAt      string  `json:"created_at"`
	LinkedLexemeID *string `json:"linked_lexeme_id,omitempty"`
	Width          *int    `json:"width,omitempty"`
	Height         *int    `json:"height,omitempty"`
	DurationMs     *int    `json:"duration_ms,omitempty"`
	ExpiresAt      *string `json:"expires_at,omitempty"`
}
