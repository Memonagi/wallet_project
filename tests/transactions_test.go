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
		WalletID: models.WalletID(uuid.New()),
		UserID:   models.UserID(uuid.New()),
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
			ID:            models.TxID(uuid.New()),
			FirstWalletID: createdWallet.WalletID,
			Money:         1000.0,
			Currency:      "RUB",
		}
		uuidString := uuid.UUID(createdWallet.WalletID).String()
		walletIDPath := walletPath + "/" + uuidString + "/deposit"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusOK, &transaction, nil, existingUser)
	})

	s.Run("zero amount of money", func() {
		transaction := models.Transaction{
			ID:            models.TxID(uuid.New()),
			FirstWalletID: createdWallet.WalletID,
			Money:         0.0,
			Currency:      "RUB",
		}
		uuidString := uuid.UUID(createdWallet.WalletID).String()
		walletIDPath := walletPath + "/" + uuidString + "/deposit"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusBadRequest, &transaction, nil, existingUser)
	})

	s.Run("negative amount of money", func() {
		transaction := models.Transaction{
			ID:            models.TxID(uuid.New()),
			FirstWalletID: createdWallet.WalletID,
			Money:         -1000.0,
			Currency:      "RUB",
		}
		uuidString := uuid.UUID(createdWallet.WalletID).String()
		walletIDPath := walletPath + "/" + uuidString + "/deposit"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusBadRequest, &transaction, nil, existingUser)
	})

	s.Run("wrong currency", func() {
		transaction := models.Transaction{
			ID:            models.TxID(uuid.New()),
			FirstWalletID: createdWallet.WalletID,
			Money:         1000.0,
			Currency:      "USD",
		}
		uuidString := uuid.UUID(createdWallet.WalletID).String()
		walletIDPath := walletPath + "/" + uuidString + "/deposit"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusBadRequest, &transaction, nil, existingUser)
	})

	s.Run("wallet not found", func() {
		transaction := models.Transaction{
			ID:            models.TxID(uuid.New()),
			FirstWalletID: models.WalletID(uuid.New()),
			Money:         1000.0,
			Currency:      "RUB",
		}
		uuidString := uuid.UUID(createdWallet.WalletID).String()
		walletIDPath := walletPath + "/" + uuidString + "/deposit"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusNotFound, &transaction, nil, existingUser)
	})

	s.Run("user is not the owner of the wallet", func() {
		transaction := models.Transaction{
			ID:            models.TxID(uuid.New()),
			FirstWalletID: createdWallet.WalletID,
			Money:         1000.0,
			Currency:      "RUB",
		}

		userFromAnotherMother := models.User{
			UserID: models.UserID(uuid.New()),
		}

		err = s.db.UpsertUser(context.Background(), userFromAnotherMother)
		s.Require().NoError(err)

		uuidString := uuid.UUID(createdWallet.WalletID).String()
		walletIDPath := walletPath + "/" + uuidString + "/deposit"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusNotFound, &transaction, nil, userFromAnotherMother)
	})

	s.Run("user not found", func() {
		transaction := models.Transaction{
			ID:            models.TxID(uuid.New()),
			FirstWalletID: createdWallet.WalletID,
			Money:         1000.0,
			Currency:      "RUB",
		}

		newUser := models.User{
			UserID: models.UserID(uuid.New()),
		}

		createdWallet.UserID = newUser.UserID
		uuidString := uuid.UUID(createdWallet.WalletID).String()
		walletIDPath := walletPath + "/" + uuidString + "/deposit"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusNotFound, &transaction, nil, newUser)
	})
}

