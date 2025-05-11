package friend

import (
	"context"
	"fmt"
)

func (f *friendUseCase) GetFriendList(ctx context.Context, userID string) ([]*FriendOutput, error) {
	// Check if user ID is provided
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	// Fetch friends from repository with pagination, setting a default limit
	limit, offset := 100, 0
	friends, err := f.friendShipRepo.FindFriendsByUserId(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch friends: %w", err)
	}

	// Convert domain accounts to FriendOutput objects
	result := make([]*FriendOutput, 0, len(friends))
	for _, friend := range friends {
		result = append(result, &FriendOutput{
			ID:        friend.Id,
			Name:      friend.Name,
			Email:     friend.Email,
			AvatarURL: friend.AvatarURL,
		})
	}

	return result, nil
}
