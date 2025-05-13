package socket

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID   string
	Conn *websocket.Conn
	Send chan []byte
	Hub  *Hub
}

type SocketManager struct {
	Hub *Hub
}

// NewSocketManager khởi tạo SocketManager mới
func NewSocketManager() *SocketManager {
	hub := NewHub()
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
	go client.readPump()
	go client.writePump()
}

// Đọc message từ client và chuyển đến Hub xử lý
func (c *Client) readPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(4096) // Giới hạn kích thước tin nhắn

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error reading message: %v", err)
			}
			break
		}

		// Chuyển tin nhắn đến Hub xử lý
		c.Hub.HandleMessage(c, message)
	}
}

// Gửi message tới client
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Kênh đã đóng
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(msg)

			// Gửi các tin nhắn đang đợi trong hàng đợi
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			// Gửi ping định kỳ để giữ kết nối
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
