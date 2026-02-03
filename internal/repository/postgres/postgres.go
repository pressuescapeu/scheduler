package postgres

import (
	"context"
	_ "embed"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/schema.sql
var schemaSQL string

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

// RunMigrations creates all tables if they don't exist
func (s *Storage) RunMigrations(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, schemaSQL)
	return err
}

// ResetCourseData wipes course-related data to allow reseeding
func (s *Storage) ResetCourseData(ctx context.Context) error {
	const query = `
		TRUNCATE schedule_sections,
		         section_meetings,
		         sections,
		         courses,
		         professors
		RESTART IDENTITY CASCADE;
	`
	_, err := s.pool.Exec(ctx, query)
	return err
}

// Close closes the database connection pool
func (s *Storage) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}
