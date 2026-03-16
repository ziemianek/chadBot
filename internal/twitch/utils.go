package twitch

import (
	"bytes"
	tea "charm.land/bubbletea/v2"
	"encoding/json"
	"github.com/charmbracelet/log"
	_ "github.com/joho/godotenv/autoload"
	"io"
	"net/http"
	"os"
)

func SendPOST(url string, content content, headers map[string]string) (*http.Response, error) {
	var err error
	var body []byte
	var request *http.Request
	body, err = json.Marshal(content)
	if err != nil {
		return nil, err
	}
	request, err = http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		request.Header.Set(k, v)
	}
	return http.DefaultClient.Do(request)
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
	var content content = content{
		Type:    "channel.chat.message",
		Version: "1",
		Condition: condition{
			BroadcasterUserId: "482260799", //TODO: get broadcaster id from code
			UserId:            "482260799", // WHY THE SAME???
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

func ReadChatMsg(msgChan chan string) tea.Msg {
	// log.Info(<-msgChan)
	return ChatMsg(<-msgChan)
}
