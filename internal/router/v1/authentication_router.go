package v1

import (
	"gochat-backend/internal/handler"
	"gochat-backend/internal/middleware"
	"gochat-backend/internal/usecase/auth"

	"github.com/gin-gonic/gin"
)

func InitAuthRouter(
	r gin.IRouter,
	middleware middleware.Middleware,
	authUseCase auth.AuthUseCase,
) {
	r.POST("/login", func(c *gin.Context) {
		handler.Login(c, authUseCase)
	})

	r.GET("/refresh-token", func(c *gin.Context) {
		handler.RefreshToken(c, authUseCase)
	})

	r.PUT("/change-password", middleware.Authentication, func(c *gin.Context) {
		handler.ChangePassword(c, authUseCase)
	})

	r.PUT("/reset-password", func(context *gin.Context) {
		handler.ResetPassword(context, authUseCase)
	})

	r.GET("/reset-password", func(context *gin.Context) {
		handler.CheckTokenResetPassword(context, authUseCase)
	})

	r.POST("/reset-password", func(context *gin.Context) {
		handler.RequestResetPassword(context, authUseCase)
	})
}
