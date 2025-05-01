package domain

import (
	"gochat-backend/pkg/verification"
	"time"
)

type RegistrationVerificationCode struct {
	ID             string                            `json:"id"`
	UserID         string                            `json:"user_id"`
	Email          string                            `json:"email"`
	Name           string                            `json:"name"`            // Thêm tên người dùng
	HashedPassword string                            `json:"hashed_password"` // Lưu mật khẩu đã hash
	Avatar         string                            `json:"avatar"`
	Code           string                            `json:"code"`
	Type           verification.VerificationCodeType `json:"type"`
	Verified       bool                              `json:"verified"`
	ExpiresAt      time.Time                         `json:"expires_at"`
	CreatedAt      time.Time                         `json:"created_at"`
	VerifiedAt     *time.Time                        `json:"verified_at"`
}
