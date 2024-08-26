package jwtclaims

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID uuid.UUID `json:"userId"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

const hours = 24

var (
	ErrInvalidToken         = errors.New("invalid token")
	ErrInvalidSigningMethod = errors.New("invalid signing method")
)

func New() *Claims {
	tokenTime := time.Now().Add(hours * time.Hour)

	return &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(tokenTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
}

func (c *Claims) GenerateToken(secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)

	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenStr, nil
}

func (c *Claims) ValidateToken(tokenStr, secret string) error {
	token, err := jwt.ParseWithClaims(tokenStr, c, func(token *jwt.Token) (interface{}, error) {
		if method, ok := token.Method.(*jwt.SigningMethodHMAC); !ok || method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidSigningMethod
		}

		return []byte(secret), nil
	})
	if err != nil {
		return fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return ErrInvalidToken
	}

	return nil
}
