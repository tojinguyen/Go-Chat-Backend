package repository

import (
	"context"
	"database/sql"
	"errors"
	domainFriend "gochat-backend/internal/domain/friend"
	"gochat-backend/internal/infra/mysqlinfra"
	"time"
)

type FriendRequestRepository interface {
	CreateFriendRequest(senderID, receiverID int) error
	GetFriendRequestByID(id int) (domainFriend.FriendRequest, error)
	GetFriendRequestsByUserID(userID int) ([]*domainFriend.FriendRequest, error)
	RemoveFriendRequest(id int) error
}

type friendRequestRepo struct {
	database *mysqlinfra.Database
}

func NewFriendRequestRepo(db *mysqlinfra.Database) FriendRequestRepository {
	return &friendRequestRepo{database: db}
}

// CreateFriendRequest tạo một yêu cầu kết bạn mới
func (r *friendRequestRepo) CreateFriendRequest(senderID, receiverID int) error {
	query := `
        INSERT INTO friend_requests (user_id_requester, user_id_receiver, created_at, status)
        VALUES (?, ?, ?, ?)
    `
	_, err := r.database.DB.ExecContext(
		context.Background(),
		query,
		senderID,
		receiverID,
		time.Now(),
		domainFriend.Pending,
	)
	return err
}

// GetFriendRequestByID lấy thông tin yêu cầu kết bạn theo ID
func (r *friendRequestRepo) GetFriendRequestByID(id int) (domainFriend.FriendRequest, error) {
	query := `
        SELECT user_id_requester, user_id_receiver, created_at, status
        FROM friend_requests
        WHERE id = ?
    `
	var req domainFriend.FriendRequest
	var status string
	err := r.database.DB.QueryRowContext(context.Background(), query, id).
		Scan(&req.UserIdRequester, &req.UserIdReceiver, &req.CreatedAt, &status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return req, nil
		}
		return req, err
	}
	req.Status = domainFriend.RequestFriendStatus(status)
	return req, nil
}

// GetFriendRequestsByUserID lấy danh sách yêu cầu kết bạn của một user
func (r *friendRequestRepo) GetFriendRequestsByUserID(userID int) ([]*domainFriend.FriendRequest, error) {
	query := `
        SELECT user_id_requester, user_id_receiver, created_at, status
        FROM friend_requests
        WHERE user_id_receiver = ?
    `
	rows, err := r.database.DB.QueryContext(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*domainFriend.FriendRequest
	for rows.Next() {
		var req domainFriend.FriendRequest
		var status string
		if err := rows.Scan(&req.UserIdRequester, &req.UserIdReceiver, &req.CreatedAt, &status); err != nil {
			return nil, err
		}
		req.Status = domainFriend.RequestFriendStatus(status)
		requests = append(requests, &req)
	}
	return requests, nil
}

// RemoveFriendRequest xóa một yêu cầu kết bạn theo ID
func (r *friendRequestRepo) RemoveFriendRequest(id int) error {
	query := `DELETE FROM friend_requests WHERE id = ?`
	_, err := r.database.DB.ExecContext(context.Background(), query, id)
	return err
}
