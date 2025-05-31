package status

import (
	"context"
	"gochat-backend/internal/domain/status"
)

type UserStatusOutput struct {
	UserID      string                `json:"user_id"`
	Status      status.UserStatusType `json:"status"`
	RawLastSeen int64                 `json:"raw_last_seen_unix"` // Timestamp Unix để client tự tính toán nếu cần
}

type StatusUseCase interface {
	SetUserOnline(ctx context.Context, userID string) error
	SetUserOffline(ctx context.Context, userID string) error
	GetUserDisplayStatus(ctx context.Context, userID string) (*UserStatusOutput, error)
}
