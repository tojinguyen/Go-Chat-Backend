package handler

import (
	"gochat-backend/internal/usecase/auth"

	"github.com/gin-gonic/gin"
)

type RegisterRequest struct {
	Name     string `json:"name" binding:"required,max=255"`
	Email    string `json:"email" binding:"required,email,customEmail,max=255"`
	Password string `json:"password" binding:"required,min=6,max=255,customPassword"`
}

func Register(c *gin.Context, authUseCase auth.AuthUseCase) {
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(400, gin.H{"error": "Failed to parse form data"})
		return
	}

	var req RegisterRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	file, err := c.FormFile("avatar")

	if err != nil {
		c.JSON(400, gin.H{"error": "Failed to get avatar file"})
		return
	}

	input := auth.RegisterInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Avatar:   file,
	}

	result, err := authUseCase.Register(c, input)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{
		"message": "Registration successful! Please check your email for verification.",
		"data":    result,
	})
}

func VerifyRegistrationCode(c *gin.Context, authUseCase auth.AuthUseCase) {
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
