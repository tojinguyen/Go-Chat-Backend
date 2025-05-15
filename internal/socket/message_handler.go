package socket

import (
	"encoding/json"
	"log"
	"time"
)

type MessageHandler struct {
	hub *Hub
}

func NewMessageHandler(hub *Hub) *MessageHandler {
	return &MessageHandler{
		hub: hub,
	}
}

// HandleMessage xử lý tin nhắn từ client
func (h *MessageHandler) HandleMessage(client *Client, data []byte) {
	var socketMsg SocketMessage
	if err := json.Unmarshal(data, &socketMsg); err != nil {
		h.sendErrorToClient(client, "Tin nhắn không hợp lệ")
		log.Printf("Error parsing message: %v", err)
		return
	}

	// Đảm bảo gán người gửi từ client
	socketMsg.SenderID = client.ID
	socketMsg.Timestamp = time.Now().UnixMilli()

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
		h.sendErrorToClient(client, "Loại tin nhắn không được hỗ trợ")
	}
}

// handleChatMessage xử lý tin nhắn chat
func (h *MessageHandler) handleChatMessage(client *Client, socketMsg SocketMessage) {
	// Kiểm tra client đã join phòng này chưa
	if !h.hub.IsClientInRoom(socketMsg.ChatRoomID, client.ID) {
		h.sendErrorToClient(client, "Bạn chưa tham gia phòng chat này")
		return
	}
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
