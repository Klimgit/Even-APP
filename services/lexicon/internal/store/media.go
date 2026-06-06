package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
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

type MediaStore struct {
	pool *pgxpool.Pool
}

func NewMediaStore(pool *pgxpool.Pool) *MediaStore {
	return &MediaStore{pool: pool}
}

func (s *MediaStore) LanguageIDByCode(ctx context.Context, code string) (uuid.UUID, error) {
	var id uuid.UUID
	err := s.pool.QueryRow(ctx, `SELECT id FROM languages WHERE code = $1`, code).Scan(&id)
	return id, err
}

func (s *MediaStore) SizeBytesByObjectKey(ctx context.Context, objectKey string) (int64, error) {
	var size int64
	err := s.pool.QueryRow(ctx, `
		SELECT size_bytes FROM media_assets WHERE object_key = $1 AND scope = 'platform'
	`, objectKey).Scan(&size)
	return size, err
}

func (s *MediaStore) InsertPending(ctx context.Context, id, languageID, uploadedBy uuid.UUID, objectKey, bucket string, sizeBytes int64) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO media_assets (
			id, scope, language_id, object_key, bucket, mime_type, media_kind,
			size_bytes, display_name, uploaded_by
		) VALUES ($1, 'platform', $2, $3, $4, 'application/octet-stream', 'image', $5, '(uploading)', $6)
	`, id, languageID, objectKey, bucket, sizeBytes, uploadedBy)
	return err
}

// SumActiveSizeByUser totals stored media for quota checks (includes pending uploads).
func (s *MediaStore) SumActiveSizeByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	var sum int64
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(size_bytes), 0)
		FROM media_assets
		WHERE uploaded_by = $1
		  AND (expires_at IS NULL OR expires_at > now())
	`, userID).Scan(&sum)
	return sum, err
}

func (a *MediaAsset) IsExpired() bool {
	return a.ExpiresAt != nil && !a.ExpiresAt.After(time.Now().UTC())
}

type ConfirmInput struct {
	ObjectKey      string
	LanguageID     uuid.UUID
	MimeType       string
	MediaKind      string
	SizeBytes      int64
	Width          *int
	Height         *int
	DurationMs     *int
	DisplayName    string
	LinkedLexemeID *uuid.UUID
	UploadedBy     uuid.UUID
	Bucket         string
	ExpiresAt      *time.Time
}

