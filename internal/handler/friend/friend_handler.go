package handler

import (
	"gochat-backend/internal/usecase/friend"

	"github.com/gin-gonic/gin"
)

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
// @Router /friends [get]
func GetFriends(c *gin.Context, friendUseCase friend.FriendUseCase) {
}

// AddFriend godoc
// @Summary Send friend request
// @Description Send a friend request to another user
// @Tags friends
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param friendID body string true "ID of user to add as friend"
// @Success 200 {object} handler.APIResponse "Friend request sent successfully"
// @Failure 400 {object} handler.APIResponse "Invalid request"
// @Failure 401 {object} handler.APIResponse "Unauthorized"
// @Failure 404 {object} handler.APIResponse "User not found"
// @Failure 409 {object} handler.APIResponse "Friend request already sent"
// @Failure 500 {object} handler.APIResponse "Internal server error"
// @Router /friends/requests [post]
func AddFriend(c *gin.Context, friendUseCase friend.FriendUseCase) {
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
// @Router /friends/requests [get]
func GetFriendRequestList(c *gin.Context, friendUseCase friend.FriendUseCase) {
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
// @Router /friends/requests/{requestID}/accept [post]
func AcceptFriendRequest(c *gin.Context, friendUseCase friend.FriendUseCase) {
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
// @Router /friends/requests/{requestID}/reject [post]
func RejectFriendRequest(c *gin.Context, friendUseCase friend.FriendUseCase) {
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
// @Router /friends/{friendID} [delete]
func DeleteFriend(c *gin.Context, friendUseCase friend.FriendUseCase) {
}
