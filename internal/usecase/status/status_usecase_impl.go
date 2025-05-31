package status

import (
	"context"
	"gochat-backend/config"
	"gochat-backend/internal/domain/status"
	"gochat-backend/internal/repository"
	"time"
)

const (
	maxOfflineDisplayDuration = 1 * time.Hour
)

type statusUseCase struct {
	statusRepo repository.StatusRepository
	cfg        *config.Environment
}

func NewStatusUseCase(repo repository.StatusRepository, cfg *config.Environment) StatusUseCase {
	return &statusUseCase{
		statusRepo: repo,
		cfg:        cfg,
	}
}

func (uc *statusUseCase) SetUserOnline(ctx context.Context, userID string) error {
	userStatus := &status.UserStatus{
		UserID:   userID,
		Status:   status.Online,
		LastSeen: time.Now().UTC(),
	}
	return uc.statusRepo.SetUserStatus(ctx, userID, userStatus)
}

func (uc *statusUseCase) SetUserOffline(ctx context.Context, userID string) error {
	userStatus := &status.UserStatus{
		UserID:   userID,
		Status:   status.Offline,
		LastSeen: time.Now().UTC(),
	}
	return uc.statusRepo.SetUserStatus(ctx, userID, userStatus)
}

func (uc *statusUseCase) GetUserDisplayStatus(ctx context.Context, userID string) (*UserStatusOutput, error) {
	userStatus, err := uc.statusRepo.GetUserStatus(ctx, userID)
	if err != nil {
		// Coi như offline > 1 giờ
		return &UserStatusOutput{
			UserID:      userID,
			Status:      status.Offline,
			RawLastSeen: time.Now().UTC().Add(-(maxOfflineDisplayDuration + time.Minute)).Unix(), // Thời gian cũ
		}, nil
	}

	if userStatus.Status == status.Online {
		return &UserStatusOutput{
			UserID:      userID,
			Status:      status.Online,
			RawLastSeen: userStatus.LastSeen.Unix(),
		}, nil
	}

	// User is Offline
	timeOffline := time.Since(userStatus.LastSeen)

	if timeOffline > maxOfflineDisplayDuration {
		return &UserStatusOutput{
			UserID:      userID,
			Status:      status.Offline,
			RawLastSeen: userStatus.LastSeen.Unix(),
		}, nil
	}

	return &UserStatusOutput{
		UserID:      userID,
		Status:      status.Offline,
		RawLastSeen: userStatus.LastSeen.Unix(),
	}, nil
}
