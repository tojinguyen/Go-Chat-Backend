package socket

import (
	"context"
	"encoding/json"
	"fmt"
	"gochat-backend/internal/infra/kafkainfra"
	"gochat-backend/internal/repository"
	"gochat-backend/internal/usecase"
	"gochat-backend/internal/usecase/status"
	"log"
	"sync"
	"time"
)

// ChatRoomActiveView quản lý các client đang chủ động xem một phòng chat cụ thể.
type ChatRoomActiveView struct {
	ID      string
	Clients map[string]*Client
	mutex   sync.RWMutex
}

// Hub quản lý các phòng chat và kết nối
type Hub struct {
	ActiveRoomViews map[string]*ChatRoomActiveView
	Clients         map[string]*Client // Map userId -> client

	Register   chan *Client
	Unregister chan *Client

	mutex sync.RWMutex

	MessageHandler *MessageHandler

	statusUseCase status.StatusUseCase

	accountRepo  repository.AccountRepository
	chatRoomRepo repository.ChatRoomRepository

	kafkaService *kafkainfra.KafkaService
}

// NewHub khởi tạo Hub mới
func NewHub(deps *usecase.SharedDependencies, statusUseCase status.StatusUseCase) *Hub {
	hub := &Hub{
		ActiveRoomViews: make(map[string]*ChatRoomActiveView),
		Clients:         make(map[string]*Client),
		Register:        make(chan *Client),
		Unregister:      make(chan *Client),
		statusUseCase:   statusUseCase,
		accountRepo:     deps.AccountRepo,
		chatRoomRepo:    deps.ChatRoomRepo,
		kafkaService:    deps.KafkaService,
	}

	hub.MessageHandler = NewMessageHandler(
		hub,
		deps.ChatRoomRepo,
		deps.MessageRepo,
		deps.AccountRepo,
	)
	return hub
}

// Run khởi chạy Hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mutex.Lock()
			h.Clients[client.ID] = client
			h.mutex.Unlock()
			log.Printf("Client %s registered to hub", client.ID)

		case client := <-h.Unregister:
			userID := client.ID

			// Cập nhật trạng thái offline
			if err := h.statusUseCase.SetUserOffline(context.Background(), userID); err != nil {
				log.Printf("Error setting user %s offline: %v", userID, err)
			}

			h.removeClientFromAllActiveViews(client)

			h.mutex.Lock()
			if _, exists := h.Clients[userID]; exists {
				delete(h.Clients, userID)
				log.Printf("Client %s (UserID) unregistered from Hub. Total online: %d", userID, len(h.Clients))
			}

			h.mutex.Unlock()
			log.Printf("Client %s unregistered from hub", client.ID)
		}
	}
}

func (h *Hub) HandleMessageWithContext(client *Client, data []byte, ctx context.Context) {
	h.MessageHandler.HandleSocketMessageWithContext(client, data, ctx)
}

// Gửi tin nhắn đến TẤT CẢ THÀNH VIÊN (DB) của phòng đang online
func (h *Hub) DeliverMessageToRoomRecipients(ctx context.Context, chatRoomID string, message SocketMessage) {
	messageJSON, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshalling message for room %s: %v", chatRoomID, err)
		return
	}

	// 1. Lấy danh sách thành viên của phòng từ DB
	roomMembersDB, err := h.chatRoomRepo.FindChatRoomMembers(ctx, chatRoomID)
	if err != nil {
		log.Printf("Hub: Error fetching DB members for room %s: %v", chatRoomID, err)
		return
	}

	if len(roomMembersDB) == 0 {
		log.Printf("Hub: No DB members found for room %s. Message not delivered.", chatRoomID)
		return
	}

	log.Printf("Hub: Delivering message type '%s' to %d DB members of room %s", message.Type, len(roomMembersDB), chatRoomID)

	for _, member := range roomMembersDB {
		recipientID := member.UserId

		// 2. Kiểm tra xem thành viên có online không và lấy *Client
		h.mutex.RLock()
		recipientClient, isOnline := h.Clients[recipientID]
		h.mutex.RUnlock()

		if isOnline && recipientClient != nil {
			select {
			case recipientClient.Send <- messageJSON:
				log.Printf("Hub: Message sent to online client %s for room %s", recipientID, chatRoomID)
			default:
				log.Printf("Hub: Send channel for client %s is full or closed. Message for room %s might be dropped for this client.", recipientID, chatRoomID)
			}
		} else {
			// Tin nhắn đã được lưu vào DB, user offline sẽ thấy khi online lại
			log.Printf("Hub: Recipient %s for room %s is offline. Message stored in DB for later retrieval.", recipientID, chatRoomID)
		}
	}
}

