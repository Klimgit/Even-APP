package service

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/even-app/even-app/libs/media"
	libs3 "github.com/even-app/even-app/libs/s3"
	"github.com/even-app/even-app/services/media/internal/domain"
	"github.com/even-app/even-app/services/media/internal/gen/query"
	"github.com/google/uuid"
)

type MediaService struct {
	q              *query.Queries
	s3             *libs3.Client
	bucket         string
	userQuotaBytes int64
}

func NewMediaService(q *query.Queries, s3 *libs3.Client, bucket string, userQuotaBytes int64) *MediaService {
	return &MediaService{q: q, s3: s3, bucket: bucket, userQuotaBytes: userQuotaBytes}
}

func (s *MediaService) Presign(ctx context.Context, in domain.PresignInput) (*domain.PresignResult, error) {
	if in.SizeBytes <= 0 {
		return nil, fmt.Errorf("size_bytes required")
	}
	langID, err := uuid.Parse(in.LanguageID)
	if err != nil {
		langID, err = s.q.GetLanguageIDByCode(ctx, query.GetLanguageIDByCodeParams{Code: "evn"})
		if err != nil {
			return nil, fmt.Errorf("language_id required")
		}
	}
	if err := s.CheckQuota(ctx, in.UserID, in.SizeBytes, in.IsAdmin); err != nil {
		return nil, err
	}
	assetID := uuid.New()
	ext := path.Ext(in.Filename)
	if ext == "" {
		ext = ".bin"
	}
	objectKey := "media/" + assetID.String() + ext
	if err := s.q.InsertPendingMedia(ctx, query.InsertPendingMediaParams{
		ID: assetID, LanguageID: langID, ObjectKey: objectKey, Bucket: s.bucket,
		SizeBytes: in.SizeBytes, UploadedBy: in.UserID,
	}); err != nil {
		return nil, err
	}
	uploadURL, err := s.s3.PresignPut(ctx, objectKey, in.MimeType)
	if err != nil {
		return nil, err
	}
	return &domain.PresignResult{
		UploadURL: uploadURL, ObjectKey: objectKey, MediaAssetID: assetID.String(),
	}, nil
}

func (s *MediaService) Confirm(ctx context.Context, in domain.MediaConfirmInput) (*domain.MediaAsset, error) {
	if err := media.ValidateDisplayName(in.DisplayName); err != nil {
		return nil, err
	}
	if in.SizeBytes <= 0 {
		return nil, fmt.Errorf("size_bytes required")
	}
	expiresAt, err := media.ResolveExpires(in.TTLSeconds, in.ExpiresAt)
	if err != nil {
		return nil, err
	}
	kind, err := media.KindFromMIME(in.MimeType)
	if err != nil {
		return nil, err
	}
	langID, err := uuid.Parse(in.LanguageID)
	if err != nil {
		langID, _ = s.q.GetLanguageIDByCode(ctx, query.GetLanguageIDByCodeParams{Code: "evn"})
	}
	var linked *uuid.UUID
	if in.LinkedLexemeID != nil && *in.LinkedLexemeID != "" {
		id, err := uuid.Parse(*in.LinkedLexemeID)
		if err != nil {
			return nil, fmt.Errorf("invalid linked_lexeme_id")
		}
		linked = &id
	}
	pendingSize, _ := s.q.GetMediaSizeByObjectKey(ctx, query.GetMediaSizeByObjectKeyParams{ObjectKey: in.ObjectKey})
	delta := in.SizeBytes - pendingSize
	if delta > 0 {
		if err := s.CheckQuota(ctx, in.UserID, delta, in.IsAdmin); err != nil {
			return nil, err
		}
	}
	row, err := s.q.ConfirmMedia(ctx, query.ConfirmMediaParams{
		ObjectKey: in.ObjectKey, MimeType: in.MimeType, MediaKind: kind, SizeBytes: in.SizeBytes,
		Width: intPtrToInt32(in.Width), Height: intPtrToInt32(in.Height), DurationMs: intPtrToInt32(in.DurationMs),
		DisplayName: strings.TrimSpace(in.DisplayName), LinkedLexemeID: linked,
		LanguageID: langID, Bucket: s.bucket, ExpiresAt: expiresAt,
	})
	if err != nil {
		return nil, err
	}
	a := mapConfirmRow(row)
	return &a, nil
}

