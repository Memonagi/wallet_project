package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
)

type Store struct {
	db  *pgxpool.Pool
	dsn string
}

//go:embed migrations
var migrations embed.FS

func New(ctx context.Context) (*Store, error) {
	dsn := "postgresql://user:password@localhost:5432/mydatabase"

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

func (s *Store) Migrate(direction migrate.MigrationDirection) error {
	db, err := sql.Open("pgx", s.dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			logrus.Panicf("failed to close database: %v", err)
		}
	}()

	assetDir := func() func(string) ([]string, error) {
		return func(path string) ([]string, error) {
			dirEntry, err := migrations.ReadDir(path)
			if err != nil {
				return nil, fmt.Errorf("failed to read directory: %w", err)
			}

			entries := make([]string, 0)

			for _, e := range dirEntry {
				entries = append(entries, e.Name())
			}

			return entries, nil
		}
	}()

	asset := migrate.AssetMigrationSource{
		Asset:    migrations.ReadFile,
		AssetDir: assetDir,
		Dir:      "migrations",
	}

	if _, err := migrate.Exec(db, "postgres", asset, direction); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}