func (h *Hub) BroadcastToRoom(chatRoomID string, message SocketMessage) {
	h.mutex.RLock()
	room, exists := h.ActiveRoomViews[chatRoomID]
	h.mutex.RUnlock()

	if !exists {
		log.Printf("Cannot broadcast to non-existent chat room: %s", chatRoomID)
		return
	}

	messageJSON, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshalling message: %v", err)
		return
	}

	room.mutex.RLock()
	defer room.mutex.RUnlock()

	var failedClients []*Client

	for clientID, client := range room.Clients {
		select {
		case client.Send <- messageJSON:
			// Gửi thành công
		default:
			// Kênh đầy hoặc bị đóng
			failedClients = append(failedClients, client)
			log.Printf("Failed to send message to client %s", clientID)
		}
	}

	// Xử lý các client không nhận được tin nhắn
	for _, client := range failedClients {
		h.removeClientFromAllActiveViews(client)
	}
}

// --- Quản lý Active Room Views ---
// JoinActiveRoomView xử lý khi client muốn chủ động xem một phòng chat.
func (h *Hub) JoinActiveRoomView(chatRoomID string, client *Client) {
	h.mutex.Lock() // Lock để truy cập ActiveRoomViews
	if _, exists := h.ActiveRoomViews[chatRoomID]; !exists {
		h.ActiveRoomViews[chatRoomID] = &ChatRoomActiveView{
			ID:      chatRoomID,
			Clients: make(map[string]*Client),
		}
	}
	activeView := h.ActiveRoomViews[chatRoomID]
	h.mutex.Unlock()

	activeView.mutex.Lock()
	_, alreadyInView := activeView.Clients[client.ID]
	if !alreadyInView {
		activeView.Clients[client.ID] = client
	}
	activeView.mutex.Unlock()

	// Gửi phản hồi join thành công cho client
	joinSuccessPayload := JoinSuccessPayload{ChatRoomID: chatRoomID, Status: "joined_active_view"}
	successMsg := SocketMessage{
		Type:      SocketMessageTypeJoinSuccess,
		SenderID:  "system",
		Timestamp: time.Now().UnixMilli(),
		Data:      mustMarshal(joinSuccessPayload),
	}
	client.Send <- mustMarshal(successMsg)

	if alreadyInView {
		log.Printf("Client %s re-confirmed active view for room %s", client.ID, chatRoomID)
	} else {
		log.Printf("Client %s joined active view for room %s", client.ID, chatRoomID)
		// Thông báo cho các client *khác* đang active trong view này
		userJoinedPayload := UserEventPayload{ChatRoomID: chatRoomID, UserID: client.ID} // Cần lấy thêm name, avatar từ AccountRepo nếu muốn
		account, _ := h.accountRepo.FindById(context.TODO(), client.ID)
		if account != nil {
			userJoinedPayload.UserName = account.Name
			userJoinedPayload.AvatarURL = account.AvatarURL
		}

		payloadBytes, err := json.Marshal(userJoinedPayload)

		if err != nil {
			log.Printf("Hub: Failed to marshal UserJoined payload: %v", err)
		} else {
			kafkaEvent := &kafkainfra.MQEvent{
				EventType:  kafkainfra.UserJoinedRoom,
				ChatRoomID: chatRoomID,
				SenderID:   client.ID,
				Timestamp:  time.Now().UTC(),
				Metadata:   payloadBytes,
			}

			h.kafkaService.PublishChatEvent(context.Background(), kafkaEvent)
		}
	}
}

