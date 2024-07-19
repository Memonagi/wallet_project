package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/Memonagi/wallet_project/internal/producer"
	usersgenerator "github.com/Memonagi/wallet_project/internal/users-generator"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)
	defer cancel()

	kafkaProducer, err := producer.New("localhost:9094")
	if err != nil {
		logrus.Panic(err)
	}

	generator := usersgenerator.New(kafkaProducer)

	if err = generator.Run(ctx); err != nil {
		logrus.Panic(err)
	}
}