func (s *MediaService) CheckQuota(ctx context.Context, userID uuid.UUID, addBytes int64, isAdmin bool) error {
	if isAdmin || s.userQuotaBytes <= 0 {
		return nil
	}
	used, err := s.q.SumActiveMediaSizeByUser(ctx, query.SumActiveMediaSizeByUserParams{UploadedBy: userID})
	if err != nil {
		return fmt.Errorf("quota check failed")
	}
	if used+addBytes > s.userQuotaBytes {
		return fmt.Errorf("media storage quota exceeded")
	}
	return nil
}

func (s *MediaService) LanguageIDByCode(ctx context.Context, code string) (uuid.UUID, error) {
	return s.q.GetLanguageIDByCode(ctx, query.GetLanguageIDByCodeParams{Code: code})
}

func (s *MediaService) List(ctx context.Context, f domain.ListFilter) ([]domain.MediaAsset, int, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}
	offset := (f.Page - 1) * f.Limit

	var kind, search *string
	if f.Kind != "" {
		kind = &f.Kind
	}
	if q := strings.TrimSpace(f.Query); q != "" {
		s := "%" + q + "%"
		search = &s
	}
	filter := query.CountPlatformMediaParams{LanguageID: f.LanguageID, MediaKind: kind, Search: search}
	total, err := s.q.CountPlatformMedia(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	rows, err := s.q.ListPlatformMedia(ctx, query.ListPlatformMediaParams{
		LanguageID: f.LanguageID, MediaKind: kind, Search: search,
		RowOffset: int32(offset), RowLimit: int32(f.Limit),
	})
	if err != nil {
		return nil, 0, err
	}
	items := make([]domain.MediaAsset, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapListRow(row))
	}
	return items, int(total), nil
}

func (s *MediaService) GetByID(ctx context.Context, id uuid.UUID) (*domain.MediaAsset, error) {
	row, err := s.q.GetMediaByID(ctx, query.GetMediaByIDParams{ID: id})
	if err != nil {
		return nil, err
	}
	a := mapGetRow(row)
	return &a, nil
}

func (s *MediaService) Patch(ctx context.Context, id uuid.UUID, in domain.MediaPatchInput) (*domain.MediaAsset, error) {
	cur, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if in.DisplayName != nil {
		if err := media.ValidateDisplayName(*in.DisplayName); err != nil {
			return nil, err
		}
		cur.DisplayName = strings.TrimSpace(*in.DisplayName)
	}
	clearLinked := in.LinkedLexemeID != nil && *in.LinkedLexemeID == ""
	if clearLinked {
		cur.LinkedLexemeID = nil
	} else if in.LinkedLexemeID != nil && *in.LinkedLexemeID != "" {
		lid, err := uuid.Parse(*in.LinkedLexemeID)
		if err != nil {
			return nil, fmt.Errorf("invalid linked_lexeme_id")
		}
		cur.LinkedLexemeID = &lid
	}
	clearExpires := in.TTLSeconds != nil && *in.TTLSeconds == 0 &&
		(in.ExpiresAt == nil || strings.TrimSpace(*in.ExpiresAt) == "")
	if clearExpires {
		cur.ExpiresAt = nil
	} else if in.TTLSeconds != nil || in.ExpiresAt != nil {
		expiresAt, err := media.ResolveExpires(in.TTLSeconds, in.ExpiresAt)
		if err != nil {
			return nil, err
		}
		cur.ExpiresAt = expiresAt
	}
	row, err := s.q.UpdateMedia(ctx, query.UpdateMediaParams{
		ID: id, DisplayName: cur.DisplayName, LinkedLexemeID: cur.LinkedLexemeID, ExpiresAt: cur.ExpiresAt,
	})
	if err != nil {
		return nil, err
	}
	a := mapUpdateRow(row)
	return &a, nil
}

func (s *MediaService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.q.DeleteMedia(ctx, query.DeleteMediaParams{ID: id})
}

