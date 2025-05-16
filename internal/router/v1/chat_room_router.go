package v1

import (
	chatHandler "gochat-backend/internal/handler/chat_room"
	"gochat-backend/internal/middleware"
	"gochat-backend/internal/usecase/chat"

	"github.com/gin-gonic/gin"
)

func InitChatRoomRouter(
	router gin.IRouter,
	middleware middleware.Middleware,
	chatUseCase chat.ChatUseCase,
) {
	// Create a new chat room
	router.POST("", middleware.Authentication, func(c *gin.Context) {
		chatHandler.CreateChatRoom(c, chatUseCase)
	})

	// Get all chat rooms for the current user
	router.GET("", middleware.Authentication, func(c *gin.Context) {
		chatHandler.GetChatRooms(c, chatUseCase)
	})

	// Get a specific chat room by ID
	router.GET("/:id", middleware.Authentication, func(c *gin.Context) {
		chatHandler.GetChatRoomByID(c, chatUseCase)
	})

	// Add members to a chat room
	router.POST("/:id/members", middleware.Authentication, func(c *gin.Context) {
		chatHandler.AddChatRoomMembers(c, chatUseCase)
	})

	// Remove a member from a chat room
	router.DELETE("/:id/members/:userID", middleware.Authentication, func(c *gin.Context) {
		chatHandler.RemoveChatRoomMember(c, chatUseCase)
	})

	// Get all messages in a chat room with pagination
	router.GET("/:id/messages", middleware.Authentication, func(c *gin.Context) {
		chatHandler.GetChatRoomMessages(c, chatUseCase)
	})

	// Send a message to a chat room
	router.POST("/:id/messages", middleware.Authentication, func(c *gin.Context) {
		chatHandler.SendMessage(c, chatUseCase)
	})

	// Leave a chat room
	router.POST("/:id/leave", middleware.Authentication, func(c *gin.Context) {
		chatHandler.LeaveChatRoom(c, chatUseCase)
	})
}
