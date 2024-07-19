package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
)

type ProduceUsers struct {
	UserID           int       `json:"userId"`
	UserName         string    `json:"userName"`
	UserSurname      string    `json:"userSurname"`
	UserAge          int       `json:"userAge"`
	UserGender       string    `json:"userGender"`
	UserEmail        string    `json:"userEmail"`
	Country          string    `json:"country"`
	EngagementSource string    `json:"engagementSource"`
	Status           string    `json:"status"`
	Archived         bool      `json:"archived"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

func (u *ProduceUsers) Run(ctx context.Context) error {
	usersJSON, err := json.Marshal(u)
	if err != nil {
		return fmt.Errorf("failed to marshal users info: %w", err)
	}

	producer, err := sarama.NewSyncProducer([]string{"localhost:9094"}, nil)
	if err != nil {
		return fmt.Errorf("error creating sync producer: %w", err)
	}

	defer func() {
		if err := producer.Close(); err != nil {
			log.Printf("error closing producer: %v", err)

			return
		}
	}()

	t := time.NewTicker(time.Minute)
	defer t.Stop()

	for {
		msg := &sarama.ProducerMessage{
			Topic: "users_info",
			Value: sarama.StringEncoder(usersJSON),
		}

		if _, _, err := producer.SendMessage(msg); err != nil {
			return fmt.Errorf("error sending message: %w", err)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
		}
	}
}
