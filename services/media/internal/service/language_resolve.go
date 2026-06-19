package service

import (
	"context"
	"fmt"

	"github.com/even-app/even-app/services/media/internal/gen/query"
	"github.com/google/uuid"
)

// resolveLanguageID maps a client language_id to a row in the media DB.
// Lexicon and other services may use different UUIDs for the same language code;
// when the ID is unknown here, fall back to evn so uploads still work in dev.
func (s *MediaService) resolveLanguageID(ctx context.Context, idStr string) (uuid.UUID, error) {
	if id, err := uuid.Parse(idStr); err == nil {
		ok, err := s.q.LanguageExists(ctx, query.LanguageExistsParams{ID: id})
		if err != nil {
			return uuid.Nil, err
		}
		if ok {
			return id, nil
		}
	}
	id, err := s.q.GetLanguageIDByCode(ctx, query.GetLanguageIDByCodeParams{Code: "evn"})
	if err != nil {
		return uuid.Nil, fmt.Errorf("language_id required")
	}
	return id, nil
}
