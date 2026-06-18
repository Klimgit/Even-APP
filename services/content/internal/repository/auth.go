package repository

import (
	"context"

	"github.com/even-app/even-app/services/content/internal/gen/authquery"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthReader struct {
	queries *authquery.Queries
}

func NewAuthReader(pool *pgxpool.Pool) *AuthReader {
	if pool == nil {
		return nil
	}
	return &AuthReader{queries: authquery.New(pool)}
}

func (r *AuthReader) Available() bool {
	return r != nil && r.queries != nil
}

func (r *AuthReader) ListUsersByIDs(ctx context.Context, ids []uuid.UUID) ([]authquery.ListUsersByIDsRow, error) {
	return r.queries.ListUsersByIDs(ctx, authquery.ListUsersByIDsParams{Column1: ids})
}

func (r *AuthReader) GetUserByEmail(ctx context.Context, email string) (authquery.GetUserByEmailRow, error) {
	return r.queries.GetUserByEmail(ctx, authquery.GetUserByEmailParams{Email: email})
}
