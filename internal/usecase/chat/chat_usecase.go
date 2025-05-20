package chat

import (
	"context"
	"fmt"
	domain "gochat-backend/internal/domain/chat"
	"gochat-backend/internal/repository"
	"time"

	"github.com/google/uuid"
)

type ChatRoomCreateInput struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"` // "GROUP" or "PRIVATE"
	Members []string `json:"members"`
}

type ChatRoomOutput struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Type        string             `json:"type"`
	CreatedAt   string             `json:"created_at"`
	MemberCount int                `json:"member_count"`
	Members     []ChatMemberOutput `json:"members,omitempty"`
	LastMessage *MessageOutput     `json:"last_message,omitempty"`
}

type ChatRoomMembersInput struct {
	Members []string `json:"members"`
}

type ChatMemberOutput struct {
	UserID    string `json:"user_id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	JoinedAt  string `json:"joined_at"`
}

type MessageInput struct {
	Type     domain.MessageType `json:"type"`
	MimeType string             `json:"mime_type,omitempty"`
	Content  string             `json:"content"`
}

type MessageOutput struct {
	ID         string             `json:"id"`
	SenderID   string             `json:"sender_id"`
	SenderName string             `json:"sender_name"`
	AvatarURL  string             `json:"avatar_url"`
	Type       domain.MessageType `json:"type"`
	MimeType   string             `json:"mime_type,omitempty"`
	Content    string             `json:"content"`
	CreatedAt  string             `json:"created_at"`
	ChatRoomID string             `json:"chat_room_id"`
}

type ChatUseCase interface {
	CreateChatRoom(ctx context.Context, userID string, input ChatRoomCreateInput) (*ChatRoomOutput, error)
	GetChatRooms(ctx context.Context, userID string) ([]*ChatRoomOutput, error)
	GetChatRoomByID(ctx context.Context, userID, chatRoomID string) (*ChatRoomOutput, error)
	AddChatRoomMembers(ctx context.Context, userID, chatRoomID string, memberIDs []string) error
	RemoveChatRoomMember(ctx context.Context, userID, chatRoomID, memberID string) error
	GetChatRoomMessages(ctx context.Context, userID, chatRoomID string, page, limit int) ([]*MessageOutput, error)
	SendMessage(ctx context.Context, userID, chatRoomID string, input MessageInput) (*MessageOutput, error)
	LeaveChatRoom(ctx context.Context, userID, chatRoomID string) error
	FindOrCreatePrivateChatRoom(ctx context.Context, currentUserID, otherUserID string) (*ChatRoomOutput, error)
}

type chatUseCase struct {
	chatRoomRepository repository.ChatRoomRepository
	messageRepository  repository.MessageRepository
	accountRepository  repository.AccountRepository
}

func NewChatUseCase(
	chatRoomRepository repository.ChatRoomRepository,
	messageRepository repository.MessageRepository,
	accountRepository repository.AccountRepository,
) ChatUseCase {
	return &chatUseCase{
		chatRoomRepository: chatRoomRepository,
		messageRepository:  messageRepository,
		accountRepository:  accountRepository,
	}
}

// CreateChatRoom creates a new chat room
func (c *chatUseCase) CreateChatRoom(ctx context.Context, userID string, input ChatRoomCreateInput) (*ChatRoomOutput, error) {
	// Validate input
	if input.Name == "" {
		return nil, fmt.Errorf("chat room name cannot be empty")
	}

	if input.Type != "GROUP" && input.Type != "PRIVATE" {
		return nil, fmt.Errorf("invalid chat room type: must be GROUP or PRIVATE")
	}

	// For private chats, only allow 2 members
	if input.Type == "PRIVATE" && len(input.Members) != 1 {
		return nil, fmt.Errorf("private chats must have exactly one other member")
	}

	// Check if members exist
	for _, memberID := range input.Members {
		account, err := c.accountRepository.FindById(ctx, memberID)
		if err != nil {
			return nil, fmt.Errorf("error finding member %s: %w", memberID, err)
		}
		if account == nil {
			return nil, fmt.Errorf("member %s does not exist", memberID)
		}
	}

	// For private chats, check if a chat already exists between these users
	if input.Type == "PRIVATE" {
		otherUserID := input.Members[0]
		existingRoom, err := c.chatRoomRepository.FindPrivateChatRoom(ctx, userID, otherUserID)
		if err != nil {
			return nil, fmt.Errorf("error checking existing chat rooms: %w", err)
		}

		if existingRoom != nil {
			// Convert to output format
			return c.convertChatRoomToOutput(ctx, existingRoom)
		}
	}

	// Create chat room
	now := time.Now().Format(time.RFC3339)
	chatRoom := &domain.ChatRoom{
		ID:        uuid.New().String(),
		Name:      input.Name,
		Type:      input.Type,
		CreatedAt: now,
	}

	// Start a transaction
	err := c.chatRoomRepository.CreateChatRoom(ctx, chatRoom)
	if err != nil {
		return nil, fmt.Errorf("error creating chat room: %w", err)
	}

	// Add the creator as a member
	creatorMember := &domain.ChatRoomMember{
		ChatRoomId: chatRoom.ID,
		UserId:     userID,
		JoinedAt:   now,
	}

	err = c.chatRoomRepository.AddChatRoomMember(ctx, creatorMember)
	if err != nil {
		return nil, fmt.Errorf("error adding creator to chat room: %w", err)
	}

	// Add other members
	for _, memberID := range input.Members {
		member := &domain.ChatRoomMember{
			ChatRoomId: chatRoom.ID,
			UserId:     memberID,
			JoinedAt:   now,
		}
		err = c.chatRoomRepository.AddChatRoomMember(ctx, member)
		if err != nil {
			return nil, fmt.Errorf("error adding member %s to chat room: %w", memberID, err)
		}
	}

	// Get the complete chat room with members
	createdChatRoom, err := c.chatRoomRepository.FindChatRoomByID(ctx, chatRoom.ID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving created chat room: %w", err)
	}

	return c.convertChatRoomToOutput(ctx, createdChatRoom)
}

// GetChatRooms gets all chat rooms for a user
func (c *chatUseCase) GetChatRooms(ctx context.Context, userID string) ([]*ChatRoomOutput, error) {
	// Get all chat rooms where the user is a member
	chatRooms, err := c.chatRoomRepository.FindChatRoomsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error finding chat rooms: %w", err)
	}

	// Convert to output format
	output := make([]*ChatRoomOutput, 0, len(chatRooms))
	for _, chatRoom := range chatRooms {
		chatRoomOutput, err := c.convertChatRoomToOutput(ctx, chatRoom)
		if err != nil {
			return nil, err
		}
		output = append(output, chatRoomOutput)
	}

	return output, nil
}

// GetChatRoomByID gets a specific chat room by ID
func (c *chatUseCase) GetChatRoomByID(ctx context.Context, userID, chatRoomID string) (*ChatRoomOutput, error) {
	// Check if the chat room exists
	chatRoom, err := c.chatRoomRepository.FindChatRoomByID(ctx, chatRoomID)
	if err != nil {
		return nil, fmt.Errorf("error finding chat room: %w", err)
	}

	if chatRoom == nil {
		return nil, fmt.Errorf("chat room not found")
	}

	// Check if the user is a member of the chat room
	isMember, err := c.chatRoomRepository.IsUserMemberOfChatRoom(ctx, userID, chatRoomID)
	if err != nil {
		return nil, fmt.Errorf("error checking chat room membership: %w", err)
	}

	if !isMember {
		return nil, fmt.Errorf("user is not a member of this chat room")
	}

	// Convert to output format with members
	return c.convertChatRoomToOutput(ctx, chatRoom)
}

// AddChatRoomMembers adds members to a chat room
func (c *chatUseCase) AddChatRoomMembers(ctx context.Context, userID, chatRoomID string, memberIDs []string) error {
	// Check if the chat room exists
	chatRoom, err := c.chatRoomRepository.FindChatRoomByID(ctx, chatRoomID)
	if err != nil {
		return fmt.Errorf("error finding chat room: %w", err)
	}

	if chatRoom == nil {
		return fmt.Errorf("chat room not found")
	}

	// Check if the user is a member of the chat room
	isMember, err := c.chatRoomRepository.IsUserMemberOfChatRoom(ctx, userID, chatRoomID)
	if err != nil {
		return fmt.Errorf("error checking chat room membership: %w", err)
	}

	if !isMember {
		return fmt.Errorf("user is not a member of this chat room")
	}

	// Private chats can't have additional members
	if chatRoom.Type == "PRIVATE" {
		return fmt.Errorf("cannot add members to a private chat")
	}

	// Add members
	now := time.Now().Format(time.RFC3339)
	for _, memberID := range memberIDs {
		// Check if the user exists
		account, err := c.accountRepository.FindById(ctx, memberID)
		if err != nil {
			return fmt.Errorf("error finding user %s: %w", memberID, err)
		}
		if account == nil {
			return fmt.Errorf("user %s not found", memberID)
		}

		// Check if already a member
		isMember, err := c.chatRoomRepository.IsUserMemberOfChatRoom(ctx, memberID, chatRoomID)
		if err != nil {
			return fmt.Errorf("error checking if user %s is already a member: %w", memberID, err)
		}

		if isMember {
			continue // Skip if already a member
		}

		// Add member
		member := &domain.ChatRoomMember{
			ChatRoomId: chatRoomID,
			UserId:     memberID,
			JoinedAt:   now,
		}
		err = c.chatRoomRepository.AddChatRoomMember(ctx, member)
		if err != nil {
			return fmt.Errorf("error adding member %s to chat room: %w", memberID, err)
		}
	}

	return nil
}

// RemoveChatRoomMember removes a member from a chat room
func (c *chatUseCase) RemoveChatRoomMember(ctx context.Context, userID, chatRoomID, memberID string) error {
	// Check if the chat room exists
	chatRoom, err := c.chatRoomRepository.FindChatRoomByID(ctx, chatRoomID)
	if err != nil {
		return fmt.Errorf("error finding chat room: %w", err)
	}

	if chatRoom == nil {
		return fmt.Errorf("chat room not found")
	}

	// Check if the user is a member of the chat room
	isMember, err := c.chatRoomRepository.IsUserMemberOfChatRoom(ctx, userID, chatRoomID)
	if err != nil {
		return fmt.Errorf("error checking chat room membership: %w", err)
	}

	if !isMember {
		return fmt.Errorf("user is not a member of this chat room")
	}

	// Private chats can't remove members
	if chatRoom.Type == "PRIVATE" {
		return fmt.Errorf("cannot remove members from a private chat")
	}

	// Remove member
	err = c.chatRoomRepository.RemoveChatRoomMember(ctx, chatRoomID, memberID)
	if err != nil {
		return fmt.Errorf("error removing member from chat room: %w", err)
	}

	return nil
}

// GetChatRoomMessages gets messages in a chat room with pagination
func (c *chatUseCase) GetChatRoomMessages(ctx context.Context, userID, chatRoomID string, page, limit int) ([]*MessageOutput, error) {
	// Check if the chat room exists
	chatRoom, err := c.chatRoomRepository.FindChatRoomByID(ctx, chatRoomID)
	if err != nil {
		return nil, fmt.Errorf("error finding chat room: %w", err)
	}

	if chatRoom == nil {
		return nil, fmt.Errorf("chat room not found")
	}

	// Check if the user is a member of the chat room
	isMember, err := c.chatRoomRepository.IsUserMemberOfChatRoom(ctx, userID, chatRoomID)
	if err != nil {
		return nil, fmt.Errorf("error checking chat room membership: %w", err)
	}

	if !isMember {
		return nil, fmt.Errorf("user is not a member of this chat room")
	}

	// Get messages with pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20 // default limit
	}

	offset := (page - 1) * limit
	messages, err := c.messageRepository.FindMessagesByChatRoomID(ctx, chatRoomID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error finding messages: %w", err)
	}

	// Convert to output format
	output := make([]*MessageOutput, 0, len(messages))
	for _, message := range messages {
		// Get sender information
		sender, err := c.accountRepository.FindById(ctx, message.SenderId)
		if err != nil {
			return nil, fmt.Errorf("error finding sender: %w", err)
		}

		senderName := "Unknown User"
		avatarURL := ""
		if sender != nil {
			senderName = sender.Name
			avatarURL = sender.AvatarURL
		}

		output = append(output, &MessageOutput{
			ID:         message.ID,
			SenderID:   message.SenderId,
			SenderName: senderName,
			AvatarURL:  avatarURL,
			Type:       message.Type,
			MimeType:   message.MimeType,
			Content:    message.Content,
			CreatedAt:  message.CreatedAt,
			ChatRoomID: message.ChatRoomId,
		})
	}

	return output, nil
}

// SendMessage sends a message to a chat room
func (c *chatUseCase) SendMessage(ctx context.Context, userID, chatRoomID string, input MessageInput) (*MessageOutput, error) {
	// Check if the chat room exists
	chatRoom, err := c.chatRoomRepository.FindChatRoomByID(ctx, chatRoomID)
	if err != nil {
		return nil, fmt.Errorf("error finding chat room: %w", err)
	}

	if chatRoom == nil {
		return nil, fmt.Errorf("chat room not found")
	}

	// Check if the user is a member of the chat room
	isMember, err := c.chatRoomRepository.IsUserMemberOfChatRoom(ctx, userID, chatRoomID)
	if err != nil {
		return nil, fmt.Errorf("error checking chat room membership: %w", err)
	}

	if !isMember {
		return nil, fmt.Errorf("user is not a member of this chat room")
	}

	// Validate message type
	switch input.Type {
	case domain.TextMessageType,
		domain.ImageMessageType,
		domain.VideoMessageType,
		domain.AudioMessageType,
		domain.FileMessageType:
		// Valid types
	default:
		return nil, fmt.Errorf("invalid message type")
	}

	// Create message
	now := time.Now().Format(time.RFC3339)
	message := &domain.Message{
		ID:         uuid.New().String(),
		SenderId:   userID,
		Type:       input.Type,
		MimeType:   input.MimeType,
		Content:    input.Content,
		CreatedAt:  now,
		ChatRoomId: chatRoomID,
	}

	// Save message
	err = c.messageRepository.CreateMessage(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("error creating message: %w", err)
	}

	// Update chat room's last message
	err = c.chatRoomRepository.UpdateLastMessage(ctx, chatRoomID, message)
	if err != nil {
		return nil, fmt.Errorf("error updating last message: %w", err)
	}

	// Get sender information
	sender, err := c.accountRepository.FindById(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error finding sender: %w", err)
	}

	senderName := "Unknown User"
	avatarURL := ""
	if sender != nil {
		senderName = sender.Name
		avatarURL = sender.AvatarURL
	}

	// Return the created message
	return &MessageOutput{
		ID:         message.ID,
		SenderID:   message.SenderId,
		SenderName: senderName,
		AvatarURL:  avatarURL,
		Type:       message.Type,
		MimeType:   message.MimeType,
		Content:    message.Content,
		CreatedAt:  message.CreatedAt,
		ChatRoomID: message.ChatRoomId,
	}, nil
}

// LeaveChatRoom allows a user to leave a chat room
func (c *chatUseCase) LeaveChatRoom(ctx context.Context, userID, chatRoomID string) error {
	// Check if the chat room exists
	chatRoom, err := c.chatRoomRepository.FindChatRoomByID(ctx, chatRoomID)
	if err != nil {
		return fmt.Errorf("error finding chat room: %w", err)
	}

	if chatRoom == nil {
		return fmt.Errorf("chat room not found")
	}

	// Check if the user is a member of the chat room
	isMember, err := c.chatRoomRepository.IsUserMemberOfChatRoom(ctx, userID, chatRoomID)
	if err != nil {
		return fmt.Errorf("error checking chat room membership: %w", err)
	}

	if !isMember {
		return fmt.Errorf("user is not a member of this chat room")
	}

	// For private chats, leaving means deleting the chat
	if chatRoom.Type == "PRIVATE" {
		// Delete all messages
		err = c.messageRepository.DeleteMessagesByChatRoomID(ctx, chatRoomID)
		if err != nil {
			return fmt.Errorf("error deleting messages: %w", err)
		}

		// Remove all members
		err = c.chatRoomRepository.RemoveAllChatRoomMembers(ctx, chatRoomID)
		if err != nil {
			return fmt.Errorf("error removing chat room members: %w", err)
		}

		// Delete the chat room
		err = c.chatRoomRepository.DeleteChatRoom(ctx, chatRoomID)
		if err != nil {
			return fmt.Errorf("error deleting chat room: %w", err)
		}

		return nil
	}

	// For group chats, just remove the user
	return c.chatRoomRepository.RemoveChatRoomMember(ctx, chatRoomID, userID)
}

// Helper function to convert domain.ChatRoom to ChatRoomOutput
func (c *chatUseCase) convertChatRoomToOutput(ctx context.Context, chatRoom *domain.ChatRoom) (*ChatRoomOutput, error) {
	// Get members
	members, err := c.chatRoomRepository.FindChatRoomMembers(ctx, chatRoom.ID)
	if err != nil {
		return nil, fmt.Errorf("error finding chat room members: %w", err)
	}

	// Convert members to output format
	memberOutputs := make([]ChatMemberOutput, 0, len(members))
	for _, member := range members {
		// Get user information
		user, err := c.accountRepository.FindById(ctx, member.UserId)
		if err != nil {
			return nil, fmt.Errorf("error finding user: %w", err)
		}

		userName := "Unknown User"
		avatarURL := ""
		if user != nil {
			userName = user.Name
			avatarURL = user.AvatarURL
		}

		memberOutputs = append(memberOutputs, ChatMemberOutput{
			UserID:    member.UserId,
			Name:      userName,
			AvatarURL: avatarURL,
			JoinedAt:  member.JoinedAt,
		})
	}

	// Convert last message if exists
	var lastMessageOutput *MessageOutput
	if chatRoom.LastMessage != nil {
		// Get sender information
		sender, err := c.accountRepository.FindById(ctx, chatRoom.LastMessage.SenderId)
		if err != nil {
			return nil, fmt.Errorf("error finding sender: %w", err)
		}

		senderName := "Unknown User"
		avatarURL := ""
		if sender != nil {
			senderName = sender.Name
			avatarURL = sender.AvatarURL
		}

		lastMessageOutput = &MessageOutput{
			ID:         chatRoom.LastMessage.ID,
			SenderID:   chatRoom.LastMessage.SenderId,
			SenderName: senderName,
			AvatarURL:  avatarURL,
			Type:       chatRoom.LastMessage.Type,
			MimeType:   chatRoom.LastMessage.MimeType,
			Content:    chatRoom.LastMessage.Content,
			CreatedAt:  chatRoom.LastMessage.CreatedAt,
			ChatRoomID: chatRoom.LastMessage.ChatRoomId,
		}
	}

	return &ChatRoomOutput{
		ID:          chatRoom.ID,
		Name:        chatRoom.Name,
		Type:        chatRoom.Type,
		CreatedAt:   chatRoom.CreatedAt,
		MemberCount: len(members),
		Members:     memberOutputs,
		LastMessage: lastMessageOutput,
	}, nil
}

// Cài đặt phương thức này
func (c *chatUseCase) FindOrCreatePrivateChatRoom(ctx context.Context, currentUserID, otherUserID string) (*ChatRoomOutput, error) {
	// Kiểm tra người dùng thứ 2 có tồn tại không
	otherUser, err := c.accountRepository.FindById(ctx, otherUserID)
	if err != nil {
		return nil, fmt.Errorf("error finding user: %w", err)
	}

	if otherUser == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Tìm chat room riêng tư giữa hai người dùng
	existingRoom, err := c.chatRoomRepository.FindPrivateChatRoom(ctx, currentUserID, otherUserID)
	if err != nil {
		return nil, fmt.Errorf("error checking existing chat rooms: %w", err)
	}

	// Nếu đã tồn tại, trả về chat room đó
	if existingRoom != nil {
		return c.convertChatRoomToOutput(ctx, existingRoom)
	}

	// Tạo tên mặc định cho chat room riêng tư
	roomName := otherUser.Name

	// Tạo chat room mới
	input := ChatRoomCreateInput{
		Name:    roomName,
		Type:    "PRIVATE",
		Members: []string{otherUserID},
	}

	return c.CreateChatRoom(ctx, currentUserID, input)
}
