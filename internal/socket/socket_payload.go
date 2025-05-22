package socket

// Payload cho CHAT
type ChatMessagePayload struct {
	Content   string `json:"content"`
	MessageID string `json:"message_id,omitempty"`
}

// Payload cho JOIN
type JoinPayload struct {
	UserName string `json:"user_name"`
}

// Payload cho LEAVE (nếu cần thêm thông tin)
type LeavePayload struct {
	Reason string `json:"reason,omitempty"`
}

// Payload cho TYPING
type TypingPayload struct {
	IsTyping bool `json:"is_typing"`
}

// Payload cho READ_RECEIPT
type ReadReceiptPayload struct {
	MessageID string `json:"message_id"`
}

// Payload cho USERS (danh sách user trong phòng)
type UsersPayload struct {
	UserIDs []string `json:"user_ids"`
}

// Payload cho ERROR
type ErrorPayload struct {
	Message string `json:"message"`
}
