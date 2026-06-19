package service

import (
	"context"
	"time"

	"github.com/even-app/even-app/services/content/internal/domain"
	"github.com/even-app/even-app/services/content/internal/gen/query"
	"github.com/google/uuid"
)

type Course struct {
	ID               uuid.UUID
	Title            string
	TargetLanguageID uuid.UUID
	UiLanguageID     uuid.UUID
	OwnerID          uuid.UUID
	IsPublished      bool
	Visibility       string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type CourseView struct {
	Course     Course
	InviteCode string
}

func courseFromRow(id uuid.UUID, title string, targetLangID, uiLangID, ownerID uuid.UUID, isPublished bool, visibility string, createdAt, updatedAt time.Time) Course {
	return Course{
		ID: id, Title: title, TargetLanguageID: targetLangID, UiLanguageID: uiLangID,
		OwnerID: ownerID, IsPublished: isPublished, Visibility: visibility,
		CreatedAt: createdAt, UpdatedAt: updatedAt,
	}
}

func courseFromCreate(row query.CreateCourseRow) Course {
	return courseFromRow(row.ID, row.Title, row.TargetLanguageID, row.UiLanguageID, row.OwnerID, row.IsPublished, row.Visibility, row.CreatedAt, row.UpdatedAt)
}

func courseFromGet(row query.GetCourseByIDRow) Course {
	return courseFromRow(row.ID, row.Title, row.TargetLanguageID, row.UiLanguageID, row.OwnerID, row.IsPublished, row.Visibility, row.CreatedAt, row.UpdatedAt)
}

func courseFromList(row query.ListCoursesByOwnerRow) Course {
	return courseFromRow(row.ID, row.Title, row.TargetLanguageID, row.UiLanguageID, row.OwnerID, row.IsPublished, row.Visibility, row.CreatedAt, row.UpdatedAt)
}

func courseFromUpdate(row query.UpdateCourseRow) Course {
	return courseFromRow(row.ID, row.Title, row.TargetLanguageID, row.UiLanguageID, row.OwnerID, row.IsPublished, row.Visibility, row.CreatedAt, row.UpdatedAt)
}

func courseFromPublish(row query.PublishCourseRow) Course {
	return courseFromRow(row.ID, row.Title, row.TargetLanguageID, row.UiLanguageID, row.OwnerID, row.IsPublished, row.Visibility, row.CreatedAt, row.UpdatedAt)
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
		out = append(out, CourseView{Course: courseFromList(row), InviteCode: code})
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
	return CourseView{Course: courseFromGet(row), InviteCode: code}, nil
}

func (s *ContentService) CreateCourse(ctx context.Context, ownerID uuid.UUID, title string, targetLangID, uiLangID uuid.UUID, visibility *string) (CourseView, error) {
	if title == "" {
		return CourseView{}, domain.ErrValidation
	}
	if visibility != nil && !domain.ValidCourseVisibility(*visibility) {
		return CourseView{}, domain.ErrValidation
	}
	vis := domain.CourseVisibilityInviteOnly
	if visibility != nil {
		vis = *visibility
	}
	row, err := s.q.CreateCourse(ctx, query.CreateCourseParams{
		Title:            title,
		TargetLanguageID: targetLangID,
		UiLanguageID:     uiLangID,
		OwnerID:          ownerID,
		Visibility:       vis,
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
	return CourseView{Course: courseFromCreate(row), InviteCode: invite.Code}, nil
}

func (s *ContentService) PatchCourse(ctx context.Context, courseID, userID uuid.UUID, isAdmin bool, params query.UpdateCourseParams) (CourseView, error) {
	if err := s.assertCourseOwner(ctx, courseID, userID, isAdmin); err != nil {
		return CourseView{}, err
	}
	if params.Visibility != nil && !domain.ValidCourseVisibility(*params.Visibility) {
		return CourseView{}, domain.ErrValidation
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
	return CourseView{Course: courseFromUpdate(row), InviteCode: code}, nil
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
	return CourseView{Course: courseFromPublish(row), InviteCode: code}, nil
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
