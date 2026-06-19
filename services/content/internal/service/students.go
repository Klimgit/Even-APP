package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/even-app/even-app/services/content/internal/domain"
	"github.com/even-app/even-app/services/content/internal/gen/authquery"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type StudentView struct {
	ID              uuid.UUID
	Email           string
	DisplayName     *string
	EnrolledCourses []StudentEnrolledCourse
}

type StudentEnrolledCourse struct {
	ID    uuid.UUID
	Title string
}

func (s *ContentService) ListCourseStudents(ctx context.Context, courseID, ownerID uuid.UUID, isAdmin bool) ([]StudentView, error) {
	if err := s.assertCourseOwner(ctx, courseID, ownerID, isAdmin); err != nil {
		return nil, err
	}
	if s.learning == nil || !s.learning.Available() {
		return []StudentView{}, nil
	}
	if s.auth == nil || !s.auth.Available() {
		return []StudentView{}, nil
	}

	course, err := s.q.GetCourseByID(ctx, courseID)
	if err != nil {
		return nil, mapNotFound(err)
	}

	enrolls, err := s.learning.ListEnrollmentsByCourse(ctx, courseID)
	if err != nil {
		return nil, err
	}
	if len(enrolls) == 0 {
		return []StudentView{}, nil
	}

	userIDs := make([]uuid.UUID, len(enrolls))
	for i, e := range enrolls {
		userIDs[i] = e.UserID
	}
	users, err := s.auth.ListUsersByIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	byID := make(map[uuid.UUID]authquery.ListUsersByIDsRow, len(users))
	for _, u := range users {
		byID[u.ID] = u
	}

	out := make([]StudentView, 0, len(enrolls))
	for _, e := range enrolls {
		u, ok := byID[e.UserID]
		if !ok {
			continue
		}
		out = append(out, StudentView{
			ID:          u.ID,
			Email:       u.Email,
			DisplayName: u.DisplayName,
			EnrolledCourses: []StudentEnrolledCourse{{
				ID: course.ID, Title: course.Title,
			}},
		})
	}
	return out, nil
}

func (s *ContentService) GetStudentProgress(ctx context.Context, courseID, studentID, ownerID uuid.UUID, isAdmin bool) (StudentProgressView, error) {
	if err := s.assertCourseOwner(ctx, courseID, ownerID, isAdmin); err != nil {
		return StudentProgressView{}, err
	}
	if s.learning == nil || !s.learning.Available() {
		return StudentProgressView{UserID: studentID, CourseID: courseID, Lessons: nil}, nil
	}

	rows, err := s.learning.GetStudentLessonProgress(ctx, studentID, courseID)
	if err != nil {
		return StudentProgressView{}, err
	}
	lessons := make([]StudentProgressLesson, len(rows))
	for i, row := range rows {
		lessons[i] = StudentProgressLesson{
			LessonID:        row.LessonID,
			Title:           row.Title,
			CompletedBlocks: int(row.CompletedBlocks),
			TotalBlocks:     int(row.TotalBlocks),
			ScoreAvg:        row.ScoreAvg,
		}
	}
	return StudentProgressView{
		UserID: studentID, CourseID: courseID, Lessons: lessons,
	}, nil
}

