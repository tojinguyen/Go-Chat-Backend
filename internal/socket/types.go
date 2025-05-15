package socket

import (
	"encoding/json"
	domain "gochat-backend/internal/domain/chat"
	"sync"
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
