package kafkainfra

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
)

type MQEvent struct {
	Type       MQEventTopic `json:"type"`
	ChatRoomID string       `json:"chat_room_id"`
	UserID     string       `json:"user_id"`
	Timestamp  time.Time    `json:"timestamp"`
	Metadata   interface{}  `json:"metadata"`
}

type MQEventProducer struct {
	topic    string
	producer sarama.SyncProducer
}

func NewChatEventProducer(brokers []string, topic string) (*MQEventProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 3
	config.Producer.Compression = sarama.CompressionSnappy

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &MQEventProducer{
		topic:    topic,
		producer: producer,
	}, nil
}

func (p *MQEventProducer) PublishEvent(ctx context.Context, event *MQEvent) error {
	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Serialize event to JSON
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create Kafka message
	msg := &sarama.ProducerMessage{
		Topic:     p.topic,
		Key:       sarama.StringEncoder(event.ChatRoomID), // Partition by room ID
		Value:     sarama.ByteEncoder(eventBytes),
		Timestamp: event.Timestamp,
	}

	// Publish message
	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		log.Printf("Failed to publish event to Kafka: %v", err)
		return fmt.Errorf("failed to publish event: %w", err)
	}

	log.Printf("Event published successfully - Topic: %s, Partition: %d, Offset: %d", p.topic, partition, offset)
	return nil
}

func (p *MQEventProducer) Close() error {
	if err := p.producer.Close(); err != nil {
		return fmt.Errorf("failed to close producer: %w", err)
	}
	return nil
}
