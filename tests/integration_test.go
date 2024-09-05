package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/Memonagi/wallet_project/internal/database"
	jwtclaims "github.com/Memonagi/wallet_project/internal/jwt-claims"
	"github.com/Memonagi/wallet_project/internal/models"
	"github.com/Memonagi/wallet_project/internal/server"
	"github.com/Memonagi/wallet_project/internal/service"
	xrclient "github.com/Memonagi/wallet_project/internal/xr-client"
	xrserver "github.com/Memonagi/wallet_project/internal/xr-server"
	xrservice "github.com/Memonagi/wallet_project/internal/xr-service"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/stretchr/testify/suite"
)

const (
	pgDSN      = "postgresql://user:password@localhost:5432/mydatabase"
	port       = 5003
	xrAddress  = "http://localhost:2607"
	xrPort     = 2607
	walletPath = `/api/v1/wallets`
)

var existingUser = models.User{
	UserID:    uuid.New(),
	Status:    "active",
	Archived:  false,
	CreatedAt: time.Now(),
	UpdatedAt: time.Now(),
}

type IntegrationTestSuite struct {
	suite.Suite
	cancelFn  context.CancelFunc
	db        *database.Store
	service   *service.Service
	server    *server.Server
	client    *xrclient.Client
	xrService *xrservice.Service
	xrServer  *xrserver.Server
	jwtClaims *jwtclaims.Claims
}

func (s *IntegrationTestSuite) SetupSuite() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFn = cancel

	var err error

	s.db, err = database.New(ctx, database.Config{Dsn: pgDSN})
	s.Require().NoError(err)

	err = s.db.Migrate(migrate.Up)
	s.Require().NoError(err)

	s.xrService = xrservice.New()
	s.xrServer = xrserver.New(xrPort, s.xrService)

	go func() {
		err := s.xrServer.Run(ctx)
		s.Require().NoError(err)
	}()

	s.client = xrclient.New(xrclient.Config{ServerAddress: xrAddress})
	s.service = service.New(s.db, s.client)
	s.jwtClaims = jwtclaims.New()
	s.server = server.New(server.Config{Port: port}, s.service, s.jwtClaims.GetPublicKey())

	go func() {
		err = s.server.Run(ctx)
		s.Require().NoError(err)
	}()

	time.Sleep(50 * time.Millisecond)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.cancelFn()
}

func (s *IntegrationTestSuite) SetupTest() {
	err := s.db.Truncate(context.Background(), "wallets", "users")
	s.Require().NoError(err)
}

func TestIntegrationSetupSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) sendRequest(method, path string, status int, entity, result any, user models.User) {
	body, err := json.Marshal(entity)
	s.Require().NoError(err)

	req, err := http.NewRequestWithContext(context.Background(), method,
		fmt.Sprintf("http://localhost:%d%s", port, path), bytes.NewReader(body))
	s.Require().NoError(err)

	token := s.getToken(user)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := http.Client{}

	resp, err := client.Do(req)
	s.Require().NoError(err)

	defer func() {
		err = resp.Body.Close()
		s.Require().NoError(err)
	}()

	s.Require().Equal(status, resp.StatusCode)

	if result == nil {
		return
	}

	respBody, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)

	err = json.Unmarshal(respBody, result)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) getToken(user models.User) string {
	claims := jwtclaims.Claims{
		UserID: user.UserID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	privateKey, err := jwtclaims.ReadPrivateKey()
	s.Require().NoError(err)

	token, err := claims.GenerateToken(privateKey)
	s.Require().NoError(err)

	return token
}
