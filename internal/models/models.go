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
}

var (
	ErrEmptyName      = errors.New("wallet name is empty")
	ErrWrongCurrency  = errors.New("wallet currency is invalid")
	ErrEmptyID        = errors.New("wallet ID is empty")
	ErrWalletNotFound = errors.New("wallet not found")
	ErrUserNotFound   = errors.New("user not found")
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