func (s *MediaStore) Confirm(ctx context.Context, in ConfirmInput) (*MediaAsset, error) {
	var a MediaAsset
	err := s.pool.QueryRow(ctx, `
		UPDATE media_assets SET
			mime_type = $2, media_kind = $3, size_bytes = $4,
			width = $5, height = $6, duration_ms = $7,
			display_name = $8, linked_lexeme_id = $9,
			language_id = $10, bucket = $11, expires_at = $12, updated_at = now()
		WHERE object_key = $1 AND scope = 'platform'
		RETURNING id, scope, language_id, owner_id, object_key, bucket, mime_type, media_kind,
			size_bytes, width, height, duration_ms, display_name, linked_lexeme_id,
			uploaded_by, expires_at, created_at
	`, in.ObjectKey, in.MimeType, in.MediaKind, in.SizeBytes, in.Width, in.Height, in.DurationMs,
		in.DisplayName, in.LinkedLexemeID, in.LanguageID, in.Bucket, in.ExpiresAt,
	).Scan(
		&a.ID, &a.Scope, &a.LanguageID, &a.OwnerID, &a.ObjectKey, &a.Bucket, &a.MimeType, &a.MediaKind,
		&a.SizeBytes, &a.Width, &a.Height, &a.DurationMs, &a.DisplayName, &a.LinkedLexemeID,
		&a.UploadedBy, &a.ExpiresAt, &a.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

type ListFilter struct {
	LanguageID uuid.UUID
	Query      string
	Kind       string
	Page       int
	Limit      int
}

func (s *MediaStore) List(ctx context.Context, f ListFilter) ([]MediaAsset, int, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}
	offset := (f.Page - 1) * f.Limit
	where := []string{
		"scope = 'platform'",
		"language_id = $1",
		"mime_type != 'application/octet-stream'",
		"(expires_at IS NULL OR expires_at > now())",
	}
	args := []any{f.LanguageID}
	n := 2
	if f.Kind != "" {
		where = append(where, fmt.Sprintf("media_kind = $%d", n))
		args = append(args, f.Kind)
		n++
	}
	if q := strings.TrimSpace(f.Query); q != "" {
		where = append(where, fmt.Sprintf("display_name ILIKE $%d", n))
		args = append(args, "%"+q+"%")
		n++
	}
	whereSQL := strings.Join(where, " AND ")
	var total int
	if err := s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM media_assets WHERE "+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, f.Limit, offset)
	rows, err := s.pool.Query(ctx, `
		SELECT id, scope, language_id, owner_id, object_key, bucket, mime_type, media_kind,
			size_bytes, width, height, duration_ms, display_name, linked_lexeme_id,
			uploaded_by, expires_at, created_at
		FROM media_assets WHERE `+whereSQL+`
		ORDER BY created_at DESC LIMIT $`+fmt.Sprint(n)+` OFFSET $`+fmt.Sprint(n+1), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []MediaAsset
	for rows.Next() {
		var a MediaAsset
		if err := rows.Scan(
			&a.ID, &a.Scope, &a.LanguageID, &a.OwnerID, &a.ObjectKey, &a.Bucket, &a.MimeType, &a.MediaKind,
			&a.SizeBytes, &a.Width, &a.Height, &a.DurationMs, &a.DisplayName, &a.LinkedLexemeID,
			&a.UploadedBy, &a.ExpiresAt, &a.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, a)
	}
	return items, total, rows.Err()
}

func (s *MediaStore) GetByID(ctx context.Context, id uuid.UUID) (*MediaAsset, error) {
	var a MediaAsset
	err := s.pool.QueryRow(ctx, `
		SELECT id, scope, language_id, owner_id, object_key, bucket, mime_type, media_kind,
			size_bytes, width, height, duration_ms, display_name, linked_lexeme_id,
			uploaded_by, expires_at, created_at
		FROM media_assets WHERE id = $1 AND scope = 'platform'
	`, id).Scan(
		&a.ID, &a.Scope, &a.LanguageID, &a.OwnerID, &a.ObjectKey, &a.Bucket, &a.MimeType, &a.MediaKind,
		&a.SizeBytes, &a.Width, &a.Height, &a.DurationMs, &a.DisplayName, &a.LinkedLexemeID,
		&a.UploadedBy, &a.ExpiresAt, &a.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *MediaStore) Patch(ctx context.Context, id uuid.UUID, displayName *string, linked *uuid.UUID, clearLinked bool, expiresAt *time.Time, clearExpires bool) (*MediaAsset, error) {
	a, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if displayName != nil {
		if err := validateDisplayName(*displayName); err != nil {
			return nil, err
		}
		a.DisplayName = strings.TrimSpace(*displayName)
	}
	if clearLinked {
		a.LinkedLexemeID = nil
	} else if linked != nil {
		a.LinkedLexemeID = linked
	}
	if clearExpires {
		a.ExpiresAt = nil
	} else if expiresAt != nil {
		a.ExpiresAt = expiresAt
	}
	err = s.pool.QueryRow(ctx, `
		UPDATE media_assets SET display_name = $2, linked_lexeme_id = $3, expires_at = $4, updated_at = now()
		WHERE id = $1
		RETURNING id, scope, language_id, owner_id, object_key, bucket, mime_type, media_kind,
			size_bytes, width, height, duration_ms, display_name, linked_lexeme_id,
			uploaded_by, expires_at, created_at
	`, id, a.DisplayName, a.LinkedLexemeID, a.ExpiresAt).Scan(
		&a.ID, &a.Scope, &a.LanguageID, &a.OwnerID, &a.ObjectKey, &a.Bucket, &a.MimeType, &a.MediaKind,
		&a.SizeBytes, &a.Width, &a.Height, &a.DurationMs, &a.DisplayName, &a.LinkedLexemeID,
		&a.UploadedBy, &a.ExpiresAt, &a.CreatedAt,
	)
	return a, err
}

func (s *MediaStore) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM media_assets WHERE id = $1 AND scope = 'platform'`, id)
	return err
}

func validateDisplayName(name string) error {
	n := strings.TrimSpace(name)
	if n == "" || len(n) > 120 {
		return fmt.Errorf("display_name must be 1–120 characters")
	}
	return nil
}
