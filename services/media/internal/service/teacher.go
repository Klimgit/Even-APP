package service

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/even-app/even-app/libs/media"
	"github.com/even-app/even-app/services/media/internal/domain"
	"github.com/even-app/even-app/services/media/internal/gen/query"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type TeacherListFilter struct {
	OwnerID      uuid.UUID
	LanguageCode string
	Query        string
	Kind         string
	Page         int
	Limit        int
}

func (s *MediaService) TeacherPresign(ctx context.Context, in domain.PresignInput) (*domain.PresignResult, error) {
	if in.SizeBytes <= 0 {
		return nil, fmt.Errorf("size_bytes required")
	}
	langID, err := s.resolveLanguageID(ctx, in.LanguageID)
	if err != nil {
		return nil, err
	}
	if err := s.CheckQuota(ctx, in.UserID, in.SizeBytes, false); err != nil {
		return nil, err
	}
	assetID := uuid.New()
	ext := path.Ext(in.Filename)
	if ext == "" {
		ext = ".bin"
	}
	objectKey := "media/" + assetID.String() + ext
	if err := s.q.InsertPendingTeacherMedia(ctx, query.InsertPendingTeacherMediaParams{
		ID: assetID, LanguageID: langID, OwnerID: uuidPtr(in.UserID), ObjectKey: objectKey,
		Bucket: s.bucket, SizeBytes: in.SizeBytes, UploadedBy: in.UserID,
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

func (s *MediaService) TeacherConfirm(ctx context.Context, in domain.MediaConfirmInput) (*domain.MediaAsset, error) {
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
	langID, err := s.resolveLanguageID(ctx, in.LanguageID)
	if err != nil {
		return nil, err
	}
	var linked *uuid.UUID
	if in.LinkedLexemeID != nil && *in.LinkedLexemeID != "" {
		id, err := uuid.Parse(*in.LinkedLexemeID)
		if err != nil {
			return nil, fmt.Errorf("invalid linked_lexeme_id")
		}
		linked = &id
	}
	pendingSize, _ := s.q.GetMediaSizeByObjectKeyAny(ctx, query.GetMediaSizeByObjectKeyAnyParams{ObjectKey: in.ObjectKey})
	delta := in.SizeBytes - pendingSize
	if delta > 0 {
		if err := s.CheckQuota(ctx, in.UserID, delta, false); err != nil {
			return nil, err
		}
	}
	row, err := s.q.ConfirmTeacherMedia(ctx, query.ConfirmTeacherMediaParams{
		ObjectKey: in.ObjectKey, MimeType: in.MimeType, MediaKind: kind, SizeBytes: in.SizeBytes,
		Width: intPtrToInt32(in.Width), Height: intPtrToInt32(in.Height), DurationMs: intPtrToInt32(in.DurationMs),
		DisplayName: strings.TrimSpace(in.DisplayName), LinkedLexemeID: linked,
		LanguageID: langID, Bucket: s.bucket, ExpiresAt: expiresAt, OwnerID: uuidPtr(in.UserID),
	})
	if err != nil {
		return nil, err
	}
	a := mapTeacherConfirmRow(row)
	return &a, nil
}

func (s *MediaService) TeacherList(ctx context.Context, f TeacherListFilter) ([]domain.MediaAsset, int, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}
	offset := (f.Page - 1) * f.Limit
	var langID *uuid.UUID
	if f.LanguageCode != "" {
		id, err := s.q.GetLanguageIDByCode(ctx, query.GetLanguageIDByCodeParams{Code: f.LanguageCode})
		if err != nil {
			return nil, 0, err
		}
		langID = &id
	}
	var kind, search *string
	if f.Kind != "" {
		kind = &f.Kind
	}
	if q := strings.TrimSpace(f.Query); q != "" {
		s := "%" + q + "%"
		search = &s
	}
	total, err := s.q.CountTeacherMedia(ctx, query.CountTeacherMediaParams{
		OwnerID: uuidPtr(f.OwnerID), LanguageID: langID, MediaKind: kind, Search: search,
	})
	if err != nil {
		return nil, 0, err
	}
	rows, err := s.q.ListTeacherMedia(ctx, query.ListTeacherMediaParams{
		OwnerID: uuidPtr(f.OwnerID), LanguageID: langID, MediaKind: kind, Search: search,
		RowOffset: int32(offset), RowLimit: int32(f.Limit),
	})
	if err != nil {
		return nil, 0, err
	}
	items := make([]domain.MediaAsset, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapTeacherListRow(row))
	}
	return items, int(total), nil
}

func (s *MediaService) TeacherGetByID(ctx context.Context, id, ownerID uuid.UUID) (*domain.MediaAsset, error) {
	row, err := s.q.GetTeacherMediaByID(ctx, query.GetTeacherMediaByIDParams{ID: id, OwnerID: uuidPtr(ownerID)})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	a := mapTeacherGetRow(row)
	return &a, nil
}

func (s *MediaService) TeacherPatch(ctx context.Context, id, ownerID uuid.UUID, in domain.MediaPatchInput) (*domain.MediaAsset, error) {
	cur, err := s.TeacherGetByID(ctx, id, ownerID)
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
	row, err := s.q.UpdateTeacherMedia(ctx, query.UpdateTeacherMediaParams{
		ID: id, OwnerID: uuidPtr(ownerID), DisplayName: cur.DisplayName,
		LinkedLexemeID: cur.LinkedLexemeID, ExpiresAt: cur.ExpiresAt,
	})
	if err != nil {
		return nil, err
	}
	a := mapTeacherUpdateRow(row)
	return &a, nil
}

func (s *MediaService) TeacherDelete(ctx context.Context, id, ownerID uuid.UUID) error {
	n, err := s.q.DeleteTeacherMedia(ctx, query.DeleteTeacherMediaParams{ID: id, OwnerID: uuidPtr(ownerID)})
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func uuidPtr(id uuid.UUID) *uuid.UUID {
	return &id
}

func mapTeacherConfirmRow(row query.ConfirmTeacherMediaRow) domain.MediaAsset {
	return mapMediaFields(
		row.ID, row.Scope, row.LanguageID, row.OwnerID, row.ObjectKey, row.Bucket,
		row.MimeType, row.MediaKind, row.SizeBytes, row.Width, row.Height, row.DurationMs,
		row.DisplayName, row.LinkedLexemeID, row.UploadedBy, row.ExpiresAt, row.CreatedAt,
	)
}

func mapTeacherListRow(row query.ListTeacherMediaRow) domain.MediaAsset {
	return mapMediaFields(
		row.ID, row.Scope, row.LanguageID, row.OwnerID, row.ObjectKey, row.Bucket,
		row.MimeType, row.MediaKind, row.SizeBytes, row.Width, row.Height, row.DurationMs,
		row.DisplayName, row.LinkedLexemeID, row.UploadedBy, row.ExpiresAt, row.CreatedAt,
	)
}

func mapTeacherGetRow(row query.GetTeacherMediaByIDRow) domain.MediaAsset {
	return mapMediaFields(
		row.ID, row.Scope, row.LanguageID, row.OwnerID, row.ObjectKey, row.Bucket,
		row.MimeType, row.MediaKind, row.SizeBytes, row.Width, row.Height, row.DurationMs,
		row.DisplayName, row.LinkedLexemeID, row.UploadedBy, row.ExpiresAt, row.CreatedAt,
	)
}

func mapTeacherUpdateRow(row query.UpdateTeacherMediaRow) domain.MediaAsset {
	return mapMediaFields(
		row.ID, row.Scope, row.LanguageID, row.OwnerID, row.ObjectKey, row.Bucket,
		row.MimeType, row.MediaKind, row.SizeBytes, row.Width, row.Height, row.DurationMs,
		row.DisplayName, row.LinkedLexemeID, row.UploadedBy, row.ExpiresAt, row.CreatedAt,
	)
}
