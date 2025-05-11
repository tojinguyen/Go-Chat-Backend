package friend

import (
	"context"
	"fmt"
	"log"
)

func (f *friendUseCase) GetFriendList(ctx context.Context, userID string, page int, limit int) ([]*FriendOutput, error) {
	// Check if user ID is provided
	if userID == "" {
		log.Println("User ID is required")
		return nil, fmt.Errorf("user ID is required")
	}

	// Validate and apply pagination parameters
	if limit <= 0 {
		limit = 100 // Default limit
	}
	if page <= 0 {
		page = 1 // Default to first page
	}

	// Calculate offset based on page and limit
	offset := (page - 1) * limit

	// Fetch friends from repository with pagination
	friends, err := f.friendShipRepo.FindFriendsByUserId(ctx, userID, limit, offset)
	if err != nil {
		log.Printf("Error fetching friends for user %s: %v", userID, err)
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
