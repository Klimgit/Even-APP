package service

import (
	"context"

	"github.com/even-app/even-app/services/auth/internal/domain"
)

func (s *AuthService) ListDemoNotes(ctx context.Context) ([]domain.DemoNote, error) {
	rows, err := s.q.ListDemoNotes(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.DemoNote, 0, len(rows))
	for _, r := range rows {
		out = append(out, domain.DemoNote{
			ID:        r.ID,
			Text:      r.Text,
			CreatedAt: r.CreatedAt,
		})
	}
	return out, nil
}
