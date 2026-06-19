package service

import (
	"context"
	"errors"

	"github.com/even-app/even-app/services/content/internal/domain"
	"github.com/even-app/even-app/services/content/internal/gen/query"
	"github.com/even-app/even-app/services/content/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ContentService struct {
	q        *query.Queries
	learning repository.LearningSource
	auth     repository.AuthSource
}

func NewContentService(q *query.Queries, learning repository.LearningSource, auth repository.AuthSource) *ContentService {
	return &ContentService{q: q, learning: learning, auth: auth}
}

func mapNotFound(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	return err
}

func (s *ContentService) assertCourseOwner(ctx context.Context, courseID, userID uuid.UUID, isAdmin bool) error {
	owner, err := s.q.GetCourseOwner(ctx, courseID)
	if err != nil {
		return mapNotFound(err)
	}
	if !isAdmin && owner != userID {
		return domain.ErrForbidden
	}
	return nil
}

func (s *ContentService) assertLessonOwner(ctx context.Context, lessonID, userID uuid.UUID, isAdmin bool) error {
	owner, err := s.q.GetLessonCourseOwner(ctx, lessonID)
	if err != nil {
		return mapNotFound(err)
	}
	if !isAdmin && owner != userID {
		return domain.ErrForbidden
	}
	return nil
}

func (s *ContentService) assertSectionOwner(ctx context.Context, sectionID, userID uuid.UUID, isAdmin bool) error {
	owner, err := s.q.GetSectionCourseOwner(ctx, sectionID)
	if err != nil {
		return mapNotFound(err)
	}
	if !isAdmin && owner != userID {
		return domain.ErrForbidden
	}
	return nil
}

func (s *ContentService) assertBlockOwner(ctx context.Context, blockID, userID uuid.UUID, isAdmin bool) error {
	owner, err := s.q.GetBlockCourseOwner(ctx, blockID)
	if err != nil {
		return mapNotFound(err)
	}
	if !isAdmin && owner != userID {
		return domain.ErrForbidden
	}
	return nil
}

func (s *ContentService) courseInviteCode(ctx context.Context, courseID uuid.UUID) (string, error) {
	row, err := s.q.GetInviteCodeByCourseID(ctx, courseID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return row.Code, nil
}

func (s *ContentService) ensureInviteCode(ctx context.Context, courseID uuid.UUID) (string, error) {
	code, err := s.courseInviteCode(ctx, courseID)
	if err != nil {
		return "", err
	}
	if code != "" {
		return code, nil
	}
	newCode, err := domain.GenerateInviteCode()
	if err != nil {
		return "", err
	}
	row, err := s.q.UpsertInviteCode(ctx, query.UpsertInviteCodeParams{
		CourseID: courseID,
		Code:     newCode,
	})
	if err != nil {
		return "", err
	}
	return row.Code, nil
}
