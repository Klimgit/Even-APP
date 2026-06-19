package repository

import (
	"context"

	"github.com/even-app/even-app/services/content/internal/gen/learningquery"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LearningReader struct {
	pool    *pgxpool.Pool
	queries *learningquery.Queries
}

func NewLearningReader(pool *pgxpool.Pool) *LearningReader {
	if pool == nil {
		return nil
	}
	return &LearningReader{pool: pool, queries: learningquery.New(pool)}
}

func (r *LearningReader) Available() bool {
	return r != nil && r.pool != nil
}

func (r *LearningReader) ListEnrollmentsByCourse(ctx context.Context, courseID uuid.UUID) ([]learningquery.ListEnrollmentsByCourseRow, error) {
	return r.queries.ListEnrollmentsByCourse(ctx, learningquery.ListEnrollmentsByCourseParams{CourseID: courseID})
}

func (r *LearningReader) GetStudentLessonProgress(ctx context.Context, userID, courseID uuid.UUID) ([]learningquery.GetStudentLessonProgressRow, error) {
	return r.queries.GetStudentLessonProgress(ctx, learningquery.GetStudentLessonProgressParams{
		UserID: userID, CourseID: courseID,
	})
}

func (r *LearningReader) CreateEnrollment(ctx context.Context, userID, courseID uuid.UUID, enrolledBy *uuid.UUID) (learningquery.CourseEnrollment, error) {
	return r.queries.CreateEnrollment(ctx, learningquery.CreateEnrollmentParams{
		UserID: userID, CourseID: courseID, EnrolledBy: enrolledBy,
	})
}

func (r *LearningReader) GetEnrollmentByUserAndCourse(ctx context.Context, userID, courseID uuid.UUID) (learningquery.CourseEnrollment, error) {
	return r.queries.GetEnrollmentByUserAndCourse(ctx, learningquery.GetEnrollmentByUserAndCourseParams{
		UserID: userID, CourseID: courseID,
	})
}

func (r *LearningReader) UpsertPublishedLessonSnapshot(ctx context.Context, lessonID, courseID uuid.UUID, version int32, snapshot []byte) error {
	return r.queries.UpsertPublishedLessonSnapshot(ctx, learningquery.UpsertPublishedLessonSnapshotParams{
		LessonID: lessonID, CourseID: courseID, Version: version, Snapshot: snapshot,
	})
}
