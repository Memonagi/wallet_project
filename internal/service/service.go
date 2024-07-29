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
	UpdateWallet(ctx context.Context, walletID uuid.UUID, wallet models.WalletUpdate) (models.Wallet, error)
	DeleteWallet(ctx context.Context, walletID uuid.UUID, wallet models.Wallet) error
}

type Service struct {
	wallets wallets
}

func New(wallets wallets) *Service {
	return &Service{wallets: wallets}
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
		updateWallet models.Wallet
		err          error
	)

	if updateWallet, err = s.wallets.UpdateWallet(ctx, walletID, wallet); err != nil {
		return models.Wallet{}, fmt.Errorf("failed update wallet info: %w", err)
	}

	return updateWallet, nil
}

func (s *Service) DeleteWallet(ctx context.Context, walletID uuid.UUID) error {
	if walletID == uuid.Nil {
		return fmt.Errorf("%w", models.ErrEmptyID)
	}

	var wallet models.Wallet

	if err := s.wallets.DeleteWallet(ctx, walletID, wallet); err != nil {
		return fmt.Errorf("failed delete wallet: %w", err)
	}

	return nil
}
