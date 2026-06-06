package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// PgTx begins Postgres transactions (see cloudtraining service-template).
type PgTx interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}
