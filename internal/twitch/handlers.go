package twitch

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/charmbracelet/log"
)

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
	var err error
	var response *http.Response
	var url string = "https://api.twitch.tv/helix/eventsub/subscriptions"
	var headers map[string]string = map[string]string{
		"Authorization": "Bearer " + os.Getenv("ACCESS_TOKEN"),
		"Client-Id":     os.Getenv("CLIENT_ID"),
		"Content-Type":  "application/json",
	}
	var bID string = getBroadcasterUserID()
	var content content = content{
		Type:    "channel.chat.message",
		Version: "1",
		Condition: condition{
			BroadcasterUserId: bID,
			UserId:            bID,
		},
		Transport: transport{
			Method:    "websocket",
			SessionId: sessionId,
		},
	}
	response, err = SendPOST(url, content, headers)
	if err != nil {
		log.Errorf("Got unexpected error: %v", err)
	}
	if response.StatusCode == http.StatusAccepted {
		log.Info("Successfully subscribed to chat")
	} else {
		out, _ := io.ReadAll(response.Body)
		log.Errorf("Could not authorize: %v", string(out))
	}
	return err
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
	var err error
	var request *http.Request
	var url string = "https://api.twitch.tv/helix/users"
	request, err = http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return ""
	}
	var headers map[string]string = map[string]string{
		"Authorization": "Bearer " + os.Getenv("ACCESS_TOKEN"),
		"Client-Id":     os.Getenv("CLIENT_ID"),
	}
	for k, v := range headers {
		request.Header.Set(k, v)
	}
	var response *http.Response
	response, err = http.DefaultClient.Do(request)
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
	err = json.NewDecoder(response.Body).Decode(&respBody)
	if err != nil {
		log.Errorf("Could not unmarshal response: %v", err)
		return ""
	}
	log.Infof("Extracted Broadcaster User ID: %v", respBody.Data[0].ID)
	return respBody.Data[0].ID
}
