package friend

import "context"

type FriendOutput struct {
}

type FriendRequestOutput struct {
}

type FriendUseCase interface {
	GetFriendList(ctx context.Context, userID string) ([]*FriendOutput, error)
	AddFriend(ctx context.Context, userID, friendID string) error
	DeleteFriend(ctx context.Context, userID, friendID string) error
	GetFriendRequestList(ctx context.Context, userID string) ([]*FriendRequestOutput, error)
	AcceptFriendRequest(ctx context.Context, userID, requestID string) error
	RejectFriendRequest(ctx context.Context, userID, requestID string) error
}

type friendUseCase struct {
}

func NewFriendUseCase() FriendUseCase {
	return &friendUseCase{}
}
