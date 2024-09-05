package tests

import (
	"context"
	"net/http"

	"github.com/Memonagi/wallet_project/internal/models"
	"github.com/google/uuid"
)

func (s *IntegrationTestSuite) TestCreateWallet() {
	// Arrange
	wallet := models.Wallet{
		WalletID: uuid.New(),
		UserID:   uuid.New(),
		Name:     "proverkaPOST",
		Currency: "USD",
	}

	s.Run("user not found", func() {
		// Act
		s.sendRequest(http.MethodPost, walletPath, http.StatusNotFound, &wallet, nil, existingUser)
	})

	s.Run("created successfully", func() {
		err := s.db.UpsertUser(context.Background(), existingUser)
		s.Require().NoError(err)

		wallet.UserID = existingUser.UserID
		createdWallet := models.Wallet{}

		// Act
		s.sendRequest(http.MethodPost, walletPath, http.StatusCreated, &wallet, &createdWallet, existingUser)

		// Assert
		s.Require().Equal(wallet.UserID, createdWallet.UserID)
		s.Require().Equal(wallet.Name, createdWallet.Name)
		s.Require().Equal(wallet.Currency, createdWallet.Currency)
	})

	s.Run("user is not the owner of the wallet", func() {
		err := s.db.UpsertUser(context.Background(), existingUser)
		s.Require().NoError(err)

		userFromAnotherMother := models.User{
			UserID: uuid.New(),
		}

		err = s.db.UpsertUser(context.Background(), userFromAnotherMother)
		s.Require().NoError(err)

		wallet.UserID = userFromAnotherMother.UserID

		// Act
		s.sendRequest(http.MethodPost, walletPath, http.StatusNotFound, &wallet, nil, existingUser)
	})
}

func (s *IntegrationTestSuite) TestGetWallet() {
	// Arrange
	wallet := models.Wallet{
		WalletID: uuid.New(),
		UserID:   uuid.New(),
		Name:     "proverkaGET",
		Currency: "USD",
	}

	err := s.db.UpsertUser(context.Background(), existingUser)
	s.Require().NoError(err)

	wallet.UserID = existingUser.UserID
	createdWallet := models.Wallet{}

	s.sendRequest(http.MethodPost, walletPath, http.StatusCreated, &wallet, &createdWallet, existingUser)

	s.Run("user not found", func() {
		uuidString := wallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString

		// Act
		s.sendRequest(http.MethodGet, walletIDPath, http.StatusNotFound, nil, nil, existingUser)
	})

	s.Run("get wallet successfully", func() {
		uuidString := createdWallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString

		// Act
		s.sendRequest(http.MethodGet, walletIDPath, http.StatusOK, nil, &createdWallet, existingUser)

		// Assert
		s.Require().Equal(wallet.UserID, createdWallet.UserID)
		s.Require().Equal(wallet.Name, createdWallet.Name)
		s.Require().Equal(wallet.Currency, createdWallet.Currency)
	})

	s.Run("user is not the owner of the wallet", func() {
		userFromAnotherMother := models.User{
			UserID: uuid.New(),
		}

		err = s.db.UpsertUser(context.Background(), userFromAnotherMother)
		s.Require().NoError(err)

		uuidString := createdWallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString

		// Act
		s.sendRequest(http.MethodGet, walletIDPath, http.StatusNotFound, nil, nil, userFromAnotherMother)
	})
}

