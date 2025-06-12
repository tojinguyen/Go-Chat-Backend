package kafkainfra

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/google/uuid"
)

// KafkaService manages Kafka producers and consumers
type KafkaService struct {
	chatProducer  *MQEventProducer
	chatConsumer  *MQEventConsumer
	brokers       []string
	chatTopic     string
	consumerGroup string
	mu            sync.RWMutex
	instanceID    string // Unique ID for this instance
}

// NewKafkaService creates a new Kafka service
func NewKafkaService(brokers []string, chatTopic, consumerGroup string) *KafkaService {
	// Tạo instanceID duy nhất cho mỗi instance bằng UUID
	instanceID := uuid.New().String()[:8]

	return &KafkaService{
		brokers:       brokers,
		chatTopic:     chatTopic,
		consumerGroup: consumerGroup,
		instanceID:    instanceID,
	}
}

// Initialize initializes the Kafka service
func (s *KafkaService) Initialize() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Initialize producer
	producer, err := NewChatEventProducer(s.brokers, s.chatTopic)
	if err != nil {
		return fmt.Errorf("failed to initialize chat producer: %w", err)
	}
	s.chatProducer = producer

	// Tạo consumer group ID duy nhất cho mỗi instance
	uniqueConsumerGroup := fmt.Sprintf("%s-%s", s.consumerGroup, s.instanceID)
	log.Printf("Initializing Kafka consumer with unique consumer group: %s", uniqueConsumerGroup)

	// Initialize consumer
	consumer, err := NewChatEventConsumer(s.brokers, s.chatTopic, uniqueConsumerGroup)
	if err != nil {
		return fmt.Errorf("failed to initialize chat consumer: %w", err)
	}
	s.chatConsumer = consumer

	log.Printf("Kafka service initialized successfully")
	return nil
}

// PublishChatEvent publishes a chat event
func (s *KafkaService) PublishChatEvent(ctx context.Context, event *MQEvent) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.chatProducer == nil {
		return fmt.Errorf("chat producer not initialized")
	}

	return s.chatProducer.PublishEvent(ctx, event)
}

// StartChatConsumer starts consuming chat events
func (s *KafkaService) StartChatConsumer(ctx context.Context, eventHandler func(*MQEvent) error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.chatConsumer == nil {
		return fmt.Errorf("chat consumer not initialized")
	}

	return s.chatConsumer.StartConsuming(ctx, eventHandler)
}

// Close closes all Kafka connections
func (s *KafkaService) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var errors []error

	if s.chatProducer != nil {
		if err := s.chatProducer.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close chat producer: %w", err))
		}
	}

	if s.chatConsumer != nil {
		if err := s.chatConsumer.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close chat consumer: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing Kafka service: %v", errors)
	}

	log.Printf("Kafka service closed successfully")
	return nil
}
