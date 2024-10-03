package producer

import (
	"fmt"

	"github.com/IBM/sarama"
)

type Producer struct {
	producer sarama.SyncProducer
}

func New(address string) (*Producer, error) {
	producer, err := sarama.NewSyncProducer([]string{address}, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating sync producer: %w", err)
	}

	return &Producer{producer: producer}, nil
}

func (p *Producer) Close() error {
	if err := p.producer.Close(); err != nil {
		return fmt.Errorf("error closing producer: %w", err)
	}

	return nil
}

func (p *Producer) produceMessage(topic, key, message string) error {
	if _, _, err := p.producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.StringEncoder(message),
	}); err != nil {
		return fmt.Errorf("error sending message: %w", err)
	}

	return nil
}

func (p *Producer) ProduceUsers(key, value string) error {
	return p.produceMessage("user_updates", key, value)
}

func (p *Producer) ProduceTx(key, value string) error {
	return p.produceMessage("transaction_updates", key, value)
}
