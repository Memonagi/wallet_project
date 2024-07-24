package models

import (
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

type WalletInfo struct {
	WalletID  uuid.UUID `json:"walletId"`
	UserID    uuid.UUID `json:"userId"`
	Name      string    `json:"name"`
	Currency  string    `json:"currency"`
	Balance   string    `json:"balance"`
	Archived  bool      `json:"archived"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
