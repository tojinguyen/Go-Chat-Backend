package socket

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID   string
	Conn *websocket.Conn
	Send chan []byte
	Hub  *SocketManager
}

type SocketManager struct {
	Clients    map[string]*Client
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan []byte
}

// NewSocketManager khởi tạo SocketManager mới
func NewSocketManager() *SocketManager {
	return &SocketManager{
		Clients:    make(map[string]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan []byte),
	}
}

// Run lắng nghe các sự kiện đăng ký/hủy đăng ký/broadcast
func (sm *SocketManager) Run() {
	for {
		select {
		case client := <-sm.Register:
			sm.Clients[client.ID] = client
			log.Printf("Client %s connected", client.ID)
		case client := <-sm.Unregister:
			if _, ok := sm.Clients[client.ID]; ok {
				delete(sm.Clients, client.ID)
				close(client.Send)
				log.Printf("Client %s disconnected", client.ID)
			}
		case message := <-sm.Broadcast:
			for _, client := range sm.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(sm.Clients, client.ID)
				}
			}
		}
	}
}

// ServeWS xử lý upgrade HTTP lên WebSocket và đăng ký client mới
func (sm *SocketManager) ServeWS(w http.ResponseWriter, r *http.Request, clientID string) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
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
		Hub:  sm,
	}
	sm.Register <- client

	// Goroutine đọc và ghi
	go client.readPump()
	go client.writePump()
}

// Đọc message từ client và broadcast
func (c *Client) readPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
		c.Hub.Broadcast <- message
	}
}

// Gửi message tới client
func (c *Client) writePump() {
	defer c.Conn.Close()
	for msg := range c.Send {
		err := c.Conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			break
		}
	}
}
