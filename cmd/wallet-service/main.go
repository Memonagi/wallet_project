package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/Memonagi/wallet_project/internal/consumer"
	"github.com/Memonagi/wallet_project/internal/database"
	"github.com/Memonagi/wallet_project/internal/server"
	"github.com/Memonagi/wallet_project/internal/service"
	_ "github.com/jackc/pgx/v5/stdlib"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

const (
	port = 8080
	dsn  = "postgresql://user:password@localhost:5432/mydatabase"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)
	defer cancel()

	db, err := database.New(ctx, dsn)
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

	svc := service.New(db)
	httpServer := server.New(port, svc)

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		err = kafkaConsumer.Run(ctx)

		return fmt.Errorf("kafka consumer stopped: %w", err)
	})

	eg.Go(func() error {
		err = httpServer.Run(ctx)

		return fmt.Errorf("server stopped: %w", err)
	})

	if err := eg.Wait(); err != nil {
		logrus.Panicf("eg.Wait(): %v", err)
	}
}
