package service

import (
	"context"
	"fmt"

	"github.com/Memonagi/wallet_project/internal/models"
	"github.com/google/uuid"
)

type wallets interface {
	CreateWallet(ctx context.Context, wallet models.Wallet, userID uuid.UUID) (models.Wallet, error)
	GetWallet(ctx context.Context, walletID, userID uuid.UUID, wallet models.Wallet) (models.Wallet, error)
	UpdateWallet(ctx context.Context, walletID, userID uuid.UUID, wallet models.WalletUpdate,
		rate float64) (models.Wallet, error)
	DeleteWallet(ctx context.Context, walletID, userID uuid.UUID) error
	GetWallets(ctx context.Context, request models.GetWalletsRequest, userID uuid.UUID) ([]models.Wallet, error)
	GetCurrency(ctx context.Context, walletID uuid.UUID) (models.WalletUpdate, error)
	Deposit(ctx context.Context, userID uuid.UUID, transaction models.Transaction) error
	WithdrawMoney(ctx context.Context, userID uuid.UUID, transaction models.Transaction) error
	Transfer(ctx context.Context, userID uuid.UUID, transaction models.Transaction, rate float64) error
}

type xrClient interface {
	GetRate(ctx context.Context, from, to string) (float64, error)
}

type Service struct {
	wallets  wallets
	xrClient xrClient
}

func New(wallets wallets, xrClient xrClient) *Service {
	return &Service{
		wallets:  wallets,
		xrClient: xrClient,
	}
}

func (s *Service) CreateWallet(ctx context.Context, wallet models.Wallet, userID uuid.UUID) (models.Wallet, error) {
	if err := wallet.Validate(); err != nil {
		return models.Wallet{}, fmt.Errorf("%w", err)
	}

	var (
		newWallet models.Wallet
		err       error
	)

	if wallet.UserID != userID {
		return models.Wallet{}, fmt.Errorf("%w", models.ErrWrongUserID)
	}

	if newWallet, err = s.wallets.CreateWallet(ctx, wallet, userID); err != nil {
		return models.Wallet{}, fmt.Errorf("failed create new wallet: %w", err)
	}

	return newWallet, nil
}

func (s *Service) GetWallet(ctx context.Context, walletID, userID uuid.UUID) (models.Wallet, error) {
	if walletID == uuid.Nil {
		return models.Wallet{}, fmt.Errorf("%w", models.ErrEmptyID)
	}

	if userID == uuid.Nil {
		return models.Wallet{}, fmt.Errorf("%w", models.ErrUserID)
	}

	var wallet models.Wallet

	walletInfo, err := s.wallets.GetWallet(ctx, walletID, userID, wallet)
	if err != nil {
		return models.Wallet{}, fmt.Errorf("wallet not found: %w", err)
	}

	return walletInfo, nil
}

func (s *Service) UpdateWallet(ctx context.Context, walletID, userID uuid.UUID,
	wallet models.WalletUpdate,
) (models.Wallet, error) {
	if err := wallet.Validate(); err != nil {
		return models.Wallet{}, fmt.Errorf("%w", err)
	}

	var (
		updatedWallet models.Wallet
		baseWallet    models.WalletUpdate
		err           error
		rate          = 1.00
	)

	baseWallet, err = s.wallets.GetCurrency(ctx, walletID)
	if err != nil {
		return models.Wallet{}, fmt.Errorf("wallet not found: %w", err)
	}

	if baseWallet.Currency != wallet.Currency {
		rate, err = s.xrClient.GetRate(ctx, *baseWallet.Currency, *wallet.Currency)
		if err != nil {
			return models.Wallet{}, fmt.Errorf("failed get rate: %w", err)
		}
	}

	baseWallet.Currency = wallet.Currency
	baseWallet.Name = wallet.Name

	if updatedWallet, err = s.wallets.UpdateWallet(ctx, walletID, userID, baseWallet, rate); err != nil {
		return models.Wallet{}, fmt.Errorf("failed update wallet info: %w", err)
	}

	return updatedWallet, nil
}

func (s *Service) DeleteWallet(ctx context.Context, walletID, userID uuid.UUID) error {
	if walletID == uuid.Nil {
		return fmt.Errorf("%w", models.ErrEmptyID)
	}

	if err := s.wallets.DeleteWallet(ctx, walletID, userID); err != nil {
		return fmt.Errorf("failed delete wallet: %w", err)
	}

	return nil
}

func (s *Service) GetWallets(ctx context.Context, request models.GetWalletsRequest,
	userID uuid.UUID,
) ([]models.Wallet, error) {
	wallets, err := s.wallets.GetWallets(ctx, request, userID)
	if err != nil {
		return nil, fmt.Errorf("failed get all wallets: %w", err)
	}

	return wallets, nil
}

func (s *Service) Deposit(ctx context.Context, userID uuid.UUID, transaction models.Transaction) error {
	if err := transaction.Validate(); err != nil {
		return fmt.Errorf("error validating transaction: %w", err)
	}

	if err := s.wallets.Deposit(ctx, userID, transaction); err != nil {
		return fmt.Errorf("failed deposit: %w", err)
	}

	return nil
}

func (s *Service) WithdrawMoney(ctx context.Context, userID uuid.UUID, transaction models.Transaction) error {
	if err := transaction.Validate(); err != nil {
		return fmt.Errorf("error validating transaction: %w", err)
	}

	if err := s.wallets.WithdrawMoney(ctx, userID, transaction); err != nil {
		return fmt.Errorf("failed withdraw money: %w", err)
	}

	return nil
}

func (s *Service) Transfer(ctx context.Context, userID uuid.UUID, transaction models.Transaction) error {
	if err := transaction.Validate(); err != nil {
		return fmt.Errorf("error validating transaction: %w", err)
	}

	if transaction.SecondWalletID == uuid.Nil {
		return fmt.Errorf("%w", models.ErrEmptyID)
	}

	secondWallet, err := s.wallets.GetCurrency(ctx, transaction.SecondWalletID)
	if err != nil {
		return fmt.Errorf("failed to get second wallet: %w", err)
	}

	rate := 1.00

	if secondWallet.Currency != &transaction.Currency {
		rate, err = s.xrClient.GetRate(ctx, transaction.Currency, *secondWallet.Currency)
		if err != nil {
			return fmt.Errorf("failed get rate: %w", err)
		}
	}

	if err = s.wallets.Transfer(ctx, userID, transaction, rate); err != nil {
		return fmt.Errorf("failed transfer transaction: %w", err)
	}

	return nil
}
