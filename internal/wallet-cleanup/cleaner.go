package walletcleanup

import (
	"context"
	"fmt"
	"time"
)

type cleaner interface {
	WalletCleaner(ctx context.Context) error
}
type Cleanup struct {
	cleaner cleaner
}

const cleanupTicker = 24 * time.Hour

func New(cleaner cleaner) *Cleanup {
	return &Cleanup{
		cleaner: cleaner,
	}
}

func (c *Cleanup) Run(ctx context.Context) error {
	t := time.NewTicker(cleanupTicker)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
			if err := c.cleanupWallet(ctx); err != nil {
				return fmt.Errorf("failed to cleanup inactive wallets: %w", err)
			}
		}
	}
}

func (c *Cleanup) cleanupWallet(ctx context.Context) error {
	if err := c.cleaner.WalletCleaner(ctx); err != nil {
		return fmt.Errorf("failed to cleanup wallets: %w", err)
	}

	return nil
}
