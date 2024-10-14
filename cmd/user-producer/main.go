package main

import (
	"context"
	"os/signal"
	"syscall"

	generateusers "github.com/Memonagi/wallet_project/internal/generate-users"
	"github.com/Memonagi/wallet_project/internal/producer"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)
	defer cancel()

	kafkaProducer, err := producer.New(producer.Config{Address: "localhost:9094"})
	if err != nil {
		logrus.Panicf("Failed to create producer: %v", err)
	}

	defer func() {
		if err = kafkaProducer.Close(); err != nil {
			logrus.Warnf("Failed to close producer: %v", err)
		}
	}()

	generator := generateusers.New(kafkaProducer)

	if err = generator.Run(ctx); err != nil {
		logrus.Panicf("Failed to run generator: %v", err)
	}
}
