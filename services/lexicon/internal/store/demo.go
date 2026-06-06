package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DemoStore struct {
	pool *pgxpool.Pool
}

func NewDemoStore(pool *pgxpool.Pool) *DemoStore {
	return &DemoStore{pool: pool}
}

func (s *DemoStore) Ping(ctx context.Context) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `SELECT 1`).Scan(&n)
	return n, err
}
