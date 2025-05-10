package handler

import (
	"gochat-backend/internal/handler"
	"gochat-backend/internal/usecase/friend"

	"github.com/gin-gonic/gin"
)

type AddFriendRequest struct {
	FriendID string `json:"friendId" binding:"required"`
}

type AcceptFriendRequestRequest struct {
	RequestID string `json:"requestId" binding:"required"`
}

// GetFriends godoc
// @Summary Get user's friends list
// @Description Retrieves a list of all friends for the authenticated user
// @Tags friends
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} handler.APIResponse{data=[]friend.FriendOutput} "List of friends"
// @Failure 401 {object} handler.APIResponse "Unauthorized"
// @Failure 500 {object} handler.APIResponse "Internal server error"
// @Router /api/v1/friends [get]
func GetFriends(c *gin.Context, friendUseCase friend.FriendUseCase) {
	userId, exists := c.Get("userId")
	if !exists {
		handler.SendErrorResponse(c, 401, "Unauthorized: User ID not found")
		return
	}

	userIdStr, ok := userId.(string)
	if !ok {
		handler.SendErrorResponse(c, 500, "Failed to process user identity")
		return
	}

	friends, err := friendUseCase.GetFriendList(c, userIdStr)

	if err != nil {
		handler.SendErrorResponse(c, 500, "Failed to get friends list")
		return
	}

	handler.SendSuccessResponse(c, 200, "Friends list retrieved successfully", friends)
}

// AddFriend godoc
// @Summary Send friend request
// @Description Send a friend request to another user
// @Tags friends
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body handler.AddFriendRequest true "Friend request data"
// @Success 200 {object} handler.APIResponse "Friend request sent successfully"
// @Failure 400 {object} handler.APIResponse "Invalid request"
// @Failure 401 {object} handler.APIResponse "Unauthorized"
// @Failure 404 {object} handler.APIResponse "User not found"
// @Failure 409 {object} handler.APIResponse "Friend request already sent"
// @Failure 500 {object} handler.APIResponse "Internal server error"
// @Router /api/v1/friends/requests [post]
func AddFriend(c *gin.Context, friendUseCase friend.FriendUseCase) {
	var req AddFriendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handler.SendErrorResponse(c, 400, "Invalid request: "+err.Error())
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		handler.SendErrorResponse(c, 401, "Unauthorized: User ID not found")
		return
	}

	userIdStr, ok := userId.(string)
	if !ok {
		handler.SendErrorResponse(c, 500, "Failed to process user identity")
		return
	}

	err := friendUseCase.AddFriend(c, userIdStr, req.FriendID)

	if err != nil {
		handler.SendErrorResponse(c, 500, "Failed to send friend request")
		return
	}

	handler.SendSuccessResponse(c, 200, "Friend request sent successfully", nil)
}

// GetFriendRequestList godoc
// @Summary Get list of friend requests
// @Description Retrieves a list of pending friend requests for the authenticated user
// @Tags friends
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} handler.APIResponse{data=[]friend.FriendRequestOutput} "List of friend requests"
// @Failure 401 {object} handler.APIResponse "Unauthorized"
// @Failure 500 {object} handler.APIResponse "Internal server error"
// @Router /api/v1/friends/requests [get]
func GetFriendRequestList(c *gin.Context, friendUseCase friend.FriendUseCase) {
	userId, exists := c.Get("userId")
	if !exists {
		handler.SendErrorResponse(c, 401, "Unauthorized: User ID not found")
		return
	}

	userIdStr, ok := userId.(string)
	if !ok {
		handler.SendErrorResponse(c, 500, "Failed to process user identity")
		return
	}

	friendRequests, err := friendUseCase.GetFriendRequestList(c, userIdStr)

	if err != nil {
		handler.SendErrorResponse(c, 500, "Failed to get friend request list")
		return
	}

	handler.SendSuccessResponse(c, 200, "Friend request list retrieved successfully", friendRequests)
}