func (s *ContentService) EnrollStudent(ctx context.Context, courseID uuid.UUID, email string, ownerID uuid.UUID, isAdmin bool) (StudentView, error) {
	if err := s.assertCourseOwner(ctx, courseID, ownerID, isAdmin); err != nil {
		return StudentView{}, err
	}
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return StudentView{}, domain.ErrValidation
	}
	if s.learning == nil || !s.learning.Available() {
		return StudentView{}, errors.New("learning service not configured")
	}
	if s.auth == nil || !s.auth.Available() {
		return StudentView{}, errors.New("auth service not configured")
	}

	course, err := s.q.GetCourseByID(ctx, courseID)
	if err != nil {
		return StudentView{}, mapNotFound(err)
	}

	user, err := s.auth.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return StudentView{}, domain.ErrNotFound
		}
		return StudentView{}, err
	}

	if _, err := s.learning.GetEnrollmentByUserAndCourse(ctx, user.ID, courseID); err == nil {
		return StudentView{}, domain.ErrConflict
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return StudentView{}, err
	}

	enrolledBy := ownerID
	if _, err := s.learning.CreateEnrollment(ctx, user.ID, courseID, &enrolledBy); err != nil {
		return StudentView{}, err
	}
	if err := s.syncPublishedLessonsForCourse(ctx, courseID); err != nil {
		return StudentView{}, err
	}

	return StudentView{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		EnrolledCourses: []StudentEnrolledCourse{{
			ID: course.ID, Title: course.Title,
		}},
	}, nil
}

func (s *ContentService) syncPublishedLessonsForCourse(ctx context.Context, courseID uuid.UUID) error {
	lessons, err := s.q.ListLessonsByCourseID(ctx, courseID)
	if err != nil {
		return err
	}
	for _, lesson := range lessons {
		if lesson.Status != "published" {
			continue
		}
		full, err := s.buildLessonFull(ctx, Lesson(lesson))
		if err != nil {
			return err
		}
		raw, err := json.Marshal(buildLessonSnapshot(full))
		if err != nil {
			return err
		}
		if err := s.learning.UpsertPublishedLessonSnapshot(ctx, lesson.ID, courseID, lesson.Version, raw); err != nil {
			return err
		}
	}
	return nil
}

type lessonSnapshot struct {
	ID        uuid.UUID         `json:"id"`
	CourseID  uuid.UUID         `json:"course_id"`
	Title     string            `json:"title"`
	SortOrder int               `json:"sort_order"`
	Version   int               `json:"version"`
	Status    string            `json:"status"`
	Sections  []sectionSnapshot `json:"sections"`
	Blocks    []blockSnapshot   `json:"blocks"`
}

type sectionSnapshot struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	SortOrder   int       `json:"sort_order"`
	SectionKind string    `json:"section_kind"`
}

type blockSnapshot struct {
	ID           uuid.UUID       `json:"id"`
	SectionID    *uuid.UUID      `json:"section_id,omitempty"`
	SortOrder    int             `json:"sort_order"`
	DisplayLabel *string         `json:"display_label,omitempty"`
	Title        *string         `json:"title,omitempty"`
	BlockType    string          `json:"block_type"`
	Config       json.RawMessage `json:"config"`
	IsHomework   bool            `json:"is_homework"`
	IsGradable   bool            `json:"is_gradable"`
}

func buildLessonSnapshot(full LessonFull) lessonSnapshot {
	snap := lessonSnapshot{
		ID: full.Lesson.ID, CourseID: full.Lesson.CourseID, Title: full.Lesson.Title,
		SortOrder: int(full.Lesson.SortOrder), Version: int(full.Lesson.Version),
		Status:   full.Lesson.Status,
		Sections: make([]sectionSnapshot, 0, len(full.Sections)),
		Blocks:   make([]blockSnapshot, 0),
	}
	for _, sec := range full.Sections {
		snap.Sections = append(snap.Sections, sectionSnapshot{
			ID: sec.Section.ID, Title: sec.Section.Title,
			SortOrder: int(sec.Section.SortOrder), SectionKind: sec.Section.SectionKind,
		})
		for _, b := range sec.Blocks {
			cfg := json.RawMessage(b.Config)
			if len(cfg) == 0 {
				cfg = json.RawMessage("{}")
			}
			snap.Blocks = append(snap.Blocks, blockSnapshot{
				ID: b.ID, SectionID: b.SectionID, SortOrder: int(b.SortOrder),
				DisplayLabel: b.DisplayLabel, Title: b.Title,
				BlockType: b.BlockType, Config: cfg, IsHomework: b.IsHomework,
				IsGradable: domain.IsGradableBlockType(b.BlockType),
			})
		}
	}
	return snap
}
