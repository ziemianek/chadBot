package twitch

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
)

type payloadSubscribeToChat struct {
	Type      string `json:"type"`
	Version   string `json:"version"`
	Condition struct {
		BroadcasterUserId string `json:"broadcaster_user_id"`
		UserId            string `json:"user_id"`
	} `json:"condition"`
	Transport struct {
		Method    string `json:"method"`
		SessionId string `json:"session_id"`
	} `json:"transport"`
}

type welcomeMessage struct {
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

func (m *welcomeMessage) handle(msg []byte) error {
	var err error
	var sessionId string
	err = json.Unmarshal(msg, m)
	sessionId = m.Payload.Session.Id
	log.Infof("Extracted session id: %v", sessionId)
	subscribe(sessionId)
	return err
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

func (m *notificationMessage) handle(ch chan string, msg []byte) error {
	var err error
	err = json.Unmarshal(msg, m)
	log.Debugf("Parsing notification message: %v", string(msg))
	// TODO: refactor this maybe?
	ch <- fmt.Sprintf("[%v] - [%v]: %v",
		parseTimestamp(m.Metadata.MsgTimestamp),
		m.Payload.Event.ChatterUserName,
		m.Payload.Event.Message.Text,
	)
	log.Infof("Sent new message: \"%v\" to message channel", m.Payload.Event.Message.Text)
	return err
}

func getMessageType(msg []byte) (string, error) {
	var err error
	var envelope struct {
		Metadata struct {
			MessageType string `json:"message_type"`
		} `json:"metadata"`
	}
	err = json.Unmarshal(msg, &envelope)
	return envelope.Metadata.MessageType, err
}

func subscribe(sessionId string) error {
	var url string = "https://api.twitch.tv/helix/eventsub/subscriptions"
	var err error
	var resp *http.Response
	//TODO: if all post requests get this, then move to callAPI method
	var headers RequestHeaders = RequestHeaders{"Content-Type": "application/json"}
	var broadcasterID string = getBroadcasterUserID()
	var body []byte
	body, err = json.Marshal(payloadSubscribeToChat{
		Type:    "channel.chat.message",
		Version: "1",
		Condition: struct {
			BroadcasterUserId string `json:"broadcaster_user_id"`
			UserId            string `json:"user_id"`
		}{
			BroadcasterUserId: broadcasterID,
			UserId:            broadcasterID,
		},
		Transport: struct {
			Method    string `json:"method"`
			SessionId string `json:"session_id"`
		}{
			Method:    "websocket",
			SessionId: sessionId,
		},
	})
	resp, err = SendPost(url, headers, body)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusAccepted {
		log.Info("Successfully subscribed to chat")
	} else {
		// TODO: make this more efficient by streaming data?
		out, _ := io.ReadAll(resp.Body)
		log.Errorf("Could not authorize: %v", string(out))
	}
	return nil
}

func parseTimestamp(ts string) string {
	var t time.Time
	var err error
	t, err = time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		log.Errorf("error parsing timestamp: %v\n", err)
		return ""
	}
	return t.Format("15:04:05")
}

func getBroadcasterUserID() string {
	var url string = "https://api.twitch.tv/helix/users"
	var err error
	var resp *http.Response
	resp, err = SendGet(url)
	if err != nil {
		log.Errorf("Could not send request: %v", err)
		return ""
	}
	var respBody struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	// stream data directly to struct
	// more efficient than io.ReadAll + json.Unmarshal
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		log.Errorf("Could not unmarshal response: %v", err)
		return ""
	}
	// SAFETY CHECK: Ensure we actually got data back before accessing index 0
	if len(respBody.Data) == 0 {
		log.Errorf("Twitch returned 0 users for this token. Check your credentials.")
		return ""
	}
	var id string = respBody.Data[0].ID
	log.Infof("Extracted Broadcaster User ID: %s", id)
	return id
}
