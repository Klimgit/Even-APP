package service

import (
	"context"

	"github.com/even-app/even-app/services/lexicon/internal/domain"
	"github.com/even-app/even-app/services/lexicon/internal/gen/query"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *LexiconService) GetLexemeUsage(ctx context.Context, lexemeID, courseID uuid.UUID) (*domain.LexemeUsageResult, error) {
	if _, err := s.q.GetLexeme(ctx, query.GetLexemeParams{ID: lexemeID}); err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	_ = courseID
	return &domain.LexemeUsageResult{
		LexemeID: lexemeID,
		Usages:   []domain.LexemeUsageEntry{},
	}, nil
}
