package tests

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Memonagi/wallet_project/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *IntegrationTestSuite) TestWalletCleanup() {
	// Arrange
	ctx := context.Background()

	err := s.db.UpsertUser(ctx, existingUser)
	s.Require().NoError(err)

	inactiveTime := time.Date(2022, 7, 26, 14, 0, 0, 0, time.UTC)
	validTime := time.Now().AddDate(0, -6, 0)

	firstWallet := models.Wallet{
		UserID:    existingUser.UserID,
		Name:      "Inactive Wallet",
		Currency:  "RUB",
		Archived:  false,
		CreatedAt: inactiveTime.AddDate(0, -1, 0),
		UpdatedAt: inactiveTime,
	}

	secondWallet := models.Wallet{
		UserID:    existingUser.UserID,
		Name:      "Valid Wallet",
		Currency:  "USD",
		Archived:  false,
		CreatedAt: validTime.AddDate(-1, 0, 0),
		UpdatedAt: validTime,
	}

	thirdWallet := models.Wallet{
		UserID:    existingUser.UserID,
		Name:      "New Wallet",
		Currency:  "EUR",
		Archived:  false,
		CreatedAt: time.Now().AddDate(0, 0, -3),
		UpdatedAt: time.Now(),
	}

	inactiveWallet, err := s.createWallet(ctx, firstWallet, existingUser.UserID)
	s.Require().NoError(err)

	validWallet, err := s.createWallet(ctx, secondWallet, existingUser.UserID)
	s.Require().NoError(err)

	newWallet, err := s.createWallet(ctx, thirdWallet, existingUser.UserID)
	s.Require().NoError(err)

	// Act
	err = s.db.WalletCleaner(ctx)
	s.Require().NoError(err)

	inactiveWalletUpdate, err := s.getWalletArchived(ctx, inactiveWallet.WalletID, existingUser.UserID)
	s.Require().NoError(err)

	validWalletUpdate, err := s.getWalletArchived(ctx, validWallet.WalletID, existingUser.UserID)
	s.Require().NoError(err)

	newWalletUpdate, err := s.getWalletArchived(ctx, newWallet.WalletID, existingUser.UserID)
	s.Require().NoError(err)

	// Assert
	s.Require().True(inactiveWalletUpdate.Archived)
	s.Require().False(validWalletUpdate.Archived)
	s.Require().False(newWalletUpdate.Archived)
}

func (s *IntegrationTestSuite) createWallet(ctx context.Context, wallet models.Wallet,
	userID models.UserID,
) (models.Wallet, error) {
	query := `INSERT INTO wallets 
    (id, user_id, name, currency, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, user_id, name, currency, balance, archived, created_at, updated_at`

	err := s.db.QueryRowFunc(ctx, query, models.WalletID(uuid.New()), userID, wallet.Name, wallet.Currency,
		wallet.CreatedAt, wallet.UpdatedAt).Scan(
		&wallet.WalletID,
		&wallet.UserID,
		&wallet.Name,
		&wallet.Currency,
		&wallet.Balance,
		&wallet.Archived,
		&wallet.CreatedAt,
		&wallet.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ForeignKeyViolation {
			return models.Wallet{}, models.ErrUserNotFound
		}

		return models.Wallet{}, fmt.Errorf("failed to create wallet: %w", err)
	}

	return wallet, nil
}

func (s *IntegrationTestSuite) getWalletArchived(ctx context.Context, walletID models.WalletID, userID models.UserID) (models.Wallet, error) {
	var wallet models.Wallet

	query := `SELECT archived FROM wallets WHERE id = $1 AND user_id = $2`

	err := s.db.QueryRowFunc(ctx, query, walletID, userID).Scan(&wallet.Archived)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Wallet{}, fmt.Errorf("failed to read wallet info: %w", models.ErrWalletNotFound)
		}

		return models.Wallet{}, fmt.Errorf("failed to read wallet info: %w", err)
	}

	return wallet, nil
}
