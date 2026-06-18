package service

import (
	"context"

	"github.com/even-app/even-app/services/content/internal/domain"
	"github.com/even-app/even-app/services/content/internal/gen/query"
	"github.com/google/uuid"
)

type LessonFull struct {
	Lesson   query.Lesson
	Sections []SectionFull
}

type SectionFull struct {
	Section query.LessonSection
	Blocks  []query.LessonBlock
}

func (s *ContentService) buildLessonFull(ctx context.Context, lesson query.Lesson) (LessonFull, error) {
	sections, err := s.q.ListSectionsByLessonID(ctx, lesson.ID)
	if err != nil {
		return LessonFull{}, err
	}
	blocks, err := s.q.ListBlocksByLessonID(ctx, lesson.ID)
	if err != nil {
		return LessonFull{}, err
	}
	bySection := make(map[uuid.UUID][]query.LessonBlock)
	for _, b := range blocks {
		if b.SectionID == nil {
			continue
		}
		bySection[*b.SectionID] = append(bySection[*b.SectionID], b)
	}
	outSections := make([]SectionFull, 0, len(sections))
	for _, sec := range sections {
		outSections = append(outSections, SectionFull{
			Section: sec,
			Blocks:  bySection[sec.ID],
		})
	}
	return LessonFull{Lesson: lesson, Sections: outSections}, nil
}

func (s *ContentService) ListLessons(ctx context.Context, courseID, userID uuid.UUID, isAdmin bool) ([]query.Lesson, error) {
	if err := s.assertCourseOwner(ctx, courseID, userID, isAdmin); err != nil {
		return nil, err
	}
	return s.q.ListLessonsByCourseID(ctx, courseID)
}

func (s *ContentService) CreateLesson(ctx context.Context, courseID, userID uuid.UUID, isAdmin bool, title string, sortOrder *int32) (query.Lesson, error) {
	if err := s.assertCourseOwner(ctx, courseID, userID, isAdmin); err != nil {
		return query.Lesson{}, err
	}
	if title == "" {
		return query.Lesson{}, domain.ErrValidation
	}
	order := int32(0)
	if sortOrder != nil {
		order = *sortOrder
	} else {
		max, err := s.q.MaxLessonSortOrder(ctx, courseID)
		if err != nil {
			return query.Lesson{}, err
		}
		order = int32(max + 1)
	}
	return s.q.CreateLesson(ctx, query.CreateLessonParams{
		CourseID:  courseID,
		Title:     title,
		SortOrder: order,
	})
}

func (s *ContentService) GetLesson(ctx context.Context, lessonID, userID uuid.UUID, isAdmin bool) (LessonFull, error) {
	if err := s.assertLessonOwner(ctx, lessonID, userID, isAdmin); err != nil {
		return LessonFull{}, err
	}
	lesson, err := s.q.GetLessonByID(ctx, lessonID)
	if err != nil {
		return LessonFull{}, mapNotFound(err)
	}
	return s.buildLessonFull(ctx, lesson)
}

func (s *ContentService) PatchLesson(ctx context.Context, lessonID, userID uuid.UUID, isAdmin bool, expectedVersion int32, params query.UpdateLessonParams) (LessonFull, error) {
	if err := s.assertLessonOwner(ctx, lessonID, userID, isAdmin); err != nil {
		return LessonFull{}, err
	}
	params.ID = lessonID
	params.ExpectedVersion = expectedVersion
	lesson, err := s.q.UpdateLesson(ctx, params)
	if err != nil {
		if mapNotFound(err) == domain.ErrNotFound {
			if _, getErr := s.q.GetLessonByID(ctx, lessonID); getErr != nil {
				return LessonFull{}, domain.ErrNotFound
			}
			return LessonFull{}, domain.ErrConflict
		}
		return LessonFull{}, err
	}
	return s.buildLessonFull(ctx, lesson)
}

func (s *ContentService) DeleteLesson(ctx context.Context, lessonID, userID uuid.UUID, isAdmin bool) error {
	if err := s.assertLessonOwner(ctx, lessonID, userID, isAdmin); err != nil {
		return err
	}
	return s.q.DeleteLesson(ctx, lessonID)
}

func (s *ContentService) PublishLesson(ctx context.Context, lessonID, userID uuid.UUID, isAdmin bool) (LessonFull, error) {
	if err := s.assertLessonOwner(ctx, lessonID, userID, isAdmin); err != nil {
		return LessonFull{}, err
	}
	lesson, err := s.q.PublishLesson(ctx, lessonID)
	if err != nil {
		return LessonFull{}, mapNotFound(err)
	}
	return s.buildLessonFull(ctx, lesson)
}

func (s *ContentService) CreateSection(ctx context.Context, lessonID, userID uuid.UUID, isAdmin bool, title, kind string, sortOrder *int32) (query.LessonSection, error) {
	if err := s.assertLessonOwner(ctx, lessonID, userID, isAdmin); err != nil {
		return query.LessonSection{}, err
	}
	if title == "" {
		return query.LessonSection{}, domain.ErrValidation
	}
	if kind == "" {
		kind = "content"
	}
	order := int32(0)
	if sortOrder != nil {
		order = *sortOrder
	} else {
		max, err := s.q.MaxSectionSortOrder(ctx, lessonID)
		if err != nil {
			return query.LessonSection{}, err
		}
		order = int32(max + 1)
	}
	return s.q.CreateSection(ctx, query.CreateSectionParams{
		LessonID:    lessonID,
		Title:       title,
		SortOrder:   order,
		SectionKind: kind,
	})
}

