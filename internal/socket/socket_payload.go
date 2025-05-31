package socket

import "encoding/json"

type SocketMessage struct {
	Type       SocketMessageType `json:"type"`
	ChatRoomID string            `json:"chat_room_id,omitempty"`
	SenderID   string            `json:"sender_id"`
	Timestamp  int64             `json:"timestamp"`
	Data       json.RawMessage   `json:"data,omitempty"` // Dữ liệu tùy chọn
}

// --- Payloads cho Client -> Server ---

type ChatMessageSendPayload struct {
	Content       string `json:"content"`
	MimeType      string `json:"mime_type,omitempty"`
	TempMessageID string `json:"temp_message_id,omitempty"`
}

type JoinRoomPayload struct {
}

type LeaveRoomPayload struct {
}

type TypingPayload struct {
	IsTyping bool `json:"is_typing"`
}

type ReadReceiptPayload struct {
	MessageID string `json:"message_id"` // ID của tin nhắn đã đọc
}

// --- Payloads cho Server -> Client ---
type ChatMessageReceivePayload struct {
	MessageID  string `json:"message_id"`
	SenderName string `json:"sender_name,omitempty"`
	AvatarURL  string `json:"avatar_url,omitempty"`
	Content    string `json:"content"`
	MimeType   string `json:"mime_type,omitempty"`
}

type UserEventPayload struct {
	UserID    string `json:"user_id"`
	UserName  string `json:"user_name,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type ActiveUsersListPayload struct {
	Users []UserEventPayload `json:"users"`
}

type JoinSuccessPayload struct {
	RoomID          string                      `json:"room_id"`
	Status          string                      `json:"status"`
	InitialMessages []ChatMessageReceivePayload `json:"initial_messages,omitempty"`
}

type ErrorPayload struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

func ParsePayload[T any](data json.RawMessage) (*T, error) {
	var payload T
	if len(data) == 0 {
		return new(T), nil
	}
	err := json.Unmarshal(data, &payload)
	return &payload, err
}

func mustMarshal(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
