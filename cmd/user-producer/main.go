package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/Memonagi/wallet_project/internal/apiserver"
	"github.com/Memonagi/wallet_project/internal/producer"
	"github.com/sirupsen/logrus"
)

const producePort = 7540

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)
	defer cancel()

	users := producer.GenerateUsersInfo()

	if err := users.Run(ctx); err != nil {
		logrus.Panicf("Failed to run producer: %v", err)
	}

	server := apiserver.NewProducerAPI(producePort)

	if err := server.RunProducerAPI(ctx); err != nil {
		logrus.Panicf("Failed to start producer api server: %v", err)
	}
}
