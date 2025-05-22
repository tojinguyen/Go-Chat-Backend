package repository

import (
	"context"
	"database/sql"
	"errors"
	domain "gochat-backend/internal/domain/chat"
	"gochat-backend/internal/infra/mysqlinfra"
	"time"
)

type ChatRoomRepository interface {
	CreateChatRoom(ctx context.Context, chatRoom *domain.ChatRoom) error
	FindChatRoomByID(ctx context.Context, chatRoomID string) (*domain.ChatRoom, error)
	FindChatRoomsByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.ChatRoom, error)
	FindPrivateChatRoom(ctx context.Context, userID1, userID2 string) (*domain.ChatRoom, error)
	UpdateLastMessage(ctx context.Context, chatRoomID string, message *domain.Message) error
	DeleteChatRoom(ctx context.Context, chatRoomID string) error

	AddChatRoomMember(ctx context.Context, member *domain.ChatRoomMember) error
	IsUserMemberOfChatRoom(ctx context.Context, userID, chatRoomID string) (bool, error)
	FindChatRoomMembers(ctx context.Context, chatRoomID string) ([]*domain.ChatRoomMember, error)
	RemoveChatRoomMember(ctx context.Context, chatRoomID, userID string) error
	RemoveAllChatRoomMembers(ctx context.Context, chatRoomID string) error
}

type chatRoomRepo struct {
	database *mysqlinfra.Database
}

func NewChatRoomRepo(db *mysqlinfra.Database) ChatRoomRepository {
	return &chatRoomRepo{database: db}
}

// CreateChatRoom creates a new chat room
func (r *chatRoomRepo) CreateChatRoom(ctx context.Context, chatRoom *domain.ChatRoom) error {
	query := `INSERT INTO chat_rooms (id, name, type, created_at) VALUES (?, ?, ?, ?)`

	// Set created_at if not provided
	if chatRoom.CreatedAt == "" {
		chatRoom.CreatedAt = time.Now().Format(time.RFC3339)
	}

	_, err := r.database.DB.ExecContext(
		ctx,
		query,
		chatRoom.ID,
		chatRoom.Name,
		chatRoom.Type,
		chatRoom.CreatedAt,
	)

	return err
}

// FindChatRoomByID retrieves a chat room by its ID
func (r *chatRoomRepo) FindChatRoomByID(ctx context.Context, chatRoomID string) (*domain.ChatRoom, error) {
	query := `
        SELECT cr.id, cr.name, cr.type, cr.created_at
        FROM chat_rooms cr
        WHERE cr.id = ?
    `

	var chatRoom domain.ChatRoom
	err := r.database.DB.QueryRowContext(ctx, query, chatRoomID).
		Scan(
			&chatRoom.ID,
			&chatRoom.Name,
			&chatRoom.Type,
			&chatRoom.CreatedAt,
		)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// Get the last message for this chat room
	lastMessage, err := r.getLastMessage(ctx, chatRoomID)
	if err != nil {
		return nil, err
	}

	chatRoom.LastMessage = lastMessage
	return &chatRoom, nil
}

// getLastMessage retrieves the last message for a chat room
func (r *chatRoomRepo) getLastMessage(ctx context.Context, chatRoomID string) (*domain.Message, error) {
	query := `
        SELECT id, sender_id, chat_room_id, type, mime_type, content, created_at
        FROM messages
        WHERE chat_room_id = ?
        ORDER BY created_at DESC
        LIMIT 1
    `

	var message domain.Message
	var messageType string

	err := r.database.DB.QueryRowContext(ctx, query, chatRoomID).
		Scan(
			&message.ID,
			&message.SenderId,
			&messageType,
			&message.MimeType,
			&message.Content,
			&message.CreatedAt,
			&message.ChatRoomId,
		)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	message.Type = domain.MessageType(messageType)
	return &message, nil
}

