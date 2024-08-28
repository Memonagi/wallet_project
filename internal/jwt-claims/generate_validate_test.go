package jwtclaims_test

import (
	"crypto/rand"
	"crypto/rsa"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"

	jwtclaims "github.com/Memonagi/wallet_project/internal/jwt-claims"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type JWTTestSuite struct {
	suite.Suite
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func (s *JWTTestSuite) SetupSuite() {
	var err error

	s.privateKey, err = jwtclaims.ReadPrivateKey()
	s.Require().NoError(err)

	s.publicKey, err = jwtclaims.ReadPublicKey()
	s.Require().NoError(err)
}

func TestJWTSetupSuite(t *testing.T) {
	suite.Run(t, new(JWTTestSuite))
}

func (s *JWTTestSuite) TestGenerateToken() {
	claims := jwtclaims.New()
	claims.UserID = uuid.New()
	claims.Email = "test@yandex.ru"
	claims.Role = "moderator"

	tokenStr, err := claims.GenerateToken(s.privateKey)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), tokenStr, "tokenStr should not be empty")
}

func (s *JWTTestSuite) TestValidateToken() {
	claims := jwtclaims.New()
	claims.UserID = uuid.New()
	claims.Email = "test@yandex.ru"
	claims.Role = "moderator"

	tokenStr, err := claims.GenerateToken(s.privateKey)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), tokenStr, "tokenStr should not be empty")

	newClaims := &jwtclaims.Claims{}
	err = newClaims.ValidateToken(tokenStr, s.publicKey)
	require.NoError(s.T(), err)

	require.Equal(s.T(), claims.UserID, newClaims.UserID)
	require.Equal(s.T(), claims.Email, newClaims.Email)
	require.Equal(s.T(), claims.Role, newClaims.Role)
}

func (s *JWTTestSuite) TestValidateInvalidSignature() {
	randomPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	randomPublicKey := &randomPrivateKey.PublicKey

	claims := jwtclaims.New()
	claims.UserID = uuid.New()
	claims.Email = "test@yandex.ru"
	claims.Role = "moderator"

	tokenStr, err := claims.GenerateToken(s.privateKey)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), tokenStr, "tokenStr should not be empty")

	newClaims := &jwtclaims.Claims{}
	err = newClaims.ValidateToken(tokenStr, randomPublicKey)
	require.Error(s.T(), err, "Must return error with invalid signature")
}

func (s *JWTTestSuite) TestValidateInvalidSignatureMethod() {
	claims := jwtclaims.New()
	claims.UserID = uuid.New()
	claims.Email = "test@yandex.ru"
	claims.Role = "moderator"

	token := jwt.NewWithClaims(jwt.SigningMethodRS384, claims)
	tokenStr, err := token.SignedString(s.privateKey)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), tokenStr, "tokenStr should not be empty")

	newClaims := &jwtclaims.Claims{}
	err = newClaims.ValidateToken(tokenStr, s.publicKey)
	require.Error(s.T(), err, "Must return error with invalid signature method")
}

func (s *JWTTestSuite) TestValidateExpiredToken() {
	claims := &jwtclaims.Claims{
		UserID: uuid.New(),
		Email:  "test@yandex.ru",
		Role:   "moderator",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-48 * time.Hour)),
		},
	}

	tokenStr, err := claims.GenerateToken(s.privateKey)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), tokenStr, "tokenStr should not be empty")

	newClaims := &jwtclaims.Claims{}
	err = newClaims.ValidateToken(tokenStr, s.publicKey)
	require.Error(s.T(), err, "Must return error with expired token")
}
