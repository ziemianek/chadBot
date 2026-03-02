package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"

	"github.com/charmbracelet/log"
	"github.com/gorilla/websocket"
	_ "github.com/joho/godotenv/autoload"
	"github.com/ziemianek/chadbot/internal/logger"
	"github.com/ziemianek/chadbot/pkg/twitch"
)

var (
	CLIENT_ID    = os.Getenv("CLIENT_ID")
	ACCESS_TOKEN = os.Getenv("ACCESS_TOKEN")
)

const WS_URL string = "wss://eventsub.wss.twitch.tv/ws"

// NEW: Struct for Chat Messages
type ChatNotification struct {
	Payload struct {
		Event struct {
			ChatterUserName string `json:"chatter_user_name"`
			ChatMessage     struct {
				Text string `json:"text"`
			} `json:"message"`
		} `json:"event"`
	} `json:"payload"`
}

type TwitchSubscribeCondition map[string]string

func Subscribe(sessionId string) (*http.Response, error) {
	var url string = "https://api.twitch.tv/helix/eventsub/subscriptions"
	payload, err := json.Marshal(twitch.TwitchEventSubscription{
		Type:      "channel.chat.message",
		Version:   "1",
		Condition: TwitchSubscribeCondition{"broadcaster_user_id": "482260799", "user_id": "482260799"},
		Transport: twitch.Transport{
			Method:    "websocket",
			SessionId: sessionId,
		},
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+ACCESS_TOKEN)
	req.Header.Set("Client-Id", CLIENT_ID)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	return client.Do(req)
}

func main() {
	var logger *log.Logger = logger.New(true)
	var check = func(err error) {
		if err != nil {
			logger.Error(err)
		}
	}

	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(WS_URL, nil)
	check(err)
	defer conn.Close()

	logger.Info("Connected to Twitch EventSub")

	for {
		_, msg, err := conn.ReadMessage()
		check(err)

		// 1. Peek at the message type using a simple anonymous struct
		var meta struct {
			Metadata struct {
				MessageType string `json:"message_type"`
			} `json:"metadata"`
		}
		json.Unmarshal(msg, &meta)

		switch meta.Metadata.MessageType {
		case "session_welcome":
			var welcomeMsg twitch.TwitchWelcomeMessage
			json.Unmarshal(msg, &welcomeMsg)
			sessionId := welcomeMsg.Payload.Session.Id

			logger.Infof("Welcome received. Subscribing with session: %s", sessionId)
			resp, subErr := Subscribe(sessionId)
			check(subErr)
			if resp.StatusCode == 202 {
				logger.Info("Subscription request ACCEPTED by Twitch")
			} else {
				logger.Warnf("Subscription failed with status: %s", resp.Status)
			}

		case "notification":
			var chatMsg ChatNotification
			json.Unmarshal(msg, &chatMsg)

			logger.Debug(chatMsg.Payload)
			// SUCCESS: This is where you see your chat!
			user := chatMsg.Payload.Event.ChatterUserName
			text := chatMsg.Payload.Event.ChatMessage.Text
			logger.Infof("CHAT [%s]: %s", user, text)

		case "session_keepalive":
			// Silence keepalives so logs stay clean
		}
	}
}
