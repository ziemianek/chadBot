package twitch

type WelcomeMessage struct {
	Payload struct {
		Session struct {
			Id                  string `json:"id"`
			Status              string `json:"status"`
			ConnectedAt         string `json:"connected_at"`
			KeepaliveTimeoutSec int    `json:"keepalive_timeout_seconds"`
			ReconnectUrl        string `json:"reconnect_url"`
			RecoveryUrl         string `json:"recovery_url"`
		} `json:"session"`
	} `json:"payload"`
}

type NotificationMessage struct {
	Payload struct {
		Event struct {
			ChatterUserName string `json:"chatter_user_name"`
			Message         struct {
				Text string `json:"text"`
			} `json:"message"`
		} `json:"event"`
	} `json:"payload"`
}

// structs below are used to send POST request
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
