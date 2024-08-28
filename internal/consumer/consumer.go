package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/Memonagi/wallet_project/internal/models"
	"github.com/sirupsen/logrus"
)

type infoSaver interface {
	UpsertUser(ctx context.Context, users models.User) error
}

type Consumer struct {
	infoSaver infoSaver
	consumer  sarama.Consumer
}

type Config struct {
	Port string
}

func New(infoSaver infoSaver, cfg Config) (*Consumer, error) {
	consumer, err := sarama.NewConsumer([]string{cfg.Port}, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating new consumer: %w", err)
	}

	return &Consumer{
		infoSaver: infoSaver,
		consumer:  consumer,
	}, nil
}

func (c *Consumer) Run(ctx context.Context) error {
	partConsumer, err := c.consumeUsers()
	if err != nil {
		return fmt.Errorf("error consuming users: %w", err)
	}

	defer func() {
		if err := partConsumer.Close(); err != nil {
			logrus.Warnf("error closing consumer: %v", err)

			return
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg := <-partConsumer.Messages():
			var users models.User

			if err := json.Unmarshal(msg.Value, &users); err != nil {
				return fmt.Errorf("error unmarshalling users: %w", err)
			}

			if err := c.infoSaver.UpsertUser(ctx, users); err != nil {
				return fmt.Errorf("error upserting users: %w", err)
			}
		case err := <-partConsumer.Errors():
			if err != nil {
				return fmt.Errorf("error consuming users: %w", err)
			}
		}
	}
}

func (c *Consumer) Close() error {
	if err := c.consumer.Close(); err != nil {
		return fmt.Errorf("error closing consumer: %w", err)
	}

	return nil
}

func (c *Consumer) consumeMessage(topic string) (sarama.PartitionConsumer, error) {
	partConsumer, err := c.consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		return nil, fmt.Errorf("error consuming message from topic %s: %w", topic, err)
	}

	return partConsumer, nil
}

func (c *Consumer) consumeUsers() (sarama.PartitionConsumer, error) {
	return c.consumeMessage("user_updates")
}
