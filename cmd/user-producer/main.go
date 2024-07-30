package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Memonagi/wallet_project/internal/generateusers"
	"github.com/Memonagi/wallet_project/internal/producer"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)
	defer cancel()

	kafkaHost := os.Getenv("KAFKA_HOST")
	if kafkaHost == "" {
		kafkaHost = "localhost"
	}

	kafkaProducer, err := producer.New(fmt.Sprintf("%s:9094", kafkaHost))
	if err != nil {
		logrus.Panicf("Failed to create producer: %v", err)
	}

	defer func() {
		if err := kafkaProducer.Close(); err != nil {
			logrus.Warnf("Failed to close producer: %v", err)
		}
	}()

	generator := generateusers.New(kafkaProducer)

	if err := generator.Run(ctx); err != nil {
		logrus.Panicf("Failed to run generator: %v", err)
	}
}
