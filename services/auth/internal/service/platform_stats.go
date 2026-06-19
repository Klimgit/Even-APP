package service

import (
	"context"

	"github.com/even-app/even-app/services/auth/internal/domain"
)

type PlatformStats struct {
	TotalUsers        int
	Students          int
	Teachers          int
	Admins            int
	PublishedCourses  int
	ActiveEnrollments int
	TotalCourses      int
	PlatformLexemes   int
}

func (s *AuthService) GetPlatformStats(ctx context.Context, isAdmin bool) (*PlatformStats, error) {
	if !isAdmin {
		return nil, domain.ErrForbidden
	}
	row, err := s.q.UserStats(ctx)
	if err != nil {
		return nil, err
	}
	out := &PlatformStats{
		TotalUsers: int(row.TotalUsers),
		Students:   int(row.Students),
		Teachers:   int(row.Teachers),
		Admins:     int(row.Admins),
	}
	if s.stats != nil {
		if n, err := s.stats.PublishedCourses(ctx); err == nil {
			out.PublishedCourses = n
		}
		if n, err := s.stats.ActiveEnrollments(ctx); err == nil {
			out.ActiveEnrollments = n
		}
		if n, err := s.stats.TotalCourses(ctx); err == nil {
			out.TotalCourses = n
		}
		if n, err := s.stats.PlatformLexemes(ctx); err == nil {
			out.PlatformLexemes = n
		}
	}
	return out, nil
}
