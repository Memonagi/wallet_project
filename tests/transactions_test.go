package tests

import (
	"context"
	"net/http"

	"github.com/Memonagi/wallet_project/internal/models"
	"github.com/google/uuid"
)

func (s *IntegrationTestSuite) TestDeposit() {
	// Arrange
	wallet := models.Wallet{
		WalletID: uuid.New(),
		UserID:   uuid.New(),
		Name:     "proverkaDEPOSIT",
		Currency: "RUB",
	}

	err := s.db.UpsertUser(context.Background(), existingUser)
	s.Require().NoError(err)

	wallet.UserID = existingUser.UserID
	createdWallet := models.Wallet{}

	s.sendRequest(http.MethodPost, walletPath, http.StatusCreated, &wallet, &createdWallet, existingUser)

	s.Run("deposit successfully", func() {
		transaction := models.Transaction{
			ID:            uuid.New(),
			FirstWalletID: createdWallet.WalletID,
			Money:         1000.0,
			Currency:      "RUB",
		}
		uuidString := createdWallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString + "/deposit"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusOK, &transaction, nil, existingUser)
	})

	s.Run("zero amount of money", func() {
		transaction := models.Transaction{
			ID:            uuid.New(),
			FirstWalletID: createdWallet.WalletID,
			Money:         0.0,
			Currency:      "RUB",
		}
		uuidString := createdWallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString + "/deposit"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusBadRequest, &transaction, nil, existingUser)
	})

	s.Run("negative amount of money", func() {
		transaction := models.Transaction{
			ID:            uuid.New(),
			FirstWalletID: createdWallet.WalletID,
			Money:         -1000.0,
			Currency:      "RUB",
		}
		uuidString := createdWallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString + "/deposit"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusBadRequest, &transaction, nil, existingUser)
	})

	s.Run("wrong currency", func() {
		transaction := models.Transaction{
			ID:            uuid.New(),
			FirstWalletID: createdWallet.WalletID,
			Money:         1000.0,
			Currency:      "USD",
		}
		uuidString := createdWallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString + "/deposit"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusBadRequest, &transaction, nil, existingUser)
	})

	s.Run("wallet not found", func() {
		transaction := models.Transaction{
			ID:            uuid.New(),
			FirstWalletID: uuid.New(),
			Money:         1000.0,
			Currency:      "RUB",
		}
		uuidString := createdWallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString + "/deposit"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusNotFound, &transaction, nil, existingUser)
	})

	s.Run("user is not the owner of the wallet", func() {
		transaction := models.Transaction{
			ID:            uuid.New(),
			FirstWalletID: createdWallet.WalletID,
			Money:         1000.0,
			Currency:      "RUB",
		}

		userFromAnotherMother := models.User{
			UserID: uuid.New(),
		}

		err = s.db.UpsertUser(context.Background(), userFromAnotherMother)
		s.Require().NoError(err)

		uuidString := createdWallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString + "/deposit"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusNotFound, &transaction, nil, userFromAnotherMother)
	})

	s.Run("user not found", func() {
		transaction := models.Transaction{
			ID:            uuid.New(),
			FirstWalletID: createdWallet.WalletID,
			Money:         1000.0,
			Currency:      "RUB",
		}

		newUser := models.User{
			UserID: uuid.New(),
		}

		createdWallet.UserID = newUser.UserID
		uuidString := createdWallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString + "/deposit"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusNotFound, &transaction, nil, newUser)
	})
}

