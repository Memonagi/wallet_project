package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
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

const (
	userIDMax      = 1000
	userIDMin      = 1
	userNameLen    = 7
	userSurnameLen = 8
	userAgeMax     = 82
	userAgeMin     = 18
	firstBool      = 2
	secBool        = 0
)

func GenerateUsersInfo() *ProduceUsers {
	//nolint:gosec
	rand.New(rand.NewSource(time.Now().UnixNano()))
	//nolint:gosec
	createdAt := time.Now().Add(-time.Duration(rand.Intn(userIDMax)) * time.Hour)

	return &ProduceUsers{
		//nolint:gosec
		UserID:      rand.Intn(userIDMax) + userIDMin,
		UserName:    randomString(userNameLen),
		UserSurname: randomString(userSurnameLen),
		//nolint:gosec
		UserAge:          rand.Intn(userAgeMax) + userAgeMin,
		UserGender:       randomGender(),
		UserEmail:        randomString(userNameLen) + "@" + randomString(userNameLen) + ".com",
		Country:          randomCountry(),
		EngagementSource: randomSource(),
		Status:           randomStatus(),
		//nolint:gosec
		Archived:  rand.Intn(firstBool) == secBool,
		CreatedAt: createdAt,
		//nolint:gosec
		UpdatedAt: createdAt.Add(-time.Duration(rand.Intn(userIDMax)) * time.Hour),
	}
}

func randomString(length int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	b := make([]byte, length)

	for i := range b {
		//nolint:gosec
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}

func randomGender() string {
	genders := []string{"male", "female"}

	//nolint:gosec
	return genders[rand.Intn(len(genders))]
}

func randomCountry() string {
	countries := []string{"Russia", "Belarus", "Kazakhstan", "Uzbekistan"}

	//nolint:gosec
	return countries[rand.Intn(len(countries))]
}

func randomSource() string {
	sources := []string{"advertising", "website", "referral", "mail"}

	//nolint:gosec
	return sources[rand.Intn(len(sources))]
}

func randomStatus() string {
	statuses := []string{"active", "inactive"}

	//nolint:gosec
	return statuses[rand.Intn(len(statuses))]
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
	defer producer.Close()

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
