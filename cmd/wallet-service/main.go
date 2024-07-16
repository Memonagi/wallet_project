package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/Memonagi/wallet_project/internal/database"
)

const dbFile = "wallet.db"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)
	defer cancel()

	db, err := database.New(ctx, dbFile)
	if err != nil {
		logrus.Panicf("failed to connect to database: %w", err)
	}
	defer func() {
		if err := db.CloseDB(); err != nil {
			logrus.Panicf("failed to close database: %w", err)
		}
	}()
}