func (s *MediaService) ToDTO(ctx context.Context, a *domain.MediaAsset) (*domain.MediaAssetDTO, error) {
	url, err := s.s3.PresignGet(ctx, a.ObjectKey)
	if err != nil {
		return nil, err
	}
	dto := &domain.MediaAssetDTO{
		ID: a.ID.String(), Scope: a.Scope, LanguageID: a.LanguageID.String(),
		DisplayName: a.DisplayName, MimeType: a.MimeType, MediaKind: a.MediaKind,
		URL: url, SizeBytes: a.SizeBytes,
		CreatedAt: a.CreatedAt.UTC().Format(time.RFC3339),
	}
	if a.LinkedLexemeID != nil {
		s := a.LinkedLexemeID.String()
		dto.LinkedLexemeID = &s
	}
	if a.Width != nil {
		dto.Width = a.Width
	}
	if a.Height != nil {
		dto.Height = a.Height
	}
	if a.DurationMs != nil {
		dto.DurationMs = a.DurationMs
	}
	if a.ExpiresAt != nil {
		s := a.ExpiresAt.UTC().Format(time.RFC3339)
		dto.ExpiresAt = &s
	}
	return dto, nil
}

func mapConfirmRow(row query.ConfirmMediaRow) domain.MediaAsset {
	return mapMediaFields(
		row.ID, row.Scope, row.LanguageID, row.OwnerID, row.ObjectKey, row.Bucket,
		row.MimeType, row.MediaKind, row.SizeBytes, row.Width, row.Height, row.DurationMs,
		row.DisplayName, row.LinkedLexemeID, row.UploadedBy, row.ExpiresAt, row.CreatedAt,
	)
}

func mapListRow(row query.ListPlatformMediaRow) domain.MediaAsset {
	return mapMediaFields(
		row.ID, row.Scope, row.LanguageID, row.OwnerID, row.ObjectKey, row.Bucket,
		row.MimeType, row.MediaKind, row.SizeBytes, row.Width, row.Height, row.DurationMs,
		row.DisplayName, row.LinkedLexemeID, row.UploadedBy, row.ExpiresAt, row.CreatedAt,
	)
}

func mapGetRow(row query.GetMediaByIDRow) domain.MediaAsset {
	return mapMediaFields(
		row.ID, row.Scope, row.LanguageID, row.OwnerID, row.ObjectKey, row.Bucket,
		row.MimeType, row.MediaKind, row.SizeBytes, row.Width, row.Height, row.DurationMs,
		row.DisplayName, row.LinkedLexemeID, row.UploadedBy, row.ExpiresAt, row.CreatedAt,
	)
}

func mapUpdateRow(row query.UpdateMediaRow) domain.MediaAsset {
	return mapMediaFields(
		row.ID, row.Scope, row.LanguageID, row.OwnerID, row.ObjectKey, row.Bucket,
		row.MimeType, row.MediaKind, row.SizeBytes, row.Width, row.Height, row.DurationMs,
		row.DisplayName, row.LinkedLexemeID, row.UploadedBy, row.ExpiresAt, row.CreatedAt,
	)
}

func mapMediaFields(
	id uuid.UUID, scope string, languageID uuid.UUID, ownerID *uuid.UUID,
	objectKey, bucket, mimeType, mediaKind string, sizeBytes int64,
	width, height, durationMs *int32, displayName string, linkedLexemeID *uuid.UUID,
	uploadedBy uuid.UUID, expiresAt *time.Time, createdAt time.Time,
) domain.MediaAsset {
	return domain.MediaAsset{
		ID: id, Scope: scope, LanguageID: languageID, OwnerID: ownerID,
		ObjectKey: objectKey, Bucket: bucket, MimeType: mimeType, MediaKind: mediaKind,
		SizeBytes: sizeBytes, Width: int32PtrToInt(width), Height: int32PtrToInt(height),
		DurationMs: int32PtrToInt(durationMs), DisplayName: displayName,
		LinkedLexemeID: linkedLexemeID, UploadedBy: uploadedBy,
		ExpiresAt: expiresAt, CreatedAt: createdAt,
	}
}

func int32PtrToInt(p *int32) *int {
	if p == nil {
		return nil
	}
	v := int(*p)
	return &v
}

func intPtrToInt32(p *int) *int32 {
	if p == nil {
		return nil
	}
	v := int32(*p)
	return &v
}
