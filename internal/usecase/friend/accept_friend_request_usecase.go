package friend

import (
	"context"
	"fmt"
	"strconv"
	"time"

	domainFriend "gochat-backend/internal/domain/friend"
)

func (f *friendUseCase) AcceptFriendRequest(ctx context.Context, userID, requestID string) error {
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
		return fmt.Errorf("unauthorized to accept this friend request")
	}

	// Check if the request is in pending status
	if request.Status != domainFriend.Pending {
		return fmt.Errorf("friend request is not in pending status")
	}

	// Start a transaction to update request and create friendship

	// 1. Create friendship record
	newFriendship := &domainFriend.FriendShip{
		UserIdA:   request.UserIdRequester,
		UserIdB:   request.UserIdReceiver,
		CreatedAt: time.Now(),
	}

	// Check if they are already friends
	exists, err := f.friendShipRepo.HasFriendShip(ctx, request.UserIdRequester, request.UserIdReceiver)
	if err != nil {
		return fmt.Errorf("error checking friendship: %w", err)
	}
	if exists {
		return fmt.Errorf("friendship already exists")
	}

	// Create the friendship record
	if err := f.friendShipRepo.CreateFriendShip(ctx, newFriendship); err != nil {
		return fmt.Errorf("error creating friendship: %w", err)
	}

	// 2. Update friend request status to accepted
	if err := f.friendRequestRepo.UpdateFriendRequestStatus(ctx, reqID, domainFriend.Accepted); err != nil {
		return fmt.Errorf("error updating friend request status: %w", err)
	}

	return nil
}
