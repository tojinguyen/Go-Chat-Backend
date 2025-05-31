package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"gochat-backend/internal/infra/redisinfra"
	"time"

	"gochat-backend/internal/domain/status"

	"github.com/redis/go-redis/v9"
)

const (
	userStatusKeyPrefix        = "user_status:"
	userStatusOfflineTTL       = 1 * time.Hour  // TTL khi user offline
	userStatusOnlineDefaultTTL = 24 * time.Hour // TTL mặc định dài cho user online nếu không có explicit offline
)

// StatusRepository defines the interface for status persistence
type StatusRepository interface {
	SetUserStatus(ctx context.Context, userID string, userStatus *status.UserStatus) error
	GetUserStatus(ctx context.Context, userID string) (*status.UserStatus, error)
	DeleteUserStatus(ctx context.Context, userID string) error // Có thể không cần nếu dùng TTL
}

type redisStatusRepository struct {
	redisService redisinfra.RedisService
}

func NewRedisStatusRepository(redisService redisinfra.RedisService) StatusRepository { // Trả về interface
	return &redisStatusRepository{
		redisService: redisService,
	}
}

func (r *redisStatusRepository) generateKey(userID string) string {
	return userStatusKeyPrefix + userID
}

func (r *redisStatusRepository) SetUserStatus(ctx context.Context, userID string, userStatus *status.UserStatus) error {
	key := r.generateKey(userID)
	data, err := json.Marshal(userStatus)
	if err != nil {
		return fmt.Errorf("failed to marshal user status: %w", err)
	}

	var ttl time.Duration
	if userStatus.Status == status.Offline {
		ttl = userStatusOfflineTTL
	} else {
		// Có thể không cần TTL cho online nếu disconnect luôn set offline.
		// Hoặc đặt TTL dài để tự dọn dẹp nếu client crash.
		ttl = userStatusOnlineDefaultTTL
	}

	return r.redisService.Set(ctx, key, data, ttl)
}

func (r *redisStatusRepository) GetUserStatus(ctx context.Context, userID string) (*status.UserStatus, error) {
	key := r.generateKey(userID)
	var userStatus status.UserStatus
	err := r.redisService.Get(ctx, key, &userStatus)
	if err != nil {
		if err == redis.Nil {
			return nil, redis.Nil // Trả về lỗi redis.Nil để usecase biết key không tồn tại
		}
		return nil, fmt.Errorf("failed to get user status from redis: %w", err)
	}
	return &userStatus, nil
}

func (r *redisStatusRepository) DeleteUserStatus(ctx context.Context, userID string) error {
	key := r.generateKey(userID)
	return r.redisService.Delete(ctx, key)
}
