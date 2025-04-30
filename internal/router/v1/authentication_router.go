package v1

import (
	"gochat-backend/internal/handler"
	"gochat-backend/internal/middleware"
	"gochat-backend/internal/usecase/auth"

	"github.com/gin-gonic/gin"
)

func InitAuthRouter(
	router gin.IRouter,
	middleware middleware.Middleware,
	authUseCase auth.AuthUseCase,
) {
	router.POST("/login", func(c *gin.Context) {
		handler.Login(c, authUseCase)
	})

	router.GET("/refresh-token", func(c *gin.Context) {
		handler.RefreshToken(c, authUseCase)
	})

	router.PUT("/change-password", middleware.Authentication, func(c *gin.Context) {
		handler.ChangePassword(c, authUseCase)
	})

	router.PUT("/reset-password", func(context *gin.Context) {
		handler.ResetPassword(context, authUseCase)
	})

	router.GET("/reset-password", func(context *gin.Context) {
		handler.CheckTokenResetPassword(context, authUseCase)
	})

	router.POST("/reset-password", func(context *gin.Context) {
		handler.RequestResetPassword(context, authUseCase)
	})
}