func (s *IntegrationTestSuite) TestWithdrawMoney() {
	// Arrange
	wallet := models.Wallet{
		WalletID: models.WalletID(uuid.New()),
		UserID:   models.UserID(uuid.New()),
		Name:     "proverkaWITHDRAW",
		Currency: "RUB",
	}

	err := s.db.UpsertUser(context.Background(), existingUser)
	s.Require().NoError(err)

	wallet.UserID = existingUser.UserID
	createdWallet := models.Wallet{}

	s.sendRequest(http.MethodPost, walletPath, http.StatusCreated, &wallet, &createdWallet, existingUser)

	transaction := models.Transaction{
		ID:            models.TxID(uuid.New()),
		FirstWalletID: createdWallet.WalletID,
		Money:         10000.0,
		Currency:      "RUB",
	}
	walletIDString := uuid.UUID(createdWallet.WalletID).String()
	depositPath := walletPath + "/" + walletIDString + "/deposit"

	s.sendRequest(http.MethodPut, depositPath, http.StatusOK, &transaction, nil, existingUser)

	s.Run("withdraw successfully", func() {
		transaction = models.Transaction{
			ID:            models.TxID(uuid.New()),
			FirstWalletID: createdWallet.WalletID,
			Money:         500.0,
			Currency:      "RUB",
		}
		uuidString := uuid.UUID(createdWallet.WalletID).String()
		walletIDPath := walletPath + "/" + uuidString + "/withdraw"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusOK, &transaction, nil, existingUser)
	})

	s.Run("zero amount of money", func() {
		transaction = models.Transaction{
			ID:            models.TxID(uuid.New()),
			FirstWalletID: createdWallet.WalletID,
			Money:         0.0,
			Currency:      "RUB",
		}
		uuidString := uuid.UUID(createdWallet.WalletID).String()
		walletIDPath := walletPath + "/" + uuidString + "/withdraw"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusBadRequest, &transaction, nil, existingUser)
	})

	s.Run("negative amount of money", func() {
		transaction = models.Transaction{
			ID:            models.TxID(uuid.New()),
			FirstWalletID: createdWallet.WalletID,
			Money:         -1000.0,
			Currency:      "RUB",
		}
		uuidString := uuid.UUID(createdWallet.WalletID).String()
		walletIDPath := walletPath + "/" + uuidString + "/withdraw"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusBadRequest, &transaction, nil, existingUser)
	})

	s.Run("wrong currency", func() {
		transaction = models.Transaction{
			ID:            models.TxID(uuid.New()),
			FirstWalletID: createdWallet.WalletID,
			Money:         1000.0,
			Currency:      "USD",
		}
		uuidString := uuid.UUID(createdWallet.WalletID).String()
		walletIDPath := walletPath + "/" + uuidString + "/withdraw"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusBadRequest, &transaction, nil, existingUser)
	})

	s.Run("wallet not found", func() {
		transaction = models.Transaction{
			ID:            models.TxID(uuid.New()),
			FirstWalletID: models.WalletID(uuid.New()),
			Money:         1000.0,
			Currency:      "RUB",
		}
		uuidString := uuid.UUID(createdWallet.WalletID).String()
		walletIDPath := walletPath + "/" + uuidString + "/withdraw"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusNotFound, &transaction, nil, existingUser)
	})

	s.Run("user is not the owner of the wallet", func() {
		transaction = models.Transaction{
			ID:            models.TxID(uuid.New()),
			FirstWalletID: createdWallet.WalletID,
			Money:         1000.0,
			Currency:      "RUB",
		}

		userFromAnotherMother := models.User{
			UserID: models.UserID(uuid.New()),
		}

		err = s.db.UpsertUser(context.Background(), userFromAnotherMother)
		s.Require().NoError(err)

		uuidString := uuid.UUID(createdWallet.WalletID).String()
		walletIDPath := walletPath + "/" + uuidString + "/withdraw"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusNotFound, &transaction, nil, userFromAnotherMother)
	})

	s.Run("user not found", func() {
		transaction = models.Transaction{
			ID:            models.TxID(uuid.New()),
			FirstWalletID: createdWallet.WalletID,
			Money:         1000.0,
			Currency:      "RUB",
		}

		newUser := models.User{
			UserID: models.UserID(uuid.New()),
		}

		createdWallet.UserID = newUser.UserID
		uuidString := uuid.UUID(createdWallet.WalletID).String()
		walletIDPath := walletPath + "/" + uuidString + "/withdraw"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusNotFound, &transaction, nil, newUser)
	})
}

