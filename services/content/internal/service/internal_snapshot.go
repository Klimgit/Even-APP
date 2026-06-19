package service

import (
	"context"
	"encoding/json"

	"github.com/even-app/even-app/libs/clients/dto"
	"github.com/even-app/even-app/services/content/internal/domain"
	"github.com/google/uuid"
)

func (s *ContentService) PublishedLessonSnapshots(ctx context.Context, courseID uuid.UUID) ([]dto.PublishedLessonSnapshot, error) {
	lessons, err := s.q.ListPublishedLessonsByCourse(ctx, courseID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.PublishedLessonSnapshot, 0, len(lessons))
	for _, summary := range lessons {
		lesson, err := s.q.GetLessonByID(ctx, summary.ID)
		if err != nil {
			return nil, err
		}
		full, err := s.buildLessonFull(ctx, lesson)
		if err != nil {
			return nil, err
		}
		raw, err := json.Marshal(buildLessonSnapshot(full))
		if err != nil {
			return nil, err
		}
		out = append(out, dto.PublishedLessonSnapshot{
			LessonID: lesson.ID,
			CourseID: courseID,
			Version:  lesson.Version,
			Snapshot: raw,
		})
	}
	return out, nil
}

func (s *ContentService) LessonSnapshotJSON(ctx context.Context, lessonID uuid.UUID) (json.RawMessage, error) {
	lesson, err := s.q.GetLessonByID(ctx, lessonID)
	if err != nil {
		return nil, err
	}
	if lesson.Status != "published" {
		return nil, domain.ErrNotFound
	}
	full, err := s.buildLessonFull(ctx, lesson)
	if err != nil {
		return nil, err
	}
	return json.Marshal(buildLessonSnapshot(full))
}

func (s *ContentService) BlockView(ctx context.Context, blockID uuid.UUID) (dto.BlockView, error) {
	row, err := s.q.GetLessonBlockWithCourse(ctx, blockID)
	if err != nil {
		return dto.BlockView{}, err
	}
	cfg := json.RawMessage(row.Config)
	if len(cfg) == 0 {
		cfg = json.RawMessage("{}")
	}
	block := map[string]any{
		"id": row.ID, "sort_order": row.SortOrder, "block_type": row.BlockType,
		"config": cfg, "is_homework": row.IsHomework,
		"is_gradable": domain.IsGradableBlockType(row.BlockType),
	}
	if row.SectionID != nil {
		block["section_id"] = *row.SectionID
	}
	if row.DisplayLabel != nil {
		block["display_label"] = *row.DisplayLabel
	}
	if row.Title != nil {
		block["title"] = *row.Title
	}
	raw, err := json.Marshal(block)
	if err != nil {
		return dto.BlockView{}, err
	}
	return dto.BlockView{Block: raw, CourseID: row.CourseID, LessonTitle: row.LessonTitle}, nil
}
