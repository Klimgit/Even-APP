package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PlatformStatsReader struct {
	content  *pgxpool.Pool
	learning *pgxpool.Pool
}

func NewPlatformStatsReader(content, learning *pgxpool.Pool) *PlatformStatsReader {
	return &PlatformStatsReader{content: content, learning: learning}
}

func (r *PlatformStatsReader) PublishedCourses(ctx context.Context) (int, error) {
	if r.content == nil {
		return 0, nil
	}
	var n int
	err := r.content.QueryRow(ctx, `SELECT COUNT(*)::int FROM courses WHERE is_published = true`).Scan(&n)
	return n, err
}

func (r *PlatformStatsReader) ActiveEnrollments(ctx context.Context) (int, error) {
	if r.learning == nil {
		return 0, nil
	}
	var n int
	err := r.learning.QueryRow(ctx, `SELECT COUNT(*)::int FROM course_enrollments WHERE status = 'active'`).Scan(&n)
	return n, err
}
