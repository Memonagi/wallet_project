package models

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type UserExternal struct {
	UserID           uuid.UUID `json:"userId"`
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

type UsersInfo struct {
	UserID    uuid.UUID `json:"userId"`
	Status    string    `json:"status"`
	Archived  bool      `json:"archived"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Wallet struct {
	WalletID  uuid.UUID `json:"walletId"`
	UserID    uuid.UUID `json:"userId"`
	Name      string    `json:"name"`
	Currency  string    `json:"currency"`
	Balance   string    `json:"balance"`
	Archived  bool      `json:"archived"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type WalletUpdate struct {
	Name     *string `json:"name"`
	Currency *string `json:"currency"`
	Balance  *string `json:"balance,omitempty"`
}

type GetWalletsRequest struct {
	Sorting    string `json:"sorting,omitempty"`
	Descending bool   `json:"descending,omitempty"`
	Limit      int    `json:"limit,omitempty"`
	Filter     string `json:"filter,omitempty"`
	Offset     int    `json:"offset,omitempty"`
}

type XRRequest struct {
	FromCurrency string `json:"fromCurrency"`
	ToCurrency   string `json:"toCurrency"`
}

type XRResponse struct {
	Rate float64 `json:"rate"`
}

var (
	errEmptyName      = errors.New("wallet name is empty")
	errWrongCurrency  = errors.New("wallet currency is invalid")
	ErrEmptyID        = errors.New("wallet ID is empty")
	ErrWalletNotFound = errors.New("wallet not found")
	ErrUserNotFound   = errors.New("user not found")
	ErrWrongCurrency  = errors.New("currency is invalid")
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
	//nolint:gochecknoglobals
	ExchangeRates = map[string]float64{
		"USD": 1.5, //nolint:mnd
		"EUR": 1.6, //nolint:mnd
		"RUB": 1,
		"JPY": 0.8, //nolint:mnd
		"CNY": 1.2, //nolint:mnd
		"CAD": 1.3, //nolint:mnd
		"AUD": 1.1, //nolint:mnd
	}
)

func (w *Wallet) Validate() error {
	if w.Name == "" {
		return errEmptyName
	}

	_, ok := currencies[strings.ToUpper(w.Currency)]
	if !ok {
		return errWrongCurrency
	}

	return nil
}

func (u *WalletUpdate) Validate() error {
	if *u.Name == "" {
		return errEmptyName
	}

	_, ok := currencies[strings.ToUpper(*u.Currency)]
	if !ok {
		return errWrongCurrency
	}

	return nil
}
