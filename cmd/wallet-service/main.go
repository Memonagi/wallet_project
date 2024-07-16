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

	if _, err := database.New(ctx); err != nil {
		logrus.Panicf("failed to connect to database: %v", err)
	}

	server := handlers.New(port)

	if err := server.Run(ctx); err != nil {
		logrus.Panicf("failed to start server: %v", err)
	}
}
