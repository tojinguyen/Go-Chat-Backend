package handler

import (
	"gochat-backend/internal/usecase/friend"

	"github.com/gin-gonic/gin"
)

func GetFriends(c *gin.Context, friendUseCase friend.FriendUseCase) {
}

func AddFriend(c *gin.Context, friendUseCase friend.FriendUseCase) {
}

func GetFriendRequestList(c *gin.Context, friendUseCase friend.FriendUseCase) {
}

func AcceptFriendRequest(c *gin.Context, friendUseCase friend.FriendUseCase) {
}

func RejectFriendRequest(c *gin.Context, friendUseCase friend.FriendUseCase) {
}

func DeleteFriend(c *gin.Context, friendUseCase friend.FriendUseCase) {
}
