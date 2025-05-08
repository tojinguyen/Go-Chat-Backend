package domain

import "time"

type Status string

type FriendShip struct {
	Id        string    `json:"id"`
	UserIdA   string    `json:"user_id_a"`
	UserIdB   string    `json:"user_id_b"`
	CreatedAt time.Time `json:"created_at"`
}
