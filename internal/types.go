package internal

type Session struct {
	Id                  string `json:"id"`
	Status              string `json:"status"`
	ConnectedAt         string `json:"connected_at"`
	KeepaliveTimeoutSec int    `json:"keepalive_timeout_seconds"`
	ReconnectUrl        string `json:"reconnect_url"`
	RecoveryUrl         string `json:"recovery_url"`
}

type WelcomePayload struct {
	Session Session `json:"session"`
}

type WelcomeMessage struct {
	Payload WelcomePayload `json:"payload"`
}

type NotificationPayload struct {
	Event Event `json:"event"`
}

type Event struct {
	ChatterUserName string      `json:"chatter_user_name"`
	Message         ChatMessage `json:"message"`
}

type ChatMessage struct {
	Text string `json:"text"`
}

type NotificationMessage struct {
	Payload NotificationPayload `json:"payload"`
}

type Metadata struct {
	MessageId   string `json:"message_id"`
	MessageType string `json:"message_type"`
	MessageTs   string `json:"message_timestamp"`
}

type BaseMessage struct {
	Metadata Metadata `json:"metadata"`
}

type headers map[string]string

type content struct {
	Type      string    `json:"type"`
	Version   string    `json:"version"`
	Condition condition `json:"condition"`
	Transport transport `json:"transport"`
}

type condition struct {
	BroadcasterUserId string `json:"broadcaster_user_id"`
	UserId            string `json:"user_id"`
}

type transport struct {
	Method    string `json:"method"`
	SessionId string `json:"session_id"`
}
