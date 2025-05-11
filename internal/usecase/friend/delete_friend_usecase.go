package friend

import (
	"context"
	"fmt"
)

func (f *friendUseCase) DeleteFriend(ctx context.Context, userID, friendID string) error {
	// Validate input parameters
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}

	if friendID == "" {
		return fmt.Errorf("friend ID is required")
	}

	// Check if userID and friendID are the same
	if userID == friendID {
		return fmt.Errorf("user cannot delete themselves")
	}

	// Check if they are actually friends
	exists, err := f.friendShipRepo.HasFriendShip(ctx, userID, friendID)
	if err != nil {
		return fmt.Errorf("error checking friendship: %w", err)
	}

	if !exists {
		return fmt.Errorf("friendship does not exist")
	}

	// Remove the friendship
	if err := f.friendShipRepo.RemoveFriendShip(ctx, userID, friendID); err != nil {
		return fmt.Errorf("error deleting friendship: %w", err)
	}

	return nil
}
