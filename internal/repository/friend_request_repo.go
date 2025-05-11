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
	CreateFriendRequest(ctx context.Context, senderID, receiverID string) error
	GetFriendRequestByID(ctx context.Context, id int) (*domainFriend.FriendRequest, error)
	GetFriendRequestsByUserID(ctx context.Context, userID string) ([]*domainFriend.FriendRequest, error)
	UpdateFriendRequestStatus(ctx context.Context, id int, status domainFriend.RequestFriendStatus) error
	RemoveFriendRequest(ctx context.Context, id int) error
}

type friendRequestRepo struct {
	database *mysqlinfra.Database
}

func NewFriendRequestRepo(db *mysqlinfra.Database) FriendRequestRepository {
	return &friendRequestRepo{database: db}
}

// CreateFriendRequest tạo một yêu cầu kết bạn mới
func (r *friendRequestRepo) CreateFriendRequest(ctx context.Context, senderID, receiverID string) error {
	query := `
        INSERT INTO friend_requests (user_id_requester, user_id_receiver, created_at, status)
        VALUES (?, ?, ?, ?)
    `
	_, err := r.database.DB.ExecContext(
		ctx,
		query,
		senderID,
		receiverID,
		time.Now(),
		domainFriend.Pending,
	)
	return err
}

// GetFriendRequestByID lấy thông tin yêu cầu kết bạn theo ID
func (r *friendRequestRepo) GetFriendRequestByID(ctx context.Context, id int) (*domainFriend.FriendRequest, error) {
	query := `
        SELECT user_id_requester, user_id_receiver, created_at, status
        FROM friend_requests
        WHERE id = ?
    `
	var req domainFriend.FriendRequest
	var status string
	err := r.database.DB.QueryRowContext(ctx, query, id).
		Scan(&req.UserIdRequester, &req.UserIdReceiver, &req.CreatedAt, &status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	req.Status = domainFriend.RequestFriendStatus(status)
	return &req, nil
}

// GetFriendRequestsByUserID lấy danh sách yêu cầu kết bạn của một user
func (r *friendRequestRepo) GetFriendRequestsByUserID(ctx context.Context, userID string) ([]*domainFriend.FriendRequest, error) {
	query := `
        SELECT user_id_requester, user_id_receiver, created_at, status
        FROM friend_requests
        WHERE user_id_receiver = ?
    `
	rows, err := r.database.DB.QueryContext(ctx, query, userID)
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

// UpdateFriendRequestStatus cập nhật trạng thái yêu cầu kết bạn
func (r *friendRequestRepo) UpdateFriendRequestStatus(ctx context.Context, id int, status domainFriend.RequestFriendStatus) error {
	query := `UPDATE friend_requests SET status = ? WHERE id = ?`

	return r.database.ExecuteTransaction(func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, query, status, id)
		return err
	})
}

// RemoveFriendRequest xóa một yêu cầu kết bạn theo ID
func (r *friendRequestRepo) RemoveFriendRequest(ctx context.Context, id int) error {
	query := `DELETE FROM friend_requests WHERE id = ?`
	_, err := r.database.DB.ExecContext(ctx, query, id)
	return err
}
