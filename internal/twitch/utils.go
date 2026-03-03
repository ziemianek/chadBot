package twitch

import (
	"bytes"
	"encoding/json"
	"github.com/charmbracelet/log"
	_ "github.com/joho/godotenv/autoload"
	"io"
	"net/http"
	"os"
)

func SendPOST(url string, content content, headers headers) (*http.Response, error) {
	body, err := json.Marshal(content)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		log.Error(err)
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return http.DefaultClient.Do(req)
}

func subscribe(sessionId string) {
	var url string = "https://api.twitch.tv/helix/eventsub/subscriptions"
	var headers headers = headers{
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
	res, err := SendPOST(url, content, headers)
	if err != nil {
		log.Error(err)
	}
	if res.StatusCode == http.StatusAccepted {
		log.Info("Successfully subscribed to chat")
	} else {
		out, _ := io.ReadAll(res.Body)
		log.Errorf("Could not authorize: %v", string(out))
	}
}

func logChatMessage(event Event) {
	log.Infof("[%v]: %v", event.ChatterUserName, event.Message.Text)
}
