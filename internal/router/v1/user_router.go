package v1

import (
	"gochat-backend/internal/middleware"
	"gochat-backend/internal/usecase/profile"

	profileHandler "gochat-backend/internal/handler/profile"

	"github.com/gin-gonic/gin"
)

func InitUserRouter(
	router gin.IRouter,
	middleware middleware.Middleware,
	profileUseCase profile.ProfileUseCase,
) {
	router.GET("/:id", middleware.Authentication, func(c *gin.Context) {
		profileHandler.GetUserProfile(c, profileUseCase)
	})

	router.GET("", middleware.Authentication, func(c *gin.Context) {
		profileHandler.SearchUsersByName(c, profileUseCase)
	})
}
