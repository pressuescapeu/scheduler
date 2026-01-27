package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pool *pgxpool.Pool
}

// NewConnection returns *Storage so the pool is shared
func NewConnection(connString string) (*Storage, error) {
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, err
	}

	return &Storage{
		pool: pool,
	}, nil
}

// Close closes the database connection pool
func (s *Storage) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}