func (s *IntegrationTestSuite) TestTransfer() {
	// Arrange
	firstWallet := models.Wallet{
		WalletID: models.WalletID(uuid.New()),
		UserID:   models.UserID(uuid.New()),
		Name:     "proverkaTRANSFER_1",
		Currency: "RUB",
	}

	secondWallet := models.Wallet{
		WalletID: models.WalletID(uuid.New()),
		UserID:   models.UserID(uuid.New()),
		Name:     "proverkaTRANSFER_2",
		Currency: "RUB",
	}

	err := s.db.UpsertUser(context.Background(), existingUser)
	s.Require().NoError(err)

	secondUser := models.User{
		UserID: models.UserID(uuid.New()),
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
		ID:            models.TxID(uuid.New()),
		FirstWalletID: firstCreatedWallet.WalletID,
		Money:         10000.0,
		Currency:      "RUB",
	}
	walletIDString := uuid.UUID(firstCreatedWallet.WalletID).String()
	depositPath := walletPath + "/" + walletIDString + "/deposit"

	s.sendRequest(http.MethodPut, depositPath, http.StatusOK, &transaction, nil, existingUser)

	s.Run("successful transfer", func() {
		transaction = models.Transaction{
			ID:             models.TxID(uuid.New()),
			FirstWalletID:  firstCreatedWallet.WalletID,
			SecondWalletID: &secondCreatedWallet.WalletID,
			Money:          1000.0,
			Currency:       "RUB",
		}
		uuidString := uuid.UUID(firstCreatedWallet.WalletID).String()
		walletIDPath := walletPath + "/" + uuidString + "/transfer"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusOK, &transaction, nil, existingUser)
	})

	s.Run("zero amount of money", func() {
		transaction = models.Transaction{
			ID:             models.TxID(uuid.New()),
			FirstWalletID:  firstCreatedWallet.WalletID,
			SecondWalletID: &secondCreatedWallet.WalletID,
			Money:          0.0,
			Currency:       "RUB",
		}
		uuidString := uuid.UUID(firstCreatedWallet.WalletID).String()
		walletIDPath := walletPath + "/" + uuidString + "/transfer"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusBadRequest, &transaction, nil, existingUser)
	})

	s.Run("negative amount of money", func() {
		transaction = models.Transaction{
			ID:             models.TxID(uuid.New()),
			FirstWalletID:  firstCreatedWallet.WalletID,
			SecondWalletID: &secondCreatedWallet.WalletID,
			Money:          -1000.0,
			Currency:       "RUB",
		}
		uuidString := uuid.UUID(firstCreatedWallet.WalletID).String()
		walletIDPath := walletPath + "/" + uuidString + "/transfer"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusBadRequest, &transaction, nil, existingUser)
	})

	s.Run("wrong currency of first wallet", func() {
		transaction = models.Transaction{
			ID:             models.TxID(uuid.New()),
			FirstWalletID:  firstCreatedWallet.WalletID,
			SecondWalletID: &secondCreatedWallet.WalletID,
			Money:          1000.0,
			Currency:       "USD",
		}
		uuidString := uuid.UUID(firstCreatedWallet.WalletID).String()
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
		updateWalletID := uuid.UUID(secondCreatedWallet.WalletID).String()
		updatePath := walletPath + "/" + updateWalletID

		s.sendRequest(http.MethodPatch, updatePath, http.StatusOK, &updateSecondWallet, &secondCreatedWallet, secondUser)

		transaction = models.Transaction{
			ID:             models.TxID(uuid.New()),
			FirstWalletID:  firstCreatedWallet.WalletID,
			SecondWalletID: &secondCreatedWallet.WalletID,
			Money:          1000.0,
			Currency:       "RUB",
		}

		uuidString := uuid.UUID(firstCreatedWallet.WalletID).String()
		walletIDPath := walletPath + "/" + uuidString + "/transfer"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusOK, &transaction, nil, existingUser)
	})

	s.Run("wallet not found", func() {
		transaction = models.Transaction{
			ID:             models.TxID(uuid.New()),
			FirstWalletID:  models.WalletID(uuid.New()),
			SecondWalletID: &secondCreatedWallet.WalletID,
			Money:          1000.0,
			Currency:       "RUB",
		}
		uuidString := uuid.UUID(firstCreatedWallet.WalletID).String()
		walletIDPath := walletPath + "/" + uuidString + "/transfer"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusNotFound, &transaction, nil, existingUser)
	})

	s.Run("user is not the owner of the wallet", func() {
		transaction = models.Transaction{
			ID:             models.TxID(uuid.New()),
			FirstWalletID:  firstCreatedWallet.WalletID,
			SecondWalletID: &secondCreatedWallet.WalletID,
			Money:          1000.0,
			Currency:       "RUB",
		}

		userFromAnotherMother := models.User{
			UserID: models.UserID(uuid.New()),
		}

		err = s.db.UpsertUser(context.Background(), userFromAnotherMother)
		s.Require().NoError(err)

		uuidString := uuid.UUID(firstCreatedWallet.WalletID).String()
		walletIDPath := walletPath + "/" + uuidString + "/transfer"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusNotFound, &transaction, nil, userFromAnotherMother)
	})

	s.Run("first user not found", func() {
		transaction = models.Transaction{
			ID:             models.TxID(uuid.New()),
			FirstWalletID:  firstCreatedWallet.WalletID,
			SecondWalletID: &secondCreatedWallet.WalletID,
			Money:          1000.0,
			Currency:       "RUB",
		}

		newUser := models.User{
			UserID: models.UserID(uuid.New()),
		}

		firstCreatedWallet.UserID = newUser.UserID
		uuidString := uuid.UUID(firstCreatedWallet.WalletID).String()
		walletIDPath := walletPath + "/" + uuidString + "/transfer"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusNotFound, &transaction, nil, newUser)
	})

	s.Run("second user not found", func() {
		newWalletID := models.WalletID(uuid.New())
		transaction = models.Transaction{
			ID:             models.TxID(uuid.New()),
			FirstWalletID:  firstCreatedWallet.WalletID,
			SecondWalletID: &newWalletID,
			Money:          1000.0,
			Currency:       "RUB",
		}
		uuidString := uuid.UUID(firstCreatedWallet.WalletID).String()
		walletIDPath := walletPath + "/" + uuidString + "/transfer"

		// Act
		s.sendRequest(http.MethodPut, walletIDPath, http.StatusNotFound, &transaction, nil, existingUser)
	})
}

