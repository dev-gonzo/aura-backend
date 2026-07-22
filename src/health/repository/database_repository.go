package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DatabaseStatusRepository struct {
	pool *pgxpool.Pool
}

func NewDatabaseStatusRepository(pool *pgxpool.Pool) *DatabaseStatusRepository {
	return &DatabaseStatusRepository{pool: pool}
}

func (r *DatabaseStatusRepository) Check(ctx context.Context) string {
	if err := r.pool.Ping(ctx); err != nil {
		return "offline"
	}

	return "online"
}
