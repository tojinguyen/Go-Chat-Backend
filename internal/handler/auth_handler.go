package handler

import "gochat-backend/internal/repository"

type AuthHandler struct {
	repo *repository.AccountRepo
}

func NewAuthHandler(repo *repository.AccountRepo) *AuthHandler {
	return &AuthHandler{
		repo: repo,
	}
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

