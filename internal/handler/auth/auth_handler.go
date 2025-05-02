package handler

import (
	"gochat-backend/internal/usecase/auth"
	"log"
	"net/http"

	"gochat-backend/internal/handler"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type RegisterRequest struct {
	Name     string `form:"name" binding:"required,max=255"`
	Email    string `form:"email" binding:"required,email,customEmail,max=255"`
	Password string `form:"password" binding:"required,min=6,max=255"`
}

type VerifyRegistrationRequest struct {
	ID    string `json:"id" binding:"required"`
	Email string `json:"email" binding:"required,email,customEmail,max=255"`
	Code  string `json:"code" binding:"required"`
}

func Register(c *gin.Context, authUseCase auth.AuthUseCase) {
	var req RegisterRequest

	if err := c.ShouldBindWith(&req, binding.FormMultipart); err != nil {
		log.Println("Error binding request:", err)
		handler.SendErrorResponse(c, 400, err.Error())
		return
	}

	file, err := c.FormFile("avatar")

	if err != nil {
		log.Println("Error getting avatar file:", err)
		handler.SendErrorResponse(c, 400, "Failed to get avatar file")
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
		handler.SendErrorResponse(c, 400, err.Error())
		return
	}

	handler.SendSuccessResponse(c, http.StatusCreated, "Registration successful! Please check your email for verification.", result)
}

func VerifyRegistrationCode(c *gin.Context, authUseCase auth.AuthUseCase) {
	var req VerifyRegistrationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("Error binding verification request:", err)
		handler.SendErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	input := auth.VerifyRegistrationInput{
		ID:    req.ID,
		Email: req.Email,
		Code:  req.Code,
	}

	result, err := authUseCase.VerifyRegistration(c, input)
	if err != nil {
		log.Println("Error during verification:", err)
		handler.SendErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	handler.SendSuccessResponse(c, http.StatusOK, "Email verification successful! You can now log in.", result)
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
