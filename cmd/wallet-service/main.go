package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/Memonagi/wallet_project/internal/database"
	"github.com/Memonagi/wallet_project/internal/handlers"
	_ "github.com/jackc/pgx/v5/stdlib"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
)

const port = 8080

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)
	defer cancel()

	db, err := database.New(ctx)
	if err != nil {
		logrus.Panicf("failed to connect to database: %v", err)
	}

	if err := db.Migrate(migrate.Up); err != nil {
		logrus.Panicf("failed to migrate users: %v", err)
	}

	logrus.Info("migrated successfully")

	server := handlers.New(port)

	if err := server.Run(ctx); err != nil {
		logrus.Panicf("failed to start server: %v", err)
	}
}
