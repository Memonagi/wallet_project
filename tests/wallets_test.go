package tests

import (
	"net/http"

	"github.com/Memonagi/wallet_project/internal/models"
	"github.com/google/uuid"
)

func (s *IntegrationTestSuite) TestCreateWallet() {
	wallet := models.Wallet{
		UserID:   uuid.New(),
		Name:     "biba",
		Currency: "USD",
	}

	s.Run("non-existent user", func() {
		s.sendRequest(http.MethodPost, walletPath, http.StatusNotFound, &wallet, nil)
	})

	s.Run("successful creation", func() {
		wallet.UserID = existingUser.UserID
		createdWallet := models.Wallet{}

		s.sendRequest(http.MethodPost, walletPath, http.StatusCreated, &wallet, &createdWallet)

		s.Require().Equal(wallet.UserID, createdWallet.UserID)
		s.Require().Equal(wallet.Currency, createdWallet.Currency)
		s.Require().Equal(wallet.Name, createdWallet.Name)
	})
}

// TODO Dockerfile
// TODO deploy workers
// TODO convert currency
// TODO config
// TODO update endpoint
// TODO test get, update, delete
// TODO get wallets and test it
// TODO authorization