func (s *IntegrationTestSuite) TestWithdrawMoney() {
	// Arrange
	wallet := models.Wallet{
		WalletID: uuid.New(),
		UserID:   uuid.New(),
		Name:     "proverkaWITHDRAW",
		Currency: "RUB",
	}

	err := s.db.UpsertUser(context.Background(), existingUser)
	s.Require().NoError(err)

	wallet.UserID = existingUser.UserID
	createdWallet := models.Wallet{}

	s.sendRequest(http.MethodPost, walletPath, http.StatusCreated, &wallet, &createdWallet, existingUser)

	transaction := models.Transaction{
		ID:            uuid.New(),
		FirstWalletID: createdWallet.WalletID,
		Money:         10000.0,
		Currency:      "RUB",
	}
	walletIDString := createdWallet.WalletID.String()
	depositPath := walletPath + "/" + walletIDString + "/deposit"

	s.sendRequest(http.MethodPut, depositPath, http.StatusOK, &transaction, nil, existingUser)

	s.Run("withdraw successfully", func() {
		transaction = models.Transaction{
			ID:            uuid.New(),
			FirstWalletID: createdWallet.WalletID,
			Money:         500.0,
			Currency:      "RUB",
		}
		uuidString := createdWallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString + "/withdraw"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusOK, &transaction, nil, existingUser)
	})

	s.Run("zero amount of money", func() {
		transaction = models.Transaction{
			ID:            uuid.New(),
			FirstWalletID: createdWallet.WalletID,
			Money:         0.0,
			Currency:      "RUB",
		}
		uuidString := createdWallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString + "/withdraw"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusBadRequest, &transaction, nil, existingUser)
	})

	s.Run("negative amount of money", func() {
		transaction = models.Transaction{
			ID:            uuid.New(),
			FirstWalletID: createdWallet.WalletID,
			Money:         -1000.0,
			Currency:      "RUB",
		}
		uuidString := createdWallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString + "/withdraw"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusBadRequest, &transaction, nil, existingUser)
	})

	s.Run("wrong currency", func() {
		transaction = models.Transaction{
			ID:            uuid.New(),
			FirstWalletID: createdWallet.WalletID,
			Money:         1000.0,
			Currency:      "USD",
		}
		uuidString := createdWallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString + "/withdraw"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusBadRequest, &transaction, nil, existingUser)
	})

	s.Run("wallet not found", func() {
		transaction = models.Transaction{
			ID:            uuid.New(),
			FirstWalletID: uuid.New(),
			Money:         1000.0,
			Currency:      "RUB",
		}
		uuidString := createdWallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString + "/withdraw"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusNotFound, &transaction, nil, existingUser)
	})

	s.Run("user is not the owner of the wallet", func() {
		transaction = models.Transaction{
			ID:            uuid.New(),
			FirstWalletID: createdWallet.WalletID,
			Money:         1000.0,
			Currency:      "RUB",
		}

		userFromAnotherMother := models.User{
			UserID: uuid.New(),
		}

		err = s.db.UpsertUser(context.Background(), userFromAnotherMother)
		s.Require().NoError(err)

		uuidString := createdWallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString + "/withdraw"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusNotFound, &transaction, nil, userFromAnotherMother)
	})

	s.Run("user not found", func() {
		transaction = models.Transaction{
			ID:            uuid.New(),
			FirstWalletID: createdWallet.WalletID,
			Money:         1000.0,
			Currency:      "RUB",
		}

		newUser := models.User{
			UserID: uuid.New(),
		}

		createdWallet.UserID = newUser.UserID
		uuidString := createdWallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString + "/withdraw"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusNotFound, &transaction, nil, newUser)
	})
}