// LeaveActiveRoomView xử lý khi client không còn xem phòng chat đó nữa.
func (h *Hub) LeaveActiveRoomView(chatRoomID string, client *Client) {
	h.mutex.RLock() // RLock để đọc ActiveRoomViews
	activeView, exists := h.ActiveRoomViews[chatRoomID]
	h.mutex.RUnlock()

	if !exists {
		log.Printf("Client %s tried to leave non-existent active view %s", client.ID, chatRoomID)
		return
	}

	userActuallyLeftView := false
	activeView.mutex.Lock()
	if _, clientWasInView := activeView.Clients[client.ID]; clientWasInView {
		delete(activeView.Clients, client.ID)
		userActuallyLeftView = true
	}
	currentActiveViewersCount := len(activeView.Clients)
	activeView.mutex.Unlock()

	if userActuallyLeftView {
		log.Printf("Client %s left active view for room %s", client.ID, chatRoomID)
		userLeftPayload := UserEventPayload{
			ChatRoomID: chatRoomID,
			UserID:     client.ID,
		}

		payloadBytes, err := json.Marshal(userLeftPayload)

		if err != nil {
			log.Printf("Hub: Failed to marshal UserLeft payload: %v", err)
		} else {
			kafkaEvent := &kafkainfra.MQEvent{
				EventType:  kafkainfra.UserLeftRoom,
				ChatRoomID: chatRoomID,
				SenderID:   client.ID,
				Timestamp:  time.Now().UTC(),
				Metadata:   payloadBytes,
			}

			h.kafkaService.PublishChatEvent(context.Background(), kafkaEvent)
		}

		userLeftMsg := SocketMessage{
			Type:      SocketMessageTypeUserLeft,
			SenderID:  "system",
			Timestamp: time.Now().UnixMilli(),
			Data:      mustMarshal(userLeftPayload),
		}
		h.broadcastToActiveView(chatRoomID, userLeftMsg, "") // Gửi cho tất cả active (bao gồm cả client nếu họ vẫn còn listen)

		if currentActiveViewersCount == 0 {
			h.mutex.Lock() // Lock để xóa ActiveRoomViews[chatRoomID]
			delete(h.ActiveRoomViews, chatRoomID)
			h.mutex.Unlock()
			log.Printf("Active room view %s deleted (no active viewers)", chatRoomID)
		} else {
			h.sendActiveUsersListToView(chatRoomID) // Cập nhật danh sách cho những người còn lại
		}
	}
}

// IsClientInActiveView kiểm tra client có đang active trong view không.
func (h *Hub) IsClientInActiveView(chatRoomID, clientID string) bool {
	h.mutex.RLock()
	activeView, exists := h.ActiveRoomViews[chatRoomID]
	h.mutex.RUnlock()

	if !exists {
		return false
	}

	activeView.mutex.RLock()
	_, clientIsViewing := activeView.Clients[clientID]
	activeView.mutex.RUnlock()
	return clientIsViewing
}

// removeClientFromAllActiveViews xóa client khỏi tất cả active views khi client disconnect.
func (h *Hub) removeClientFromAllActiveViews(client *Client) {
	h.mutex.RLock()
	// Sao chép key để tránh deadlock khi gọi LeaveActiveRoomView
	activeViewIDs := make([]string, 0, len(h.ActiveRoomViews))
	for id := range h.ActiveRoomViews {
		activeViewIDs = append(activeViewIDs, id)
	}
	h.mutex.RUnlock()

	for _, viewID := range activeViewIDs {
		h.LeaveActiveRoomView(viewID, client) // Hàm này đã có logging bên trong
	}
	log.Printf("Client %s removed from all active views.", client.ID)
}

