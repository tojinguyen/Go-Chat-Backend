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
// @Summary Create a new chat room
// @Description Creates a new chat room with the authenticated user as owner
// @Tags Chat Room
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body chat.ChatRoomCreateInput true "Chat room creation data"
// @Success 201 {object} handler.APIResponse{data=chat.ChatRoomOutput} "Chat room created successfully"
// @Failure 400 {object} handler.APIResponse "Invalid request format"
// @Failure 401 {object} handler.APIResponse "Unauthorized"
// @Failure 500 {object} handler.APIResponse "Internal server error"
// @Router /chat-rooms [post]
func CreateChatRoom(c *gin.Context, chatUseCase chat.ChatUseCase) {
	// Get user ID from context
	userID := c.GetString("userId")
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
// @Summary Get all user's chat rooms
// @Description Retrieves all chat rooms the authenticated user belongs to
// @Tags Chat Room
// @Produce json
// @Security BearerAuth
// @Success 200 {object} handler.APIResponse{data=[]chat.ChatRoomOutput} "Chat rooms retrieved successfully"
// @Failure 401 {object} handler.APIResponse "Unauthorized"
// @Failure 500 {object} handler.APIResponse "Internal server error"
// @Router /chat-rooms [get]
func GetChatRooms(c *gin.Context, chatUseCase chat.ChatUseCase) {
	// Get user ID from context
	userID := c.GetString("userId")
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
// @Summary Get chat room by ID
// @Description Retrieves a specific chat room by its ID if user is a member
// @Tags Chat Room
// @Produce json
// @Security BearerAuth
// @Param id path string true "Chat Room ID"
// @Success 200 {object} handler.APIResponse{data=chat.ChatRoomOutput} "Chat room retrieved successfully"
// @Failure 400 {object} handler.APIResponse "Chat room ID is required"
// @Failure 401 {object} handler.APIResponse "Unauthorized"
// @Failure 404 {object} handler.APIResponse "Chat room not found"
// @Failure 500 {object} handler.APIResponse "Internal server error"
// @Router /chat-rooms/{id} [get]
func GetChatRoomByID(c *gin.Context, chatUseCase chat.ChatUseCase) {
	// Get user ID from context
	userID := c.GetString("userId")
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
// @Summary Add members to chat room
// @Description Adds new members to an existing chat room
// @Tags Chat Room
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Chat Room ID"
// @Param request body chat.ChatRoomMembersInput true "Members to add"
// @Success 200 {object} handler.APIResponse "Members added successfully"
// @Failure 400 {object} handler.APIResponse "Invalid request format or no members specified"
// @Failure 401 {object} handler.APIResponse "Unauthorized"
// @Failure 403 {object} handler.APIResponse "User not authorized to add members"
// @Failure 404 {object} handler.APIResponse "Chat room not found"
// @Failure 500 {object} handler.APIResponse "Internal server error"
// @Router /chat-rooms/{id}/members [post]
func AddChatRoomMembers(c *gin.Context, chatUseCase chat.ChatUseCase) {
	// Get user ID from context
	userID := c.GetString("userId")
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
	var input chat.ChatRoomMembersInput
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
// @Summary Remove member from chat room
// @Description Removes a specific member from a chat room
// @Tags Chat Room
// @Produce json
// @Security BearerAuth
// @Param id path string true "Chat Room ID"
// @Param userID path string true "User ID to remove"
// @Success 200 {object} handler.APIResponse "Member removed successfully"
// @Failure 400 {object} handler.APIResponse "Chat room ID or Member ID is required"
// @Failure 401 {object} handler.APIResponse "Unauthorized"
// @Failure 403 {object} handler.APIResponse "User not authorized to remove members"
// @Failure 404 {object} handler.APIResponse "Chat room or member not found"
// @Failure 500 {object} handler.APIResponse "Internal server error"
// @Router /chat-rooms/{id}/members/{userID} [delete]
func RemoveChatRoomMember(c *gin.Context, chatUseCase chat.ChatUseCase) {
	// Get user ID from context
	userID := c.GetString("userId")
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
// @Summary Get chat room messages
// @Description Retrieves messages from a chat room with pagination
// @Tags Chat Room
// @Produce json
// @Security BearerAuth
// @Param id path string true "Chat Room ID"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 20)"
// @Success 200 {object} handler.APIResponse{data=[]chat.MessageOutput} "Messages retrieved successfully"
// @Failure 400 {object} handler.APIResponse "Chat room ID is required"
// @Failure 401 {object} handler.APIResponse "Unauthorized"
// @Failure 403 {object} handler.APIResponse "User not a member of chat room"
// @Failure 404 {object} handler.APIResponse "Chat room not found"
// @Failure 500 {object} handler.APIResponse "Internal server error"
// @Router /chat-rooms/{id}/messages [get]
func GetChatRoomMessages(c *gin.Context, chatUseCase chat.ChatUseCase) {
	// Get user ID from context
	userID := c.GetString("userId")
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
// @Summary Send message to chat room
// @Description Sends a new message to a chat room
// @Tags Chat Room
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Chat Room ID"
// @Param request body chat.MessageInput true "Message data"
// @Success 201 {object} handler.APIResponse{data=chat.MessageOutput} "Message sent successfully"
// @Failure 400 {object} handler.APIResponse "Invalid request format or chat room ID is required"
// @Failure 401 {object} handler.APIResponse "Unauthorized"
// @Failure 403 {object} handler.APIResponse "User not a member of chat room"
// @Failure 404 {object} handler.APIResponse "Chat room not found"
// @Failure 500 {object} handler.APIResponse "Internal server error"
// @Router /chat-rooms/{id}/messages [post]
func SendMessage(c *gin.Context, chatUseCase chat.ChatUseCase) {
	// Get user ID from context
	userID := c.GetString("userId")
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
// @Summary Leave chat room
// @Description Allows the authenticated user to leave a chat room
// @Tags Chat Room
// @Produce json
// @Security BearerAuth
// @Param id path string true "Chat Room ID"
// @Success 200 {object} handler.APIResponse "Left chat room successfully"
// @Failure 400 {object} handler.APIResponse "Chat room ID is required"
// @Failure 401 {object} handler.APIResponse "Unauthorized"
// @Failure 404 {object} handler.APIResponse "Chat room not found"
// @Failure 500 {object} handler.APIResponse "Internal server error"
// @Router /chat-rooms/{id}/leave [post]
func LeaveChatRoom(c *gin.Context, chatUseCase chat.ChatUseCase) {
	// Get user ID from context
	userID := c.GetString("userId")
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

// FindOrCreatePrivateChatRoom tìm chat room giữa người dùng hiện tại và một người dùng khác
// Nếu không tồn tại, tạo mới chat room riêng tư
// @Summary Find or create private chat room
// @Description Finds existing private chat room between current user and specified user, or creates a new one
// @Tags Chat Room
// @Produce json
// @Security BearerAuth
// @Param userID path string true "Other User ID"
// @Success 200 {object} handler.APIResponse{data=chat.ChatRoomOutput} "Chat room found or created successfully"
// @Failure 400 {object} handler.APIResponse "User ID is required"
// @Failure 401 {object} handler.APIResponse "Unauthorized"
// @Failure 404 {object} handler.APIResponse "User not found"
// @Failure 500 {object} handler.APIResponse "Internal server error"
// @Router /chat-rooms/private/{userID} [get]
func FindOrCreatePrivateChatRoom(c *gin.Context, chatUseCase chat.ChatUseCase) {
	// Lấy user ID từ context
	currentUserID := c.GetString("userId")
	if currentUserID == "" {
		handler.SendErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Lấy ID người dùng khác từ URL
	otherUserID := c.Param("userID")
	if otherUserID == "" {
		handler.SendErrorResponse(c, http.StatusBadRequest, "User ID is required")
		return
	}

	// Không thể tạo chat với chính mình
	if currentUserID == otherUserID {
		handler.SendErrorResponse(c, http.StatusBadRequest, "Cannot create chat with yourself")
		return
	}

	// Tìm hoặc tạo chat room riêng tư
	chatRoom, err := chatUseCase.FindOrCreatePrivateChatRoom(c.Request.Context(), currentUserID, otherUserID)
	if err != nil {
		handler.SendErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to find or create chat room: %v", err))
		return
	}

	handler.SendSuccessResponse(c, http.StatusOK, "Chat room found or created successfully", chatRoom)
}
