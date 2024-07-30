package database

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Memonagi/wallet_project/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Store) CreateWallet(ctx context.Context, wallet models.Wallet) (models.Wallet, error) {
	query := `INSERT INTO wallets 
    (id, user_id, name, currency)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, name, currency, balance, archived, created_at, updated_at`

	err := s.db.QueryRow(ctx, query, uuid.New(), wallet.UserID, wallet.Name, wallet.Currency).Scan(
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

func (s *Store) GetWallet(ctx context.Context, walletID uuid.UUID,
	wallet models.Wallet,
) (models.Wallet, error) {
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
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Wallet{}, fmt.Errorf("failed to read wallet info: %w", models.ErrWalletNotFound)
		}

		return models.Wallet{}, fmt.Errorf("failed to read wallet info: %w", err)
	}

	return wallet, nil
}

func (s *Store) UpdateWallet(ctx context.Context, walletID uuid.UUID,
	wallet models.WalletUpdate,
) (models.Wallet, error) {
	var (
		sb           strings.Builder
		args         []any
		updateWallet = models.Wallet{}
	)

	sb.WriteString("UPDATE wallets SET ")

	if wallet.Name != nil {
		sb.WriteString(fmt.Sprintf("name = $%d, ", len(args)))
		args = append(args, wallet.Name)
	}

	if wallet.Currency != nil {
		if len(args) > 0 {
			sb.WriteString(", ")
		}

		sb.WriteString(fmt.Sprintf("currency = $%d, ", len(args)))
		args = append(args, wallet.Currency)
	}

	if len(args) == 0 {
		return s.GetWallet(ctx, walletID, updateWallet)
	}

	sb.WriteString(fmt.Sprintf("updated_at = NOW() WHERE id = $%d AND archived = false", len(args)))
	args = append(args, walletID)

	sb.WriteString(" RETURNING id, user_id, name, currency, balance, archived, created_at, updated_at")

	query := sb.String()

	err := s.db.QueryRow(ctx, query, args...).Scan(
		&updateWallet.WalletID,
		&updateWallet.UserID,
		&updateWallet.Name,
		&updateWallet.Currency,
		&updateWallet.Balance,
		&updateWallet.Archived,
		&updateWallet.CreatedAt,
		&updateWallet.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Wallet{}, fmt.Errorf("failed to update wallet info: %w", models.ErrWalletNotFound)
		}

		return models.Wallet{}, fmt.Errorf("failed to update wallet info: %w", err)
	}

	return updateWallet, nil
}

func (s *Store) DeleteWallet(ctx context.Context, walletID uuid.UUID, wallet models.Wallet) error {
	query := `UPDATE wallets SET
	archived = $2, 
	updated_at = NOW()
	WHERE id = $1 AND archived = false`

	if _, err := s.db.Exec(ctx, query, walletID, wallet.Archived); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("failed to delete wallet: %w", models.ErrWalletNotFound)
		}

		return fmt.Errorf("failed to delete wallet: %w", err)
	}

	return nil
}
