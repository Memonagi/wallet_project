package generateusers

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/Memonagi/wallet_project/internal/models"
)

type producer interface {
	ProduceUsers(key, value string) error
}

type Generator struct {
	producer producer
}

const (
	userIDMax       = 1000
	userIDMin       = 1
	userNameLen     = 7
	userSurnameLen  = 8
	userAgeMax      = 82
	userAgeMin      = 18
	firstBool       = 2
	secBool         = 0
	generatorTicker = 200 * time.Millisecond
)

func New(producer producer) *Generator {
	return &Generator{producer: producer}
}

func (g *Generator) Run(ctx context.Context) error {
	t := time.NewTicker(generatorTicker)
	defer t.Stop()

	for {
		users := generateInfo()

		usersJSON, err := json.Marshal(users)
		if err != nil {
			return fmt.Errorf("failed to marshal users: %w", err)
		}

		if err := g.producer.ProduceUsers("", string(usersJSON)); err != nil {
			return fmt.Errorf("failed to produce users: %w", err)
		}

		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
		}
	}
}

func generateInfo() *models.UserExternal {
	//nolint:gosec
	rand.New(rand.NewSource(time.Now().UnixNano()))
	//nolint:gosec
	createdAt := time.Now().Add(-time.Duration(rand.Intn(userIDMax)) * time.Hour)

	return &models.UserExternal{
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