func (s *IntegrationTestSuite) TestTransfer() {
	// Arrange
	firstWallet := models.Wallet{
		WalletID: uuid.New(),
		UserID:   uuid.New(),
		Name:     "proverkaTRANSFER_1",
		Currency: "RUB",
	}

	secondWallet := models.Wallet{
		WalletID: uuid.New(),
		UserID:   uuid.New(),
		Name:     "proverkaTRANSFER_2",
		Currency: "RUB",
	}

	err := s.db.UpsertUser(context.Background(), existingUser)
	s.Require().NoError(err)

	secondUser := models.User{
		UserID: uuid.New(),
	}

	err = s.db.UpsertUser(context.Background(), secondUser)
	s.Require().NoError(err)

	firstWallet.UserID = existingUser.UserID
	secondWallet.UserID = secondUser.UserID
	firstCreatedWallet := models.Wallet{}
	secondCreatedWallet := models.Wallet{}

	s.sendRequest(http.MethodPost, walletPath, http.StatusCreated, &firstWallet, &firstCreatedWallet, existingUser)

	s.sendRequest(http.MethodPost, walletPath, http.StatusCreated, &secondWallet, &secondCreatedWallet, secondUser)

	transaction := models.Transaction{
		ID:            uuid.New(),
		FirstWalletID: firstCreatedWallet.WalletID,
		Money:         10000.0,
		Currency:      "RUB",
	}
	walletIDString := firstCreatedWallet.WalletID.String()
	depositPath := walletPath + "/" + walletIDString + "/deposit"

	s.sendRequest(http.MethodPut, depositPath, http.StatusOK, &transaction, nil, existingUser)

	s.Run("successful transfer", func() {
		transaction = models.Transaction{
			ID:             uuid.New(),
			FirstWalletID:  firstCreatedWallet.WalletID,
			SecondWalletID: secondCreatedWallet.WalletID,
			Money:          1000.0,
			Currency:       "RUB",
		}
		uuidString := firstCreatedWallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString + "/transfer"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusOK, &transaction, nil, existingUser)
	})

	s.Run("zero amount of money", func() {
		transaction = models.Transaction{
			ID:             uuid.New(),
			FirstWalletID:  firstCreatedWallet.WalletID,
			SecondWalletID: secondCreatedWallet.WalletID,
			Money:          0.0,
			Currency:       "RUB",
		}
		uuidString := firstCreatedWallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString + "/transfer"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusBadRequest, &transaction, nil, existingUser)
	})

	s.Run("negative amount of money", func() {
		transaction = models.Transaction{
			ID:             uuid.New(),
			FirstWalletID:  firstCreatedWallet.WalletID,
			SecondWalletID: secondCreatedWallet.WalletID,
			Money:          -1000.0,
			Currency:       "RUB",
		}
		uuidString := firstCreatedWallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString + "/transfer"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusBadRequest, &transaction, nil, existingUser)
	})

	s.Run("wrong currency of first wallet", func() {
		transaction = models.Transaction{
			ID:             uuid.New(),
			FirstWalletID:  firstCreatedWallet.WalletID,
			SecondWalletID: secondCreatedWallet.WalletID,
			Money:          1000.0,
			Currency:       "USD",
		}
		uuidString := firstCreatedWallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString + "/transfer"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusBadRequest, &transaction, nil, existingUser)
	})

	s.Run("different currency of second wallet", func() {
		updateSecondWallet := models.Wallet{
			WalletID: secondCreatedWallet.WalletID,
			UserID:   secondCreatedWallet.UserID,
			Name:     secondCreatedWallet.Name,
			Currency: "USD",
		}
		updateWalletID := secondCreatedWallet.WalletID.String()
		updatePath := walletPath + "/" + updateWalletID

		s.sendRequest(http.MethodPatch, updatePath, http.StatusOK, &updateSecondWallet, &secondCreatedWallet, secondUser)

		transaction = models.Transaction{
			ID:             uuid.New(),
			FirstWalletID:  firstCreatedWallet.WalletID,
			SecondWalletID: secondCreatedWallet.WalletID,
			Money:          1000.0,
			Currency:       "RUB",
		}

		uuidString := firstCreatedWallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString + "/transfer"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusOK, &transaction, nil, existingUser)
	})

	s.Run("wallet not found", func() {
		transaction = models.Transaction{
			ID:             uuid.New(),
			FirstWalletID:  uuid.New(),
			SecondWalletID: secondCreatedWallet.WalletID,
			Money:          1000.0,
			Currency:       "RUB",
		}
		uuidString := firstCreatedWallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString + "/transfer"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusNotFound, &transaction, nil, existingUser)
	})

	s.Run("user is not the owner of the wallet", func() {
		transaction = models.Transaction{
			ID:             uuid.New(),
			FirstWalletID:  firstCreatedWallet.WalletID,
			SecondWalletID: secondCreatedWallet.WalletID,
			Money:          1000.0,
			Currency:       "RUB",
		}

		userFromAnotherMother := models.User{
			UserID: uuid.New(),
		}

		err = s.db.UpsertUser(context.Background(), userFromAnotherMother)
		s.Require().NoError(err)

		uuidString := firstCreatedWallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString + "/transfer"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusNotFound, &transaction, nil, userFromAnotherMother)
	})

	s.Run("first user not found", func() {
		transaction = models.Transaction{
			ID:             uuid.New(),
			FirstWalletID:  firstCreatedWallet.WalletID,
			SecondWalletID: secondCreatedWallet.WalletID,
			Money:          1000.0,
			Currency:       "RUB",
		}

		newUser := models.User{
			UserID: uuid.New(),
		}

		firstCreatedWallet.UserID = newUser.UserID
		uuidString := firstCreatedWallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString + "/transfer"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusNotFound, &transaction, nil, newUser)
	})

	s.Run("second user not found", func() {
		transaction = models.Transaction{
			ID:             uuid.New(),
			FirstWalletID:  firstCreatedWallet.WalletID,
			SecondWalletID: uuid.New(),
			Money:          1000.0,
			Currency:       "RUB",
		}
		uuidString := firstCreatedWallet.WalletID.String()
		walletIDPath := walletPath + "/" + uuidString + "/transfer"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusNotFound, &transaction, nil, existingUser)
	})
}
