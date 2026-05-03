package models

type User struct {
	ID           string `json:"-"`
	Login        string `json:"login"`
	PasswordHash string `json:"-"`
}
