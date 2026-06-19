package service

import (
	"context"

	"github.com/even-app/even-app/services/content/internal/domain"
	"github.com/even-app/even-app/services/content/internal/gen/query"
	"github.com/google/uuid"
)

func (s *ContentService) ListModules(ctx context.Context, courseID, userID uuid.UUID, isAdmin bool) ([]query.CourseModule, error) {
	if err := s.assertCourseOwner(ctx, courseID, userID, isAdmin); err != nil {
		return nil, err
	}
	return s.q.ListModulesByCourseID(ctx, courseID)
}

func (s *ContentService) CreateModule(ctx context.Context, courseID, userID uuid.UUID, isAdmin bool, title string, sortOrder *int32) (query.CourseModule, error) {
	if err := s.assertCourseOwner(ctx, courseID, userID, isAdmin); err != nil {
		return query.CourseModule{}, err
	}
	if title == "" {
		return query.CourseModule{}, domain.ErrValidation
	}
	order := int32(0)
	if sortOrder != nil {
		order = *sortOrder
	} else {
		max, err := s.q.MaxModuleSortOrder(ctx, courseID)
		if err != nil {
			return query.CourseModule{}, err
		}
		order = int32(max + 1)
	}
	return s.q.CreateModule(ctx, query.CreateModuleParams{
		CourseID: courseID, Title: title, SortOrder: order,
	})
}

func (s *ContentService) PatchModule(ctx context.Context, moduleID, userID uuid.UUID, isAdmin bool, params query.UpdateModuleParams) (query.CourseModule, error) {
	if err := s.assertModuleOwner(ctx, moduleID, userID, isAdmin); err != nil {
		return query.CourseModule{}, err
	}
	params.ID = moduleID
	row, err := s.q.UpdateModule(ctx, params)
	if err != nil {
		return query.CourseModule{}, mapNotFound(err)
	}
	return row, nil
}

func (s *ContentService) DeleteModule(ctx context.Context, moduleID, userID uuid.UUID, isAdmin bool) error {
	if err := s.assertModuleOwner(ctx, moduleID, userID, isAdmin); err != nil {
		return err
	}
	n, err := s.q.CountLessonsInModule(ctx, moduleID)
	if err != nil {
		return err
	}
	if n > 0 {
		return domain.ErrConflict
	}
	return s.q.DeleteModule(ctx, moduleID)
}

func (s *ContentService) ReorderModules(ctx context.Context, courseID, userID uuid.UUID, isAdmin bool, ids []uuid.UUID) error {
	if err := s.assertCourseOwner(ctx, courseID, userID, isAdmin); err != nil {
		return err
	}
	modules, err := s.q.ListModulesByCourseID(ctx, courseID)
	if err != nil {
		return err
	}
	owned := make(map[uuid.UUID]struct{}, len(modules))
	for _, m := range modules {
		owned[m.ID] = struct{}{}
	}
	for i, id := range ids {
		if _, ok := owned[id]; !ok {
			return domain.ErrValidation
		}
		if err := s.q.SetModuleSortOrder(ctx, query.SetModuleSortOrderParams{ID: id, SortOrder: int32(i)}); err != nil {
			return err
		}
	}
	return nil
}

func (s *ContentService) ensureDefaultModule(ctx context.Context, courseID uuid.UUID) (query.CourseModule, error) {
	modules, err := s.q.ListModulesByCourseID(ctx, courseID)
	if err != nil {
		return query.CourseModule{}, err
	}
	if len(modules) > 0 {
		return modules[0], nil
	}
	return s.q.CreateModule(ctx, query.CreateModuleParams{
		CourseID: courseID, Title: "Основной", SortOrder: 0,
	})
}

func (s *ContentService) assertModuleOwner(ctx context.Context, moduleID, userID uuid.UUID, isAdmin bool) error {
	owner, err := s.q.GetModuleCourseOwner(ctx, moduleID)
	if err != nil {
		return mapNotFound(err)
	}
	if !isAdmin && owner != userID {
		return domain.ErrForbidden
	}
	return nil
}
