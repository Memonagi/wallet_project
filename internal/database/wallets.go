package database

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Memonagi/wallet_project/internal/models"
	"github.com/Memonagi/wallet_project/internal/server"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sirupsen/logrus"
)

func (s *Store) CreateWallet(ctx context.Context, wallet models.Wallet, userID uuid.UUID) (models.Wallet, error) {
	query := `INSERT INTO wallets 
    (id, user_id, name, currency)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, name, currency, balance, archived, created_at, updated_at`

	err := s.db.QueryRow(ctx, query, uuid.New(), userID, wallet.Name, wallet.Currency).Scan(
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

func (s *Store) GetWallet(ctx context.Context, walletID, userID uuid.UUID,
	wallet models.Wallet,
) (models.Wallet, error) {
	query := `SELECT id, user_id, name, currency, balance, archived, created_at, updated_at 
FROM wallets WHERE id = $1 AND user_id = $2 AND archived = false`

	err := s.db.QueryRow(ctx, query, walletID, userID).Scan(
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

func (s *Store) UpdateWallet(ctx context.Context, walletID, userID uuid.UUID,
	wallet models.WalletUpdate, rate float64,
) (models.Wallet, error) {
	var (
		sb            strings.Builder
		args          []any
		updatedWallet = models.Wallet{}
	)

	sb.WriteString("UPDATE wallets SET ")

	if wallet.Name != nil {
		args = append(args, wallet.Name)
		sb.WriteString(fmt.Sprintf("name = $%d, ", len(args)))
	}

	if wallet.Currency != nil {
		args = append(args, wallet.Currency)
		sb.WriteString(fmt.Sprintf("currency = $%d, ", len(args)))
	}

	if rate != 1 {
		args = append(args, rate)
		sb.WriteString(fmt.Sprintf("balance = $%d * balance, ", len(args)))
	}

	if len(args) == 0 {
		return s.GetWallet(ctx, walletID, userID, updatedWallet)
	}

	args = append(args, walletID, userID)
	sb.WriteString(fmt.Sprintf(`updated_at = NOW() 
WHERE id = $%d AND user_id = $%d AND archived = false`, len(args)-1, len(args)))

	sb.WriteString(" RETURNING id, user_id, name, currency, balance, archived, created_at, updated_at")

	query := sb.String()

	err := s.db.QueryRow(ctx, query, args...).Scan(
		&updatedWallet.WalletID,
		&updatedWallet.UserID,
		&updatedWallet.Name,
		&updatedWallet.Currency,
		&updatedWallet.Balance,
		&updatedWallet.Archived,
		&updatedWallet.CreatedAt,
		&updatedWallet.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Wallet{}, fmt.Errorf("failed to update wallet info: %w", models.ErrWalletNotFound)
		}

		return models.Wallet{}, fmt.Errorf("failed to update wallet info: %w", err)
	}

	return updatedWallet, nil
}

func (s *Store) DeleteWallet(ctx context.Context, walletID, userID uuid.UUID) error {
	query := `UPDATE wallets SET
	archived = true, 
	updated_at = NOW()
	WHERE id = $1 AND user_id = $2 AND archived = false`

	res, err := s.db.Exec(ctx, query, walletID, userID)
	if err != nil || res.RowsAffected() == 0 {
		return fmt.Errorf("failed to delete wallet: %w", models.ErrWalletNotFound)
	}

	return nil
}

func (s *Store) GetWallets(ctx context.Context, request models.GetWalletsRequest,
	userID uuid.UUID,
) ([]models.Wallet, error) {
	var (
		wallets []models.Wallet
		rows    pgx.Rows
		err     error
	)

	query, args := s.getWalletsQuery(request, userID)
	if rows, err = s.db.Query(ctx, query, args...); err != nil {
		return nil, fmt.Errorf("failed to get wallets: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var wallet models.Wallet
		if err = rows.Scan(
			&wallet.WalletID,
			&wallet.UserID,
			&wallet.Name,
			&wallet.Currency,
			&wallet.Balance,
			&wallet.Archived,
			&wallet.CreatedAt,
			&wallet.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan wallets: %w", err)
		}

		wallets = append(wallets, wallet)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to get wallets: %w", err)
	}

	if len(wallets) == 0 {
		return []models.Wallet{}, nil
	}

	return wallets, nil
}

func (s *Store) getWalletsQuery(request models.GetWalletsRequest, userID uuid.UUID) (string, []any) {
	var (
		sb             strings.Builder
		args           []any
		validSortParam = map[string]string{
			"id":         "id",
			"name":       "name",
			"currency":   "currency",
			"balance":    "balance",
			"created_at": "created_at",
			"updated_at": "updated_at",
		}
	)

	sb.WriteString(`SELECT id, user_id, name, currency, balance, archived, created_at, updated_at 
FROM wallets WHERE archived = false`)

	args = append(args, userID)
	sb.WriteString(fmt.Sprintf(` AND user_id = $%d`, len(args)))

	if request.Filter != "" {
		args = append(args, "%"+request.Filter+"%")
		sb.WriteString(fmt.Sprintf(` AND concat_ws(' ', id, name, currency, balance, created_at, updated_at) 
ILIKE $%d`, len(args)))
	}

	sorting, ok := validSortParam[request.Sorting]

	if !ok {
		sorting = "id"
	}

	sb.WriteString(" ORDER BY " + sorting)

	if request.Descending {
		sb.WriteString(" DESC")
	}

	if request.Limit == 0 {
		request.Limit = server.DefaultLimit
	}

	args = append(args, request.Limit)
	sb.WriteString(fmt.Sprintf(" LIMIT $%d", len(args)))

	if request.Offset > 0 {
		args = append(args, request.Offset)
		sb.WriteString(fmt.Sprintf(" OFFSET $%d", len(args)))
	}

	return sb.String(), args
}

func (s *Store) GetCurrency(ctx context.Context, walletID uuid.UUID) (models.WalletUpdate, error) {
	var wallet models.WalletUpdate

	query := `SELECT currency FROM wallets WHERE id = $1 AND archived = false`

	err := s.db.QueryRow(ctx, query, walletID).Scan(&wallet.Currency)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.WalletUpdate{}, fmt.Errorf("failed to read wallet info: %w", models.ErrWalletNotFound)
		}

		return models.WalletUpdate{}, fmt.Errorf("failed to read wallet info: %w", err)
	}

	return wallet, nil
}

func (s *Store) getWalletTx(ctx context.Context, walletID, userID uuid.UUID) (models.Wallet, error) {
	var wallet models.Wallet

	query := `SELECT id, user_id, name, currency, balance, archived, created_at, updated_at 
FROM wallets WHERE id = $1 AND user_id = $2 AND archived = false FOR UPDATE `

	err := s.db.QueryRow(ctx, query, walletID, userID).Scan(
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

	wallet, err = s.getWalletTx(ctx, transaction.FirstWalletID, userID)
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

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

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

	wallet, err = s.getWalletTx(ctx, transaction.FirstWalletID, userID)
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

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

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

	wallet, err = s.getWalletTx(ctx, transaction.FirstWalletID, userID)
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
