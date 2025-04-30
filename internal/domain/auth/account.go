package domain

import "time"

type Account struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
