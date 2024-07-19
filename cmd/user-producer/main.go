package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/Memonagi/wallet_project/internal/generateusers"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)
	defer cancel()

	users := generateusers.GenerateInfo()

	if err := users.Run(ctx); err != nil {
		logrus.Panicf("Failed to run producer: %v", err)
	}
}
