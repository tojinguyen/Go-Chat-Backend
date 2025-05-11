package friend

import (
	"context"
	"fmt"
	"strconv"

	domainFriend "gochat-backend/internal/domain/friend"
)

func (f *friendUseCase) RejectFriendRequest(ctx context.Context, userID, requestID string) error {
	// Convert requestID from string to int
	reqID, err := strconv.Atoi(requestID)
	if err != nil {
		return fmt.Errorf("invalid request ID format: %w", err)
	}

	// Get the friend request
	request, err := f.friendRequestRepo.GetFriendRequestByID(ctx, reqID)
	if err != nil {
		return fmt.Errorf("failed to get friend request: %w", err)
	}

	// Check if request exists
	if request == nil || request.UserIdReceiver == "" {
		return fmt.Errorf("friend request not found")
	}

	// Validate that the current user is the receiver of the request
	if request.UserIdReceiver != userID {
		return fmt.Errorf("unauthorized to reject this friend request")
	}

	// Check if the request is in pending status
	if request.Status != domainFriend.Pending {
		return fmt.Errorf("friend request is not in pending status")
	}

	// Update the friend request status to rejected
	if err := f.friendRequestRepo.UpdateFriendRequestStatus(ctx, reqID, domainFriend.Rejected); err != nil {
		return fmt.Errorf("error updating friend request status: %w", err)
	}

	return nil
}
