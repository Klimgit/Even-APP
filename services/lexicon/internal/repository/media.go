package repository

import (
	"context"
	"errors"

	"github.com/even-app/even-app/services/lexicon/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MediaReader struct {
	pool *pgxpool.Pool
}

func NewMediaReader(pool *pgxpool.Pool) *MediaReader {
	if pool == nil {
		return nil
	}
	return &MediaReader{pool: pool}
}

func (r *MediaReader) Available() bool {
	return r != nil && r.pool != nil
}

func (r *MediaReader) AssertTeacherMediaOwner(ctx context.Context, assetID, ownerID uuid.UUID) error {
	var uploadedBy uuid.UUID
	err := r.pool.QueryRow(ctx,
		`SELECT uploaded_by FROM media_assets WHERE id = $1 AND scope = 'teacher'`,
		assetID,
	).Scan(&uploadedBy)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	}
	if uploadedBy != ownerID {
		return domain.ErrForbidden
	}
	return nil
}
