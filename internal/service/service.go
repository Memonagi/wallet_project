package service

import (
	"context"
	"fmt"

	"github.com/Memonagi/wallet_project/internal/models"
	"github.com/google/uuid"
)

type wallets interface {
	CreateWallet(ctx context.Context, wallet models.Wallet) (models.Wallet, error)
	GetWallet(ctx context.Context, walletID uuid.UUID, wallet models.Wallet) (models.Wallet, error)
	UpdateWallet(ctx context.Context, walletID uuid.UUID, wallet models.WalletUpdate, rate float64) (models.Wallet, error)
	DeleteWallet(ctx context.Context, walletID uuid.UUID) error
	GetWallets(ctx context.Context, request models.GetWalletsRequest, userID uuid.UUID) ([]models.Wallet, error)
	GetCurrency(ctx context.Context, walletID uuid.UUID) (models.WalletUpdate, error)
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

func (s *Service) CreateWallet(ctx context.Context, wallet models.Wallet) (models.Wallet, error) {
	if err := wallet.Validate(); err != nil {
		return models.Wallet{}, fmt.Errorf("%w", err)
	}

	var (
		newWallet models.Wallet
		err       error
	)

	if newWallet, err = s.wallets.CreateWallet(ctx, wallet); err != nil {
		return models.Wallet{}, fmt.Errorf("failed create new wallet: %w", err)
	}

	return newWallet, nil
}

func (s *Service) GetWallet(ctx context.Context, walletID uuid.UUID) (models.Wallet, error) {
	if walletID == uuid.Nil {
		return models.Wallet{}, fmt.Errorf("%w", models.ErrEmptyID)
	}

	var wallet models.Wallet

	walletInfo, err := s.wallets.GetWallet(ctx, walletID, wallet)
	if err != nil {
		return models.Wallet{}, fmt.Errorf("wallet not found: %w", err)
	}

	return walletInfo, nil
}

func (s *Service) UpdateWallet(ctx context.Context, walletID uuid.UUID,
	wallet models.WalletUpdate,
) (models.Wallet, error) {
	if err := wallet.Validate(); err != nil {
		return models.Wallet{}, fmt.Errorf("%w", err)
	}

	var (
		updatedWallet models.Wallet
		err           error
		rate          = 1.00
	)

	baseWallet, err := s.wallets.GetCurrency(ctx, walletID)
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

	if updatedWallet, err = s.wallets.UpdateWallet(ctx, walletID, baseWallet, rate); err != nil {
		return models.Wallet{}, fmt.Errorf("failed update wallet info: %w", err)
	}

	return updatedWallet, nil
}

func (s *Service) DeleteWallet(ctx context.Context, walletID uuid.UUID) error {
	if walletID == uuid.Nil {
		return fmt.Errorf("%w", models.ErrEmptyID)
	}

	if err := s.wallets.DeleteWallet(ctx, walletID); err != nil {
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
