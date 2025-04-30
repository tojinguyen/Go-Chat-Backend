package handler

import (
	"gochat-backend/internal/repository"
	"gochat-backend/internal/usecase/auth"

	"github.com/gin-gonic/gin"
)

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

func Login(c *gin.Context, authUseCase auth.AuthUseCase) {
}

func RefreshToken(c *gin.Context, authUseCase auth.AuthUseCase) {
}

func ChangePassword(c *gin.Context, authUseCase auth.AuthUseCase) {
}

func ResetPassword(c *gin.Context, authUseCase auth.AuthUseCase) {
}

func CheckTokenResetPassword(c *gin.Context, authUseCase auth.AuthUseCase) {
}

func RequestResetPassword(c *gin.Context, authUseCase auth.AuthUseCase) {
}
