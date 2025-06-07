package v1

import (
	"gochat-backend/internal/middleware"
	"gochat-backend/internal/socket"
	"log"

	wsHandler "gochat-backend/internal/handler/websocket"

	"github.com/gin-gonic/gin"
)

func InitWebSocketRouter(
	router gin.IRouter,
	middleware middleware.Middleware,
	socketManager *socket.SocketManager,
) {
	// Route chính để kết nối WebSocket
	router.GET("", middleware.Authentication, func(c *gin.Context) {
		log.Printf("WebSocket connection request from user: %s", c.GetString("userId"))
		wsHandler.HandleWebSocketConnection(c, socketManager)
	})
}
