package socket

import (
	"encoding/json"
	domain "gochat-backend/internal/domain/chat"
	"log"
	"sync"
	"time"
)

type SocketMessageType string

const (
	// Tin nhắn từ client
	SocketMessageTypeChat        SocketMessageType = "CHAT"         // Gửi tin nhắn chat
	SocketMessageTypeJoin        SocketMessageType = "JOIN"         // Tham gia phòng chat
	SocketMessageTypeLeave       SocketMessageType = "LEAVE"        // Rời phòng chat
	SocketMessageTypeTyping      SocketMessageType = "TYPING"       // Đang nhập
	SocketMessageTypeReadReceipt SocketMessageType = "READ_RECEIPT" // Đánh dấu đã đọc

	// Tin nhắn từ server
	SocketMessageTypeUsers       SocketMessageType = "USERS"        // Danh sách người dùng
	SocketMessageTypeJoinSuccess SocketMessageType = "JOIN_SUCCESS" // Tham gia phòng thành công
	SocketMessageTypeJoinError   SocketMessageType = "JOIN_ERROR"   // Lỗi khi tham gia phòng
	SocketMessageTypeUserJoined  SocketMessageType = "USER_JOINED"  // Thông báo người dùng khác tham gia
	SocketMessageTypeUserLeft    SocketMessageType = "USER_LEFT"    // Thông báo người dùng khác rời đi
	SocketMessageTypeError       SocketMessageType = "ERROR"        // Thông báo lỗi
)

type SocketMessage struct {
	Type       SocketMessageType `json:"type"`
	ChatRoomID string            `json:"chat_room_id,omitempty"`
	SenderID   string            `json:"sender_id"`
	Message    *domain.Message   `json:"message,omitempty"`
	Timestamp  int64             `json:"timestamp"`
	Data       json.RawMessage   `json:"data,omitempty"` // Dữ liệu tùy chọn
}

type ChatRoomSocket struct {
	ID      string
	Clients map[string]*Client
	mutex   sync.RWMutex
}

// Hub quản lý các phòng chat và kết nối
type Hub struct {
	ChatRooms      map[string]*ChatRoomSocket
	Clients        map[string]*Client // Map clientID -> client
	Register       chan *Client
	Unregister     chan *Client
	mutex          sync.RWMutex
	MessageHandler *MessageHandler
}

// NewHub khởi tạo Hub mới
func NewHub() *Hub {
	hub := &Hub{
		ChatRooms:  make(map[string]*ChatRoomSocket),
		Clients:    make(map[string]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}

	hub.MessageHandler = NewMessageHandler(hub)
	return hub
}

// Run khởi chạy Hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mutex.Lock()
			h.Clients[client.ID] = client
			h.mutex.Unlock()
			log.Printf("Client %s registered to hub", client.ID)

		case client := <-h.Unregister:
			h.removeClientFromAllRooms(client)
			h.mutex.Lock()
			if _, exists := h.Clients[client.ID]; exists {
				delete(h.Clients, client.ID)
				close(client.Send)
			}
			h.mutex.Unlock()
			log.Printf("Client %s unregistered from hub", client.ID)
		}
	}
}

// HandleMessage xử lý tin nhắn từ client
// HandleMessage xử lý tin nhắn từ client
func (h *Hub) HandleMessage(client *Client, data []byte) {
	h.MessageHandler.HandleMessage(client, data)
}

// JoinRoomWithResponse là phiên bản mở rộng của JoinRoom với phản hồi JOIN_SUCCESS
func (h *Hub) JoinRoomWithResponse(chatRoomID string, client *Client) {
	h.mutex.Lock()
	if _, exists := h.ChatRooms[chatRoomID]; !exists {
		h.ChatRooms[chatRoomID] = &ChatRoomSocket{
			ID:      chatRoomID,
			Clients: make(map[string]*Client),
		}
	}
	h.mutex.Unlock()

	room := h.ChatRooms[chatRoomID]
	room.mutex.Lock()
	// Nếu client đã ở trong phòng này rồi, không cần thông báo lại
	alreadyJoined := false
	if _, exists := room.Clients[client.ID]; exists {
		alreadyJoined = true
	} else {
		room.Clients[client.ID] = client
	}
	room.mutex.Unlock()

	// Gửi phản hồi thành công cho client
	joinSuccessMsg := SocketMessage{
		Type:       SocketMessageTypeJoinSuccess,
		ChatRoomID: chatRoomID,
		SenderID:   "system",
		Timestamp:  time.Now().UnixMilli(),
	}

	// Thêm thông tin vào Data
	successData, _ := json.Marshal(map[string]string{
		"room_id": chatRoomID,
		"status":  "joined",
	})
	joinSuccessMsg.Data = successData

	messageJSON, _ := json.Marshal(joinSuccessMsg)
	client.Send <- messageJSON

	// Nếu đã tham gia rồi, không cần thông báo và cập nhật danh sách
	if alreadyJoined {
		return
	}

	// Thông báo người dùng mới tham gia cho các client khác
	joinMsg := SocketMessage{
		Type:       SocketMessageTypeUserJoined,
		ChatRoomID: chatRoomID,
		SenderID:   client.ID,
		Timestamp:  time.Now().UnixMilli(),
	}

	// Broadcast cho tất cả người dùng khác trong phòng
	h.BroadcastToRoom(chatRoomID, joinMsg)

	// Gửi danh sách người dùng cho tất cả
	h.sendUserList(chatRoomID)

	log.Printf("Client %s joined chat room %s", client.ID, chatRoomID)
}

