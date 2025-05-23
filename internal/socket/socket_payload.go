package socket

import "encoding/json"

type SocketMessage struct {
	Type       SocketMessageType `json:"type"`
	ChatRoomID string            `json:"chat_room_id,omitempty"`
	SenderID   string            `json:"sender_id"`
	Timestamp  int64             `json:"timestamp"`
	Data       json.RawMessage   `json:"data,omitempty"` // Dữ liệu tùy chọn
}

type ChatMessagePayload struct {
	Content   string `json:"content"`
	MessageID string `json:"message_id,omitempty"`
	MimeType  string `json:"mime_type,omitempty"`
}

type JoinPayload struct {
	RoomID string `json:"room_id"`
}

type LeavePayload struct {
	Reason string `json:"reason,omitempty"`
}

type TypingPayload struct {
	IsTyping bool `json:"is_typing"`
}

type ReadReceiptPayload struct {
	MessageID string `json:"message_id"`
}

type UsersPayload struct {
	UserIDs []string `json:"user_ids"`
}

type JoinSuccessPayload struct {
	RoomID string `json:"room_id"`
	Status string `json:"status"`
}

type UserEventPayload struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name,omitempty"`
}

type ErrorPayload struct {
	Message string `json:"message"`
}

func ParsePayload[T any](data json.RawMessage) (*T, error) {
	var payload T
	err := json.Unmarshal(data, &payload)
	return &payload, err
}
