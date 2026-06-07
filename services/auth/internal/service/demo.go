package service

import (
	"context"

	"github.com/even-app/even-app/services/auth/internal/domain"
	"github.com/google/uuid"
)

type DemoStats struct {
	TotalUsers int
	Students   int
	Teachers   int
	Admins     int
}

func (s *AuthService) DemoPublic(ctx context.Context) (int, error) {
	count, err := s.q.CountUsers(ctx)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (s *AuthService) DemoMe(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	return s.Me(ctx, userID)
}

func (s *AuthService) DemoTeacher(ctx context.Context, userID uuid.UUID, role string, isAdmin bool) (*domain.User, error) {
	if role != "teacher" && !isAdmin {
		return nil, domain.ErrForbidden
	}
	return s.Me(ctx, userID)
}

func (s *AuthService) DemoAdminStats(ctx context.Context, isAdmin bool) (*DemoStats, error) {
	if !isAdmin {
		return nil, domain.ErrForbidden
	}
	row, err := s.q.UserStats(ctx)
	if err != nil {
		return nil, err
	}
	return &DemoStats{
		TotalUsers: int(row.TotalUsers),
		Students:   int(row.Students),
		Teachers:   int(row.Teachers),
		Admins:     int(row.Admins),
	}, nil
}
