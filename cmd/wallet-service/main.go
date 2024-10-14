package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/Memonagi/wallet_project/internal/config"
	"github.com/Memonagi/wallet_project/internal/consumer"
	"github.com/Memonagi/wallet_project/internal/database"
	jwtclaims "github.com/Memonagi/wallet_project/internal/jwt-claims"
	"github.com/Memonagi/wallet_project/internal/producer"
	"github.com/Memonagi/wallet_project/internal/server"
	"github.com/Memonagi/wallet_project/internal/service"
	xrclient "github.com/Memonagi/wallet_project/internal/xr/xr-client"
	_ "github.com/jackc/pgx/v5/stdlib"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)
	defer cancel()

	cfg := config.New()

	db, err := database.New(ctx, database.Config{Dsn: cfg.GetPostgresDSN()})
	if err != nil {
		logrus.Panicf("failed to connect to database: %v", err)
	}

	if err = db.Migrate(migrate.Up); err != nil {
		logrus.Panicf("failed to migrate users: %v", err)
	}

	logrus.Info("migrated successfully")

	kafkaConsumer, err := consumer.New(db, consumer.Config{Port: cfg.GetKafkaPort()})
	if err != nil {
		logrus.Panicf("failed to connect to consumer: %v", err)
	}

	defer func() {
		if err = kafkaConsumer.Close(); err != nil {
			logrus.Warnf("failed to close consumer: %v", err)
		}
	}()

	txProducer, err := producer.New(producer.Config{Address: cfg.GetKafkaPort()})
	if err != nil {
		logrus.Panicf("Failed to create producer: %v", err)
	}

	defer func() {
		if err = txProducer.Close(); err != nil {
			logrus.Warnf("Failed to close producer: %v", err)
		}
	}()

	client := xrclient.New(xrclient.Config{ServerAddress: cfg.GetXRServerAddress()})
	svc := service.New(db, client, txProducer)
	jwtClaims := jwtclaims.New()
	httpServer := server.New(server.Config{Port: cfg.GetAppPort()}, svc, jwtClaims.GetPublicKey())

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		err := kafkaConsumer.Run(ctx)

		return fmt.Errorf("kafka consumer stopped: %w", err)
	})

	eg.Go(func() error {
		err := httpServer.Run(ctx)

		return fmt.Errorf("server stopped: %w", err)
	})

	if err = eg.Wait(); err != nil {
		logrus.Panicf("eg.Wait(): %v", err)
	}
}
