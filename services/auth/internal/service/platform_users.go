package service

import (
	"context"
	"errors"

	"github.com/even-app/even-app/services/auth/internal/domain"
	"github.com/even-app/even-app/services/auth/internal/gen/query"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserListResult struct {
	Items []domain.User
	Total int
}

type PatchPlatformUserInput struct {
	Role    *string
	IsAdmin *bool
}

type CreatePlatformUserInput struct {
	Email       string
	Password    string
	DisplayName *string
	Role        string
	IsAdmin     bool
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

func (s *AuthService) PatchPlatformUser(ctx context.Context, isAdmin bool, actorID, userID uuid.UUID, in PatchPlatformUserInput) (*domain.User, error) {
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	u := mapUser(row)
	details := map[string]any{}
	if in.Role != nil {
		details["role"] = *in.Role
	}
	if in.IsAdmin != nil {
		details["is_admin"] = *in.IsAdmin
	}
	_ = s.LogAudit(ctx, &actorID, "user.update", "user", &u.ID, details)
	return &u, nil
}

func (s *AuthService) GetPlatformUser(ctx context.Context, isAdmin bool, userID uuid.UUID) (*domain.User, error) {
	if !isAdmin {
		return nil, domain.ErrForbidden
	}
	row, err := s.q.GetUserByID(ctx, query.GetUserByIDParams{ID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	u := mapUser(row)
	return &u, nil
}

func (s *AuthService) CreatePlatformUser(ctx context.Context, isAdmin bool, actorID uuid.UUID, in CreatePlatformUserInput) (*domain.User, error) {
	if !isAdmin {
		return nil, domain.ErrForbidden
	}
	if in.Role != "student" && in.Role != "teacher" {
		return nil, domain.ErrValidation
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	row, err := s.q.CreateUser(ctx, query.CreateUserParams{
		Email: in.Email, PasswordHash: string(hash), DisplayName: in.DisplayName, Role: in.Role,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, domain.ErrConflict
		}
		return nil, err
	}
	if in.IsAdmin {
		row, err = s.q.UpdateUserPlatform(ctx, query.UpdateUserPlatformParams{
			ID: row.ID, IsAdmin: &in.IsAdmin,
		})
		if err != nil {
			return nil, err
		}
	}
	u := mapUser(row)
	_ = s.LogAudit(ctx, &actorID, "user.create", "user", &u.ID, map[string]any{
		"email": u.Email, "role": u.Role, "is_admin": u.IsAdmin,
	})
	return &u, nil
}

func (s *AuthService) DeletePlatformUser(ctx context.Context, isAdmin bool, actorID, userID uuid.UUID) error {
	if !isAdmin {
		return domain.ErrForbidden
	}
	if actorID == userID {
		return domain.ErrValidation
	}
	if _, err := s.q.GetUserByID(ctx, query.GetUserByIDParams{ID: userID}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	}
	if err := s.q.DeleteUser(ctx, query.DeleteUserParams{ID: userID}); err != nil {
		return err
	}
	_ = s.LogAudit(ctx, &actorID, "user.delete", "user", &userID, nil)
	return nil
}

func (s *AuthService) ResetPlatformUserPassword(ctx context.Context, isAdmin bool, actorID, userID uuid.UUID, password string) error {
	if !isAdmin {
		return domain.ErrForbidden
	}
	if len(password) < 8 {
		return domain.ErrValidation
	}
	if _, err := s.q.GetUserByID(ctx, query.GetUserByIDParams{ID: userID}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.q.UpdateUserPassword(ctx, query.UpdateUserPasswordParams{
		ID: userID, PasswordHash: string(hash),
	}); err != nil {
		return err
	}
	_ = s.LogAudit(ctx, &actorID, "user.reset_password", "user", &userID, nil)
	return nil
}
