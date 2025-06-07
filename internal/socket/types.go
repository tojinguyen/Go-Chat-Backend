package socket

type SocketMessageType string

const (
	// Tin nhắn từ client
	SocketMessageTypeChat        SocketMessageType = "CHAT"         // Gửi tin nhắn chat
	SocketMessageTypeJoin        SocketMessageType = "JOIN_ROOM"    // Tham gia phòng chat
	SocketMessageTypeLeave       SocketMessageType = "LEAVE"        // Rời phòng chat
	SocketMessageTypeTyping      SocketMessageType = "TYPING"       // Đang nhập
	SocketMessageTypeReadReceipt SocketMessageType = "READ_RECEIPT" // Đánh dấu đã đọc
	SocketMessageTypePing        SocketMessageType = "PING"         // Tin nhắn ping để kiểm tra kết nối

	// Tin nhắn từ server
	SocketMessageTypeUsers       SocketMessageType = "USERS"        // Danh sách người dùng
	SocketMessageTypeJoinSuccess SocketMessageType = "JOIN_SUCCESS" // Tham gia phòng thành công
	SocketMessageTypeJoinError   SocketMessageType = "JOIN_ERROR"   // Lỗi khi tham gia phòng
	SocketMessageTypeUserJoined  SocketMessageType = "USER_JOINED"  // Thông báo người dùng khác tham gia
	SocketMessageTypeUserLeft    SocketMessageType = "USER_LEFT"    // Thông báo người dùng khác rời đi
	SocketMessageTypeError       SocketMessageType = "ERROR"        // Thông báo lỗi
	SocketMessageTypePong        SocketMessageType = "PONG"         // Tin nhắn pong để phản hồi ping
	SocketMessageTypeNewMessage  SocketMessageType = "NEW_MESSAGE"  // Tin nhắn chat mới (có thể dùng CHAT, nhưng NEW_MESSAGE rõ hơn cho server -> client)
)
