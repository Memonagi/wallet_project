package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/Memonagi/wallet_project/internal/database"
	"github.com/Memonagi/wallet_project/internal/handlers"
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

	if err := db.MigrateUsers(ctx); err != nil {
		logrus.Panicf("failed to migrate users: %v", err)
	}

	server := handlers.New(port)

	if err := server.Run(ctx); err != nil {
		logrus.Panicf("failed to start server: %v", err)
	}
}
