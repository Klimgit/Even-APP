package repository

import (
	"context"

	"github.com/even-app/even-app/libs/clients"
	authclient "github.com/even-app/even-app/libs/clients/auth"
	"github.com/even-app/even-app/libs/clients/dto"
	learningclient "github.com/even-app/even-app/libs/clients/learning"
	"github.com/even-app/even-app/services/content/internal/gen/authquery"
	"github.com/even-app/even-app/services/content/internal/gen/learningquery"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type LearningSource interface {
	Available() bool
	ListEnrollmentsByCourse(ctx context.Context, courseID uuid.UUID) ([]learningquery.ListEnrollmentsByCourseRow, error)
	GetStudentLessonProgress(ctx context.Context, userID, courseID uuid.UUID) ([]learningquery.GetStudentLessonProgressRow, error)
	CreateEnrollment(ctx context.Context, userID, courseID uuid.UUID, enrolledBy *uuid.UUID) (learningquery.CourseEnrollment, error)
	GetEnrollmentByUserAndCourse(ctx context.Context, userID, courseID uuid.UUID) (learningquery.CourseEnrollment, error)
	UpsertPublishedLessonSnapshot(ctx context.Context, lessonID, courseID uuid.UUID, version int32, snapshot []byte) error
}

type AuthSource interface {
	Available() bool
	ListUsersByIDs(ctx context.Context, ids []uuid.UUID) ([]authquery.ListUsersByIDsRow, error)
	GetUserByEmail(ctx context.Context, email string) (authquery.GetUserByEmailRow, error)
}

var (
	_ LearningSource = (*LearningReader)(nil)
	_ AuthSource     = (*AuthReader)(nil)
)

type LearningRemote struct {
	client *learningclient.Client
}

func NewLearningRemote(baseURL, token string) *LearningRemote {
	return &LearningRemote{client: learningclient.New(baseURL, token)}
}

func (r *LearningRemote) Available() bool {
	return r != nil && r.client.Available()
}

func (r *LearningRemote) ListEnrollmentsByCourse(ctx context.Context, courseID uuid.UUID) ([]learningquery.ListEnrollmentsByCourseRow, error) {
	rows, err := r.client.ListEnrollmentsByCourse(ctx, courseID)
	if err != nil {
		return nil, err
	}
	out := make([]learningquery.ListEnrollmentsByCourseRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, learningquery.ListEnrollmentsByCourseRow{
			UserID: row.UserID, CourseID: row.CourseID, Status: row.Status, EnrolledBy: row.EnrolledBy,
		})
	}
	return out, nil
}

func (r *LearningRemote) GetStudentLessonProgress(ctx context.Context, userID, courseID uuid.UUID) ([]learningquery.GetStudentLessonProgressRow, error) {
	rows, err := r.client.GetStudentLessonProgress(ctx, userID, courseID)
	if err != nil {
		return nil, err
	}
	out := make([]learningquery.GetStudentLessonProgressRow, 0, len(rows))
	for _, row := range rows {
		score := float32(0)
		if row.ScoreAvg != nil {
			score = float32(*row.ScoreAvg)
		}
		out = append(out, learningquery.GetStudentLessonProgressRow{
			LessonID: row.LessonID, Title: row.LessonTitle,
			TotalBlocks: int32(row.BlocksTotal), CompletedBlocks: int32(row.BlocksDone), ScoreAvg: score,
		})
	}
	return out, nil
}

func (r *LearningRemote) CreateEnrollment(ctx context.Context, userID, courseID uuid.UUID, enrolledBy *uuid.UUID) (learningquery.CourseEnrollment, error) {
	row, err := r.client.CreateEnrollment(ctx, dto.CreateEnrollmentRequest{
		UserID: userID, CourseID: courseID, EnrolledBy: enrolledBy,
	})
	if err != nil {
		return learningquery.CourseEnrollment{}, err
	}
	return learningquery.CourseEnrollment{
		ID: row.ID, UserID: row.UserID, CourseID: row.CourseID, Status: row.Status, EnrolledBy: row.EnrolledBy,
	}, nil
}

func (r *LearningRemote) GetEnrollmentByUserAndCourse(ctx context.Context, userID, courseID uuid.UUID) (learningquery.CourseEnrollment, error) {
	row, err := r.client.GetEnrollmentByUserAndCourse(ctx, userID, courseID)
	if err != nil {
		if apiErr, ok := err.(*clients.APIError); ok && apiErr.Status == 404 {
			return learningquery.CourseEnrollment{}, pgx.ErrNoRows
		}
		return learningquery.CourseEnrollment{}, err
	}
	return learningquery.CourseEnrollment{
		ID: row.ID, UserID: row.UserID, CourseID: row.CourseID, Status: row.Status, EnrolledBy: row.EnrolledBy,
	}, nil
}

func (r *LearningRemote) UpsertPublishedLessonSnapshot(ctx context.Context, lessonID, courseID uuid.UUID, version int32, snapshot []byte) error {
	return r.client.UpsertPublishedLessonSnapshot(ctx, dto.UpsertSnapshotRequest{
		LessonID: lessonID, CourseID: courseID, Version: version, Snapshot: snapshot,
	})
}

type AuthRemote struct {
	client *authclient.Client
}

func NewAuthRemote(baseURL, token string) *AuthRemote {
	return &AuthRemote{client: authclient.New(baseURL, token)}
}

func (r *AuthRemote) Available() bool {
	return r != nil && r.client.Available()
}

func (r *AuthRemote) ListUsersByIDs(ctx context.Context, ids []uuid.UUID) ([]authquery.ListUsersByIDsRow, error) {
	rows, err := r.client.ListUsersByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make([]authquery.ListUsersByIDsRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, authquery.ListUsersByIDsRow{
			ID: row.ID, Email: row.Email, DisplayName: row.DisplayName, Role: row.Role, IsAdmin: row.IsAdmin,
		})
	}
	return out, nil
}

func (r *AuthRemote) GetUserByEmail(ctx context.Context, email string) (authquery.GetUserByEmailRow, error) {
	row, err := r.client.GetUserByEmail(ctx, email)
	if err != nil {
		if apiErr, ok := err.(*clients.APIError); ok && apiErr.Status == 404 {
			return authquery.GetUserByEmailRow{}, pgx.ErrNoRows
		}
		return authquery.GetUserByEmailRow{}, err
	}
	return authquery.GetUserByEmailRow{
		ID: row.ID, Email: row.Email, DisplayName: row.DisplayName, Role: row.Role, IsAdmin: row.IsAdmin,
	}, nil
}

var (
	_ LearningSource = (*LearningRemote)(nil)
	_ AuthSource     = (*AuthRemote)(nil)
)
