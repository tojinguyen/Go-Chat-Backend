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

type MQEventConsumer struct {
	consumer sarama.ConsumerGroup
	topic    string
	groupID  string
}

func NewChatEventConsumer(brokers []string, topic string, groupID string) (*MQEventConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Group.Session.Timeout = 10 * time.Second
	config.Consumer.Group.Heartbeat.Interval = 3 * time.Second
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	return &MQEventConsumer{
		consumer: consumer,
		topic:    topic,
		groupID:  groupID,
	}, nil
}

type ConsumerGroupHandler struct {
	eventHandler func(*MQEvent) error
}

func (h *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			var event MQEvent
			if err := json.Unmarshal(message.Value, &event); err != nil {
				log.Printf("Failed to unmarshal event: %v", err)
				session.MarkMessage(message, "")
				continue
			}

			if err := h.eventHandler(&event); err != nil {
				log.Printf("Failed to handle event: %v", err)
				// Continue processing other messages even if one fails
			}

			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

// StartConsuming starts consuming events with the provided handler
func (c *MQEventConsumer) StartConsuming(ctx context.Context, eventHandler func(*MQEvent) error) error {
	handler := &ConsumerGroupHandler{
		eventHandler: eventHandler,
	}

	for {
		if err := c.consumer.Consume(ctx, []string{c.topic}, handler); err != nil {
			log.Printf("Error from consumer: %v", err)
			return err
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

// Close closes the consumer
func (c *MQEventConsumer) Close() error {
	return c.consumer.Close()
}
