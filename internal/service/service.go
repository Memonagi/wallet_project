package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Memonagi/wallet_project/internal/models"
	"github.com/google/uuid"
)

type wallets interface {
	CreateWallet(ctx context.Context, wallet models.WalletInfo) error
	ReadWalletInfo(ctx context.Context, walletID uuid.UUID, wallet models.WalletInfo) (models.WalletInfo, error)
	UpdateWalletInfo(ctx context.Context, wallet models.WalletInfo) error
	DeleteWallet(ctx context.Context, walletID uuid.UUID, wallet models.WalletInfo) error
}

type Service struct {
	wallets wallets
}

var (
	errName     = errors.New("wallet name is empty")
	errCurrency = errors.New("wallet currency is invalid")
	errID       = errors.New("wallet ID is empty")
)

func New(wallets wallets) *Service {
	return &Service{wallets: wallets}
}

func (s *Service) CreateWallet(ctx context.Context, wallet models.WalletInfo) error {
	if wallet.Name == "" {
		return fmt.Errorf("%w", errName)
	}

	if !s.checkCurrency(wallet) {
		return fmt.Errorf("%w", errCurrency)
	}

	if err := s.wallets.CreateWallet(ctx, wallet); err != nil {
		return fmt.Errorf("failed create new wallet: %w", err)
	}

	return nil
}

func (s *Service) ReadWalletInfo(ctx context.Context, walletID uuid.UUID) (models.WalletInfo, error) {
	if walletID == uuid.Nil {
		return models.WalletInfo{}, fmt.Errorf("%w", errID)
	}

	var wallet models.WalletInfo

	walletInfo, err := s.wallets.ReadWalletInfo(ctx, walletID, wallet)
	if err != nil {
		return models.WalletInfo{}, fmt.Errorf("wallet not found: %w", err)
	}

	return walletInfo, nil
}

func (s *Service) UpdateWalletInfo(ctx context.Context, wallet models.WalletInfo) error {
	if wallet.Name == "" {
		return fmt.Errorf("%w", errName)
	}

	if !s.checkCurrency(wallet) {
		return fmt.Errorf("%w", errCurrency)
	}

	if err := s.wallets.UpdateWalletInfo(ctx, wallet); err != nil {
		return fmt.Errorf("failed update wallet info: %w", err)
	}

	return nil
}

func (s *Service) DeleteWallet(ctx context.Context, walletID uuid.UUID) error {
	if walletID == uuid.Nil {
		return fmt.Errorf("%w", errID)
	}

	var wallet models.WalletInfo

	if err := s.wallets.DeleteWallet(ctx, walletID, wallet); err != nil {
		return fmt.Errorf("failed delete wallet: %w", err)
	}

	return nil
}

func (s *Service) checkCurrency(wallet models.WalletInfo) bool {
	currencies := map[string]struct{}{
		"USD": {},
		"EUR": {},
		"RUB": {},
		"JPY": {},
		"CNY": {},
		"CAD": {},
		"AUD": {},
	}

	_, ok := currencies[strings.ToUpper(wallet.Currency)]

	return ok
}