func (s *ContentService) PatchSection(ctx context.Context, sectionID, userID uuid.UUID, isAdmin bool, params query.UpdateSectionParams) (query.LessonSection, error) {
	if err := s.assertSectionOwner(ctx, sectionID, userID, isAdmin); err != nil {
		return query.LessonSection{}, err
	}
	params.ID = sectionID
	row, err := s.q.UpdateSection(ctx, params)
	if err != nil {
		return query.LessonSection{}, mapNotFound(err)
	}
	return row, nil
}

func (s *ContentService) DeleteSection(ctx context.Context, sectionID, userID uuid.UUID, isAdmin bool) error {
	if err := s.assertSectionOwner(ctx, sectionID, userID, isAdmin); err != nil {
		return err
	}
	return s.q.DeleteSection(ctx, sectionID)
}

func (s *ContentService) ReorderSections(ctx context.Context, lessonID, userID uuid.UUID, isAdmin bool, ids []uuid.UUID) error {
	if err := s.assertLessonOwner(ctx, lessonID, userID, isAdmin); err != nil {
		return err
	}
	all, err := s.q.ListSectionsByLessonID(ctx, lessonID)
	if err != nil {
		return err
	}
	byID := make(map[uuid.UUID]struct{}, len(all))
	for _, row := range all {
		byID[row.ID] = struct{}{}
	}
	seen := make(map[uuid.UUID]struct{}, len(ids))
	for i, id := range ids {
		if _, ok := byID[id]; !ok {
			return domain.ErrNotFound
		}
		seen[id] = struct{}{}
		if err := s.q.SetSectionSortOrder(ctx, query.SetSectionSortOrderParams{ID: id, SortOrder: int32(i)}); err != nil {
			return err
		}
	}
	next := len(ids)
	for _, row := range all {
		if _, ok := seen[row.ID]; ok {
			continue
		}
		if err := s.q.SetSectionSortOrder(ctx, query.SetSectionSortOrderParams{ID: row.ID, SortOrder: int32(next)}); err != nil {
			return err
		}
		next++
	}
	return nil
}

func (s *ContentService) CreateBlock(ctx context.Context, lessonID, userID uuid.UUID, isAdmin bool, params query.CreateBlockParams) (query.LessonBlock, error) {
	if err := s.assertLessonOwner(ctx, lessonID, userID, isAdmin); err != nil {
		return query.LessonBlock{}, err
	}
	if params.BlockType == "" {
		return query.LessonBlock{}, domain.ErrValidation
	}
	params.LessonID = lessonID
	if len(params.Config) == 0 {
		params.Config = []byte("{}")
	}
	return s.q.CreateBlock(ctx, params)
}

func (s *ContentService) GetBlock(ctx context.Context, blockID, userID uuid.UUID, isAdmin bool) (query.LessonBlock, error) {
	if err := s.assertBlockOwner(ctx, blockID, userID, isAdmin); err != nil {
		return query.LessonBlock{}, err
	}
	row, err := s.q.GetBlockByID(ctx, blockID)
	if err != nil {
		return query.LessonBlock{}, mapNotFound(err)
	}
	return row, nil
}

func (s *ContentService) PatchBlock(ctx context.Context, blockID, userID uuid.UUID, isAdmin bool, params query.UpdateBlockParams) (query.LessonBlock, error) {
	if err := s.assertBlockOwner(ctx, blockID, userID, isAdmin); err != nil {
		return query.LessonBlock{}, err
	}
	params.ID = blockID
	row, err := s.q.UpdateBlock(ctx, params)
	if err != nil {
		return query.LessonBlock{}, mapNotFound(err)
	}
	return row, nil
}

func (s *ContentService) DeleteBlock(ctx context.Context, blockID, userID uuid.UUID, isAdmin bool) error {
	if err := s.assertBlockOwner(ctx, blockID, userID, isAdmin); err != nil {
		return err
	}
	return s.q.DeleteBlock(ctx, blockID)
}

func (s *ContentService) ReorderBlocks(ctx context.Context, lessonID, userID uuid.UUID, isAdmin bool, ids []uuid.UUID) error {
	if err := s.assertLessonOwner(ctx, lessonID, userID, isAdmin); err != nil {
		return err
	}
	all, err := s.q.ListBlocksByLessonID(ctx, lessonID)
	if err != nil {
		return err
	}
	byID := make(map[uuid.UUID]struct{}, len(all))
	for _, row := range all {
		byID[row.ID] = struct{}{}
	}
	seen := make(map[uuid.UUID]struct{}, len(ids))
	for i, id := range ids {
		if _, ok := byID[id]; !ok {
			return domain.ErrNotFound
		}
		seen[id] = struct{}{}
		if err := s.q.SetBlockSortOrder(ctx, query.SetBlockSortOrderParams{ID: id, SortOrder: int32(i)}); err != nil {
			return err
		}
	}
	next := len(ids)
	for _, row := range all {
		if _, ok := seen[row.ID]; ok {
			continue
		}
		if err := s.q.SetBlockSortOrder(ctx, query.SetBlockSortOrderParams{ID: row.ID, SortOrder: int32(next)}); err != nil {
			return err
		}
		next++
	}
	return nil
}
