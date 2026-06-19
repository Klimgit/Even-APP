package service

import (
	"context"

	"github.com/even-app/even-app/services/content/internal/domain"
	"github.com/even-app/even-app/services/content/internal/gen/query"
	"github.com/google/uuid"
)

type Lesson = query.GetLessonByIDRow

type LessonFull struct {
	Lesson   Lesson
	Sections []SectionFull
}

type SectionFull struct {
	Section query.LessonSection
	Blocks  []query.LessonBlock
}

func (s *ContentService) buildLessonFull(ctx context.Context, lesson Lesson) (LessonFull, error) {
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

func (s *ContentService) ListLessons(ctx context.Context, courseID, userID uuid.UUID, isAdmin bool) ([]Lesson, error) {
	if err := s.assertCourseOwner(ctx, courseID, userID, isAdmin); err != nil {
		return nil, err
	}
	rows, err := s.q.ListLessonsByCourseID(ctx, courseID)
	if err != nil {
		return nil, err
	}
	out := make([]Lesson, len(rows))
	for i, r := range rows {
		out[i] = Lesson(r)
	}
	return out, nil
}

func (s *ContentService) CreateLessonInCourse(ctx context.Context, courseID, userID uuid.UUID, isAdmin bool, title string, sortOrder *int32, moduleID *uuid.UUID) (Lesson, error) {
	if err := s.assertCourseOwner(ctx, courseID, userID, isAdmin); err != nil {
		return Lesson{}, err
	}
	modID := uuid.Nil
	if moduleID != nil {
		modID = *moduleID
	} else {
		mod, err := s.ensureDefaultModule(ctx, courseID)
		if err != nil {
			return Lesson{}, err
		}
		modID = mod.ID
	}
	return s.CreateLesson(ctx, modID, userID, isAdmin, title, sortOrder)
}

func (s *ContentService) ListPlatformCourses(ctx context.Context, search *string, page, limit int) ([]query.ListAllCoursesRow, int32, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	total, err := s.q.CountAllCourses(ctx, search)
	if err != nil {
		return nil, 0, err
	}
	rows, err := s.q.ListAllCourses(ctx, query.ListAllCoursesParams{
		Search: search, Limit: int32(limit), Offset: int32((page - 1) * limit),
	})
	return rows, total, err
}

func (s *ContentService) CreateLesson(ctx context.Context, moduleID, userID uuid.UUID, isAdmin bool, title string, sortOrder *int32) (Lesson, error) {
	if err := s.assertModuleOwner(ctx, moduleID, userID, isAdmin); err != nil {
		return Lesson{}, err
	}
	mod, err := s.q.GetModuleByID(ctx, moduleID)
	if err != nil {
		return Lesson{}, mapNotFound(err)
	}
	if title == "" {
		return Lesson{}, domain.ErrValidation
	}
	order := int32(0)
	if sortOrder != nil {
		order = *sortOrder
	} else {
		max, err := s.q.MaxLessonSortOrderByModule(ctx, moduleID)
		if err != nil {
			return Lesson{}, err
		}
		order = int32(max + 1)
	}
	row, err := s.q.CreateLesson(ctx, query.CreateLessonParams{
		CourseID: mod.CourseID, ModuleID: moduleID, Title: title, SortOrder: order,
	})
	if err != nil {
		return Lesson{}, err
	}
	return Lesson(row), nil
}

func (s *ContentService) GetLesson(ctx context.Context, lessonID, userID uuid.UUID, isAdmin bool) (LessonFull, error) {
	if err := s.assertLessonOwner(ctx, lessonID, userID, isAdmin); err != nil {
		return LessonFull{}, err
	}
	row, err := s.q.GetLessonByID(ctx, lessonID)
	if err != nil {
		return LessonFull{}, mapNotFound(err)
	}
	return s.buildLessonFull(ctx, Lesson(row))
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
	return s.buildLessonFull(ctx, Lesson(lesson))
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
	return s.buildLessonFull(ctx, Lesson(lesson))
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
	if err := validateBlockInput(params.BlockType, params.Config); err != nil {
		return query.LessonBlock{}, err
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
	if params.BlockType != nil || len(params.Config) > 0 {
		blockType := ""
		if params.BlockType != nil {
			blockType = *params.BlockType
		} else {
			existing, err := s.q.GetBlockByID(ctx, blockID)
			if err != nil {
				return query.LessonBlock{}, mapNotFound(err)
			}
			blockType = existing.BlockType
		}
		if err := validateBlockInput(blockType, params.Config); err != nil {
			return query.LessonBlock{}, err
		}
	}
	params.ID = blockID
	row, err := s.q.UpdateBlock(ctx, params)
	if err != nil {
		return query.LessonBlock{}, mapNotFound(err)
	}
	return row, nil
}

func validateBlockInput(blockType string, config []byte) error {
	return domain.ValidateBlockConfig(blockType, config)
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
