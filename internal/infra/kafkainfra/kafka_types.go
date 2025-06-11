package kafkainfra

type MQEventType string

const (
	MessageSent    MQEventType = "message_sent"
	TypingStarted  MQEventType = "typing_started"
	TypingStopped  MQEventType = "typing_stopped"
	UserOnline     MQEventType = "user_online"
	UserOffline    MQEventType = "user_offline"
	UserJoinedRoom MQEventType = "user_joined_room"
	UserLeftRoom   MQEventType = "user_left_room"
)
