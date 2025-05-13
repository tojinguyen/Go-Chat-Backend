package handler

import (
	"gochat-backend/internal/handler"
	"gochat-backend/internal/socket"
	"gochat-backend/pkg/jwt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Tạo kết nối WebSocket từ client
func HandleWebSocketConnection(c *gin.Context, socketManager *socket.SocketManager, jwtService jwt.JwtService) {
	// Lấy token từ query parameter hoặc header
	token := c.Query("token")
	if token == "" {
		token = c.GetHeader("Authorization")
		// Xóa tiền tố "Bearer " nếu có
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
	}

	// Xác thực token
	claims, err := jwtService.ValidateAccessToken(token)
	if err != nil {
		log.Printf("Unauthorized WebSocket connection: %v", err)
		c.JSON(http.StatusUnauthorized, handler.APIResponse{
			Success: false,
			Message: "Unauthorized",
			Data:    nil,
		})
		return
	}

	// Lấy userID từ claims
	userID := claims.UserId

	// Nếu client_id được cung cấp, kiểm tra xem nó có khớp với userID trong token không
	clientID := c.Query("client_id")
	if clientID != "" && clientID != userID {
		log.Printf("Client ID mismatch: %s vs %s", clientID, userID)
		c.JSON(http.StatusForbidden, handler.APIResponse{
			Success: false,
			Message: "Client ID mismatch",
			Data:    nil,
		})
		return
	}

	// Sử dụng userID từ token nếu không có client_id
	if clientID == "" {
		clientID = userID
	}

	log.Printf("Establishing WebSocket connection for user: %s", clientID)

	// Chuyển từ gin context sang http context
	// Vì gin.Context.Writer và gin.Context.Request là http.ResponseWriter và *http.Request
	socketManager.ServeWS(c.Writer, c.Request, clientID)
}

// JoinChatRoom godoc
// @Summary Join a chat room
// @Description Join a specific chat room via WebSocket connection
// @Tags WebSocket
// @Accept json
// @Produce json
// @Param chat_room_id path string true "Chat Room ID"
// @Param Authorization header string true "Bearer token"
// @Success 200 {string} string "WebSocket connection established"
// @Failure 401 {object} handler.APIResponse "Unauthorized"
// @Router /ws/rooms/{chat_room_id}/join [get]
func JoinChatRoom(c *gin.Context, socketManager *socket.SocketManager) {
	chatRoomID := c.Param("chat_room_id")
	clientID := c.GetString("user_id") // Đã được xác thực qua middleware

	if chatRoomID == "" {
		c.JSON(http.StatusBadRequest, handler.APIResponse{
			Success: false,
			Message: "Chat room ID is required",
		})
		return
	}

	// Trong trường hợp thực, việc join room sẽ được xử lý thông qua kết nối WebSocket,
	// nhưng chúng ta có thể cung cấp một endpoint API để bắt đầu quá trình
	c.JSON(http.StatusOK, handler.APIResponse{
		Success: true,
		Message: "Join request accepted. Connect via WebSocket to complete.",
		Data: map[string]string{
			"chat_room_id": chatRoomID,
			"client_id":    clientID,
		},
	})
}
