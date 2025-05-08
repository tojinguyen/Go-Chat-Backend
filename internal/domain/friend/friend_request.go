package domain

import "time"

type RequestFriendStatus string

const (
	Pending   RequestFriendStatus = "pending"
	Accepted  RequestFriendStatus = "accepted"
	Rejected  RequestFriendStatus = "rejected"
	Cancelled RequestFriendStatus = "cancelled"
)

type FriendRequest struct {
	UserIdRequester string              `json:"user_id_requester"`
	UserIdReceiver  string              `json:"user_id_receiver"`
	CreatedAt       time.Time           `json:"created_at"`
	Status          RequestFriendStatus `json:"status"`
}
