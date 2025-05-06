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
	Email string `json:"email" binding:"required,email,customEmail,max=255"`
	Code  string `json:"code" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email,customEmail,max=255"`
	Password string `json:"password" binding:"required,min=6,max=255"`
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with name, email, password and avatar
// @Tags Auth
// @Accept multipart/form-data
// @Produce json
// @Param name formData string true "User name"
// @Param email formData string true "Email"
// @Param password formData string true "Password"
// @Param avatar formData file true "Avatar image"
// @Success 201 {object} handler.APIResponse{data=auth.RegisterOutput}
// @Failure 400 {object} handler.APIResponse
// @Router /auth/register [post]
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

// VerifyRegistrationCode godoc
// @Summary Verify user registration code
// @Description Verify the registration code sent to user's email
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body VerifyRegistrationRequest true "Email and verification code"
// @Success 200 {object} handler.APIResponse{data=auth.RegisterOutput}
// @Failure 400 {object} handler.APIResponse "Invalid request or verification failed"
// @Router /auth/verify [post]
func VerifyRegistrationCode(c *gin.Context, authUseCase auth.AuthUseCase) {
	var req VerifyRegistrationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("Error binding verification request:", err)
		handler.SendErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	input := auth.VerifyRegistrationInput{
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

// Login godoc
// @Summary Login
// @Description Login with email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} handler.APIResponse{data=auth.LoginOutput}
// @Failure 401 {object} handler.APIResponse
// @Router /auth/login [post]
func Login(c *gin.Context, authUseCase auth.AuthUseCase) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("Error binding login request:", err)
		handler.SendErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	input := auth.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	}

	result, err := authUseCase.Login(c, input)
	if err != nil {
		log.Println("Error during login:", err)
		handler.SendErrorResponse(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	handler.SendSuccessResponse(c, http.StatusOK, "Login successful", result)
}

func VerifyToken(c *gin.Context, authUseCase auth.AuthUseCase) {
	// Extract the token from the request header
	token := c.Request.Header.Get("Authorization")
	log.Println("Verify Token:", token)
	if token == "" {
		handler.SendErrorResponse(c, http.StatusUnauthorized, "Token is required")
		return
	}

	// Verify the token using the authUseCase
	result, err := authUseCase.VerifyToken(c, token)
	if err != nil {
		log.Println("Error verifying token:", err)
		handler.SendErrorResponse(c, http.StatusUnauthorized, "Invalid or expired token")
		return
	}

	handler.SendSuccessResponse(c, http.StatusOK, "Token is valid", result)
}

func RefreshToken(c *gin.Context, authUseCase auth.AuthUseCase) {
	// Extract the refresh token from the request header
	refreshToken := c.Request.Header.Get("Authorization")
	log.Println("Refresh Token:", refreshToken)
	if refreshToken == "" {
		handler.SendErrorResponse(c, http.StatusUnauthorized, "Refresh token is required")
		return
	}

	// Refresh the token using the authUseCase
	result, err := authUseCase.RefreshToken(c, refreshToken)
	if err != nil {
		log.Println("Error refreshing token:", err)
		handler.SendErrorResponse(c, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}

	handler.SendSuccessResponse(c, http.StatusOK, "Token refreshed successfully", result)
}

func ChangePassword(c *gin.Context, authUseCase auth.AuthUseCase) {
}

func ResetPassword(c *gin.Context, authUseCase auth.AuthUseCase) {
}

func CheckTokenResetPassword(c *gin.Context, authUseCase auth.AuthUseCase) {
}

func RequestResetPassword(c *gin.Context, authUseCase auth.AuthUseCase) {
}
