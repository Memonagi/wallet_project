package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Memonagi/wallet_project/internal/models"
	"github.com/google/uuid"
)

//nolint:interfacebloat
type wallets interface {
	CreateWallet(ctx context.Context, wallet models.Wallet, userID models.UserID) (models.Wallet, error)
	GetWallet(ctx context.Context, walletID models.WalletID, userID models.UserID,
		wallet models.Wallet) (models.Wallet, error)
	UpdateWallet(ctx context.Context, walletID models.WalletID, userID models.UserID, wallet models.WalletUpdate,
		rate float64) (models.Wallet, error)
	DeleteWallet(ctx context.Context, walletID models.WalletID, userID models.UserID) error
	GetWallets(ctx context.Context, request models.GetWalletsRequest, userID models.UserID) ([]models.Wallet, error)
	GetCurrency(ctx context.Context, walletID models.WalletID) (models.WalletUpdate, error)
	Deposit(ctx context.Context, userID models.UserID, transaction models.Transaction) error
	WithdrawMoney(ctx context.Context, userID models.UserID, transaction models.Transaction) error
	Transfer(ctx context.Context, userID models.UserID, transaction models.Transaction, rate float64) error
	GetTransactions(ctx context.Context, request models.GetWalletsRequest,
		walletID models.WalletID) ([]models.Transaction, error)
	WalletCleaner(ctx context.Context) error
}

type xrClient interface {
	GetRate(ctx context.Context, from, to string) (float64, error)
}

//go:generate mockgen -source=service.go -destination=../mocks/mock_txproducer.gen.go -package=mocks txProducer
type txProducer interface {
	ProduceTx(key, value string) error
}

type Service struct {
	wallets  wallets
	xrClient xrClient
	producer txProducer
}

const cleanupTicker = 24 * time.Hour

func New(wallets wallets, xrClient xrClient, producer txProducer) *Service {
	return &Service{
		wallets:  wallets,
		xrClient: xrClient,
		producer: producer,
	}
}

func (s *Service) Run(ctx context.Context) error {
	t := time.NewTicker(cleanupTicker)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
			if err := s.cleanupWallet(ctx); err != nil {
				return fmt.Errorf("failed to cleanup inactive wallets: %w", err)
			}
		}
	}
}

func (s *Service) cleanupWallet(ctx context.Context) error {
	if err := s.wallets.WalletCleaner(ctx); err != nil {
		return fmt.Errorf("failed to cleanup wallets: %w", err)
	}

	return nil
}

func (s *Service) CreateWallet(ctx context.Context, wallet models.Wallet, userID models.UserID) (models.Wallet, error) {
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

func (s *Service) GetWallet(ctx context.Context, walletID models.WalletID,
	userID models.UserID,
) (models.Wallet, error) {
	if walletID == models.WalletID(uuid.Nil) {
		return models.Wallet{}, fmt.Errorf("%w", models.ErrEmptyID)
	}

	if userID == models.UserID(uuid.Nil) {
		return models.Wallet{}, fmt.Errorf("%w", models.ErrUserID)
	}

	var wallet models.Wallet

	walletInfo, err := s.wallets.GetWallet(ctx, walletID, userID, wallet)
	if err != nil {
		return models.Wallet{}, fmt.Errorf("wallet not found: %w", err)
	}

	return walletInfo, nil
}

func (s *Service) UpdateWallet(ctx context.Context, walletID models.WalletID, userID models.UserID,
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

func (s *Service) DeleteWallet(ctx context.Context, walletID models.WalletID, userID models.UserID) error {
	if walletID == models.WalletID(uuid.Nil) {
		return fmt.Errorf("%w", models.ErrEmptyID)
	}

	if err := s.wallets.DeleteWallet(ctx, walletID, userID); err != nil {
		return fmt.Errorf("failed delete wallet: %w", err)
	}

	return nil
}

func (s *Service) GetWallets(ctx context.Context, request models.GetWalletsRequest,
	userID models.UserID,
) ([]models.Wallet, error) {
	wallets, err := s.wallets.GetWallets(ctx, request, userID)
	if err != nil {
		return nil, fmt.Errorf("failed get all wallets: %w", err)
	}

	return wallets, nil
}

func (s *Service) Deposit(ctx context.Context, userID models.UserID, transaction models.Transaction) error {
	if err := transaction.Validate(); err != nil {
		return fmt.Errorf("error validating transaction: %w", err)
	}

	if err := s.wallets.Deposit(ctx, userID, transaction); err != nil {
		return fmt.Errorf("failed deposit: %w", err)
	}

	txJSON, err := json.Marshal(transaction)
	if err != nil {
		return fmt.Errorf("failed to marshal deposit transaction: %w", err)
	}

	if err = s.producer.ProduceTx("", string(txJSON)); err != nil {
		return fmt.Errorf("failed to produce deposit transaction: %w", err)
	}

	return nil
}

func (s *Service) WithdrawMoney(ctx context.Context, userID models.UserID, transaction models.Transaction) error {
	if err := transaction.Validate(); err != nil {
		return fmt.Errorf("error validating transaction: %w", err)
	}

	if err := s.wallets.WithdrawMoney(ctx, userID, transaction); err != nil {
		return fmt.Errorf("failed withdraw money: %w", err)
	}

	txJSON, err := json.Marshal(transaction)
	if err != nil {
		return fmt.Errorf("failed to marshal withdraw transaction: %w", err)
	}

	if err = s.producer.ProduceTx("", string(txJSON)); err != nil {
		return fmt.Errorf("failed to produce withdraw transaction: %w", err)
	}

	return nil
}

func (s *Service) Transfer(ctx context.Context, userID models.UserID, transaction models.Transaction) error {
	if err := transaction.Validate(); err != nil {
		return fmt.Errorf("error validating transaction: %w", err)
	}

	if transaction.SecondWalletID == nil {
		return fmt.Errorf("%w", models.ErrEmptyID)
	}

	secondWallet, err := s.wallets.GetCurrency(ctx, *transaction.SecondWalletID)
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

	txJSON, err := json.Marshal(transaction)
	if err != nil {
		return fmt.Errorf("failed to marshal transfer transaction: %w", err)
	}

	if err = s.producer.ProduceTx("", string(txJSON)); err != nil {
		return fmt.Errorf("failed to produce transfer transaction: %w", err)
	}

	return nil
}

func (s *Service) GetTransactions(ctx context.Context, request models.GetWalletsRequest,
	walletID models.WalletID, userID models.UserID,
) ([]models.Transaction, error) {
	_, err := s.GetWallet(ctx, walletID, userID)
	if err != nil {
		return nil, fmt.Errorf("%w", models.ErrWrongUserID)
	}

	transactions, err := s.wallets.GetTransactions(ctx, request, walletID)
	if err != nil {
		return nil, fmt.Errorf("failed to get all transactions: %w", err)
	}

	return transactions, nil
}
