package handler

import (
	"gochat-backend/internal/usecase/auth"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type RegisterRequest struct {
	Name     string `form:"name" binding:"required,max=255"`
	Email    string `form:"email" binding:"required,email,customEmail,max=255"`
	Password string `form:"password" binding:"required,min=6,max=255"`
}

func Register(c *gin.Context, authUseCase auth.AuthUseCase) {
	var req RegisterRequest

	if err := c.ShouldBindWith(&req, binding.FormMultipart); err != nil {
		log.Println("Error binding request:", err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	file, err := c.FormFile("avatar")

	if err != nil {
		log.Println("Error getting avatar file:", err)
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
		log.Println("Error during registration:", err)
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
