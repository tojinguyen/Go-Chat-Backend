package v1

import (
	"gochat-backend/internal/middleware"
	"gochat-backend/internal/socket"

	"gochat-backend/pkg/jwt"

	wsHandler "gochat-backend/internal/handler/websocket"

	"github.com/gin-gonic/gin"
)

func InitWebSocketRouter(
	router gin.IRouter,
	middleware middleware.Middleware,
	socketManager *socket.SocketManager,
	jwtService jwt.JwtService,
) {
	// Route chính để kết nối WebSocket
	router.GET("/ws", func(c *gin.Context) {
		wsHandler.HandleWebSocketConnection(c, socketManager, jwtService)
	})

	authorized := router.Group("/")
	authorized.Use(middleware.Authentication)
	{
		// API endpoints đi kèm với WebSocket
		authorized.GET("/rooms/:chat_room_id/join", func(c *gin.Context) {
			wsHandler.JoinChatRoom(c, socketManager)
		})
	}
}
