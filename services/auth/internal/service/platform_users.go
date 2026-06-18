package service

import (
	"context"

	"github.com/even-app/even-app/services/auth/internal/domain"
	"github.com/even-app/even-app/services/auth/internal/gen/query"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserListResult struct {
	Items []domain.User
	Total int
}

type PatchPlatformUserInput struct {
	Role    *string
	IsAdmin *bool
}

func (s *AuthService) ListPlatformUsers(ctx context.Context, isAdmin bool, q, role string, page, limit int) (*UserListResult, error) {
	if !isAdmin {
		return nil, domain.ErrForbidden
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit
	total, err := s.q.CountUsersFiltered(ctx, query.CountUsersFilteredParams{
		Column1: q,
		Column2: role,
	})
	if err != nil {
		return nil, err
	}
	rows, err := s.q.ListUsers(ctx, query.ListUsersParams{
		Column1: q,
		Column2: role,
		Offset:  int32(offset),
		Limit:   int32(limit),
	})
	if err != nil {
		return nil, err
	}
	items := make([]domain.User, 0, len(rows))
	for _, r := range rows {
		items = append(items, mapUser(r))
	}
	return &UserListResult{Items: items, Total: int(total)}, nil
}

func (s *AuthService) PatchPlatformUser(ctx context.Context, isAdmin bool, userID uuid.UUID, in PatchPlatformUserInput) (*domain.User, error) {
	if !isAdmin {
		return nil, domain.ErrForbidden
	}
	if in.Role != nil && *in.Role != "student" && *in.Role != "teacher" {
		return nil, domain.ErrValidation
	}
	row, err := s.q.UpdateUserPlatform(ctx, query.UpdateUserPlatformParams{
		ID:      userID,
		Role:    in.Role,
		IsAdmin: in.IsAdmin,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	u := mapUser(row)
	return &u, nil
}
