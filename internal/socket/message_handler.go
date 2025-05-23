package socket

import (
	"encoding/json"
	"gochat-backend/internal/repository"
	"log"
	"time"
)

type MessageHandler struct {
	hub                *Hub
	chatRoomRepository repository.ChatRoomRepository
	messageRepository  repository.MessageRepository
	accountRepository  repository.AccountRepository
}

func NewMessageHandler(
	hub *Hub,
	chatRoomRepository repository.ChatRoomRepository,
	messageRepository repository.MessageRepository,
	accountRepository repository.AccountRepository,
) *MessageHandler {
	return &MessageHandler{
		hub:                hub,
		chatRoomRepository: chatRoomRepository,
		messageRepository:  messageRepository,
		accountRepository:  accountRepository,
	}
}

// HandleSocketMessage xử lý tin nhắn từ client
func (h *MessageHandler) HandleSocketMessage(client *Client, data []byte) {
	var socketMsg SocketMessage
	if err := json.Unmarshal(data, &socketMsg); err != nil {
		h.sendErrorToClient(client, "Message format is invalid")
		log.Printf("Error parsing message: %v", err)
		return
	}

	// Suren sender ID and timestamp
	socketMsg.SenderID = client.ID
	socketMsg.Timestamp = time.Now().UnixMilli()

	log.Printf("Received message from client %s: %s", client.ID, string(data))

	switch socketMsg.Type {
	case SocketMessageTypeChat:
		h.handleChatMessage(client, socketMsg)
	case SocketMessageTypeJoin:
		h.handleJoinMessage(client, socketMsg)
	case SocketMessageTypeLeave:
		h.handleLeaveMessage(client, socketMsg)
	case SocketMessageTypeTyping:
		h.handleTypingMessage(client, socketMsg)
	case SocketMessageTypeReadReceipt:
		h.handleReadReceiptMessage(client, socketMsg)
	default:
		h.sendErrorToClient(client, "Message type is not supported")
		log.Printf("Unsupported message type: %s", socketMsg.Type)
	}
}

// handleChatMessage xử lý tin nhắn chat
func (h *MessageHandler) handleChatMessage(client *Client, socketMsg SocketMessage) {
	// Kiểm tra client đã join phòng này chưa
	if !h.hub.IsClientInRoom(socketMsg.ChatRoomID, client.ID) {
		h.sendErrorToClient(client, "Bạn chưa tham gia phòng chat này")
		return
	}

	// Parse payload từ Data
	// payload, err := ParsePayload[ChatMessagePayload](socketMsg.Data)
	// if err != nil {
	// 	h.sendErrorToClient(client, "Dữ liệu chat không hợp lệ")
	// 	return
	// }

	h.hub.BroadcastToRoom(socketMsg.ChatRoomID, socketMsg)
}

// handleJoinMessage xử lý yêu cầu tham gia phòng
func (h *MessageHandler) handleJoinMessage(client *Client, socketMsg SocketMessage) {
	// Kiểm tra xem có ChatRoomID không
	if socketMsg.ChatRoomID == "" {
		h.sendErrorToClient(client, "Thiếu thông tin phòng chat")
		return
	}

	// Kiểm tra quyền tham gia phòng (có thể thêm logic ở đây)
	// ...

	// Tham gia phòng
	h.hub.JoinRoomWithResponse(socketMsg.ChatRoomID, client)
}

// handleLeaveMessage xử lý yêu cầu rời phòng
func (h *MessageHandler) handleLeaveMessage(client *Client, socketMsg SocketMessage) {
	if socketMsg.ChatRoomID == "" {
		h.sendErrorToClient(client, "Thiếu thông tin phòng chat")
		return
	}
	h.hub.LeaveRoom(socketMsg.ChatRoomID, client)
}

// handleTypingMessage xử lý thông báo đang gõ
func (h *MessageHandler) handleTypingMessage(client *Client, socketMsg SocketMessage) {
	if !h.hub.IsClientInRoom(socketMsg.ChatRoomID, client.ID) {
		return // Bỏ qua nếu không ở trong phòng
	}
	h.hub.BroadcastToRoom(socketMsg.ChatRoomID, socketMsg)
}

// handleReadReceiptMessage xử lý xác nhận đã đọc
func (h *MessageHandler) handleReadReceiptMessage(client *Client, socketMsg SocketMessage) {
	if !h.hub.IsClientInRoom(socketMsg.ChatRoomID, client.ID) {
		return
	}
	h.hub.BroadcastToRoom(socketMsg.ChatRoomID, socketMsg)
}

// sendErrorToClient gửi thông báo lỗi cho client
func (h *MessageHandler) sendErrorToClient(client *Client, errorMsg string) {
	msg := SocketMessage{
		Type:      SocketMessageTypeError,
		SenderID:  "system",
		Timestamp: time.Now().UnixMilli(),
	}

	// Chuyển errorMsg thành JSON và gán vào Data
	data, _ := json.Marshal(map[string]string{"message": errorMsg})
	msg.Data = data

	messageJSON, _ := json.Marshal(msg)
	select {
	case client.Send <- messageJSON:
		// Gửi thành công
	default:
		// Không thể gửi
		log.Printf("Failed to send error message to client %s", client.ID)
	}
}