func (s *IntegrationTestSuite) TestUpdateWallet() {
	// Arrange
	wallet := models.Wallet{
		WalletID: uuid.New(),
		UserID:   uuid.New(),
		Name:     "proverkaPATCH",
		Currency: "USD",
	}

	err := s.db.UpsertUser(context.Background(), existingUser)
	s.Require().NoError(err)

	wallet.UserID = existingUser.UserID
	createdWallet := models.Wallet{}

	s.sendRequest(http.MethodPost, walletPath, http.StatusCreated, &wallet, &createdWallet, existingUser)

	s.Run("user not found", func() {
		uuidString := wallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString

		// Act
		s.sendRequest(http.MethodPatch, walletIDPath, http.StatusNotFound, &wallet, nil, existingUser)
	})

	s.Run("name updated successfully", func() {
		updatedWallet := models.Wallet{
			WalletID: createdWallet.WalletID,
			UserID:   createdWallet.UserID,
			Name:     "renamedWallet",
			Currency: createdWallet.Currency,
		}
		uuidString := createdWallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString

		// Act
		s.sendRequest(http.MethodPatch, walletIDPath, http.StatusOK, &updatedWallet, &createdWallet, existingUser)

		// Assert
		s.Require().Equal(updatedWallet.UserID, createdWallet.UserID)
		s.Require().Equal(updatedWallet.Name, createdWallet.Name)
		s.Require().Equal(updatedWallet.Currency, createdWallet.Currency)
	})

	s.Run("currency updated successfully", func() {
		updatedWallet := models.Wallet{
			WalletID: createdWallet.WalletID,
			UserID:   createdWallet.UserID,
			Name:     createdWallet.Name,
			Currency: "RUB",
		}
		uuidString := createdWallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString

		// Act
		s.sendRequest(http.MethodPatch, walletIDPath, http.StatusOK, &updatedWallet, &createdWallet, existingUser)

		// Assert
		s.Require().Equal(updatedWallet.UserID, createdWallet.UserID)
		s.Require().Equal(updatedWallet.Name, createdWallet.Name)
		s.Require().Equal(updatedWallet.Currency, createdWallet.Currency)
	})

	s.Run("all info updated successfully", func() {
		updatedWallet := models.Wallet{
			WalletID: createdWallet.WalletID,
			UserID:   createdWallet.UserID,
			Name:     "renamedWallet",
			Currency: "RUB",
		}
		uuidString := createdWallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString

		// Act
		s.sendRequest(http.MethodPatch, walletIDPath, http.StatusOK, &updatedWallet, &createdWallet, existingUser)

		// Assert
		s.Require().Equal(updatedWallet.UserID, createdWallet.UserID)
		s.Require().Equal(updatedWallet.Name, createdWallet.Name)
		s.Require().Equal(updatedWallet.Currency, createdWallet.Currency)
	})

	s.Run("nothing to update", func() {
		updatedWallet := createdWallet
		uuidString := createdWallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString

		// Act
		s.sendRequest(http.MethodPatch, walletIDPath, http.StatusOK, &updatedWallet, &createdWallet, existingUser)

		// Assert
		s.Require().Equal(updatedWallet.UserID, createdWallet.UserID)
		s.Require().Equal(updatedWallet.Name, createdWallet.Name)
		s.Require().Equal(updatedWallet.Currency, createdWallet.Currency)
	})

	s.Run("user is not the owner of the wallet", func() {
		userFromAnotherMother := models.User{
			UserID: uuid.New(),
		}
		err = s.db.UpsertUser(context.Background(), userFromAnotherMother)
		s.Require().NoError(err)

		updatedWallet := models.Wallet{
			WalletID: uuid.New(),
			UserID:   userFromAnotherMother.UserID,
			Name:     "renamedWallet",
			Currency: "RUB",
		}
		uuidString := createdWallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString

		// Act
		s.sendRequest(http.MethodPatch, walletIDPath, http.StatusNotFound, &updatedWallet, nil, userFromAnotherMother)
	})
}

func (s *IntegrationTestSuite) TestDeleteWallet() {
	// Arrange
	wallet := models.Wallet{
		WalletID: uuid.New(),
		UserID:   uuid.New(),
		Name:     "proverkaDELETE",
		Currency: "USD",
	}

	s.Run("user not found", func() {
		uuidString := wallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString

		// Act
		s.sendRequest(http.MethodDelete, walletIDPath, http.StatusNotFound, nil, nil, existingUser)
	})

	err := s.db.UpsertUser(context.Background(), existingUser)
	s.Require().NoError(err)

	wallet.UserID = existingUser.UserID
	createdWallet := models.Wallet{}

	s.sendRequest(http.MethodPost, walletPath, http.StatusCreated, &wallet, &createdWallet, existingUser)

	s.Run("deleted successfully", func() {
		uuidString := createdWallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString

		// Act
		s.sendRequest(http.MethodDelete, walletIDPath, http.StatusOK, nil, nil, existingUser)
	})

	s.Run("user is not the owner of the wallet", func() {
		userFromAnotherMother := models.User{
			UserID: uuid.New(),
		}
		err = s.db.UpsertUser(context.Background(), userFromAnotherMother)
		s.Require().NoError(err)

		uuidString := createdWallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString

		// Act
		s.sendRequest(http.MethodDelete, walletIDPath, http.StatusNotFound, nil, nil, userFromAnotherMother)
	})
}