// LeaveRoom xóa client khỏi phòng
func (h *Hub) LeaveRoom(chatRoomID string, client *Client) {
	h.mutex.RLock()
	room, exists := h.ChatRooms[chatRoomID]
	h.mutex.RUnlock()

	if !exists {
		return
	}

	room.mutex.Lock()
	delete(room.Clients, client.ID)
	clientsCount := len(room.Clients)
	room.mutex.Unlock()

	// Thông báo người dùng đã rời đi
	leaveMsg := SocketMessage{
		Type:       SocketMessageTypeLeave,
		ChatRoomID: chatRoomID,
		SenderID:   client.ID,
		Timestamp:  time.Now().UnixMilli(),
	}

	h.BroadcastToRoom(chatRoomID, leaveMsg)

	// Nếu phòng trống, xóa phòng
	if clientsCount == 0 {
		h.mutex.Lock()
		delete(h.ChatRooms, chatRoomID)
		h.mutex.Unlock()
		log.Printf("Chat room %s deleted (empty)", chatRoomID)
	} else {
		h.sendUserList(chatRoomID)
	}

	log.Printf("Client %s left chat room %s", client.ID, chatRoomID)
}

// BroadcastToRoom gửi tin nhắn tới tất cả client trong phòng
func (h *Hub) BroadcastToRoom(chatRoomID string, message SocketMessage) {
	h.mutex.RLock()
	room, exists := h.ChatRooms[chatRoomID]
	h.mutex.RUnlock()

	if !exists {
		log.Printf("Cannot broadcast to non-existent chat room: %s", chatRoomID)
		return
	}

	messageJSON, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshalling message: %v", err)
		return
	}

	room.mutex.RLock()
	defer room.mutex.RUnlock()

	var failedClients []*Client

	for clientID, client := range room.Clients {
		select {
		case client.Send <- messageJSON:
			// Gửi thành công
		default:
			// Kênh đầy hoặc bị đóng
			failedClients = append(failedClients, client)
			log.Printf("Failed to send message to client %s", clientID)
		}
	}

	// Xử lý các client không nhận được tin nhắn
	for _, client := range failedClients {
		h.removeClientFromAllRooms(client)
	}
}

// GetClientCount trả về số lượng client trong một phòng
func (h *Hub) GetClientCount(chatRoomID string) int {
	h.mutex.RLock()
	room, exists := h.ChatRooms[chatRoomID]
	h.mutex.RUnlock()

	if !exists {
		return 0
	}

	room.mutex.RLock()
	defer room.mutex.RUnlock()
	return len(room.Clients)
}

// IsClientInRoom kiểm tra một client có trong phòng không
func (h *Hub) IsClientInRoom(chatRoomID, clientID string) bool {
	h.mutex.RLock()
	room, exists := h.ChatRooms[chatRoomID]
	h.mutex.RUnlock()

	if !exists {
		return false
	}

	room.mutex.RLock()
	defer room.mutex.RUnlock()
	_, found := room.Clients[clientID]
	return found
}

// removeClientFromAllRooms xóa client khỏi tất cả các phòng
func (h *Hub) removeClientFromAllRooms(client *Client) {
	h.mutex.RLock()
	roomIDs := make([]string, 0, len(h.ChatRooms))
	for roomID := range h.ChatRooms {
		roomIDs = append(roomIDs, roomID)
	}
	h.mutex.RUnlock()

	for _, roomID := range roomIDs {
		h.mutex.RLock()
		room, exists := h.ChatRooms[roomID]
		h.mutex.RUnlock()

		if !exists {
			continue
		}

		room.mutex.Lock()
		if _, clientExists := room.Clients[client.ID]; clientExists {
			delete(room.Clients, client.ID)
			clientsCount := len(room.Clients)
			room.mutex.Unlock()

			// Thông báo người dùng đã rời đi
			leaveMsg := SocketMessage{
				Type:       SocketMessageTypeLeave,
				ChatRoomID: roomID,
				SenderID:   client.ID,
				Timestamp:  time.Now().UnixMilli(),
			}
			h.BroadcastToRoom(roomID, leaveMsg)

			// Nếu phòng trống, xóa phòng
			if clientsCount == 0 {
				h.mutex.Lock()
				delete(h.ChatRooms, roomID)
				h.mutex.Unlock()
				log.Printf("Chat room %s deleted (empty)", roomID)
			} else {
				h.sendUserList(roomID)
			}
		} else {
			room.mutex.Unlock()
		}
	}
}

// sendUserList gửi danh sách người dùng trong phòng
func (h *Hub) sendUserList(chatRoomID string) {
	h.mutex.RLock()
	room, exists := h.ChatRooms[chatRoomID]
	h.mutex.RUnlock()

	if !exists {
		return
	}

	room.mutex.RLock()
	userIDs := make([]string, 0, len(room.Clients))
	for userID := range room.Clients {
		userIDs = append(userIDs, userID)
	}
	room.mutex.RUnlock()

	userListJSON, _ := json.Marshal(userIDs)

	usersMsg := SocketMessage{
		Type:       SocketMessageTypeUsers,
		ChatRoomID: chatRoomID,
		SenderID:   "system",
		Timestamp:  time.Now().UnixMilli(),
		Data:       userListJSON,
	}

	h.BroadcastToRoom(chatRoomID, usersMsg)
}
