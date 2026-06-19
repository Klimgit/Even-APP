package repository

import (
	"context"
	"fmt"

	"github.com/even-app/even-app/services/learning/internal/gen/mediaquery"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MediaReader struct {
	queries *mediaquery.Queries
}

func NewMediaReader(pool *pgxpool.Pool) *MediaReader {
	if pool == nil {
		return nil
	}
	return &MediaReader{queries: mediaquery.New(pool)}
}

func (r *MediaReader) Available() bool {
	return r != nil && r.queries != nil
}

type MediaView struct {
	ID          uuid.UUID
	Scope       string
	DisplayName string
	MimeType    string
	MediaKind   string
}

func MediaRefURL(scope string, id uuid.UUID) string {
	if scope == "teacher" {
		return fmt.Sprintf("/api/v1/teacher/media/%s", id)
	}
	return fmt.Sprintf("/api/v1/platform/media/%s", id)
}

func (r *MediaReader) MediaByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]MediaView, error) {
	out := make(map[uuid.UUID]MediaView)
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := r.queries.ListMediaByIDs(ctx, mediaquery.ListMediaByIDsParams{Column1: ids})
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		out[row.ID] = MediaView{
			ID: row.ID, Scope: row.Scope, DisplayName: row.DisplayName,
			MimeType: row.MimeType, MediaKind: row.MediaKind,
		}
	}
	return out, nil
}
