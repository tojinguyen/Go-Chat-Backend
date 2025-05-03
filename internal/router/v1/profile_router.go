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
	router.GET("/profile/:id", middleware.Authentication, func(c *gin.Context) {
		profileHandler.GetUserProfile(c, profileUseCase)
	})

	router.GET("/profiles/search", middleware.Authentication, func(c *gin.Context) {
		profileHandler.SearchUsersByName(c, profileUseCase)
	})
}
