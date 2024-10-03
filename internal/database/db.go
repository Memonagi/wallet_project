package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/Memonagi/wallet_project/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
)

type Store struct {
	db       *pgxpool.Pool
	dsn      string
	producer txProducer
}

type Config struct {
	Dsn string
}

//go:embed migrations
var migrations embed.FS

func New(ctx context.Context, cfg Config, producer txProducer) (*Store, error) {
	db, err := pgxpool.New(ctx, cfg.Dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logrus.Info("connected to database")

	return &Store{
		db:       db,
		dsn:      cfg.Dsn,
		producer: producer,
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

func (s *Store) UpsertUser(ctx context.Context, users models.User) error {
	query := `INSERT INTO users (id, status, archived, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (id) DO UPDATE SET 
    status = excluded.status, 
    archived = excluded.archived,
    updated_at = NOW()`

	_, err := s.db.Exec(ctx, query, users.UserID, users.Status, users.Archived, users.CreatedAt, users.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to upsert users: %w", err)
	}

	return nil
}

func (s *Store) Truncate(ctx context.Context, tables ...string) error {
	for _, table := range tables {
		if _, err := s.db.Exec(ctx, "DELETE FROM "+table); err != nil {
			return fmt.Errorf("failed to truncate table: %w", err)
		}
	}

	return nil
}
