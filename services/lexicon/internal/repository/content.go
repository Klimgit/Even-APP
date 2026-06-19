package repository

import (
	"context"

	"github.com/even-app/even-app/services/lexicon/internal/gen/contentquery"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ContentReader struct {
	pool    *pgxpool.Pool
	queries *contentquery.Queries
}

func NewContentReader(pool *pgxpool.Pool) *ContentReader {
	if pool == nil {
		return nil
	}
	return &ContentReader{pool: pool, queries: contentquery.New(pool)}
}

func (r *ContentReader) Available() bool {
	return r != nil && r.pool != nil
}

func (r *ContentReader) ListBlocksByCourseID(ctx context.Context, courseID uuid.UUID) ([]contentquery.ListBlocksByCourseIDRow, error) {
	return r.queries.ListBlocksByCourseID(ctx, contentquery.ListBlocksByCourseIDParams{CourseID: courseID})
}
