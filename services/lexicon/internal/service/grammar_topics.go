package service

import (
	"context"
	"errors"

	"github.com/even-app/even-app/services/lexicon/internal/domain"
	"github.com/even-app/even-app/services/lexicon/internal/gen/query"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *LexiconService) ListGrammarTopics(ctx context.Context, code string) ([]query.GrammarTopic, error) {
	if _, err := s.languageIDByCode(ctx, code); err != nil {
		return nil, err
	}
	return s.q.ListGrammarTopicsByLanguageCode(ctx, query.ListGrammarTopicsByLanguageCodeParams{Code: code})
}

func (s *LexiconService) CreateGrammarTopic(ctx context.Context, code, title, description string, sortOrder int32) (query.GrammarTopic, error) {
	langID, err := s.languageIDByCode(ctx, code)
	if err != nil {
		return query.GrammarTopic{}, err
	}
	return s.q.CreateGrammarTopic(ctx, query.CreateGrammarTopicParams{
		LanguageID:  langID,
		Title:       title,
		Description: description,
		SortOrder:   sortOrder,
	})
}

func (s *LexiconService) PatchGrammarTopic(ctx context.Context, topicID uuid.UUID, title, description *string, sortOrder *int32) (query.GrammarTopic, error) {
	row, err := s.q.PatchGrammarTopic(ctx, query.PatchGrammarTopicParams{
		ID:          topicID,
		Title:       title,
		Description: description,
		SortOrder:   sortOrder,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return query.GrammarTopic{}, domain.ErrNotFound
		}
		return query.GrammarTopic{}, err
	}
	return row, nil
}

func (s *LexiconService) DeleteGrammarTopic(ctx context.Context, topicID uuid.UUID) error {
	if err := s.q.DeleteGrammarTopic(ctx, query.DeleteGrammarTopicParams{ID: topicID}); err != nil {
		return err
	}
	return nil
}