// AcceptFriendRequest godoc
// @Summary Accept a friend request
// @Description Accept a pending friend request from another user
// @Tags friends
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param requestID path string true "Friend request ID to accept"
// @Success 200 {object} handler.APIResponse "Friend request accepted successfully"
// @Failure 400 {object} handler.APIResponse "Invalid request"
// @Failure 401 {object} handler.APIResponse "Unauthorized"
// @Failure 404 {object} handler.APIResponse "Friend request not found"
// @Failure 500 {object} handler.APIResponse "Internal server error"
// @Router /api/v1/friends/requests/{requestID}/accept [post]
func AcceptFriendRequest(c *gin.Context, friendUseCase friend.FriendUseCase) {
	requestID := c.Param("requestID")
	if requestID == "" {
		handler.SendErrorResponse(c, 400, "Request ID is required")
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		handler.SendErrorResponse(c, 401, "Unauthorized: User ID not found")
		return
	}

	userIdStr, ok := userId.(string)
	if !ok {
		handler.SendErrorResponse(c, 500, "Failed to process user identity")
		return
	}

	err := friendUseCase.AcceptFriendRequest(c, userIdStr, requestID)
	if err != nil {
		handler.SendErrorResponse(c, 500, "Failed to accept friend request")
		return
	}

	handler.SendSuccessResponse(c, 200, "Friend request accepted successfully", nil)
}

// RejectFriendRequest godoc
// @Summary Reject a friend request
// @Description Reject a pending friend request from another user
// @Tags friends
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param requestID path string true "Friend request ID to reject"
// @Success 200 {object} handler.APIResponse "Friend request rejected successfully"
// @Failure 400 {object} handler.APIResponse "Invalid request"
// @Failure 401 {object} handler.APIResponse "Unauthorized"
// @Failure 404 {object} handler.APIResponse "Friend request not found"
// @Failure 500 {object} handler.APIResponse "Internal server error"
// @Router /api/v1/friends/requests/{requestID}/reject [post]
func RejectFriendRequest(c *gin.Context, friendUseCase friend.FriendUseCase) {
	requestID := c.Param("requestID")
	if requestID == "" {
		handler.SendErrorResponse(c, 400, "Request ID is required")
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		handler.SendErrorResponse(c, 401, "Unauthorized: User ID not found")
		return
	}

	userIdStr, ok := userId.(string)
	if !ok {
		handler.SendErrorResponse(c, 500, "Failed to process user identity")
		return
	}

	err := friendUseCase.RejectFriendRequest(c, userIdStr, requestID)
	if err != nil {
		handler.SendErrorResponse(c, 500, "Failed to reject friend request")
		return
	}

	handler.SendSuccessResponse(c, 200, "Friend request rejected successfully", nil)
}

// DeleteFriend godoc
// @Summary Remove a friend
// @Description Remove a user from the authenticated user's friends list
// @Tags friends
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param friendID path string true "ID of friend to remove"
// @Success 200 {object} handler.APIResponse "Friend removed successfully"
// @Failure 400 {object} handler.APIResponse "Invalid request"
// @Failure 401 {object} handler.APIResponse "Unauthorized"
// @Failure 404 {object} handler.APIResponse "Friend not found"
// @Failure 500 {object} handler.APIResponse "Internal server error"
// @Router /api/v1/friends/{friendID} [delete]
func DeleteFriend(c *gin.Context, friendUseCase friend.FriendUseCase) {
	friendID := c.Param("friendID")
	if friendID == "" {
		handler.SendErrorResponse(c, 400, "Friend ID is required")
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		handler.SendErrorResponse(c, 401, "Unauthorized: User ID not found")
		return
	}

	userIdStr, ok := userId.(string)
	if !ok {
		handler.SendErrorResponse(c, 500, "Failed to process user identity")
		return
	}

	err := friendUseCase.DeleteFriend(c, userIdStr, friendID)
	if err != nil {
		handler.SendErrorResponse(c, 500, "Failed to delete friend")
		return
	}

	handler.SendSuccessResponse(c, 200, "Friend deleted successfully", nil)
}
