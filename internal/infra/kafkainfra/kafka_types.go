package kafkainfra

type MQEventTopic string

const (
	MessageSent   MQEventTopic = "message_sent"
	TypingStarted MQEventTopic = "typing_started"
	TypingStopped MQEventTopic = "typing_stopped"
	UserOnline    MQEventTopic = "user_online"
	UserOffline   MQEventTopic = "user_offline"
)
