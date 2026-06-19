package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PlatformStatsReader struct {
	content  *pgxpool.Pool
	learning *pgxpool.Pool
	lexicon  *pgxpool.Pool
}

func NewPlatformStatsReader(content, learning, lexicon *pgxpool.Pool) *PlatformStatsReader {
	return &PlatformStatsReader{content: content, learning: learning, lexicon: lexicon}
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

func (r *PlatformStatsReader) TotalCourses(ctx context.Context) (int, error) {
	if r.content == nil {
		return 0, nil
	}
	var n int
	err := r.content.QueryRow(ctx, `SELECT COUNT(*)::int FROM courses`).Scan(&n)
	return n, err
}

func (r *PlatformStatsReader) PlatformLexemes(ctx context.Context) (int, error) {
	if r.lexicon == nil {
		return 0, nil
	}
	var n int
	err := r.lexicon.QueryRow(ctx, `SELECT COUNT(*)::int FROM lexemes WHERE scope = 'platform'`).Scan(&n)
	return n, err
}
