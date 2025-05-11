package friend

import (
	"context"
	"fmt"
	"time"
)

func (f *friendUseCase) GetFriendRequestList(ctx context.Context, userID string) ([]*FriendRequestOutput, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	// Get friend requests for the user from repository
	requests, err := f.friendRequestRepo.GetFriendRequestsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch friend requests: %w", err)
	}

	// Transform to FriendRequestOutput format
	result := make([]*FriendRequestOutput, 0, len(requests))

	// We would need the account repository to get the requester's info
	// This is not injected yet, so for now we'll just use the available data
	for _, req := range requests {
		// Format the created time as a string
		createdAtStr := req.CreatedAt.Format(time.RFC3339)

		// In a complete implementation, we'd fetch the requester's name and avatar from the account repository
		result = append(result, &FriendRequestOutput{
			ID:          "", // We'd need to add an ID field to the FriendRequest model
			RequesterID: req.UserIdRequester,
			// These fields would come from the account repo in a complete implementation
			RequesterName: "",
			AvatarURL:     "",
			CreatedAt:     createdAtStr,
			Status:        string(req.Status),
		})
	}

	return result, nil
}
