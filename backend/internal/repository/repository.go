package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Ping verifies the database connection. Domain query methods hang off this
// struct as feature work lands.
func (r *Repository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

// BeginTx starts a new transaction.
func (r *Repository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}

func isNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
