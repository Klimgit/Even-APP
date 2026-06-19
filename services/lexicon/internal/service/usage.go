package service

import (
	"context"
	"errors"

	"github.com/even-app/even-app/services/lexicon/internal/domain"
	"github.com/even-app/even-app/services/lexicon/internal/gen/query"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *LexiconService) GetLexemeUsage(ctx context.Context, lexemeID, courseID uuid.UUID) (*domain.LexemeUsageResult, error) {
	if _, err := s.q.GetLexeme(ctx, query.GetLexemeParams{ID: lexemeID}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	result := &domain.LexemeUsageResult{
		LexemeID: lexemeID,
		Usages:   []domain.LexemeUsageEntry{},
	}
	if s.content == nil || !s.content.Available() {
		return result, nil
	}

	blocks, err := s.content.ListBlocksByCourseID(ctx, courseID)
	if err != nil {
		return nil, err
	}

	lexemeStr := lexemeID.String()
	for _, b := range blocks {
		for _, ref := range domain.ExtractLexemeRefs(domain.BlockLexemeSource{
			BlockType:    b.BlockType,
			Config:       b.Config,
			DisplayLabel: b.DisplayLabel,
		}) {
			if ref.LexemeID != lexemeStr {
				continue
			}
			label := ref.BlockDisplayLabel
			if label == "" && b.DisplayLabel != nil {
				label = *b.DisplayLabel
			}
			result.Usages = append(result.Usages, domain.LexemeUsageEntry{
				LessonID:     b.LessonID,
				LessonTitle:  b.LessonTitle,
				BlockID:      b.ID,
				DisplayLabel: label,
				UsageKind:    string(ref.UsageKind),
			})
		}
	}
	return result, nil
}
