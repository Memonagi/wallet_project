package jwtclaims

import (
	"crypto/rsa"
	_ "embed"
	"fmt"
	"time"

	"github.com/Memonagi/wallet_project/internal/models"
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

//go:embed keys/private_key.pem
var privateKeyData []byte

//go:embed keys/public_key.pem
var publicKeyData []byte

func New() *Claims {
	tokenTime := time.Now().Add(hours * time.Hour)

	return &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(tokenTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
}

func ReadPrivateKey() (*rsa.PrivateKey, error) {
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyData)
	if err != nil {
		return nil, fmt.Errorf("error parsing private key: %w", err)
	}

	return privateKey, nil
}

func ReadPublicKey() (*rsa.PublicKey, error) {
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyData)
	if err != nil {
		return nil, fmt.Errorf("error parsing public key: %w", err)
	}

	return publicKey, nil
}

func (c *Claims) GenerateToken(secret *rsa.PrivateKey) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, c)

	tokenStr, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenStr, nil
}

func (c *Claims) ValidateToken(tokenStr string, secret *rsa.PublicKey) error {
	token, err := jwt.ParseWithClaims(tokenStr, c, func(token *jwt.Token) (interface{}, error) {
		if method, ok := token.Method.(*jwt.SigningMethodRSA); !ok || method != jwt.SigningMethodRS256 {
			return nil, models.ErrInvalidSigningMethod
		}

		return secret, nil
	})
	if err != nil {
		return fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return models.ErrInvalidToken
	}

	return nil
}

func (c *Claims) GetPublicKey() *rsa.PublicKey {
	key, err := ReadPublicKey()
	if err != nil {
		return nil
	}

	return key
}