func (s *IntegrationTestSuite) TestGetTransactions() {
	// Arrange
	firstWallet := models.Wallet{
		WalletID: models.WalletID(uuid.New()),
		UserID:   models.UserID(uuid.New()),
		Name:     "proverkaGET_TX_1",
		Currency: "RUB",
	}

	secondWallet := models.Wallet{
		WalletID: models.WalletID(uuid.New()),
		UserID:   models.UserID(uuid.New()),
		Name:     "proverkaGET_TX_2",
		Currency: "RUB",
	}

	err := s.db.UpsertUser(context.Background(), existingUser)
	s.Require().NoError(err)

	secondUser := models.User{
		UserID: models.UserID(uuid.New()),
	}

	err = s.db.UpsertUser(context.Background(), secondUser)
	s.Require().NoError(err)

	firstWallet.UserID = existingUser.UserID
	secondWallet.UserID = secondUser.UserID
	firstCreatedWallet := models.Wallet{}
	secondCreatedWallet := models.Wallet{}

	s.sendRequest(http.MethodPost, walletPath, http.StatusCreated, &firstWallet, &firstCreatedWallet, existingUser)

	s.sendRequest(http.MethodPost, walletPath, http.StatusCreated, &secondWallet, &secondCreatedWallet, secondUser)

	s.Run("empty list", func() {
		var transactions []models.Transaction

		uuidString := uuid.UUID(firstCreatedWallet.WalletID).String()
		walletTxPath := walletPath + "/" + uuidString + "/transactions"
		// Act
		s.sendRequest(http.MethodGet, walletTxPath, http.StatusOK, nil, &transactions, existingUser)

		// Assert
		s.Require().Len(transactions, 0)
	})

	firstTx := models.Transaction{
		ID:            models.TxID(uuid.New()),
		Name:          "deposit",
		FirstWalletID: firstCreatedWallet.WalletID,
		Money:         10000.0,
		Currency:      "RUB",
	}

	uuidString := uuid.UUID(firstCreatedWallet.WalletID).String()
	firstWalletIDPath := walletPath + "/" + uuidString + "/deposit"

	s.sendRequest(http.MethodPut, firstWalletIDPath, http.StatusOK, &firstTx, nil, existingUser)

	secondTx := models.Transaction{
		ID:             models.TxID(uuid.New()),
		Name:           "transfer",
		FirstWalletID:  firstCreatedWallet.WalletID,
		SecondWalletID: &secondCreatedWallet.WalletID,
		Money:          1000.0,
		Currency:       "RUB",
	}

	secWalletIDPath := walletPath + "/" + uuidString + "/transfer"

	s.sendRequest(http.MethodPut, secWalletIDPath, http.StatusOK, &secondTx, nil, existingUser)

	thirdTx := models.Transaction{
		ID:            models.TxID(uuid.New()),
		Name:          "withdraw",
		FirstWalletID: firstCreatedWallet.WalletID,
		Money:         5000.0,
		Currency:      "RUB",
	}

	thirdWalletIDPath := walletPath + "/" + uuidString + "/withdraw"

	s.sendRequest(http.MethodPut, thirdWalletIDPath, http.StatusOK, &thirdTx, nil, existingUser)

	txPath := walletPath + "/" + uuidString + "/transactions"

	s.Run("read successfully", func() {
		var transactions []models.Transaction

		s.sendRequest(http.MethodGet, txPath, http.StatusOK, nil, &transactions, existingUser)

		s.Require().Len(transactions, 3)
	})

	s.Run("descending names read successfully", func() {
		var transactions []models.Transaction
		descTxPath := txPath + "?sorting=name&descending=true"

		// Act
		s.sendRequest(http.MethodGet, descTxPath, http.StatusOK, nil, &transactions, existingUser)

		// Assert
		s.Require().Equal(thirdTx.Name, transactions[0].Name)
		s.Require().Equal(thirdTx.Currency, transactions[0].Currency)

		s.Require().Equal(secondTx.Name, transactions[1].Name)
		s.Require().Equal(secondTx.Currency, transactions[1].Currency)

		s.Require().Equal(firstTx.Name, transactions[2].Name)
		s.Require().Equal(firstTx.Currency, transactions[2].Currency)
	})

	s.Run("filter read successfully", func() {
		var transactions []models.Transaction
		filterTxPath := txPath + "?filter=eur"

		// Act
		s.sendRequest(http.MethodGet, filterTxPath, http.StatusOK, nil, &transactions, existingUser)

		// Assert
		s.Require().Len(transactions, 0)
	})

	s.Run("limit and offset read successfully", func() {
		var transactions []models.Transaction
		limitTxPath := txPath + "?sorting=name&limit=2&offset=2"

		// Act
		s.sendRequest(http.MethodGet, limitTxPath, http.StatusOK, nil, &transactions, existingUser)

		// Assert
		s.Require().Len(transactions, 1)
	})

	s.Run("user is not the owner of the wallet", func() {
		userFromAnotherMother := models.User{
			UserID: models.UserID(uuid.New()),
		}

		err = s.db.UpsertUser(context.Background(), userFromAnotherMother)
		s.Require().NoError(err)

		// Act
		s.sendRequest(http.MethodGet, txPath, http.StatusNotFound, nil, nil, userFromAnotherMother)
	})
}
