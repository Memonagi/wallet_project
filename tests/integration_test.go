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
	"github.com/Memonagi/wallet_project/internal/models"
	"github.com/Memonagi/wallet_project/internal/server"
	"github.com/Memonagi/wallet_project/internal/service"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/stretchr/testify/suite"
)

const (
	pgDSN      = "postgresql://user:password@localhost:5432/mydatabase"
	port       = 5003
	walletPath = `/api/v1/wallets`
)

var existingUser = models.UsersInfo{
	UserID:    uuid.New(),
	Status:    "active",
	Archived:  false,
	CreatedAt: time.Now(),
	UpdatedAt: time.Now(),
}

type IntegrationTestSuite struct {
	suite.Suite
	cancelFn context.CancelFunc
	db       *database.Store
	service  *service.Service
	server   *server.Server
}

func (s *IntegrationTestSuite) SetupSuite() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFn = cancel

	var err error

	s.db, err = database.New(ctx, pgDSN)
	s.Require().NoError(err)

	err = s.db.Migrate(migrate.Up)
	s.Require().NoError(err)

	s.service = service.New(s.db)
	s.server = server.New(port, s.service)

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

func (s *IntegrationTestSuite) sendRequest(method, path string, status int, entity, result any) {
	body, err := json.Marshal(entity)
	s.Require().NoError(err)

	req, err := http.NewRequestWithContext(context.Background(), method,
		fmt.Sprintf("http://localhost:%d%s", port, path), bytes.NewReader(body))
	s.Require().NoError(err)

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