package socket

import "encoding/json"

type SocketMessage struct {
	Type      SocketMessageType `json:"type"`
	SenderID  string            `json:"sender_id"`
	Timestamp int64             `json:"timestamp"`
	Data      json.RawMessage   `json:"data,omitempty"` // Dữ liệu tùy chọn
}

// --- Payloads cho Client -> Server ---

type ChatMessageSendPayload struct {
	ChatRoomID    string `json:"chat_room_id,omitempty"`
	Content       string `json:"content"`
	MimeType      string `json:"mime_type,omitempty"`
	TempMessageID string `json:"temp_message_id,omitempty"`
}

type JoinRoomPayload struct {
	ChatRoomID string `json:"chat_room_id,omitempty"`
}

type LeaveRoomPayload struct {
	ChatRoomID string `json:"chat_room_id,omitempty"`
}

type TypingPayload struct {
	ChatRoomID string `json:"chat_room_id,omitempty"`
	IsTyping   bool   `json:"is_typing"`
}

type ReadReceiptPayload struct {
	ChatRoomID string `json:"chat_room_id,omitempty"`
	MessageID  string `json:"message_id"` // ID của tin nhắn đã đọc
}

// --- Payloads cho Server -> Client ---
type ChatMessageReceivePayload struct {
	ChatRoomID string `json:"chat_room_id,omitempty"`
	MessageID  string `json:"message_id"`
	SenderName string `json:"sender_name,omitempty"`
	AvatarURL  string `json:"avatar_url,omitempty"`
	Content    string `json:"content"`
	MimeType   string `json:"mime_type,omitempty"`
}

type UserEventPayload struct {
	ChatRoomID string `json:"chat_room_id,omitempty"`
	UserID     string `json:"user_id"`
	UserName   string `json:"user_name,omitempty"`
	AvatarURL  string `json:"avatar_url,omitempty"`
}

type ActiveUsersListPayload struct {
	ChatRoomID string             `json:"chat_room_id,omitempty"`
	Users      []UserEventPayload `json:"users"`
}

type JoinSuccessPayload struct {
	ChatRoomID      string                      `json:"chat_room_id,omitempty"`
	Status          string                      `json:"status"`
	InitialMessages []ChatMessageReceivePayload `json:"initial_messages,omitempty"`
}

type ErrorPayload struct {
	ChatRoomID string `json:"chat_room_id,omitempty"`
	Message    string `json:"message"`
	Code       string `json:"code,omitempty"`
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
