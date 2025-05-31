package status

import "time"

type UserStatusType string

const (
	Online  UserStatusType = "online"
	Offline UserStatusType = "offline"
)

type UserStatus struct {
	UserID   string         `json:"user_id"`
	Status   UserStatusType `json:"status"`
	LastSeen time.Time      `json:"last_seen"` // Thời điểm cuối cùng user online (nếu offline) hoặc thời điểm set online
}