// sendActiveUsersListToView gửi danh sách client đang active trong một view.
func (h *Hub) sendActiveUsersListToView(chatRoomID string) {
	h.mutex.RLock()
	activeView, exists := h.ActiveRoomViews[chatRoomID]
	h.mutex.RUnlock()

	if !exists {
		return
	}

	activeView.mutex.RLock()
	// Lấy thông tin chi tiết của user từ AccountRepo
	usersPayloadList := make([]UserEventPayload, 0, len(activeView.Clients))
	for userID := range activeView.Clients {
		account, err := h.accountRepo.FindById(context.TODO(), userID) // Sử dụng context phù hợp
		if err == nil && account != nil {
			usersPayloadList = append(usersPayloadList, UserEventPayload{
				UserID:    account.Id,
				UserName:  account.Name,
				AvatarURL: account.AvatarURL,
			})
		} else {
			usersPayloadList = append(usersPayloadList, UserEventPayload{UserID: userID}) // Fallback
			log.Printf("Hub: Could not get account details for active user %s in view %s: %v", userID, chatRoomID, err)
		}
	}
	activeView.mutex.RUnlock()

	activeUsersListPayload := ActiveUsersListPayload{ChatRoomID: chatRoomID, Users: usersPayloadList}
	usersListMsg := SocketMessage{
		Type:      SocketMessageTypeUsers,
		SenderID:  "system",
		Timestamp: time.Now().UnixMilli(),
		Data:      mustMarshal(activeUsersListPayload),
	}
	h.broadcastToActiveView(chatRoomID, usersListMsg, "") // Gửi cho tất cả trong active view
}

// broadcastToActiveView gửi message đến tất cả client đang active trong một view.
// `excludeClientID` để không gửi lại cho sender (nếu cần).
func (h *Hub) broadcastToActiveView(chatRoomID string, message SocketMessage, excludeClientID string) {
	h.mutex.RLock()
	activeView, exists := h.ActiveRoomViews[chatRoomID]
	h.mutex.RUnlock()

	if !exists {
		log.Printf("Hub: Cannot broadcast to non-existent active view: %s", chatRoomID)
		return
	}

	messageJSON, err := json.Marshal(message)
	if err != nil {
		log.Printf("Hub: Error marshalling message for active view %s: %v", chatRoomID, err)
		return
	}

	activeView.mutex.RLock() // Chỉ cần RLock để đọc danh sách client
	// Tạo một slice copy của clients để tránh giữ lock lâu khi gửi
	clientsToSend := make([]*Client, 0, len(activeView.Clients))
	for _, client := range activeView.Clients {
		if client.ID != excludeClientID {
			clientsToSend = append(clientsToSend, client)
		}
	}
	activeView.mutex.RUnlock()

	for _, client := range clientsToSend {
		select {
		case client.Send <- messageJSON:
		default:
			log.Printf("Hub: Send channel for client %s in active view %s is full/closed.", client.ID, chatRoomID)
		}
	}
}

// startKafkaConsumer khởi động consumer Kafka để nhận tin nhắn
func (h *Hub) startKafkaConsumer() {
	ctx := context.Background()
	err := h.kafkaService.StartChatConsumer(ctx, h.handleKafkaEvent)
	if err != nil {
		log.Printf("Error starting Kafka consumer: %v", err)
	}
}

