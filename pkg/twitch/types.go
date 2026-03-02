package twitch

type TwitchResponseMetadata struct {
	MessageId        string `json:"message_id"`
	MessageType      string `json:"message_type"`
	MessageTimestamp string `json:"message_timestamp"`
}

type TwitchSession struct {
	Id                  string `json:"id"`
	Status              string `json:"status"`
	ConnectedAt         string `json:"connected_at"`
	KeepaliveTimeoutSec int    `json:"keepalive_timeout_seconds"`
	ReconnectUrl        any    `json:"reconnect_url"`
	RecoveryUrl         any    `json:"recovery_url"`
}

type TwitchMessagePayload struct {
	Session TwitchSession
}

type TwitchWelcomeMessage struct {
	Metadata TwitchResponseMetadata
	Payload  TwitchMessagePayload
}

type TwitchEventSubscription struct {
	Type      string            `json:"type"`
	Version   string            `json:"version"`
	Condition map[string]string `json:"condition"`
	Transport Transport         `json:"transport"`
}

type Transport struct {
	Method    string `json:"method"`
	SessionId string `json:"session_id"`
}
