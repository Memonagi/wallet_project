//nolint:gosec
package usersgenerator

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/Memonagi/wallet_project/internal/models"
)

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

type producer interface {
	ProduceMessage(topic string, key, message string) error
}

type Generator struct {
	producer producer
}

func New(producer producer) *Generator {
	return &Generator{
		producer: producer,
	}
}

const (
	usersTopic      = "user_updates"
	generatorPeriod = 200 * time.Millisecond
)

func (g *Generator) Run(ctx context.Context) error {
	t := time.NewTicker(generatorPeriod)
	defer t.Stop()

	for {
		user := generateInfo()

		val, err := json.Marshal(user)
		if err != nil {
			return fmt.Errorf("failed to marshal user: %w", err)
		}

		if err = g.producer.ProduceMessage(usersTopic, "", string(val)); err != nil {
			return fmt.Errorf("failed to produce user: %w", err)
		}

		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
		}
	}
}

func generateInfo() *models.UserExternal {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	createdAt := time.Now().Add(-time.Duration(rand.Intn(userIDMax)) * time.Hour)

	return &models.UserExternal{
		UserID:           rand.Intn(userIDMax) + userIDMin,
		UserName:         randomString(userNameLen),
		UserSurname:      randomString(userSurnameLen),
		UserAge:          rand.Intn(userAgeMax) + userAgeMin,
		UserGender:       randomGender(),
		UserEmail:        randomString(userNameLen) + "@" + randomString(userNameLen) + ".com",
		Country:          randomCountry(),
		EngagementSource: randomSource(),
		Status:           randomStatus(),
		Archived:         rand.Intn(firstBool) == secBool,
		CreatedAt:        createdAt,
		UpdatedAt:        createdAt.Add(-time.Duration(rand.Intn(userIDMax)) * time.Hour),
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
