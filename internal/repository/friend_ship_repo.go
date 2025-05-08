package repository

import (
	"context"
	"database/sql"
	domainAuth "gochat-backend/internal/domain/auth"
	domainFriendShip "gochat-backend/internal/domain/friend"
	"gochat-backend/internal/infra/mysqlinfra"
)

type FriendShipRepository interface {
	CreateFriendShip(ctx context.Context, friendShip *domainFriendShip.FriendShip) error
	HasFriendShip(ctx context.Context, userId, friendId string) (bool, error)
	FindFriendsByUserId(ctx context.Context, userId string, limit, offset int) ([]*domainAuth.Account, error)
	CountFriendsByUserId(ctx context.Context, userId string) (int, error)
	RemoveFriendShip(ctx context.Context, userId, friendId string) error
}

type FriendRepo struct {
	database *mysqlinfra.Database
}

func NewFriendShipRepo(db *mysqlinfra.Database) FriendShipRepository {
	return &FriendRepo{database: db}
}

func (r *FriendRepo) CreateFriendShip(ctx context.Context, friendShip *domainFriendShip.FriendShip) error {
	query := `INSERT INTO friend_ships (user_id_a, user_id_b, created_at) VALUES (?, ?, ?)`

	return r.database.ExecuteTransaction(func(tx *sql.Tx) error {
		_, err := tx.ExecContext(
			ctx,
			query,
			friendShip.UserIdA,
			friendShip.UserIdB,
			friendShip.CreatedAt,
		)
		return err
	})
}

func (r *FriendRepo) HasFriendShip(ctx context.Context, userId, friendId string) (bool, error) {
	query := `SELECT COUNT(*) FROM friend_ships WHERE (user_id_a = ? AND user_id_b = ?) OR (user_id_a = ? AND user_id_b = ?)`

	var count int
	err := r.database.DB.QueryRowContext(ctx, query, userId, friendId, friendId, userId).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *FriendRepo) FindFriendsByUserId(ctx context.Context, userId string, limit, offset int) ([]*domainAuth.Account, error) {
	query := `
		SELECT u.id, u.name, u.email, u.avatar_url, u.created_at, u.updated_at 
		FROM users AS u 
		JOIN friend_ships AS fs ON (u.id = fs.user_id_a OR u.id = fs.user_id_b) 
		WHERE (fs.user_id_a = ? OR fs.user_id_b = ?) AND u.id != ? 
		LIMIT ? OFFSET ?`

	rows, err := r.database.DB.QueryContext(ctx, query, userId, userId, userId, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var friends []*domainAuth.Account
	for rows.Next() {
		var friend domainAuth.Account
		if err := rows.Scan(&friend.Id, &friend.Name, &friend.Email, &friend.AvatarURL, &friend.CreatedAt, &friend.UpdatedAt); err != nil {
			return nil, err
		}
		friends = append(friends, &friend)
	}

	return friends, nil
}

func (r *FriendRepo) CountFriendsByUserId(ctx context.Context, userId string) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM users AS u 
		JOIN friend_ships AS fs ON (u.id = fs.user_id_a OR u.id = fs.user_id_b) 
		WHERE (fs.user_id_a = ? OR fs.user_id_b = ?) AND u.id != ?`

	var count int
	err := r.database.DB.QueryRowContext(ctx, query, userId, userId, userId).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *FriendRepo) RemoveFriendShip(ctx context.Context, userId, friendId string) error {
	query := `DELETE FROM friend_ships WHERE (user_id_a = ? AND user_id_b = ?) OR (user_id_a = ? AND user_id_b = ?)`

	return r.database.ExecuteTransaction(func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, query, userId, friendId, friendId, userId)
		return err
	})
}
