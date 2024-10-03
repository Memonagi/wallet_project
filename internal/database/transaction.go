package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Memonagi/wallet_project/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sirupsen/logrus"
)

type txProducer interface {
	ProduceTx(key, value string) error
}

func (s *Store) getWalletTx(ctx context.Context, walletID, userID uuid.UUID, dbTx pgx.Tx) (models.Wallet, error) {
	var wallet models.Wallet

	query := `SELECT id, user_id, name, currency, balance, archived, created_at, updated_at 
FROM wallets WHERE id = $1 AND user_id = $2 AND archived = false FOR UPDATE `

	err := dbTx.QueryRow(ctx, query, walletID, userID).Scan(
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

func (s *Store) createTx(ctx context.Context, transaction models.Transaction, dbTx pgx.Tx) error {
	query := `INSERT INTO transactions 
    (id, name, first_wallet, second_wallet, currency, money) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	args := []any{
		uuid.New(),
		transaction.Name,
		transaction.FirstWalletID,
		nil,
		transaction.Currency,
		transaction.Money,
	}

	if transaction.SecondWalletID != uuid.Nil {
		args[3] = transaction.SecondWalletID
	}

	err := dbTx.QueryRow(ctx, query, args...).Scan(&transaction.ID)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ForeignKeyViolation {
			return models.ErrWalletNotFound
		}

		return fmt.Errorf("failed to save history of transaction in database: %w", err)
	}

	txJSON, err := json.Marshal(transaction)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %w", err)
	}

	if err = s.producer.ProduceTx("", string(txJSON)); err != nil {
		return fmt.Errorf("failed to produce transaction: %w", err)
	}

	return nil
}

func (s *Store) Deposit(ctx context.Context, userID uuid.UUID, transaction models.Transaction) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err = tx.Rollback(ctx); err != nil && errors.Is(err, pgx.ErrTxClosed) {
			logrus.Warnf("failed to rollback transaction: %v", err)
		}
	}()

	var wallet models.Wallet

	wallet, err = s.getWalletTx(ctx, transaction.FirstWalletID, userID, tx)
	if err != nil {
		return fmt.Errorf("failed to get wallet: %w", err)
	}

	if wallet.Currency != transaction.Currency {
		return fmt.Errorf("%w", models.ErrWrongCurrency)
	}

	query := `UPDATE wallets 
SET balance = balance + $3, updated_at = NOW() WHERE id = $1 AND user_id = $2 AND archived = false`

	res, err := tx.Exec(ctx, query, transaction.FirstWalletID, userID, transaction.Money)
	if err != nil {
		return fmt.Errorf("failed to update wallet info: %w", err)
	}

	if res.RowsAffected() == 0 {
		return models.ErrWalletNotFound
	}

	transaction.Name = "deposit"

	if err = s.createTx(ctx, transaction, tx); err != nil {
		return fmt.Errorf("failed to save history of transaction: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

//nolint:cyclop
func (s *Store) WithdrawMoney(ctx context.Context, userID uuid.UUID, transaction models.Transaction) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err = tx.Rollback(ctx); err != nil && errors.Is(err, pgx.ErrTxClosed) {
			logrus.Warnf("failed to rollback transaction: %v", err)
		}
	}()

	var wallet models.Wallet

	wallet, err = s.getWalletTx(ctx, transaction.FirstWalletID, userID, tx)
	if err != nil {
		return fmt.Errorf("failed to get wallet: %w", err)
	}

	switch {
	case wallet.Currency != transaction.Currency:
		return fmt.Errorf("%w", models.ErrWrongCurrency)
	case wallet.Balance < transaction.Money:
		return fmt.Errorf("%w", models.ErrInsufficientFunds)
	}

	query := `UPDATE wallets 
SET balance = balance - $3, updated_at = NOW() WHERE id = $1 AND user_id = $2 AND archived = false`

	res, err := tx.Exec(ctx, query, transaction.FirstWalletID, userID, transaction.Money)
	if err != nil {
		return fmt.Errorf("failed to update wallet info: %w", err)
	}

	if res.RowsAffected() == 0 {
		return models.ErrWalletNotFound
	}

	transaction.Name = "withdraw"

	if err = s.createTx(ctx, transaction, tx); err != nil {
		return fmt.Errorf("failed to save history of transaction: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

//nolint:cyclop
func (s *Store) Transfer(ctx context.Context, userID uuid.UUID, transaction models.Transaction, rate float64) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err = tx.Rollback(ctx); err != nil && errors.Is(err, pgx.ErrTxClosed) {
			logrus.Warnf("failed to rollback transaction: %v", err)
		}
	}()

	var wallet models.Wallet

	wallet, err = s.getWalletTx(ctx, transaction.FirstWalletID, userID, tx)
	if err != nil {
		return fmt.Errorf("failed to get wallet: %w", err)
	}

	if err = currencyBalanceCheck(wallet, transaction); err != nil {
		return fmt.Errorf("failed to check wallet balance: %w", err)
	}

	firstQuery := `UPDATE wallets 
SET balance = balance - $3, updated_at = NOW() WHERE id = $1 AND user_id = $2 AND archived = false`

	firstRow, err := tx.Exec(ctx, firstQuery, transaction.FirstWalletID, userID, transaction.Money)
	if err != nil {
		return fmt.Errorf("failed to update wallet info: %w", err)
	}

	secondQuery := `UPDATE wallets 
SET balance = balance + ($2::numeric * $3::numeric), updated_at = NOW() WHERE id = $1 AND archived = false`

	secondRow, err := tx.Exec(ctx, secondQuery, transaction.SecondWalletID, transaction.Money, rate)
	if err != nil {
		return fmt.Errorf("failed to update wallet info: %w", err)
	}

	if err = checkRow(firstRow, secondRow); err != nil {
		return fmt.Errorf("nothing changed: %w", err)
	}

	transaction.Name = "transfer"

	if err = s.createTx(ctx, transaction, tx); err != nil {
		return fmt.Errorf("failed to save history of transaction: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func currencyBalanceCheck(wallet models.Wallet, transaction models.Transaction) error {
	switch {
	case wallet.Currency != transaction.Currency:
		return fmt.Errorf("%w", models.ErrWrongCurrency)
	case wallet.Balance < transaction.Money:
		return fmt.Errorf("%w", models.ErrInsufficientFunds)
	}

	return nil
}

func checkRow(firstRow, secondRow pgconn.CommandTag) error {
	if firstRow.RowsAffected() == 0 || secondRow.RowsAffected() == 0 {
		return models.ErrWalletNotFound
	}

	return nil
}
