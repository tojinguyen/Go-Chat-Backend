package friend

import (
	"context"
	"gochat-backend/internal/repository"
)

type FriendOutput struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

type FriendRequestOutput struct {
	ID            string `json:"id"`
	RequesterID   string `json:"requester_id"`
	RequesterName string `json:"requester_name"`
	AvatarURL     string `json:"avatar_url"`
	CreatedAt     string `json:"created_at"`
	Status        string `json:"status"`
}

type FriendUseCase interface {
	GetFriendList(ctx context.Context, userID string, page int, limit int) ([]*FriendOutput, error)
	AddFriend(ctx context.Context, userID, friendID string) error
	DeleteFriend(ctx context.Context, userID, friendID string) error
	GetFriendRequestList(ctx context.Context, userID string) ([]*FriendRequestOutput, error)
	AcceptFriendRequest(ctx context.Context, userID, requestID string) error
	RejectFriendRequest(ctx context.Context, userID, requestID string) error
}

type friendUseCase struct {
	friendShipRepo    repository.FriendShipRepository
	friendRequestRepo repository.FriendRequestRepository
}

func NewFriendUseCase(
	friendShipRepo repository.FriendShipRepository,
	friendRequestRepo repository.FriendRequestRepository,
) FriendUseCase {
	return &friendUseCase{
		friendShipRepo:    friendShipRepo,
		friendRequestRepo: friendRequestRepo,
	}
}
