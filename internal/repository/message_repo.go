package repository

import (
	"context"
	"database/sql"
	"errors"
	domain "gochat-backend/internal/domain/chat"
	"gochat-backend/internal/infra/mysqlinfra"
	"time"
)

type MessageRepository interface {
	CreateMessage(ctx context.Context, message *domain.Message) error
	FindMessageByID(ctx context.Context, messageID string) (*domain.Message, error)
	FindMessagesByChatRoomID(ctx context.Context, chatRoomID string, limit, offset int) ([]*domain.Message, error)
	DeleteMessage(ctx context.Context, messageID string) error
	DeleteMessagesByChatRoomID(ctx context.Context, chatRoomID string) error
}

type messageRepo struct {
	database *mysqlinfra.Database
}

func NewMessageRepo(db *mysqlinfra.Database) MessageRepository {
	return &messageRepo{database: db}
}

// CreateMessage creates a new message
func (r *messageRepo) CreateMessage(ctx context.Context, message *domain.Message) error {
	query := `
        INSERT INTO messages (id, sender_id, receiver_id, type, mime_type, content, created_at, chat_room_id)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `

	// Set created_at if not provided
	if message.CreatedAt == "" {
		message.CreatedAt = time.Now().Format(time.RFC3339)
	}

	_, err := r.database.DB.ExecContext(
		ctx,
		query,
		message.ID,
		message.SenderId,
		message.ReceiverId,
		message.Type,
		message.MimeType,
		message.Content,
		message.CreatedAt,
		message.ChatRoomId,
	)

	return err
}

// FindMessageByID retrieves a message by its ID
func (r *messageRepo) FindMessageByID(ctx context.Context, messageID string) (*domain.Message, error) {
	query := `
        SELECT id, sender_id, receiver_id, type, mime_type, content, created_at, chat_room_id
        FROM messages
        WHERE id = ?
    `

	var message domain.Message
	var messageType string

	err := r.database.DB.QueryRowContext(ctx, query, messageID).
		Scan(
			&message.ID,
			&message.SenderId,
			&message.ReceiverId,
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

// FindMessagesByChatRoomID retrieves messages for a chat room with pagination
func (r *messageRepo) FindMessagesByChatRoomID(ctx context.Context, chatRoomID string, limit, offset int) ([]*domain.Message, error) {
	query := `
        SELECT id, sender_id, receiver_id, type, mime_type, content, created_at, chat_room_id
        FROM messages
        WHERE chat_room_id = ?
        ORDER BY created_at DESC
        LIMIT ? OFFSET ?
    `

	rows, err := r.database.DB.QueryContext(ctx, query, chatRoomID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*domain.Message
	for rows.Next() {
		var message domain.Message
		var messageType string

		if err := rows.Scan(
			&message.ID,
			&message.SenderId,
			&message.ReceiverId,
			&messageType,
			&message.MimeType,
			&message.Content,
			&message.CreatedAt,
			&message.ChatRoomId,
		); err != nil {
			return nil, err
		}

		message.Type = domain.MessageType(messageType)
		messages = append(messages, &message)
	}

	return messages, nil
}

// DeleteMessage deletes a message by its ID
func (r *messageRepo) DeleteMessage(ctx context.Context, messageID string) error {
	query := `DELETE FROM messages WHERE id = ?`
	_, err := r.database.DB.ExecContext(ctx, query, messageID)
	return err
}

// DeleteMessagesByChatRoomID deletes all messages in a chat room
func (r *messageRepo) DeleteMessagesByChatRoomID(ctx context.Context, chatRoomID string) error {
	query := `DELETE FROM messages WHERE chat_room_id = ?`
	_, err := r.database.DB.ExecContext(ctx, query, chatRoomID)
	return err
}
