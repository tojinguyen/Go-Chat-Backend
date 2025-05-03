package v1

import (
	"gochat-backend/internal/middleware"
	"gochat-backend/internal/usecase/profile"

	profileHandler "gochat-backend/internal/handler/profile"

	"github.com/gin-gonic/gin"
)

func InitProfileRouter(
	router gin.IRouter,
	middleware middleware.Middleware,
	profileUseCase profile.ProfileUseCase,
) {
	router.GET("/users/:id", middleware.Authentication, func(c *gin.Context) {
		profileHandler.GetUserProfile(c, profileUseCase)
	})

	router.GET("/users", middleware.Authentication, func(c *gin.Context) {
		profileHandler.SearchUsersByName(c, profileUseCase)
	})
}
