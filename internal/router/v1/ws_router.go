package v1

import (
	"gochat-backend/internal/middleware"
	"gochat-backend/internal/socket"

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
		wsHandler.HandleWebSocketConnection(c, socketManager)
	})
}
