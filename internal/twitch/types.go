package twitch

type RequestHeaders map[string]string

// TwitchUserResponse matches the Helix Users API response.
type TwitchUserResponse struct {
	Data []TwitchUser `json:"data"`
}

type TwitchUser struct {
	ID              string `json:"id"`
	Login           string `json:"login"`
	DisplayName     string `json:"display_name"`
	Description     string `json:"description"`
	ProfileImageURL string `json:"profile_image_url"`
}

// EventSub structs for internal processing.
type welcomeMessage struct {
	Payload struct {
		Session struct {
			Id string `json:"id"`
		} `json:"session"`
	} `json:"payload"`
}

type notificationMessage struct {
	Metadata struct {
		MsgTimestamp string `json:"message_timestamp"`
	} `json:"metadata"`
	Payload struct {
		Event struct {
			ChatterUserName string `json:"chatter_user_name"`
			Message         struct {
				Text string `json:"text"`
			} `json:"message"`
		} `json:"event"`
	} `json:"payload"`
}

type payloadSubscribeToChat struct {
	Type      string            `json:"type"`
	Version   string            `json:"version"`
	Condition map[string]string `json:"condition"`
	Transport map[string]string `json:"transport"`
}
