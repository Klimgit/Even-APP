package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	libjwt "github.com/even-app/even-app/libs/jwt"
	"github.com/even-app/even-app/services/auth/internal/domain"
	"github.com/even-app/even-app/services/auth/internal/gen/query"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	q          *query.Queries
	jwt        *libjwt.Manager
	refreshTTL time.Duration
}

func NewAuthService(q *query.Queries, jwt *libjwt.Manager, refreshTTL time.Duration) *AuthService {
	return &AuthService{q: q, jwt: jwt, refreshTTL: refreshTTL}
}

func (s *AuthService) Register(ctx context.Context, email, password, role string, displayName *string) (*domain.AuthOutcome, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("hash failed")
	}
	if role != "student" && role != "teacher" {
		role = "student"
	}
	row, err := s.q.CreateUser(ctx, query.CreateUserParams{
		Email: email, PasswordHash: string(hash), DisplayName: displayName, Role: role,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, domain.ErrConflict
		}
		return nil, err
	}
	u := mapUser(row)
	tokens, err := s.issueTokens(ctx, &u)
	if err != nil {
		return nil, err
	}
	return &domain.AuthOutcome{Tokens: tokens, User: &u}, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*domain.AuthOutcome, error) {
	row, err := s.q.GetUserByEmail(ctx, query.GetUserByEmailParams{Email: email})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUnauthorized
		}
		return nil, err
	}
	u := mapUser(row)
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, domain.ErrUnauthorized
	}
	tokens, err := s.issueTokens(ctx, &u)
	if err != nil {
		return nil, err
	}
	return &domain.AuthOutcome{Tokens: tokens, User: &u}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*domain.Tokens, error) {
	hash := hashToken(refreshToken)
	userID, err := s.q.GetUserIDByRefreshHash(ctx, query.GetUserIDByRefreshHashParams{TokenHash: hash})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUnauthorized
		}
		return nil, err
	}
	_ = s.q.DeleteRefreshToken(ctx, query.DeleteRefreshTokenParams{TokenHash: hash})
	row, err := s.q.GetUserByID(ctx, query.GetUserByIDParams{ID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUnauthorized
		}
		return nil, err
	}
	u := mapUser(row)
	tokens, err := s.issueTokens(ctx, &u)
	if err != nil {
		return nil, err
	}
	return &tokens, nil
}

func (s *AuthService) Me(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
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

func (s *AuthService) issueTokens(ctx context.Context, u *domain.User) (domain.Tokens, error) {
	access, err := s.jwt.IssueAccess(u.ID, u.Role, u.IsAdmin)
	if err != nil {
		return domain.Tokens{}, err
	}
	refresh, err := newRefreshToken()
	if err != nil {
		return domain.Tokens{}, err
	}
	if err := s.q.SaveRefreshToken(ctx, query.SaveRefreshTokenParams{
		UserID: u.ID, TokenHash: hashToken(refresh), ExpiresAt: time.Now().Add(s.refreshTTL),
	}); err != nil {
		return domain.Tokens{}, err
	}
	return domain.Tokens{AccessToken: access, RefreshToken: refresh}, nil
}

func mapUser(row query.User) domain.User {
	return domain.User{
		ID: row.ID, Email: row.Email, PasswordHash: row.PasswordHash,
		DisplayName: row.DisplayName, Role: row.Role, IsAdmin: row.IsAdmin,
		CreatedAt: row.CreatedAt,
	}
}

func isUniqueViolation(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique")
}

func newRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
