package v1

import (
	"gochat-backend/internal/middleware"
	"gochat-backend/internal/usecase/friend"

	"github.com/gin-gonic/gin"
)

func InitFriendRouter(
	router gin.IRouter,
	middleware middleware.Middleware,
	friendUseCase friend.FriendUseCase,
) {
	// Get all friends
	router.GET("/", middleware.Authentication, func(c *gin.Context) {
	})

	// Request to add a friend
	router.GET("/requests", middleware.Authentication, func(c *gin.Context) {
	})

	// Get all friend requests
	router.POST("/requests", middleware.Authentication, func(c *gin.Context) {
	})

	// Accept a friend request
	router.POST("/requests/:requestID/accept", middleware.Authentication, func(c *gin.Context) {
	})

	// Reject a friend request
	router.POST("/requests/:requestID/reject", middleware.Authentication, func(c *gin.Context) {
	})

	// Delete a friend
	router.DELETE("/:friendID", middleware.Authentication, func(c *gin.Context) {
	})
}
