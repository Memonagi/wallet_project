package database

import (
	"context"
	"fmt"

	_ "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	db  *pgxpool.Pool
	dsn string
}

func New(ctx context.Context) (*DB, error) {
	dsn := "postgresql://user:password@localhost:5432/dbname"

	db, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{
		db:  db,
		dsn: dsn,
	}, nil
}