func (s *IntegrationTestSuite) TestGetWallets() {
	// Arrange
	wallet := models.Wallet{
		WalletID: uuid.New(),
		UserID:   uuid.New(),
		Name:     "1_proverkaGETALL",
		Currency: "USD",
	}

	secWallet := models.Wallet{
		WalletID: uuid.New(),
		UserID:   uuid.New(),
		Name:     "2_proverkaGETALL",
		Currency: "RUB",
	}

	thirdWallet := models.Wallet{
		WalletID: uuid.New(),
		UserID:   uuid.New(),
		Name:     "3_proverkaGETALL",
		Currency: "EUR",
	}

	s.Run("empty list", func() {
		var wallets []models.Wallet

		// Act
		s.sendRequest(http.MethodGet, walletPath, http.StatusOK, nil, &wallets, existingUser)

		// Assert
		s.Require().Len(wallets, 0)
	})

	// Arrange
	err := s.db.UpsertUser(context.Background(), existingUser)
	s.Require().NoError(err)

	wallet.UserID = existingUser.UserID
	createdWallet := models.Wallet{}

	s.sendRequest(http.MethodPost, walletPath, http.StatusCreated, &wallet, &createdWallet, existingUser)

	secWallet.UserID = existingUser.UserID
	secCreatedWallet := models.Wallet{}

	s.sendRequest(http.MethodPost, walletPath, http.StatusCreated, &secWallet, &secCreatedWallet, existingUser)

	thirdWallet.UserID = existingUser.UserID
	thirdCreatedWallet := models.Wallet{}

	s.sendRequest(http.MethodPost, walletPath, http.StatusCreated, &thirdWallet, &thirdCreatedWallet, existingUser)

	s.Run("read successfully", func() {
		var wallets []models.Wallet

		// Act
		s.sendRequest(http.MethodGet, walletPath, http.StatusOK, nil, &wallets, existingUser)

		// Assert
		s.Require().Len(wallets, 3)
	})

	s.Run("descending names read successfully", func() {
		var wallets []models.Wallet
		descWalletPath := walletPath + "?sorting=name&descending=true"

		// Act
		s.sendRequest(http.MethodGet, descWalletPath, http.StatusOK, nil, &wallets, existingUser)

		// Assert
		s.Require().Equal(createdWallet.UserID, wallets[2].UserID)
		s.Require().Equal(createdWallet.Name, wallets[2].Name)
		s.Require().Equal(createdWallet.Currency, wallets[2].Currency)

		s.Require().Equal(secCreatedWallet.UserID, wallets[1].UserID)
		s.Require().Equal(secCreatedWallet.Name, wallets[1].Name)
		s.Require().Equal(secCreatedWallet.Currency, wallets[1].Currency)

		s.Require().Equal(thirdCreatedWallet.UserID, wallets[0].UserID)
		s.Require().Equal(thirdCreatedWallet.Name, wallets[0].Name)
		s.Require().Equal(thirdCreatedWallet.Currency, wallets[0].Currency)
	})

	s.Run("filter read successfully", func() {
		var wallets []models.Wallet
		filterWalletPath := walletPath + "?filter=rub"

		// Act
		s.sendRequest(http.MethodGet, filterWalletPath, http.StatusOK, nil, &wallets, existingUser)

		// Assert
		s.Require().Len(wallets, 1)
	})

	s.Run("limit and offset read successfully", func() {
		var wallets []models.Wallet
		limitWalletPath := walletPath + "?sorting=name&limit=2&offset=2"

		// Act
		s.sendRequest(http.MethodGet, limitWalletPath, http.StatusOK, nil, &wallets, existingUser)

		// Assert
		s.Require().Len(wallets, 1)
		s.Require().Equal(wallets[0], thirdCreatedWallet)
	})

	s.Run("user is not the owner of the wallet", func() {
		userFromAnotherMother := models.User{
			UserID: uuid.New(),
		}

		err = s.db.UpsertUser(context.Background(), userFromAnotherMother)
		s.Require().NoError(err)

		var wallets []models.Wallet

		// Act
		s.sendRequest(http.MethodGet, walletPath, http.StatusOK, nil, &wallets, userFromAnotherMother)

		// Assert
		s.Require().Len(wallets, 0)
	})
}

// TODO authorization
// TODO technical debt - update is transaction
// TODO finance transactions
// TODO history of transactions
// TODO metrics
// TODO id