// FindChatRoomsByUserID retrieves all chat rooms a user is a member of
func (r *chatRoomRepo) FindChatRoomsByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.ChatRoom, error) {
	query := `
		SELECT cr.id, cr.name, cr.type, cr.created_at
		FROM chat_rooms cr
		JOIN chat_room_members crm ON cr.id = crm.chat_room_id
		WHERE crm.user_id = ?
		ORDER BY cr.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.database.DB.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chatRooms []*domain.ChatRoom
	for rows.Next() {
		var chatRoom domain.ChatRoom

		if err := rows.Scan(
			&chatRoom.ID,
			&chatRoom.Name,
			&chatRoom.Type,
			&chatRoom.CreatedAt,
		); err != nil {
			return nil, err
		}

		// Get the last message for each chat room
		lastMessage, err := r.getLastMessage(ctx, chatRoom.ID)
		if err != nil {
			return nil, err
		}

		chatRoom.LastMessage = lastMessage
		chatRooms = append(chatRooms, &chatRoom)
	}

	return chatRooms, nil
}

// FindPrivateChatRoom finds a private chat room between two users
func (r *chatRoomRepo) FindPrivateChatRoom(ctx context.Context, userID1, userID2 string) (*domain.ChatRoom, error) {
	query := `
        SELECT cr.id
        FROM chat_rooms cr
        JOIN chat_room_members crm1 ON cr.id = crm1.chat_room_id
        JOIN chat_room_members crm2 ON cr.id = crm2.chat_room_id
        WHERE cr.type = 'PRIVATE'
        AND crm1.user_id = ?
        AND crm2.user_id = ?
    `

	var chatRoomID string
	err := r.database.DB.QueryRowContext(ctx, query, userID1, userID2).Scan(&chatRoomID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return r.FindChatRoomByID(ctx, chatRoomID)
}

// UpdateLastMessage updates the last message reference for a chat room
func (r *chatRoomRepo) UpdateLastMessage(ctx context.Context, chatRoomID string, message *domain.Message) error {
	// Since we're retrieving the last message based on the most recent timestamp,
	// we don't need to actually update anything in the chat_rooms table
	// The last message will be retrieved dynamically when fetching the chat room
	return nil
}

// DeleteChatRoom deletes a chat room by its ID
func (r *chatRoomRepo) DeleteChatRoom(ctx context.Context, chatRoomID string) error {
	return r.database.ExecuteTransaction(func(tx *sql.Tx) error {
		// Delete all members first (due to foreign key constraints)
		memberQuery := `DELETE FROM chat_room_members WHERE chat_room_id = ?`
		if _, err := tx.ExecContext(ctx, memberQuery, chatRoomID); err != nil {
			return err
		}

		// Delete all messages
		messageQuery := `DELETE FROM messages WHERE chat_room_id = ?`
		if _, err := tx.ExecContext(ctx, messageQuery, chatRoomID); err != nil {
			return err
		}

		// Delete the chat room
		roomQuery := `DELETE FROM chat_rooms WHERE id = ?`
		_, err := tx.ExecContext(ctx, roomQuery, chatRoomID)
		return err
	})
}

// AddChatRoomMember adds a member to a chat room
func (r *chatRoomRepo) AddChatRoomMember(ctx context.Context, member *domain.ChatRoomMember) error {
	query := `INSERT INTO chat_room_members (chat_room_id, user_id, joined_at) VALUES (?, ?, ?)`

	// Set joined_at if not provided
	if member.JoinedAt == "" {
		member.JoinedAt = time.Now().Format(time.RFC3339)
	}

	_, err := r.database.DB.ExecContext(
		ctx,
		query,
		member.ChatRoomId,
		member.UserId,
		member.JoinedAt,
	)

	return err
}

// IsUserMemberOfChatRoom checks if a user is a member of a chat room
func (r *chatRoomRepo) IsUserMemberOfChatRoom(ctx context.Context, userID, chatRoomID string) (bool, error) {
	query := `
        SELECT COUNT(*)
        FROM chat_room_members
        WHERE chat_room_id = ? AND user_id = ?
    `

	var count int
	err := r.database.DB.QueryRowContext(ctx, query, chatRoomID, userID).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// FindChatRoomMembers retrieves all members of a chat room
func (r *chatRoomRepo) FindChatRoomMembers(ctx context.Context, chatRoomID string) ([]*domain.ChatRoomMember, error) {
	query := `
        SELECT chat_room_id, user_id, joined_at
        FROM chat_room_members
        WHERE chat_room_id = ?
    `

	rows, err := r.database.DB.QueryContext(ctx, query, chatRoomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*domain.ChatRoomMember
	for rows.Next() {
		var member domain.ChatRoomMember

		if err := rows.Scan(
			&member.ChatRoomId,
			&member.UserId,
			&member.JoinedAt,
		); err != nil {
			return nil, err
		}

		members = append(members, &member)
	}

	return members, nil
}

// RemoveChatRoomMember removes a member from a chat room
func (r *chatRoomRepo) RemoveChatRoomMember(ctx context.Context, chatRoomID, userID string) error {
	query := `DELETE FROM chat_room_members WHERE chat_room_id = ? AND user_id = ?`
	_, err := r.database.DB.ExecContext(ctx, query, chatRoomID, userID)
	return err
}

// RemoveAllChatRoomMembers removes all members from a chat room
func (r *chatRoomRepo) RemoveAllChatRoomMembers(ctx context.Context, chatRoomID string) error {
	query := `DELETE FROM chat_room_members WHERE chat_room_id = ?`
	_, err := r.database.DB.ExecContext(ctx, query, chatRoomID)
	return err
}
