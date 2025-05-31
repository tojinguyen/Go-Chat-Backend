package socket

import (
	"context"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second    // Thời gian chờ tối đa để ghi một message.
	pongWait       = 60 * time.Second    // Thời gian chờ tối đa cho một pong message từ client.
	pingPeriod     = (pongWait * 9) / 10 // Phải nhỏ hơn pongWait.
	maxMessageSize = 4096                // Kích thước tối đa của message.
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
		log.Printf("Client %s: ReadPump stopped and unregistered.", c.ID)
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().UTC().Add(pongWait)) // Thiết lập deadline đọc ban đầu
	c.Conn.SetPongHandler(func(string) error {             // Xử lý khi nhận được Pong message
		c.Conn.SetReadDeadline(time.Now().UTC().Add(pongWait))
		return nil
	})

	for {
		select {
		case <-c.ctx.Done():
			log.Printf("Client %s: Context done, exiting ReadPump.", c.ID)
			return
		default:
			// Đọc message từ WebSocket
			_, message, err := c.Conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNoStatusReceived) {
					log.Printf("Client %s: Error reading message: %v", c.ID, err)
				} else {
					log.Printf("Client %s: WebSocket closed: %v", c.ID, err) // Client tự đóng hoặc lỗi mạng
				}
				return // Thoát vòng lặp, defer sẽ được gọi
			}
			// Chuyển tin nhắn đến Hub xử lý với context
			c.Hub.HandleMessageWithContext(c, message, c.ctx)
		}
	}
}

// Gửi message tới client
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod) // Tạo ticker để gửi Ping định kỳ
	defer func() {
		ticker.Stop()
		c.Conn.Close()
		log.Printf("Client %s: WritePump stopped.", c.ID)
	}()

	for {
		select {
		case <-c.ctx.Done():
			log.Printf("Client %s: Context done, exiting WritePump.", c.ID)
			// Gửi CloseMessage khi context bị hủy (ví dụ từ ReadPump)
			c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait)) // Đặt deadline cho việc ghi
			if !ok {
				// Kênh Send đã bị đóng (thường do Hub đóng khi unregister)
				log.Printf("Client %s: Send channel closed.", c.ID)
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Ghi message ra WebSocket
			err := c.Conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Printf("Client %s: Error writing message: %v", c.ID, err)
				return // Thoát vòng lặp, defer sẽ được gọi
			}
		case <-ticker.C:
			// Gửi ping định kỳ để giữ kết nối
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Client %s: Error sending ping: %v", c.ID, err)
				return // Thoát vòng lặp, defer sẽ được gọi
			}
		}
	}
}
