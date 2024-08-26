package jwtclaims_test

import (
	"testing"
	"time"

	jwtclaims "github.com/Memonagi/wallet_project/internal/jwt-claims"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGenerateToken(t *testing.T) {
	secret := "my_secret_string"

	claims := jwtclaims.New()
	claims.UserID = uuid.New()
	claims.Email = "test@yandex.ru"
	claims.Role = "moderator"

	tokenStr, err := claims.GenerateToken(secret)
	require.NoError(t, err)
	require.NotEmpty(t, tokenStr, "tokenStr should not be empty")
}

func TestValidateToken(t *testing.T) {
	secret := "my_secret_string"

	claims := jwtclaims.New()
	claims.UserID = uuid.New()
	claims.Email = "test@yandex.ru"
	claims.Role = "moderator"

	tokenStr, err := claims.GenerateToken(secret)
	require.NoError(t, err)
	require.NotEmpty(t, tokenStr, "tokenStr should not be empty")

	newClaims := &jwtclaims.Claims{}
	err = newClaims.ValidateToken(tokenStr, secret)
	require.NoError(t, err)

	require.Equal(t, claims.UserID, newClaims.UserID)
	require.Equal(t, claims.Email, newClaims.Email)
	require.Equal(t, claims.Role, newClaims.Role)
}

func TestValidateInvalidSignature(t *testing.T) {
	secret := "my_secret_string"
	invalidSecret := "bla_bla_bla"

	claims := jwtclaims.New()
	claims.UserID = uuid.New()
	claims.Email = "test@yandex.ru"
	claims.Role = "moderator"

	tokenStr, err := claims.GenerateToken(secret)
	require.NoError(t, err)
	require.NotEmpty(t, tokenStr, "tokenStr should not be empty")

	newClaims := &jwtclaims.Claims{}
	err = newClaims.ValidateToken(tokenStr, invalidSecret)
	require.Error(t, err, "Must return error with invalid signature")
}

func TestValidateInvalidSignatureMethod(t *testing.T) {
	secret := "my_secret_string"

	claims := jwtclaims.New()
	claims.UserID = uuid.New()
	claims.Email = "test@yandex.ru"
	claims.Role = "moderator"

	token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	require.NoError(t, err)
	require.NotEmpty(t, tokenStr, "tokenStr should not be empty")

	newClaims := &jwtclaims.Claims{}
	err = newClaims.ValidateToken(tokenStr, secret)
	require.Error(t, err, "Must return error with invalid signature method")
}

func TestValidateExpiredToken(t *testing.T) {
	secret := "my_secret_string"

	claims := &jwtclaims.Claims{
		UserID: uuid.New(),
		Email:  "test@yandex.ru",
		Role:   "moderator",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-48 * time.Hour)),
		},
	}

	tokenStr, err := claims.GenerateToken(secret)
	require.NoError(t, err)
	require.NotEmpty(t, tokenStr, "tokenStr should not be empty")

	newClaims := &jwtclaims.Claims{}
	err = newClaims.ValidateToken(tokenStr, secret)
	require.Error(t, err, "Must return error with expired token")
}
