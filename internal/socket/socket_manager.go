package socket

import (
	"context"
	"gochat-backend/internal/usecase"
	"gochat-backend/internal/usecase/status"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type SocketManager struct {
	Hub           *Hub
	statusUseCase status.StatusUseCase
}

// NewSocketManager khởi tạo SocketManager mới
func NewSocketManager(deps *usecase.SharedDependencies, statusUseCase status.StatusUseCase) *SocketManager {
	hub := NewHub(deps, statusUseCase)
	// Khởi chạy hub trong goroutine riêng
	go hub.Run()

	return &SocketManager{
		Hub:           hub,
		statusUseCase: statusUseCase,
	}
}

// ServeWS xử lý upgrade HTTP lên WebSocket và đăng ký client mới
// clientID ở đây CHÍNH LÀ UserID của người dùng đã xác thực
func (sm *SocketManager) ServeWS(w http.ResponseWriter, r *http.Request, userID string) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// TODO: Implement a proper origin check for production
			// Ví dụ: return r.Header.Get("Origin") == "http://yourfrontend.com"
			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("SocketManager: Failed to upgrade connection for user %s: %v", userID, err)
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}
	log.Printf("SocketManager: WebSocket connection upgraded for user %s", userID)

	// Cập nhật trạng thái user online
	if err := sm.statusUseCase.SetUserOnline(r.Context(), userID); err != nil {
		log.Printf("SocketManager: Error setting user %s online: %v", userID, err)
		// Không hủy kết nối ở đây, vẫn cho phép user kết nối
	}

	// Tạo context cho client này, sẽ bị hủy khi client disconnect
	clientCtx, clientCancel := context.WithCancel(context.Background())

	client := &Client{
		ID:     userID, // Client.ID chính là UserID
		Conn:   conn,
		Send:   make(chan []byte, 256), // Kênh buffered để tránh block
		Hub:    sm.Hub,
		ctx:    clientCtx,
		cancel: clientCancel,
	}

	// Đăng ký client với Hub
	sm.Hub.Register <- client

	// Goroutine đọc và ghi message cho client này
	// ReadPump và WritePump sẽ tự xử lý việc đóng kết nối và unregister khi cần
	go client.WritePump()
	go client.ReadPump() // ReadPump nên chạy sau WritePump để WritePump có thể gửi CloseMessage nếu ReadPump thoát trước

	log.Printf("SocketManager: Client %s registered and pumps started.", userID)
}
