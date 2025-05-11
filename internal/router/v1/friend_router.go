package v1

import (
	"gochat-backend/internal/middleware"
	"gochat-backend/internal/usecase/friend"

	friendHandler "gochat-backend/internal/handler/friend"

	"github.com/gin-gonic/gin"
)

func InitFriendRouter(
	router gin.IRouter,
	middleware middleware.Middleware,
	friendUseCase friend.FriendUseCase,
) {
	// Get all friends with optional pagination (page and limit)
	router.GET("/", middleware.Authentication, func(c *gin.Context) {
		friendHandler.GetFriends(c, friendUseCase)
	})

	// Request to add a friend
	router.POST("/requests", middleware.Authentication, func(c *gin.Context) {
		// Would need to modify handler to use c.Param("friendId")
		friendHandler.AddFriend(c, friendUseCase)
	})

	// Get all friend requests
	router.GET("/requests", middleware.Authentication, func(c *gin.Context) {
		friendHandler.GetFriendRequestList(c, friendUseCase)
	})

	// Accept a friend request
	router.POST("/requests/:requestID/accept", middleware.Authentication, func(c *gin.Context) {
		friendHandler.AcceptFriendRequest(c, friendUseCase)
	})

	// Reject a friend request
	router.POST("/requests/:requestID/reject", middleware.Authentication, func(c *gin.Context) {
		friendHandler.RejectFriendRequest(c, friendUseCase)
	})

	// Delete a friend
	router.DELETE("/:friendID", middleware.Authentication, func(c *gin.Context) {
		friendHandler.DeleteFriend(c, friendUseCase)
	})
}
