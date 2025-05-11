package friend

import (
	"context"
	"fmt"
	domainFriend "gochat-backend/internal/domain/friend"
	"time"
)

func (f *friendUseCase) AddFriend(ctx context.Context, userID, friendID string) error {
	// Check if userID and friendID are the same
	if userID == friendID {
		return fmt.Errorf("user cannot add themselves as a friend")
	}

	// Check if they are already friends
	exists, err := f.friendShipRepo.HasFriendShip(ctx, userID, friendID)
	if err != nil {
		return fmt.Errorf("error checking friendship: %w", err)
	}
	if exists {
		return fmt.Errorf("friendship already exists")
	}

	// Create a new friendship
	newFriendship := &domainFriend.FriendShip{
		UserIdA:   userID,
		UserIdB:   friendID,
		CreatedAt: time.Now(),
	}

	if err := f.friendShipRepo.CreateFriendShip(ctx, newFriendship); err != nil {
		return fmt.Errorf("error creating friendship: %w", err)
	}

	return nil
}
