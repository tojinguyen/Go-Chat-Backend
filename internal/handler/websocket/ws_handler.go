package handler

import (
	"gochat-backend/internal/socket"
	"log"

	"github.com/gin-gonic/gin"
)

// Tạo kết nối WebSocket từ client
func HandleWebSocketConnection(c *gin.Context, socketManager *socket.SocketManager) {
	// Get userID from Context
	userID := c.GetString("userId")

	log.Printf("Establishing WebSocket connection for user: %s", userID)

	// Chuyển từ gin context sang http context
	// Vì gin.Context.Writer và gin.Context.Request là http.ResponseWriter và *http.Request
	socketManager.ServeWS(c.Writer, c.Request, userID)
}
