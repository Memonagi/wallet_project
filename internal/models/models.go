package models

import "time"

type UserExternal struct {
	UserID           int       `json:"userId"`
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
