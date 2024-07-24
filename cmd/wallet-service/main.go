package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/Memonagi/wallet_project/internal/consumer"
	"github.com/Memonagi/wallet_project/internal/database"
	"github.com/Memonagi/wallet_project/internal/server"
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

	kafkaConsumer, err := consumer.New(db, "localhost:9094")
	if err != nil {
		logrus.Panicf("failed to connect to consumer: %v", err)
	}

	defer func() {
		if err := kafkaConsumer.Close(); err != nil {
			logrus.Warnf("failed to close consumer: %v", err)
		}
	}()

	if err := kafkaConsumer.Run(ctx); err != nil {
		logrus.Panicf("failed to start consumer: %v", err)
	}

	newServer := server.New(port)

	if err := newServer.Run(ctx); err != nil {
		logrus.Panicf("failed to start server: %v", err)
	}
}
