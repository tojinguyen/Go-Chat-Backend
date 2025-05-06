package v1

import (
	authHandler "gochat-backend/internal/handler/auth"
	"gochat-backend/internal/middleware"
	"gochat-backend/internal/usecase/auth"

	"github.com/gin-gonic/gin"
)

func InitAuthRouter(
	router gin.IRouter,
	middleware middleware.Middleware,
	authUseCase auth.AuthUseCase,
) {
	router.POST("/register", func(c *gin.Context) {
		authHandler.Register(c, authUseCase)
	})

	router.POST("/verify-registration-code", func(c *gin.Context) {
		authHandler.VerifyRegistrationCode(c, authUseCase)
	})

	router.POST("/login", func(c *gin.Context) {
		authHandler.Login(c, authUseCase)
	})

	router.GET("/verify-token", middleware.Authentication, func(c *gin.Context) {
		authHandler.VerifyToken(c, authUseCase)
	})

	router.GET("/refresh-token", func(c *gin.Context) {
		authHandler.RefreshToken(c, authUseCase)
	})

	router.PUT("/change-password", middleware.Authentication, func(c *gin.Context) {
		authHandler.ChangePassword(c, authUseCase)
	})

	router.PUT("/reset-password", func(context *gin.Context) {
		authHandler.ResetPassword(context, authUseCase)
	})

	router.GET("/reset-password", func(context *gin.Context) {
		authHandler.CheckTokenResetPassword(context, authUseCase)
	})

	router.POST("/reset-password", func(context *gin.Context) {
		authHandler.RequestResetPassword(context, authUseCase)
	})
}
