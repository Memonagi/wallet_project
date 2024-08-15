package tests

import (
	"context"
	"net/http"

	"github.com/Memonagi/wallet_project/internal/models"
	"github.com/google/uuid"
)

func (s *IntegrationTestSuite) TestCreateWallet() {
	wallet := models.Wallet{
		WalletID: uuid.New(),
		UserID:   uuid.New(),
		Name:     "proverkaPOST",
		Currency: "USD",
	}

	s.Run("user not found", func() {
		s.sendRequest(http.MethodPost, walletPath, http.StatusNotFound, &wallet, nil)
	})

	s.Run("created successfully", func() {
		err := s.db.UpsertUser(context.Background(), existingUser)
		s.Require().NoError(err)

		wallet.UserID = existingUser.UserID
		createdWallet := models.Wallet{}

		s.sendRequest(http.MethodPost, walletPath, http.StatusCreated, &wallet, &createdWallet)

		s.Require().Equal(wallet.UserID, createdWallet.UserID)
		s.Require().Equal(wallet.Name, createdWallet.Name)
		s.Require().Equal(wallet.Currency, createdWallet.Currency)
	})
}

func (s *IntegrationTestSuite) TestGetWallet() {
	wallet := models.Wallet{
		WalletID: uuid.New(),
		UserID:   uuid.New(),
		Name:     "proverkaGET",
		Currency: "USD",
	}

	s.Run("user not found", func() {
		uuidString := wallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString

		s.sendRequest(http.MethodGet, walletIDPath, http.StatusNotFound, nil, nil)
	})

	s.Run("get wallet successfully", func() {
		err := s.db.UpsertUser(context.Background(), existingUser)
		s.Require().NoError(err)

		wallet.UserID = existingUser.UserID
		createdWallet := models.Wallet{}

		s.sendRequest(http.MethodPost, walletPath, http.StatusCreated, &wallet, &createdWallet)

		uuidString := createdWallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString

		s.sendRequest(http.MethodGet, walletIDPath, http.StatusOK, nil, &createdWallet)

		s.Require().Equal(wallet.UserID, createdWallet.UserID)
		s.Require().Equal(wallet.Name, createdWallet.Name)
		s.Require().Equal(wallet.Currency, createdWallet.Currency)
	})
}

func (s *IntegrationTestSuite) TestUpdateWallet() {
	wallet := models.Wallet{
		WalletID: uuid.New(),
		UserID:   uuid.New(),
		Name:     "proverkaPATCH",
		Currency: "USD",
	}

	s.Run("user not found", func() {
		uuidString := wallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString
		s.sendRequest(http.MethodPatch, walletIDPath, http.StatusNotFound, &wallet, nil)
	})

	s.Run("updated successfully", func() {
		err := s.db.UpsertUser(context.Background(), existingUser)
		s.Require().NoError(err)

		wallet.UserID = existingUser.UserID
		createdWallet := models.Wallet{}

		s.sendRequest(http.MethodPost, walletPath, http.StatusCreated, &wallet, &createdWallet)

		updatedWallet := models.Wallet{}
		uuidString := createdWallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString

		s.sendRequest(http.MethodPatch, walletIDPath, http.StatusOK, &createdWallet, &updatedWallet)

		s.Require().Equal(wallet.UserID, updatedWallet.UserID)
		s.Require().Equal(wallet.Name, updatedWallet.Name)
		s.Require().Equal(wallet.Currency, updatedWallet.Currency)
	})
}

func (s *IntegrationTestSuite) TestDeleteWallet() {
	wallet := models.Wallet{
		WalletID: uuid.New(),
		UserID:   uuid.New(),
		Name:     "proverkaDELETE",
		Currency: "USD",
	}

	s.Run("user not found", func() {
		uuidString := wallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString

		s.sendRequest(http.MethodDelete, walletIDPath, http.StatusNotFound, nil, nil)
	})

	s.Run("deleted successfully", func() {
		err := s.db.UpsertUser(context.Background(), existingUser)
		s.Require().NoError(err)

		wallet.UserID = existingUser.UserID
		createdWallet := models.Wallet{}

		s.sendRequest(http.MethodPost, walletPath, http.StatusCreated, &wallet, &createdWallet)

		uuidString := createdWallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString

		s.sendRequest(http.MethodDelete, walletIDPath, http.StatusOK, nil, nil)
	})
}

func (s *IntegrationTestSuite) TestGetWallets() {
	wallet := models.Wallet{
		WalletID: uuid.New(),
		UserID:   uuid.New(),
		Name:     "proverkaGETALL",
		Currency: "USD",
	}

	// Правильно ли я поняла, что в данном случае проверка "user not found" не нужна,
	// так как мы в любом случае выводим содержимое всех существующих кошельков?
	s.Run("read successfully", func() {
		err := s.db.UpsertUser(context.Background(), existingUser)
		s.Require().NoError(err)

		wallet.UserID = existingUser.UserID
		createdWallet := models.Wallet{}

		s.sendRequest(http.MethodPost, walletPath, http.StatusCreated, &wallet, &createdWallet)

		s.sendRequest(http.MethodGet, walletPath, http.StatusOK, nil, &createdWallet)

		s.Require().Equal(wallet.UserID, createdWallet.UserID)
		s.Require().Equal(wallet.Name, createdWallet.Name)
		s.Require().Equal(wallet.Currency, createdWallet.Currency)
	})
}