// handleKafkaEvent xử lý sự kiện từ Kafka
func (h *Hub) handleKafkaEvent(event *kafkainfra.MQEvent) error {
	log.Printf("Received Kafka event: %s for room %s from user %s",
		event.EventType, event.ChatRoomID, event.SenderID)

	var socketMsg SocketMessage

	switch event.EventType {
	case kafkainfra.MessageSent:
		var payload ChatMessageReceivePayload
		metadataBytes, ok := event.Metadata.([]byte)
		if !ok {
			metadataBytes, ok = event.Metadata.(json.RawMessage)
			if !ok {
				return fmt.Errorf("expected event.Metadata to be []byte or json.RawMessage, got %T", event.Metadata)
			}
		}

		if err := json.Unmarshal(metadataBytes, &payload); err != nil {
			return fmt.Errorf("failed to unmarshal message payload: %w", err)
		}

		socketMsg = SocketMessage{
			Type:      SocketMessageTypeNewMessage,
			SenderID:  event.SenderID,
			Timestamp: event.Timestamp.UnixMilli(),
			Data:      mustMarshal(payload),
		}

		h.DeliverMessageToRoomRecipients(context.Background(), event.ChatRoomID, socketMsg)

	case kafkainfra.TypingStarted:
		var payload TypingPayload
		metadataBytes, ok := event.Metadata.([]byte)
		if !ok {
			metadataBytes, ok = event.Metadata.(json.RawMessage)
			if !ok {
				return fmt.Errorf("expected event.Metadata to be []byte or json.RawMessage, got %T", event.Metadata)
			}
		}

		if err := json.Unmarshal(metadataBytes, &payload); err != nil {
			return fmt.Errorf("failed to unmarshal typing payload: %w", err)
		}

		typingMsg := SocketMessage{
			Type:      SocketMessageTypeTyping,
			SenderID:  event.SenderID,
			Timestamp: time.Now().UTC().UnixMilli(),
			Data:      mustMarshal(payload),
		}

		h.broadcastToActiveView(payload.ChatRoomID, typingMsg, event.SenderID)
	case kafkainfra.TypingStopped:
		var payload TypingPayload
		metadataBytes, ok := event.Metadata.([]byte)
		if !ok {
			metadataBytes, ok = event.Metadata.(json.RawMessage)
			if !ok {
				return fmt.Errorf("expected event.Metadata to be []byte or json.RawMessage, got %T", event.Metadata)
			}
		}

		if err := json.Unmarshal(metadataBytes, &payload); err != nil {
			return fmt.Errorf("failed to unmarshal typing payload: %w", err)
		}

		typingMsg := SocketMessage{
			Type:      SocketMessageTypeTyping,
			SenderID:  event.SenderID,
			Timestamp: time.Now().UTC().UnixMilli(),
			Data:      mustMarshal(payload),
		}

		h.broadcastToActiveView(payload.ChatRoomID, typingMsg, event.SenderID)

	case kafkainfra.UserJoinedRoom:
		var payload UserEventPayload
		metadataBytes, ok := event.Metadata.([]byte)
		if !ok {
			metadataBytes, ok = event.Metadata.(json.RawMessage)
			if !ok {
				return fmt.Errorf("expected event.Metadata to be []byte or json.RawMessage, got %T", event.Metadata)
			}
		}

		if err := json.Unmarshal(metadataBytes, &payload); err != nil {
			return fmt.Errorf("failed to unmarshal user joined payload: %w", err)
		}

		userJoinedMsg := SocketMessage{
			Type:      SocketMessageTypeUserJoined,
			SenderID:  "system",
			Timestamp: time.Now().UTC().UnixMilli(),
			Data:      mustMarshal(payload),
		}

		h.broadcastToActiveView(payload.ChatRoomID, userJoinedMsg, payload.UserID)
		h.sendActiveUsersListToView(payload.ChatRoomID)

	case kafkainfra.UserLeftRoom:
		var payload UserEventPayload
		metadataBytes, ok := event.Metadata.([]byte)
		if !ok {
			metadataBytes, ok = event.Metadata.(json.RawMessage)
			if !ok {
				return fmt.Errorf("expected event.Metadata to be []byte or json.RawMessage, got %T", event.Metadata)
			}
		}

		if err := json.Unmarshal(metadataBytes, &payload); err != nil {
			return fmt.Errorf("failed to unmarshal user left payload: %w", err)
		}

		userLeftMsg := SocketMessage{
			Type:      SocketMessageTypeUserLeft,
			SenderID:  "system",
			Timestamp: time.Now().UTC().UnixMilli(),
			Data:      mustMarshal(payload),
		}

		h.broadcastToActiveView(payload.ChatRoomID, userLeftMsg, payload.UserID)
	default:
		return fmt.Errorf("unsupported event type: %s", event.EventType)
	}
	return nil
}
