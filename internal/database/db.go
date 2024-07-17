package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type Store struct {
	db  *pgxpool.Pool
	dsn string
}

func New(ctx context.Context) (*Store, error) {
	dsn := "postgresql://user:password@localhost:5432/dbname"

	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logrus.Info("connected to database")

	return &Store{
		db:  db,
		dsn: dsn,
	}, nil
}

func (s *Store) MigrateUsers(ctx context.Context) error {
	createUserTable := `CREATE TABLE IF NOT EXISTS users (
    user_id UUID NOT NULL UNIQUE PRIMARY KEY,
    status VARCHAR(128) NOT NULL,
    archived BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);`

	if _, err := s.db.Exec(ctx, createUserTable); err != nil {
		return fmt.Errorf("failed to create user table: %w", err)
	}

	logrus.Info("migrated users")

	return nil
}
