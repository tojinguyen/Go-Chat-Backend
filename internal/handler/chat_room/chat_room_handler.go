package chat_room

import (
	"fmt"
	"gochat-backend/internal/handler"
	"gochat-backend/internal/usecase/chat"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateChatRoom handles the creation of a new chat room
func CreateChatRoom(c *gin.Context, chatUseCase chat.ChatUseCase) {
	// Get user ID from context
	userID := c.GetString("user_id")
	if userID == "" {
		handler.SendErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse request body
	var input chat.ChatRoomCreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		handler.SendErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Invalid request format: %v", err))
		return
	}

	// Create chat room
	chatRoom, err := chatUseCase.CreateChatRoom(c.Request.Context(), userID, input)
	if err != nil {
		handler.SendErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create chat room: %v", err))
		return
	}

	handler.SendSuccessResponse(c, http.StatusCreated, "Chat room created successfully", chatRoom)
}

// GetChatRooms retrieves all chat rooms for the current user
func GetChatRooms(c *gin.Context, chatUseCase chat.ChatUseCase) {
	// Get user ID from context
	userID := c.GetString("user_id")
	if userID == "" {
		handler.SendErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get all chat rooms
	chatRooms, err := chatUseCase.GetChatRooms(c.Request.Context(), userID)
	if err != nil {
		handler.SendErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to get chat rooms: %v", err))
		return
	}

	handler.SendSuccessResponse(c, http.StatusOK, "Chat rooms retrieved successfully", chatRooms)
}

// GetChatRoomByID retrieves a specific chat room by ID
func GetChatRoomByID(c *gin.Context, chatUseCase chat.ChatUseCase) {
	// Get user ID from context
	userID := c.GetString("user_id")
	if userID == "" {
		handler.SendErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get chat room ID from URL params
	chatRoomID := c.Param("id")
	if chatRoomID == "" {
		handler.SendErrorResponse(c, http.StatusBadRequest, "Chat room ID is required")
		return
	}

	// Get chat room
	chatRoom, err := chatUseCase.GetChatRoomByID(c.Request.Context(), userID, chatRoomID)
	if err != nil {
		handler.SendErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to get chat room: %v", err))
		return
	}

	handler.SendSuccessResponse(c, http.StatusOK, "Chat room retrieved successfully", chatRoom)
}

// AddChatRoomMembers adds members to a chat room
func AddChatRoomMembers(c *gin.Context, chatUseCase chat.ChatUseCase) {
	// Get user ID from context
	userID := c.GetString("user_id")
	if userID == "" {
		handler.SendErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get chat room ID from URL params
	chatRoomID := c.Param("id")
	if chatRoomID == "" {
		handler.SendErrorResponse(c, http.StatusBadRequest, "Chat room ID is required")
		return
	}

	// Parse request body
	var input struct {
		Members []string `json:"members"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		handler.SendErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Invalid request format: %v", err))
		return
	}

	if len(input.Members) == 0 {
		handler.SendErrorResponse(c, http.StatusBadRequest, "No members specified")
		return
	}

	// Add members
	err := chatUseCase.AddChatRoomMembers(c.Request.Context(), userID, chatRoomID, input.Members)
	if err != nil {
		handler.SendErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to add members: %v", err))
		return
	}

	handler.SendSuccessResponse(c, http.StatusOK, "Members added successfully", nil)
}

// RemoveChatRoomMember removes a member from a chat room
func RemoveChatRoomMember(c *gin.Context, chatUseCase chat.ChatUseCase) {
	// Get user ID from context
	userID := c.GetString("user_id")
	if userID == "" {
		handler.SendErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get chat room ID and member ID from URL params
	chatRoomID := c.Param("id")
	if chatRoomID == "" {
		handler.SendErrorResponse(c, http.StatusBadRequest, "Chat room ID is required")
		return
	}

	memberID := c.Param("userID")
	if memberID == "" {
		handler.SendErrorResponse(c, http.StatusBadRequest, "Member ID is required")
		return
	}

	// Remove member
	err := chatUseCase.RemoveChatRoomMember(c.Request.Context(), userID, chatRoomID, memberID)
	if err != nil {
		handler.SendErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to remove member: %v", err))
		return
	}

	handler.SendSuccessResponse(c, http.StatusOK, "Member removed successfully", nil)
}

// GetChatRoomMessages retrieves messages from a chat room with pagination
func GetChatRoomMessages(c *gin.Context, chatUseCase chat.ChatUseCase) {
	// Get user ID from context
	userID := c.GetString("user_id")
	if userID == "" {
		handler.SendErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get chat room ID from URL params
	chatRoomID := c.Param("id")
	if chatRoomID == "" {
		handler.SendErrorResponse(c, http.StatusBadRequest, "Chat room ID is required")
		return
	}

	// Parse pagination params
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 20
	}

	// Get messages
	messages, err := chatUseCase.GetChatRoomMessages(c.Request.Context(), userID, chatRoomID, page, limit)
	if err != nil {
		handler.SendErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to get messages: %v", err))
		return
	}

	handler.SendSuccessResponse(c, http.StatusOK, "Messages retrieved successfully", messages)
}

// SendMessage sends a message to a chat room
func SendMessage(c *gin.Context, chatUseCase chat.ChatUseCase) {
	// Get user ID from context
	userID := c.GetString("user_id")
	if userID == "" {
		handler.SendErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get chat room ID from URL params
	chatRoomID := c.Param("id")
	if chatRoomID == "" {
		handler.SendErrorResponse(c, http.StatusBadRequest, "Chat room ID is required")
		return
	}

	// Parse request body
	var input chat.MessageInput
	if err := c.ShouldBindJSON(&input); err != nil {
		handler.SendErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Invalid request format: %v", err))
		return
	}

	// Send message
	message, err := chatUseCase.SendMessage(c.Request.Context(), userID, chatRoomID, input)
	if err != nil {
		handler.SendErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to send message: %v", err))
		return
	}

	handler.SendSuccessResponse(c, http.StatusCreated, "Message sent successfully", message)
}

// LeaveChatRoom allows a user to leave a chat room
func LeaveChatRoom(c *gin.Context, chatUseCase chat.ChatUseCase) {
	// Get user ID from context
	userID := c.GetString("user_id")
	if userID == "" {
		handler.SendErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get chat room ID from URL params
	chatRoomID := c.Param("id")
	if chatRoomID == "" {
		handler.SendErrorResponse(c, http.StatusBadRequest, "Chat room ID is required")
		return
	}

	// Leave chat room
	err := chatUseCase.LeaveChatRoom(c.Request.Context(), userID, chatRoomID)
	if err != nil {
		handler.SendErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to leave chat room: %v", err))
		return
	}

	handler.SendSuccessResponse(c, http.StatusOK, "Left chat room successfully", nil)
}
