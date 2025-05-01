package domain

import "time"

type VerificationCode struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	Email      string     `json:"email"`
	Code       string     `json:"code"`
	Type       string     `json:"type"`
	Verified   bool       `json:"verified"`
	ExpiresAt  time.Time  `json:"expires_at"`
	CreatedAt  time.Time  `json:"created_at"`
	VerifiedAt *time.Time `json:"verified_at"` // Nullable field for when the code is verified
}
