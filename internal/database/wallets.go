package database

import (
	"context"
	"fmt"

	"github.com/Memonagi/wallet_project/internal/models"
	"github.com/google/uuid"
)

func (s *Store) CreateWallet(ctx context.Context, wallet models.WalletInfo) error {
	query := `INSERT INTO wallets 
    (id, user_id, name, currency, balance, archived, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := s.db.Exec(
		ctx,
		query,
		wallet.WalletID,
		wallet.UserID,
		wallet.Name,
		wallet.Currency,
		wallet.Balance,
		wallet.Archived,
		wallet.CreatedAt,
		wallet.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create wallet: %w", err)
	}

	return nil
}

func (s *Store) ReadWalletInfo(ctx context.Context, walletID uuid.UUID,
	wallet models.WalletInfo,
) (models.WalletInfo, error) {
	query := `SELECT * FROM wallets WHERE id = $1 AND archived = false`

	err := s.db.QueryRow(ctx, query, walletID).Scan(
		&wallet.WalletID,
		&wallet.UserID,
		&wallet.Name,
		&wallet.Currency,
		&wallet.Balance,
		&wallet.Archived,
		&wallet.CreatedAt,
		&wallet.UpdatedAt)
	if err != nil {
		return models.WalletInfo{}, fmt.Errorf("failed to read wallet info: %w", err)
	}

	return wallet, nil
}

func (s *Store) UpdateWalletInfo(ctx context.Context, wallet models.WalletInfo) error {
	query := `UPDATE wallets SET
	name = $2, 
	currency = $3,
	updated_at = NOW()
	WHERE id = $1 AND archived = false`

	if _, err := s.db.Exec(ctx, query, wallet.WalletID, wallet.Name, wallet.Currency); err != nil {
		return fmt.Errorf("failed to update wallet info: %w", err)
	}

	return nil
}

func (s *Store) DeleteWallet(ctx context.Context, walletID uuid.UUID, wallet models.WalletInfo) error {
	query := `UPDATE wallets SET
	archived = $2, 
	updated_at = NOW()
	WHERE id = $1`

	if _, err := s.db.Exec(ctx, query, walletID, wallet.Archived); err != nil {
		return fmt.Errorf("failed to delete wallet: %w", err)
	}

	return nil
}
