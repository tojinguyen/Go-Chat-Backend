package socket

import (
	"context"
	"gochat-backend/internal/usecase"
	"gochat-backend/internal/usecase/status"
	"log"
	"net/http"
	"time"

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
func (sm *SocketManager) ServeWS(w http.ResponseWriter, r *http.Request, clientID string) {
	upgrader := websocket.Upgrader{
		CheckOrigin:     func(r *http.Request) bool { return true },
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	if err := sm.statusUseCase.SetUserOnline(r.Context(), clientID); err != nil {
		log.Printf("Error setting user %s online: %v", clientID, err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	client := &Client{
		ID:     clientID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Hub:    sm.Hub,
		ctx:    ctx,
		cancel: cancel,
	}

	// Đăng ký client với Hub
	sm.Hub.Register <- client

	// Thiết lập ping/pong để giữ kết nối
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Goroutine đọc và ghi
	go client.ReadPump()
	go client.WritePump()
}
