package socket

import (
	"gochat-backend/internal/usecase"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type SocketManager struct {
	Hub *Hub
}

// NewSocketManager khởi tạo SocketManager mới
func NewSocketManager(deps *usecase.SharedDependencies) *SocketManager {
	hub := NewHub(deps)
	// Khởi chạy hub trong goroutine riêng
	go hub.Run()

	return &SocketManager{
		Hub: hub,
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

	client := &Client{
		ID:   clientID,
		Conn: conn,
		Send: make(chan []byte, 256),
		Hub:  sm.Hub,
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
