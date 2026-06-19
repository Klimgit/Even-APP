package service

import (
	"context"

	"github.com/even-app/even-app/services/content/internal/domain"
	"github.com/even-app/even-app/services/content/internal/gen/query"
	"github.com/google/uuid"
)

func (s *ContentService) ListFavoriteBlockTypes(ctx context.Context, userID uuid.UUID) (map[string]struct{}, error) {
	rows, err := s.q.ListFavoriteBlockTypes(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make(map[string]struct{}, len(rows))
	for _, blockType := range rows {
		out[blockType] = struct{}{}
	}
	return out, nil
}

func (s *ContentService) AddFavoriteBlockType(ctx context.Context, userID uuid.UUID, blockType string) error {
	if !domain.IsKnownBlockType(blockType) {
		return domain.ErrValidation
	}
	return s.q.AddFavoriteBlockType(ctx, query.AddFavoriteBlockTypeParams{
		UserID: userID, BlockType: blockType,
	})
}

func (s *ContentService) RemoveFavoriteBlockType(ctx context.Context, userID uuid.UUID, blockType string) error {
	return s.q.RemoveFavoriteBlockType(ctx, query.RemoveFavoriteBlockTypeParams{
		UserID: userID, BlockType: blockType,
	})
}
