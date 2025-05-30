package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	domain "gochat-backend/internal/domain/chat"
	"gochat-backend/internal/infra/mysqlinfra"
	"gochat-backend/internal/infra/redisinfra"
	"time"
)

const (
	chatRoomsListCacheKeyPrefix = "chat_rooms_list:"
	chatRoomsListCacheTTLExpiry = 15 * time.Minute // Ví dụ: cache 15 phút

	chatRoomDetailsCacheKeyPrefix = "chat_room_details:"
	chatRoomDetailsCacheTTLExpiry = 5 * time.Minute // Cache chi tiết phòng ngắn hơn
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
	database     *mysqlinfra.Database
	redisService redisinfra.RedisService
}

func NewChatRoomRepo(db *mysqlinfra.Database, redisService redisinfra.RedisService) ChatRoomRepository {
	return &chatRoomRepo{
		database:     db,
		redisService: redisService,
	}
}

// CreateChatRoom creates a new chat room
func (r *chatRoomRepo) CreateChatRoom(ctx context.Context, chatRoom *domain.ChatRoom) error {
	query := `INSERT INTO chat_rooms (id, name, type, created_at) VALUES (?, ?, ?, ?)`

	// Set created_at if not provided
	if chatRoom.CreatedAt.IsZero() {
		chatRoom.CreatedAt = time.Now().UTC()
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
	cacheKey := r.generateChatRoomDetailsCacheKey(chatRoomID)
	var cachedChatRoom domain.ChatRoom

	// 1. Thử lấy từ Cache
	if err := r.redisService.Get(ctx, cacheKey, &cachedChatRoom); err == nil {
		return &cachedChatRoom, nil
	}

	// 2. Nếu không có trong Cache, truy vấn từ CSDL
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

	// 3. Lưu vào Cache
	if err := r.redisService.Set(ctx, cacheKey, &chatRoom, chatRoomDetailsCacheTTLExpiry); err != nil {
		fmt.Printf("Warning: Failed to cache chat room details (chatRoomID: %s): %v\n", chatRoomID, err)
	}

	chatRoom.LastMessage = lastMessage
	return &chatRoom, nil
}

// FindChatRoomsByUserID retrieves all chat rooms a user is a member of
func (r *chatRoomRepo) FindChatRoomsByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.ChatRoom, error) {
	cacheKey := r.generateChatRoomsListCacheKey(userID, limit, offset)
	var cachedChatRooms []*domain.ChatRoom

	// 1. Thử lấy từ Cache
	if err := r.redisService.Get(ctx, cacheKey, &cachedChatRooms); err == nil {
		return cachedChatRooms, nil
	}

	// 2. Nếu không có trong Cache, truy vấn từ CSDL
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

	// 3. Lưu vào Cache
	if len(chatRooms) > 0 {
		if err := r.redisService.Set(ctx, cacheKey, chatRooms, chatRoomsListCacheTTLExpiry); err != nil {
			fmt.Printf("Warning: Failed to cache chat rooms list (userID: %s): %v\n", userID, err)
		}
	} else {
		// Cache mảng rỗng
		if err := r.redisService.Set(ctx, cacheKey, []*domain.ChatRoom{}, chatRoomsListCacheTTLExpiry); err != nil {
			fmt.Printf("Warning: Failed to cache empty chat rooms list (userID: %s): %v\n", userID, err)
		}
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

	r.invalidateChatRoomDetailsCache(ctx, chatRoomID)
	// Nếu bạn muốn cập nhật cache ngay lập tức (thay vì đợi lần đọc tiếp theo):
	// 1. Lấy ChatRoom từ DB (hoặc từ cache nếu có nhưng sẽ hơi cũ)
	// 2. Cập nhật chatRoom.LastMessage = message
	// 3. Set lại vào cache
	// Tuy nhiên, việc này có thể dẫn đến race condition nếu có nhiều tin nhắn đến cùng lúc.
	// Invalidate và để lần đọc sau tự fill cache thường an toàn hơn.
	return nil
}

// DeleteChatRoom deletes a chat room by its ID
func (r *chatRoomRepo) DeleteChatRoom(ctx context.Context, chatRoomID string) error {
	// ... (code xóa DB giữ nguyên)
	err := r.database.ExecuteTransaction(func(tx *sql.Tx) error {
		memberQuery := `DELETE FROM chat_room_members WHERE chat_room_id = ?`
		if _, err := tx.ExecContext(ctx, memberQuery, chatRoomID); err != nil {
			return err
		}
		messageQuery := `DELETE FROM messages WHERE chat_room_id = ?`
		if _, err := tx.ExecContext(ctx, messageQuery, chatRoomID); err != nil {
			return err
		}
		roomQuery := `DELETE FROM chat_rooms WHERE id = ?`
		_, err := tx.ExecContext(ctx, roomQuery, chatRoomID)
		return err
	})
	if err != nil {
		return err
	}
	// Invalidate cache chi tiết phòng
	r.invalidateChatRoomDetailsCache(ctx, chatRoomID)
	// Cần invalidate cache danh sách phòng của TẤT CẢ member của phòng này.
	// Điều này đòi hỏi phải lấy danh sách member trước khi xóa.
	// Hoặc chấp nhận eventual consistency.
	return nil
}

// AddChatRoomMember adds a member to a chat room
func (r *chatRoomRepo) AddChatRoomMember(ctx context.Context, member *domain.ChatRoomMember) error {
	query := `INSERT INTO chat_room_members (chat_room_id, user_id, joined_at) VALUES (?, ?, ?)`
	if member.JoinedAt.IsZero() {
		member.JoinedAt = time.Now().UTC()
	}
	_, err := r.database.DB.ExecContext(
		ctx,
		query,
		member.ChatRoomId,
		member.UserId,
		member.JoinedAt,
	)
	if err != nil {
		return err
	}

	// Invalidate cache chi tiết phòng
	r.invalidateChatRoomDetailsCache(ctx, member.ChatRoomId)
	// Invalidate cache danh sách phòng của user vừa được thêm
	r.invalidateUserChatRoomsListCache(ctx, member.UserId)
	return nil
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
	// ... (code xóa member giữ nguyên)
	query := `DELETE FROM chat_room_members WHERE chat_room_id = ? AND user_id = ?`
	_, err := r.database.DB.ExecContext(ctx, query, chatRoomID, userID)
	if err != nil {
		return err
	}

	// Invalidate cache chi tiết phòng
	r.invalidateChatRoomDetailsCache(ctx, chatRoomID)
	// Invalidate cache danh sách phòng của user vừa bị xóa
	r.invalidateUserChatRoomsListCache(ctx, userID)
	return nil
}

// RemoveAllChatRoomMembers removes all members from a chat room
func (r *chatRoomRepo) RemoveAllChatRoomMembers(ctx context.Context, chatRoomID string) error {
	// ... (code xóa tất cả member giữ nguyên)
	// Cần lấy danh sách member TRƯỚC KHI xóa để invalidate cache của họ.
	members, _ := r.FindChatRoomMembers(ctx, chatRoomID) // Bỏ qua lỗi ở đây để cố gắng xóa

	query := `DELETE FROM chat_room_members WHERE chat_room_id = ?`
	_, err := r.database.DB.ExecContext(ctx, query, chatRoomID)
	if err != nil {
		return err
	}
	// Invalidate cache chi tiết phòng
	r.invalidateChatRoomDetailsCache(ctx, chatRoomID)
	// Invalidate cache danh sách phòng của tất cả member
	if members != nil {
		for _, m := range members {
			r.invalidateUserChatRoomsListCache(ctx, m.UserId)
		}
	}
	return nil
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
			&message.ChatRoomId,
			&messageType,
			&message.MimeType,
			&message.Content,
			&message.CreatedAt,
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

func (r *chatRoomRepo) generateChatRoomsListCacheKey(userID string, limit, offset int) string {
	return fmt.Sprintf("%s%s:limit%d:offset%d", chatRoomsListCacheKeyPrefix, userID, limit, offset)
}

func (r *chatRoomRepo) generateChatRoomDetailsCacheKey(chatRoomID string) string {
	return fmt.Sprintf("%s%s", chatRoomDetailsCacheKeyPrefix, chatRoomID)
}

// invalidateChatRoomDetailsCache xóa cache chi tiết của một phòng.
func (r *chatRoomRepo) invalidateChatRoomDetailsCache(ctx context.Context, chatRoomID string) {
	cacheKey := r.generateChatRoomDetailsCacheKey(chatRoomID)
	if err := r.redisService.Delete(ctx, cacheKey); err != nil {
		fmt.Printf("Warning: Failed to delete chat room details cache (chatRoomID: %s): %v\n", chatRoomID, err)
	}
}

// invalidateUserChatRoomsListCache xóa cache danh sách phòng của một user (trang đầu tiên).
// Cần chiến lược tốt hơn cho phân trang.
func (r *chatRoomRepo) invalidateUserChatRoomsListCache(ctx context.Context, userID string) {
	// Đây là cách đơn giản, chỉ xóa trang đầu.
	// Thực tế cần xóa tất cả các trang hoặc dùng versioning.
	cacheKey := r.generateChatRoomsListCacheKey(userID, 20, 0) // Giả sử limit mặc định là 20
	if err := r.redisService.Delete(ctx, cacheKey); err != nil {
		fmt.Printf("Warning: Failed to delete user's chat rooms list cache (userID: %s): %v\n", userID, err)
	}
}
