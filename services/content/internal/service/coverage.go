package service

import (
	"context"

	"github.com/even-app/even-app/services/content/internal/domain"
	"github.com/google/uuid"
)

type CoverageSummary struct {
	IntroducedCount int
	ExercisedCount  int
	LexemeCount     int
}

type CoverageByLesson struct {
	LessonID    uuid.UUID
	LessonTitle string
	Items       []domain.LexemeRef
}

type FormCoverage struct {
	LexemeID string
	Lemma    string
	Forms    []FormCoverageEntry
}

type FormCoverageEntry struct {
	FormID     string
	Form       string
	Introduced bool
	Exercised  bool
}

type StudentProgressView struct {
	UserID   uuid.UUID
	CourseID uuid.UUID
	Lessons  []StudentProgressLesson
}

type StudentProgressLesson struct {
	LessonID        uuid.UUID
	Title           string
	CompletedBlocks int
	TotalBlocks     int
	ScoreAvg        float32
}

func (s *ContentService) GetCoverageSummary(ctx context.Context, courseID, userID uuid.UUID, isAdmin bool) (CoverageSummary, error) {
	if err := s.assertCourseOwner(ctx, courseID, userID, isAdmin); err != nil {
		return CoverageSummary{}, err
	}
	blocks, err := s.q.ListBlocksByCourseID(ctx, courseID)
	if err != nil {
		return CoverageSummary{}, err
	}
	introduced := make(map[string]struct{})
	exercised := make(map[string]struct{})
	all := make(map[string]struct{})
	for _, b := range blocks {
		for _, ref := range domain.ExtractLexemeRefs(domain.BlockLexemeSource{
			BlockType:    b.BlockType,
			Config:       b.Config,
			DisplayLabel: b.DisplayLabel,
		}) {
			all[ref.LexemeID] = struct{}{}
			switch ref.UsageKind {
			case domain.UsageIntroduced:
				introduced[ref.LexemeID] = struct{}{}
			case domain.UsageExercised:
				exercised[ref.LexemeID] = struct{}{}
			}
		}
	}
	return CoverageSummary{
		IntroducedCount: len(introduced),
		ExercisedCount:  len(exercised),
		LexemeCount:     len(all),
	}, nil
}

func (s *ContentService) GetCoverageByLesson(ctx context.Context, courseID, userID uuid.UUID, isAdmin bool) ([]CoverageByLesson, error) {
	if err := s.assertCourseOwner(ctx, courseID, userID, isAdmin); err != nil {
		return nil, err
	}
	blocks, err := s.q.ListBlocksByCourseID(ctx, courseID)
	if err != nil {
		return nil, err
	}
	byLesson := make(map[uuid.UUID]*CoverageByLesson)
	order := make([]uuid.UUID, 0)
	for _, b := range blocks {
		entry, ok := byLesson[b.LessonID]
		if !ok {
			entry = &CoverageByLesson{LessonID: b.LessonID, LessonTitle: b.LessonTitle}
			byLesson[b.LessonID] = entry
			order = append(order, b.LessonID)
		}
		entry.Items = append(entry.Items, domain.ExtractLexemeRefs(domain.BlockLexemeSource{
			BlockType:    b.BlockType,
			Config:       b.Config,
			DisplayLabel: b.DisplayLabel,
		})...)
	}
	out := make([]CoverageByLesson, 0, len(order))
	for _, id := range order {
		out = append(out, *byLesson[id])
	}
	return out, nil
}

func (s *ContentService) GetFormsCoverage(ctx context.Context, courseID, userID uuid.UUID, isAdmin bool) ([]FormCoverage, error) {
	if err := s.assertCourseOwner(ctx, courseID, userID, isAdmin); err != nil {
		return nil, err
	}
	blocks, err := s.q.ListBlocksByCourseID(ctx, courseID)
	if err != nil {
		return nil, err
	}
	type formKey struct {
		lexemeID string
		formID   string
	}
	lexemeForms := make(map[string]map[formKey]FormCoverageEntry)
	for _, b := range blocks {
		for _, ref := range domain.ExtractLexemeRefs(domain.BlockLexemeSource{
			BlockType:    b.BlockType,
			Config:       b.Config,
			DisplayLabel: b.DisplayLabel,
		}) {
			if ref.LexemeID == "" {
				continue
			}
			if _, ok := lexemeForms[ref.LexemeID]; !ok {
				lexemeForms[ref.LexemeID] = make(map[formKey]FormCoverageEntry)
			}
			key := formKey{lexemeID: ref.LexemeID, formID: ref.FormID}
			entry := lexemeForms[ref.LexemeID][key]
			entry.FormID = ref.FormID
			if ref.FormID != "" {
				entry.Form = ref.FormID
			}
			switch ref.UsageKind {
			case domain.UsageIntroduced:
				entry.Introduced = true
			case domain.UsageExercised:
				entry.Exercised = true
			}
			lexemeForms[ref.LexemeID][key] = entry
		}
	}
	out := make([]FormCoverage, 0, len(lexemeForms))
	for lexemeID, forms := range lexemeForms {
		fc := FormCoverage{LexemeID: lexemeID, Lemma: lexemeID}
		for _, f := range forms {
			fc.Forms = append(fc.Forms, f)
		}
		out = append(out, fc)
	}
	return out, nil
}

func (s *ContentService) GetStudentProgress(ctx context.Context, courseID, studentID, userID uuid.UUID, isAdmin bool) (StudentProgressView, error) {
	if err := s.assertCourseOwner(ctx, courseID, userID, isAdmin); err != nil {
		return StudentProgressView{}, err
	}
	return StudentProgressView{
		UserID:   studentID,
		CourseID: courseID,
		Lessons:  nil,
	}, nil
}
