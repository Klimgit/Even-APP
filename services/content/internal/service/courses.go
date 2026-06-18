package service

import (
	"context"

	"github.com/even-app/even-app/services/content/internal/domain"
	"github.com/even-app/even-app/services/content/internal/gen/query"
	"github.com/google/uuid"
)

type CourseView struct {
	Course     query.Course
	InviteCode string
}

func (s *ContentService) ListCourses(ctx context.Context, ownerID uuid.UUID) ([]CourseView, error) {
	rows, err := s.q.ListCoursesByOwner(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	out := make([]CourseView, 0, len(rows))
	for _, row := range rows {
		code, err := s.courseInviteCode(ctx, row.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, CourseView{Course: row, InviteCode: code})
	}
	return out, nil
}

func (s *ContentService) GetCourse(ctx context.Context, courseID, userID uuid.UUID, isAdmin bool) (CourseView, error) {
	if err := s.assertCourseOwner(ctx, courseID, userID, isAdmin); err != nil {
		return CourseView{}, err
	}
	row, err := s.q.GetCourseByID(ctx, courseID)
	if err != nil {
		return CourseView{}, mapNotFound(err)
	}
	code, err := s.ensureInviteCode(ctx, courseID)
	if err != nil {
		return CourseView{}, err
	}
	return CourseView{Course: row, InviteCode: code}, nil
}

func (s *ContentService) CreateCourse(ctx context.Context, ownerID uuid.UUID, title string, targetLangID, uiLangID uuid.UUID) (CourseView, error) {
	if title == "" {
		return CourseView{}, domain.ErrValidation
	}
	row, err := s.q.CreateCourse(ctx, query.CreateCourseParams{
		Title:            title,
		TargetLanguageID: targetLangID,
		UiLanguageID:     uiLangID,
		OwnerID:          ownerID,
	})
	if err != nil {
		return CourseView{}, err
	}
	code, err := domain.GenerateInviteCode()
	if err != nil {
		return CourseView{}, err
	}
	invite, err := s.q.UpsertInviteCode(ctx, query.UpsertInviteCodeParams{
		CourseID: row.ID,
		Code:     code,
	})
	if err != nil {
		return CourseView{}, err
	}
	return CourseView{Course: row, InviteCode: invite.Code}, nil
}

func (s *ContentService) PatchCourse(ctx context.Context, courseID, userID uuid.UUID, isAdmin bool, params query.UpdateCourseParams) (CourseView, error) {
	if err := s.assertCourseOwner(ctx, courseID, userID, isAdmin); err != nil {
		return CourseView{}, err
	}
	params.ID = courseID
	row, err := s.q.UpdateCourse(ctx, params)
	if err != nil {
		return CourseView{}, mapNotFound(err)
	}
	code, err := s.ensureInviteCode(ctx, courseID)
	if err != nil {
		return CourseView{}, err
	}
	return CourseView{Course: row, InviteCode: code}, nil
}

func (s *ContentService) DeleteCourse(ctx context.Context, courseID, userID uuid.UUID, isAdmin bool) error {
	if err := s.assertCourseOwner(ctx, courseID, userID, isAdmin); err != nil {
		return err
	}
	return s.q.DeleteCourse(ctx, courseID)
}

func (s *ContentService) PublishCourse(ctx context.Context, courseID, userID uuid.UUID, isAdmin bool) (CourseView, error) {
	if err := s.assertCourseOwner(ctx, courseID, userID, isAdmin); err != nil {
		return CourseView{}, err
	}
	row, err := s.q.PublishCourse(ctx, courseID)
	if err != nil {
		return CourseView{}, mapNotFound(err)
	}
	code, err := s.ensureInviteCode(ctx, courseID)
	if err != nil {
		return CourseView{}, err
	}
	return CourseView{Course: row, InviteCode: code}, nil
}

func (s *ContentService) GetInviteCode(ctx context.Context, courseID, userID uuid.UUID, isAdmin bool) (string, error) {
	if err := s.assertCourseOwner(ctx, courseID, userID, isAdmin); err != nil {
		return "", err
	}
	return s.ensureInviteCode(ctx, courseID)
}

func (s *ContentService) RegenerateInviteCode(ctx context.Context, courseID, userID uuid.UUID, isAdmin bool) (string, error) {
	if err := s.assertCourseOwner(ctx, courseID, userID, isAdmin); err != nil {
		return "", err
	}
	if _, err := s.q.GetCourseByID(ctx, courseID); err != nil {
		return "", mapNotFound(err)
	}
	code, err := domain.GenerateInviteCode()
	if err != nil {
		return "", err
	}
	row, err := s.q.UpsertInviteCode(ctx, query.UpsertInviteCodeParams{
		CourseID: courseID,
		Code:     code,
	})
	if err != nil {
		return "", err
	}
	return row.Code, nil
}
