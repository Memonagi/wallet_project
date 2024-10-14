package models

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type (
	WalletID uuid.UUID
	UserID   uuid.UUID
	TxID     uuid.UUID
)

type UserExternal struct {
	UserID           UserID    `json:"userId"`
	UserName         string    `json:"userName"`
	UserSurname      string    `json:"userSurname"`
	UserAge          int       `json:"userAge"`
	UserGender       string    `json:"userGender"`
	UserEmail        string    `json:"userEmail"`
	Country          string    `json:"country"`
	EngagementSource string    `json:"engagementSource"`
	Status           string    `json:"status"`
	Archived         bool      `json:"archived"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type User struct {
	UserID    UserID    `json:"userId"`
	Status    string    `json:"status"`
	Archived  bool      `json:"archived"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Wallet struct {
	WalletID  WalletID  `json:"walletId"`
	UserID    UserID    `json:"userId"`
	Name      string    `json:"name"`
	Currency  string    `json:"currency"`
	Balance   float64   `json:"balance"`
	Archived  bool      `json:"archived"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type WalletUpdate struct {
	Name     *string `json:"name"`
	Currency *string `json:"currency"`
}

type GetWalletsRequest struct {
	Sorting    string `json:"sorting,omitempty"`
	Descending bool   `json:"descending,omitempty"`
	Limit      int    `json:"limit,omitempty"`
	Filter     string `json:"filter,omitempty"`
	Offset     int    `json:"offset,omitempty"`
}

type UserInfo struct {
	UserID UserID `json:"userId"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

type XRRequest struct {
	FromCurrency string `json:"fromCurrency"`
	ToCurrency   string `json:"toCurrency"`
}

type XRResponse struct {
	Rate float64 `json:"rate"`
}

type Transaction struct {
	ID             TxID      `json:"id"`
	Name           string    `json:"name"`
	FirstWalletID  WalletID  `json:"firstWallet"`
	SecondWalletID *WalletID `json:"secondWallet"`
	Money          float64   `json:"money"`
	Currency       string    `json:"currency"`
	CreatedAt      time.Time `json:"createdAt"`
}

var (
	ErrEmptyName            = errors.New("wallet name is empty")
	ErrEmptyID              = errors.New("wallet ID is empty")
	ErrWalletNotFound       = errors.New("wallet not found")
	ErrUserNotFound         = errors.New("user not found")
	ErrWrongCurrency        = errors.New("currency is invalid")
	ErrInvalidToken         = errors.New("invalid token")
	ErrInvalidSigningMethod = errors.New("invalid signing method")
	ErrInsufficientFunds    = errors.New("insufficient funds")
	ErrWrongMoney           = errors.New("zero or negative amount of money")
	ErrUserID               = errors.New("user ID is empty")
	ErrWrongUserID          = errors.New("user is not the owner of the wallet")
	//nolint:gochecknoglobals
	currencies = map[string]struct{}{
		"USD": {},
		"EUR": {},
		"RUB": {},
		"JPY": {},
		"CNY": {},
		"CAD": {},
		"AUD": {},
	}
)

func (w *Wallet) Validate() error {
	if w.Name == "" {
		return ErrEmptyName
	}

	_, ok := currencies[strings.ToUpper(w.Currency)]
	if !ok {
		return ErrWrongCurrency
	}

	return nil
}

func (u *WalletUpdate) Validate() error {
	if *u.Name == "" {
		return ErrEmptyName
	}

	_, ok := currencies[strings.ToUpper(*u.Currency)]
	if !ok {
		return ErrWrongCurrency
	}

	return nil
}

func (t *Transaction) Validate() error {
	switch {
	case t.Money == 0:
		return ErrWrongMoney
	case t.Money < 0:
		return ErrWrongMoney
	case t.FirstWalletID == WalletID(uuid.Nil):
		return ErrWalletNotFound
	}

	_, ok := currencies[strings.ToUpper(t.Currency)]
	if !ok {
		return ErrWrongCurrency
	}

	return nil
}
