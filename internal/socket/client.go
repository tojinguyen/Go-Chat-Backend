package socket

import (
	"context"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID     string
	Conn   *websocket.Conn
	Send   chan []byte
	Hub    *Hub
	ctx    context.Context
	cancel context.CancelFunc
}

// Đọc message từ client và chuyển đến Hub xử lý
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
		c.cancel()
	}()

	c.Conn.SetReadLimit(4096) // Giới hạn kích thước tin nhắn

	for {
		select {
		case <-c.ctx.Done():
			// Context đã bị hủy, thoát khỏi goroutine
			return
		default:
			_, message, err := c.Conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("Error reading message: %v", err)
				}
				return
			}

			// Chuyển tin nhắn đến Hub xử lý với context
			c.Hub.HandleMessageWithContext(c, message, c.ctx)
		}
	}
}

// Gửi message tới client
func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case <-c.ctx.Done():
			// Context đã bị hủy, thoát khỏi goroutine
			return
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
