package domain

type MessageType string

const (
	TextMessageType  MessageType = "TEXT"
	ImageMessageType MessageType = "IMAGE"
	VideoMessageType MessageType = "VIDEO"
	AudioMessageType MessageType = "AUDIO"
	FileMessageType  MessageType = "FILE"
)

type Message struct {
	ID         string      `json:"id"`
	SenderId   string      `json:"sender_id"`
	ReceiverId string      `json:"receiver_id"`
	Type       MessageType `json:"type"`
	MimeType   string      `json:"mime_type,omitempty"`
	Content    string      `json:"content"`
	CreatedAt  string      `json:"created_at"`
}

type ChatRoom struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Type        string   `json:"type"` // "GROUP" or "PRIVATE"
	CreatedAt   string   `json:"created_at"`
	LastMessage *Message `json:"last_message"`
}
